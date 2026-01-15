package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/metrics"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/orchestrator"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/report"
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services"        // Register services
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/services/services" // Register service implementations
	_ "github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery/services" // Register discoverers
)

func main() {
	cloudName := flag.String("cloud", "", "The name of the cloud in clouds.yaml")
	policyPath := flag.String("policy", "", "Path to policies.yaml")
	outPath := flag.String("out", "", "Write findings to this file (default: policy defaults.output if set)")
	outFormat := flag.String("out-format", "jsonl", "Output format: jsonl, csv")
	workers := flag.Int("workers", runtime.NumCPU()*8, "Number of concurrent workers")
	fix := flag.Bool("fix", false, "Apply remediations for enforce-mode rules (default: false, dry-run)")
	allTenants := flag.Bool("all-tenants", false, "Scan all tenants/projects (requires admin). Default: false")
	jobsBuffer := flag.Int("jobs-buffer", 1000, "Jobs channel buffer size")
	resultsBuffer := flag.Int("results-buffer", 100, "Results channel buffer size")
	allowActions := flag.String("allow-actions", "", "Comma-separated list of remediation actions to allow (default: allow all)")
	metricsAddr := flag.String("metrics-addr", "", "Prometheus metrics listen address (e.g., :9090)")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	logFormat := flag.String("log-format", "text", "Log format: text, json")
	flag.Parse()

	if *cloudName == "" {
		*cloudName = os.Getenv("OS_CLOUD")
	}
	if *cloudName == "" {
		log.Fatal("Error: Please provide a cloud name via --cloud or OS_CLOUD env var")
	}

	if *policyPath == "" {
		log.Fatal("Error: Please provide a policy file via --policy")
	}

	configureLogger(*logLevel, *logFormat)

	fmt.Printf("Initializing Session for cloud: %q...\n", *cloudName)

	session, err := auth.NewSession(*cloudName)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("Authentication successful!")

	fmt.Printf("Loading policy from %q...\n", *policyPath)
	p, err := policy.Load(*policyPath)
	if err != nil {
		log.Fatalf("Failed to load policy: %v", err)
	}
	fmt.Printf("Policy loaded: %d service policies\n", len(p.Policies))

	workersCount := p.EffectiveWorkers(*workers)
	fmt.Printf("Using %d workers\n", workersCount)

	if *outPath == "" && p.Defaults.Output != "" {
		*outPath = p.Defaults.Output
	}

	var findingsWriter report.ResultWriter
	var findingsFile *os.File

	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("Failed to create output file %q: %v", *outPath, err)
		}
		findingsFile = f
		defer func() { _ = findingsFile.Close() }()
		writer, err := report.NewWriter(*outFormat, findingsFile)
		if err != nil {
			log.Fatalf("Failed to create output writer: %v", err)
		}
		findingsWriter = writer
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(p, session, workersCount, *fix, *allTenants)
	orch.SetBuffers(*jobsBuffer, *resultsBuffer)
	orch.SetRemediationAllowlist(parseAllowlist(*allowActions))
	defer orch.Stop()

	fmt.Println("Starting policy audit...")
	resultsChan, err := orch.Run()
	if err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	if *metricsAddr != "" {
		go func() {
			if err := metrics.StartServer(*metricsAddr); err != nil {
				log.Printf("Metrics server error: %v", err)
			}
		}()
	}

	summaryChan := make(chan report.Summary, 1)
	go func() {
		summaryChan <- report.ConsumeResults(resultsChan, findingsWriter)
	}()

	summary := <-summaryChan
	report.PrintSummary(os.Stdout, summary)
	if findingsWriter != nil {
		fmt.Printf("Findings written: %d\nOutput: %s\n", summary.Written, *outPath)
	} else {
		fmt.Println("Findings written: 0 (no --out specified)")
	}
	if summary.Violations > 0 {
		os.Exit(2)
	}
}

func parseAllowlist(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var actions []string
	for _, part := range parts {
		action := strings.TrimSpace(part)
		if action == "" {
			continue
		}
		actions = append(actions, action)
	}
	return actions
}

func configureLogger(level, format string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: parseLogLevel(level)}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

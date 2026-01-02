package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
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
	outPath := flag.String("out", "", "Write JSONL findings to this file (default: policy defaults.output if set)")
	workers := flag.Int("workers", runtime.NumCPU()*8, "Number of concurrent workers")
	apply := flag.Bool("apply", false, "Apply remediations for enforce-mode rules (default: false, dry-run)")
	fix := flag.Bool("fix", false, "Alias for --apply")
	allTenants := flag.Bool("all-tenants", false, "Scan all tenants/projects (requires admin). Default: false")
	flag.Parse()

	// --fix is an alias for --apply
	if *fix {
		*apply = true
	}

	if *cloudName == "" {
		*cloudName = os.Getenv("OS_CLOUD")
	}
	if *cloudName == "" {
		log.Fatal("Error: Please provide a cloud name via --cloud or OS_CLOUD env var")
	}

	if *policyPath == "" {
		log.Fatal("Error: Please provide a policy file via --policy")
	}

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

	var findingsWriter *report.JSONLWriter
	var findingsFile *os.File

	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("Failed to create output file %q: %v", *outPath, err)
		}
		findingsFile = f
		defer func() { _ = findingsFile.Close() }()
		findingsWriter = report.NewJSONLWriter(findingsFile)
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(p, session, workersCount, *apply, *allTenants)
	defer orch.Stop()

	fmt.Println("Starting policy audit...")
	resultsChan, err := orch.Run()
	if err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	var scanned, violations, errors int
	var written int

	for result := range resultsChan {
		scanned++
		if result.Error != nil {
			errors++
		}
		if result.RemediationError != nil {
			errors++
		}
		if !result.Compliant {
			violations++
		}

		// Stream findings: write only non-compliant (including errors/remediation errors).
		if findingsWriter != nil && (!result.Compliant || result.Error != nil || result.RemediationError != nil) {
			if err := findingsWriter.WriteResult(result); err != nil {
				log.Printf("Failed to write finding: %v", err)
			} else {
				written++
			}
		}
	}

	report.PrintSummary(os.Stdout, scanned, violations, errors)
	if findingsWriter != nil {
		fmt.Printf("Findings written: %d\nOutput: %s\n", written, *outPath)
	} else {
		fmt.Println("Findings written: 0 (no --out specified)")
	}
	if violations > 0 {
		os.Exit(2)
	}
}

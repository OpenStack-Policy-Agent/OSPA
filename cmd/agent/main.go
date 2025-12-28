package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/auth"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/engine"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/report"
)

func main() {
	cloudName := flag.String("cloud", "", "The name of the cloud in clouds.yaml")
	policyPath := flag.String("policy", "", "Path to rules.yaml (MVP mode)")
	outPath := flag.String("out", "", "Write JSONL findings to this file (default: policy defaults.output if set)")
	days := flag.Int("days", 30, "Find SHUTOFF servers whose Updated timestamp is older than this many days (approximation)")
	workers := flag.Int("workers", runtime.NumCPU()*8, "Number of concurrent workers")
	apply := flag.Bool("apply", false, "Apply remediations for enforce-mode rules (default: false, dry-run)")
	allTenants := flag.Bool("all-tenants", false, "Scan all tenants/projects (requires admin). Default: false")
	flag.Parse()

	if *cloudName == "" {
		*cloudName = os.Getenv("OS_CLOUD")
	}
	if *cloudName == "" {
		log.Fatal("Error: Please provide a cloud name via --cloud or OS_CLOUD env var")
	}

	fmt.Println("cloudName", *cloudName)

	fmt.Printf("Initializing Session for cloud: %q...\n", *cloudName)

	session, err := auth.NewSession(*cloudName)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("Authentication successful!")

	fmt.Println("Creating Compute (Nova) client...")
	computeClient, err := session.GetComputeClient()
	if err != nil {
		log.Fatalf("Failed to get compute client: %v", err)
	}
	fmt.Printf("Connected to Compute Endpoint: %s\n", computeClient.Endpoint)

	jobs := make(chan engine.Job, 1024)
	results := make(chan engine.Result, 1024)

	var evaluator engine.Evaluator
	var findingsWriter *report.JSONLWriter
	var findingsFile *os.File

	if *policyPath != "" {
		p, err := policy.Load(*policyPath)
		if err != nil {
			log.Fatalf("Failed to load policy: %v", err)
		}

		rules := make([]engine.ServerRule, 0, len(p.Rules))
		for _, r := range p.Rules {
			// MVP: only supports compute.server + stopped instances rule shape.
			rules = append(rules, engine.StoppedOlderThanRule{
				RuleID:            r.ID,
				MatchStatus:       r.Filters.Status,
				OlderThan:         time.Duration(r.Conditions.UpdatedOlderThanDays) * 24 * time.Hour,
				RecommendedAction: r.EffectiveRemediation(),
				Mode:              r.Mode,
			})
		}

		*workers = p.EffectiveWorkers(*workers)
		evaluator = engine.RuleSet{Now: time.Now, Rules: rules}

		if *outPath == "" && p.Defaults.Output != "" {
			*outPath = p.Defaults.Output
		}
	} else {
		// POC fallback (no policy file): single hardcoded rule driven by flags.
		evaluator = engine.RuleSet{
			Now: time.Now,
			Rules: []engine.ServerRule{
				engine.StoppedOlderThanRule{
					RuleID:            "poc.stopped_instances_older_than",
					MatchStatus:       "SHUTOFF",
					OlderThan:         time.Duration(*days) * 24 * time.Hour,
					RecommendedAction: "delete",
					Mode:              "audit",
				},
			},
		}
	}

	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("Failed to create output file %q: %v", *outPath, err)
		}
		findingsFile = f
		defer func() { _ = findingsFile.Close() }()
		findingsWriter = report.NewJSONLWriter(findingsFile)
	}

	var wg sync.WaitGroup
	engine.StartWorkerPool(*workers, *apply, computeClient, evaluator, jobs, results, &wg)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Discovery closes the jobs channel when done
	discovery.DiscoverServers(computeClient, *allTenants, jobs)

	var scanned, violations, errors int
	var written int
	for r := range results {
		scanned++
		if r.Error != nil {
			errors++
		}
		if r.RemediationError != nil {
			errors++
		}
		if !r.Compliant {
			violations++
		}

		// Stream findings: write only non-compliant (including errors/remediation errors).
		if findingsWriter != nil && (!r.Compliant || r.Error != nil || r.RemediationError != nil) {
			if err := findingsWriter.WriteResult(r); err != nil {
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

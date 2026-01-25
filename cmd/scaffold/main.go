package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/generators"
	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/registry"
)

var (
	serviceName = flag.String("service", "", "Service name (e.g., nova, neutron, cinder)")
	resources   = flag.String("resources", "", "Comma-separated list of resource types (e.g., instance,keypair)")
	list        = flag.Bool("list", false, "List all available OpenStack services and resources")
)

func main() {
	flag.Parse()

	// Handle list command
	if *list {
		listServices()
		return
	}

	if *serviceName == "" {
		fmt.Fprintf(os.Stderr, "Error: --service is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --service <name> --resources <list>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --service nova --resources instance,keypair\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nList available services:\n")
		fmt.Fprintf(os.Stderr, "  %s --list\n", os.Args[0])
		os.Exit(1)
	}

	// Normalize service name
	serviceNameLower := strings.ToLower(*serviceName)

	// Validate service exists in registry
	if err := registry.ValidateService(serviceNameLower); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUse --list to see all available services\n")
		os.Exit(1)
	}

	// Get service info from registry (provides display name, service type, etc.)
	serviceInfo, err := registry.GetServiceInfo(serviceNameLower)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse resources
	resourceList := []string{}
	if *resources != "" {
		resourceList = strings.Split(*resources, ",")
		for i, r := range resourceList {
			resourceList[i] = strings.TrimSpace(strings.ToLower(r))
		}
	}

	if len(resourceList) == 0 {
		fmt.Fprintf(os.Stderr, "Error: at least one resource type is required (--resources)\n")
		availableResources, _ := registry.ListResources(serviceNameLower)
		if len(availableResources) > 0 {
			fmt.Fprintf(os.Stderr, "Available resources for %s: %v\n", serviceNameLower, availableResources)
		}
		os.Exit(1)
	}

	// Validate resources exist for this service
	if err := registry.ValidateResources(serviceNameLower, resourceList); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate files (preserves existing resources, adds new ones)
	if err := generators.GenerateService(serviceNameLower, serviceInfo.DisplayName, serviceInfo.ServiceType, resourceList); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully generated files for service '%s'\n", serviceNameLower)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("───────────────────────────────────────────────────────────────────────")
	fmt.Printf("1. Update checks in cmd/scaffold/internal/registry/config/%s.yaml\n", serviceNameLower)
	fmt.Println("   to match what the OpenStack API actually supports for each resource.")
	fmt.Println()
	fmt.Println("2. Implement the generated discoverers in:")
	fmt.Printf("   pkg/discovery/services/%s.go\n", serviceNameLower)
	fmt.Println("   using gophercloud. See: https://pkg.go.dev/github.com/gophercloud/gophercloud")
	fmt.Println()
	fmt.Println("3. Implement the generated auditors in:")
	fmt.Printf("   pkg/audit/%s/\n", serviceNameLower)
	fmt.Println("   with real field extraction and check logic.")
	fmt.Println()
	fmt.Printf("4. Run tests: go test ./pkg/audit/%s/...\n", serviceNameLower)
	fmt.Printf("5. Review policy guide: docs/reference/services/%s.md\n", serviceNameLower)
}

// listServices prints all available OpenStack services and their resources
func listServices() {
	services := registry.ListServices()
	if len(services) == 0 {
		fmt.Println("No services available")
		return
	}

	fmt.Println("Available OpenStack Services:")
	fmt.Println("==============================")

	for _, svcName := range services {
		info, err := registry.GetServiceInfo(svcName)
		if err != nil {
			continue
		}

		fmt.Printf("\n%s (%s)\n", info.DisplayName, svcName)
		fmt.Printf("  Service Type: %s\n", info.ServiceType)
		fmt.Printf("  Resources: ")

		resList, err := registry.ListResources(svcName)
		if err != nil {
			fmt.Println("(error listing resources)")
			continue
		}

		for i, res := range resList {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(res)
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Printf("  %s --service nova --resources instance,keypair\n", os.Args[0])
}

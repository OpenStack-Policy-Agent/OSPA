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
	serviceName = flag.String("service", "", "Service name (e.g., glance, keystone)")
	displayName = flag.String("display-name", "", "Display name for the service (e.g., Glance, Keystone)")
	resources   = flag.String("resources", "", "Comma-separated list of resource types (e.g., image,member)")
	serviceType = flag.String("type", "", "OpenStack service type for client (e.g., image, identity)")
	force       = flag.Bool("force", false, "Overwrite existing files")
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
		fmt.Fprintf(os.Stderr, "Usage: %s --service <name> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --service glance --display-name Glance --resources image,member --type image\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nList available services:\n")
		fmt.Fprintf(os.Stderr, "  %s --list\n", os.Args[0])
		os.Exit(1)
	}

	// Normalize service name
	serviceNameLower := strings.ToLower(*serviceName)

	// Validate service exists in OpenStack
	if err := registry.ValidateService(serviceNameLower); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUse --list to see all available services\n")
		os.Exit(1)
	}

	// Get service info for defaults
	serviceInfo, err := registry.GetServiceInfo(serviceNameLower)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Set defaults from registry if not provided
	if *displayName == "" {
		*displayName = serviceInfo.DisplayName
	}
	if *serviceType == "" {
		*serviceType = serviceInfo.ServiceType
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

	// Analyze existing service
	analysis, err := generators.AnalyzeService(".", serviceNameLower, resourceList)
	if err == nil && analysis.ServiceExists {
		// Service exists - show what will be updated
		if len(analysis.MissingResources) > 0 {
			fmt.Printf("Service '%s' already exists. Adding new resources: %v\n", serviceNameLower, analysis.MissingResources)
			if len(analysis.ExistingResources) > 0 {
				fmt.Printf("Existing resources: %v\n", analysis.ExistingResources)
			}
		} else {
			fmt.Printf("Service '%s' already exists with all requested resources.\n", serviceNameLower)
			if len(analysis.MissingFiles) > 0 {
				fmt.Printf("Generating missing files: %v\n", analysis.MissingFiles)
			} else {
				fmt.Println("All files already exist. Use --force to overwrite.")
				return
			}
		}
	}

	// Generate files
	if err := generators.GenerateService(serviceNameLower, *displayName, *serviceType, resourceList, *force); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully generated/updated files for service '%s'\n", serviceNameLower)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Review and customize the generated files\n")
	fmt.Printf("2. Test with: go test ./pkg/audit/%s/...\n", serviceNameLower)
	fmt.Printf("3. Run e2e tests: go test -tags=e2e ./e2e/%s_test.go\n", serviceNameLower)
	fmt.Printf("4. Review policy guide: examples/policies/%s-policy-guide.md\n", serviceNameLower)
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
	
	fmt.Printf("\nExample usage:\n")
	fmt.Printf("  %s --service glance --resources image,member --type image\n", os.Args[0])
}

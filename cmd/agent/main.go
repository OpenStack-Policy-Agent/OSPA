package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yourname/openstack-agent/pkg/auth"
)

func main() {
	// 1. Parse Command Line Arguments
	cloudName := flag.String("cloud", "", "The name of the cloud in clouds.yaml")
	flag.Parse()

	// If no flag provided, try the environment variable
	if *cloudName == "" {
		*cloudName = os.Getenv("OS_CLOUD")
	}

	if *cloudName == "" {
		log.Fatal("Error: Please provide a cloud name via --cloud or OS_CLOUD env var")
	}

	fmt.Printf("Initializing Session for cloud: '%s'...\n", *cloudName)

	// 2. Initialize Session
	session, err := auth.NewSession(*cloudName)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Authentication successful!")

	// 3. Test Connection (Compute)
	fmt.Println("Testing Compute (Nova) connection...")
	computeClient, err := session.GetComputeClient()
	if err != nil {
		log.Fatalf("Failed to get compute client: %v", err)
	}

	fmt.Printf("✅ Connected to Compute Endpoint: %s\n", computeClient.Endpoint)
}

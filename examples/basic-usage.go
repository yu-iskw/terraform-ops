package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yu/terraform-ops/internal/terraform/config"
	"github.com/yu/terraform-ops/pkg/terraform"
)

func main() {
	// Example 1: Using the Terraform client
	fmt.Println("=== Example 1: Terraform Client Usage ===")
	client := terraform.NewClient("terraform", "/path/to/terraform/project")

	ctx := context.Background()

	// Initialize Terraform
	fmt.Println("Running terraform init...")
	if err := client.Init(ctx); err != nil {
		log.Printf("Failed to initialize Terraform: %v", err)
	}

	// Plan the changes
	fmt.Println("Running terraform plan...")
	if err := client.Plan(ctx); err != nil {
		log.Printf("Failed to plan Terraform changes: %v", err)
	}

	fmt.Println("Terraform operations completed successfully!")

	// Example 2: Using show-terraform functionality
	fmt.Println("\n=== Example 2: Show Terraform Configuration ===")

	// Example workspace path (this would be a real workspace in practice)
	workspacePath := "./test/workspaces/gcs-backend"

	// Get Terraform information from the workspace using the new config parser
	parser := config.NewParser()
	infos, err := parser.ParseConfigFiles([]string{workspacePath})
	if err != nil {
		log.Printf("Error getting Terraform info: %v", err)
		return
	}

	if len(infos) == 0 {
		log.Printf("No Terraform information found")
		return
	}

	info := infos[0]
	fmt.Printf("Workspace: %s\n", info.Path)
	fmt.Printf("Required Version: %s\n", info.RequiredVersion)

	// Check if backend is configured
	if info.Backend != nil {
		fmt.Printf("Backend Type: %s\n", info.Backend.Type)
		fmt.Printf("Backend Config:\n")
		for key, value := range info.Backend.Config {
			fmt.Printf("  %s: %s\n", key, value)
		}

		// Specifically check for impersonate_service_account
		if impersonateSA, exists := info.Backend.Config["impersonate_service_account"]; exists {
			fmt.Printf("\nService Account Impersonation: %s\n", impersonateSA)
		}
	}

	// Display required providers
	fmt.Printf("\nRequired Providers:\n")
	for provider, version := range info.RequiredProviders {
		if version == "" {
			fmt.Printf("  %s: (no version constraint)\n", provider)
		} else {
			fmt.Printf("  %s: %s\n", provider, version)
		}
	}

	// Example of JSON output
	fmt.Printf("\nJSON Output:\n")
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Println(string(jsonData))
	}
}

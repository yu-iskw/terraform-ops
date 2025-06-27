package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yu/terraform-ops/pkg/terraform"
)

func main() {
	// Create a new Terraform client
	client := terraform.NewClient("terraform", "/path/to/terraform/project")

	ctx := context.Background()

	// Initialize Terraform
	fmt.Println("Running terraform init...")
	if err := client.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize Terraform: %v", err)
	}

	// Plan the changes
	fmt.Println("Running terraform plan...")
	if err := client.Plan(ctx); err != nil {
		log.Fatalf("Failed to plan Terraform changes: %v", err)
	}

	fmt.Println("Terraform operations completed successfully!")
}

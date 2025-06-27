package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terraform-ops",
	Short: "Terraform operations CLI tool",
	Long:  `A CLI tool for managing Terraform operations and workflows`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Terraform Ops CLI v1.0.0")
		fmt.Println("Use --help for more information")
	},
}

func init() {
	// Add global flags here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

// Run executes the root command
func Run() error {
	return rootCmd.Execute()
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

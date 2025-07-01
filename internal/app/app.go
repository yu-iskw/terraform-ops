// Copyright 2025 yu-iskw
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yu/terraform-ops/internal/commands"
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
	rootCmd.AddCommand(commands.DefaultShowTerraformCommand().Command())
	rootCmd.AddCommand(commands.DefaultPlanGraphCommand().Command())
	rootCmd.AddCommand(commands.DefaultSummarizePlanCommand().Command())
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

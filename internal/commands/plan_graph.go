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

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/terraform/graph"
	"github.com/yu/terraform-ops/internal/terraform/graph/generators"
	"github.com/yu/terraform-ops/internal/terraform/plan"
)

// PlanGraphCommand represents the plan-graph command with dependency injection
type PlanGraphCommand struct {
	planParser   core.PlanParser
	graphBuilder core.GraphBuilder
	genFactory   *generators.Factory
}

// NewPlanGraphCommand creates a new plan-graph command with injected dependencies
func NewPlanGraphCommand(
	planParser core.PlanParser,
	graphBuilder core.GraphBuilder,
	genFactory *generators.Factory,
) *PlanGraphCommand {
	return &PlanGraphCommand{
		planParser:   planParser,
		graphBuilder: graphBuilder,
		genFactory:   genFactory,
	}
}

// Command returns the cobra command for plan-graph
func (c *PlanGraphCommand) Command() *cobra.Command {
	var opts core.GraphOptions

	cmd := &cobra.Command{
		Use:   "plan-graph <PLAN_FILE>",
		Short: "Generate a visual graph representation of Terraform plan changes",
		Long: `Generate a visual graph representation of Terraform plan changes for the given workspace.
The generated graph shows relationships between resources, grouped by modules, with clear indication of resource lifecycle changes (create, update, delete).
Dependencies between resources are always shown. Output values, variables, local values, and data sources are shown by default with different colors and shapes for each type.

Supported output formats:
- graphviz: Graphviz DOT format (default)
- mermaid: Mermaid diagram format
- plantuml: PlantUML format

Node types and their visual representation:
- Resources: House shapes (with action-based colors)
- Data Sources: Cyan diamonds
- Outputs: Blue inverted houses (exports)
- Variables: Yellow cylinders (inputs)
- Locals: Pink octagons (computed values)

Examples:
  terraform-ops plan-graph plan.json
  terraform-ops plan-graph --format mermaid plan.json
  terraform-ops plan-graph --no-outputs plan.json
  terraform-ops plan-graph --no-variables --no-locals plan.json
  terraform-ops plan-graph --no-data-sources --no-outputs --no-variables plan.json
  terraform-ops plan-graph --output graph.dot plan.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runPlanGraph(args[0], opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP((*string)(&opts.Format), "format", "f", string(core.FormatGraphviz), "Output format (graphviz, mermaid, plantuml)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP((*string)(&opts.GroupBy), "group-by", "g", string(core.GroupByModule), "Grouping strategy (module, action, resource_type)")
	cmd.Flags().BoolVar(&opts.NoDataSources, "no-data-sources", false, "Exclude data source resources from the graph")
	cmd.Flags().BoolVar(&opts.NoOutputs, "no-outputs", false, "Exclude output values from the graph")
	cmd.Flags().BoolVar(&opts.NoVariables, "no-variables", false, "Exclude variable values from the graph")
	cmd.Flags().BoolVar(&opts.NoLocals, "no-locals", false, "Exclude local values from the graph")
	cmd.Flags().BoolVarP(&opts.Compact, "compact", "c", false, "Generate a more compact graph layout")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output for debugging")

	return cmd
}

// runPlanGraph executes the plan-graph command
func (c *PlanGraphCommand) runPlanGraph(planFile string, opts core.GraphOptions) error {
	// Validate format
	if !isValidFormat(opts.Format) {
		return &core.UnsupportedFormatError{Format: string(opts.Format)}
	}

	// Validate grouping strategy
	if !isValidGrouping(opts.GroupBy) {
		return fmt.Errorf("unsupported grouping: %s. Supported groupings: module, action, resource_type", opts.GroupBy)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Parsing plan file: %s\n", planFile)
	}

	// Parse the plan file
	plan, err := c.planParser.ParsePlanFile(planFile)
	if err != nil {
		return fmt.Errorf("failed to parse plan file: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d resource changes\n", len(plan.ResourceChanges))
	}

	// Build graph data
	graphData, err := c.graphBuilder.BuildGraph(plan, opts)
	if err != nil {
		return fmt.Errorf("failed to build graph data: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generated graph with %d nodes and %d edges\n", len(graphData.Nodes), len(graphData.Edges))
	}

	// Create generator for the specified format
	generator, err := c.genFactory.CreateGenerator(opts.Format)
	if err != nil {
		return fmt.Errorf("failed to create graph generator: %w", err)
	}

	// Generate the graph
	graphOutput, err := generator.Generate(graphData, opts)
	if err != nil {
		return fmt.Errorf("failed to generate graph: %w", err)
	}

	// Write output
	if opts.Output != "" {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", opts.Output)
		}
		if err := os.WriteFile(opts.Output, []byte(graphOutput), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Print(graphOutput)
	}

	return nil
}

// isValidFormat checks if the format is supported
func isValidFormat(format core.GraphFormat) bool {
	switch format {
	case core.FormatGraphviz, core.FormatMermaid, core.FormatPlantUML:
		return true
	default:
		return false
	}
}

// isValidGrouping checks if the grouping strategy is supported
func isValidGrouping(grouping core.GroupingStrategy) bool {
	switch grouping {
	case core.GroupByModule, core.GroupByAction, core.GroupByResourceType:
		return true
	default:
		return false
	}
}

// DefaultPlanGraphCommand creates a plan-graph command with default dependencies
func DefaultPlanGraphCommand() *PlanGraphCommand {
	return NewPlanGraphCommand(
		plan.NewParser(),
		graph.NewBuilder(),
		generators.NewFactory(),
	)
}

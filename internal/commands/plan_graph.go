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
	"github.com/yu/terraform-ops/internal/ir"
	terraformsource "github.com/yu/terraform-ops/internal/source/terraform"
	"github.com/yu/terraform-ops/internal/terraform/graph"
	"github.com/yu/terraform-ops/internal/terraform/graph/generators"
)

// PlanGraphCommand renders the dependency graph already normalized into a
// ChangeSet. It does not parse Terraform configuration or rediscover edges.
type PlanGraphCommand struct {
	graphBuilder core.GraphBuilder
	genFactory   *generators.Factory
}

func NewPlanGraphCommand(
	graphBuilder core.GraphBuilder,
	genFactory *generators.Factory,
) *PlanGraphCommand {
	return &PlanGraphCommand{graphBuilder: graphBuilder, genFactory: genFactory}
}

func (c *PlanGraphCommand) Command() *cobra.Command {
	var opts core.GraphOptions

	cmd := &cobra.Command{
		Use:   "plan-graph <PLAN_FILE>",
		Short: "Generate a visual graph representation of Terraform plan changes",
		Long: `Generate a visual graph representation of Terraform plan changes.
The command loads the same sanitized ChangeSet used by terraform-ops analyze and renders its normalized dependency graph.
Resources, data sources, root outputs, and provided root variables are represented when present in machine-readable plan JSON.

Supported output formats:
- graphviz: Graphviz DOT format (default)
- mermaid: Mermaid diagram format
- plantuml: PlantUML format

Examples:
  terraform-ops plan-graph plan.json
  terraform-ops plan-graph --format mermaid plan.json
  terraform-ops plan-graph --no-outputs plan.json
  terraform-ops plan-graph --no-variables plan.json
  terraform-ops plan-graph --no-data-sources --no-outputs --no-variables plan.json
  terraform-ops plan-graph --no-modules plan.json
  terraform-ops plan-graph --output graph.dot plan.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runPlanGraph(args[0], opts)
		},
	}

	cmd.Flags().StringVarP((*string)(&opts.Format), "format", "f", string(core.FormatGraphviz), "Output format (graphviz, mermaid, plantuml)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP((*string)(&opts.GroupBy), "group-by", "g", string(core.GroupByModule), "Grouping strategy (module, action, resource_type)")
	cmd.Flags().BoolVar(&opts.NoDataSources, "no-data-sources", false, "Exclude data source resources from the graph")
	cmd.Flags().BoolVar(&opts.NoOutputs, "no-outputs", false, "Exclude root output values from the graph")
	cmd.Flags().BoolVar(&opts.NoVariables, "no-variables", false, "Exclude provided root variables from the graph")
	cmd.Flags().BoolVar(&opts.NoLocals, "no-locals", false, "Compatibility flag; local declarations are not exposed by plan JSON")
	cmd.Flags().BoolVar(&opts.NoModules, "no-modules", false, "Exclude resources from modules from the graph")
	cmd.Flags().BoolVarP(&opts.Compact, "compact", "c", false, "Generate a more compact graph layout")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output for debugging")

	return cmd
}

func (c *PlanGraphCommand) runPlanGraph(planFile string, opts core.GraphOptions) error {
	if !isValidFormat(opts.Format) {
		return &core.UnsupportedFormatError{Format: string(opts.Format)}
	}
	if !isValidGrouping(opts.GroupBy) {
		return fmt.Errorf("unsupported grouping: %s. Supported groupings: module, action, resource_type", opts.GroupBy)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Loading normalized plan: %s\n", planFile)
	}
	changeSet, err := terraformsource.LoadFile(
		planFile,
		terraformsource.DefaultMaxPlanBytes,
		ir.EngineUnknown,
		ir.RedactionStandard,
	)
	if err != nil {
		return fmt.Errorf("failed to load plan file: %w", err)
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d normalized graph nodes and %d evidence edges\n", len(changeSet.Graph.Nodes), len(changeSet.Graph.Edges))
	}

	graphData, err := c.graphBuilder.BuildGraph(changeSet, opts)
	if err != nil {
		return fmt.Errorf("failed to build graph data: %w", err)
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generated graph with %d nodes and %d edges\n", len(graphData.Nodes), len(graphData.Edges))
	}

	generator, err := c.genFactory.CreateGenerator(opts.Format)
	if err != nil {
		return fmt.Errorf("failed to create graph generator: %w", err)
	}
	graphOutput, err := generator.Generate(graphData, opts)
	if err != nil {
		return fmt.Errorf("failed to generate graph: %w", err)
	}

	if opts.Output != "" {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", opts.Output)
		}
		if err := os.WriteFile(opts.Output, []byte(graphOutput), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Print(graphOutput)
	}
	return nil
}

func isValidFormat(format core.GraphFormat) bool {
	switch format {
	case core.FormatGraphviz, core.FormatMermaid, core.FormatPlantUML:
		return true
	default:
		return false
	}
}

func isValidGrouping(grouping core.GroupingStrategy) bool {
	switch grouping {
	case core.GroupByModule, core.GroupByAction, core.GroupByResourceType:
		return true
	default:
		return false
	}
}

func DefaultPlanGraphCommand() *PlanGraphCommand {
	return NewPlanGraphCommand(
		graph.NewBuilder(),
		generators.NewFactory(),
	)
}

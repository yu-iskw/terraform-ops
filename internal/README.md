# Software Architecture

## Overview

The `internal` package contains the core application logic for the terraform-ops CLI tool. The architecture follows clean architecture principles with clear separation of concerns, dependency injection, and interface-based design.

## High-Level Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        A[cmd/terraform-ops] --> B[internal/app]
    end

    subgraph "Application Layer"
        B --> C[internal/commands]
        B --> D[internal/config]
    end

    subgraph "Core Domain"
        E[internal/core]
    end

    subgraph "Infrastructure Layer"
        F[internal/terraform]
        F --> G[internal/terraform/config]
        F --> H[internal/terraform/plan]
        F --> I[internal/terraform/graph]
        I --> J[internal/terraform/graph/generators]
    end

    C --> E
    D --> E
    F --> E

    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style E fill:#e8f5e8
    style F fill:#fff3e0
```

## Package Structure

```mermaid
graph LR
    subgraph "internal/"
        subgraph "app/"
            A1[app.go] --> A2[app_test.go]
        end

        subgraph "commands/"
            B1[plan_graph.go]
        end

        subgraph "config/"
            C1[config.go]
        end

        subgraph "core/"
            D1[types.go] --> D2[errors.go]
            D2 --> D3[errors_test.go]
            D1 --> D4[types_test.go]
        end

        subgraph "terraform/"
            subgraph "config/"
                E1[parser.go] --> E2[parser_test.go]
            end

            subgraph "plan/"
                F1[parser.go] --> F2[parser_test.go]
            end

            subgraph "graph/"
                G1[builder.go] --> G2[builder_test.go]
                G1 --> G3[dependencies.go]

                subgraph "generators/"
                    H1[factory.go] --> H2[factory_test.go]
                    H3[graphviz.go] --> H4[graphviz_test.go]
                    H5[mermaid.go]
                    H6[plantuml.go]
                end
            end
        end
    end

    A1 --> B1
    A1 --> E1
    B1 --> G1
    B1 --> F1
    G1 --> H1
    H1 --> H3
    H1 --> H5
    H1 --> H6
```

## Core Domain Model

```mermaid
classDiagram
    class PlanParser {
        <<interface>>
        +ParsePlanFile(filename string) (*TerraformPlan, error)
    }

    class ConfigParser {
        <<interface>>
        +ParseConfigFiles(paths []string) ([]TerraformConfig, error)
    }

    class GraphBuilder {
        <<interface>>
        +BuildGraph(plan *TerraformPlan, opts GraphOptions) (*GraphData, error)
    }

    class GraphGenerator {
        <<interface>>
        +Generate(graphData *GraphData, opts GraphOptions) (string, error)
    }

    class TerraformPlan {
        +FormatVersion string
        +ResourceChanges []ResourceChange
        +OutputChanges map[string]OutputChange
        +Configuration Configuration
        +Variables map[string]Variable
        +Applicable bool
        +Complete bool
        +Errored bool
    }

    class GraphData {
        +Nodes []GraphNode
        +Edges []GraphEdge
    }

    class GraphOptions {
        +Format GraphFormat
        +Output string
        +GroupBy GroupingStrategy
        +NoDataSources bool
        +NoOutputs bool
        +NoVariables bool
        +NoLocals bool
        +Compact bool
        +Verbose bool
    }

    PlanParser --> TerraformPlan
    ConfigParser --> TerraformConfig
    GraphBuilder --> GraphData
    GraphGenerator --> GraphData
    GraphBuilder --> GraphOptions
    GraphGenerator --> GraphOptions
```

## Dependency Flow

```mermaid
flowchart TD
    A[CLI Command] --> B[Command Handler]
    B --> C[Plan Parser]
    B --> D[Graph Builder]
    B --> E[Graph Generator Factory]

    C --> F[Parse JSON Plan]
    F --> G[Validate Plan Structure]
    G --> H[TerraformPlan Object]

    D --> I[Extract Resource Changes]
    D --> J[Extract Output Changes]
    D --> K[Extract Variables]
    D --> L[Extract Locals]
    D --> M[Analyze Dependencies]

    I --> N[Create Graph Nodes]
    J --> N
    K --> N
    L --> N
    M --> O[Create Graph Edges]

    N --> P[GraphData Object]
    O --> P

    E --> Q[Select Generator Type]
    Q --> R[Graphviz Generator]
    Q --> S[Mermaid Generator]
    Q --> T[PlantUML Generator]

    P --> R
    P --> S
    P --> T

    R --> U[Generate DOT Format]
    S --> V[Generate Mermaid Format]
    T --> W[Generate PlantUML Format]

    style A fill:#e3f2fd
    style H fill:#e8f5e8
    style P fill:#fff3e0
    style U fill:#fce4ec
    style V fill:#fce4ec
    style W fill:#fce4ec
```

## Error Handling Architecture

```mermaid
graph TD
    A[Application Error] --> B{Error Type?}

    B -->|Plan Parse| C[PlanParseError]
    B -->|Config Parse| D[ConfigParseError]
    B -->|Graph Build| E[GraphBuildError]
    B -->|Graph Generation| F[GraphGenerationError]
    B -->|Validation| G[ValidationError]
    B -->|Unsupported Format| H[UnsupportedFormatError]

    C --> I[File Path + Message + Cause]
    D --> J[Path + Message + Cause]
    E --> K[Message + Cause]
    F --> L[Format + Message + Cause]
    G --> M[Field + Message]
    H --> N[Format]

    I --> O[Error Wrapping]
    J --> O
    K --> O
    L --> O
    M --> O
    N --> O

    O --> P[User-Friendly Error Message]

    style A fill:#ffebee
    style P fill:#e8f5e8
```

## Graph Generation Pipeline

```mermaid
sequenceDiagram
    participant CLI as CLI Command
    participant Parser as Plan Parser
    participant Builder as Graph Builder
    participant Factory as Generator Factory
    participant Generator as Graph Generator
    participant Output as Output Handler

    CLI->>Parser: ParsePlanFile(plan.json)
    Parser->>Parser: Read & Validate JSON
    Parser-->>CLI: TerraformPlan Object

    CLI->>Builder: BuildGraph(plan, options)
    Builder->>Builder: Extract Resource Changes
    Builder->>Builder: Extract Output Changes
    Builder->>Builder: Extract Variables & Locals
    Builder->>Builder: Analyze Dependencies
    Builder-->>CLI: GraphData Object

    CLI->>Factory: CreateGenerator(format)
    Factory-->>CLI: GraphGenerator Interface

    CLI->>Generator: Generate(graphData, options)
    Generator->>Generator: Group Nodes by Module
    Generator->>Generator: Apply Styling & Colors
    Generator->>Generator: Generate Format-Specific Output
    Generator-->>CLI: Formatted Graph String

    CLI->>Output: Write to File or Stdout
    Output-->>CLI: Success/Error
```

## Design Patterns

### 1. Dependency Injection

- All major components accept interfaces rather than concrete implementations
- Enables easy testing and component swapping
- Example: `PlanGraphCommand` accepts `PlanParser`, `GraphBuilder`, and `GeneratorFactory` interfaces

### 2. Factory Pattern

- `GeneratorFactory` creates appropriate graph generators based on format
- Encapsulates generator creation logic
- Supports easy addition of new formats

### 3. Strategy Pattern

- Different graph generators implement the same `GraphGenerator` interface
- Allows runtime selection of output format
- Each generator handles format-specific rendering logic

### 4. Builder Pattern

- `GraphBuilder` constructs complex `GraphData` objects
- Handles different node types and dependency analysis
- Provides fluent interface for graph construction

### 5. Error Wrapping

- Custom error types with context preservation
- Proper error chain maintenance
- User-friendly error messages with technical details

## Testing Strategy

```mermaid
graph LR
    subgraph "Unit Tests"
        A[Interface Tests]
        B[Component Tests]
        C[Error Handling Tests]
    end

    subgraph "Integration Tests"
        D[End-to-End Workflows]
        E[Real Terraform Plans]
        F[Format Generation Tests]
    end

    subgraph "Test Coverage"
        G[>80% Code Coverage]
        H[Table-Driven Tests]
        I[Mock Dependencies]
    end

    A --> G
    B --> G
    C --> G
    D --> H
    E --> H
    F --> H
```

## Key Architectural Principles

1. **Separation of Concerns**: Each package has a single responsibility
2. **Interface Segregation**: Small, focused interfaces
3. **Dependency Inversion**: High-level modules don't depend on low-level modules
4. **Single Responsibility**: Each function/class has one clear purpose
5. **Error Handling**: Comprehensive error types with proper context
6. **Testability**: All components are easily unit testable
7. **Extensibility**: Easy to add new graph formats and features

## Package Responsibilities

- **`app/`**: CLI application setup and command registration
- **`commands/`**: Command implementations with dependency injection
- **`config/`**: Application configuration management
- **`core/`**: Domain models, interfaces, and error types
- **`terraform/config/`**: Terraform configuration file parsing
- **`terraform/plan/`**: Terraform plan JSON parsing and validation
- **`terraform/graph/`**: Graph construction and dependency analysis
- **`terraform/graph/generators/`**: Format-specific graph generation

## File Organization

### Core Files

- **`app/app.go`**: Main CLI application with Cobra command setup
- **`app/app_test.go`**: Unit tests for CLI application
- **`commands/plan_graph.go`**: Plan-graph command implementation with dependency injection
- **`config/config.go`**: Application configuration management
- **`core/types.go`**: Core domain types and interfaces
- **`core/errors.go`**: Custom error types with proper wrapping
- **`core/*_test.go`**: Comprehensive unit tests for core functionality

### Terraform Integration

- **`terraform/config/parser.go`**: HCL parser for Terraform configuration files
- **`terraform/config/parser_test.go`**: Tests for configuration parsing
- **`terraform/plan/parser.go`**: JSON parser for Terraform plan files
- **`terraform/plan/parser_test.go`**: Tests for plan parsing and validation

### Graph Generation

- **`terraform/graph/builder.go`**: Graph construction from Terraform plans
- **`terraform/graph/builder_test.go`**: Tests for graph building logic
- **`terraform/graph/dependencies.go`**: Dependency analysis and edge creation
- **`terraform/graph/generators/factory.go`**: Factory for creating graph generators
- **`terraform/graph/generators/factory_test.go`**: Tests for generator factory
- **`terraform/graph/generators/graphviz.go`**: Graphviz DOT format generator
- **`terraform/graph/generators/graphviz_test.go`**: Tests for Graphviz generation
- **`terraform/graph/generators/mermaid.go`**: Mermaid format generator
- **`terraform/graph/generators/plantuml.go`**: PlantUML format generator

This architecture provides a solid foundation for the terraform-ops tool, ensuring maintainability, testability, and extensibility while following Go best practices and clean architecture principles.

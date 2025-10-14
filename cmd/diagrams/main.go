// Package main provides diagram generation utilities for the waffles project.
//
// This application generates architectural and component diagrams for the waffles
// LLM toolchain orchestrator using the go-diagrams library. It creates visual
// representations of the project structure and component relationships to aid in
// documentation and understanding.
//
// The generated diagrams are saved as .dot files in the docs/diagrams/go-diagrams/
// directory and can be converted to various image formats using Graphviz.
//
// Usage:
//
//	go run cmd/diagrams/main.go
//
// This will generate:
//   - architecture.dot: High-level architecture showing user interaction flow
//   - components.dot: Component relationships and dependencies
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/generic"
	"github.com/blushft/go-diagrams/nodes/programming"
)

// main is the entry point for the diagram generation utility.
//
// This function orchestrates the entire diagram generation process:
//  1. Creates the output directory structure
//  2. Changes to the appropriate working directory
//  3. Generates architecture and component diagrams
//  4. Reports successful completion
//
// The function will terminate with log.Fatal if any critical operation fails,
// such as directory creation, navigation, or diagram rendering.
func main() {
	// Ensure output directory exists
	if err := os.MkdirAll("docs/diagrams", 0750); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Change to docs/diagrams directory
	if err := os.Chdir("docs/diagrams"); err != nil {
		log.Fatal("Failed to change directory:", err)
	}

	// Generate architecture diagram
	generateArchitectureDiagram()

	// Generate component diagram
	generateComponentDiagram()

	fmt.Println("Diagram .dot files generated successfully in ./docs/diagrams/go-diagrams/")
}

// generateArchitectureDiagram creates a high-level architecture diagram showing
// the interaction flow between users and the waffles LLM toolchain orchestrator.
//
// The diagram illustrates:
//   - User interaction with the CLI application
//   - Configuration management flow (env/godotenv)
//   - Pipeline execution with external tools (wheresmyprompt, files2prompt, llm)
//   - Logging and query system integration
//
// The diagram is rendered in top-to-bottom (TB) direction and saved as
// "architecture.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateArchitectureDiagram() {
	d, err := diagram.New(diagram.Filename("architecture"), diagram.Label("Waffles LLM Toolchain Architecture"), diagram.Direction("TB"))
	if err != nil {
		log.Fatal(err)
	}

	// Define components
	user := generic.Blank.Blank(diagram.NodeLabel("User"))
	cli := programming.Language.Go(diagram.NodeLabel("Waffles CLI"))
	config := generic.Blank.Blank(diagram.NodeLabel("Configuration\n(env/godotenv)"))
	pipeline := programming.Language.Go(diagram.NodeLabel("Pipeline Executor"))
	wheresmyprompt := generic.Blank.Blank(diagram.NodeLabel("wheresmyprompt\n(Prompt Retrieval)"))
	files2prompt := generic.Blank.Blank(diagram.NodeLabel("files2prompt\n(Context Extraction)"))
	llm := generic.Blank.Blank(diagram.NodeLabel("llm CLI\n(LLM Execution)"))
	logging := generic.Blank.Blank(diagram.NodeLabel("SQLite Logging\n(Query System)"))

	// Create connections
	d.Connect(user, cli, diagram.Forward())
	d.Connect(cli, config, diagram.Forward())
	d.Connect(cli, pipeline, diagram.Forward())
	d.Connect(pipeline, wheresmyprompt, diagram.Forward())
	d.Connect(pipeline, files2prompt, diagram.Forward())
	d.Connect(pipeline, llm, diagram.Forward())
	d.Connect(cli, logging, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}

// generateComponentDiagram creates a detailed component diagram showing the
// relationships and dependencies between different packages in the waffles project.
//
// The diagram illustrates:
//   - main.go as the entry point
//   - cmd/waffles package handling CLI operations and subcommands
//   - Integration with configuration, pipeline, logging, version, and man packages
//   - Internal packages for setup, query, and export functionality
//   - Data flow between components
//
// The diagram is rendered in left-to-right (LR) direction and saved as
// "components.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateComponentDiagram() {
	d, err := diagram.New(diagram.Filename("components"), diagram.Label("Waffles Components"), diagram.Direction("LR"))
	if err != nil {
		log.Fatal(err)
	}

	// Main components
	main := programming.Language.Go(diagram.NodeLabel("main.go"))
	rootCmd := programming.Language.Go(diagram.NodeLabel("cmd/waffles\nroot.go"))
	config := programming.Language.Go(diagram.NodeLabel("pkg/config\nconfig.go"))
	pipeline := programming.Language.Go(diagram.NodeLabel("pkg/pipeline\nexecutor.go"))
	logging := programming.Language.Go(diagram.NodeLabel("pkg/logging\ndatabase.go"))
	version := programming.Language.Go(diagram.NodeLabel("pkg/version\nversion.go"))
	man := programming.Language.Go(diagram.NodeLabel("pkg/man\nman.go"))

	// Internal components
	setup := programming.Language.Go(diagram.NodeLabel("internal/setup\nwizard.go"))
	query := programming.Language.Go(diagram.NodeLabel("internal/query\nengine.go"))
	export := programming.Language.Go(diagram.NodeLabel("internal/export\nexporters.go"))

	// Create connections showing the flow
	d.Connect(main, rootCmd, diagram.Forward())
	d.Connect(rootCmd, config, diagram.Forward())
	d.Connect(rootCmd, pipeline, diagram.Forward())
	d.Connect(rootCmd, logging, diagram.Forward())
	d.Connect(rootCmd, version, diagram.Forward())
	d.Connect(rootCmd, man, diagram.Forward())
	d.Connect(rootCmd, setup, diagram.Forward())
	d.Connect(rootCmd, query, diagram.Forward())
	d.Connect(rootCmd, export, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}

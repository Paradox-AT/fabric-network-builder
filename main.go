package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"network-builder/src/cli"
	"network-builder/src/config"
	"network-builder/src/generator"
	"network-builder/src/generator/configtx"
	"network-builder/src/generator/crypto"
	"network-builder/src/generator/docker"
	"network-builder/src/generator/scripts"

	"github.com/charmbracelet/huh"
)

func main() {
	fmt.Println("=== Hyperledger Fabric Network Builder ===")
	fmt.Println("This wizard will help you configure your enterprise blockchain network.")
	fmt.Println("Use Tab/Shift+Tab to navigate back and forth between pages.")

	outputDir := "./network"
	configPath := filepath.Join(outputDir, "network-config.json")

	// Load existing configuration if available
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading configuration: %v", err)
	}

	var cfg *config.NetworkConfig
	skipWizard := false

	if loadedCfg != nil {
		var action string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("A previous configuration was found. What would you like to do?").
					Options(
						huh.NewOption("Start fresh (overwrite existing config)", "fresh"),
						huh.NewOption("Make changes to existing config", "modify"),
						huh.NewOption("Reset (Regenerate artifacts from saved config)", "reset"),
					).
					Value(&action),
			),
		).Run()

		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("\nAborted.")
				os.Exit(0)
			}
			log.Fatalf("Error in prompt: %v", err)
		}

		switch action {
		case "fresh":
			cfg = nil
		case "modify":
			cfg = loadedCfg
		case "reset":
			cfg = loadedCfg
			skipWizard = true
		}
	}

	if !skipWizard {
		// Run the wizard to collect/update configuration
		cfg, err = cli.RunWizard(cfg)
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("\nAborted.")
				os.Exit(0)
			}
			log.Fatalf("Wizard failed: %v", err)
		}
	}

	cli.PrintSummary(cfg)

	generators := []generator.Generator{
		crypto.NewCryptogenGenerator(),
		crypto.NewFabricCAGenerator(),
		configtx.NewConfigtxGenerator(),
		docker.NewDockerComposeGenerator(),
		scripts.NewScriptsGenerator(),
	}

	fmt.Println("\n=== Generating Network Artifacts ===")

	cleanArtifacts(outputDir)

	for _, gen := range generators {
		err := gen.Generate(cfg, outputDir)
		if err != nil {
			log.Fatalf("Failed during generation: %v", err)
		}
	}

	// Save the configuration state for future runs
	if err := config.SaveConfig(cfg, configPath); err != nil {
		log.Printf("Warning: Failed to save configuration: %v", err)
	}

	fmt.Printf("\nGeneration complete. Artifacts are located in: %s\n", outputDir)
}

// cleanArtifacts removes all previously generated files to prevent stale
// artifacts from interfering with the new generation run.
func cleanArtifacts(outputDir string) {
	fmt.Println("Cleaning up old artifacts...")

	// Generated subdirectories
	subDirs := []string{"compose", "channel-artifacts", "configtx", "organizations", "scripts"}
	for _, d := range subDirs {
		target := filepath.Join(outputDir, d)
		if err := os.RemoveAll(target); err == nil {
			fmt.Printf("  removed: %s/\n", d)
		}
	}

	// Generated top-level files
	files := []string{"network.sh", "network.config", "log.txt"}
	for _, f := range files {
		target := filepath.Join(outputDir, f)
		if _, err := os.Stat(target); err == nil {
			os.Remove(target)
			fmt.Printf("  removed: %s\n", f)
		}
	}

	fmt.Println("Cleanup complete.")
}

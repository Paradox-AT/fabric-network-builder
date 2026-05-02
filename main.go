package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"network-builder/src/cli"
	"network-builder/src/config"
	"network-builder/src/generator"
	"network-builder/src/generator/configtx"
	"network-builder/src/generator/crypto"
	"network-builder/src/generator/docker"
	"network-builder/src/generator/scripts"
)

func main() {
	fmt.Println("=== Hyperledger Fabric Network Builder ===")
	fmt.Println("This wizard will help you configure your enterprise blockchain network.")
	fmt.Println("Use Tab/Shift+Tab to navigate back and forth between pages.")

	configPath := "network-config.json"
	outputDir := "./network"

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
		configtx.NewConfigtxGenerator(),
		docker.NewDockerComposeGenerator(),
		scripts.NewScriptsGenerator(),
	}

	fmt.Println("\n=== Generating Network Artifacts ===")

	// Clean up old artifacts to prevent stale files from previous runs
	subDirs := []string{"organizations", "configtx", "scripts", "compose"}
	for _, d := range subDirs {
		os.RemoveAll(filepath.Join(outputDir, d))
	}
	// Also remove top-level scripts if they exist (legacy paths)
	os.Remove(filepath.Join(outputDir, "network.sh"))
	os.Remove(filepath.Join(outputDir, "bootstrap.sh"))

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

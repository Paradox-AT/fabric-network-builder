package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"

	"network-builder/src/cli"
	"network-builder/src/config"
	"network-builder/src/generator"
	"network-builder/src/generator/configtx"
	"network-builder/src/generator/crypto"
	"network-builder/src/generator/docker"
)

func main() {
	outputDir := "./network"
	configPath := filepath.Join(outputDir, "network-config.json")

	var existingCfg *config.NetworkConfig

	// Check if existing config exists
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Warning: Failed to load existing config: %v", err)
	}

	if loadedCfg != nil {
		var action string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("A previous configuration was found. What would you like to do?").
					Options(
						huh.NewOption("Make changes to existing config", "modify"),
						huh.NewOption("Start fresh (overwrite existing config)", "fresh"),
						huh.NewOption("Back it up and start fresh", "backup"),
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
		case "modify":
			existingCfg = loadedCfg
		case "backup":
			backupDir := fmt.Sprintf("./network.backup_%d", time.Now().Unix())
			err := os.Rename(outputDir, backupDir)
			if err != nil {
				log.Fatalf("Failed to backup directory: %v", err)
			}
			fmt.Printf("Backed up existing configuration to %s\n", backupDir)
			// Proceed with existingCfg = nil to start fresh
		case "fresh":
			// Proceed with existingCfg = nil to start fresh
		}
	}

	cfg, err := cli.RunWizard(existingCfg)
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("\nWizard aborted by user. Exiting gracefully.")
			os.Exit(0)
		}
		log.Printf("Wizard failed: %v", err)
		os.Exit(1)
	}

	cli.PrintSummary(cfg)

	generators := []generator.Generator{
		crypto.NewCryptogenGenerator(),
		configtx.NewConfigtxGenerator(),
		docker.NewDockerComposeGenerator(),
	}

	fmt.Println("\n=== Generating Network Artifacts ===")
	for _, gen := range generators {
		err := gen.Generate(cfg, outputDir)
		if err != nil {
			log.Fatalf("Failed during generation: %v", err)
		}
	}

	// Save the configuration state for future runs
	if err := config.SaveConfig(cfg, configPath); err != nil {
		log.Printf("Warning: Failed to save configuration state: %v", err)
	}

	fmt.Println("\nGeneration complete. Artifacts are located in:", outputDir)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"network-builder/src/cli"
	"network-builder/src/config"
	"network-builder/src/generator"
	"network-builder/src/generator/ccp"
	"network-builder/src/generator/configtx"
	"network-builder/src/generator/crypto"
	"network-builder/src/generator/docker"
	"network-builder/src/generator/monitoring"
	"network-builder/src/generator/scripts"

	"github.com/charmbracelet/huh"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Parse CLI flags for verbose and skip-cleanup modes
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--verbose" {
			config.Verbose = true
		}
		if arg == "-s" || arg == "--skip-cleanup" || arg == "--no-cleanup" {
			config.SkipCleanup = true
		}
	}

	// Set up root context with signal cancellation for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("=== Hyperledger Fabric Network Builder ===")
	fmt.Println("This wizard will help you configure your enterprise blockchain network.")
	fmt.Println("Use Tab/Shift+Tab to navigate back and forth between pages.")

	outputDir := "./network"
	configPath := filepath.Join(outputDir, "network-config.json")

	// Load existing configuration if available
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Notice: Could not parse previous config cleanly (%v). Starting fresh.", err)
		loadedCfg = nil
	}

	var cfg *config.NetworkConfig
	skipWizard := false
	addOrgFlow := false

	if loadedCfg != nil {
		var action string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("A previous configuration was found. What would you like to do?").
					Options(
						huh.NewOption("Start fresh (overwrite existing config)", "fresh"),
						huh.NewOption("Make changes to existing config", "modify"),
						huh.NewOption("Add a new organization to existing network", "add-org"),
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
		case "add-org":
			cfg = loadedCfg
			addOrgFlow = true
		case "reset":
			cfg = loadedCfg
			skipWizard = true
		}
	}

	if !skipWizard {
		// Run the wizard to collect/update configuration
		if addOrgFlow {
			cfg, err = cli.RunAddOrgWizard(cfg)
		} else {
			cfg, err = cli.RunWizard(cfg)
		}
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("\nAborted.")
				os.Exit(0)
			}
			log.Fatalf("Wizard failed: %v", err)
		}
	}

	// Validate configuration before generation
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation error: %v", err)
	}

	cli.PrintSummary(cfg)

	generators := []generator.Generator{
		crypto.NewCryptogenGenerator(),
		crypto.NewFabricCAGenerator(),
		configtx.NewConfigtxGenerator(),
		docker.NewDockerComposeGenerator(),
		scripts.NewScriptsGenerator(),
		ccp.NewCCPGenerator(),
		monitoring.NewMonitoringGenerator(),
	}

	fmt.Println("\n=== Generating Network Artifacts ===")

	if config.Verbose {
		if !addOrgFlow && !config.SkipCleanup {
			cleanArtifacts(outputDir)
		}
		g, gCtx := errgroup.WithContext(ctx)
		for _, gen := range generators {
			genInstance := gen
			g.Go(func() error {
				return genInstance.Generate(gCtx, cfg, outputDir)
			})
		}
		if err := g.Wait(); err != nil {
			log.Fatalf("Failed during artifact generation: %v", err)
		}
	} else {
		if !addOrgFlow && !config.SkipCleanup {
			cleanArtifacts(outputDir)
		}
		fmt.Println()
		err = runWithSpinner("Generating network artifacts", func() error {
			g, gCtx := errgroup.WithContext(ctx)
			for _, gen := range generators {
				genInstance := gen
				g.Go(func() error {
					return genInstance.Generate(gCtx, cfg, outputDir)
				})
			}
			return g.Wait()
		})
		if err != nil {
			log.Fatalf("Failed during artifact generation: %v", err)
		}
	}

	// Save the configuration state for future runs with atomic writes
	if err := config.SaveConfig(cfg, configPath); err != nil {
		log.Printf("Warning: Failed to save configuration: %v", err)
	}

	fmt.Printf("\nGeneration complete. Artifacts are located in: %s\n", outputDir)

	fmt.Println()
	fmt.Println("⚠️  SECURITY REMINDER: The generated .env file contains WEAK DEFAULT credentials.  ⚠️")
	fmt.Println("⚠️  Rotate ALL passwords before deploying to any shared or production environment. ⚠️")
	fmt.Println("⚠️  See the warning header inside network/.env for details.                        ⚠️")
	fmt.Println()
}

// cleanArtifacts removes old generated config files while preserving runtime directories and downloaded Fabric configs (core.yaml, orderer.yaml)
func cleanArtifacts(outputDir string) {
	subDirs := []string{
		"compose",
		"channel-artifacts",
		"configtx",
		"scripts",
		filepath.Join("config", "prometheus"),
		filepath.Join("config", "grafana"),
	}
	files := []string{"network.sh", "network.config", "log.txt"}

	if config.Verbose {
		fmt.Println("Cleaning up old artifacts")
		for _, d := range subDirs {
			target := filepath.Join(outputDir, d)
			if err := os.RemoveAll(target); err == nil {
				fmt.Printf("removed: %s/\n", d)
			}
		}
		for _, f := range files {
			target := filepath.Join(outputDir, f)
			if _, err := os.Stat(target); err == nil {
				_ = os.Remove(target)
				fmt.Printf("removed: %s\n", f)
			}
		}
		fmt.Println("Cleanup complete.")
	} else {
		_ = runWithSpinner("Cleaning up old artifacts", func() error {
			for _, d := range subDirs {
				target := filepath.Join(outputDir, d)
				_ = os.RemoveAll(target)
			}
			for _, f := range files {
				target := filepath.Join(outputDir, f)
				if _, err := os.Stat(target); err == nil {
					_ = os.Remove(target)
				}
			}
			return nil
		})
	}
}

// runWithSpinner runs fn asynchronously while printing an animated braille spinner to stdout.
func runWithSpinner(title string, fn func() error) error {
	done := make(chan struct{})
	var err error
	go func() {
		err = fn()
		close(done)
	}()

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			if err != nil {
				fmt.Printf("\r  %s ❌", title)
			} else {
				fmt.Printf("\r  %s ✔ ", title)
			}
			return err
		case <-time.After(80 * time.Millisecond):
			fmt.Printf("\r%s  %s", title, frames[i%len(frames)])
			i++
		}
	}
}

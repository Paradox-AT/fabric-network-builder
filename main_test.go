package main

import (
	"context"
	"path/filepath"
	"testing"

	"network-builder/src/config"
	"network-builder/src/generator"
	"network-builder/src/generator/ccp"
	"network-builder/src/generator/configtx"
	"network-builder/src/generator/crypto"
	"network-builder/src/generator/docker"
	"network-builder/src/generator/monitoring"
	"network-builder/src/generator/scripts"

	"golang.org/x/sync/errgroup"
)

func TestRegenerateNetworkArtifacts(t *testing.T) {
	outputDir := "./network"
	configPath := filepath.Join(outputDir, "network-config.json")

	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Skipf("No network-config.json found: %v", err)
	}

	generators := []generator.Generator{
		crypto.NewFabricCAGenerator(),
		crypto.NewCryptogenGenerator(),
		configtx.NewConfigtxGenerator(),
		docker.NewDockerComposeGenerator(),
		scripts.NewScriptsGenerator(),
		monitoring.NewMonitoringGenerator(),
		ccp.NewCCPGenerator(),
	}

	g, ctx := errgroup.WithContext(context.Background())
	for _, gen := range generators {
		gen := gen
		g.Go(func() error {
			return gen.Generate(ctx, loadedCfg, outputDir)
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatalf("Failed to generate artifacts: %v", err)
	}
}

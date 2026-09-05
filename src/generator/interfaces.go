package generator

import (
	"context"
	"network-builder/src/config"
)

// Generator defines the interface for creating physical network artifacts from in-memory configuration.
type Generator interface {
	Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error
}


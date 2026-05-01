package generator

import "network-builder/src/config"

// Generator defines the interface for creating physical network artifacts from in-memory configuration.
type Generator interface {
	Generate(cfg *config.NetworkConfig, outputDir string) error
}

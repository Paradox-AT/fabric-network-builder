package configtx

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"

	"network-builder/src/config"
	"network-builder/src/generator/utils"
)

//go:embed templates/configtx.yaml.tmpl
var templateFS embed.FS

// ConfigtxGenerator generates a configtx.yaml file
type ConfigtxGenerator struct{}

// NewConfigtxGenerator returns a new instance of ConfigtxGenerator
func NewConfigtxGenerator() *ConfigtxGenerator {
	return &ConfigtxGenerator{}
}

// Generate implements the Generator interface
func (g *ConfigtxGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// nextID is generator-local state; merge it into the shared FuncMap cleanly.
	currentID := 0
	funcs := utils.GetFuncMap()
	funcs["nextID"] = func() int {
		currentID++
		return currentID
	}

	tmpl, err := template.New("configtx.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/configtx.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse configtx template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute configtx template: %w", err)
	}

	filePath := filepath.Join(outputDir, "configtx", "configtx.yaml")
	if err := config.WriteFileAtomic(filePath, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write configtx.yaml: %w", err)
	}

	return nil
}


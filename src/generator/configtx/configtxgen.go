package configtx

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"network-builder/src/config"
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
func (g *ConfigtxGenerator) Generate(cfg *config.NetworkConfig, outputDir string) error {
	currentID := 0
	funcs := template.FuncMap{
		"until": func(count int) []int {
			var r []int
			for i := 0; i < count; i++ {
				r = append(r, i)
			}
			return r
		},
		"toLower": strings.ToLower,
		"nextID": func() int {
			currentID++
			return currentID
		},
	}

	tmpl, err := template.New("configtx.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/configtx.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse configtx template: %w", err)
	}

	configtxDir := filepath.Join(outputDir, "configtx")
	err = os.MkdirAll(configtxDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create configtx output directory: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute configtx template: %w", err)
	}

	filePath := filepath.Join(configtxDir, "configtx.yaml")
	err = os.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write configtx.yaml: %w", err)
	}

	fmt.Printf("Generated %s\n", filePath)
	return nil
}

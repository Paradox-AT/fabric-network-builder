package crypto

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

//go:embed templates/crypto-config.yaml.tmpl
var templateFS embed.FS

// CryptogenGenerator generates a crypto-config.yaml for the cryptogen binary.
type CryptogenGenerator struct{}

// NewCryptogenGenerator returns a new instance of CryptogenGenerator.
func NewCryptogenGenerator() *CryptogenGenerator {
	return &CryptogenGenerator{}
}

// Generate implements the Generator interface.
func (g *CryptogenGenerator) Generate(cfg *config.NetworkConfig, outputDir string) error {
	if cfg.CryptoStrategy != "cryptogen" {
		// Silently return if this strategy isn't chosen
		return nil
	}

	// Parse the template
	funcs := template.FuncMap{
		"until": func(count int) []int {
			var r []int
			for i := 0; i < count; i++ {
				r = append(r, i)
			}
			return r
		},
	}

	tmpl, err := template.New("crypto-config.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/crypto-config.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse crypto-config template: %w", err)
	}

	// Ensure output directory exists
	err = os.MkdirAll(filepath.Join(outputDir, "organizations"), 0755)
	if err != nil {
		return fmt.Errorf("failed to create organizations directory: %w", err)
	}

	for _, org := range cfg.Orgs {
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, org)
		if err != nil {
			return fmt.Errorf("failed to execute crypto-config template for org %s: %w", org.Name, err)
		}

		orgDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "identity-config", "cryptogen")
		err = os.MkdirAll(orgDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for org %s: %w", org.Name, err)
		}

		filePath := filepath.Join(orgDir, "crypto-config.yaml")
		err = os.WriteFile(filePath, buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("failed to write crypto-config yaml for org %s: %w", org.Name, err)
		}
		fmt.Printf("Generated %s\n", filePath)
	}

	return nil
}

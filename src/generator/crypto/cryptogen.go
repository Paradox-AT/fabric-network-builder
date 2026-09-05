package crypto

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"network-builder/src/config"
	"network-builder/src/generator/utils"
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
func (g *CryptogenGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	if cfg.CryptoStrategy != "cryptogen" {
		// Silently return if this strategy isn't chosen
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tmpl, err := template.New("crypto-config.yaml.tmpl").Funcs(utils.GetFuncMap()).ParseFS(templateFS, "templates/crypto-config.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse crypto-config template: %w", err)
	}

	// Ensure output directory exists with restricted directory permissions
	err = os.MkdirAll(filepath.Join(outputDir, "organizations"), 0700)
	if err != nil {
		return fmt.Errorf("failed to create organizations directory: %w", err)
	}

	for _, org := range cfg.Orgs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, org)
		if err != nil {
			return fmt.Errorf("failed to execute crypto-config template for org %s: %w", org.Name, err)
		}

		orgDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "identity-config", "cryptogen")
		filePath := filepath.Join(orgDir, "crypto-config.yaml")

		if err := config.WriteFileAtomic(filePath, buf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write crypto-config yaml for org %s: %w", org.Name, err)
		}
	}

	return nil
}


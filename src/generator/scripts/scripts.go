package scripts

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"network-builder/src/config"
	"network-builder/src/generator/utils"
)

// sensitiveFiles lists generated output files that must have restrictive permissions.
var sensitiveFiles = map[string]bool{
	".env": true,
}

//go:embed templates
var templateFS embed.FS

// ScriptsGenerator generates automation scripts for the network
type ScriptsGenerator struct{}

// NewScriptsGenerator returns a new instance of ScriptsGenerator
func NewScriptsGenerator() *ScriptsGenerator {
	return &ScriptsGenerator{}
}

// Generate implements the Generator interface
func (g *ScriptsGenerator) Generate(cfg *config.NetworkConfig, outputDir string) error {
	// 1. Generate scripts that go into network/scripts/
	scriptFiles := map[string]string{
		"envVar.sh.tmpl":        "scripts/envVar.sh",
		"utils.sh.tmpl":         "scripts/utils.sh",
		"createChannel.sh.tmpl": "scripts/createChannel.sh",
		"bootstrap.sh.tmpl":     "scripts/bootstrap.sh",
		"deployCC.sh.tmpl":      "scripts/deployCC.sh",
		"ccutils.sh.tmpl":       "scripts/ccutils.sh",
		"packageCC.sh.tmpl":     "scripts/packageCC.sh",
	}

	for tmplName, fileName := range scriptFiles {
		err := g.generateFile(cfg, tmplName, filepath.Join(outputDir, fileName))
		if err != nil {
			return err
		}
	}

	// 2. Generate scripts that go into network/ root
	rootFiles := map[string]string{
		"network.sh.tmpl":     "network.sh",
		"network.config.tmpl": "network.config",
		"env.tmpl":            ".env",
	}

	for tmplName, fileName := range rootFiles {
		destPath := filepath.Join(outputDir, fileName)
		err := g.generateFile(cfg, tmplName, destPath)
		if err != nil {
			return err
		}
		// Apply restrictive permissions to sensitive files (e.g., .env)
		if sensitiveFiles[fileName] {
			if chmodErr := os.Chmod(destPath, 0600); chmodErr != nil {
				fmt.Printf("Warning: failed to set permissions on %s: %v\n", destPath, chmodErr)
			}
		}
	}

	fmt.Println()
	fmt.Println("⚠️  SECURITY REMINDER: The generated .env file contains WEAK DEFAULT credentials.")
	fmt.Println("   Rotate ALL passwords before deploying to any shared or production environment.")
	fmt.Println("   See the warning header inside network/.env for details.")
	fmt.Println()

	return nil
}

func (g *ScriptsGenerator) generateFile(cfg *config.NetworkConfig, tmplName, destPath string) error {
	tmpl, err := template.New(tmplName).Funcs(utils.GetFuncMap()).ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplName, err)
	}

	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute template %s: %w", tmplName, err)
	}

	err = os.WriteFile(destPath, buf.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	fmt.Printf("Generated %s\n", destPath)
	return nil
}

package scripts

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		"addOrg.sh.tmpl":        "scripts/addOrg.sh",
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
	}

	for tmplName, fileName := range rootFiles {
		destPath := filepath.Join(outputDir, fileName)
		err := g.generateFile(cfg, tmplName, destPath)
		if err != nil {
			return err
		}
	}

	// 3. Generate `.env` with merge logic to append missing variables
	envPath := filepath.Join(outputDir, ".env")
	envContent, err := g.executeTemplate(cfg, "env.tmpl")
	if err != nil {
		return err
	}
	mergedEnv, err := g.mergeEnvFile(envContent, envPath)
	if err != nil {
		return fmt.Errorf("failed to merge env file: %w", err)
	}
	err = os.WriteFile(envPath, mergedEnv, 0600)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}
	fmt.Printf("Generated %s\n", envPath)

	fmt.Println()
	fmt.Println("⚠️  SECURITY REMINDER: The generated .env file contains WEAK DEFAULT credentials.")
	fmt.Println("   Rotate ALL passwords before deploying to any shared or production environment.")
	fmt.Println("   See the warning header inside network/.env for details.")
	fmt.Println()

	return nil
}

func (g *ScriptsGenerator) executeTemplate(cfg *config.NetworkConfig, tmplName string) ([]byte, error) {
	tmpl, err := template.New(tmplName).Funcs(utils.GetFuncMap()).ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", tmplName, err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", tmplName, err)
	}
	return buf.Bytes(), nil
}

func (g *ScriptsGenerator) generateFile(cfg *config.NetworkConfig, tmplName, destPath string) error {
	content, err := g.executeTemplate(cfg, tmplName)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	err = os.WriteFile(destPath, content, 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	fmt.Printf("Generated %s\n", destPath)
	return nil
}

func (g *ScriptsGenerator) mergeEnvFile(newContent []byte, destPath string) ([]byte, error) {
	existingBytes, err := os.ReadFile(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newContent, nil
		}
		return nil, err
	}

	// Parse existing env file into key-value pairs
	existingLines := strings.Split(string(existingBytes), "\n")
	existingVars := make(map[string]string)
	for _, line := range existingLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			existingVars[parts[0]] = parts[1]
		}
	}

	// Read new generated lines, append only missing ones
	newLines := strings.Split(string(newContent), "\n")
	var appendedVars []string
	for _, line := range newLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if _, exists := existingVars[key]; !exists {
				appendedVars = append(appendedVars, line)
			}
		}
	}

	if len(appendedVars) > 0 {
		var merged strings.Builder
		merged.WriteString(strings.TrimRight(string(existingBytes), "\n"))
		merged.WriteString("\n\n# Dynamically Appended Credentials for New Organization(s)\n")
		for _, v := range appendedVars {
			merged.WriteString(v)
			merged.WriteString("\n")
		}
		return []byte(merged.String()), nil
	}

	return existingBytes, nil
}

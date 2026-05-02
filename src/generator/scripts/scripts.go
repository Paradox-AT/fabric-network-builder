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
)

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
	}

	for tmplName, fileName := range scriptFiles {
		err := g.generateFile(cfg, tmplName, filepath.Join(outputDir, fileName))
		if err != nil {
			return err
		}
	}

	// 2. Generate scripts that go into network/ root
	rootFiles := map[string]string{
		"network.sh.tmpl": "network.sh",
	}

	for tmplName, fileName := range rootFiles {
		err := g.generateFile(cfg, tmplName, filepath.Join(outputDir, fileName))
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *ScriptsGenerator) generateFile(cfg *config.NetworkConfig, tmplName, destPath string) error {
	funcs := template.FuncMap{
		"until": func(count int) []int {
			var r []int
			for i := 0; i < count; i++ {
				r = append(r, i)
			}
			return r
		},
		"toLower": strings.ToLower,
		"inc": func(i int) int {
			return i + 1
		},
		"add": func(a, b int) int {
			return a + b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
	}

	tmpl, err := template.New(tmplName).Funcs(funcs).ParseFS(templateFS, "templates/"+tmplName)
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

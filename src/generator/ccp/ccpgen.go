package ccp

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

//go:embed templates/*.tmpl
var templateFS embed.FS

type CCPGenerator struct{}

func NewCCPGenerator() *CCPGenerator {
	return &CCPGenerator{}
}

// CCPData is the template data for CCP generation.
type CCPData struct {
	Org         config.OrgConfig
	OrgIndex    int
	NetworkName string
}

func (g *CCPGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	if !cfg.GenerateCCP {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	funcs := utils.GetFuncMap()

	// Parse all three templates
	genTmpl, err := template.New("ccp-generate.sh.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/ccp-generate.sh.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse ccp-generate.sh template: %w", err)
	}
	jsonTmpl, err := template.New("ccp-template.json.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/ccp-template.json.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse ccp-template.json template: %w", err)
	}
	yamlTmpl, err := template.New("ccp-template.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/ccp-template.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse ccp-template.yaml template: %w", err)
	}

	data := struct {
		NetworkName string
		Orgs        []config.OrgConfig
	}{
		NetworkName: cfg.NetworkName,
		Orgs:        cfg.Orgs,
	}

	orgsDir := filepath.Join(outputDir, "organizations")

	// 1. Generate single central ccp-generate.sh
	var genBuf bytes.Buffer
	if err := genTmpl.Execute(&genBuf, data); err != nil {
		return fmt.Errorf("failed to execute ccp-generate.sh: %w", err)
	}
	genPath := filepath.Join(orgsDir, "ccp-generate.sh")
	if err := config.WriteFileAtomic(genPath, genBuf.Bytes(), 0700); err != nil {
		return fmt.Errorf("failed to write %s: %w", genPath, err)
	}

	// 2. Generate single central ccp-template.json
	var jsonBuf bytes.Buffer
	if err := jsonTmpl.Execute(&jsonBuf, data); err != nil {
		return fmt.Errorf("failed to execute ccp-template.json: %w", err)
	}
	jsonPath := filepath.Join(orgsDir, "ccp-template.json")
	if err := config.WriteFileAtomic(jsonPath, jsonBuf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", jsonPath, err)
	}

	// 3. Generate single central ccp-template.yaml
	var yamlBuf bytes.Buffer
	if err := yamlTmpl.Execute(&yamlBuf, data); err != nil {
		return fmt.Errorf("failed to execute ccp-template.yaml: %w", err)
	}
	yamlPath := filepath.Join(orgsDir, "ccp-template.yaml")
	if err := config.WriteFileAtomic(yamlPath, yamlBuf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", yamlPath, err)
	}

	return nil
}

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
	"network-builder/src/generator/utils"
)

//go:embed templates/registerEnroll.sh.tmpl templates/fabric-ca-server-config.yaml.tmpl
var caTmplFS embed.FS

// FabricCAGenerator generates Fabric CA configuration and docker-compose files
type FabricCAGenerator struct{}

// NewFabricCAGenerator returns a new instance of FabricCAGenerator
func NewFabricCAGenerator() *FabricCAGenerator {
	return &FabricCAGenerator{}
}

// OrgCAData is passed to the per-org CA compose template
type OrgCAData struct {
	Org            config.OrgConfig
	OrgIndex       int
	CAVersion      string
	CADatabaseType string
	PostgresVersion string
}

// Generate implements the Generator interface
func (g *FabricCAGenerator) Generate(cfg *config.NetworkConfig, outputDir string) error {
	if cfg.CryptoStrategy != "Fabric CA" {
		return nil
	}


	// 1. Parse templates
	enrollTmpl, err := template.New("registerEnroll.sh.tmpl").Funcs(utils.GetFuncMap()).ParseFS(caTmplFS, "templates/registerEnroll.sh.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse registerEnroll template: %w", err)
	}

	serverConfigTmpl, err := template.New("fabric-ca-server-config.yaml.tmpl").Funcs(utils.GetFuncMap()).ParseFS(caTmplFS, "templates/fabric-ca-server-config.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse CA server config template: %w", err)
	}

	// 2. Generate per-org identity scripts and configs
	for i, org := range cfg.Orgs {
		orgScriptDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "scripts")
		if err := os.MkdirAll(orgScriptDir, 0755); err != nil {
			return fmt.Errorf("failed to create scripts dir for org %s: %w", org.Name, err)
		}

		data := OrgCAData{
			Org:             org,
			OrgIndex:        i,
			CAVersion:       cfg.CAVersion,
			CADatabaseType:  cfg.CADatabaseType,
			PostgresVersion: cfg.PostgresVersion,
		}

		// 2a. Generate registerEnroll.sh
		var enrollBuf bytes.Buffer
		if err := enrollTmpl.Execute(&enrollBuf, data); err != nil {
			return fmt.Errorf("failed to execute registerEnroll template for %s: %w", org.Name, err)
		}

		scriptPath := filepath.Join(orgScriptDir, "registerEnroll.sh")
		if err := os.WriteFile(scriptPath, enrollBuf.Bytes(), 0755); err != nil {
			return fmt.Errorf("failed to write %s: %w", scriptPath, err)
		}
		fmt.Printf("Generated %s\n", scriptPath)

		// 2b. Create the CA data directory
		caDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "ca")
		if err := os.MkdirAll(caDir, 0755); err != nil {
			return fmt.Errorf("failed to create CA directory for org %s: %w", org.Name, err)
		}

		// 2c. Generate fabric-ca-server-config.yaml
		caConfigDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "identity-config", "ca")
		if err := os.MkdirAll(caConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create CA config directory for org %s: %w", org.Name, err)
		}

		var configBuf bytes.Buffer
		if err := serverConfigTmpl.Execute(&configBuf, data); err != nil {
			return fmt.Errorf("failed to execute CA server config template for %s: %w", org.Name, err)
		}

		configPath := filepath.Join(caConfigDir, "fabric-ca-server-config.yaml")
		if err := os.WriteFile(configPath, configBuf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", configPath, err)
		}
		fmt.Printf("Generated %s\n", configPath)
	}

	return nil
}

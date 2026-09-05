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
	Org             config.OrgConfig
	OrgIndex        int
	CAVersion       string
	CADatabaseType  string
	PostgresVersion string
}

// Generate implements the Generator interface
func (g *FabricCAGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	if cfg.CryptoStrategy != "Fabric CA" {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data := OrgCAData{
			Org:             org,
			OrgIndex:        i,
			CAVersion:       cfg.CAVersion,
			CADatabaseType:  cfg.CADatabaseType,
			PostgresVersion: cfg.PostgresVersion,
		}

		// 2a. Generate registerEnroll.sh with executable permissions (0700)
		var enrollBuf bytes.Buffer
		if err := enrollTmpl.Execute(&enrollBuf, data); err != nil {
			return fmt.Errorf("failed to execute registerEnroll template for %s: %w", org.Name, err)
		}

		scriptPath := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "scripts", "registerEnroll.sh")
		if err := config.WriteFileAtomic(scriptPath, enrollBuf.Bytes(), 0700); err != nil {
			return fmt.Errorf("failed to write %s: %w", scriptPath, err)
		}

		// 2b. Create the CA data directory with restricted permissions (0700)
		caDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "ca")
		if err := os.MkdirAll(caDir, 0700); err != nil {
			return fmt.Errorf("failed to create CA directory for org %s: %w", org.Name, err)
		}

		// 2c. Generate fabric-ca-server-config.yaml with secure permissions (0600)
		var configBuf bytes.Buffer
		if err := serverConfigTmpl.Execute(&configBuf, data); err != nil {
			return fmt.Errorf("failed to execute CA server config template for %s: %w", org.Name, err)
		}

		configPath := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "identity-config", "ca", "fabric-ca-server-config.yaml")
		if err := config.WriteFileAtomic(configPath, configBuf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", configPath, err)
		}
	}

	return nil
}


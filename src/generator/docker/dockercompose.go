package docker

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

//go:embed templates/*.tmpl
var templateFS embed.FS

// DockerComposeGenerator generates organizational and hub docker-compose YAML files
type DockerComposeGenerator struct{}

// NewDockerComposeGenerator returns a new instance of DockerComposeGenerator
func NewDockerComposeGenerator() *DockerComposeGenerator {
	return &DockerComposeGenerator{}
}

type OrgComposeData struct {
	Org               config.OrgConfig
	OrgIndex          int
	FabricVersion     string
	CAVersion         string
	CADatabaseType    string
	PostgresVersion   string
	CouchDBVersion    string
	BindAddress       string
	PeerStateDatabase func(int) string // returns "leveldb" or "couchdb" for a given peer index
}

// Generate implements the Generator interface
func (g *DockerComposeGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	funcs := utils.GetFuncMap()

	// 1. Parse all templates
	peerTmpl, err := template.New("peer-compose.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/peer-compose.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse peer template: %w", err)
	}
	ordererTmpl, err := template.New("orderer-compose.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/orderer-compose.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse orderer template: %w", err)
	}
	hubPeersTmpl, err := template.New("hub-peers.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/hub-peers.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse hub-peers template: %w", err)
	}
	hubOrderersTmpl, err := template.New("hub-orderers.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/hub-orderers.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse hub-orderers template: %w", err)
	}
	hubMasterTmpl, err := template.New("hub-master.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/hub-master.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse hub-master template: %w", err)
	}
	couchTmpl, err := template.New("couch-compose.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/couch-compose.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse couch template: %w", err)
	}
	caComposeTmpl, err := template.New("ca-compose.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/ca-compose.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse ca-compose template: %w", err)
	}
	hubCAsTmpl, err := template.New("hub-cas.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/hub-cas.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse hub-cas template: %w", err)
	}
	utilsTmpl, err := template.New("compose-utils.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/compose-utils.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse compose-utils template: %w", err)
	}
	serversTmpl, err := template.New("pgadmin-servers.json.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/pgadmin-servers.json.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse servers template: %w", err)
	}

	// 2. Generate Organizational Compose Files
	for i, org := range cfg.Orgs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		orgComposeDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "compose")
		orgCopy := org // capture for closure
		bindAddress := cfg.BindAddress
		if bindAddress == "" {
			bindAddress = "0.0.0.0"
		}

		data := OrgComposeData{
			Org:             org,
			OrgIndex:        i,
			FabricVersion:   cfg.FabricVersion,
			CAVersion:       cfg.CAVersion,
			CADatabaseType:  cfg.CADatabaseType,
			PostgresVersion: cfg.PostgresVersion,
			CouchDBVersion:  cfg.CouchDBVersion,
			BindAddress:     bindAddress,
			PeerStateDatabase: func(peerIndex int) string {
				return orgCopy.PeerStateDatabase(peerIndex)
			},
		}

		if org.PeerCount > 0 {
			var buf bytes.Buffer
			err = peerTmpl.Execute(&buf, data)
			if err != nil {
				return fmt.Errorf("failed to execute peer template for %s: %w", org.Name, err)
			}
			filePath := filepath.Join(orgComposeDir, "compose-peer.yaml")
			if err := config.WriteFileAtomic(filePath, buf.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}

			// Generate couch overlay if any peer in this org uses CouchDB
			hasCouchDB := false
			for p := 0; p < org.PeerCount; p++ {
				if org.PeerStateDatabase(p) == "couchdb" {
					hasCouchDB = true
					break
				}
			}
			if hasCouchDB {
				var couchBuf bytes.Buffer
				err = couchTmpl.Execute(&couchBuf, data)
				if err != nil {
					return fmt.Errorf("failed to execute couch template for %s: %w", org.Name, err)
				}
				couchFilePath := filepath.Join(orgComposeDir, "compose-couch.yaml")
				if err := config.WriteFileAtomic(couchFilePath, couchBuf.Bytes(), 0600); err != nil {
					return fmt.Errorf("failed to write %s: %w", couchFilePath, err)
				}
			}
		}

		if org.OrdererCount > 0 {
			var buf bytes.Buffer
			err = ordererTmpl.Execute(&buf, data)
			if err != nil {
				return fmt.Errorf("failed to execute orderer template for %s: %w", org.Name, err)
			}
			filePath := filepath.Join(orgComposeDir, "compose-orderer.yaml")
			if err := config.WriteFileAtomic(filePath, buf.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}
		}

		// Generate CA compose if using Fabric CA
		if cfg.CryptoStrategy == "Fabric CA" {
			var buf bytes.Buffer
			err = caComposeTmpl.Execute(&buf, data)
			if err != nil {
				return fmt.Errorf("failed to execute CA compose template for %s: %w", org.Name, err)
			}
			filePath := filepath.Join(orgComposeDir, "compose-ca.yaml")
			if err := config.WriteFileAtomic(filePath, buf.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}
		}
	}

	// 3. Generate Hub Compose Files
	hubDir := filepath.Join(outputDir, "compose")

	// Generate compose/compose-peers.yaml
	var peersBuf bytes.Buffer
	err = hubPeersTmpl.Execute(&peersBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-peers template: %w", err)
	}
	peersPath := filepath.Join(hubDir, "compose-peers.yaml")
	if err := config.WriteFileAtomic(peersPath, peersBuf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", peersPath, err)
	}

	// Generate compose/compose-orderers.yaml
	var orderersBuf bytes.Buffer
	err = hubOrderersTmpl.Execute(&orderersBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-orderers template: %w", err)
	}
	orderersPath := filepath.Join(hubDir, "compose-orderers.yaml")
	if err := config.WriteFileAtomic(orderersPath, orderersBuf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", orderersPath, err)
	}

	// Generate compose/docker-compose.yaml (Master)
	var masterBuf bytes.Buffer
	err = hubMasterTmpl.Execute(&masterBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-master template: %w", err)
	}
	masterPath := filepath.Join(hubDir, "docker-compose.yaml")
	if err := config.WriteFileAtomic(masterPath, masterBuf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", masterPath, err)
	}

	// Generate compose/compose-cas.yaml if using Fabric CA
	if cfg.CryptoStrategy == "Fabric CA" {
		var casBuf bytes.Buffer
		err = hubCAsTmpl.Execute(&casBuf, cfg)
		if err != nil {
			return fmt.Errorf("failed to execute hub-cas template: %w", err)
		}
		casPath := filepath.Join(hubDir, "compose-cas.yaml")
		if err := config.WriteFileAtomic(casPath, casBuf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", casPath, err)
		}
	}

	// Generate compose/compose-utils.yaml and config/servers.json if using Postgres
	if cfg.CADatabaseType == "postgres" && cfg.CryptoStrategy == "Fabric CA" {
		configDir := filepath.Join(outputDir, "config")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		var serversBuf bytes.Buffer
		err = serversTmpl.Execute(&serversBuf, cfg)
		if err != nil {
			return fmt.Errorf("failed to execute servers template: %w", err)
		}
		serversPath := filepath.Join(configDir, "pgadmin-servers.json")
		if err := config.WriteFileAtomic(serversPath, serversBuf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", serversPath, err)
		}

		var utilsBuf bytes.Buffer
		err = utilsTmpl.Execute(&utilsBuf, cfg)
		if err != nil {
			return fmt.Errorf("failed to execute compose-utils template: %w", err)
		}
		utilsPath := filepath.Join(hubDir, "compose-utils.yaml")
		if err := config.WriteFileAtomic(utilsPath, utilsBuf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", utilsPath, err)
		}
	}

	return nil
}


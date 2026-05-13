package docker

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
func (g *DockerComposeGenerator) Generate(cfg *config.NetworkConfig, outputDir string) error {
	funcs := template.FuncMap{
		"until": func(count int) []int {
			var r []int
			for i := 0; i < count; i++ {
				r = append(r, i)
			}
			return r
		},
		"calculatePeerPort": func(orgIndex, peerIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 4000 + (peerIndex * 100) + (offset % 100)
		},
		"calculateOrdererPort": func(orgIndex, ordererIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 3000 + (ordererIndex * 100) + (offset % 100)
		},
		"calculateCAPort": func(orgIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 2000 + (offset % 100)
		},
		"calculateDatabasePort": func(orgIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 1000 + (offset % 100)
		},
		"toLower": strings.ToLower,
		"orgHasCouchDB": func(org config.OrgConfig) bool {
			for p := 0; p < org.PeerCount; p++ {
				if org.PeerStateDatabase(p) == "couchdb" {
					return true
				}
			}
			return false
		},
	}

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

	// 2. Generate Organizational Compose Files
	for i, org := range cfg.Orgs {
		orgComposeDir := filepath.Join(outputDir, "organizations", strings.ToLower(org.Name), "compose")
		err := os.MkdirAll(orgComposeDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create compose directory for org %s: %w", org.Name, err)
		}

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
			err = os.WriteFile(filePath, buf.Bytes(), 0644)
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}
			fmt.Printf("Generated %s\n", filePath)

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
				err = os.WriteFile(couchFilePath, couchBuf.Bytes(), 0644)
				if err != nil {
					return fmt.Errorf("failed to write %s: %w", couchFilePath, err)
				}
				fmt.Printf("Generated %s\n", couchFilePath)
			}
		}

		if org.OrdererCount > 0 {
			var buf bytes.Buffer
			err = ordererTmpl.Execute(&buf, data)
			if err != nil {
				return fmt.Errorf("failed to execute orderer template for %s: %w", org.Name, err)
			}
			filePath := filepath.Join(orgComposeDir, "compose-orderer.yaml")
			err = os.WriteFile(filePath, buf.Bytes(), 0644)
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}
			fmt.Printf("Generated %s\n", filePath)
		}

		// Generate CA compose if using Fabric CA
		if cfg.CryptoStrategy == "Fabric CA" {
			var buf bytes.Buffer
			err = caComposeTmpl.Execute(&buf, data)
			if err != nil {
				return fmt.Errorf("failed to execute CA compose template for %s: %w", org.Name, err)
			}
			filePath := filepath.Join(orgComposeDir, "compose-ca.yaml")
			err = os.WriteFile(filePath, buf.Bytes(), 0644)
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", filePath, err)
			}
			fmt.Printf("Generated %s\n", filePath)
		}
	}

	// 3. Generate Hub Compose Files
	hubDir := filepath.Join(outputDir, "compose")
	err = os.MkdirAll(hubDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create hub directory: %w", err)
	}

	// Generate compose/compose-peers.yaml
	var peersBuf bytes.Buffer
	err = hubPeersTmpl.Execute(&peersBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-peers template: %w", err)
	}
	err = os.WriteFile(filepath.Join(hubDir, "compose-peers.yaml"), peersBuf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write compose/compose-peers.yaml: %w", err)
	}
	fmt.Printf("Generated %s\n", filepath.Join(hubDir, "compose-peers.yaml"))

	// Generate compose/compose-orderers.yaml
	var orderersBuf bytes.Buffer
	err = hubOrderersTmpl.Execute(&orderersBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-orderers template: %w", err)
	}
	err = os.WriteFile(filepath.Join(hubDir, "compose-orderers.yaml"), orderersBuf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write compose/compose-orderers.yaml: %w", err)
	}
	fmt.Printf("Generated %s\n", filepath.Join(hubDir, "compose-orderers.yaml"))

	// Generate compose/docker-compose.yaml (Master)
	var masterBuf bytes.Buffer
	err = hubMasterTmpl.Execute(&masterBuf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute hub-master template: %w", err)
	}
	err = os.WriteFile(filepath.Join(hubDir, "docker-compose.yaml"), masterBuf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write compose/docker-compose.yaml: %w", err)
	}
	fmt.Printf("Generated %s\n", filepath.Join(hubDir, "docker-compose.yaml"))

	// Generate compose/compose-cas.yaml if using Fabric CA
	if cfg.CryptoStrategy == "Fabric CA" {
		var casBuf bytes.Buffer
		err = hubCAsTmpl.Execute(&casBuf, cfg)
		if err != nil {
			return fmt.Errorf("failed to execute hub-cas template: %w", err)
		}
		err = os.WriteFile(filepath.Join(hubDir, "compose-cas.yaml"), casBuf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("failed to write compose/compose-cas.yaml: %w", err)
		}
		fmt.Printf("Generated %s\n", filepath.Join(hubDir, "compose-cas.yaml"))
	}

	return nil
}

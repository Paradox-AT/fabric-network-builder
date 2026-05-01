package docker

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"network-builder/src/config"
)

//go:embed templates/docker-compose.yaml.tmpl
var templateFS embed.FS

// DockerComposeGenerator generates a docker-compose.yaml file
type DockerComposeGenerator struct{}

// NewDockerComposeGenerator returns a new instance of DockerComposeGenerator
func NewDockerComposeGenerator() *DockerComposeGenerator {
	return &DockerComposeGenerator{}
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
			return 10000 + (orgIndex * 1000) + (peerIndex * 100) + offset
		},
		"calculateOrdererPort": func(orgIndex, ordererIndex, offset int) int {
			return 20000 + (orgIndex * 1000) + (ordererIndex * 100) + offset
		},
	}

	tmpl, err := template.New("docker-compose.yaml.tmpl").Funcs(funcs).ParseFS(templateFS, "templates/docker-compose.yaml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	composeDir := filepath.Join(outputDir, "compose")
	err = os.MkdirAll(composeDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create compose output directory: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	if err != nil {
		return fmt.Errorf("failed to execute docker-compose template: %w", err)
	}

	filePath := filepath.Join(composeDir, "compose-test-net.yaml")
	err = os.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write compose-test-net.yaml: %w", err)
	}

	fmt.Printf("Generated %s\n", filePath)
	return nil
}

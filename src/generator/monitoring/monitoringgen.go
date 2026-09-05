package monitoring

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

type MonitoringGenerator struct{}

func NewMonitoringGenerator() *MonitoringGenerator {
	return &MonitoringGenerator{}
}

func (g *MonitoringGenerator) Generate(ctx context.Context, cfg *config.NetworkConfig, outputDir string) error {
	if !cfg.EnableMonitoring {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	funcs := utils.GetFuncMap()

	type tmplTarget struct {
		TmplName string
		DestPath string
		Perm     uint32
	}

	targets := []tmplTarget{
		{"compose-monitoring.yaml.tmpl", filepath.Join(outputDir, "compose", "compose-monitoring.yaml"), 0600},
		{"prometheus.yml.tmpl", filepath.Join(outputDir, "config", "prometheus", "prometheus.yml"), 0600},
		{"grafana-datasource.yaml.tmpl", filepath.Join(outputDir, "config", "grafana", "provisioning", "datasources", "datasource.yaml"), 0600},
	}

	for _, t := range targets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tmpl, err := template.New(t.TmplName).Funcs(funcs).ParseFS(templateFS, "templates/"+t.TmplName)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", t.TmplName, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to execute %s: %w", t.TmplName, err)
		}

		if err := config.WriteFileAtomic(t.DestPath, buf.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", t.DestPath, err)
		}
	}

	return nil
}

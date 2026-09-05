package scripts_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"network-builder/src/config"
	"network-builder/src/generator/scripts"
)

func TestScriptsGenerator_BashSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scripts_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.NetworkConfig{
		NetworkName:    "test-network",
		FabricVersion:  "3.1.5",
		CAVersion:      "1.5.15",
		CryptoStrategy: "Fabric CA",
		Orgs: []config.OrgConfig{
			{
				Name:         "Org1",
				Domain:       "org1.example.com",
				MSPID:        "Org1MSP",
				CAName:       "Org1CA",
				PeerCount:    2,
				OrdererCount: 1,
			},
			{
				Name:         "Org2",
				Domain:       "org2.example.com",
				MSPID:        "Org2MSP",
				CAName:       "Org2CA",
				PeerCount:    1,
				OrdererCount: 0,
			},
		},
	}

	gen := scripts.NewScriptsGenerator()
	err = gen.Generate(context.Background(), cfg, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify all generated .sh files pass `bash -n`
	shFiles := []string{
		"network.sh",
		"setOrgEnv.sh",
		"monitordocker.sh",
		"scripts/envVar.sh",
		"scripts/utils.sh",
		"scripts/createChannel.sh",
		"scripts/bootstrap.sh",
		"scripts/deployCC.sh",
		"scripts/ccutils.sh",
		"scripts/packageCC.sh",
		"scripts/addOrg.sh",
	}

	for _, relPath := range shFiles {
		fullPath := filepath.Join(tmpDir, relPath)
		cmd := exec.Command("bash", "-n", fullPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("bash -n failed on %s: %v\nOutput:\n%s", relPath, err, string(output))
		}
	}
}

func TestRegenerateNetwork(t *testing.T) {
	networkDir := filepath.Join("..", "..", "..", "network")
	cfgPath := filepath.Join(networkDir, "network-config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Skip("network-config.json not found")
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	gen := scripts.NewScriptsGenerator()
	if err := gen.Generate(context.Background(), cfg, networkDir); err != nil {
		t.Fatalf("ScriptsGenerator failed: %v", err)
	}
}

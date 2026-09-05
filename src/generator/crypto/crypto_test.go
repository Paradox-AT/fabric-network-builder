package crypto_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"network-builder/src/config"
	"network-builder/src/generator/crypto"
)

func TestFabricCAGenerator_BashSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "crypto_test_*")
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
		},
	}

	gen := crypto.NewFabricCAGenerator()
	err = gen.Generate(context.Background(), cfg, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	registerEnrollPath := filepath.Join(tmpDir, "organizations", "org1", "scripts", "registerEnroll.sh")
	cmd := exec.Command("bash", "-n", registerEnrollPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash -n failed on registerEnroll.sh: %v\nOutput:\n%s", err, string(output))
	}
}

func TestRegenerateCrypto(t *testing.T) {
	networkDir := filepath.Join("..", "..", "..", "network")
	cfgPath := filepath.Join(networkDir, "network-config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Skip("network-config.json not found")
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	gen := crypto.NewFabricCAGenerator()
	if err := gen.Generate(context.Background(), cfg, networkDir); err != nil {
		t.Fatalf("FabricCAGenerator failed: %v", err)
	}
}

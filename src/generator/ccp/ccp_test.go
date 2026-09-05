package ccp_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"network-builder/src/config"
	"network-builder/src/generator/ccp"
)

func TestCCPGenerator_BashSyntax(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ccp_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.NetworkConfig{
		NetworkName: "test-network",
		GenerateCCP: true,
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

	gen := ccp.NewCCPGenerator()
	err = gen.Generate(context.Background(), cfg, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	ccpGenPath := filepath.Join(tmpDir, "organizations", "ccp-generate.sh")
	cmd := exec.Command("bash", "-n", ccpGenPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash -n failed on ccp-generate.sh: %v\nOutput:\n%s", err, string(output))
	}
}

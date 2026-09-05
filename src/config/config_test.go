package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"network-builder/src/config"
)

func TestNetworkConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.NetworkConfig
		wantErr bool
	}{
		{
			name: "Valid Raft Configuration",
			cfg: config.NetworkConfig{
				NetworkName:   "valid-network",
				OrdererType:   "etcdraft",
				ChannelCount:  1,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 1,
						PeerCount:    2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid SmartBFT Configuration",
			cfg: config.NetworkConfig{
				NetworkName:   "bft-network",
				OrdererType:   "BFT",
				ChannelCount:  1,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 4,
						PeerCount:    2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid SmartBFT Quorum (< 4 orderers)",
			cfg: config.NetworkConfig{
				NetworkName:   "bft-invalid",
				OrdererType:   "BFT",
				ChannelCount:  1,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 3,
						PeerCount:    2,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Duplicate MSP IDs",
			cfg: config.NetworkConfig{
				NetworkName:   "dup-msp",
				OrdererType:   "etcdraft",
				ChannelCount:  1,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 1,
						PeerCount:    1,
					},
					{
						Name:         "Org2",
						MSPID:        "Org1MSP",
						Domain:       "org2.example.com",
						OrdererCount: 0,
						PeerCount:    1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Config with Monitoring and CCP enabled",
			cfg: config.NetworkConfig{
				NetworkName:      "test-net",
				OrdererType:      "etcdraft",
				EnableMonitoring: true,
				GenerateCCP:      true,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 1,
						PeerCount:    1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Network Name with uppercase spaces",
			cfg: config.NetworkConfig{
				NetworkName:  "Invalid Network Name",
				OrdererType:  "etcdraft",
				ChannelCount: 1,
				Orgs: []config.OrgConfig{
					{
						Name:         "Org1",
						MSPID:        "Org1MSP",
						Domain:       "org1.example.com",
						OrdererCount: 1,
						PeerCount:    1,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoadConfig_Atomic(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "network-config.json")

	originalCfg := &config.NetworkConfig{
		NetworkName:      "test-network",
		FabricVersion:    "3.1.5",
		CAVersion:        "1.5.19",
		CouchDBVersion:   "3.5.2",
		CADatabaseType:   "sqlite",
		OrdererType:      "etcdraft",
		CryptoStrategy:   "cryptogen",
		DeploymentTarget: "Docker Compose",
		ChaincodeMode:    "Embedded",
		ChannelCount:     1,
		BindAddress:      "0.0.0.0",
		EnableMonitoring: true,
		GenerateCCP:      true,
		Orgs: []config.OrgConfig{
			{
				Name:         "Org1",
				MSPID:        "Org1MSP",
				Domain:       "org1.example.com",
				CAName:       "ca-org1",
				OrdererCount: 1,
				PeerCount:    2,
			},
		},
	}

	if err := config.SaveConfig(originalCfg, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file permissions (0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("os.Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}

	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedCfg.NetworkName != originalCfg.NetworkName {
		t.Errorf("expected network name %s, got %s", originalCfg.NetworkName, loadedCfg.NetworkName)
	}
	if loadedCfg.EnableMonitoring != true {
		t.Errorf("expected EnableMonitoring true, got false")
	}
	if loadedCfg.GenerateCCP != true {
		t.Errorf("expected GenerateCCP true, got false")
	}
}

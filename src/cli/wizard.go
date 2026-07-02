package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"network-builder/src/config"
)

// RunWizard executes the interactive prompt and returns the populated NetworkConfig
func RunWizard(existingCfg *config.NetworkConfig) (*config.NetworkConfig, error) {
	fmt.Println("=== Hyperledger Fabric Network Builder ===")
	fmt.Println("This wizard will help you configure your enterprise blockchain network.")
	fmt.Println("Use Tab/Shift+Tab to navigate back and forth between pages.")
	fmt.Println()

	var cfg *config.NetworkConfig
	var orgCountStr string
	var channelCountStr string

	if existingCfg != nil {
		cfg = existingCfg
		orgCountStr = strconv.Itoa(len(cfg.Orgs))
		channelCountStr = strconv.Itoa(cfg.ChannelCount)
	} else {
		cfg = &config.NetworkConfig{
			NetworkName:      "fabric-network",
			FabricVersion:    "3.1.4",
			CAVersion:        "1.5.19",
			CouchDBVersion:   "3.5.1",
			CADatabaseType:   "sqlite",
			PostgresVersion:  "16.2",
			OrdererType:      "etcdraft",
			CryptoStrategy:   "cryptogen",
			DeploymentTarget: "Docker Compose",
			ChaincodeMode:    "Embedded",
			ChannelCount:     1,
			BindAddress:      "0.0.0.0",
		}
		orgCountStr = "3"
		channelCountStr = "1"
	}

	for {
		// 1. Global Network Setup
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What is the name of your network?").
					Value(&cfg.NetworkName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Fabric Version:").
					Value(&cfg.FabricVersion).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Fabric CA Version (if used):").
					Value(&cfg.CAVersion).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title("CouchDB Version (if used):").
					Value(&cfg.CouchDBVersion).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose the ordering service consensus type:").
					Options(
						huh.NewOption("Raft (CFT)", "etcdraft"),
						huh.NewOption("SmartBFT (BFT)", "BFT"),
					).
					Value(&cfg.OrdererType),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose the cryptographic material strategy:").
					Options(
						huh.NewOption("cryptogen (Local testing only)", "cryptogen"),
						huh.NewOption("Fabric CA (Enterprise standard)", "Fabric CA"),
					).
					Value(&cfg.CryptoStrategy),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose the CA database type:").
					Options(
						huh.NewOption("SQLite (Embedded, testing)", "sqlite"),
						huh.NewOption("PostgreSQL (External, production)", "postgres"),
					).
					Value(&cfg.CADatabaseType),
			).WithHideFunc(func() bool {
				return cfg.CryptoStrategy != "Fabric CA"
			}),
			huh.NewGroup(
				huh.NewInput().
					Title("PostgreSQL Version:").
					Value(&cfg.PostgresVersion),
			).WithHideFunc(func() bool {
				return cfg.CADatabaseType != "postgres" || cfg.CryptoStrategy != "Fabric CA"
			}),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What is your deployment target?").
					Options(
						huh.NewOption("Docker Compose", "Docker Compose"),
						huh.NewOption("Kubernetes", "Kubernetes"),
					).
					Value(&cfg.DeploymentTarget),
				huh.NewSelect[string]().
					Title("Chaincode Execution Model:").
					Options(
						huh.NewOption("Embedded (Default)", "Embedded"),
						huh.NewOption("External Chaincode-as-a-Service (CCaaS)", "CCaaS"),
					).
					Value(&cfg.ChaincodeMode),
				huh.NewSelect[string]().
					Title("Bind Address (Port Exposure):").
					Description("Choose 0.0.0.0 for external access or 127.0.0.1 for local only.").
					Options(
						huh.NewOption("0.0.0.0 (All interfaces - recommended for Portainer)", "0.0.0.0"),
						huh.NewOption("127.0.0.1 (Localhost only - more secure)", "127.0.0.1"),
					).
					Value(&cfg.BindAddress),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("How many Organizations will participate in the network?").
					Value(&orgCountStr).
					Validate(func(s string) error {
						v, err := strconv.Atoi(s)
						if err != nil || v <= 0 {
							return errors.New("must be a positive number")
						}
						return nil
					}),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("How many application channels do you need at launch?").
					Value(&channelCountStr).
					Validate(func(s string) error {
						v, err := strconv.Atoi(s)
						if err != nil || v <= 0 {
							return errors.New("must be a positive number")
						}
						return nil
					}),
			),
		).Run()

		if err != nil {
			return nil, err
		}

		// 2. Allocate Organizations
		orgCount, _ := strconv.Atoi(orgCountStr)
		if len(cfg.Orgs) != orgCount {
			newOrgs := make([]config.OrgConfig, orgCount)
			for i := 0; i < orgCount; i++ {
				if i < len(cfg.Orgs) {
					newOrgs[i] = cfg.Orgs[i]
				} else {
					// Initialize new org with defaults
					newOrgs[i] = config.OrgConfig{
						Name:          fmt.Sprintf("Org%d", i+1),
						MSPID:         fmt.Sprintf("Org%dMSP", i+1),
						Domain:        fmt.Sprintf("org%d.example.com", i+1),
						CAName:        fmt.Sprintf("ca-org%d", i+1),
						OrdererCount:  1,
						PeerCount:     2,
						StateDatabase: "leveldb",
					}
				}
			}
			cfg.Orgs = newOrgs
		}

		// 3. Organization Specific Details
		for i := 0; i < orgCount; i++ {
			ordererCountStr := strconv.Itoa(cfg.Orgs[i].OrdererCount)
			peerCountStr := strconv.Itoa(cfg.Orgs[i].PeerCount)

			err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] Organization Name:", i+1)).
						Value(&cfg.Orgs[i].Name),
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] MSP ID:", i+1)).
						Value(&cfg.Orgs[i].MSPID),
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] Base Domain:", i+1)).
						Value(&cfg.Orgs[i].Domain),
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] Certificate Authority (CA) Name:", i+1)).
						Value(&cfg.Orgs[i].CAName),
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] Number of orderer nodes:", i+1)).
						Value(&ordererCountStr),
					huh.NewInput().
						Title(fmt.Sprintf("[Org %d] Number of peers:", i+1)).
						Value(&peerCountStr),
				),
			).Run()

			if err != nil {
				return nil, err
			}

			cfg.Orgs[i].OrdererCount, _ = strconv.Atoi(ordererCountStr)
			cfg.Orgs[i].PeerCount, _ = strconv.Atoi(peerCountStr)

			// Initialize the Peers slice to match the declared peer count
			if len(cfg.Orgs[i].Peers) != cfg.Orgs[i].PeerCount {
				newPeers := make([]config.PeerConfig, cfg.Orgs[i].PeerCount)
				for p := 0; p < cfg.Orgs[i].PeerCount && p < len(cfg.Orgs[i].Peers); p++ {
					newPeers[p] = cfg.Orgs[i].Peers[p]
				}
				cfg.Orgs[i].Peers = newPeers
			}

			// Ask state database preference per peer
			for p := 0; p < cfg.Orgs[i].PeerCount; p++ {
				if cfg.Orgs[i].Peers[p].StateDatabase == "" {
					cfg.Orgs[i].Peers[p].StateDatabase = "leveldb"
				}
				peerDB := cfg.Orgs[i].Peers[p].StateDatabase
				err := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title(fmt.Sprintf("[Org %d / Peer %d] State database:", i+1, p)).
							Options(
								huh.NewOption("LevelDB (Default, embedded)", "leveldb"),
								huh.NewOption("CouchDB (Rich JSON queries)", "couchdb"),
							).
							Value(&peerDB),
					),
				).Run()
				if err != nil {
					return nil, err
				}
				cfg.Orgs[i].Peers[p].StateDatabase = peerDB
			}
		} // end org loop

		// 4. Network-wide Validation
		totalOrderers := 0
		for _, org := range cfg.Orgs {
			totalOrderers += org.OrdererCount
		}

		isValid := true
		var errMsg string

		switch cfg.OrdererType {
		case "etcdraft":
			if totalOrderers < 1 {
				isValid = false
				errMsg = "Raft (etcdraft) requires at least 1 orderer node."
			}
		case "BFT":
			if totalOrderers < 4 || (totalOrderers-1)%3 != 0 {
				isValid = false
				errMsg = fmt.Sprintf("SmartBFT requires 3f+1 nodes (4, 7, 10...). You have %d.", totalOrderers)
			}
		}

		if isValid {
			break
		}

		fmt.Println("\n[!] VALIDATION ERROR:", errMsg)
		fmt.Println("Press Enter to fix your configuration...")
		var dummy string
		fmt.Scanln(&dummy)
	}

	cfg.ChannelCount, _ = strconv.Atoi(channelCountStr)
	return cfg, nil
}

// PrintSummary outputs a human-readable summary of the collected configuration
func PrintSummary(cfg *config.NetworkConfig) {
	fmt.Println("\n=== Network Configuration Summary ===")
	fmt.Printf("Network Name:    %s\n", cfg.NetworkName)
	fmt.Printf("Fabric Version:  %s\n", cfg.FabricVersion)
	fmt.Printf("CA Version:      %s\n", cfg.CAVersion)
	fmt.Printf("CouchDB Version: %s\n", cfg.CouchDBVersion)
	fmt.Printf("Consensus:       %s\n", cfg.OrdererType)
	fmt.Printf("Crypto Strategy: %s\n", cfg.CryptoStrategy)
	if cfg.CryptoStrategy == "Fabric CA" {
		fmt.Printf("  CA Database:   %s\n", cfg.CADatabaseType)
		if cfg.CADatabaseType == "postgres" {
			fmt.Printf("  Postgres Ver:  %s\n", cfg.PostgresVersion)
		}
	}
	fmt.Printf("Deploy Target:   %s\n", cfg.DeploymentTarget)
	fmt.Printf("Chaincode Mode:  %s\n", cfg.ChaincodeMode)
	fmt.Printf("Bind Address:    %s\n", cfg.BindAddress)
	fmt.Printf("Channels:        %d\n", cfg.ChannelCount)

	totalOrderers := 0
	totalPeers := 0
	for _, o := range cfg.Orgs {
		totalOrderers += o.OrdererCount
		totalPeers += o.PeerCount
	}
	fmt.Printf("Total Organizations: %d\n", len(cfg.Orgs))
	fmt.Printf("Total Orderers:      %d\n", totalOrderers)
	fmt.Printf("Total Peers:         %d\n", totalPeers)

	fmt.Println("\n--- Organizations ---")
	for i, org := range cfg.Orgs {
		fmt.Printf("  %d. %s (MSP: %s, Domain: %s, Orderers: %d, Peers: %d)\n", i+1, org.Name, org.MSPID, org.Domain, org.OrdererCount, org.PeerCount)
	}

	// Warnings block
	hasWarnings := false
	warnings := "\n[!] ATTENTION: The following selections are captured but not yet fully implemented in our physical generators:\n"

	if cfg.DeploymentTarget == "Kubernetes" {
		hasWarnings = true
		warnings += "    - Kubernetes manifests (falling back to Docker Compose conceptually)\n"
	}
	if cfg.ChaincodeMode == "CCaaS" {
		hasWarnings = true
		warnings += "    - External Chaincode-as-a-Service builders in core.yaml\n"
	}
	if cfg.ChannelCount > 1 {
		hasWarnings = true
		warnings += "    - Multi-channel configtx.yaml profiles\n"
	}

	if hasWarnings {
		fmt.Println(warnings)
	}

	fmt.Println("\nConfiguration captured successfully. Ready to generate cryptographic material and channel artifacts.")
}

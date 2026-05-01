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
	const maxOrgs = 10
	ordererNodeCounts := make([]string, maxOrgs)
	peerNodeCounts := make([]string, maxOrgs)

	if existingCfg != nil {
		cfg = existingCfg
		orgCountStr = strconv.Itoa(len(cfg.Orgs))
		channelCountStr = strconv.Itoa(cfg.ChannelCount)

		// Pad existing orgs to maxOrgs so the form has enough fields
		for i := len(cfg.Orgs); i < maxOrgs; i++ {
			cfg.Orgs = append(cfg.Orgs, config.OrgConfig{
				Name:   fmt.Sprintf("Org%d", i+1),
				MSPID:  fmt.Sprintf("Org%dMSP", i+1),
				Domain: fmt.Sprintf("org%d.example.com", i+1),
				CAName: fmt.Sprintf("ca.org%d.example.com", i+1),
			})
		}
		
		for i := 0; i < maxOrgs; i++ {
			if i < len(existingCfg.Orgs) && i < cap(existingCfg.Orgs) && existingCfg.Orgs[i].OrdererCount > 0 { // Just populate counts based on what exists
                ordererNodeCounts[i] = strconv.Itoa(cfg.Orgs[i].OrdererCount)
            } else if i < len(existingCfg.Orgs) && existingCfg.Orgs[i].OrdererCount == 0 && existingCfg.Orgs[i].PeerCount > 0 {
				ordererNodeCounts[i] = "0"
			} else {
				ordererNodeCounts[i] = "1"
			}

			if i < len(existingCfg.Orgs) && i < cap(existingCfg.Orgs) && existingCfg.Orgs[i].PeerCount > 0 {
                peerNodeCounts[i] = strconv.Itoa(cfg.Orgs[i].PeerCount)
            } else if i < len(existingCfg.Orgs) && existingCfg.Orgs[i].PeerCount == 0 && existingCfg.Orgs[i].OrdererCount > 0 {
				peerNodeCounts[i] = "0"
			} else {
				peerNodeCounts[i] = "2"
			}
		}

	} else {
		cfg = &config.NetworkConfig{
			NetworkName:      "fabric-network",
			FabricVersion:    "3.1.4",
			OrdererType:      "etcdraft",
			CryptoStrategy:   "cryptogen",
			StateDatabase:    "LevelDB",
			DeploymentTarget: "Docker Compose",
			ChaincodeMode:    "Embedded",
			ChannelCount:     1,
		}

		orgCountStr = "3"
		channelCountStr = "1"

		cfg.Orgs = make([]config.OrgConfig, maxOrgs)

		for i := 0; i < maxOrgs; i++ {
			cfg.Orgs[i] = config.OrgConfig{
				Name:   fmt.Sprintf("Org%d", i+1),
				MSPID:  fmt.Sprintf("Org%dMSP", i+1),
				Domain: fmt.Sprintf("org%d.example.com", i+1),
				CAName: fmt.Sprintf("ca.org%d.example.com", i+1),
			}
			// Default to 1 orderer and 2 peers per org
			ordererNodeCounts[i] = "1"
			peerNodeCounts[i] = "2"
		}
	}

	for {
		groups := []*huh.Group{}

		// 1. General Network Setup (Split into 4 pages for readability)
		groups = append(groups,
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
					Title("Which Fabric Version will you use?").
					Value(&cfg.FabricVersion).
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
						huh.NewOption("etcdraft", "CFT"),
						huh.NewOption("SmartBFT", "BFT"),
					).
					Value(&cfg.OrdererType),
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
					Title("Which state database will you use?").
					Options(
						huh.NewOption("LevelDB (Default, Key-Value)", "LevelDB"),
						huh.NewOption("CouchDB (Rich JSON Queries)", "CouchDB"),
					).
					Value(&cfg.StateDatabase),
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
			),
			huh.NewGroup(
				huh.NewInput().
					Title("How many Organizations will participate in the network? (Max 10)").
					Value(&orgCountStr).
					Validate(func(s string) error {
						v, err := strconv.Atoi(s)
						if err != nil || v <= 0 || v > 10 {
							return errors.New("must be between 1 and 10")
						}
						return nil
					}),
			),
		)

		// 2. Organization Setup (Dynamic pages)
		for i := 0; i < maxOrgs; i++ {
			idx := i // Capture loop variable
			groups = append(groups, huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] Organization Name:", idx+1)).
					Value(&cfg.Orgs[idx].Name).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] MSP ID:", idx+1)).
					Value(&cfg.Orgs[idx].MSPID).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] Base Domain:", idx+1)).
					Value(&cfg.Orgs[idx].Domain).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] Certificate Authority (CA) Name:", idx+1)).
					Value(&cfg.Orgs[idx].CAName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("required")
						}
						return nil
					}),
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] Number of orderer nodes:", idx+1)).
					Value(&ordererNodeCounts[idx]).
					Validate(func(s string) error {
						v, err := strconv.Atoi(s)
						if err != nil || v < 0 {
							return errors.New("must be 0 or positive number")
						}
						return nil
					}),
				huh.NewInput().
					Title(fmt.Sprintf("[Org %d] Number of peers:", idx+1)).
					Value(&peerNodeCounts[idx]).
					Validate(func(s string) error {
						v, err := strconv.Atoi(s)
						if err != nil || v < 0 {
							return errors.New("must be 0 or positive number")
						}
						return nil
					}),
			).WithHideFunc(func() bool {
				count, _ := strconv.Atoi(orgCountStr)
				return idx >= count
			}))
		}

		// 3. Channel Configuration at the end
		groups = append(groups, huh.NewGroup(
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
		))

		err := huh.NewForm(groups...).Run()
		if err != nil {
			return nil, err
		}

		// Post-process counts
		orgCount, _ := strconv.Atoi(orgCountStr)
		totalOrderers := 0

		for i := 0; i < orgCount; i++ {
			oCount, _ := strconv.Atoi(ordererNodeCounts[i])
			totalOrderers += oCount
		}

		// Validation logic
		isValid := true
		var errMsg string

		if cfg.OrdererType == "CFT" {
			if totalOrderers < 1 {
				isValid = false
				errMsg = "Raft (CFT) requires at least 1 orderer node across the network."
			}
		} else if cfg.OrdererType == "BFT" {
			// SmartBFT requires 3f+1, meaning total-1 is divisible by 3, and minimum 4
			if totalOrderers < 4 || (totalOrderers-1)%3 != 0 {
				isValid = false
				errMsg = fmt.Sprintf("SmartBFT requires 3f+1 nodes (e.g. 4, 7, 10). You specified %d.", totalOrderers)
			}
		}

		if isValid {
			break
		}

		fmt.Println("\n[!] VALIDATION ERROR:", errMsg)
		fmt.Println("Please adjust your orderer node counts.")
		fmt.Println("Press Enter to return to the wizard and fix your configuration...")
		var dummy string
		fmt.Scanln(&dummy)
	}

	// Finalize struct
	orgCount, _ := strconv.Atoi(orgCountStr)
	cfg.ChannelCount, _ = strconv.Atoi(channelCountStr)
	cfg.Orgs = cfg.Orgs[:orgCount]

	for i := 0; i < orgCount; i++ {
		cfg.Orgs[i].OrdererCount, _ = strconv.Atoi(ordererNodeCounts[i])
		cfg.Orgs[i].PeerCount, _ = strconv.Atoi(peerNodeCounts[i])
	}

	return cfg, nil
}

// PrintSummary outputs a human-readable summary of the collected configuration
func PrintSummary(cfg *config.NetworkConfig) {
	fmt.Println("\n=== Network Configuration Summary ===")
	fmt.Printf("Network Name:    %s\n", cfg.NetworkName)
	fmt.Printf("Fabric Version:  %s\n", cfg.FabricVersion)
	fmt.Printf("Consensus:       %s\n", cfg.OrdererType)
	fmt.Printf("Crypto Strategy: %s\n", cfg.CryptoStrategy)
	fmt.Printf("State Database:  %s\n", cfg.StateDatabase)
	fmt.Printf("Deploy Target:   %s\n", cfg.DeploymentTarget)
	fmt.Printf("Chaincode Mode:  %s\n", cfg.ChaincodeMode)
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

	if cfg.StateDatabase == "CouchDB" {
		hasWarnings = true
		warnings += "    - CouchDB container orchestration and configuration\n"
	}
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

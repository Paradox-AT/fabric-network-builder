package config

import (
	"fmt"
	"regexp"
	"strings"
)

// NetworkConfig holds the overall Hyperledger Fabric network configuration
type NetworkConfig struct {
	NetworkName      string      `json:"networkName"`
	FabricVersion    string      `json:"fabricVersion"`
	CAVersion        string      `json:"caVersion"`
	CouchDBVersion   string      `json:"couchDBVersion"`
	CADatabaseType   string      `json:"caDatabaseType"` // "sqlite" or "postgres"
	PostgresVersion  string      `json:"postgresVersion"`
	OrdererType      string      `json:"ordererType"` // "etcdraft" or "BFT"
	CryptoStrategy   string      `json:"cryptoStrategy"` // "cryptogen" or "Fabric CA"
	DeploymentTarget string      `json:"deploymentTarget"`
	ChaincodeMode    string      `json:"chaincodeMode"`
	ChannelCount     int         `json:"channelCount"`
	BindAddress      string      `json:"bindAddress"` // IP address to bind ports to (e.g., "0.0.0.0" or "127.0.0.1")
	EnableMonitoring bool        `json:"enableMonitoring"` // Generate Prometheus + Grafana stack
	GenerateCCP      bool        `json:"generateCCP"`      // Generate Client Connection Profiles
	Orgs             []OrgConfig `json:"orgs"`
}

// PeerConfig holds per-peer configuration
type PeerConfig struct {
	StateDatabase string `json:"stateDatabase"` // "leveldb" or "couchdb"
}

// OrgConfig holds configuration for a Fabric organization
type OrgConfig struct {
	Name          string       `json:"name"`
	MSPID         string       `json:"mspID"`
	Domain        string       `json:"domain"`
	CAName        string       `json:"caName"`
	OrdererCount  int          `json:"ordererCount"`
	PeerCount     int          `json:"peerCount"`
	StateDatabase string       `json:"stateDatabase"` // org-level default (legacy / fallback)
	Peers         []PeerConfig `json:"peers"`          // per-peer overrides; len == PeerCount
}

// PeerStateDatabase returns the effective state database for a specific peer index.
// It prefers the per-peer setting if populated, otherwise falls back to the org default.
func (o *OrgConfig) PeerStateDatabase(peerIndex int) string {
	if peerIndex < len(o.Peers) && o.Peers[peerIndex].StateDatabase != "" {
		return o.Peers[peerIndex].StateDatabase
	}
	if o.StateDatabase != "" {
		return o.StateDatabase
	}
	return "leveldb"
}

// Validate checks the domain validity of the entire network topology configuration.
func (c *NetworkConfig) Validate() error {
	if strings.TrimSpace(c.NetworkName) == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	nameRegex := regexp.MustCompile("^[a-z0-9][a-z0-9_-]*$")
	if !nameRegex.MatchString(c.NetworkName) {
		return fmt.Errorf("invalid network name %q: must be lowercase alphanumeric with hyphens or underscores", c.NetworkName)
	}

	if len(c.Orgs) == 0 {
		return fmt.Errorf("network must contain at least one organization")
	}

	totalOrderers := 0
	totalPeers := 0
	seenOrgs := make(map[string]bool)
	seenMSPs := make(map[string]bool)
	seenDomains := make(map[string]bool)

	for i, org := range c.Orgs {
		if strings.TrimSpace(org.Name) == "" {
			return fmt.Errorf("org[%d] name cannot be empty", i)
		}
		if seenOrgs[strings.ToLower(org.Name)] {
			return fmt.Errorf("duplicate org name %q", org.Name)
		}
		seenOrgs[strings.ToLower(org.Name)] = true

		if strings.TrimSpace(org.MSPID) == "" {
			return fmt.Errorf("org %s: MSP ID cannot be empty", org.Name)
		}
		if seenMSPs[org.MSPID] {
			return fmt.Errorf("duplicate MSP ID %q in org %s", org.MSPID, org.Name)
		}
		seenMSPs[org.MSPID] = true

		if strings.TrimSpace(org.Domain) == "" {
			return fmt.Errorf("org %s: domain cannot be empty", org.Name)
		}
		if seenDomains[strings.ToLower(org.Domain)] {
			return fmt.Errorf("duplicate domain %q in org %s", org.Domain, org.Name)
		}
		seenDomains[strings.ToLower(org.Domain)] = true

		if org.PeerCount < 0 {
			return fmt.Errorf("org %s: peer count cannot be negative", org.Name)
		}
		if org.OrdererCount < 0 {
			return fmt.Errorf("org %s: orderer count cannot be negative", org.Name)
		}

		totalOrderers += org.OrdererCount
		totalPeers += org.PeerCount
	}

	if totalPeers == 0 {
		return fmt.Errorf("network must have at least one peer across all organizations")
	}

	switch c.OrdererType {
	case "etcdraft":
		if totalOrderers < 1 {
			return fmt.Errorf("Raft (etcdraft) consensus requires at least 1 orderer node across the network")
		}
	case "BFT":
		if totalOrderers < 4 {
			return fmt.Errorf("SmartBFT consensus requires at least 4 orderers to tolerate 1 Byzantine fault (3f+1). Got %d", totalOrderers)
		}
	default:
		return fmt.Errorf("unsupported consensus type %q: must be 'etcdraft' or 'BFT'", c.OrdererType)
	}

	return nil
}


package config

// NetworkConfig holds the overall Hyperledger Fabric network configuration
type NetworkConfig struct {
	NetworkName      string
	FabricVersion    string
	CAVersion        string
	CouchDBVersion   string
	CADatabaseType   string // "sqlite" or "postgres"
	PostgresVersion  string
	OrdererType      string
	CryptoStrategy   string
	DeploymentTarget string
	ChaincodeMode    string
	ChannelCount     int
	BindAddress      string // IP address to bind ports to (e.g., "0.0.0.0" or "127.0.0.1")
	Orgs             []OrgConfig
}

// PeerConfig holds per-peer configuration
type PeerConfig struct {
	StateDatabase string // "leveldb" or "couchdb"
}

// OrgConfig holds configuration for a Fabric organization
type OrgConfig struct {
	Name          string
	MSPID         string
	Domain        string
	CAName        string
	OrdererCount  int
	PeerCount     int
	StateDatabase string        // org-level default (legacy / fallback)
	Peers         []PeerConfig  // per-peer overrides; len == PeerCount
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

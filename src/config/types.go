package config

// NetworkConfig holds the overall Hyperledger Fabric network configuration
type NetworkConfig struct {
	NetworkName      string
	FabricVersion    string
	OrdererType      string
	CryptoStrategy   string
	StateDatabase    string
	DeploymentTarget string
	ChaincodeMode    string
	ChannelCount     int
	Orgs             []OrgConfig
}

// OrgConfig holds configuration for a Fabric organization
type OrgConfig struct {
	Name         string
	MSPID        string
	Domain       string
	CAName       string
	OrdererCount int
	PeerCount    int
}

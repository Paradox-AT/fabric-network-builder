# Hyperledger Fabric Network Builder

An enterprise-grade CLI wizard for scaffolding Hyperledger Fabric networks. This tool automates the generation of complex network artifacts including cryptographic material configurations, channel transaction profiles, and docker orchestration manifests.

## Key Features

- **Interactive Wizard**: A beautiful, terminal-based multi-page form powered by `charmbracelet/huh`.
- **Enterprise Topology**: Unified organization model allowing organizations to own both peers and orderers.
- **Consensus Validation**: Built-in verification for CFT (Raft) and BFT (SmartBFT) node counts (e.g., enforcing the `3f+1` rule).
- **Test-Network Compliance**: Generates folder structures and YAML files that strictly follow the official `fabric-samples/test-network` conventions.
- **Heavily Documented Artifacts**: Injects official Fabric documentation blocks into all generated YAML files.
- **State Persistence**: Saves your configuration to `network-config.json`, allowing you to resume, modify, or back up previous sessions.
- **Dynamic Port Mapping**: Automatically handles port collisions by assigning predictable ranges per organization.

## Quick Start

### Prerequisites
- Go 1.21+
- Hyperledger Fabric Binaries (for executing the generated artifacts)
- Docker & Docker Compose

### Running the Builder
```bash
go run main.go
```

## Folder Structure

The tool generates artifacts in the `./network` directory:
- `configtx/`: Contains the `configtx.yaml` channel profile.
- `organizations/cryptogen/`: Contains individual `crypto-config-*.yaml` files for each organization.
- `compose/`: Contains `compose-test-net.yaml` for container orchestration.
- `network-config.json`: Persisted state of your last wizard session.

## Consensus Rules
The builder strictly enforces network health:
- **etcdraft (CFT)**: Requires at least 1 orderer node.
- **SmartBFT (BFT)**: Requires at least 4 orderer nodes and follows the `3f+1` rule (4, 7, 10, etc.).

## Development

The project is modularly structured:
- `src/cli/`: Interactive wizard logic.
- `src/config/`: Domain models and IO persistence.
- `src/generator/`: Polymorphic engines for Crypto, Configtx, and Docker manifests.

---
Built with ❤️ for the Hyperledger Fabric community.

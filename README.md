# Hyperledger Fabric Network Builder 🚀

[![Hyperledger Fabric Version](https://img.shields.io/badge/Fabric-3.1.5-blue.svg)](https://hyperledger-fabric.readthedocs.io/)
[![Hyperledger Fabric CA](https://img.shields.io/badge/Fabric%20CA-1.5.19-green.svg)](https://hyperledger-fabric-ca.readthedocs.io/)
[![CouchDB Version](https://img.shields.io/badge/CouchDB-3.5.2-red.svg)](https://couchdb.apache.org/)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev/)

A production-grade, organization-centric CLI wizard for architecting and deploying Hyperledger Fabric 3.1.5 networks. This tool automates the entire lifecycle from cryptographic identity issuance and channel creation to chaincode lifecycle and metrics monitoring.

## ✨ Enterprise Features

-   **Modular Identity Management**:
    *   **Cryptogen**: Rapid development and local testing.
    *   **Fabric CA**: Production-ready PKI with support for **SQLite** or **PostgreSQL** backends and sovereign org enrollment scripts.
-   **Advanced Consensus**: Built-in validation and generation for **Raft (CFT)** and **SmartBFT (BFT)** with strict quorum rules (3f+1).
-   **Pluggable State Database**: Per-peer selection of **LevelDB** or **CouchDB** (v3.5.2+).
-   **Client Connection Profiles (CCP)**: Automated generation of `connection-{orgName}.json` and `connection-{orgName}.yaml` for client application SDKs.
-   **Full-Stack Observability**: Optional **Prometheus** metrics scraping and **Grafana** dashboard provisioning for peer, orderer, and system monitoring.
-   **Dynamic Organization Expansion**: Add new organizations to an already running network on-the-fly without downtime.
-   **Concurrent & Atomic Artifact Pipeline**: High-performance generation using Go `errgroup` concurrency, atomic file writes (`0600`/`0700` secure permissions), and interactive progress spinners.
-   **Modular Orchestration**: Generates organization-specific docker manifests and a centralized "Hub" for total network control.

## 🚀 Getting Started

### 1. Build the Network
Run the interactive CLI wizard to design your topology:
```bash
go run main.go
```

**CLI Flags:**
- `-v`, `--verbose`: Enable detailed per-file output during artifact generation.
- `-s`, `--skip-cleanup`, `--no-cleanup`: Retain existing artifacts and skip cleanup before generating.

### 2. Bootstrap Binaries
Download the required Fabric binaries and Docker images:
```bash
./network/network.sh bootstrap
```

### 3. Launch & Deploy
Bring up the network and create your first channel:
```bash
# Start network containers (add -u / --utility to include pgAdmin)
./network/network.sh up

# Create and join channel
./network/network.sh createChannel -c mychannel

# Deploy smart contracts
./network/network.sh deployCC -c mychannel -ccn basic -ccp ../chaincode/asset-transfer-basic/chaincode-go -ccl go
```

## 📂 Directory Structure

The builder creates a clean, sovereign organizational hierarchy:

```text
network/
├── bin/                       # Fabric binaries (post-bootstrap)
├── compose/                   # Centralized Hub orchestration (peers, orderers, compose-monitoring.yaml)
├── config/                    # Monitoring configurations (Prometheus & Grafana provisioning)
│   ├── grafana/
│   └── prometheus/
├── configtx/                  # Channel & genesis block profiles (configtx.yaml)
├── organizations/             # Sovereign organizational data
│   └── {orgName}/
│       ├── compose/           # Org-specific docker manifests (Peers, Orderers, CAs)
│       ├── connection-{org}.* # Connection profiles (JSON & YAML)
│       ├── identities/        # Generated certificates & MSP (excluded from git)
│       ├── identity-config/   # Cryptogen / CA configuration files
│       └── scripts/           # Org-specific enrollment & registration scripts
├── scripts/                   # Internal lifecycle, channel, & chaincode helpers
├── network-config.json        # Saved topology state for quick rebuilds / modifications
├── network.config             # Generated environment constants for shell scripts
└── network.sh                 # Unified management CLI
```

## 🛠️ Commands Reference

| Command | Description |
| :--- | :--- |
| `./network.sh up [-u]` | Generates crypto material and launches all network services. Pass `-u` or `--utility` to also launch utility containers (e.g. pgAdmin). |
| `./network.sh down` | Gracefully stops containers, cleans docker networks, and removes runtime artifacts. |
| `./network.sh restart` | Bounces all network containers while preserving volume states. |
| `./network.sh createChannel [-c <name>]` | Automates genesis block creation, channel creation, peer joins, and anchor peer definitions. |
| `./network.sh packageCC` | Packages smart contract into a deployable chaincode tarball (`.tar.gz`). |
| `./network.sh deployCC` | Executes the complete 4-step lifecycle: package, install, approve, and commit chaincode across orgs. |
| `./network.sh addOrg` | Generates artifacts and executes channel configuration update transactions to onboard a new organization. |
| `./network.sh monitor` | Streams real-time Docker container CPU, memory, and I/O metrics. |
| `./network.sh setOrgEnv` | Exports environment variables (`CORE_PEER_*`) for targeting specific organizations in CLI interactions. |

## 🧪 Tested Environment

- **Hyperledger Fabric**: 3.1.5
- **Fabric CA**: 1.5.19
- **Go**: 1.22+
- **Docker**: 24.0+ / Docker Compose v2+
- **OS**: Linux (Arch, Ubuntu 22.04/24.04), macOS

## 🛡️ Security Best Practices

- **Pre-configured Git Ignore**: Automatically ignores sensitive private keys and credentials inside `identities/` while preserving configurations and templates.
- **Secure File Permissions**: Generated configuration files and identity states are written atomically with restricted `0600` file permissions and `0700` directory permissions.
- **Default Credentials Warning**: Generated `.env` files contain default development passwords; ensure you rotate them prior to any production or shared deployment.

---
Built with ❤️ for the Hyperledger Community.

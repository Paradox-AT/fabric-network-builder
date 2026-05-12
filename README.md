# Hyperledger Fabric Network Builder 🚀

[![Fabric Version](https://img.shields.io/badge/Fabric-3.1.4-blue.svg)](https://hyperledger-fabric.readthedocs.io/)
[![Fabric CA](https://img.shields.io/badge/Fabric%20CA-1.5.12-green.svg)](https://hyperledger-fabric-ca.readthedocs.io/)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev/)

A production-grade, organization-centric CLI wizard for architecting and deploying Hyperledger Fabric 3.1.4 networks. This tool automates the entire lifecycle from identity issuance to channel creation.

## ✨ Enterprise Features

-   **Modular Identity Management**:
    *   **Cryptogen**: Rapid development and local testing.
    *   **Fabric CA**: Production-ready PKI with support for **SQLite** or **PostgreSQL** backends.
-   **Advanced Consensus**: Built-in validation and generation for **Raft (CFT)** and **SmartBFT (BFT)**.
-   **Pluggable State Database**: Per-peer selection of **LevelDB** or **CouchDB** (v3.3.3+).
-   **Modular Orchestration**: Generates organization-specific docker manifests and a centralized "Hub" for total network control.
-   **Lifecycle Automation**: A robust `network.sh` script that handles:
    *   Zero-config certificate generation (CA or Cryptogen).
    *   Multi-stage docker orchestration.
    *   Automated channel creation, anchor peer updates, and organization joining.

## 🚀 Getting Started

### 1. Build the Network
Run the interactive wizard to design your topology:
```bash
go run main.go
```

### 2. Bootstrap Binaries
Download the required Fabric binaries and Docker images:
```bash
./network/network.sh bootstrap
```

### 3. Launch & Deploy
Bring up the network and create your first channel:
```bash
./network/network.sh up
./network/network.sh createChannel mychannel
```

## 📂 Directory Structure

The builder creates a clean, sovereign organizational hierarchy:

```text
network/
├── bin/                  # Fabric binaries (post-bootstrap)
├── compose/              # Centralized Hub orchestration
├── configtx/             # Global channel profiles
├── organizations/        # Sovereign organizational data
│   └── {orgName}/
│       ├── compose/      # Org-specific docker manifests (Peers/Orderers/CAs)
│       ├── identities/   # Generated certificates (MSP/TLS)
│       ├── identity-config/ # Cryptogen/CA configuration files
│       └── scripts/      # Org-specific enrollment & registration
├── scripts/              # Internal lifecycle helpers
└── network.sh            # Unified entry point
```

## 🛠️ Commands Reference

| Command | Description |
| :--- | :--- |
| `./network.sh up` | Generates identities and starts all network components. |
| `./network.sh down` | Stops containers and removes runtime artifacts. |
| `./network.sh restart` | Quickly bounces the network while preserving volumes. |
| `./network.sh createChannel` | Automates the genesis-to-join flow for a channel. |

## 🧪 Tested Environment

- **Hyperledger Fabric**: 3.1.4
- **Fabric CA**: 1.5.12
- **Go**: 1.22+
- **Docker**: 24.0+
- **OS**: Linux (Optimized for Ubuntu/Arch)

## 🛡️ Security Best Practices

The tool is pre-configured with a `.gitignore` that avoids tracking sensitive `identities/` folders while ensuring your **configuration templates** and **wizard state** (network-config.json) are safely versioned.

---
Built with ❤️ for the Hyperledger Community.

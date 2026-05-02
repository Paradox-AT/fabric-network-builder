# 🚀 Hyperledger Fabric Network Builder

An enterprise-grade, **Sovereign Architecture** CLI tool for scaffolding production-ready Hyperledger Fabric networks. This tool automates the entire lifecycle of a blockchain network, from identity generation to container orchestration and channel creation.

## 🏛️ Sovereign Architecture

Unlike the generic `test-network`, this builder implements a **Modular Sovereignty** model:
- **Independent Organizations**: Each organization is self-contained with its own `identity-config`, `identities`, and `compose` manifests.
- **Centralized Hub**: Orchestration is managed via a root-level hub that aggregates organizational infrastructure using Docker Compose `include`.
- **Environment Isolation**: Absolute pathing ensures scripts work from any directory.

## ✨ Key Features

- **Interactive Wizard**: A beautiful terminal UI for complex network configuration.
- **Automated Bootstrapping**: One-command initialization for Fabric binaries and Docker images.
- **Smart Pathing**: Intelligent `${ROOTDIR}` resolution for robust lifecycle management.
- **State Persistence**: Configuration is saved to `network-config.json` for easy regeneration.
- **Advanced Networking**: Automated port mapping (10000+ for Peers, 20000+ for Orderers).

## 📋 Prerequisites

Before you begin, ensure you have the following installed:
- **Go**: 1.21 or higher
- **Docker**: 24.0+
- **Docker Compose**: V2.20+
- **Bash**: 4.0+
- **Git** & **curl**: For bootstrapping tools

## 🚦 Quick Start

### 1. Generate the Network
Run the builder to configure your organizations and consensus:
```bash
go run main.go
```

### 2. Bootstrap Environment
Download the required Fabric binaries and pull Docker images:
```bash
./network/network.sh bootstrap
```

### 3. Launch Network
Bring the entire network up (Orderers, Peers, and Networking):
```bash
./network/network.sh up
```

### 4. Create a Channel
Initialize your first channel across all organizations:
```bash
./network/network.sh createChannel mychannel
```

## 📁 Directory Structure

```text
network/
├── bin/                  # Fabric binaries (after bootstrap)
├── channel-artifacts/    # Genesis blocks and channel transactions
├── compose/              # Centralized orchestration hub
├── configtx/             # Global channel profiles
├── organizations/        # Sovereign organizational data
│   └── {orgName}/
│       ├── compose/      # Org-specific docker manifests
│       ├── identities/   # Generated certificates (MSP/TLS)
│       └── identity-config/ # Cryptogen/CA configuration
├── scripts/              # Lifecycle automation scripts
└── network.sh            # Main entry point
```

## 🛠️ Lifecycle Commands

- `./network.sh up`: Generates certs and starts all containers.
- `./network.sh down`: Stops containers and cleans up artifacts (preserves config).
- `./network.sh bootstrap`: Downloads Fabric tools and images.
- `./network.sh createChannel <name>`: Automates channel creation and joining.

## 🧪 Tested With

- **Hyperledger Fabric**: 3.1.4
- **Fabric CA**: 1.5.12
- **Go**: 1.21+
- **Docker**: 24.0+
- **OS**: Arch Linux

---
Built with ❤️ for professional Hyperledger Fabric developers.

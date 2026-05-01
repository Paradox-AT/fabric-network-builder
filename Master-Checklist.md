# Final Master Checklist — Pre-Network Design Questions
## Hyperledger Fabric 3.1.4 — Enterprise Architecture

---

## TIER 1 — Foundation Decisions
> *Must be answered first. Everything else depends on them.*

---

### 1. Consortium & Governance

```
[ ] 1.1  How many organizations will participate at launch?
[ ] 1.2  Which orgs are founding members vs planned future joiners?
[ ] 1.3  Who has admin authority over the network?
         - Single org admin?
         - Shared majority governance?
         - Policy-based governance?
[ ] 1.4  What is the channel creation policy?
         - Any org can create?
         - Majority approval required?
[ ] 1.5  How will config updates and upgrades be approved?
         - Who signs channel config updates?
         - Majority or ALL orgs required?
[ ] 1.6  Will there be a dedicated orderer organization
         or will peer orgs run orderers?
[ ] 1.7  What is the offboarding process for an org that leaves?
[ ] 1.8  Is there a legal consortium agreement backing the network?
```

---

### 2. Consensus Type

```
[ ] 2.1  Do you trust ALL orderer operators not to act maliciously?
         - Yes → Raft (CFT)
         - No  → SmartBFT (BFT)

[ ] 2.2  What is your adversary / threat model?
         - Crash failures only          → Raft
         - Malicious or colluding nodes → SmartBFT

[ ] 2.3  Are orderers operated by a single org
         or multiple competing orgs?

[ ] 2.4  What fault tolerance level is required?
         Raft:     tolerates (n-1)/2 crash failures
                   3 nodes → tolerate 1
                   5 nodes → tolerate 2
                   7 nodes → tolerate 3
         SmartBFT: tolerates f byzantine nodes
                   needs 3f+1 nodes
                   4 nodes → tolerate 1
                   7 nodes → tolerate 2

[ ] 2.5  What is the acceptable block finality time?
         (Cross-region latency directly impacts this)

[ ] 2.6  Will orderers be geographically distributed?
         - Same datacenter?
         - Multi-AZ?
         - Multi-region?

[ ] 2.7  What is your quorum loss recovery plan?
         - Raft:     losing majority = network halt
         - SmartBFT: losing f+1 nodes = network halt
```

```
Decision Tree:
─────────────────────────────────────────────────
Are ALL orderer operators trusted?
          │
         YES                    NO
          │                      │
        RAFT                SmartBFT
      (CFT)               (BFT - 3.x only)
─────────────────────────────────────────────────
```

---

### 3. Channel Strategy

```
[ ] 3.1  How many channels are needed at launch?
[ ] 3.2  Which orgs participate in which channels?
[ ] 3.3  Is data isolation required between participant groups?
[ ] 3.4  Do you need Private Data Collections (PDC)
         within a channel for sub-org data isolation?
[ ] 3.5  What is the channel naming convention?
[ ] 3.6  What is the process for adding new channels post-launch?
[ ] 3.7  What data must never be visible across orgs
         even within the same channel?
```

---

### 4. Identity & Certificate Management

```
[ ] 4.1  What environment is this?
         - Development  → cryptogen acceptable
         - Staging      → Fabric CA required
         - Production   → Fabric CA required + HSM recommended

[ ] 4.2  Will you use Fabric CA or an External CA?
         External options:
         - HashiCorp Vault
         - AWS ACM (Amazon Certificate Manager)
         - CFSSL
         - Existing enterprise PKI

[ ] 4.3  CA hierarchy design:
         - One CA per org or shared CA?
         - Root CA only or Root CA + Intermediate CA?
         Recommended:
         Root CA → Intermediate CA → Identity CA → Identities

[ ] 4.4  Fabric CA database backend:
         - SQLite   → development only
         - PostgreSQL → staging and production
         - MariaDB    → staging and production

[ ] 4.5  Will Fabric CA be deployed in HA mode?
         (Multiple CA instances behind a load balancer)

[ ] 4.6  Is LDAP / Active Directory integration required
         for user identity?

[ ] 4.7  Will you use an HSM (Hardware Security Module)?
         - For peer signing keys?
         - For orderer signing keys?
         - For CA private keys?
         HSM options:
         - AWS CloudHSM
         - Azure Dedicated HSM
         - SoftHSM (development only)
         - PKCS#11 compatible device

[ ] 4.8  Certificate validity periods:
         - Root CA cert:          10-20 years
         - Intermediate CA cert:  5-10 years
         - Peer TLS cert:         1-2 years
         - Orderer TLS cert:      1-2 years
         - Admin cert:            1 year
         - Client / user cert:    90 days - 1 year

[ ] 4.9  What is the certificate rotation strategy?
         - Who triggers rotation?
         - What is the automation tooling?
         - Zero-downtime rotation procedure?

[ ] 4.10 What is the certificate revocation strategy?
         (Fabric has limited CRL support — plan carefully)

[ ] 4.11 What is the identity lifecycle process?
         - Onboarding new users and orgs?
         - Offboarding departed users and orgs?

[ ] 4.12 Certificate expiry monitoring:
         - Alerting at 90 / 60 / 30 days before expiry?
         - Who is responsible for acting on alerts?

[ ] 4.13 Are Node OUs enabled (client, peer, orderer, admin)?
         (Mandatory in modern Fabric for simplified policies)

[ ] 4.14 Will you use a dedicated CA for TLS certificates, 
         separate from the Identity/Enrollment CA?
```

---

## TIER 2 — Infrastructure Decisions
> *Answer after Tier 1 is resolved.*

---

### 5. Ordering Service

```
[ ] 5.1  How many orderer nodes?
         Raft     → odd numbers: 3, 5, 7
         SmartBFT → 3f+1:        4, 7, 10

[ ] 5.2  Which orgs host orderer nodes?

[ ] 5.3  Are orderers on dedicated infrastructure
         or shared with peers?

[ ] 5.4  Block configuration parameters:
         BatchTimeout:       2s      # max wait before cutting block
         MaxMessageCount:    500     # max transactions per block
         AbsoluteMaxBytes:   10 MB   # max block size
         PreferredMaxBytes:  2 MB    # target block size

[ ] 5.5  Raft tick configuration (if using Raft):
         TickInterval:          500ms
         ElectionTick:          10
         HeartbeatTick:         1
         MaxInflightBlocks:     5
         SnapshotIntervalSize:  20 MB

[ ] 5.6  Who has authority to add / remove orderer nodes
         from the consenter set?

[ ] 5.7  What is the orderer certificate rotation process?

[ ] 5.8  What is the disaster recovery plan
         if quorum is permanently lost?

[ ] 5.9  Block delivery model:
         NOTE: Gossip block dissemination is deprecated in 3.x
         Peers should be configured to receive blocks
         directly from orderer:
         peer.gossip.orgLeader: true
         peer.gossip.useLeaderElection: false
         peer.gossip.state.enabled: false
         peer.deliveryclient.blockGossipEnabled: false

[ ] 5.10 Who holds the Orderer Admin certificates required to 
         invoke the osnadmin API, and how are these highly 
         privileged keys secured?
```

---

### 6. Peer Topology

```
[ ] 6.1  How many peers per organization?

[ ] 6.2  What is the role of each peer?
         - Endorsing peer
         - Committing-only peer
         - Anchor peer (cross-org communication)

[ ] 6.3  Which peers will serve as anchor peers per org per channel?

[ ] 6.4  State database selection:
         LevelDB  → key-value only, faster, lower ops overhead
         CouchDB  → rich JSON queries, required for complex queries
                    Version: 3.4.2

[ ] 6.5  What is the expected ledger growth rate?
         (GB per month estimate)

[ ] 6.6  Will peers be co-located or on separate dedicated nodes?

[ ] 6.7  Gossip configuration:
         NOTE: Block gossip deprecated — configure direct delivery
         Gossip still used for:
         - Private data dissemination
         - Membership discovery
         MaxBlockCountToStore:     100
         MaxPropagationBurstSize:  10
         PropagateIterations:      1
         PullInterval:             4s
         PullPeerNum:              3

[ ] 6.8  What is the strategy for joining new peers? 
         - Sync from genesis block?
         - Join via State Snapshot (faster onboarding)?

[ ] 6.9  Will you dedicate specific peers purely as Gateway Peers 
         (handling client traffic but not endorsing), or will all 
         peers serve client Gateway requests?

[ ] 6.10 How will client applications load balance across the 
         Gateway peers? (e.g., k8s Ingress, HAProxy, AWS ALB)
```

---

### 7. Infrastructure & Deployment

```
[ ] 7.1  Deployment target:
         - Cloud only?
         - On-premise only?
         - Hybrid?

[ ] 7.2  Cloud provider(s):
         - AWS
         - GCP
         - Azure
         - OCI
         - Private datacenter

[ ] 7.3  Container orchestration:
         - Kubernetes (strongly recommended)
         - Which distribution?
           EKS / GKE / AKS / Rancher / bare metal k8s

[ ] 7.4  Will each org manage independent infrastructure
         or shared infrastructure?

[ ] 7.5  Network connectivity model between orgs:
         - Public internet + mTLS (most common)
         - VPN / private peering
         - SD-WAN

[ ] 7.6  Will you use a service mesh?
         - Istio
         - Linkerd
         - Consul Connect

[ ] 7.7  Hardware sizing per node type:
         ┌─────────────┬──────────┬───────────┬──────────┐
         │ Component   │ vCPU     │ RAM       │ Storage  │
         ├─────────────┼──────────┼───────────┼──────────┤
         │ Peer        │ 4-8      │ 16-32 GB  │ SSD      │
         │ Orderer     │ 4        │ 8-16 GB   │ SSD      │
         │ CA          │ 2        │ 4 GB      │ SSD      │
         │ CouchDB     │ 4-8      │ 16-32 GB  │ SSD      │
         └─────────────┴──────────┴───────────┴──────────┘

[ ] 7.8  Storage class for Kubernetes persistent volumes?
         - Fast SSD-backed storage class required
         - ReadWriteOnce access mode per component

[ ] 7.9  Disaster recovery site strategy:
         - Active-Active?
         - Active-Passive?
         - RTO and RPO targets?
```

---

## TIER 3 — Application Decisions
> *Answer in parallel with Tier 2.*

---

### 8. Chaincode

```
[ ] 8.1  Chaincode language:
         - Go       → recommended (performance, type safety)
         - Node.js  → acceptable
         - Java     → acceptable

[ ] 8.2  Chaincode execution model:
         - External Chaincode as a Service (CCaaS)
           (recommended for production — independent scaling,
            no peer restart on upgrade)
         - Which external builder image/binary will you configure 
           the peers to use? (e.g., standard k8s-builder)

[ ] 8.3  Chaincode libraries (Fabric 3.x MANDATORY):
         - github.com/hyperledger/fabric-chaincode-go/v2
         - github.com/hyperledger/fabric-protos-go-apiv2
         - github.com/hyperledger/fabric-contract-api-go/v2

[ ] 8.4  Endorsement policy per chaincode:
         AND('Org1.peer', 'Org2.peer')
         OR('Org1.peer', 'Org2.peer')
         OutOf(2, 'Org1.peer', 'Org2.peer', 'Org3.peer')

[ ] 8.5  How many chaincode packages at launch?

[ ] 8.6  Chaincode lifecycle approval policy:
         - Majority of orgs?
         - All orgs must approve?

[ ] 8.7  Chaincode upgrade governance:
         - Who approves upgrades?
         - What is the rollback plan?
         - How are breaking changes handled?

[ ] 8.8  Will you use Private Data Collections?
         - Which orgs share which collections?
         - What is the collection endorsement policy?
         - What is the BTL (Block To Live) for private data?

[ ] 8.9  Will you use GetMultipleStates() for batch reads?
         (New in Fabric 3.x — significant performance gain
          for chaincode reading many keys)

[ ] 8.10 Will any chaincodes utilize State-Based Endorsement 
         (where specific individual keys require different 
         signers than the default chaincode policy)?
```

---

### 9. Performance & Scalability

```
[ ] 9.1  Expected TPS:
         - Average TPS?
         - Peak TPS?

[ ] 9.2  Average transaction payload size?
         (KB per transaction)

[ ] 9.3  Latency SLAs:
         - End-to-end transaction finality target?
         - Read query response time target?

[ ] 9.4  Expected ledger size:
         - At 6 months?
         - At 1 year?
         - At 3 years?

[ ] 9.5  Will you implement off-chain storage for large payloads?
         - IPFS
         - AWS S3 / GCS / Azure Blob
         - External database
         Strategy: store hash on-chain, payload off-chain

[ ] 9.6  Read vs write ratio?
         (Helps size CouchDB and peer resources)

[ ] 9.7  MVCC conflict handling strategy under high concurrency?
         (Parallel writes to same key = transaction invalidation)

[ ] 9.8  Will you need parallel endorsement
         across multiple chaincodes?
```

---

### 10. Client Applications & Integration

```
[ ] 10.1  Client SDK:
          Fabric Gateway SDK ONLY (Fabric 3.x)
          Legacy SDKs are NOT supported

[ ] 10.2  Will you expose a REST API gateway
          in front of Fabric?
          - Custom Go gateway service?
          - GraphQL?
          - gRPC passthrough?

[ ] 10.3  Event listening strategy:
          - Block events
          - Chaincode events
          - Filtered block events
          - Who consumes events and what do they trigger?

[ ] 10.4  External system integrations:
          - ERP systems (SAP, Oracle)?
          - Databases (PostgreSQL, MongoDB)?
          - Message queues (Kafka, RabbitMQ, NATS)?
          - Payment systems?
          - External APIs?

[ ] 10.5  Wallet and identity management for client apps:
          - File-based wallet (dev only)
          - Database-backed wallet (production)
          - HSM-backed wallet (high security)

[ ] 10.6  Off-chain query strategy:
          - Will you maintain an indexed off-chain database
            for complex queries?
          - Event-driven sync from chaincode events
            to PostgreSQL / MongoDB / Elasticsearch?
```

---

## TIER 4 — Security & Compliance
> *Non-negotiable. Must be answered before go-live.*

---

### 11. Security

```
[ ] 11.1  Is mutual TLS (mTLS) enforced everywhere?
          - Peer to peer
          - Peer to orderer
          - Client to Gateway Peer (Port 7051)
          - Client to orderer (osnadmin)

[ ] 11.2  Secret management strategy:
          - HashiCorp Vault          (recommended)
          - AWS Secrets Manager
          - Azure Key Vault
          - Kubernetes Secrets       (NOT recommended for production)

[ ] 11.3  Network segmentation model:
          - Which ports are exposed externally?
          - Firewall rules between components?
          - Peer port:     7051 (grpc)
          - Orderer port:  7050 (grpc)
          - CA port:       7054 (https)
          - Operations:    9443 (https metrics/health)

[ ] 11.4  Container security strategy:
          - Image scanning (Trivy, Snyk, AWS ECR scanning)?
          - Non-root containers?
          - Read-only filesystems where possible?
          - Pod Security Standards on Kubernetes?

[ ] 11.5  Compliance frameworks that apply:
          - SOC 2 / ISO 27001
          - HIPAA
          - GDPR
          - PCI-DSS
          - Industry-specific regulations?

[ ] 11.6  Data residency requirements:
          - Which country / region must hold the data?
          - Any cross-border data transfer restrictions?

[ ] 11.7  Encryption at rest:
          - Ledger data encrypted at rest?
          - CA database encrypted at rest?
          - Kubernetes persistent volumes encrypted?

[ ] 11.8  Penetration testing plan:
          - Pre-launch pen test required?
          - Ongoing periodic testing?

[ ] 11.9  Vulnerability management:
          - Process for applying Fabric patch releases?
          - CVE monitoring for Fabric and dependencies?
          - Go vulnerability scanning (govulncheck)?

[ ] 11.10 Will the operations endpoint (9443) be secured via 
          mutual TLS, or isolated to a strict internal 
          management subnet?
```

---

### 12. Observability & Operations

```
[ ] 12.1  Metrics collection:
          - Prometheus + Grafana  (standard for Fabric)
          - Datadog
          - AWS CloudWatch
          Fabric exposes metrics at:
          operations endpoint: https://<host>:9443/metrics

[ ] 12.2  Log aggregation:
          - ELK Stack (Elasticsearch, Logstash, Kibana)
          - Grafana Loki + Promtail
          - Splunk
          - Datadog Logs

[ ] 12.3  Distributed tracing:
          - OpenTelemetry (recommended)
          - Jaeger
          - Zipkin

[ ] 12.4  Critical alerting rules needed:
          - Orderer leader change (Raft)
          - Peer disconnection from orderer
          - Block commit lag
          - Certificate expiry (90/60/30 days)
          - CouchDB storage thresholds
          - Pod restarts / OOMKilled events

[ ] 12.5  Backup strategy for:
          - Ledger data (peer filesystem)
          - CA data and private keys
          - Channel configuration blocks
          - Chaincode packages
          - Kubernetes etcd

[ ] 12.6  RTO / RPO targets:
          RTO (Recovery Time Objective):
               How fast must the network recover?
          RPO (Recovery Point Objective):
               How much data loss is acceptable?

[ ] 12.7  Day 2 operations ownership:
          - Certificate rotation
          - Fabric version upgrades
          - Orderer consenter set changes
          - Chaincode lifecycle management
          - Capacity planning and scaling

[ ] 12.8  On-call and incident response process:
          - Who is on-call?
          - What is the escalation path?
          - What runbooks exist?
          - SLA for incident response?
```

---

## Summary — Complete Question Map

```
┌─────────────────────────────────────────────────────────────────┐
│              FABRIC 3.1.4 PRE-NETWORK QUESTION MAP              │
├────────┬──────────────────────────────────┬─────────────────────┤
│ TIER   │ DOMAIN                           │ QUESTIONS           │
├────────┼──────────────────────────────────┼─────────────────────┤
│        │ 1. Consortium & Governance       │ 1.1  → 1.8          │
│ TIER 1 │ 2. Consensus Type                │ 2.1  → 2.7          │
│        │ 3. Channel Strategy              │ 3.1  → 3.7          │
│        │ 4. Identity & Cert Management    │ 4.1  → 4.14         │
├────────┼──────────────────────────────────┼─────────────────────┤
│        │ 5. Ordering Service              │ 5.1  → 5.10         │
│ TIER 2 │ 6. Peer Topology                 │ 6.1  → 6.10         │
│        │ 7. Infrastructure & Deployment   │ 7.1  → 7.9          │
├────────┼──────────────────────────────────┼─────────────────────┤
│        │ 8.  Chaincode                    │ 8.1  → 8.10         │
│ TIER 3 │ 9.  Performance & Scalability    │ 9.1  → 9.8          │
│        │ 10. Client Apps & Integration    │ 10.1 → 10.6         │
├────────┼──────────────────────────────────┼─────────────────────┤
│ TIER 4 │ 11. Security                     │ 11.1 → 11.10        │
│        │ 12. Observability & Operations   │ 12.1 → 12.8         │
└────────┴──────────────────────────────────┴─────────────────────┘
  Total: 12 domains  |  ~105 questions  |  4 tiers
```

---

## Minimum to Start Architecture Design

```
Answer these 10 first and we can begin immediately:

  1.1  →  How many organizations?
  2.1  →  Trusted or untrusted orderer operators?
  2.4  →  Required fault tolerance level?
  3.1  →  How many channels?
  4.1  →  Development, staging, or production?
  5.1  →  How many orderer nodes?
  6.1  →  How many peers per org?
  6.4  →  LevelDB or CouchDB?
  7.1  →  Cloud, on-prem, or Kubernetes?
  9.1  →  Expected TPS and payload size?
```

---

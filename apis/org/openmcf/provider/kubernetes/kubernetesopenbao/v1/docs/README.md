# OpenBao on Kubernetes - Technical Research

## Overview

OpenBao is an open-source secrets management solution that is a community-driven fork of HashiCorp Vault, created after Vault's license change to BSL (Business Source License). OpenBao is developed under the OpenSSF (Open Source Security Foundation) umbrella and maintains API compatibility with Vault while being released under the MPL-2.0 license.

## Architecture

### Core Components

1. **OpenBao Server**: The main secrets management engine
2. **Storage Backend**: Persistent storage for encrypted secrets
3. **Agent Injector**: Kubernetes admission webhook for automatic secret injection
4. **CSI Provider**: Secrets Store CSI driver integration (optional)

### Deployment Modes

#### Standalone Mode
- Single server instance
- File-based storage backend
- Suitable for development and small deployments
- No built-in redundancy

#### High Availability (HA) Mode
- Multiple server instances (3+ recommended)
- Raft integrated storage for consensus
- Automatic leader election
- Data replication across all nodes

## Helm Chart Details

### Repository Information
- **Chart Repository**: https://openbao.github.io/openbao-helm
- **OCI Registry**: oci://ghcr.io/openbao/charts/openbao
- **Chart Version**: 0.23.3 (as of implementation)
- **App Version**: v2.4.4

### Key Helm Values

#### Server Configuration
```yaml
server:
  enabled: true
  image:
    registry: "quay.io"
    repository: "openbao/openbao"
    tag: ""  # defaults to appVersion
  standalone:
    enabled: "-"  # enabled when HA is disabled
    config: |
      ui = true
      listener "tcp" {
        tls_disable = 1
        address = "[::]:8200"
        cluster_address = "[::]:8201"
      }
      storage "file" {
        path = "/openbao/data"
      }
  ha:
    enabled: false
    replicas: 3
    raft:
      enabled: false
      config: |
        ui = true
        listener "tcp" {
          tls_disable = 1
          address = "[::]:8200"
          cluster_address = "[::]:8201"
        }
        storage "raft" {
          path = "/openbao/data"
        }
        service_registration "kubernetes" {}
  dataStorage:
    enabled: true
    size: 10Gi
    storageClass: null
    accessMode: ReadWriteOnce
```

#### Agent Injector Configuration
```yaml
injector:
  enabled: "-"  # follows global.enabled
  replicas: 1
  image:
    registry: "docker.io"
    repository: "hashicorp/vault-k8s"
    tag: "1.7.2"
  agentImage:
    registry: "quay.io"
    repository: "openbao/openbao"
```

#### UI Configuration
```yaml
ui:
  enabled: false
  serviceType: "ClusterIP"
  externalPort: 8200
```

## Network Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 8200 | TCP | API and UI access |
| 8201 | TCP | Cluster communication (HA mode) |

## Storage Backends

### File Backend (Standalone)
- Simple file-based storage
- Single node only
- Good for development

### Raft Backend (HA)
- Built-in consensus protocol
- Leader election
- Data replication
- Recommended for production

### External Backends (via config override)
- Consul
- PostgreSQL
- MySQL
- DynamoDB
- And more

## Security Considerations

### Seal/Unseal Architecture

OpenBao encrypts all stored data with a randomly generated encryption key. This encryption key is itself protected by a **master key** that is never stored in plaintext. How the master key is protected defines the seal type.

#### Shamir Seal (Default)

The default method uses **Shamir's Secret Sharing** to split the master key into N key shares with a threshold of T shares required to reconstruct it (commonly 5 shares, threshold 3). On every pod start, T key shares must be provided via `bao operator unseal` before the server can serve requests.

**Problem in Kubernetes:** Every pod restart -- rolling update, node failure, OOM kill, or preemption -- produces a sealed instance that cannot serve traffic until a human provides unseal keys. For HA deployments with 3+ replicas, a single node preemption event requires manual intervention bounded by human response time.

#### Auto-Unseal

Auto-unseal delegates master key protection to an external KMS. The master key is encrypted (wrapped) by the KMS key and stored alongside the data. On startup, OpenBao calls the KMS to decrypt (unwrap) the master key automatically -- no human intervention required.

**Seal stanza**: Auto-unseal is configured via an HCL `seal` block in the server config. The `KubernetesOpenBao` component generates this block from the `auto_unseal` spec field and appends it to the Helm chart's server configuration.

**Migration path**: Existing deployments using Shamir can migrate to auto-unseal by adding the seal stanza and running `bao operator unseal -migrate` with the existing Shamir keys.

#### GCP Cloud KMS

- **Seal type**: `gcpckms`
- **Key requirement**: Symmetric encrypt/decrypt key in a Cloud KMS keyring
- **IAM**: The service account needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key
- **Authentication on GKE**: Use Workload Identity -- annotate the Kubernetes ServiceAccount with `iam.gke.io/gcp-service-account` pointing to a GCP service account that has a Workload Identity binding for the OpenBao namespace and SA. No credential files needed.
- **Authentication elsewhere**: Application Default Credentials or a JSON key file mounted as a volume

#### AWS KMS

- **Seal type**: `awskms`
- **Key requirement**: Symmetric KMS key (default key spec `SYMMETRIC_DEFAULT`)
- **IAM**: The IAM entity needs `kms:Encrypt` and `kms:Decrypt` on the key ARN
- **Authentication on EKS**: Use IRSA (IAM Roles for Service Accounts) -- associate an IAM role with the Kubernetes service account
- **Authentication elsewhere**: Static credentials via a Kubernetes secret with `access-key` and `secret-key` data keys, or EC2 instance profile

#### Azure Key Vault

- **Seal type**: `azurekeyvault`
- **Key requirement**: RSA or EC key in an Azure Key Vault with `wrapKey` and `unwrapKey` permissions
- **Authentication on AKS**: Azure Managed Identity (either system-assigned or user-assigned)
- **Authentication elsewhere**: Service principal credentials via a Kubernetes secret with `client-id` and `client-secret` data keys

#### Transit Seal

- **Seal type**: `transit`
- **Key requirement**: A Transit secrets engine key on another Vault/OpenBao instance (the "central" instance)
- **Dependency**: The satellite instance depends on the central instance being available and unsealed
- **Authentication**: A Vault/OpenBao token stored in a Kubernetes secret with a `token` data key. The token must have a policy granting `update` on `transit/encrypt/<key>` and `transit/decrypt/<key>`
- **Use case**: Multi-cluster setups where a single central Vault instance manages unseal for satellite instances

### Auto-Unseal Security Considerations

- **KMS key rotation**: Cloud KMS keys should have automatic rotation enabled. OpenBao re-wraps the master key on the next seal/unseal cycle after rotation.
- **Key access audit**: Enable Cloud Audit Logs (GCP), CloudTrail (AWS), or Diagnostic Logs (Azure) on the KMS key to track every encrypt/decrypt call.
- **Blast radius**: Deleting or disabling the KMS key makes all sealed OpenBao data unrecoverable. Use key destruction protection and IAM deny policies on deletion.
- **Network path**: The OpenBao pod must have network access to the KMS API endpoint. For Transit seal, the pod must reach the central Vault instance on port 8200.

### TLS Configuration
- TLS disabled by default (`global.tlsDisable: true`)
- Can be enabled for production deployments
- Supports custom certificates

### Authentication Methods
- Kubernetes Service Account
- Token-based
- LDAP
- OIDC
- And more

## Initialization Process

After deployment, OpenBao must be initialized:

```bash
# Initialize with 5 key shares, 3 required to unseal
bao operator init -key-shares=5 -key-threshold=3

# Unseal (repeat 3 times with different keys)
bao operator unseal <key>

# Login with root token
bao login <root-token>
```

## Kubernetes Integration

### Service Account Auth
```bash
# Enable Kubernetes auth
bao auth enable kubernetes

# Configure auth
bao write auth/kubernetes/config \
    kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"

# Create role
bao write auth/kubernetes/role/app \
    bound_service_account_names=app-sa \
    bound_service_account_namespaces=default \
    policies=app-policy \
    ttl=24h
```

### Agent Injection Annotations
```yaml
annotations:
  vault.hashicorp.com/agent-inject: "true"
  vault.hashicorp.com/role: "app"
  vault.hashicorp.com/agent-inject-secret-config: "secret/data/app/config"
```

## Monitoring and Telemetry

### Prometheus Integration
```yaml
serverTelemetry:
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
  prometheusRules:
    enabled: true
```

### Grafana Dashboard
The Helm chart includes a pre-built Grafana dashboard for monitoring OpenBao metrics.

## Best Practices

### Production Deployment
1. **Use HA Mode**: Deploy at least 3 replicas for fault tolerance
2. **Enable TLS**: Secure all communications with TLS
3. **Auto-Unseal**: Configure auto-unseal for operational simplicity
4. **Audit Logging**: Enable audit devices for compliance
5. **Backup Strategy**: Regular snapshots of Raft storage

### Resource Sizing
| Deployment Size | CPU Request | Memory Request | Storage |
|-----------------|-------------|----------------|---------|
| Development | 100m | 128Mi | 1Gi |
| Small | 250m | 256Mi | 10Gi |
| Medium | 500m | 512Mi | 50Gi |
| Large | 1000m | 1Gi | 100Gi |

## References

- [OpenBao Documentation](https://openbao.org/docs/)
- [OpenBao Helm Chart](https://github.com/openbao/openbao-helm)
- [OpenBao GitHub Repository](https://github.com/openbao/openbao)
- [Kubernetes Auth Method](https://openbao.org/docs/auth/kubernetes)
- [Agent Injector](https://openbao.org/docs/platform/k8s/injector)

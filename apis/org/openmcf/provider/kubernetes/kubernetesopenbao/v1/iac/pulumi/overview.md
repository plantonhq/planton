# KubernetesOpenBao Pulumi Module - Architecture Overview

## Module Structure

```
iac/pulumi/
├── main.go           # Pulumi entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and dependency management
├── debug.sh          # Debugging script
├── README.md         # Module documentation
├── overview.md       # This file
└── module/
    ├── main.go       # Resource orchestration
    ├── locals.go     # Local variables and computed values
    ├── vars.go       # Helm chart configuration
    ├── outputs.go    # Output constants
    ├── namespace.go  # Namespace creation
    └── helm_chart.go # Helm chart installation, seal config, workload identity
```

## Resource Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    KubernetesOpenBaoStackInput                   │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │ ProviderConfig  │  │            Target                    │  │
│  │ (kubeconfig)    │  │  ┌─────────────────────────────┐    │  │
│  └────────┬────────┘  │  │    KubernetesOpenBaoSpec    │    │  │
│           │           │  │  - namespace                 │    │  │
│           │           │  │  - server_container          │    │  │
│           │           │  │  - high_availability         │    │  │
│           │           │  │  - ingress                   │    │  │
│           │           │  │  - injector                  │    │  │
│           │           │  │  - ui_enabled                │    │  │
│           │           │  │  - auto_unseal                │    │  │
│           │           │  └─────────────────────────────┘    │  │
│           │           └─────────────────────────────────────┘  │
└───────────┼────────────────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────────┐
│                      Module Orchestration                          │
│                        (module/main.go)                            │
│                                                                    │
│  1. initializeLocals() ─────────► Compute namespace, labels,       │
│                                   service names, endpoints         │
│                                                                    │
│  2. pulumikubernetesprovider ───► Create Kubernetes provider       │
│                                                                    │
│  3. namespace() ────────────────► Create namespace (if enabled)    │
│                                                                    │
│  4. helmChart() ────────────────► Deploy OpenBao Helm chart        │
│                                                                    │
└───────────────────────────────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────────┐
│                     Kubernetes Resources                           │
│                                                                    │
│  ┌─────────────┐  ┌─────────────────────────────────────────────┐ │
│  │  Namespace  │  │              Helm Release                    │ │
│  │ (optional)  │  │  ┌───────────────────────────────────────┐  │ │
│  └─────────────┘  │  │ OpenBao Server (StatefulSet)          │  │ │
│                   │  │ - Standalone OR HA mode                │  │ │
│                   │  │ - PersistentVolumeClaims               │  │ │
│                   │  │ - ConfigMaps                           │  │ │
│                   │  │ - Services                             │  │ │
│                   │  └───────────────────────────────────────┘  │ │
│                   │  ┌───────────────────────────────────────┐  │ │
│                   │  │ Agent Injector (Deployment, optional) │  │ │
│                   │  │ - MutatingWebhookConfiguration        │  │ │
│                   │  │ - ServiceAccount/RBAC                 │  │ │
│                   │  └───────────────────────────────────────┘  │ │
│                   │  ┌───────────────────────────────────────┐  │ │
│                   │  │ Ingress (optional)                    │  │ │
│                   │  │ - TLS termination                     │  │ │
│                   │  │ - External hostname                   │  │ │
│                   │  └───────────────────────────────────────┘  │ │
│                   │  ┌───────────────────────────────────────┐  │ │
│                   │  │ UI Service (optional)                 │  │ │
│                   │  └───────────────────────────────────────┘  │ │
│                   └─────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────────┐
│                        Stack Outputs                               │
│                                                                    │
│  - namespace              - kube_endpoint                         │
│  - service                - external_hostname                     │
│  - port_forward_command   - api_address                          │
│  - cluster_address        - ha_enabled                           │
└───────────────────────────────────────────────────────────────────┘
```

## Deployment Modes

### Standalone Mode
```
┌────────────────────────────────────────┐
│          OpenBao Server Pod            │
│  ┌─────────────────────────────────┐  │
│  │         openbao container       │  │
│  │  - API: 8200                    │  │
│  │  - Cluster: 8201                │  │
│  └────────────┬────────────────────┘  │
│               │                        │
│  ┌────────────▼────────────────────┐  │
│  │    PersistentVolumeClaim        │  │
│  │    (file storage backend)       │  │
│  └─────────────────────────────────┘  │
└────────────────────────────────────────┘
```

### High Availability Mode (Raft)
```
┌────────────────────────────────────────────────────────────────────┐
│                       Raft Cluster (3+ replicas)                   │
│                                                                    │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐         │
│  │ openbao-0   │◄───►│ openbao-1   │◄───►│ openbao-2   │         │
│  │  (Leader)   │     │ (Standby)   │     │ (Standby)   │         │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘         │
│         │                   │                   │                 │
│  ┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐         │
│  │    PVC-0    │     │    PVC-1    │     │    PVC-2    │         │
│  │ (Raft data) │     │ (Raft data) │     │ (Raft data) │         │
│  └─────────────┘     └─────────────┘     └─────────────┘         │
│                                                                    │
│  Services:                                                         │
│  - openbao (ClusterIP): Routes to active leader                   │
│  - openbao-active: Selects active node only                       │
│  - openbao-standby: Selects standby nodes                         │
│  - openbao-internal (Headless): Pod-to-pod communication          │
└────────────────────────────────────────────────────────────────────┘
```

## Helm Values Mapping

| Spec Field | Helm Value | Description |
|------------|------------|-------------|
| `server_container.replicas` | `server.ha.replicas` | Number of server replicas (HA mode) |
| `server_container.resources` | `server.resources` | CPU/memory allocation |
| `server_container.data_storage_size` | `server.dataStorage.size` | PVC size |
| `high_availability.enabled` | `server.ha.enabled` | Enable HA mode |
| `high_availability.replicas` | `server.ha.replicas` | HA replica count |
| `ingress.enabled` | `server.ingress.enabled` | Enable ingress |
| `ingress.hostname` | `server.ingress.hosts[].host` | Ingress hostname |
| `ui_enabled` | `ui.enabled` | Enable UI service |
| `injector.enabled` | `injector.enabled` | Enable agent injector |
| `tls_enabled` | `global.tlsDisable` (inverted) | TLS configuration |
| `auto_unseal.*` | `server.standalone.config` / `server.ha.raft.config` | Appends HCL `seal` stanza to server config |
| `auto_unseal.gcp_kms.workload_identity_service_account` | `server.serviceAccount.annotations` | Adds `iam.gke.io/gcp-service-account` annotation |

## Key Implementation Details

### Auto-Unseal (`helm_chart.go`)

- `sealConfigHcl(spec)` -- Type-switches on the `auto_unseal.seal` oneof and returns the matching HCL `seal` stanza (`gcpckms`, `awskms`, `azurekeyvault`, or `transit`). Returns empty string when auto-unseal is not configured. Uses `GetValue()` on `StringValueOrRef` fields.
- `workloadIdentityServiceAccount(spec)` -- Extracts the GCP service account email from the GCP KMS seal config for the `iam.gke.io/gcp-service-account` annotation. Returns empty string when GCP KMS is not configured.
- The seal stanza is appended to both the standalone and HA server config strings.
- The service account annotation is injected into Helm values when Workload Identity is configured.

### Labels
All resources are tagged with standard OpenMCF labels:
- `planton-resource: "true"`
- `planton-resource-name: <name>`
- `planton-resource-kind: KubernetesOpenBao`
- `planton-organization: <org>` (if set)
- `planton-environment: <env>` (if set)

### Service Discovery
- Internal: `<name>.<namespace>.svc.cluster.local:8200`
- Active (HA): `<name>-active.<namespace>.svc.cluster.local:8200`
- Standby (HA): `<name>-standby.<namespace>.svc.cluster.local:8200`

### Port Forwarding
For local development access:
```bash
kubectl port-forward -n <namespace> service/<name> 8200:8200
```

## Post-Deployment Requirements

After deployment, OpenBao requires initialization:
1. Initialize: `bao operator init`
2. Unseal (Shamir only): `bao operator unseal` (repeat with threshold keys). If `auto_unseal` is configured, this step is handled automatically by the KMS provider on every pod startup.
3. Login: `bao login <root-token>`
4. Configure auth methods and policies

# Kubernetes External Secrets

Automate secret management by syncing secrets from cloud providers (GCP, AWS, Azure) into Kubernetes using External Secrets Operator.

## Overview

**KubernetesExternalSecrets** is a Planton component that deploys and configures the [External Secrets Operator](https://external-secrets.io/) on Kubernetes clusters. ESO continuously synchronizes secrets from external secret stores (Google Cloud Secret Manager, AWS Secrets Manager, Azure Key Vault) into Kubernetes Secret objects.

Instead of manually managing Kubernetes secrets or storing them in Git, ESO treats cloud secret stores as the single source of truth and automatically keeps Kubernetes secrets in sync.

## Why We Created This

Traditional Kubernetes secret management is fraught with problems:

- **Manual secret lifecycle**: Creating, updating, and rotating secrets requires manual kubectl commands
- **Git-based secrets**: Storing encrypted secrets in Git creates versioning and rotation challenges
- **Credential sprawl**: Static service account keys scattered across clusters
- **No dynamic rotation**: Secrets can't be automatically rotated with short TTLs
- **Auditability gaps**: Hard to track who accessed which secrets when

Our KubernetesExternalSecrets module solves these problems by providing:

1. **Automatic synchronization**: Cloud secrets sync to Kubernetes automatically
2. **Cloud-native identity**: Uses Workload Identity, IRSA, and Managed Identity (no static credentials)
3. **Single source of truth**: Cloud secret managers are the authoritative source
4. **Automatic rotation**: Secrets update when changed in the cloud provider
5. **Multi-cloud support**: Consistent interface for GCP, AWS, and Azure
6. **Production-ready defaults**: Optimal polling intervals, resource limits, and security settings

## Supported Providers

- **GKE (Google Kubernetes Engine)** with Google Cloud Secret Manager and Workload Identity
- **EKS (Amazon Elastic Kubernetes Service)** with AWS Secrets Manager and IRSA
- **AKS (Azure Kubernetes Service)** with Azure Key Vault and Managed Identity

## Key Features

### 🔄 Continuous Secret Synchronization
- Automatically creates Kubernetes Secrets from cloud secret stores
- Polls for changes and updates Secrets automatically
- Configurable poll interval to balance freshness vs API costs

### 🔐 Secure by Default
- Cloud-native authentication (Workload Identity, IRSA, Managed Identity)
- No long-lived credentials stored in cluster
- Encryption-at-rest via KMS integration
- Fine-grained IAM permissions for secret access

### 🌍 Multi-Cloud Ready
- Single API for GCP, AWS, and Azure secret managers
- Provider-specific optimizations built-in
- Consistent configuration across clouds

### 📦 Production-Ready
- Deployed via Helm with battle-tested operator
- Configurable CPU/memory resources
- Proper RBAC and ServiceAccount setup
- Monitoring and observability built-in

### ⚙️ Flexible Configuration
- Configurable poll intervals (balance API costs vs freshness)
- Custom resource limits for operator pods
- Regional configuration for AWS
- Support for existing IAM roles/managed identities

## Quick Start

See [examples.md](examples.md) for complete YAML manifests for each cloud provider.

### Example: External Secrets for GKE

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: external-secrets-prod
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: external-secrets
  poll_interval_seconds: 30
  container:
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 50m
        memory: 100Mi
  gke:
    project_id:
      value: my-gcp-project
    gsa_email: external-secrets@my-project.iam.gserviceaccount.com
```

This creates an External Secrets Operator deployment that:
- Uses Workload Identity to authenticate with Google Cloud Secret Manager
- Polls for secret changes every 30 seconds
- Automatically creates/updates Kubernetes Secrets from cloud secrets
- Has proper resource limits and requests

## How It Works

1. **Deploy**: Apply a KubernetesExternalSecrets manifest to your cluster
2. **Sync**: ESO controller authenticates with cloud provider using workload identity
3. **Watch**: Controller polls secret store for changes based on ExternalSecret/SecretStore CRDs
4. **Create**: When secrets exist in cloud, ESO creates corresponding Kubernetes Secret objects
5. **Update**: When cloud secrets change, Kubernetes Secrets are updated automatically

## Use Cases

- **Application secrets**: Database passwords, API keys, certificates
- **Dynamic rotation**: Short-lived credentials that rotate frequently
- **Compliance**: Centralized secret management with audit trails
- **Multi-environment**: Different secrets for dev/staging/prod from same codebase
- **Migration from static secrets**: Gradual migration path from git-stored secrets

## Benefits

### For Platform Teams
- **Standardized deployment**: Same interface across all clusters and clouds
- **Security compliance**: No static credentials, all authentication via cloud IAM
- **Audit trail**: All secret access logged through cloud provider audit logs
- **Operational simplicity**: No Vault to maintain, no custom sidecars to write

### For Application Teams
- **Zero secret management**: Secrets appear as standard Kubernetes Secrets
- **Fast deployments**: No waiting for manual secret creation
- **Self-service**: Create secrets in cloud provider, they appear in cluster automatically
- **GitOps-friendly**: Application manifests don't contain secrets

### For Everyone
- **Better security**: Cloud secret managers > Git-encrypted secrets
- **Automatic rotation**: Secrets update without deployment
- **Cost-effective**: Pay only for cloud secret manager API calls
- **Disaster recovery**: Secrets backed by cloud provider's durability guarantees

## Configuration Reference

| Field | Description | Default |
|-------|-------------|---------|
| `poll_interval_seconds` | How often to poll cloud provider for changes | `10` |
| `container.resources` | CPU/memory limits for ESO controller | See spec |
| `provider_config` | Cloud provider configuration (gke, eks, or aks) | Required |

### GKE Configuration

| Field | Description |
|-------|-------------|
| `project_id` | GCP project ID containing the secrets |
| `gsa_email` | Google Service Account email for Workload Identity |

### EKS Configuration

| Field | Description |
|-------|-------------|
| `region` | AWS region for Secrets Manager (defaults to cluster region) |
| `irsa_role_arn_override` | Optional IAM role ARN for IRSA (auto-created if empty) |

### AKS Configuration

| Field | Description |
|-------|-------------|
| `key_vault_resource_id` | Azure Key Vault resource ID |
| `managed_identity_client_id` | Azure Managed Identity client ID |

## Namespace Management

The component provides flexible namespace management to support different deployment patterns:

### Create Namespace (Default Behavior)

When `create_namespace: true` (or omitted), the component creates and manages the namespace:

```yaml
spec:
  namespace:
    value: external-secrets
  create_namespace: true  # Component creates the namespace
  # ...
```

**Use when:**
- Deploying to a new cluster where namespace doesn't exist
- You want the component to manage namespace lifecycle
- Single-component deployment (namespace is exclusive to this component)

### Use Existing Namespace

When `create_namespace: false`, the component uses an existing namespace:

```yaml
spec:
  namespace:
    value: platform-services  # Must already exist
  create_namespace: false  # Use existing namespace
  # ...
```

**Use when:**
- Deploying multiple components to the same namespace
- Namespace is managed separately (e.g., by platform team)
- Namespace has specific labels, annotations, or RBAC policies
- GitOps workflow where namespace is created via separate manifest

**Important**: When using `create_namespace: false`, ensure the namespace exists before deploying this component. The deployment will fail if the namespace doesn't exist.

## Prerequisites

### For GKE
- GKE cluster with Workload Identity enabled
- Google Cloud Secret Manager API enabled
- GCP service account with `secretmanager.secretAccessor` role
- IAM binding for Workload Identity (created by module)

### For EKS
- EKS cluster with OIDC provider configured
- AWS Secrets Manager in the same region
- IAM role with SecretsManager permissions (created by module or provided)

### For AKS
- AKS cluster with Azure AD Workload Identity enabled
- Azure Key Vault created
- User-assigned Managed Identity with "Key Vault Secrets User" role

## Using External Secrets After Deployment

Once deployed, create SecretStore and ExternalSecret resources to sync secrets:

### Create a SecretStore

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gcpsm-secret-store
  namespace: default
spec:
  provider:
    gcpsm:
      projectID: my-gcp-project
      auth:
        workloadIdentity:
          clusterLocation: us-central1
          clusterName: my-cluster
          serviceAccountRef:
            name: external-secrets-sa
```

### Create an ExternalSecret

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
  namespace: default
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: gcpsm-secret-store
    kind: SecretStore
  target:
    name: app-secrets  # Kubernetes Secret name
    creationPolicy: Owner
  data:
  - secretKey: database-password
    remoteRef:
      key: prod-db-password  # Cloud secret name
```

This creates a Kubernetes Secret named `app-secrets` with data from `prod-db-password` in Google Cloud Secret Manager.

## Documentation

- **[Research Documentation](docs/README.md)**: Deep dive into ESO landscape, alternatives, and best practices
- **[Examples](examples.md)**: Complete YAML manifests for GKE, EKS, and AKS
- **[Pulumi Module](iac/pulumi/README.md)**: Using the Pulumi implementation directly
- **[Terraform Module](iac/tf/README.md)**: Using the Terraform implementation directly

## Next Steps

1. Review the [examples](examples.md) for your cloud provider
2. Read the [research documentation](docs/README.md) to understand architecture patterns
3. Create a KubernetesExternalSecrets manifest
4. Deploy it via `planton deploy`
5. Create SecretStore and ExternalSecret CRDs to sync your secrets

## Support

For issues, questions, or contributions, see the main [Planton repository](https://github.com/plantonhq/planton).


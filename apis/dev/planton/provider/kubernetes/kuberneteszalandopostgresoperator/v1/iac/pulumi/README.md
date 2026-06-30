# Zalando Postgres Operator - Pulumi Module

This Pulumi module deploys the [Zalando Postgres Operator](https://github.com/zalando/postgres-operator) on Kubernetes clusters, providing automated PostgreSQL database management with backup support for Cloudflare R2.

## Overview

The module provides a streamlined interface for deploying and configuring the Zalando Postgres Operator, handling:
- Helm chart deployment with configurable resources
- Automatic namespace creation with Planton labels
- Optional backup configuration with Cloudflare R2 integration
- Label inheritance for all managed PostgreSQL databases

## Prerequisites

| Requirement | Purpose |
|------------|---------|
| **Pulumi CLI** | Infrastructure-as-Code deployment |
| **Kubernetes Cluster** | Target for operator deployment |
| **kubectl** | Verification and debugging |
| **Go** | Module development/debugging |
| **Cloudflare R2** (optional) | Backup storage |

## Quick Start

### Deploy with Planton CLI

```bash
# Create manifest file
cat > operator.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
EOF

# Deploy
planton pulumi up --manifest operator.yaml
```

### Deploy Standalone (for development)

```bash
cd iac/pulumi

# Initialize Pulumi stack
pulumi stack init dev

# Set configuration
pulumi config set kubernetesProvider:kubeconfig ~/.kube/config

# Deploy
pulumi up
```

## Module Structure

```
iac/pulumi/
├── main.go                 # Pulumi entry point (CLI integration)
├── Pulumi.yaml             # Pulumi project configuration
├── Makefile                # Build and test automation
├── debug.sh                # Local debugging script
├── README.md               # This file
├── overview.md             # Architecture documentation
└── module/
    ├── main.go             # Resources function (entry point)
    ├── locals.go           # Shared data transformations
    ├── outputs.go          # Output constant definitions
    ├── vars.go             # Helm chart variables
    ├── postgres_operator.go # Operator deployment logic
    └── backup_config.go    # Backup Secret/ConfigMap creation
```

## Key Files

### module/main.go

Entry point for the Pulumi module. Orchestrates the deployment:

1. Initialize locals (labels, metadata)
2. Create Kubernetes provider
3. Deploy Postgres Operator (Helm chart + backup resources)
4. Export outputs

### module/postgres_operator.go

Handles the core operator deployment:

- Creates `postgres-operator` namespace
- Conditionally creates backup resources (if configured)
- Deploys Helm chart with inherited labels configuration
- Exports namespace output

### module/backup_config.go

Manages backup configuration when `backup_config` is provided:

- Creates Kubernetes Secret with R2 credentials (`r2-postgres-backup-credentials`)
- Creates ConfigMap with WAL-G environment variables (`postgres-pod-backup-config`)
- Constructs R2 endpoint URL and S3 prefix
- Returns ConfigMap reference for operator configuration

### module/locals.go

Transforms input protobuf messages into reusable local values:

- Computes Kubernetes labels (resource, org, env, kind, id)
- Stores references to input objects
- Minimal transformation layer (follows "thin locals" pattern)

### module/vars.go

Constants and configuration:

- Helm chart name: `postgres-operator`
- Helm repository: `https://opensource.zalando.com/postgres-operator/charts/postgres-operator`
- Chart version: `1.12.2`
- Namespace: `postgres-operator` (fixed)

## Input Schema

The module accepts `KubernetesZalandoPostgresOperatorStackInput` with:

### Required Fields

```protobuf
message KubernetesZalandoPostgresOperatorSpec {
  container {
    resources {
      requests { cpu, memory }
      limits { cpu, memory }
    }
  }
}
```

### Optional Fields

```protobuf
message KubernetesZalandoPostgresOperatorSpec {
  backup_config {
    bucket                     // literal name or value_from ref to an S3-compatible bucket
    object_prefix              // base path; module appends the per-cluster/per-version suffix
    schedule                   // e.g., "0 2 * * *"
    enable_wal_g_backup        // defaults to true
    enable_wal_g_restore       // defaults to true
    enable_clone_wal_g_restore // defaults to true
    credentials {
      cloudflare_account_id
      access_key_id
      secret_access_key
    }
  }
}
```

## Outputs

| Export Key | Description |
|------------|-------------|
| `namespace` | Operator namespace (`postgres-operator`) |

Additional outputs defined in `stack_outputs.proto` but not yet implemented:
- `service`: Operator service name
- `port_forward_command`: kubectl port-forward command
- `kube_endpoint`: Internal cluster endpoint
- `ingress_endpoint`: Public endpoint (N/A for this operator)

## Configuration

### Resource Limits

The operator container resources are configurable via spec:

```yaml
spec:
  container:
    resources:
      requests:
        cpu: 100m      # Minimum guaranteed
        memory: 256Mi
      limits:
        cpu: 2000m     # Maximum allowed
        memory: 2Gi
```

### Backup Configuration

When backup_config is provided, the module:

1. Creates `r2-postgres-backup-credentials` Secret
2. Creates `postgres-pod-backup-config` ConfigMap
3. Configures operator Helm chart with `pod_environment_configmap`

All PostgreSQL databases created by the operator automatically inherit the backup configuration.

### Helm Values

The module sets these Helm values:

```yaml
configKubernetes:
  inherited_labels:
    - resource
    - organization
    - environment
    - resource_kind
    - resource_id
  pod_environment_configmap: "postgres-operator/postgres-pod-backup-config"  # if backup configured
```

## Development

### Local Testing

Use the provided `debug.sh` script:

```bash
cd iac/pulumi
./debug.sh
```

This script:
1. Reads `../hack/manifest.yaml` as input
2. Converts YAML to JSON
3. Runs Pulumi with the test manifest

### Makefile Commands

```bash
# Build the module
make build

# Run tests
make test

# Deploy with debug
make debug

# Clean build artifacts
make clean
```

### Manual Testing

```bash
# Set up test manifest
cat > hack/manifest.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: test-operator
spec:
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 500m
        memory: 512Mi
EOF

# Run Pulumi
cd iac/pulumi
export PLANTON_CLOUD_ARTIFACT_STORE_ID=test
pulumi up
```

## Troubleshooting

### Operator Not Deploying

```bash
# Check Pulumi state
pulumi stack export

# Check Helm release
helm list -n postgres-operator

# Check operator pod
kubectl get pods -n postgres-operator
kubectl logs -n postgres-operator deployment/postgres-operator
```

### Backup Configuration Issues

```bash
# Verify Secret creation
kubectl get secret -n postgres-operator r2-postgres-backup-credentials -o yaml

# Verify ConfigMap creation
kubectl get configmap -n postgres-operator postgres-pod-backup-config -o yaml

# Check if operator sees the ConfigMap
helm get values postgres-operator -n postgres-operator
```

### Label Inheritance Not Working

```bash
# Check Helm values
kubectl get configmap -n postgres-operator sh.helm.release.v1.postgres-operator.v1 -o json | jq '.data.release' | base64 -d | gunzip | jq '.config'

# Verify inherited_labels setting
kubectl get configmap -n postgres-operator sh.helm.release.v1.postgres-operator.v1 -o json | jq '.data.release' | base64 -d | gunzip | jq '.config.configKubernetes.inherited_labels'
```

## Common Patterns

### Production Deployment with Backups

```go
// In a custom Pulumi program
import (
    kuberneteszalandopostgresoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteszalandopostgresoperator/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteszalandopostgresoperator/v1/iac/pulumi/module"
)

stackInput := &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorStackInput{
    Target: &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperator{
        Metadata: &shared.CloudResourceMetadata{
            Name: "postgres-operator",
            Org:  "acme-corp",
            Env:  "production",
        },
        Spec: &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorSpec{
            Container: &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorSpecContainer{
                Resources: &kubernetes.ContainerResources{
                    Requests: &kubernetes.ContainerResourceQuantity{Cpu: "100m", Memory: "256Mi"},
                    Limits:   &kubernetes.ContainerResourceQuantity{Cpu: "2000m", Memory: "2Gi"},
                },
            },
            BackupConfig: &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorBackupConfig{
                R2Config: &kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorBackupR2Config{
                    CloudflareAccountId: "your-account-id",
                    BucketName:          "postgres-backups-prod",
                    AccessKeyId:         "your-access-key",
                    SecretAccessKey:     "your-secret-key",
                },
                BackupSchedule: "0 2 * * *",
            },
        },
    },
}

err := module.Resources(ctx, stackInput)
```

## Architecture

For detailed architecture documentation, see [overview.md](./overview.md).

Key design decisions:
- **Single Namespace**: Operator always deploys to `postgres-operator` namespace
- **Conditional Backup**: Backup resources only created when `backup_config` is provided
- **Label Inheritance**: Operator configured to propagate Planton labels to all databases
- **Helm-Based**: Uses official Zalando Helm chart for operator deployment

## References

- [Spec Definition](../../spec.proto)
- [Stack Outputs](../../stack_outputs.proto)
- [Architecture Overview](./overview.md)
- [Zalando Operator Docs](https://postgres-operator.readthedocs.io/)
- [Helm Chart](https://github.com/zalando/postgres-operator/tree/master/charts/postgres-operator)
- [WAL-G Documentation](https://github.com/wal-g/wal-g)

## Support

For issues or questions:
- Review [examples.md](../../examples.md) for usage patterns
- Check [troubleshooting section](#troubleshooting) above
- Consult [architecture docs](./overview.md) for design rationale
- File an issue in the Planton repository


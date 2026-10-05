# Kubernetes ServiceAccount - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes ServiceAccount — the in-cluster identity pods run as. It covers image-pull secrets, the tri-state token-automount flag, and cloud workload-identity binding (GKE Workload Identity, EKS IRSA, Azure AD Workload Identity) expressed as ServiceAccount annotations.

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go              # Orchestrator: provider init, resource creation, output export
    ├── locals.go            # Derived values: labels, annotation merging, workload-identity translation
    ├── service_account.go   # Creates kubernetes.core.v1.ServiceAccount resource
    └── outputs.go           # Exports service_account_name, namespace, rbac_subject, workload_identity_handle
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesServiceAccountIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels
   - Namespace defaulting (`default` when unset)
   - Workload-identity annotations merged over user annotations (workload-identity wins on collision)
   - Resolved image-pull secret names
   - The RBAC subject string `system:serviceaccount:<namespace>:<name>`
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **ServiceAccount Creation**: A single `kubernetes.core.v1.ServiceAccount` is created
5. **Output Export**: Name, namespace, RBAC subject, and workload-identity handle are exported

## Workload-Identity Annotation Mapping

| Cloud | Annotation | Value |
|-------|-----------|-------|
| GKE | `iam.gke.io/gcp-service-account` | GCP service account email |
| EKS | `eks.amazonaws.com/role-arn` | AWS IAM role ARN |
| AKS | `azure.workload.identity/client-id` | Managed-identity/Entra client ID |
| AKS (optional) | `azure.workload.identity/tenant-id` | Entra tenant ID, only when explicitly set |

## Tri-State Token Automount

`spec.automount_service_account_token` is optional:

- **unset** — the field is omitted from the ServiceAccount; the cluster default (mount the token) applies
- **false** — pods running as this identity get no API token mount unless their pod spec overrides
- **true** — the mount is explicit

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

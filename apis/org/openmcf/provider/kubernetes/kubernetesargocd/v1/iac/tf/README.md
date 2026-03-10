# Terraform Module to Deploy Argo CD on Kubernetes

This Terraform module deploys Argo CD on a Kubernetes cluster using the official Helm chart from the Argo Project.

## Prerequisites

- Kubernetes cluster with sufficient resources
- Helm provider configured
- Kubernetes provider configured
- OpenMCF CLI installed

## Usage

### Initialize Terraform

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=us-east-1" \
  --backend-config="key=kubernetes-stacks/test-argocd.tfstate"
```

### Plan Deployment

```shell
openmcf tofu plan --manifest hack/manifest.yaml
```

### Apply Deployment

```shell
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

### Destroy Deployment

```shell
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

## What Gets Deployed

This module deploys the following resources:

1. **Kubernetes Namespace**: An optionally created dedicated namespace for Argo CD (based on `create_namespace` flag)
2. **Argo CD Helm Release**: The official argo-cd chart from https://argoproj.github.io/argo-helm
3. **Core Components**:
   - API Server (with configured resource limits)
   - Application Controller (with configured resource limits)
   - Repo Server (with configured resource limits)
   - Redis (for caching and session storage)

## Configuration

The module reads configuration from the `KubernetesArgocd` manifest and applies it to the Helm chart values:

- **Container Resources**: CPU and memory limits/requests for all components
- **Ingress**: Optional ingress configuration for external access
- **Labels**: Resource labels for organization and tracking
- **Namespace Management**: Control whether the namespace should be created or use an existing one

## Namespace Management

The module supports two namespace management modes:

### Automatic Creation (create_namespace: true)

The module creates and manages the namespace with appropriate labels. This is the default behavior and is suitable for:
- Development and testing environments
- Scenarios where the deployment tool has full cluster permissions
- Simplified deployment workflows

### Use Existing Namespace (create_namespace: false)

The module uses a pre-existing namespace. This is recommended for:
- Production environments with strict namespace governance
- Multi-tenant clusters with namespace policies
- Environments where namespace creation requires elevated privileges

**Important**: When using `create_namespace: false`, ensure the namespace exists before running terraform apply:

```bash
kubectl create namespace <namespace-name>
```

If the namespace doesn't exist, the deployment will fail with a "namespace not found" error.

## Outputs

The module exports the following outputs:

- `namespace`: The Kubernetes namespace where Argo CD is deployed
- `service`: The Kubernetes service name for the Argo CD server
- `port_forward_command`: Command to set up local port-forwarding
- `kube_endpoint`: Internal cluster endpoint
- `external_hostname`: External ingress hostname (if ingress is enabled)
- `internal_hostname`: Internal ingress hostname (if ingress is enabled)

## Accessing Argo CD

### Via Port-Forward (Local Development)

Use the `port_forward_command` output:

```shell
kubectl port-forward -n <namespace> service/<service-name> 8080:80
```

Then access Argo CD at http://localhost:8080

### Via Ingress (Production)

If ingress is enabled, access Argo CD at the `external_hostname` output.

## Default Admin Credentials

The initial admin password is auto-generated. Retrieve it with:

```shell
kubectl -n <namespace> get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

**Important**: Change the admin password and configure SSO for production use.

## Notes

- The module uses Helm chart version 7.7.12 by default
- Resource requests and limits are configurable via the manifest
- The deployment is atomic - it will rollback on failure
- Redis is deployed with minimal resources; scale up for production


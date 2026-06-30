# KubernetesExternalDns Pulumi Module

Pulumi module for deploying ExternalDNS on Kubernetes clusters with multi-cloud provider support.

## Overview

This Pulumi module deploys [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) to Kubernetes clusters, enabling automatic DNS record management for Services and Ingresses. It supports four cloud DNS providers with cloud-native authentication:

- **GKE + Cloud DNS**: Workload Identity (no JSON keys)
- **EKS + Route53**: IRSA (no AWS access keys)
- **AKS + Azure DNS**: Managed Identity (no client secrets)
- **Cloudflare DNS**: API tokens (stored as Kubernetes secrets)

The module handles:
- Namespace creation (or lookup of existing namespace)
- Service account configuration with cloud provider annotations
- Secret creation (for Cloudflare)
- Helm chart deployment with provider-specific values
- Zone filtering for safety

## Architecture

```
┌─────────────────────────────────────────┐
│   KubernetesExternalDnsStackInput       │
│  ┌──────────────────────────────────┐   │
│  │ provider_config:                 │   │
│  │   kubernetes_provider_config_id  │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │ target:                          │   │
│  │   spec:                          │   │
│  │     gke/eks/aks/cloudflare       │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
              │
              ▼
     ┌────────────────────┐
     │ Pulumi Resources   │
     ├────────────────────┤
     │ • Namespace        │
     │ • ServiceAccount   │
     │ • Secret (if CF)   │
     │ • Helm Release     │
     └────────────────────┘
              │
              ▼
   ┌──────────────────────────┐
   │  ExternalDNS Deployment  │
   │  (watches K8s resources) │
   └──────────────────────────┘
              │
              ▼
    ┌───────────────────────┐
    │  External DNS Provider│
    │  (Route53, Cloud DNS, │
    │   Azure DNS, CF)      │
    └───────────────────────┘
```

## Usage

### Prerequisites

1. **Kubernetes cluster credentials** configured in Planton
2. **DNS provider setup**:
   - **GKE**: Cloud DNS zone + GCP service account with `roles/dns.admin`
   - **EKS**: Route53 hosted zone + OIDC provider configured
   - **AKS**: Azure DNS zone + Managed Identity with "DNS Zone Contributor" role
   - **Cloudflare**: DNS zone + API token with Zone:DNS:Edit permissions

### Input Structure

The module expects a `KubernetesExternalDnsStackInput` with:

```yaml
provider_config:
  kubernetes_provider_config_id: <kubernetes-cluster-credential-id>

target:
  metadata:
    name: external-dns-prod
  spec:
    target_cluster:
      kubernetes_cluster_id:
        value: prod-gke-cluster
    # Optional: override defaults
    namespace: external-dns  # default
    kubernetes_external_dns_version: v0.19.0  # default
    helm_chart_version: 1.19.0  # default
    # Provider config (choose one):
    gke:
      project_id:
        value: my-gcp-project
      dns_zone_id:
        value: my-cloud-dns-zone-id
```

### Running the Module

#### Via Planton CLI

```bash
# Create manifest
cat > external-dns.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalDns
metadata:
  name: external-dns-prod
spec:
  target_cluster:
    kubernetes_cluster_id:
      value: prod-gke-cluster
  gke:
    project_id:
      value: my-gcp-project
    dns_zone_id:
      value: my-dns-zone-id
EOF

# Deploy
planton deploy external-dns.yaml
```

#### Direct Pulumi Usage

```bash
cd iac/pulumi
pulumi stack init dev
pulumi config set --path kubernetes_external_dns_stack_input.provider_config.kubernetes_provider_config_id <cred-id>
pulumi up
```

### Debugging

Use the provided debug script to run Pulumi with detailed logging:

```bash
cd iac/pulumi
./debug.sh
```

## Namespace Management

The module provides flexible namespace management through the `create_namespace` flag:

### Automatic Namespace Creation

When `create_namespace` is set to `true`, the module creates the namespace using `corev1.NewNamespace()`:

```yaml
spec:
  namespace:
    value: external-dns
  create_namespace: true  # Module will create the namespace
```

### Using Existing Namespace

When `create_namespace` is `false` (default), the module uses `corev1.GetNamespace()` to lookup the existing namespace:

```yaml
spec:
  namespace:
    value: existing-namespace
  create_namespace: false  # Namespace must already exist
```

**Important**: If `create_namespace` is `false` and the namespace doesn't exist, the Pulumi deployment will fail with an error message:

```
failed to get existing namespace <name> - ensure it exists before deployment
```

This behavior ensures that namespace management is explicit and prevents accidental namespace creation when working with pre-configured namespaces that may have specific labels, annotations, or resource quotas.

## Module Implementation

### Key Files

- **`main.go`**: Pulumi entrypoint that calls the module
- **`module/main.go`**: Core logic for creating resources
- **`module/outputs.go`**: Output constants
- **`module/vars.go`**: Helm chart configuration

### Provider-Specific Logic

#### GKE (Google Cloud DNS)

```go
// Creates ServiceAccount with Workload Identity annotation
annotations["iam.gke.io/gcp-service-account"] = "<release-name>@<project-id>.iam.gserviceaccount.com"

// Helm values
values["provider"] = "google"
values["google"]["project"] = project_id
values["zoneIdFilters"] = [dns_zone_id]
```

**External setup required:**
- Create GCP service account: `<release-name>@<project-id>.iam.gserviceaccount.com`
- Grant `roles/dns.admin` on the DNS zone
- Create IAM binding for Workload Identity:
  ```bash
  gcloud iam service-accounts add-iam-policy-binding \
    <release-name>@<project-id>.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:<project-id>.svc.id.goog[external-dns/<release-name>]"
  ```

#### EKS (AWS Route53)

```go
// Creates ServiceAccount with IRSA annotation (if role provided)
if irsa_role_arn_override != "" {
  annotations["eks.amazonaws.com/role-arn"] = irsa_role_arn_override
}

// Helm values
values["provider"] = "aws"
values["zoneIdFilters"] = [route53_zone_id]
```

**External setup required (if not using override):**
- Create IAM role with trust policy for OIDC provider
- Attach policy granting Route53 permissions on the hosted zone
- OR provide existing role ARN via `irsa_role_arn_override`

#### AKS (Azure DNS)

```go
// Creates ServiceAccount with Managed Identity annotation
annotations["azure.workload.identity/client-id"] = managed_identity_client_id

// Helm values
values["provider"] = "azure"
values["domainFilters"] = [dns_zone_id]
```

**External setup required:**
- Create user-assigned Managed Identity
- Grant "DNS Zone Contributor" role on DNS zone
- Create federated credential linking to ServiceAccount

#### Cloudflare

```go
// Creates Secret with API token
secret, err := corev1.NewSecret(ctx, "cloudflare-api-token-<release-name>", ...)

// Helm values
values["provider"] = "cloudflare"
values["cloudflare"]["apiToken"] = secret_name
values["domainFilters"] = [dns_zone_id]
if is_proxied {
  values["cloudflare"]["proxied"] = true
}
```

**Security note**: API token is stored as a Kubernetes Secret. Consider using external secret operators (e.g., external-secrets, sealed-secrets) for production.

## Outputs

The module exports:

```go
// Namespace where ExternalDNS is deployed
ctx.Export("namespace", pulumi.String(namespace))

// Port-forward command for debugging (if needed)
ctx.Export("port_forward_command", pulumi.String(...))
```

## Multi-Instance Support

You can deploy multiple ExternalDNS instances in the same cluster:

```yaml
# Instance 1: Production domain
---
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalDns
metadata:
  name: external-dns-prod
spec:
  gke:
    dns_zone_id:
      value: prod-zone-id

# Instance 2: Staging domain
---
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalDns
metadata:
  name: external-dns-staging
spec:
  gke:
    dns_zone_id:
      value: staging-zone-id
```

Each instance:
- Gets a unique Helm release name (matches `metadata.name`)
- Has its own ServiceAccount and cloud IAM binding
- Is scoped to its own DNS zone via zone filtering

## Best Practices

### Security

1. **Use cloud-native IAM** (IRSA, Workload Identity, Managed Identity) instead of static credentials
2. **Scope to specific zones** via zone filtering (prevents managing wrong DNS records)
3. **Least privilege**: Grant only necessary DNS permissions (e.g., only the zones ExternalDNS manages)
4. **Separate instances**: Use different ExternalDNS instances for prod/staging zones

### Operations

1. **Pin versions**: Specify exact `kubernetes_external_dns_version` and `helm_chart_version` for reproducibility
2. **Monitor logs**: Check ExternalDNS pod logs for sync errors
3. **Test in staging**: Deploy to staging cluster first, verify DNS records are created correctly
4. **GitOps**: Store manifests in Git and use ArgoCD/Flux for deployment

### Troubleshooting

**ExternalDNS pod not starting?**
```bash
kubectl logs -n external-dns <pod-name>
```

**Authentication errors?**
- **GKE**: Verify Workload Identity binding: `gcloud iam service-accounts get-iam-policy <GSA>`
- **EKS**: Verify IRSA role trust policy includes your cluster's OIDC provider
- **AKS**: Verify federated credential exists for the Managed Identity
- **Cloudflare**: Verify API token has correct permissions and is not expired

**DNS records not being created?**
- Check zone filtering matches your DNS zone
- Verify the Ingress/Service has the ExternalDNS annotation
- Check ExternalDNS logs for sync errors
- Verify cloud IAM permissions are correct

## Examples

See [examples.md](../examples.md) for complete YAML manifests for all cloud providers.

## Reference

- [ExternalDNS Documentation](https://kubernetes-sigs.github.io/external-dns/)
- [Bitnami Helm Chart](https://github.com/bitnami/charts/tree/main/bitnami/external-dns)
- [Planton Documentation](../../README.md)


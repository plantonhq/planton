# Kubernetes ExternalDNS

Automate DNS record management for Kubernetes services and ingresses across multiple cloud providers.

## Overview

**KubernetesExternalDns** is a Planton component that deploys and configures [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) on Kubernetes clusters. ExternalDNS automatically creates, updates, and deletes DNS records in external DNS providers (like Route53, Cloud DNS, Azure DNS, or Cloudflare) based on Kubernetes resources such as Services and Ingresses.

Instead of manually managing DNS records every time you deploy a service, ExternalDNS watches your cluster and keeps DNS in sync automatically. This eliminates manual DNS management, reduces errors, and enables true declarative infrastructure.

## Why We Created This

Managing DNS records manually for Kubernetes services is tedious and error-prone:

- **Manual workflow overhead**: Deploy a service → grab the external IP → log into DNS provider → create/update A record → repeat for every service
- **Configuration drift**: DNS records get out of sync with actual service endpoints
- **Multi-cluster complexity**: Managing DNS across dozens of clusters in multiple cloud providers becomes unmanageable
- **Inconsistent authentication**: Each cloud provider has different credential mechanisms (IRSA, Workload Identity, Managed Identity)
- **Security risks**: Long-lived credentials scattered across clusters

Our KubernetesExternalDns module solves these problems by providing:

1. **Declarative DNS automation**: Define your hostname in a Kubernetes annotation, ExternalDNS does the rest
2. **Multi-cloud support**: Consistent interface for GKE, EKS, AKS, and Cloudflare
3. **Cloud-native security**: Uses IRSA, Workload Identity, and Managed Identity (no long-lived secrets)
4. **Zone scoping**: Automatically limits ExternalDNS to specific DNS zones for safety
5. **Production-ready defaults**: Upsert-only policy, zone filtering, and proper RBAC out of the box

## Supported Providers

- **GKE (Google Kubernetes Engine)** with Cloud DNS
- **EKS (Amazon Elastic Kubernetes Service)** with Route53
- **AKS (Azure Kubernetes Service)** with Azure DNS
- **Cloudflare DNS** (works with any Kubernetes cluster)

Each provider is configured with cloud-native authentication (no static credentials):
- **GKE**: Workload Identity
- **EKS**: IAM Roles for Service Accounts (IRSA)
- **AKS**: Azure Managed Identity
- **Cloudflare**: API token (stored in Kubernetes secret)

## Key Features

### 🔄 Automated DNS Synchronization
- Watch Kubernetes Ingresses and Services for DNS annotations
- Automatically create/update/delete DNS records in external provider
- No manual DNS management required

### 🔐 Secure by Default
- Cloud-native authentication (IRSA, Workload Identity, Managed Identity)
- No long-lived credentials in cluster
- Zone-scoped permissions (least privilege)
- Upsert-only policy prevents accidental deletions

### 🌍 Multi-Cloud Ready
- Single API for GKE, EKS, AKS, and Cloudflare
- Provider-specific optimizations built-in
- Consistent configuration across clouds

### 📦 Production-Ready
- Deployed via Helm with battle-tested charts
- Configurable versions for ExternalDNS and Helm chart
- Namespace isolation
- Proper RBAC and ServiceAccount setup

### 🎯 Zone Filtering
- Automatically scoped to specific DNS zones
- Prevents cross-contamination between environments
- Support for multiple ExternalDNS instances per cluster (different zones)

## Quick Start

See [examples.md](examples.md) for complete YAML manifests for each cloud provider.

### Example: ExternalDNS for GKE

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalDns
metadata:
  name: external-dns-prod
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: kubernetes-external-dns
  gke:
    project_id:
      value: my-gcp-project
    dns_zone_id:
      value: my-dns-zone-id
```

This creates an ExternalDNS deployment that:
- Uses Workload Identity to authenticate with Google Cloud DNS
- Is scoped to manage records only in the specified DNS zone
- Watches for Kubernetes Ingresses and Services with DNS annotations
- Automatically creates A/CNAME records pointing to LoadBalancer IPs

## Benefits

### For Platform Teams
- **Standardized deployment**: Same interface across all clusters and clouds
- **Security compliance**: No static credentials, all authentication via cloud-native IAM
- **Audit trail**: All DNS changes are logged through cloud provider audit logs
- **Multi-tenancy**: Deploy multiple ExternalDNS instances for different zones/domains

### For Application Teams  
- **Zero DNS overhead**: Just annotate your Ingress/Service, DNS is automatic
- **Faster deployments**: No waiting for manual DNS changes
- **Self-service**: No dependency on platform team for DNS records
- **GitOps-friendly**: DNS is declared in the same YAML as your service

### For Everyone
- **Fewer incidents**: No more "forgot to update DNS" outages
- **Faster iteration**: Deploy new services without DNS friction
- **Cost savings**: Reduce operational toil and manual processes
- **Better reliability**: DNS stays in sync with actual service state

## How It Works

1. **Deploy**: Apply a KubernetesExternalDns manifest to your cluster
2. **Watch**: ExternalDNS monitors Ingress/Service resources for annotations like `external-dns.alpha.kubernetes.io/hostname`
3. **Sync**: When it detects a hostname annotation, ExternalDNS creates/updates the DNS record in your cloud provider
4. **Maintain**: As services scale, move, or are deleted, DNS records are kept in sync automatically

## Use Cases

- **Public-facing services**: Automatically create DNS for LoadBalancer services
- **Ingress hostnames**: Auto-manage A/CNAME records for Ingress resources
- **Multi-region active-active**: Keep DNS updated as traffic shifts between regions
- **Blue-green deployments**: DNS points to new stack automatically when Ingress updates
- **Development environments**: Each dev cluster gets its own subdomain, managed automatically

## Documentation

- **[Research Documentation](docs/README.md)**: Deep dive into ExternalDNS landscape, authentication patterns, and best practices
- **[Examples](examples.md)**: Complete YAML manifests for GKE, EKS, AKS, and Cloudflare
- **[Pulumi Module](iac/pulumi/README.md)**: Using the Pulumi implementation directly
- **[Terraform Module](iac/tf/README.md)**: Using the Terraform implementation directly

## Configuration Reference

| Field | Description | Default |
|-------|-------------|---------|
| `target_cluster` | Kubernetes cluster to deploy ExternalDNS on | Required |
| `namespace` | Kubernetes namespace for ExternalDNS | `kubernetes-external-dns` |
| `create_namespace` | Whether to create the namespace (if false, namespace must exist) | `false` |
| `kubernetes_external_dns_version` | ExternalDNS image version | `v0.19.0` |
| `helm_chart_version` | Helm chart version | `1.19.0` |
| `provider_config` | Cloud provider configuration (gke, eks, aks, or cloudflare) | Required |

## Namespace Management

The component provides flexible namespace management through the `create_namespace` flag:

### Automatic Namespace Creation

When `create_namespace` is set to `true`, the module creates the namespace:

```yaml
spec:
  namespace:
    value: kubernetes-external-dns
  create_namespace: true  # Module will create the namespace
```

### Using Existing Namespace

When `create_namespace` is `false` (default), the namespace must exist before deployment:

```yaml
spec:
  namespace:
    value: existing-namespace
  create_namespace: false  # Namespace must already exist
```

**Important**: If `create_namespace` is `false` and the namespace doesn't exist, deployment will fail with a clear error message.

### GKE Configuration

| Field | Description |
|-------|-------------|
| `project_id` | GCP project ID hosting the DNS zone |
| `dns_zone_id` | Cloud DNS zone ID to manage |

### EKS Configuration

| Field | Description |
|-------|-------------|
| `route53_zone_id` | Route53 hosted zone ID to manage |
| `irsa_role_arn_override` | Optional IAM role ARN for IRSA (auto-created if empty) |

### AKS Configuration

| Field | Description |
|-------|-------------|
| `dns_zone_id` | Azure DNS zone ID |
| `managed_identity_client_id` | Azure Managed Identity client ID |

### Cloudflare Configuration

| Field | Description |
|-------|-------------|
| `api_token` | Cloudflare API token with Zone:DNS:Edit permissions |
| `dns_zone_id` | Cloudflare zone ID to manage |
| `is_proxied` | Enable Cloudflare proxy (orange cloud) for records |

## Prerequisites

### For GKE
- GKE cluster with Workload Identity enabled
- Google Cloud DNS zone created
- GCP service account with `roles/dns.admin` on the DNS zone
- IAM binding for Workload Identity (created by module)

### For EKS
- EKS cluster with OIDC provider configured
- Route53 hosted zone created
- IAM role with Route53 permissions (created by module or provided)

### For AKS
- AKS cluster with Azure AD Workload Identity enabled
- Azure DNS zone created
- User-assigned Managed Identity with "DNS Zone Contributor" role

### For Cloudflare
- Cloudflare account and zone
- API token with Zone:Zone:Read and Zone:DNS:Edit permissions

## Next Steps

1. Review the [examples](examples.md) for your cloud provider
2. Read the [research documentation](docs/README.md) to understand architecture patterns
3. Create a KubernetesExternalDns manifest
4. Deploy it via `planton deploy`
5. Test by creating a Service or Ingress with DNS annotation

## Support

For issues, questions, or contributions, see the main [Planton repository](https://github.com/plantonhq/planton).


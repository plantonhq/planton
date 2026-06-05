# KubernetesExternalDns Terraform Module

Terraform module for deploying ExternalDNS on Kubernetes clusters with multi-cloud DNS provider support.

## Overview

This Terraform module deploys [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) to Kubernetes clusters using the Helm provider. It supports four DNS providers with cloud-native authentication:

- **GKE + Cloud DNS**: Uses Workload Identity
- **EKS + Route53**: Uses IRSA (IAM Roles for Service Accounts)
- **AKS + Azure DNS**: Uses Managed Identity
- **Cloudflare DNS**: Uses API tokens (stored as Kubernetes secrets)

The module creates:
- Kubernetes namespace (or uses existing namespace)
- Service account with cloud provider annotations
- Secret (for Cloudflare)
- Helm release with provider-specific configuration

## Prerequisites

### Terraform Providers

This module requires the following providers:

```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.11.0"
    }
  }
}
```

### Kubernetes Cluster Access

Configure the Kubernetes and Helm providers with access to your cluster:

```hcl
provider "kubernetes" {
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}
```

### Cloud Provider Setup

#### For GKE
- GKE cluster with Workload Identity enabled
- Cloud DNS zone created
- GCP service account with `roles/dns.admin`
- IAM binding for Workload Identity

#### For EKS
- EKS cluster with OIDC provider configured
- Route53 hosted zone created
- IAM role with Route53 permissions (or module will use auto-created role)

#### For AKS
- AKS cluster with Azure AD Workload Identity enabled
- Azure DNS zone created
- User-assigned Managed Identity with "DNS Zone Contributor" role
- Federated credential configured

#### For Cloudflare
- Cloudflare account and zone
- API token with `Zone:Zone:Read` and `Zone:DNS:Edit` permissions

## Usage

### Basic Example (GKE)

```hcl
module "external_dns" {
  source = "path/to/module"

  metadata = {
    name = "external-dns-prod"
  }

  spec = {
    namespace        = "external-dns"  # required
    create_namespace = true            # optional, defaults to false

    gke = {
      project_id  = "my-gcp-project"
      dns_zone_id = "my-cloud-dns-zone-id"
    }
  }
}
```

### EKS Example

```hcl
module "external_dns" {
  source = "path/to/module"

  metadata = {
    name = "external-dns-eks"
  }

  spec = {
    namespace = "external-dns"  # required
    eks = {
      route53_zone_id        = "Z1234567890ABC"
      irsa_role_arn_override = "arn:aws:iam::123456789012:role/external-dns-role"  # optional
    }
  }
}
```

### Cloudflare Example

```hcl
module "external_dns" {
  source = "path/to/module"

  metadata = {
    name = "external-dns-cloudflare"
  }

  spec = {
    namespace = "external-dns"  # required
    cloudflare = {
      api_token   = var.cloudflare_api_token  # Use variable for sensitive data
      dns_zone_id = "1234567890abcdef1234567890abcdef"
      is_proxied  = true  # optional, enables Cloudflare proxy
    }
  }
}
```

## Namespace Management

The module provides flexible namespace management through the `create_namespace` flag:

### Automatic Namespace Creation

When `create_namespace` is set to `true`, the module creates the namespace:

```hcl
spec = {
  namespace        = "external-dns"
  create_namespace = true  # Module will create the namespace
  # ... provider config
}
```

### Using Existing Namespace

When `create_namespace` is `false` (default), the module uses a data source to lookup the existing namespace:

```hcl
spec = {
  namespace        = "existing-namespace"
  create_namespace = false  # Namespace must already exist
  # ... provider config
}
```

**Important**: If `create_namespace` is `false` and the namespace doesn't exist, the Terraform apply will fail with an error indicating the namespace was not found. This behavior ensures that namespace management is explicit and prevents accidental namespace creation when working with pre-configured namespaces that may have specific labels, annotations, or resource quotas.

## Variables

### `metadata` (Required)

```hcl
metadata = {
  name = string           # Required: Resource name
  id = string            # Optional: Resource ID (defaults to name)
  org = string           # Optional: Organization
  env = string           # Optional: Environment
  labels = map(string)   # Optional: Additional labels
  tags = list(string)    # Optional: Tags
  version = object({     # Optional: Version info
    id = string
    message = string
  })
}
```

### `spec` (Required)

The `spec` variable defines the ExternalDNS configuration. Choose **one** provider configuration:

#### GKE Configuration

```hcl
spec = {
  namespace            = "external-dns"  # Required
  create_namespace     = true            # Optional, defaults to false
  external_dns_version = "v0.19.0"        # Optional
  helm_chart_version   = "1.19.0"         # Optional

  gke = {
    project_id  = "my-gcp-project"  # GCP project ID
    dns_zone_id = "my-dns-zone-id"  # Cloud DNS zone ID
  }
}
```

#### EKS Configuration

```hcl
spec = {
  namespace = "external-dns"  # Required
  eks = {
    route53_zone_id        = "Z1234567890ABC"  # Route53 hosted zone ID
    irsa_role_arn_override = "arn:aws:iam::123456789012:role/external-dns-role"  # Optional
  }
}
```

#### AKS Configuration

```hcl
spec = {
  namespace = "external-dns"  # Required
  aks = {
    dns_zone_id                = "/subscriptions/.../dnszones/example.com"  # Azure DNS zone ID
    managed_identity_client_id = "12345678-1234-1234-1234-123456789012"     # Managed Identity client ID
  }
}
```

#### Cloudflare Configuration

```hcl
spec = {
  namespace = "external-dns"  # Required
  cloudflare = {
    api_token   = "your-api-token"                     # Cloudflare API token
    dns_zone_id = "1234567890abcdef1234567890abcdef"  # Cloudflare zone ID
    is_proxied  = true                                 # Optional: Enable Cloudflare proxy
  }
}
```

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | Namespace where ExternalDNS is deployed |
| `release_name` | Helm release name |
| `service_account_name` | Kubernetes service account name |
| `provider_type` | DNS provider type (google, aws, azure, cloudflare) |
| `gke_service_account_email` | Google Service Account email (GKE only) |
| `cloudflare_secret_name` | Kubernetes secret name for Cloudflare API token (Cloudflare only) |

## Post-Deployment Setup

### GKE (Workload Identity)

After deploying the module, create and configure the Google Service Account:

```bash
# Create GCP service account
export GSA_NAME=$(terraform output -raw service_account_name)
export PROJECT_ID="my-gcp-project"
export NAMESPACE=$(terraform output -raw namespace)

gcloud iam service-accounts create $GSA_NAME \
  --project=$PROJECT_ID

# Grant DNS admin role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/dns.admin"

# Create Workload Identity binding
gcloud iam service-accounts add-iam-policy-binding \
  $GSA_NAME@$PROJECT_ID.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:$PROJECT_ID.svc.id.goog[$NAMESPACE/$GSA_NAME]"
```

### EKS (IRSA)

If not providing `irsa_role_arn_override`, create the IAM role:

```bash
# Get OIDC provider
export OIDC_PROVIDER=$(aws eks describe-cluster --name my-cluster --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")
export NAMESPACE=$(terraform output -raw namespace)
export SA_NAME=$(terraform output -raw service_account_name)

# Create trust policy
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/$OIDC_PROVIDER"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "$OIDC_PROVIDER:sub": "system:serviceaccount:$NAMESPACE:$SA_NAME"
        }
      }
    }
  ]
}
EOF

# Create IAM role
aws iam create-role --role-name external-dns-role --assume-role-policy-document file://trust-policy.json

# Attach Route53 policy
aws iam attach-role-policy --role-name external-dns-role --policy-arn arn:aws:iam::aws:policy/AmazonRoute53FullAccess
```

### AKS (Managed Identity)

Configure the federated credential:

```bash
export MANAGED_IDENTITY_NAME="external-dns-identity"
export NAMESPACE=$(terraform output -raw namespace)
export SA_NAME=$(terraform output -raw service_account_name)

# Create federated credential
az identity federated-credential create \
  --name external-dns-credential \
  --identity-name $MANAGED_IDENTITY_NAME \
  --resource-group my-resource-group \
  --issuer https://MY_AKS_CLUSTER_OIDC_ISSUER \
  --subject system:serviceaccount:$NAMESPACE:$SA_NAME
```

## Examples

See [examples.md](examples.md) for complete Terraform configurations.

## Troubleshooting

### Helm Release Fails

Check Helm release status:

```bash
helm list -n $(terraform output -raw namespace)
helm status $(terraform output -raw release_name) -n $(terraform output -raw namespace)
```

### ExternalDNS Pod Not Starting

Check pod logs:

```bash
kubectl logs -n $(terraform output -raw namespace) -l app.kubernetes.io/name=external-dns
```

### Authentication Errors

Verify cloud IAM configuration:

**GKE:**
```bash
gcloud iam service-accounts get-iam-policy $(terraform output -raw gke_service_account_email)
```

**EKS:**
```bash
aws iam get-role --role-name external-dns-role
```

**AKS:**
```bash
az identity show --name external-dns-identity --resource-group my-rg
```

## Security Best Practices

1. **Use secrets for API tokens**: Store Cloudflare API tokens in Terraform variables, not in code
2. **Limit DNS permissions**: Scope IAM roles to specific DNS zones only
3. **Use cloud-native auth**: Prefer IRSA/Workload Identity/Managed Identity over static credentials
4. **Enable zone filtering**: Always configure zone filtering to prevent managing wrong DNS

## Reference

- [ExternalDNS Documentation](https://kubernetes-sigs.github.io/external-dns/)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)
- [Terraform Helm Provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs)


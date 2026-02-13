---
title: "Credentials"
description: "How OpenMCF loads cloud provider credentials and a quick reference for all supported providers"
icon: "security"
order: 30
---

# Credentials

OpenMCF needs credentials to create and manage resources in your cloud accounts. This page explains how credentials are loaded and provides a quick reference for all supported providers. For detailed setup instructions for the three major clouds, see the dedicated provider guides:

- [AWS Provider Setup](./aws-provider-setup)
- [GCP Provider Setup](./gcp-provider-setup)
- [Azure Provider Setup](./azure-provider-setup)

## How Credentials Are Loaded

OpenMCF supports two methods for providing credentials. Both methods work with all IaC engines (Pulumi, OpenTofu, Terraform).

### Method 1: Environment Variables

Set standard environment variables for your cloud provider. OpenMCF and the underlying IaC engines (Pulumi, Terraform, OpenTofu) read them automatically:

```bash
# Example: AWS
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-west-2"

openmcf pulumi up -f ops/aws/database.yaml
```

This is the simplest approach and works out of the box if you already have your cloud provider CLI configured (`aws`, `gcloud`, `az`).

### Method 2: Provider Config File (`-p` Flag)

Pass a YAML file containing credentials using the `-p` (or `--provider-config`) flag:

```bash
openmcf pulumi up -f ops/aws/database.yaml -p aws-credential.yaml
```

The CLI auto-detects which provider is needed from your manifest's `apiVersion` and `kind`. You do not need to specify the provider type separately.

Provider config files use `snake_case` field names matching the Protocol Buffer definitions. For AWS:

```yaml
# aws-credential.yaml
account_id: "123456789012"
access_key_id: "AKIAIOSFODNN7EXAMPLE"
secret_access_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
region: "us-west-2"
```

When a `-p` file is provided, OpenMCF parses it, validates it against the provider's proto definition, and converts the fields to the standard environment variables that the IaC engine expects.

### When to Use Each Method

| Method | Best for |
|--------|----------|
| Environment variables | Local development, CI/CD pipelines, existing cloud CLI setups |
| Provider config file (`-p`) | Multi-account scenarios, explicit credential management, scripted workflows |

## Quick Reference: Environment Variables by Provider

These are the environment variables that OpenMCF and the underlying IaC engines read for each provider.

### AWS

```bash
AWS_ACCESS_KEY_ID          # Required
AWS_SECRET_ACCESS_KEY      # Required
AWS_REGION                 # Required
AWS_SESSION_TOKEN          # Optional (temporary credentials via STS)
AWS_PROFILE                # Optional (named profile from ~/.aws/credentials)
```

See [AWS Provider Setup](./aws-provider-setup) for IAM configuration and detailed instructions.

### GCP

```bash
GOOGLE_CREDENTIALS                 # JSON service account key (set by -p flag)
GOOGLE_APPLICATION_CREDENTIALS     # Path to service account JSON key file
GOOGLE_CLOUD_PROJECT               # Default project ID
```

See [GCP Provider Setup](./gcp-provider-setup) for service account setup and detailed instructions.

### Azure

```bash
ARM_CLIENT_ID              # Service principal application (client) ID
ARM_CLIENT_SECRET          # Service principal password
ARM_TENANT_ID              # Azure AD tenant ID
ARM_SUBSCRIPTION_ID        # Azure subscription ID
```

See [Azure Provider Setup](./azure-provider-setup) for service principal setup and detailed instructions.

### Kubernetes

```bash
KUBECONFIG                 # Path to kubeconfig file (default: ~/.kube/config)
```

OpenMCF uses the default kubeconfig at `~/.kube/config` if no `KUBECONFIG` variable is set. For managed clusters, use your cloud provider's CLI to generate the kubeconfig:

```bash
# GKE
gcloud container clusters get-credentials my-cluster --region us-central1

# EKS
aws eks update-kubeconfig --name my-cluster --region us-west-2

# AKS
az aks get-credentials --resource-group my-rg --name my-cluster
```

### Cloudflare

```bash
CLOUDFLARE_API_TOKEN       # Scoped API token (recommended)
CLOUDFLARE_API_KEY         # Global API key (legacy, not recommended)
CLOUDFLARE_EMAIL           # Account email (required with legacy API key)
```

### DigitalOcean

```bash
DIGITALOCEAN_TOKEN         # Personal access token
SPACES_ACCESS_KEY_ID       # Spaces access key (for object storage)
SPACES_SECRET_ACCESS_KEY   # Spaces secret key (for object storage)
```

### Civo

```bash
CIVO_TOKEN                 # Civo API token
```

### OpenStack

```bash
OS_AUTH_URL                # Identity endpoint URL
OS_REGION_NAME             # Region name
OS_USERNAME                # Username (password auth)
OS_PASSWORD                # Password (password auth)
OS_TENANT_NAME             # Tenant/project name
OS_TENANT_ID               # Tenant/project ID
OS_USER_DOMAIN_NAME        # User domain name
OS_PROJECT_DOMAIN_NAME     # Project domain name
```

OpenStack also supports application credential and token authentication. See the OpenStack provider documentation for details.

### Scaleway

```bash
SCW_ACCESS_KEY                 # Access key
SCW_SECRET_KEY                 # Secret key
SCW_DEFAULT_PROJECT_ID         # Default project ID
SCW_DEFAULT_ORGANIZATION_ID    # Default organization ID
SCW_DEFAULT_REGION             # Default region
SCW_DEFAULT_ZONE               # Default zone
```

### MongoDB Atlas

```bash
MONGODB_ATLAS_PUBLIC_KEY       # Atlas public key
MONGODB_ATLAS_PRIVATE_KEY      # Atlas private key
```

### Confluent

```bash
CONFLUENT_API_KEY              # Confluent API key
CONFLUENT_API_SECRET           # Confluent API secret
```

### Snowflake

```bash
SNOWFLAKE_ACCOUNT              # Account identifier
SNOWFLAKE_REGION               # Region
SNOWFLAKE_USERNAME             # Username
SNOWFLAKE_PASSWORD             # Password
```

### Auth0

```bash
AUTH0_DOMAIN                   # Auth0 tenant domain
AUTH0_CLIENT_ID                # Machine-to-machine client ID
AUTH0_CLIENT_SECRET            # Machine-to-machine client secret
```

### OpenFGA

```bash
FGA_API_URL                    # OpenFGA server URL
FGA_API_TOKEN                  # API token (optional)
FGA_CLIENT_ID                  # Client ID for OAuth2 (optional)
FGA_CLIENT_SECRET              # Client secret for OAuth2 (optional)
```

## Security Best Practices

**Use environment variables or `-p` files, never inline credentials in manifests.** Manifests are typically committed to version control. Credentials should not be.

**Use least-privilege permissions.** Grant only the IAM permissions required for the specific resources you are deploying. Avoid admin or root credentials.

**Rotate credentials regularly.** Set a 90-day rotation schedule, and rotate immediately if a credential may have been exposed.

**Use temporary credentials where possible.** AWS IAM roles, GCP Workload Identity, and Azure Managed Identity provide credentials that rotate automatically without manual key management.

**Separate credentials by environment.** Use different credentials for development, staging, and production. This limits the blast radius of a compromised credential.

**In CI/CD pipelines, use your platform's secret management.** GitHub Actions secrets, GitLab CI variables, and similar mechanisms encrypt credentials at rest and inject them as environment variables. See [CI/CD Integration](./cicd-integration) for patterns.

## What's Next

- [AWS Provider Setup](./aws-provider-setup) — Detailed AWS IAM configuration
- [GCP Provider Setup](./gcp-provider-setup) — Detailed GCP service account setup
- [Azure Provider Setup](./azure-provider-setup) — Detailed Azure service principal setup
- [CI/CD Integration](./cicd-integration) — Credential injection in pipelines
- [State Backends](./state-backends) — Configure state storage (also requires credentials)

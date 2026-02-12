---
title: "Credentials Management"
description: "Complete guide to providing cloud provider credentials for OpenMCF - environment variables, credential files, and best practices"
icon: "key"
order: 3
---

# Credentials Management Guide

Your complete guide to securely providing cloud provider credentials to OpenMCF.

---

## Overview

To deploy infrastructure, OpenMCF needs permission to create resources in your cloud accounts. These permissions come from **credentials** - authentication information that proves you have the right to make changes.

Think of credentials like keys to different buildings. AWS credentials are like keys to Amazon's building, GCP credentials open Google's building, and so on. OpenMCF needs the right key for whichever building (cloud provider) you're working with.

### How Credentials Are Loaded

OpenMCF uses the same credential loading mechanism as the underlying IaC providers (Pulumi, Terraform, OpenTofu). Credentials are loaded in this order:

1. **Environment Variables** (Default) - The CLI and IaC providers automatically read credentials from standard environment variables for each provider
2. **Provider Config File** (`-p` flag) - Explicit credentials file that overrides environment variables

**The key insight**: You don't need to do anything special if you already have your cloud provider CLI configured (e.g., `aws`, `gcloud`, `az`). The environment variables and default credential files used by those tools are automatically picked up.

### When to Use Each Method

| Method | Best For | Example |
|--------|----------|---------|
| Environment Variables | Local development, CI/CD pipelines | `export AWS_ACCESS_KEY_ID=...` |
| Provider Config File (`-p`) | Multi-account scenarios, explicit credentials | `openmcf apply -f manifest.yaml -p aws-creds.yaml` |

Both methods are secure when used properly. Environment variables are simpler; config files offer more control.

---

## General Principles

### Security Best Practices

**✅ DO**:
- Use environment variables or credential files
- Store credentials in password managers
- Use IAM roles and temporary credentials when possible
- Rotate credentials regularly
- Use least-privilege permissions (only what's needed)
- Use different credentials for dev/staging/prod

**❌ DON'T**:
- Commit credentials to Git
- Hardcode credentials in manifests
- Share credentials via email or chat
- Use root/admin credentials for deployments
- Reuse personal credentials for automation

### Permission Scoping

Grant only the permissions needed for your deployments:

- Creating compute resources? Grant compute permissions.
- Managing databases? Grant database permissions.
- Don't grant `*:*` (full access) unless absolutely necessary.

Each cloud provider has guides for setting up appropriate IAM policies.

---

## AWS Credentials

### Method 1: Environment Variables (Recommended)

The simplest approach - AWS CLI and OpenMCF both read these:

```bash
# Set credentials
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_DEFAULT_REGION="us-west-2"  # Optional but recommended

# Verify they work
aws sts get-caller-identity

# Deploy
openmcf pulumi up -f ops/aws/vpc.yaml
```

**Where to get these**:
1. AWS Console → IAM → Users → Your User → Security Credentials
2. Click "Create access key"
3. Download and store securely (you won't see the secret again)

### Method 2: Provider Config Files via CLI Flag

```bash
# Create provider config file
cat > ~/.aws/openmcf-prod.yaml <<EOF
accessKeyId: AKIAIOSFODNN7EXAMPLE
secretAccessKey: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region: us-west-2
EOF

# Use with CLI (provider type auto-detected from manifest)
openmcf pulumi up \
  -f ops/aws/vpc.yaml \
  -p ~/.aws/openmcf-prod.yaml
```

The CLI automatically detects which provider is needed based on your manifest's `apiVersion` and `kind`. You don't need to specify provider-specific flags.

### Method 3: AWS Profiles (Recommended for Multiple Accounts)

Use AWS CLI profiles to manage multiple accounts:

```bash
# Configure profile
aws configure --profile production
# Enter access key, secret, region when prompted

# Use profile with OpenMCF
export AWS_PROFILE=production
openmcf pulumi up -f ops/aws/vpc.yaml
```

### Method 4: IAM Roles (Best for EC2/ECS/Lambda)

If running on AWS compute (EC2, ECS, Lambda), use IAM roles instead of access keys:

```bash
# No credentials needed - automatically provided by AWS
# Just ensure your EC2 instance/ECS task has an IAM role attached

openmcf pulumi up -f ops/aws/vpc.yaml
```

### Troubleshooting AWS Credentials

**Problem**: "Unable to locate credentials"

```bash
# Solution 1: Check if credentials are set
env | grep AWS

# Solution 2: Verify credentials are valid
aws sts get-caller-identity

# Solution 3: Check AWS CLI config
cat ~/.aws/credentials
cat ~/.aws/config
```

**Problem**: "Access Denied" errors

```bash
# Check what permissions your credentials have
aws iam get-user

# Verify you have the necessary IAM policies attached
# Contact your AWS administrator if you need additional permissions
```

---

## Google Cloud (GCP) Credentials

### Method 1: Service Account Key File (Recommended)

**Step 1**: Create service account and key:

```bash
# Via gcloud CLI
gcloud iam service-accounts create openmcf-deployer \
  --display-name "OpenMCF Deployer"

# Grant necessary roles (example: GKE admin)
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:openmcf-deployer@my-project.iam.gserviceaccount.com" \
  --role="roles/container.admin"

# Create and download key
gcloud iam service-accounts keys create ~/gcp-key.json \
  --iam-account=openmcf-deployer@my-project.iam.gserviceaccount.com
```

**Step 2**: Use the key file:

```bash
# Method A: Environment variable (most common)
export GOOGLE_APPLICATION_CREDENTIALS=~/gcp-key.json

openmcf pulumi up -f ops/gcp/gke-cluster.yaml

# Method B: CLI flag (provider auto-detected from manifest)
openmcf pulumi up \
  -f ops/gcp/gke-cluster.yaml \
  -p ~/gcp-credential.yaml
```

### Method 2: Application Default Credentials (Local Development)

```bash
# Authenticate with your personal Google account
gcloud auth application-default login

# No additional configuration needed
openmcf pulumi up -f ops/gcp/gke-cluster.yaml
```

**When to use**: Local development, personal projects.  
**When NOT to use**: Production, CI/CD (use service accounts instead).

### Method 3: Workload Identity (Best for GKE)

If deploying from within GKE, use Workload Identity:

```bash
# Configure at cluster creation - no credentials in code
# Kubernetes service accounts automatically get GCP permissions
# OpenMCF automatically uses workload identity when available
```

### GCP Credential File Format for CLI Flag

```yaml
# gcp-credential.yaml
serviceAccountKeyBase64: "<base64-encoded-json-key>"
```

To create:

```bash
# Encode your service account key
base64 -i ~/gcp-key.json | tr -d '\n' > base64-key.txt

# Create YAML file
cat > gcp-credential.yaml <<EOF
serviceAccountKeyBase64: $(cat base64-key.txt)
EOF
```

### Troubleshooting GCP Credentials

**Problem**: "Application Default Credentials not found"

```bash
# Solution: Set environment variable
export GOOGLE_APPLICATION_CREDENTIALS=~/gcp-key.json

# Or authenticate with gcloud
gcloud auth application-default login
```

**Problem**: "Permission denied" errors

```bash
# Check what project you're using
gcloud config get-value project

# List available projects
gcloud projects list

# Set correct project
gcloud config set project my-project-id

# Verify service account permissions
gcloud projects get-iam-policy my-project \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:openmcf-deployer@*"
```

---

## Azure Credentials

### Method 1: Service Principal (Recommended)

**Step 1**: Create service principal:

```bash
# Create service principal and get credentials
az ad sp create-for-rbac \
  --name "openmcf-deployer" \
  --role contributor \
  --scopes /subscriptions/<subscription-id>

# Output shows:
# {
#   "appId": "abc-123",           # This is CLIENT_ID
#   "displayName": "...",
#   "password": "xyz-789",         # This is CLIENT_SECRET
#   "tenant": "def-456"            # This is TENANT_ID
# }
```

**Step 2**: Use credentials:

```bash
# Method A: Environment variables
export ARM_CLIENT_ID="abc-123"
export ARM_CLIENT_SECRET="xyz-789"
export ARM_TENANT_ID="def-456"
export ARM_SUBSCRIPTION_ID="your-subscription-id"

openmcf pulumi up -f ops/azure/aks-cluster.yaml

# Method B: Provider config file via CLI flag
cat > azure-credential.yaml <<EOF
clientId: abc-123
clientSecret: xyz-789
tenantId: def-456
subscriptionId: your-subscription-id
EOF

openmcf pulumi up \
  -f ops/azure/aks-cluster.yaml \
  -p azure-credential.yaml
```

### Method 2: Azure CLI Authentication (Local Development)

```bash
# Login with your personal account
az login

# No additional configuration needed
openmcf pulumi up -f ops/azure/aks-cluster.yaml
```

### Troubleshooting Azure Credentials

**Problem**: "Failed to authenticate"

```bash
# Verify you're logged in
az account show

# List available subscriptions
az account list

# Set correct subscription
az account set --subscription "My Subscription"
```

**Problem**: "Insufficient permissions"

```bash
# Check service principal roles
az role assignment list \
  --assignee <client-id> \
  --output table

# Add necessary role
az role assignment create \
  --assignee <client-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>
```

---

## Cloudflare Credentials

### Method 1: API Token (Recommended)

**Step 1**: Create API token in Cloudflare dashboard:

1. Go to Cloudflare Dashboard → My Profile → API Tokens
2. Click "Create Token"
3. Select template or create custom with needed permissions
4. Copy the token (you won't see it again)

**Step 2**: Use the token:

```bash
# Method A: Environment variable
export CLOUDFLARE_API_TOKEN="your-api-token-here"

openmcf pulumi up -f ops/cloudflare/r2-bucket.yaml

# Method B: Credential file (not commonly used, environment variable preferred)
```

### Method 2: Legacy API Key (Not Recommended)

```bash
export CLOUDFLARE_API_KEY="your-api-key"
export CLOUDFLARE_EMAIL="your-email@example.com"

openmcf pulumi up -f ops/cloudflare/r2-bucket.yaml
```

**Why not recommended**: API keys have account-wide access. API tokens can be scoped to specific permissions.

### Troubleshooting Cloudflare Credentials

**Problem**: "Authentication failed"

```bash
# Verify token is set
echo $CLOUDFLARE_API_TOKEN

# Test token with Cloudflare API
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN"
```

**Problem**: "Insufficient permissions"

- Check token permissions in Cloudflare dashboard
- Create new token with required permissions
- Ensure token isn't expired

---

## Kubernetes Cluster Credentials

When deploying to Kubernetes (using `*.Kubernetes` components), you need kubeconfig credentials.

### Method 1: Default Kubeconfig File

```bash
# OpenMCF automatically uses ~/.kube/config
openmcf pulumi up -f ops/k8s/postgres.yaml
```

### Method 2: Custom Kubeconfig Path

```bash
# Set custom kubeconfig
export KUBECONFIG=~/.kube/staging-cluster-config

openmcf pulumi up -f ops/k8s/postgres.yaml
```

### Method 3: Kubeconfig via CLI Flag

```bash
# Pass kubeconfig as provider config file
openmcf pulumi up \
  -f ops/k8s/postgres.yaml \
  -p ~/.kube/prod-cluster.yaml
```

### Getting Kubeconfig Files

**For GKE**:
```bash
gcloud container clusters get-credentials my-cluster \
  --region us-central1 \
  --project my-project
```

**For EKS**:
```bash
aws eks update-kubeconfig \
  --name my-cluster \
  --region us-west-2
```

**For AKS**:
```bash
az aks get-credentials \
  --resource-group my-rg \
  --name my-cluster
```

### Troubleshooting Kubernetes Credentials

**Problem**: "Unable to connect to cluster"

```bash
# Verify kubeconfig is valid
kubectl cluster-info

# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch context
kubectl config use-context my-cluster
```

---

## Other Providers

### MongoDB Atlas

```bash
# Environment variables
export MONGODB_ATLAS_PUBLIC_KEY="your-public-key"
export MONGODB_ATLAS_PRIVATE_KEY="your-private-key"

# Or via CLI flag (provider auto-detected from manifest)
openmcf pulumi up \
  -f ops/atlas/cluster.yaml \
  -p atlas-creds.yaml
```

### Snowflake

```bash
# Environment variables
export SNOWFLAKE_ACCOUNT="account-identifier"
export SNOWFLAKE_USER="username"
export SNOWFLAKE_PASSWORD="password"

# Or via CLI flag (provider auto-detected from manifest)
openmcf pulumi up \
  -f ops/snowflake/database.yaml \
  -p snowflake-creds.yaml
```

### Confluent Cloud

```bash
# Environment variables
export CONFLUENT_CLOUD_API_KEY="api-key"
export CONFLUENT_CLOUD_API_SECRET="api-secret"

# Or via CLI flag (provider auto-detected from manifest)
openmcf pulumi up \
  -f ops/confluent/kafka.yaml \
  -p confluent-creds.yaml
```

---

## CI/CD Credential Management

### GitHub Actions

```yaml
name: Deploy Infrastructure

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to AWS
        run: |
          openmcf pulumi up \
            -f ops/aws/vpc.yaml \
            --yes
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_DEFAULT_REGION: us-west-2
```

**Store credentials in**:
- Repository Secrets (Settings → Secrets and variables → Actions)
- Organization Secrets (for sharing across repos)
- Environment Secrets (for environment-specific credentials)

### GitLab CI

```yaml
deploy:
  stage: deploy
  script:
    - openmcf pulumi up -f ops/gcp/cluster.yaml --yes
  variables:
    GOOGLE_APPLICATION_CREDENTIALS: ${GCP_SERVICE_ACCOUNT_KEY}
  only:
    - main
```

**Store credentials in**:
- CI/CD Variables (Settings → CI/CD → Variables)
- Mark as "Protected" and "Masked"

### Jenkins

```groovy
pipeline {
    agent any
    environment {
        AWS_ACCESS_KEY_ID     = credentials('aws-access-key-id')
        AWS_SECRET_ACCESS_KEY = credentials('aws-secret-access-key')
    }
    stages {
        stage('Deploy') {
            steps {
                sh 'openmcf pulumi up -f ops/aws/vpc.yaml --yes'
            }
        }
    }
}
```

**Store credentials in**: Jenkins Credentials Manager

---

## Credential Storage Solutions

### 1. Password Managers (Recommended for Individuals)

**1Password**:
```bash
# Store credentials
op item create --category login \
  --title "AWS Prod Credentials" \
  aws_access_key_id[password]=AKIA... \
  aws_secret_access_key[password]=wJal...

# Retrieve and use
export AWS_ACCESS_KEY_ID=$(op item get "AWS Prod Credentials" --fields aws_access_key_id)
export AWS_SECRET_ACCESS_KEY=$(op item get "AWS Prod Credentials" --fields aws_secret_access_key)
```

**pass (Unix password manager)**:
```bash
# Store credentials
pass insert aws/prod/access_key_id
pass insert aws/prod/secret_access_key

# Retrieve and use
export AWS_ACCESS_KEY_ID=$(pass aws/prod/access_key_id)
export AWS_SECRET_ACCESS_KEY=$(pass aws/prod/secret_access_key)
```

### 2. Secret Managers (Recommended for Teams)

**AWS Secrets Manager**:
```bash
# Store credentials
aws secretsmanager create-secret \
  --name prod/gcp/service-account \
  --secret-string file://gcp-key.json

# Retrieve and use
aws secretsmanager get-secret-value \
  --secret-id prod/gcp/service-account \
  --query SecretString \
  --output text > /tmp/gcp-key.json

export GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcp-key.json
```

**HashiCorp Vault**:
```bash
# Store credentials
vault kv put secret/aws/prod \
  access_key_id=AKIA... \
  secret_access_key=wJal...

# Retrieve and use
export AWS_ACCESS_KEY_ID=$(vault kv get -field=access_key_id secret/aws/prod)
export AWS_SECRET_ACCESS_KEY=$(vault kv get -field=secret_access_key secret/aws/prod)
```

---

## Security Checklist

Before deploying to production:

- [ ] Credentials stored in secure location (not in code)
- [ ] Using least-privilege IAM policies
- [ ] Different credentials for dev/staging/prod
- [ ] Credentials rotated regularly (every 90 days)
- [ ] Service accounts used instead of personal accounts
- [ ] Temporary credentials used where possible (IAM roles, workload identity)
- [ ] CI/CD secrets marked as protected/masked
- [ ] Credential access logged and monitored
- [ ] Revocation plan in place for compromised credentials
- [ ] Team members only have access to credentials they need

---

## Common Mistakes to Avoid

### ❌ Committing Credentials to Git

```bash
# This is BAD - credentials in git history forever
git add aws-credentials.yaml
git commit -m "Add AWS credentials"  # DON'T DO THIS!
```

**If you accidentally commit credentials**:
1. Rotate credentials IMMEDIATELY
2. Use `git-filter-branch` or BFG Repo-Cleaner to remove from history
3. Force push (if you must, and if repository is private)
4. Assume credentials are compromised - rotate them

### ❌ Using Root/Admin Credentials

```bash
# DON'T use root AWS account credentials
# DON'T use GCP owner role
# DON'T use Azure global administrator

# DO create service accounts with minimal permissions
```

### ❌ Sharing Credentials Insecurely

```bash
# DON'T send credentials via:
# - Email
# - Slack/Teams
# - Text message
# - Unencrypted files

# DO use:
# - Secret managers
# - Encrypted password managers
# - Secure credential sharing tools (1Password, Vault)
```

### ❌ Never Rotating Credentials

```bash
# DON'T use the same credentials forever
# DO rotate every 90 days or when:
# - Team member leaves
# - Credentials may have been exposed
# - As part of regular security practice
```

---

## Quick Reference

### Environment Variables by Provider

These environment variables are automatically read by the underlying IaC providers (Pulumi/Terraform) when no explicit `-p` flag is provided.

**AWS**:
```bash
AWS_ACCESS_KEY_ID          # Required
AWS_SECRET_ACCESS_KEY      # Required
AWS_DEFAULT_REGION         # Required
AWS_SESSION_TOKEN          # Optional (for temporary credentials)
AWS_PROFILE                # Optional (use named profile)
```

**GCP**:
```bash
GOOGLE_APPLICATION_CREDENTIALS  # Path to service account JSON key file
GOOGLE_CLOUD_PROJECT            # Default project ID
```

**Azure**:
```bash
ARM_CLIENT_ID           # Service principal client ID
ARM_CLIENT_SECRET       # Service principal client secret
ARM_TENANT_ID           # Azure AD tenant ID
ARM_SUBSCRIPTION_ID     # Azure subscription ID
```

**Cloudflare**:
```bash
CLOUDFLARE_API_TOKEN    # Scoped API token (recommended)
CLOUDFLARE_API_KEY      # Global API key (legacy)
CLOUDFLARE_EMAIL        # Account email (with legacy key)
```

**Kubernetes**:
```bash
KUBECONFIG              # Path to kubeconfig file (default: ~/.kube/config)
KUBE_CONTEXT            # Specific context to use
```

**OpenFGA**:
```bash
FGA_API_URL             # OpenFGA server URL
FGA_API_TOKEN           # API token for authentication
FGA_CLIENT_ID           # Client ID (for OAuth2 client credentials)
FGA_CLIENT_SECRET       # Client secret (for OAuth2 client credentials)
```

**Confluent Cloud**:
```bash
CONFLUENT_CLOUD_API_KEY      # Confluent Cloud API key
CONFLUENT_CLOUD_API_SECRET   # Confluent Cloud API secret
```

**Snowflake**:
```bash
SNOWFLAKE_ACCOUNT       # Snowflake account identifier
SNOWFLAKE_REGION        # Snowflake region
SNOWFLAKE_USER          # Username
SNOWFLAKE_PASSWORD      # Password
```

**MongoDB Atlas**:
```bash
MONGODB_ATLAS_PUBLIC_KEY   # Atlas public key
MONGODB_ATLAS_PRIVATE_KEY  # Atlas private key
```

**Auth0**:
```bash
AUTH0_DOMAIN            # Auth0 tenant domain
AUTH0_CLIENT_ID         # Machine-to-machine client ID
AUTH0_CLIENT_SECRET     # Machine-to-machine client secret
```

**DigitalOcean**:
```bash
DIGITALOCEAN_TOKEN      # Personal access token
SPACES_ACCESS_KEY_ID    # Spaces access key (for object storage)
SPACES_SECRET_ACCESS_KEY # Spaces secret key (for object storage)
```

**Civo**:
```bash
CIVO_TOKEN              # Civo API token
```

### CLI Provider Config Flag

```bash
# Unified flag - provider type is auto-detected from manifest
-p, --provider-config <file>
```

The CLI automatically determines which provider credentials are needed based on your manifest's `apiVersion` and `kind`. For example:
- `aws.openmcf.org/v1` → AWS credentials expected
- `gcp.openmcf.org/v1` → GCP credentials expected
- `kubernetes.openmcf.org/v1` → Kubernetes config expected
- `openfga.openmcf.org/v1` → OpenFGA credentials expected

---

## Related Documentation

- [Pulumi Commands](/docs/cli/pulumi-commands) - Deploying with Pulumi
- [OpenTofu Commands](/docs/cli/tofu-commands) - Deploying with OpenTofu
- [Manifest Structure](/docs/guides/manifests) - Writing manifests
- [CLI Reference](/docs/cli/cli-reference) - Complete command reference

---

## Getting Help

**Found a credential issue?** Check the troubleshooting section for your provider above.

**Security concern?** Contact your security team immediately if credentials may be compromised.

**Need help?** [Open an issue](https://github.com/plantonhq/openmcf/issues) with details (never include actual credentials in issues!).

---

**Remember**: Treat credentials like passwords. Never share them insecurely, rotate them regularly, and use the minimum permissions necessary. Your infrastructure's security depends on it. 🔐


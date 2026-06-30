---
title: "GCP Provider Setup"
description: "Configure Google Cloud credentials for Planton deployments — service accounts, environment variables, and provider config files"
icon: "cloud"
order: 80
---

# GCP Provider Setup

This guide covers everything you need to authenticate Planton with Google Cloud Platform. It applies to all GCP deployment components: `GcpCloudSql`, `GcpGkeCluster`, `GcpGcsBucket`, `GcpCloudRun`, and others.

For a quick reference of all provider credentials, see [Credentials](./credentials).

## Prerequisites

- A GCP project with billing enabled
- The [gcloud CLI](https://cloud.google.com/sdk/docs/install) installed (recommended, not required)

## Authentication Methods

### Method 1: Application Default Credentials (Local Development)

The fastest way to get started. Authenticate with your Google account and the credential is stored locally:

```bash
gcloud auth application-default login

planton pulumi up -f ops/gcp/database.yaml
```

This opens a browser for authentication. The resulting credential is stored at `~/.config/gcloud/application_default_credentials.json` and is automatically picked up by Planton and the underlying IaC engines.

Use this for local development and experimentation. For production and CI/CD, use a service account.

### Method 2: Service Account Key File

Create a service account with specific permissions and download a JSON key file:

```bash
# Set the environment variable to point to the key file
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"

planton pulumi up -f ops/gcp/database.yaml
```

| Variable | Description |
|----------|-------------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to a service account JSON key file |
| `GOOGLE_CLOUD_PROJECT` | Default project ID (optional, usually inferred from the key) |

### Method 3: Provider Config File (`-p`)

Pass credentials using the `-p` flag with a YAML file. The GCP provider config requires the service account key as a base64-encoded string, matching the `GcpProviderConfig` Protocol Buffer definition:

```yaml
# gcp-credential.yaml
service_account_key_base64: "<base64-encoded-json-key>"
```

Generate the file:

```bash
# Encode the service account key
base64 -i /path/to/service-account-key.json | tr -d '\n' > /tmp/encoded-key.txt

# Create the provider config
echo "service_account_key_base64: $(cat /tmp/encoded-key.txt)" > gcp-credential.yaml

# Clean up
rm /tmp/encoded-key.txt
```

Deploy using the `-p` flag:

```bash
planton pulumi up -f ops/gcp/database.yaml -p gcp-credential.yaml
```

The CLI decodes the base64 string and sets the `GOOGLE_CREDENTIALS` environment variable with the raw JSON key content for the IaC engine subprocess.

**Fields in the provider config file:**

| Field | Required | Description |
|-------|----------|-------------|
| `service_account_key_base64` | Yes | Base64-encoded JSON service account key |

The field name uses `snake_case`, matching the protobuf definition at `apis/dev/planton/provider/gcp/provider.proto`.

### Method 4: Workload Identity (GKE)

When running on GKE, use Workload Identity to bind Kubernetes service accounts to GCP service accounts. No key files are needed:

```bash
# No credential configuration needed
# The GKE pod's service account automatically receives GCP permissions
planton pulumi up -f ops/gcp/database.yaml
```

This is the most secure method for workloads running on GKE.

## Creating a Service Account

### Step 1: Create the Service Account

```bash
gcloud iam service-accounts create planton-deployer \
  --display-name "Planton Deployer" \
  --project my-project-id
```

### Step 2: Grant IAM Roles

Grant roles based on what you are deploying:

```bash
PROJECT_ID="my-project-id"
SA_EMAIL="planton-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

# For Cloud SQL deployments
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.admin"

# For GKE deployments
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/container.admin"

# For GCS bucket deployments
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.admin"

# For Cloud Run deployments
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"
```

### Step 3: Create a Key File

```bash
gcloud iam service-accounts keys create ~/gcp-service-account-key.json \
  --iam-account="${SA_EMAIL}"
```

Store this key file securely. You will not be able to download it again.

## Least-Privilege IAM Roles

| Component | Minimum IAM Role |
|-----------|-----------------|
| `GcpCloudSql` | `roles/cloudsql.admin` |
| `GcpGkeCluster` | `roles/container.admin`, `roles/compute.networkAdmin` |
| `GcpGcsBucket` | `roles/storage.admin` |
| `GcpCloudRun` | `roles/run.admin` |
| `GcpDnsZone` | `roles/dns.admin` |

For broader deployments, start with `roles/editor` and narrow permissions as you identify the exact resources being created. Avoid `roles/owner` in production.

## Verifying Credentials

### Environment Variable Method

```bash
# Check which account is active
gcloud auth application-default print-access-token > /dev/null 2>&1 && echo "ADC configured" || echo "No ADC"

# Check service account key
cat $GOOGLE_APPLICATION_CREDENTIALS | jq '.client_email'
```

### Service Account Permissions

```bash
# List roles granted to a service account
gcloud projects get-iam-policy my-project-id \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:planton-deployer@*" \
  --format="table(bindings.role)"
```

## Troubleshooting

### "Could not find default credentials"

No credentials are configured:

```bash
# Option 1: Set up Application Default Credentials
gcloud auth application-default login

# Option 2: Set the environment variable
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/key.json"
```

### "Permission denied" or "403 Forbidden"

The service account lacks the required IAM roles:

```bash
# Check current project
gcloud config get-value project

# Check service account roles (see Verifying Credentials above)
# Add the missing role
gcloud projects add-iam-policy-binding my-project-id \
  --member="serviceAccount:planton-deployer@my-project-id.iam.gserviceaccount.com" \
  --role="roles/cloudsql.admin"
```

### "Failed to decode service account key from base64"

The `-p` config file contains an invalid base64 string:

```bash
# Verify the encoding is correct
echo "<your-base64-string>" | base64 --decode | jq '.client_email'
```

If decoding fails, re-encode the key file:

```bash
base64 -i /path/to/service-account-key.json | tr -d '\n'
```

Ensure there are no line breaks in the base64 string.

### Provider Config File Validation Error

- The field name must be `service_account_key_base64` (snake_case)
- The value must be a valid base64-encoded string
- The decoded content must be a valid GCP service account JSON key

## What's Next

- [Credentials](./credentials) — Quick reference for all providers
- [AWS Provider Setup](./aws-provider-setup) — Configure AWS credentials
- [Azure Provider Setup](./azure-provider-setup) — Configure Azure credentials
- [CI/CD Integration](./cicd-integration) — Use GCP credentials in pipelines

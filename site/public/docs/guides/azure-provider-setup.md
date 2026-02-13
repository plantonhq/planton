---
title: "Azure Provider Setup"
description: "Configure Azure credentials for OpenMCF deployments — service principals, environment variables, and provider config files"
icon: "cloud"
order: 90
---

# Azure Provider Setup

This guide covers everything you need to authenticate OpenMCF with Microsoft Azure. It applies to all Azure deployment components: `AzureAksCluster`, `AzureResourceGroup`, and others.

For a quick reference of all provider credentials, see [Credentials](./credentials).

## Prerequisites

- An Azure subscription
- The [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) installed (recommended, not required)

## Authentication Methods

### Method 1: Azure CLI Authentication (Local Development)

The fastest way to get started. Log in with your Azure account:

```bash
az login

openmcf pulumi up -f ops/azure/cluster.yaml
```

This opens a browser for authentication. The resulting token is used automatically by OpenMCF and the underlying IaC engines.

Use this for local development. For production and CI/CD, use a service principal.

### Method 2: Environment Variables

Set the standard Azure environment variables for service principal authentication:

```bash
export ARM_CLIENT_ID="abc-123-def-456"
export ARM_CLIENT_SECRET="your-client-secret"
export ARM_TENANT_ID="tenant-id-here"
export ARM_SUBSCRIPTION_ID="subscription-id-here"

openmcf pulumi up -f ops/azure/cluster.yaml
```

| Variable | Required | Description |
|----------|----------|-------------|
| `ARM_CLIENT_ID` | Yes | Service principal application (client) ID |
| `ARM_CLIENT_SECRET` | Yes | Service principal password/secret |
| `ARM_TENANT_ID` | Yes | Azure Active Directory tenant ID |
| `ARM_SUBSCRIPTION_ID` | Yes | Azure subscription ID |

### Method 3: Provider Config File (`-p`)

Pass credentials using the `-p` flag with a YAML file matching the `AzureProviderConfig` Protocol Buffer definition:

```yaml
# azure-credential.yaml
client_id: "abc-123-def-456"
client_secret: "your-client-secret"
tenant_id: "tenant-id-here"
subscription_id: "subscription-id-here"
```

Deploy using the `-p` flag:

```bash
openmcf pulumi up -f ops/azure/cluster.yaml -p azure-credential.yaml
```

The CLI validates the config file against the proto schema, then converts the fields to environment variables (`ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_TENANT_ID`, `ARM_SUBSCRIPTION_ID`) for the IaC engine subprocess.

**Fields in the provider config file:**

| Field | Required | Description |
|-------|----------|-------------|
| `client_id` | Yes | Service principal application (client) ID |
| `client_secret` | Yes | Service principal password/secret |
| `tenant_id` | Yes | Azure AD tenant ID |
| `subscription_id` | Yes | Azure subscription ID |

All field names use `snake_case`, matching the protobuf definition at `apis/org/openmcf/provider/azure/provider.proto`.

## Creating a Service Principal

### Step 1: Create the Service Principal

```bash
az ad sp create-for-rbac \
  --name "openmcf-deployer" \
  --role Contributor \
  --scopes /subscriptions/<subscription-id>
```

Output:

```json
{
  "appId": "abc-123-def-456",
  "displayName": "openmcf-deployer",
  "password": "your-client-secret",
  "tenant": "tenant-id-here"
}
```

Map the output to credentials:

| Output field | Credential |
|-------------|------------|
| `appId` | `ARM_CLIENT_ID` / `client_id` |
| `password` | `ARM_CLIENT_SECRET` / `client_secret` |
| `tenant` | `ARM_TENANT_ID` / `tenant_id` |

The `subscription_id` comes from your Azure subscription, not the service principal output.

### Step 2: Find Your Subscription ID

```bash
az account show --query id --output tsv
```

Or list all subscriptions:

```bash
az account list --output table
```

## RBAC Role Assignments

The `Contributor` role grants broad permissions to create and manage resources. For least-privilege access, assign specific roles:

| Component | Minimum Role | Scope |
|-----------|-------------|-------|
| `AzureAksCluster` | `Azure Kubernetes Service Contributor` | Subscription or resource group |
| `AzureResourceGroup` | `Contributor` | Subscription |

### Assigning a Role

```bash
az role assignment create \
  --assignee <client-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>
```

To scope to a specific resource group:

```bash
az role assignment create \
  --assignee <client-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group-name>
```

## Verifying Credentials

### Azure CLI

```bash
# Check current login
az account show

# Verify service principal can authenticate
az login --service-principal \
  --username <client-id> \
  --password <client-secret> \
  --tenant <tenant-id>

az account show
```

### Environment Variables

```bash
# Check that all four variables are set
echo "CLIENT_ID: $ARM_CLIENT_ID"
echo "TENANT_ID: $ARM_TENANT_ID"
echo "SUBSCRIPTION_ID: $ARM_SUBSCRIPTION_ID"
echo "CLIENT_SECRET: $([ -n "$ARM_CLIENT_SECRET" ] && echo 'set' || echo 'not set')"
```

## Troubleshooting

### "Failed to authenticate"

The service principal credentials are invalid or expired:

```bash
# Test authentication directly
az login --service-principal \
  --username $ARM_CLIENT_ID \
  --password $ARM_CLIENT_SECRET \
  --tenant $ARM_TENANT_ID
```

If this fails, the client secret may have expired. Create a new one:

```bash
az ad sp credential reset --id <client-id>
```

### "AuthorizationFailed" or "Insufficient privileges"

The service principal lacks the required RBAC role:

```bash
# Check current role assignments
az role assignment list \
  --assignee <client-id> \
  --output table

# Add the required role
az role assignment create \
  --assignee <client-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>
```

### "SubscriptionNotFound"

The `ARM_SUBSCRIPTION_ID` is incorrect or the service principal does not have access to that subscription:

```bash
# List subscriptions accessible to the service principal
az login --service-principal \
  --username $ARM_CLIENT_ID \
  --password $ARM_CLIENT_SECRET \
  --tenant $ARM_TENANT_ID

az account list --output table
```

### Provider Config File Validation Error

- Field names must use `snake_case`: `client_id`, not `clientId`
- All four fields (`client_id`, `client_secret`, `tenant_id`, `subscription_id`) are required
- Values must be non-empty strings

## What's Next

- [Credentials](./credentials) — Quick reference for all providers
- [AWS Provider Setup](./aws-provider-setup) — Configure AWS credentials
- [GCP Provider Setup](./gcp-provider-setup) — Configure GCP credentials
- [CI/CD Integration](./cicd-integration) — Use Azure credentials in pipelines

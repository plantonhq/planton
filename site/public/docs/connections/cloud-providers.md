---
title: "Cloud Providers"
description: "Connect your AWS, GCP, Azure, DigitalOcean, and Cloudflare accounts to deploy infrastructure through Planton"
icon: cloud
order: 20
tags:
  - Connect
  - Cloud Providers
  - AWS
  - GCP
  - Azure
---

# Cloud Providers

To deploy infrastructure through Planton, the platform needs permission to act in your cloud accounts. Cloud provider connections give Planton the credentials it needs to create, update, and manage resources — from VPCs and Kubernetes clusters to databases and DNS zones — while keeping those credentials encrypted and scoped to the environments you authorize.

## Choosing an Authentication Mode

Every cloud provider connection starts with a choice: how should Planton authenticate?

### Inline Credentials

Provide your access keys or service account credentials directly. Planton encrypts them at rest and injects them at deployment time. This is the fastest way to get started.

**When to use**: Development environments, proof of concept, small teams where simplicity outweighs the operational overhead of managing a Runner.

**Trade-off**: The credentials are stored (encrypted) in Planton's control plane. For organizations with strict policies about where credentials can reside, consider runner-delegated authentication instead.

### Runner-Delegated Authentication

Deploy a [Planton Runner](/docs/runner) in your own infrastructure and let it authenticate using the cloud environment's native identity system — AWS IRSA, GCP Workload Identity, or Azure Managed Identity. No credentials are stored in Planton's control plane. The Runner authenticates locally and executes deployments where the credentials already exist.

**When to use**: Production environments, compliance-sensitive workloads, organizations that require credentials to never leave their network perimeter.

**Trade-off**: Requires deploying and maintaining a Runner. See the [Runner documentation](/docs/runner) for deployment options.

### Cross-Account Trust (AWS Only)

Create an IAM role in your AWS account that trusts Planton's Runner. When a deployment runs, the Runner assumes that role using AWS STS — no long-lived access keys are exchanged. This approach combines the convenience of not managing inline keys with the security of role-based access.

**When to use**: AWS organizations with multi-account strategies, teams that follow AWS best practices for cross-account access.

**Trade-off**: Requires initial IAM role setup (automated via CloudFormation Quick Create) and a Runner for execution.

---

## AWS

AWS is the most fully featured cloud provider integration. It supports all three authentication modes and includes an automated setup flow using AWS CloudFormation.

### Connecting via the Web Console

1. Navigate to **Connections** and click the **AWS** card under Infrastructure.
2. **Name your connection** — choose a descriptive name like "AWS Production" or "aws-dev-sandbox". The slug is derived from the name ("AWS Production" becomes `aws-production`), because a slug is lowercase letters and digits joined by single hyphens, like my-app-2.
3. **Choose your authentication method**:
   - **Inline API Keys** — Enter your Access Key ID and Secret Access Key directly.
   - **Cross-Account Trust** — Planton generates a CloudFormation Quick Create link. Click it to open the AWS Console, review the stack, and create the IAM role. The role ARN is captured automatically.
   - **Self-Hosted Runner** — Select an existing Runner or create a new one. The Runner handles authentication using its environment credentials.
4. **Create the connection**.

<!-- SCREENSHOT: AWS connection wizard - auth method selection
  Page: /resource/connect/aws-credential/create
  Action: Show step 2 with three auth method cards visible
  Focus: The three authentication method cards
  Alt: AWS connection wizard showing Inline API Keys, Cross-Account Trust, and Self-Hosted Runner options
-->

### Connecting via the CLI

The CLI provides a browser-based CloudFormation flow for AWS:

```bash
# Interactive setup — opens browser for CloudFormation Quick Create
planton connect aws

# Specify details upfront
planton connect aws --name my-aws-prod --account-id 123456789012

# Select specific capabilities
planton connect aws --capabilities eks,vpc_networking,s3_storage

# Reconnect an existing credential with a new IAM role
planton connect aws --reconnect my-aws-prod
```

### AWS Capabilities

When connecting via CloudFormation, you can scope the IAM role to specific capabilities. This follows the principle of least privilege — grant only the permissions Planton needs for the resource types you plan to deploy.

Available capabilities: EKS, ECS, VPC networking, S3 storage, RDS databases, Lambda compute, Route 53 DNS, CloudWatch monitoring, ECR container registry, KMS encryption.

By default, all capabilities are granted. You can narrow the scope during initial setup or when reconnecting.

### What You Need

| Field | Description |
|-------|-------------|
| Account ID | Your 12-digit AWS account number |
| Access Key ID | 20-character key starting with AKIA (for inline auth) |
| Secret Access Key | 40-character secret (for inline auth) |
| Region | Default deployment region (defaults to us-west-2) |

For cross-account trust, the CloudFormation stack handles IAM role creation automatically. For runner-delegated auth, you need a Runner deployed with appropriate IAM permissions.

---

## Google Cloud Platform

GCP connections use a service account key to authenticate. Planton acts as the service account when creating and managing resources in your GCP project.

### Connecting via the Web Console

1. Navigate to **Connections** and click the **GCP** card under Infrastructure.
2. **Name your connection** and provide the service account key JSON file (base64-encoded).
3. **Create the connection**.

<!-- SCREENSHOT: GCP connection form
  Page: /resource/connect/gcp-credential/create
  Action: Show the GCP connection form with name and service account key fields
  Focus: The connection name and service account key upload fields
  Alt: GCP connection form showing name input and service account key JSON upload
-->

### What You Need

| Field | Description |
|-------|-------------|
| Service Account Key | A JSON key file for a GCP service account with appropriate IAM roles, base64-encoded |

### Preparing Your GCP Service Account

In the Google Cloud Console:

1. Create a service account in the project where Planton will manage resources.
2. Grant the service account the IAM roles needed for your resource types (e.g., Kubernetes Engine Admin for GKE clusters, Cloud SQL Admin for databases).
3. Create a JSON key for the service account.
4. Base64-encode the key file for use in Planton.

GCP connections support both inline and runner-delegated authentication. For runner-delegated auth, deploy a Runner with GCP Workload Identity configured.

---

## Microsoft Azure

Azure connections use a service principal (app registration) to authenticate. Planton acts as the service principal when managing resources in your Azure subscription.

<!-- SCREENSHOT: Azure connection form
  Page: /resource/connect/azure-credential/create
  Action: Show the Azure connection form with service principal credential fields
  Focus: The tenant ID, client ID, client secret, and subscription ID fields
  Alt: Azure connection form showing service principal credential fields for tenant, client, and subscription
-->

### Connecting via the Web Console

1. Navigate to **Connections** and click the **Azure** card under Infrastructure.
2. **Name your connection** and provide the service principal credentials.
3. **Create the connection**.

### What You Need

| Field | Description |
|-------|-------------|
| Tenant ID | Your Azure Active Directory tenant identifier |
| Subscription ID | The Azure subscription where resources will be managed |
| Client ID | The application (client) ID of your service principal |
| Client Secret | The client secret for authentication |

### Preparing Your Azure Service Principal

In the Azure Portal:

1. Register an application in Azure Active Directory.
2. Create a client secret for the application.
3. Assign the application the Contributor role (or more restrictive custom roles) on the subscription or resource groups where Planton will manage resources.

Azure connections support both inline and runner-delegated authentication. For runner-delegated auth, deploy a Runner with Azure Managed Identity configured.

---

## DigitalOcean

DigitalOcean connections use an API token to authenticate. Optionally, you can provide Spaces credentials for object storage operations.

### What You Need

| Field | Description |
|-------|-------------|
| API Token | A personal access token from the DigitalOcean control panel |
| Default Region | The region for new resources (required) |
| Spaces Access ID | Access key for DigitalOcean Spaces (optional, for object storage) |
| Spaces Secret Key | Secret key for DigitalOcean Spaces (optional) |

Generate an API token from the DigitalOcean control panel under **API > Tokens/Keys**. For Spaces, create a separate Spaces access key under **API > Spaces Keys**.

---

## Cloudflare

Cloudflare connections support two authentication schemes. The recommended approach uses a scoped API token; the legacy approach uses a global API key with your account email.

### What You Need (Recommended)

| Field | Description |
|-------|-------------|
| API Token | A scoped API token created in the Cloudflare dashboard |

### What You Need (Legacy)

| Field | Description |
|-------|-------------|
| API Key | Your global API key from the Cloudflare dashboard |
| Email | The email address associated with your Cloudflare account |

### R2 Storage Credentials

If you deploy a Cloudflare Worker whose code is a pre-built bundle stored in R2 (the Worker's `r2Bundle` source), the connection also needs an R2 key pair. Planton reads the bundle through R2's S3-compatible API, which the API token cannot sign for. A Worker that is only bound to an R2 bucket needs no R2 keys.

| Field | Description |
|-------|-------------|
| R2 Access Key ID | Access key of an R2 API token with Object Read on the bundle bucket (minimum 20 characters) |
| R2 Secret Access Key | Secret key paired with the access key (minimum 20 characters) |
| R2 Endpoint | Custom endpoint URL (optional; defaults to `https://<account_id>.r2.cloudflarestorage.com`) |

Create an R2 key pair in the Cloudflare dashboard under **R2 > Manage R2 API Tokens**.

Create API tokens in the Cloudflare dashboard under **My Profile > API Tokens**. Use the **Create Token** button to create a scoped token with only the permissions your deployments need.

---

## Common Patterns

### One Connection Per Account

Create separate connections for each cloud account. A "production AWS account" and a "development AWS account" should be two distinct connections, even if they belong to the same organization. This makes it straightforward to authorize the right credential for the right environment.

### Naming Conventions

Use names that identify both the provider and the purpose:

- `aws-production` — clear and specific
- `gcp-analytics-project` — identifies the GCP project's role
- `azure-staging` — identifies the environment intent

Avoid generic names like `aws1` or `my-cloud` — they become confusing as the number of connections grows.

### Credential Rotation

To rotate credentials, update the connection with new values. Planton uses the updated credentials for all future deployments. In-flight deployments continue with the credentials they started with. No downtime is required.

For AWS cross-account trust connections, rotation means updating the IAM role — use the `--reconnect` flag in the CLI to set up a new role.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [Environment Mappings](/docs/connections/environment-mappings) — Authorize connections for specific environments
- [Default Connections](/docs/connections/default-connections) — Configure automatic credential selection
- [Runner](/docs/runner) — Deploy a Runner for runner-delegated authentication

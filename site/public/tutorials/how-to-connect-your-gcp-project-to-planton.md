---
title: "How to Connect Your GCP Project to Planton"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "gcp"
  - "connect"
  - "provider-connection"
  - "getting-started"
category: "connect"
excerpt: "Grant Planton secure access to deploy and manage resources in your Google Cloud project."
---

# How to Connect Your GCP Project to Planton

Before Planton can deploy infrastructure in your Google Cloud project -- GKE clusters, Cloud SQL databases, GCS buckets, Cloud Run services, or anything else -- it needs a secure way to authenticate with GCP on your behalf. This tutorial walks you through creating that connection using three different approaches, so you can choose the one that fits your security requirements and workflow.

> **Note**: The Planton web console provides a guided wizard for this workflow.
> This tutorial uses the CLI/YAML approach where possible for stability and
> reproducibility. The browser OAuth path (Path B) is initiated through the
> web console. The console UI evolves frequently — always check it for the
> latest experience.

## What You Will Learn

- What a provider connection is and why Planton needs one for GCP
- How to create a GCP Provider Connection with a service account key (most common for teams and automation)
- How to create a GCP Provider Connection with browser OAuth through the web console (quickest path for individual developers)
- How to create a GCP Provider Connection with runner-delegated authentication via Workload Identity or Application Default Credentials (maximum security)
- How to set a default provider connection so your GCP deployments use it automatically

## Prerequisites

- [ ] A Google Cloud project where you have permission to create service accounts and manage IAM roles
- [ ] A Planton account with an organization already created
- [ ] The `planton` CLI installed and authenticated (run `planton auth login` if you have not done this yet)
- [ ] The `gcloud` CLI installed and authenticated (Path A and Path C use `gcloud` commands)

## Understanding Provider Connections

A provider connection tells Planton how to authenticate when deploying infrastructure in your cloud account. Credentials are stored as secrets in Planton's secrets manager and referenced by slug in the connection manifest -- the actual values never appear in the spec. For more on provider connections, see the [connections documentation](/docs/connections/cloud-providers).

### Three Authentication Approaches for GCP

GCP supports a different set of authentication options compared to AWS. Where AWS offers cross-account trust via CloudFormation, GCP offers browser OAuth via Google's consent flow. Here is how the three approaches compare:

| | Service Account Key | Browser OAuth | Runner-Delegated |
|---|---|---|---|
| **How it works** | You create a GCP service account, download its JSON key, store the key as a Planton secret, and the connection references that secret | You sign in with your Google account through the Planton web console; Planton stores a refresh token automatically | A self-hosted Planton Runner uses its environment credentials (GKE Workload Identity, Compute Engine metadata, Application Default Credentials) |
| **Security** | Service account key is long-lived; you manage rotation | Refresh token stored as a secret; access tokens are short-lived (~1 hour) | Credentials never leave your infrastructure; nothing stored in Planton |
| **Setup time** | 10-15 minutes | 2 minutes (web console only) | 30+ minutes (requires deploying a runner) |
| **Best for** | Teams, CI/CD automation, multi-user environments | Individual developers, quick evaluation, getting started | Maximum security, air-gapped environments, regulated industries |

All three approaches result in the same outcome: a provider connection that Planton can use to deploy GCP infrastructure on your behalf.

## Path A: Service Account Key (Inline Credentials)

This is the most common approach for teams and automation. You create a GCP service account with the IAM roles Planton needs, download the JSON key file, store it as a Planton secret, and create a connection that references it.

### Step 1: Create a GCP Service Account

Use the `gcloud` CLI to create a service account in your project:

```bash
gcloud iam service-accounts create planton-deployer \
  --display-name="Planton Infrastructure Deployer" \
  --project=your-project-id
```

Replace `your-project-id` with your actual GCP project ID.

### Step 2: Grant IAM Roles

The service account needs IAM roles for the GCP services you plan to manage through Planton. For a broad starting point, the `Editor` role works, though you should scope this down for production use.

```bash
gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:planton-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/editor"
```

For more granular control, grant only the roles required for your specific use case. For example, if you plan to manage Cloud SQL databases and GKE clusters:

```bash
gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:planton-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/cloudsql.admin"

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:planton-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/container.admin"
```

### Step 3: Create and Download a JSON Key

```bash
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=planton-deployer@your-project-id.iam.gserviceaccount.com
```

This creates a file named `sa-key.json` in your current directory containing the private key for the service account. Treat this file as a sensitive credential.

### Step 4: Store the Key as an Org-Level Secret

Use `planton secret set` with `--string`, reading the JSON key file from standard input, to store it as one single-value Planton secret. The slug you choose here (`gcp-sa-key` in this example) is what you will reference in the connection manifest.

```bash
planton secret set gcp-sa-key --string < ./sa-key.json
```

After the secret is stored in config-manager, delete the local key file so it does not remain on disk:

```bash
rm sa-key.json
```

### Step 5: Write the Connection Manifest

Create a file named `gcp-connection.yaml` with the following content:

```yaml
apiVersion: connect.planton.ai/v1
kind: GcpProviderConnection
metadata:
  name: my-gcp-connection
  org: your-org
spec:
  service_account_key:
    secret: gcp-sa-key
```

Replace these values with your own:

- **`metadata.name`**: A human-readable name for this connection. Planton generates a URL-safe slug from it.
- **`metadata.org`**: Your Planton organization slug.
- **`spec.service_account_key.secret`**: The slug of the secret you created in Step 4.

A few things to note about the GCP connection spec:

- GCP's inline mode requires only **one secret reference** (`service_account_key`). The service account JSON key bundles the project ID, client email, and private key into a single file, so there are no separate fields for project, region, or individual credentials.
- **`auth_mode`** defaults to `inline` when omitted. Including it explicitly (`auth_mode: inline`) is also fine, but not required.
- The IaC provider extracts everything it needs (project ID, credentials) from the key at execution time.

### Step 6: Apply the Manifest

```bash
planton apply -f gcp-connection.yaml
```

Planton validates that the referenced secret exists in your organization's config-manager and creates the connection.

### Step 7: Verify the Connection

Retrieve the connection to confirm it was created:

```bash
planton get GcpProviderConnection my-gcp-connection
```

The connection metadata and spec will display. The `service_account_key` field will show only the secret slug, not the actual key contents -- confirming that no sensitive data is stored in the connection resource itself.

## Path B: Browser OAuth (Sign in with Google)

This is the quickest way to connect your GCP project. You sign in with your Google account through the Planton web console, and Planton handles everything else -- no service accounts, no key files, no manual secret creation. The platform stores a refresh token automatically and exchanges it for short-lived access tokens when running operations.

Browser OAuth is initiated through the web console. Unlike the service account key and runner paths, this flow cannot be completed via CLI/YAML alone.

### When to Use Browser OAuth

Browser OAuth is ideal for individual developers evaluating Planton or working in development environments. For production workloads shared across a team, a service account key (Path A) or runner-delegated authentication (Path C) provides better operational characteristics -- service account connections do not depend on any individual user's Google account.

### How the Flow Works

1. In the Planton web console, navigate to your organization's **Connections** page and choose to create a new GCP connection.
2. Select **Sign in with Google** as the connection method.
3. A popup opens to Google's consent screen. Sign in with a Google account that has access to the GCP projects you want to manage.
4. After you authorize access, Planton exchanges the authorization code for tokens, stores the refresh token as a config-manager secret, and creates the connection. The console shows a list of GCP projects accessible to your account for confirmation.

### What Gets Created

The OAuth flow produces a connection resource equivalent to this:

```yaml
apiVersion: connect.planton.ai/v1
kind: GcpProviderConnection
metadata:
  name: my-gcp-oauth
  org: your-org
spec:
  auth_mode: browser_oauth
  oauth_user_email: you@example.com
  refresh_token:
    secret: my-gcp-oauth-oauth-refresh-token
```

- **`oauth_user_email`** is a display-only field showing which Google account authorized the connection. It is populated from the Google ID token during the OAuth callback.
- **`refresh_token.secret`** points to a config-manager secret that was created automatically. The secret slug follows the pattern `{connection-slug}-oauth-refresh-token`.

At execution time, Planton exchanges the refresh token for a short-lived access token (approximately one hour lifetime) and sends only the access token to the runner. The long-lived refresh token never leaves config-manager. If Google returns a rotated refresh token during the exchange, Planton updates the config-manager secret automatically.

### Verifying a Browser OAuth Connection

After the console flow completes, verify the connection via the CLI:

```bash
planton get GcpProviderConnection my-gcp-oauth
```

The spec will show `auth_mode: browser_oauth`, the `oauth_user_email` of the authorizing account, and the refresh token secret slug.

## Path C: Runner-Delegated Authentication (Workload Identity)

This approach provides the highest security posture. Instead of storing any credentials in Planton -- even as secret references -- you deploy a Planton Runner in your own GCP infrastructure and let it authenticate using Application Default Credentials (ADC). No sensitive values are ever transmitted to or stored in the Planton control plane.

On GKE, the runner can use Workload Identity to map its Kubernetes service account to a Google service account. On Compute Engine, it can use the instance's metadata service. The runner supports any credential source in GCP's standard ADC chain.

This approach requires a running Planton Runner. If you do not have one yet, the steps below walk you through enrollment.

### Step 1: Create a Runner Token

Create a runner token from the console's Runners area. The token is a named, revocable secret that authorizes runners to join your organization — it is never a runner's identity. A runner started with the token enrolls itself with the control plane on arrival and receives its own identity (a dedicated service account, an API key, and the mutual-TLS material for the runner-to-control-plane tunnel) — these are separate from GCP credentials.

### Step 2: Deploy the Runner with GCP IAM Bindings

Deploy the runner binary or container in your GCP environment. The runner needs IAM permissions for the GCP services you plan to manage through Planton.

How you bind IAM permissions depends on where the runner runs:

- **GKE with Workload Identity**: Create a Google service account with the required IAM roles. Bind it to the runner's Kubernetes service account using Workload Identity. The runner automatically receives credentials scoped to the Google service account's permissions.
- **Compute Engine**: The runner uses the VM's default service account or a custom service account attached to the instance. Grant the service account the required IAM roles.
- **Any environment with ADC**: Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to point to a service account key file, or use `gcloud auth application-default login`. The runner uses the standard ADC credential chain.

The runner uses GCP's Application Default Credentials, so any method that makes credentials available to a standard Google Cloud SDK client will work.

### Step 3: Start the Runner

Start the runner with your runner token and a name (for example, `my-gcp-runner`). On first start it enrolls itself — the registration appears in your Runners list at that moment — then establishes an outbound mTLS tunnel to the Planton control plane. It is now ready to receive and execute operations.

### Step 4: Write the Connection Manifest

Create a file named `gcp-runner-connection.yaml`:

```yaml
apiVersion: connect.planton.ai/v1
kind: GcpProviderConnection
metadata:
  name: gcp-runner
  org: your-org
spec:
  auth_mode: runner
  runner: my-gcp-runner
```

The key differences from the service account key manifest:

- **`auth_mode: runner`** tells Planton to delegate authentication entirely to the runner's environment.
- **`runner: my-gcp-runner`** is the name the runner enrolled under in Step 3. This tells Planton which runner to route operations to.
- There is no `service_account_key` field. In runner mode, secret reference fields are ignored even if provided.

### Step 5: Apply the Manifest

```bash
planton apply -f gcp-runner-connection.yaml
```

### Step 6: Verify the Connection

```bash
planton get GcpProviderConnection gcp-runner
```

The connection spec will show `auth_mode: runner` and the runner slug. No secret references appear because the runner handles all authentication locally using its environment credentials.

## Setting the Default GCP Connection

Once you have a connection, set it as the default for your organization so that GCP deployments use it automatically without needing to specify a connection in every manifest.

```bash
planton connection default set --provider gcp --connection my-gcp-connection
```

To set a different default for a specific environment (for example, using the runner-delegated connection for production):

```bash
planton connection default set --provider gcp --connection gcp-runner --env production
```

When a deployment needs a GCP connection, Planton resolves it in this order:

1. **Explicit connection** specified in the resource manifest (highest priority)
2. **Environment-level default** for the target environment
3. **Organization-level default** (lowest priority)

If none of these exist, the deployment fails with a clear error.

Verify your defaults are set correctly:

```bash
planton connection default list
```

## What to Do Next

Your GCP project is now connected to Planton. From here, you can:

- **Deploy GCP infrastructure** through the Cloud Catalog -- Cloud SQL databases, GKE clusters, GCS buckets, Cloud Run services, and more
- **Connect additional GCP projects** by repeating this process (organizations commonly have separate connections for different projects or environments)
- **Scope connection access** by creating provider connection authorizations that control which connections are allowed in which environments

When more tutorials become available, look for guides on deploying a PostgreSQL database on Google Cloud SQL and deploying your first backend service with zero-config CI/CD.

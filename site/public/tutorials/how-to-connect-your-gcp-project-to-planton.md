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

Before Planton can deploy infrastructure in your Google Cloud project -- GKE clusters, Cloud SQL databases, GCS buckets, Cloud Run services, or anything else -- it needs a secure way to authenticate with GCP on your behalf. This tutorial walks you through creating that connection using three different approaches, keyless first, so you can choose the one that fits your security requirements and workflow.

> **Note**: The keyless path (Path A) runs in the Planton web console, which
> builds the trust in your project with you. The other two paths use the
> CLI/YAML approach for stability and reproducibility. The console UI evolves
> frequently — always check it for the latest experience.

## What You Will Learn

- What a provider connection is and why Planton needs one for GCP
- How to create a keyless GCP Provider Connection in the web console, with no key stored anywhere (recommended)
- How to create a GCP Provider Connection with a stored service account key
- How to create a GCP Provider Connection with a runner's own identity via Workload Identity or Application Default Credentials (maximum isolation)
- How to set a default provider connection so your GCP deployments use it automatically

## Prerequisites

- [ ] A Google Cloud project where you administer IAM: a project Owner, or Role Administrator, Project IAM Admin, Workload Identity Pool Admin and Service Account Admin
- [ ] A Planton account with an organization already created
- [ ] The `planton` CLI installed and authenticated (run `planton auth login` if you have not done this yet) for Paths B and C
- [ ] The `gcloud` CLI installed and authenticated, or Cloud Shell (Path A's script, Paths B and C)

## Understanding Provider Connections

A provider connection tells Planton how to authenticate when deploying infrastructure in your cloud account, and what Planton may deploy there. Credentials, when a connection holds any, are stored as secrets in Planton's secrets manager and referenced by slug in the connection manifest -- the actual values never appear in the spec. For more on provider connections, see the [connections documentation](/docs/connections/cloud-providers).

Every connection records a **grant**: what Planton may deploy through it. **Everything Planton Can Deploy** covers every GCP kind in Planton's catalog; **Only What I Pick** covers the kinds you choose. The permissions come from the catalog, so the service account Planton uses holds a custom role with exactly what those kinds need, never a broad role like Owner or Editor.

### Three Authentication Approaches for GCP

| | Keyless | Stored Key | Runner Identity |
|---|---|---|---|
| **How it works** | Your project trusts Planton's identity issuer through Workload Identity Federation; Planton exchanges a short-lived token to act as a service account in your project | You create a GCP service account, download its JSON key, store the key as a Planton secret, and the connection references that secret | A self-hosted Planton Runner uses its environment credentials (GKE Workload Identity, Compute Engine metadata, Application Default Credentials) |
| **Security** | Nothing stored in Planton: no key, no Google token; the trust admits only this connection | Service account key is long-lived; you manage rotation | Credentials never leave your infrastructure; nothing stored in Planton |
| **Setup time** | About two minutes (one consent in your project, or one script) | 10-15 minutes | 30+ minutes (requires deploying a runner) |
| **Best for** | Everyone who can consent once in their project, or run one script there | Projects where you cannot create a workload identity pool | Air-gapped environments, regulated industries, teams that already run a runner |

All three approaches result in the same outcome: a provider connection that Planton can use to deploy GCP infrastructure on your behalf.

## Path A: Keyless, in the Web Console

This approach stores nothing in Planton. Your project gets a workload identity pool and an OIDC provider that trust Planton's identity issuer for this one connection, and a service account that only this connection may act as, holding the connection's own custom role. The console builds the trust with you before the connection exists, and creates the connection once Google confirms the trust.

### Step 1: Start the Connection

In the Planton web console, open **Connections** and click the **GCP** card under Infrastructure. **Name** the connection, for example `prod-gcp`, then choose **Keyless** (listed first, and recommended).

### Step 2: Say Where It Deploys

On the **Project** step, enter the **Project ID**, as the Google Cloud console's project picker shows it; Planton looks up the project number itself. The console suggests the **Service Account** `planton-provisioner@<project-id>.iam.gserviceaccount.com`, which the setup creates if it does not exist.

### Step 3: Choose What Planton May Deploy

On the **Permissions** step, keep **Everything Planton Can Deploy** (recommended), or choose **Only What I Pick** and tick the kinds you use. **Organization Access** (folders, the organization and billing accounts) and **Access Granting** (giving people and identities access across the project) stay off unless you turn them on. Click **Show the Role** to read exactly what the setup will create.

### Step 4: Set Up the Trust in Your Project

The **Set Up** step names everything it will build: the APIs it turns on, the identity pool, the provider, the service account and its roles. Build it in one of two ways:

- **Set It Up for Me** (the step's main button): sign in with a Google account that administers the project and approve Google's consent screen once. Planton checks that your account may make every change, turns on the APIs the trust and Planton need, creates the pool, the provider, the service account and the binding, and creates this connection's own custom role and grants it to the service account. It uses your approval only for this setup and then discards it: Google issues no refresh token, and nothing about your Google access is stored.
- **Prefer to Run It Yourself? Show the Setup Script**: run the script in Cloud Shell, or any shell with `gcloud` signed in to your project. It builds the same trust and prints one line; paste it into **The Line the Script Printed** and click **Check the Trust**.

Either way, Planton then asks Google whether it trusts this connection. A new provider can take a minute or two to spread inside Google, and the step waits for it. When **GCP Trusts This Connection** appears, click **Next**.

### Step 5: Review, Create, and Read the Verdict

On **Review & Create**, click **Create**. The next screen runs the connection's first Verify and says what it can deploy, as **"N of M GCP Kinds Can Deploy"** counted against your grant. The connection stores only the provider's name and the service account's email: no key, no Google token, nothing to rotate.

For what the setup creates in detail, running it again, and a second connection in the same project, see [Keyless Cloud Connections](/docs/connections/keyless-cloud-connections#google-cloud).

## Path B: Stored Key (a Service Account Key)

This approach stores a long-lived credential. You create a GCP service account with the role Planton's catalog renders for what you want it to deploy, download the JSON key file, store it as a Planton secret, and create a connection that references it.

### Step 1: Create a GCP Service Account

Use the `gcloud` CLI to create a service account in your project:

```bash
gcloud iam service-accounts create planton-deployer \
  --display-name="Planton Infrastructure Deployer" \
  --project=your-project-id
```

Replace `your-project-id` with your actual GCP project ID.

### Step 2: Grant the Connection's Role

The service account needs the permissions of what you want Planton to deploy, and nothing broader. In the console's GCP connection wizard, choose **Stored Key**, and on its **Permissions** step choose **Everything Planton Can Deploy** or **Only What I Pick**, then click **Show the Role**. The role is written in the form `gcloud iam roles create --file` takes: save it as `planton-role.yaml`, then create it in your project and grant it to the service account:

```bash
gcloud iam roles create planton_deployer --project=your-project-id --file=planton-role.yaml

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:planton-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="projects/your-project-id/roles/planton_deployer"
```

When the step also names predefined roles (a few permissions Google keeps out of custom roles), grant those to the service account the same way.

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
  grant:
    coverage: everything
  service_account_key:
    secret: gcp-sa-key
```

Replace these values with your own:

- **`metadata.name`**: A human-readable name for this connection. Planton generates a URL-safe slug from it.
- **`metadata.org`**: Your Planton organization slug.
- **`spec.service_account_key.secret`**: The slug of the secret you created in Step 4.
- **`spec.grant`**: What Planton may deploy through the connection: `coverage: everything` for every kind Planton's catalog offers, or `coverage: chosen` with the `kinds` you pick. Every cloud connection records one, and the role you granted in Step 2 should match it.

A few things to note about the GCP connection spec:

- GCP's stored-key mode requires only **one secret reference** (`service_account_key`). The service account JSON key bundles the project ID, client email, and private key into a single file, so there are no separate fields for project, region, or individual credentials.
- **`auth_mode`** defaults to `inline` when omitted. Including it explicitly (`auth_mode: inline`) is also fine, but not required.
- The IaC provider extracts everything it needs (project ID, credentials) from the key at execution time.

### Step 6: Apply the Manifest

```bash
planton apply -f gcp-connection.yaml
```

Planton validates that the referenced secret exists in your organization's config-manager and creates the connection. A manifest without `grant` is refused with a message that says how to add it.

### Step 7: Verify the Connection

Retrieve the connection to confirm it was created:

```bash
planton get GcpProviderConnection my-gcp-connection
```

The connection metadata and spec will display. The `service_account_key` field will show only the secret slug, not the actual key contents -- confirming that no sensitive data is stored in the connection resource itself.

## Path C: Runner Identity (Workload Identity)

This approach provides the highest security posture. Instead of storing any credentials in Planton -- even as secret references -- you deploy a Planton Runner in your own GCP infrastructure and let it authenticate using Application Default Credentials (ADC). No sensitive values are ever transmitted to or stored in the Planton control plane.

On GKE, the runner can use Workload Identity to map its Kubernetes service account to a Google service account. On Compute Engine, it can use the instance's metadata service. The runner supports any credential source in GCP's standard ADC chain.

This approach requires a running Planton Runner. If you do not have one yet, the steps below walk you through enrollment.

### Step 1: Create a Runner Token

Create a runner token from the console's Runners area. The token is a named, revocable secret that authorizes runners to join your organization — it is never a runner's identity. A runner started with the token enrolls itself with the control plane on arrival and receives its own identity (a dedicated service account, an API key, and the mutual-TLS material for the runner-to-control-plane tunnel) — these are separate from GCP credentials.

### Step 2: Deploy the Runner with GCP IAM Bindings

Deploy the runner binary or container in your GCP environment. The runner's identity needs the role Planton's catalog renders for what you want it to deploy: in the console's GCP connection wizard, choose **Runner Identity**, and on its **Permissions** step click **Show the Role**.

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
  grant:
    coverage: everything
  auth_mode: runner
  runner: my-gcp-runner
```

The key differences from the service account key manifest:

- **`auth_mode: runner`** tells Planton to delegate authentication entirely to the runner's environment.
- **`runner: my-gcp-runner`** is the name the runner enrolled under in Step 3. This tells Planton which runner to route operations to.
- There is no `service_account_key` field. In runner mode, secret reference fields are ignored even if provided.
- **`grant`** says what Planton may deploy through the connection, as in Path B.

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

To set a different default for a specific environment (for example, using the runner connection for production):

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

- **Deploy GCP infrastructure** through the Infra Catalog -- Cloud SQL databases, GKE clusters, GCS buckets, Cloud Run services, and more
- **Connect additional GCP projects** by repeating this process (organizations commonly have separate connections for different projects or environments)
- **Scope connection access** by creating provider connection authorizations that control which connections are allowed in which environments

When more tutorials become available, look for guides on deploying a PostgreSQL database on Google Cloud SQL and deploying your first backend service with zero-config CI/CD.

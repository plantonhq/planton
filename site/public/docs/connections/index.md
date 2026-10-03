---
title: "Connections"
description: "Manage credentials and integrations for cloud providers, source control, registries, state backends, and managed services"
icon: connect
order: 20
tags:
  - Connections
  - Credentials
  - Integrations
---

# Connect

Connect is where you bring your existing cloud accounts, source control, registries, and third-party services into Planton. Every deployment Planton runs on your behalf — whether it is creating an AWS VPC, building a container image, or storing Terraform/OpenTofu state — requires credentials to authenticate with external providers. Connect manages those credentials so they are stored securely, scoped to the right environments, and never exposed in logs or CI/CD pipelines.

## The Problem Connect Solves

In a typical DevOps setup, credentials are scattered. AWS keys live in CI/CD environment variables. GitHub tokens are duplicated across repositories. Container registry passwords are hardcoded in build scripts. When an engineer rotates a credential, they have to track down every place it was used.

Connect eliminates this by giving credentials a single home at the organization level. You create a credential once, authorize it for specific environments, and Planton handles injection at runtime. Engineers deploying infrastructure or services never see the underlying secrets — they select which connection to use, and the platform takes care of the rest.

## What You Can Connect

The Connections page in the web console organizes integrations into four categories:

### Infrastructure

Cloud providers where Planton deploys and manages resources on your behalf.

| Provider | What It Enables |
|----------|----------------|
| **AWS** | VPCs, EKS clusters, RDS databases, S3 buckets, Lambda functions, Route 53, and 40+ other resource types |
| **Google Cloud** | GKE clusters, Cloud SQL, GCS buckets, Cloud Run services, Cloud DNS, and more |
| **Azure** | AKS clusters, Azure SQL, Blob Storage, Container Instances, Application Gateways |
| **DigitalOcean** | Kubernetes clusters (DOKS), Droplets, managed databases, Spaces object storage |
| **Cloudflare** | DNS zones, CDN, Workers, R2 storage |
| **Kubernetes** | External clusters not created by Planton — for hybrid cloud, migration, or multi-cluster deployments |

For details on connecting each provider, see [Cloud Providers](/docs/connections/cloud-providers) and [Kubernetes Clusters](/docs/connections/kubernetes-clusters).

### DevOps Pipeline

Source control providers and artifact registries that power your build and deployment pipelines.

| Integration | What It Enables |
|-------------|----------------|
| **GitHub** | Repository access for Service Hub builds and pull request deployments — as a GitHub App, or as the sign-in the machine running Planton already holds |
| **GitLab** | Repository access and pipeline triggers (self-hosted GitLab supported) |
| **Container Registry** | Where builds push images — GCP Artifact Registry, AWS ECR, Azure Container Registry, JFrog Artifactory, and GitHub Container Registry, signing in through a connection you already trust or with stored keys |
| **NPM** | Private JavaScript/TypeScript package resolution during builds |
| **Maven** | Private Java artifact resolution during builds |
| **Cloudflare Wrangler** | R2 bucket for storing Cloudflare Worker script bundles during deployment |

For details, see [Git Providers](/docs/connections/git-providers) and [Container Registries](/docs/connections/container-registries).

### Infrastructure as Code

State backends that store the state files for Pulumi, Terraform, and OpenTofu deployments.

| Backend | What It Enables |
|---------|----------------|
| **Pulumi** | State storage via Pulumi Cloud, S3, GCS, or Azure Blob — used by all Pulumi-based infrastructure deployments |
| **Terraform** | State storage via S3, GCS, or Azure RM — used by Terraform and OpenTofu-based infrastructure deployments |

Every backend also names the key that encrypts its state -- Planton's key by default, or a passphrase, cloud KMS key or Vault key you hold, which an AWS, Google Cloud, Azure or Vault connection supplies. For details, see [State Backends](/docs/connections/state-backends) and [State Encryption](/docs/connections/state-encryption).

### Managed Services

Third-party platforms that Planton can provision resources on.

| Service | What It Enables |
|---------|----------------|
| **Auth0** | Identity and access management tenant integration |
| **OpenFGA** | Fine-grained authorization server connection |

Managed service credentials follow the same create-authorize-default workflow as cloud provider credentials. To connect a managed service, navigate to the **Managed Services** section of the Connections page and follow the setup wizard for your provider.

## How Authentication Works

When you create a connection, you choose how Planton authenticates with the provider. The right choice depends on where Planton runs and on your security requirements. One rule holds across every mode: a connection describes how to sign in; wherever a sign-in already exists — on the machine running Planton, in a runner's environment, in another connection — Planton references it and stores nothing.

### Inline Credentials

You provide the credentials directly — an API key, a service account key, a client secret. Planton encrypts and stores them. This is the most straightforward option and works well for development environments or organizations getting started.

**Best for**: Quick setup, development environments, small teams.

### Runner-Delegated Authentication

Instead of storing credentials in Planton, you deploy a [Runner](/docs/runner) in your own infrastructure and let it authenticate using the environment's native identity — AWS IAM Roles for Service Accounts (IRSA), GCP Workload Identity, or Azure Managed Identity. The credentials never leave your infrastructure.

**Best for**: Production environments, compliance-sensitive workloads, organizations that require credentials to stay within their network.

### Cross-Account Trust (AWS)

Specific to AWS. You create an IAM role in your AWS account that trusts Planton's Runner. The Runner assumes that role using AWS STS when it needs to act in your account. No long-lived access keys are exchanged.

**Best for**: AWS organizations with strict credential policies, multi-account setups, cross-account deployments.

### The Sign-In on This Machine

When Planton runs on your own laptop or a self-hosted server, that machine is usually already signed in — `aws sso login`, `gcloud auth login`, `az login`, `gh auth login`. Planton detects those sign-ins and turns each into a connection that stores nothing: a cloud connection the local runner resolves at deploy time, and a GitHub connection the control plane reads at the moment it reaches GitHub. Sign out on the machine and the connection stops working the same second.

**Best for**: A developer's laptop; a self-hosted server whose environment already carries the sign-ins it needs.

### A Connection You Already Trust

Some connections need no credential of their own because another connection already has one. A container registry connection names the connection it trusts — a GitHub connection for GHCR, a cloud connection for ECR, Artifact Registry, or ACR — and the runner derives the registry's push token from it at the moment a build pushes. The registry stores no key.

**Best for**: Every registry a connection you already have can reach.

Not every provider supports every mode. Cloud infrastructure providers (AWS, GCP, Azure) support inline, runner-delegated, and the sign-in on this machine; DigitalOcean and Cloudflare support inline and an exported API token. Cross-account trust is AWS-only. GitHub signs in as an App or as the sign-in on this machine. Container registries sign in through a connection you already trust or with stored keys. GitLab, state backends, and managed services use inline credentials.

## Authorization and Defaults

Creating a credential is only the first step. Two additional concepts control how credentials are used across your organization:

### Environment Authorization

A credential exists at the organization level, but it must be explicitly authorized for each environment where it can be used. This prevents a production AWS account from being accidentally used in development, or a staging database credential from being used in production.

You can authorize a credential for all environments at once (organization-wide scope) or for a specific list of environments. See [Environment Mappings](/docs/connections/environment-mappings) for details.

### Default Connections

When an Infra Component or Service deployment doesn't explicitly specify which credential to use, Planton looks for a default connection for that provider. Defaults can be set at the organization level (applies to all environments) or per-environment (overrides the organization default for that specific environment).

The resolution is straightforward: if an environment-level default exists, it wins. Otherwise, the organization-level default is used. If neither exists, the deployment fails with a clear error message telling you to configure a default.

See [Default Connections](/docs/connections/default-connections) for the full resolution logic and CLI commands.

## Getting Started

The typical setup flow for a new organization:

1. **Connect your cloud provider** — Start with your primary cloud account (AWS, GCP, or Azure). See [Cloud Providers](/docs/connections/cloud-providers).
2. **Connect your source control** — Link your GitHub or GitLab organization so Service Hub can access your repositories. See [Git Providers](/docs/connections/git-providers).
3. **Connect a container registry** — Set up a registry for your build artifacts. See [Container Registries](/docs/connections/container-registries).
4. **Configure a state backend** — Choose where infrastructure state files are stored. See [State Backends](/docs/connections/state-backends).
5. **Authorize for environments** — Decide which credentials can be used in which environments. See [Environment Mappings](/docs/connections/environment-mappings).
6. **Set defaults** — Configure default credentials so deployments work without explicit credential selection. See [Default Connections](/docs/connections/default-connections).

<!-- SCREENSHOT: Connections Mission Control page
  Page: /orgs/{org}/connections
  Action: Show the full Mission Control layout with all four categories visible
  Focus: The complete grid of provider cards organized by category
  Alt: Connections page showing Infrastructure, DevOps Pipeline, Infrastructure as Code, and Managed Services categories with provider cards
-->

## Managing Connections

### Using the Web Console

Navigate to **Connections** in the sidebar to see the Mission Control layout. Click any provider card to start the connection wizard. Existing connections are listed in the "Connected Providers" tab on the right side of the page.

### Using the CLI

The CLI provides commands for managing connection authorization and defaults:

```bash
# Connect an AWS account via browser-based CloudFormation flow
planton connect aws

# List all connection authorizations
planton connection authorization list

# Set a default connection for a provider
planton connection default set --provider aws --connection my-aws-account

# Test which connection would be resolved for a deployment
planton connection default resolve --provider aws --env production
```

## Related Documentation

- [Cloud Providers](/docs/connections/cloud-providers) — Connect AWS, GCP, Azure, and other cloud accounts
- [Git Providers](/docs/connections/git-providers) — Connect GitHub and GitLab for source code access
- [Container Registries](/docs/connections/container-registries) — Connect image and package registries
- [State Backends](/docs/connections/state-backends) — Configure Pulumi, Terraform, and OpenTofu state storage
- [Kubernetes Clusters](/docs/connections/kubernetes-clusters) — Connect external Kubernetes clusters
- [Environment Mappings](/docs/connections/environment-mappings) — Control which credentials can be used where
- [Default Connections](/docs/connections/default-connections) — Configure automatic credential selection
- [Runner](/docs/runner) — Deploy the secure execution agent for runner-delegated authentication
- [Security Overview](/docs/security) — Platform-wide security architecture and credential isolation model

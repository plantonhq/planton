---
title: "Container Registries"
description: "Connect container image registries and package repositories so Planton can push build artifacts and resolve private dependencies"
icon: connect
order: 40
tags:
  - Connect
  - Container Registry
  - Docker
  - NPM
  - Maven
---

# Container Registries

When Service Hub builds your code, it produces artifacts — container images for most services, or script bundles for Cloudflare Workers. Those artifacts need to be stored somewhere accessible to the deployment target. Container registry connections tell Planton where to push those artifacts and how to authenticate with the registry.

This page covers three types of registry connections: container image registries (for Docker/OCI images), NPM registries (for JavaScript/TypeScript packages), and Maven repositories (for Java artifacts).

## Container Image Registries

Container image registries store the Docker images that Service Hub produces during builds. Every service that uses Buildpacks or Dockerfile builds needs a container registry connection so the built image can be pushed.

### Two Ways a Registry Connection Signs In

A registry connection describes how Planton signs in to a registry; it never holds the result. It signs in one of two ways:

- **A connection you already trust** — the registry names a connection your organization already has: a GitHub connection for GHCR, an AWS connection for ECR, a GCP connection for Artifact Registry, an Azure connection for ACR. The connection stores no key. At the moment a build pushes, the runner derives a short-lived registry token from that connection's own credential — the sign-in your laptop already holds, or the cloud identity a hosted runner already has. Sign out of the trusted connection and the registry stops working the same moment.
- **Stored keys** — the registry references a secret you created: a service account key, IAM access keys, a service principal, a personal access token. The runner reads the secret at push time. This is the arm for JFrog Artifactory, and for any registry that sits outside the connections you have.

The Container Registry card on the Connections page opens on **You Already Trust These** — one card per registry the connections in your organization can reach — and **Use This Registry** writes a connection that stores nothing. **Set Up Manually** reaches the wizard, whose **Credential** step asks the same question: **A Connection You Already Trust** or **Stored Keys**.

Planton supports five container image registry providers:

### GCP Artifact Registry

Google's managed container registry. This is the recommended choice if your deployment targets run on GCP (GKE, Cloud Run).

| Field | Description |
|-------|-------------|
| GCP Project ID | The Google Cloud project that hosts the registry |
| GCP Region | The region where the registry is located |
| Repository Name | The Artifact Registry repository name |
| Credential | A GCP connection you already trust, or a service account key (JSON) with the Artifact Registry Writer role, stored as a secret |

### AWS Elastic Container Registry (ECR)

Amazon's managed container registry. The natural choice for ECS and EKS deployments.

| Field | Description |
|-------|-------------|
| Account ID | Your 12-digit AWS account number |
| Region | The AWS region where the ECR registry is hosted |
| Credential | An AWS connection you already trust, or an IAM access key pair with ECR push permissions, stored as secrets |

ECR refuses a push to a repository that doesn't exist, and Planton creates none: give each service a repository named by its image repository path, or declare a create-on-push repository creation template (`AwsEcrRegistrySettings`) for a prefix your services share, and ECR creates each one on its first push.

### Azure Container Registry (ACR)

Microsoft's managed container registry. Pairs with AKS and Azure Container Apps deployments.

| Field | Description |
|-------|-------------|
| Login Server | The registry's login server, `<name>.azurecr.io` |
| Credential | An Azure connection you already trust, or a service principal (tenant, client ID, client secret) with the AcrPush role, stored as secrets |

### JFrog Artifactory

A self-hosted or cloud-hosted universal artifact repository. Useful for organizations that standardize on JFrog for all artifact management.

| Field | Description |
|-------|-------------|
| Server URL | The Artifactory instance, e.g. `https://mycompany.jfrog.io/artifactory` |
| Repository Key | The Docker repository within Artifactory |
| Username | The Artifactory user or service account |
| Access Token | An Artifactory access token, stored as a secret |

### GitHub Container Registry (GHCR)

GitHub's container registry, tightly integrated with GitHub Packages.

| Field | Description |
|-------|-------------|
| Credential | The GitHub connection you already trust — the sign-in on the machine running Planton, which needs the `write:packages` scope — or a GitHub username with a personal access token carrying `write:packages`, stored as a secret |

### Connecting via the Web Console

1. Navigate to **Connections** and click the **Container Registry** card under Artifact Stores.
2. **You Already Trust These** lists the registries your connections reach — *GHCR as priya-dev · via GitHub connection github-account-priya-dev*, *ECR us-east-1 for 123456789012 · via AWS connection aws-profile-acme-dev*. Fill in anything the connection cannot know (a region, a repository name) on the card, and click **Use This Registry**. Planton writes the connection and proves it: the runner derives the credential once and the registry accepts it, or you read the exact sentence saying why not.
3. Or choose **Set Up Manually**: pick the provider, answer **Credential** (**A Connection You Already Trust** or **Stored Keys**), and provide the fields listed above.
4. On the completed screen — and on the connection's page any time after — **Verify This Registry** (or `planton connect registry verify <slug>` from a terminal) asks the registry to accept the derived credential right now, and then whether it would accept a push: ACR's push grant (AcrPush, not AcrPull), Artifact Registry's repository and the Writer role on it, JFrog's Docker repository key, a GHCR token's `write:packages` scope. The verdict says what was established: **Push Proven** names who may push where; **Push Not Proven Yet** is ECR, which answers a push grant only for a named repository — name one under **Verify a Repository** (or pass `--repository`) to prove it; **Push Not Verifiable** is honesty about what the registry cannot answer (a GitHub package owner's setting), with why and a door to where that setting lives; a refusal names what to create or grant, with the command that does it on its own line to copy; **Registry Could Not Be Reached** names the runner that tried, and asks you to retry once its network path is open. Each build asks ECR before it builds, too: a repository that does not exist (with no create-on-push template covering it) or an identity without the push actions fails the build at once, naming what to create or grant.

<!-- SCREENSHOT: You Already Trust These
  Page: /orgs/{org}/connections/create/container-registry
  Action: Show the registry cards the organization's trusted connections reach, with Use This Registry
  Focus: The GHCR card via a GitHub connection and the ECR card via an AWS connection
  Alt: Container registry create page listing the registries the organization's trusted connections reach, one confirm card each
-->

### Connecting via the CLI

```bash
# Which registries do my connections reach?
planton connect registry detect

# GHCR through the GitHub sign-in on this machine, without the confirmation prompt
planton connect registry detect --from github-account-priya-dev --yes

# ECR in a region through an AWS connection
planton connect registry detect --from aws-profile-acme-dev --region us-east-1
```

The verb writes the same credential-free connection the card writes, then proves it and reads **Registry Connection Ready — ghcr.io accepted the credential derived from GitHub connection github-account-priya-dev**.

### How Container Registries Are Used During Builds

When a Service Hub pipeline runs:

1. The build stage compiles your code and produces a container image.
2. The image is tagged with the commit SHA and pushed to the connected registry.
3. The deployment stage writes that image reference into your deployment manifests and applies them. The cluster or runtime then pulls the image itself — see the next section for how it signs in.

The image path follows the convention: `<registry-host>/<repository-path>:<commit-sha>`. Planton constructs this path automatically from the registry connection configuration.

### Pulling Private Images

A registry connection is how Planton reaches a registry: builds push with it. Pulling is the workload's own affair, because a pull has a different lifetime — a build signs in once and is done, while a cluster signs in every time a pod starts, for as long as the workload runs. Planton never copies a person's sign-in or a short-lived token into a cluster, so a Kubernetes workload pulls in one of three ways, all visible in its own manifest:

- **Declare nothing** when the cluster's own cloud identity reaches the registry: an EKS node role with the ECR read policy, a GKE node service account with the storage read scope, an AKS kubelet identity holding `AcrPull`, a DOKS cluster with registry integration on. No Secret, no field.
- **`pod.imageRegistries`** — the login declared on the workload itself: the registry server, the account, and a password that is only ever a `$secret/<slug>` reference. The workload module builds one `kubernetes.io/dockerconfigjson` Secret from it, in the workload's namespace, created and destroyed with the workload. One entry per registry server.
- **`pod.imagePullSecrets`** — a `KubernetesSecret` (its docker-registry arm) or a `KubernetesExternalSecret` (the **Docker Registry Pull Secret** preset, fed from your secrets manager) declared beside the workload, for a login many workloads share or for deploys that run without a Planton backend.

**The deploy fills the common case for you.** When a service deploys to a Kubernetes target, Planton looks at the service's registry connection and, when it holds a login a cluster can keep, writes that login onto the workload's `pod.imageRegistries` as a reference — in the open, beside the image it injected. The run's environment row says what it did or why it filled nothing:

| Registry connection | What the deploy fills |
|---|---|
| GHCR with a **pull token** (either credential arm) | The pull token — a read-only login is what a cluster should hold |
| GHCR with a stored personal access token | That token, by reference |
| GHCR through a GitHub connection, no pull token | Nothing — *add a read-only pull token to the registry connection, or declare the login on the workload's imageRegistries* |
| Artifact Registry with a stored service-account key | The key, by reference |
| Artifact Registry through a GCP connection | Nothing — a GKE cluster pulls with its node service account when granted on the repository |
| ACR with a stored service principal | The principal, by reference |
| ACR through an Azure connection | Nothing — an AKS cluster pulls with its kubelet identity when it holds `AcrPull` |
| ECR, on any arm | Nothing, ever — ECR issues only twelve-hour tokens; the cluster pulls with its own AWS identity or through a pull-through cache |
| JFrog Artifactory | The access token, by reference |

The fill yields to one thing only: an `imageRegistries` entry the workload already declares for the same server. A Secret named in `imagePullSecrets` never suppresses it. The password that lands on the manifest is a reference the runner resolves inside the cluster's own account; the value never touches Planton's records.

**The GHCR pull token.** A GHCR connection that pushes through your GitHub sign-in has no long-lived login to hand a cluster. Give it a **pull token** — a bot account's personal access token with only `read:packages`, stored as an organization secret — under `spec.githubContainerRegistry.pullToken` (the connection wizard's **Pull Token** step, the connection page's **Pull Token** section, or `planton apply`). Builds keep pushing through the sign-in; clusters pull with the token.

**Targets that pull only from their own cloud's registry.** Cloud Run and Cloud Run Jobs pull private images only from Artifact Registry (public Docker Hub and GHCR images excepted); App Runner only from ECR or ECR Public; Lambda container images only from ECR in the same Region. No field on the target changes that. Push to the provider's registry, or declare a pull-through repository that proxies the registry your image lives in: a `GcpArtifactRegistryRepo` in `REMOTE_REPOSITORY` mode, or `AwsEcrRegistrySettings.pullThroughCacheRules`. ECS pulls from ECR with the task execution role and from any other registry with the Secrets Manager credential the task definition declares.

**When it cannot work, it says so.** A workload referencing a Secret that has no value yet is refused before an infra job is created, naming the field and the resource. A literal password is refused at apply, naming the `$secret/` grammar. A pod that still cannot pull shows the kubelet's own `ImagePullBackOff` line followed by the remedy: declare the login on the workload, or a pull secret beside it, or pull from a registry the cluster's own identity reaches.

---

## NPM Registries

NPM registry connections allow Service Hub builds to resolve private JavaScript and TypeScript packages. If your monorepo or service depends on internal packages published to a private NPM registry, you need this connection so the build stage can install dependencies.

Planton supports three NPM registry providers:

| Provider | What You Need |
|----------|--------------|
| **GCP Artifact Registry** | GCP Project ID, region, repository name, service account key |
| **GitHub Packages** | GitHub username, personal access token with `read:packages` scope |
| **JFrog Artifactory** | Configuration interface being implemented |

### Connecting via the Web Console

1. Navigate to **Connections** and click the **NPM** card under DevOps Pipeline.
2. **Name your connection**, select your provider, and provide the credentials.
3. **Create the connection**.

---

## Maven Repositories

Maven repository connections allow Service Hub builds to resolve private Java artifacts. If your services depend on internal libraries published to a private Maven repository, you need this connection.

Planton supports three Maven repository providers:

| Provider | What You Need |
|----------|--------------|
| **GCP Artifact Registry** | GCP Project ID, region, repository name, service account key |
| **GitHub Packages** | GitHub username, personal access token with `read:packages` scope |
| **JFrog Artifactory** | Configuration interface being implemented |

### Connecting via the Web Console

1. Navigate to **Connections** and click the **Maven** card under DevOps Pipeline.
2. **Name your connection**, select your provider, and provide the credentials.
3. **Create the connection**.

---

## Cloudflare Worker Script Storage

Cloudflare Workers are deployed differently from container-based services — they use script bundles stored in R2 buckets rather than container images. If you deploy Cloudflare Workers through Service Hub, you need a Wrangler connection that specifies the R2 bucket for storing script bundles.

This connection appears under the DevOps Pipeline category as **Wrangler** in the Connections page.

| Field | Description |
|-------|-------------|
| R2 Access Key ID | Access key for R2 API operations |
| R2 Secret Access Key | Secret key for R2 API operations |
| R2 Endpoint | Custom endpoint URL (optional) |

---

## Choosing Your Registry

If you're starting fresh and choosing which registry to use, a practical guideline:

- **Deploying to AWS (EKS, ECS)?** Use ECR — the cluster pulls with its own identity, no pull login needed, and you avoid cross-cloud network transfer costs.
- **Deploying to GCP (GKE, Cloud Run)?** Use GCP Artifact Registry — same benefits for the GCP ecosystem, and the only registry Cloud Run pulls private images from.
- **Deploying to multiple clouds?** Use GitHub Container Registry or JFrog Artifactory for a cloud-neutral registry — clusters pull with a stored login (for GHCR, a read-only pull token) the deploy fills onto the workload for you.
- **Already using JFrog?** Continue using it — Planton connects to it the same way your existing tooling does.
- **Already connected a cloud or GitHub?** Its registry is one click away with nothing to paste — GHCR through your GitHub sign-in, ECR / Artifact Registry / ACR through the cloud connection you already have.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [Git Providers](/docs/connections/git-providers) — Connect source code repositories
- [CI/CD: Build Methods](/docs/ci-cd/build-methods) — How Planton builds container images

---
title: "Kubernetes Clusters"
description: "Connect external Kubernetes clusters for service deployments, infrastructure management, and Cloud Ops access"
icon: kubernetes
order: 60
tags:
  - Connect
  - Kubernetes
  - GKE
  - EKS
  - AKS
  - DOKS
---

# Kubernetes Clusters

Planton can create Kubernetes clusters for you through Infra Hub (GKE, EKS, AKS, DOKS). But many organizations already have clusters running — production clusters that predate Planton, clusters managed by a separate platform team, or clusters in environments that Planton doesn't manage directly.

Kubernetes cluster connections let you bring those existing clusters into Planton. Once connected, you can deploy services to them through Service Hub, manage workloads through Cloud Ops, and include them in your environment authorization model alongside cloud provider credentials.

## When to Use Kubernetes Cluster Connections

- **Existing production clusters** — You have clusters already running and want to deploy services to them through Planton without recreating them.
- **Hybrid cloud** — Some clusters are managed by Planton, others are managed externally. Connecting external clusters gives you a unified deployment surface.
- **Migration** — You're moving to Planton gradually and want to deploy new services to existing clusters before migrating the cluster management itself.
- **Multi-cluster architectures** — Your workloads span multiple clusters, some of which Planton created and some of which it didn't.

## Supported Providers

Kubernetes cluster connections support four managed Kubernetes providers. GKE, EKS, and AKS clusters accept short-lived tokens from the cloud's own identity system, so their connections store no Kubernetes credential: they name the cluster and **borrow the credentials of a cloud connection** you already have (a GCP, AWS, or Azure connection). The cloud connection's own sign-in method decides how the tokens are obtained — a stored key, the runner's own identity, or keyless, where nothing is stored at all.

When Planton creates a GKE, EKS, or AKS cluster through Infra Hub, it creates the cluster's connection for you, borrowing the same cloud connection that built the cluster. A cluster built keyless is therefore reached keyless too.

### Google Kubernetes Engine (GKE)

A GKE connection authenticates in one of two ways:

- **Borrow a GCP connection (recommended).** Name a GCP connection whose service account has access to the cluster. Every sign-in method of that connection works, including keyless.
- **A dedicated service account key.** For a cluster whose team prefers a key of its own over a shared GCP connection.

The console's form offers the dedicated key today; Planton writes the borrowing form for every GKE cluster it builds.

| Field | Description |
|-------|-------------|
| Cluster Endpoint | The API server endpoint of your GKE cluster |
| Cluster CA Data | Base64-encoded certificate authority data for TLS verification |
| GCP Connection | The GCP connection whose credentials reach the cluster (the borrowing option) |
| Service Account Key | The JSON key of a GCP service account with Kubernetes Engine access (the dedicated-key option) |

To find these values:

1. In the Google Cloud Console, navigate to your GKE cluster's details page.
2. The **Endpoint** and **Cluster CA certificate** are on the cluster details page.
3. Grant the service account behind your GCP connection (or a dedicated one, for a key) the `Kubernetes Engine Developer` or `Kubernetes Engine Admin` role.

### DigitalOcean Kubernetes (DOKS)

DOKS connections use a kubeconfig file.

| Field | Description |
|-------|-------------|
| Kubeconfig | The kubeconfig content for your DOKS cluster |

To get your kubeconfig:

1. In the DigitalOcean control panel, navigate to your Kubernetes cluster.
2. Download the kubeconfig file from the cluster details page, or use the CLI: `doctl kubernetes cluster kubeconfig show my-cluster`.

### Amazon EKS

An EKS connection names the cluster (name, endpoint, CA data, and region) and borrows an AWS connection whose IAM principal has access to the cluster. Every sign-in method of that AWS connection works, including keyless.

### Azure AKS

An AKS connection names an Entra-integrated cluster (endpoint and CA data) and borrows an Azure connection whose principal holds RBAC on the cluster. Every sign-in method of that Azure connection works, including keyless. A cluster without Entra integration has no tokens to borrow; connect it with its kubeconfig instead.

## Connecting via the Web Console

1. Navigate to **Connections** and click the **Kubernetes** card under Infrastructure.
2. **Name your connection** — use a name that identifies the cluster (e.g., "prod-gke-us-east", "legacy-doks-cluster").
3. **Select the provider** — GKE or DigitalOcean DOKS.
4. **Provide the credentials** listed above for your provider.
5. **Create the connection**.

<!-- SCREENSHOT: Kubernetes cluster connection form
  Page: /resource/connect/kubernetes-cluster-credential/create
  Action: Show the provider selector with GKE selected and credentials form visible
  Focus: Provider selection and GKE-specific credential fields
  Alt: Kubernetes cluster connection form showing GKE provider selected with endpoint, CA data, and service account key fields
-->

## Authentication Modes

Like cloud provider connections, Kubernetes cluster connections support inline and runner-delegated authentication:

- **Inline** — Provide the cluster credentials directly (endpoint, CA cert, service account key or kubeconfig). Simplest option.
- **Runner-delegated** — A [Planton Runner](/docs/runner) deployed with access to the cluster handles authentication. Useful when the cluster is in a private network and credentials should not leave the network perimeter.

## How Kubernetes Connections Are Used

Once connected, external clusters become deployment targets:

- **Service Hub** — When creating a service deployment target, you can select a connected external cluster alongside clusters that Planton created through Infra Hub.
- **Cloud Ops** — You can browse pods, stream logs, and exec into containers on connected clusters through the Cloud Ops interface, as long as a Runner with access to the cluster is configured.
- **Infra Hub** — Infra Components with Kubernetes catalog kinds (Helm charts, operators, custom resources) can target connected clusters.

## How a Cluster Pulls Images

A cluster connection is how Planton reaches the cluster; how the cluster reaches a container registry is the cluster's own identity, and it decides whether a private image needs a login at all:

- **EKS** pulls from ECR with the node role (the `AmazonEC2ContainerRegistryReadOnly` policy) or with IRSA. ECR issues only twelve-hour tokens, so this is the only way an EKS cluster pulls from ECR — a registry connection's keys are never written to the cluster.
- **GKE** pulls from Artifact Registry with the node service account when that account is granted on the repository (the default node scope `devstorage.read_only` is what allows it).
- **AKS** pulls from Azure Container Registry with the kubelet identity when it holds `AcrPull` on the registry.
- **DOKS** pulls from the DigitalOcean Container Registry cluster-wide when the cluster's registry integration is on.

For any other registry — GHCR, Docker Hub, a registry in another cloud, a self-managed cluster — the workload declares its login on its own manifest (`pod.imageRegistries`, or a Secret named in `pod.imagePullSecrets`), and the service deploy fills it from the registry connection when that connection holds a login a cluster can keep. See [Pulling Private Images](/docs/connections/container-registries#pulling-private-images).

## Practical Guidance

### Cluster Naming

Name connections by cluster identity and purpose, not by how they were created:

- `prod-gke-us-central1` — identifies the cluster's role, provider, and region
- `staging-doks-nyc1` — clear and specific
- `legacy-app-cluster` — useful during migration

### Credential Scope

The credentials you provide should have the minimum permissions needed for your use case:

- **Service deployments only**: The service account needs permissions to create and manage Deployments, Services, ConfigMaps, Secrets, and Ingress resources in the target namespaces.
- **Cloud Ops access**: Additionally needs pod list, log read, and exec permissions.
- **Full Infra Hub management**: Needs cluster-admin or equivalent broad permissions.

### Keep Credentials Current

Kubernetes credentials (especially kubeconfig tokens and service account keys) have expiration policies. Monitor connection health and rotate credentials before they expire to avoid deployment failures.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [Cloud Providers](/docs/connections/cloud-providers) — Connect cloud provider accounts for creating new clusters
- [CI/CD: Deployment Targets](/docs/ci-cd/deployment-targets) — How services target Kubernetes clusters
- [Operations](/docs/operations) — Runtime operations on connected clusters

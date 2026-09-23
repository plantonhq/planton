---
title: "Deployment"
description: "Enroll runners with a runner token, install, and deploy Planton Runner to Kubernetes, AWS ECS, GCP Cloud Run, or Azure Container Apps"
icon: server
order: 30
tags:
  - Runner
  - Deployment
  - Kubernetes
  - AWS
  - GCP
  - Azure
---

# Deployment

Getting a runner operational involves two phases: creating a runner token once, and starting runners with it — each runner enrolls itself on arrival. This page walks through the model and the target-specific deployment mechanics.

## Enrollment: the Runner Token

A runner token is a named, revocable secret that authorizes runners to join your organization. You create it once (in the console: Organization Settings → Runner Tokens, or `planton runner token create`); every runner started with it registers ITSELF with Planton on arrival and receives its own identity — a dedicated [service account](/docs/security/authentication-and-authorization#service-accounts), an API key, and mTLS certificates for the secure tunnel connection, delivered together in the runner's identity document.

Three properties define the model:

- **The token is never a runner's identity.** It cannot authenticate a single API call; its only power is admitting runners through the enrollment door. One token starting ten runners yields ten distinct identities, each revocable on its own.
- **Every arrival is attributed.** Each runner permanently records which token admitted it — who created that token and when is visible history, never an application-log archaeology exercise.
- **Re-admission is guarded by lineage.** A runner that loses its stored identity (a container that lost its disk) can re-enroll under its original name — but only through the same token that admitted it originally. A different token gets a closed door, so a leaked token can never hijack another token's runner.

The runner needs exactly two inputs at start: the token and the control-plane address. Everything else — organization, identity, endpoints, work-queue coordinates — arrives in the enrollment response, derived server-side from the token's record.

### Credential Security

The identity document contains sensitive material: the runner's private key and an API key. Both are minted server-side at enrollment and delivered only to the runner — they never pass through a person's browser, clipboard, or download folder. Planton does not store the private key or the raw API key after minting.

There is no separate credential-rotation ceremony: **re-enrolling IS rotating**. When a runner re-enrolls, it receives a fresh identity and every prior API key for that runner is revoked in the same act — a runner can never accumulate multiple live keys. If both a runner's identity and its admitting token are compromised, revoke the token, reset the runner's enrollment from its detail page (an audited action that clears its lineage), and let it re-enroll with a new token.

The deployment targets below store only the runner TOKEN in the target's secret store — the identity document is born where the runner runs (minted at enrollment, persisted on the runner's own writable storage) and never moves between machines:

| Deployment Target | Where the token lives |
|-------------------|-------------------|
| Kubernetes | Kubernetes Secret |
| AWS ECS | AWS Secrets Manager |
| GCP Cloud Run | GCP Secret Manager |
| Azure Container Apps | Container App built-in secret |

## Installation

To run a runner locally (for development or testing), install the binary:

```bash
planton runner install
```

This downloads the runner binary. To install a specific version:

```bash
planton runner install --version 0.1.5
```

## Starting a Runner Locally

For development or testing, you can start a runner directly from your machine. On first start with a runner token, the runner enrolls itself and persists its identity document into the local store (`~/.planton/`). On later starts it resolves its identity in this order:

1. `--credentials-file` flag or `PLANTON_RUNNER_CREDENTIALS_FILE` environment variable pointing to the JSON file
2. `PLANTON_RUNNER_CREDENTIALS` environment variable containing the raw JSON inline
3. Named runner from the local store (scans `~/.planton/`)

Once enrolled, start it by name:

```bash
planton runner start prod-runner --org acme
```

The runner loads its identity, establishes the secure tunnel, authenticates with the control plane, and begins accepting operations.

## Deployment Targets

For production use, deploy the runner as a long-running container. The CLI supports four deployment targets through `planton runner deploy`:

```bash
planton runner deploy prod-runner --token prt_...
```

This command ships the runner token into the target's secret store and deploys the runner container — the runner enrolls itself on first boot and appears in your Runners list the moment it joins. The runner does not need to exist yet: deploying a new name registers it on arrival. The command prompts you to select a deployment target and walks through the target-specific configuration; you can also specify `--image-tag` to pin a specific runner version. Re-deploys (image bumps) never re-ask for the token — they reuse the secret already sitting in the target.

### Kubernetes

Deploys the runner using a Helm chart. This is the most common deployment target.

**Prerequisites:**
- `kubectl` configured with access to the target cluster
- Helm 3 installed
- A runner token

**What gets deployed:**
- A Deployment running the runner container, with a writable volume where the runner persists the identity document it receives at enrollment
- A Secret containing the runner token
- A Kubernetes ServiceAccount (if using runner-delegated auth with Workload Identity or IRSA)

**When to use Kubernetes:**
- Your infrastructure operations target Kubernetes clusters
- You want the runner co-located with the workloads it manages
- You are already running Kubernetes and want to keep operations consolidated

### AWS ECS (Fargate)

Deploys the runner as a Fargate task in Amazon ECS. The runner runs as a single container with no EC2 instances to manage.

**Prerequisites:**
- AWS credentials with permissions to create ECS resources
- A VPC with subnets that have outbound internet access (for the tunnel connection)

**When to use ECS:**
- Your infrastructure is primarily on AWS
- You want a serverless deployment without managing Kubernetes
- You need the runner to have AWS IAM role-based access via ECS task roles

### GCP Cloud Run

Deploys the runner as an always-on Cloud Run service. Cloud Run's minimum instance count is set to ensure the runner stays warm for tunnel connectivity.

**Prerequisites:**
- GCP credentials with permissions to deploy Cloud Run services
- A project with the Cloud Run API enabled

**When to use Cloud Run:**
- Your infrastructure is primarily on GCP
- You want a managed container runtime without Kubernetes overhead
- You need the runner to use GCP Workload Identity for authentication

### Azure Container Apps

Deploys the runner as an always-on Container App. Similar to Cloud Run, minimum replicas are configured to maintain the tunnel connection.

**Prerequisites:**
- Azure credentials with permissions to create Container App resources
- A Container Apps Environment in your target region

**When to use Container Apps:**
- Your infrastructure is primarily on Azure
- You want a managed container runtime with Azure-native identity integration
- You need the runner to use Managed Identity for authentication

## Declarative Deployment from the Catalog

Every deployment target above also exists as a first-class catalog component, so a runner can be declared in a manifest, composed into an infra-chart beside the network it serves, and managed through the same deploy pipeline as everything else:

| Catalog Kind | Substrate |
|--------------|-----------|
| `AwsPlantonRunner` | AWS ECS (Fargate) |
| `GcpPlantonRunner` | GCP Cloud Run |
| `AzurePlantonRunner` | Azure Container Apps |
| `KubernetesPlantonRunner` | Any Kubernetes cluster (the official Helm chart) |

Each kind's spec carries the runner token as a managed-secret reference — on Planton, the platform mints a token and writes it at exactly that reference before the infrastructure applies, so declaring a runner is genuinely one step:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPlantonRunner
metadata:
  name: vpc-runner
spec:
  region: us-central1
  token: $secret/vpc-runner-token
```

The deployed runner enrolls itself on first boot and appears in your Runners list the moment it joins — exactly the same arrival story as a CLI deploy. Each kind models its substrate's real placement surface (subnets and task roles on AWS, direct VPC egress and service accounts on GCP, the Container App Environment on Azure, the namespace and chart values on Kubernetes); see each component's catalog page for the full specification.

## Default Runner

Once a runner is deployed, you can make it your organization's default runner. The default serves live cloud operations (a connection's verify, `planton kubectl`, resource browsing) for every [connection](/docs/connections) that names no runner of its own.

```bash
planton runner set-default prod-runner
```

Deploys do not consult the default: a deploy runs on the runner its connection names, else on a Planton-hosted runner. To run deploys on this runner, name it on the connection and keep your organization's state in a backend the runner can reach. [Where Your Work Runs](/docs/runner/planton-hosted-runners#where-your-work-runs) lays out every order, and [Running Deploys on Your Own Runner](/docs/runner/planton-hosted-runners#running-deploys-on-your-own-runner) walks through the move.

### Managing the Default

```bash
# Show where live cloud operations run for your organization
planton runner get-default

# Clear the default
planton runner unset-default
```

## Verifying Connectivity

After deploying a runner, verify that it has connected successfully. The simplest way is to run a Cloud Ops command that routes through the runner:

```bash
planton kubectl get pods --connection my-k8s-connection -n default
```

If the runner is connected and the connection is configured correctly, you will see the pod listing from your cluster. If the runner is not connected, you will receive a connection error indicating the runner is unreachable.

## Runner Management

### Listing Runners

```bash
# List runners in your organization
planton runner list

# Include each runner's reported capabilities
planton runner list --capabilities
```

### Viewing Runner Details

```bash
planton runner get prod-runner
```

### Deleting a Runner

```bash
planton runner delete prod-runner
```

This removes the runner registration and its auto-provisioned service account (which cascades to revoke all API keys). Any deployed runner using those credentials will lose both tunnel connectivity and control-plane authentication. The deployed container or service must be cleaned up separately.

## Related Documentation

- [Runner Overview](/docs/runner) — What Runner is and why it exists
- [Security Model](/docs/runner/security-model) — Credential isolation, authentication modes, and trust boundaries
- [Connections](/docs/connections) — Managing connections that route through runners

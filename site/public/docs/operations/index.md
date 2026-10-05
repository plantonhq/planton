---
title: "Operations"
description: "Real-time operations gateway for inspecting, debugging, and managing deployed infrastructure across Kubernetes, AWS, GCP, and Azure"
icon: cloud
order: 50
tags:
  - Operations
  - Operations
  - Kubernetes
  - Multi-Cloud
---

# Cloud Ops

Cloud Ops is Planton's real-time operations gateway. It lets developers and operators inspect, debug, and manage deployed infrastructure across Kubernetes, AWS, GCP, and Azure — from the web console and CLI — without distributing credentials or opening firewall rules.

## Why Cloud Ops Exists

Infra Hub provisions infrastructure. Service Hub deploys applications. But once everything is running, a fundamental question remains: **how do you see what's happening?**

Without a unified operations layer, teams face a fragmented reality:

- **Developers need kubectl access** to view pods, read logs, and debug failures. This means distributing kubeconfig files, managing Kubernetes RBAC, and teaching every developer CLI commands for each cluster.
- **Operators need cloud console access** to inspect EC2 instances, check storage buckets, or verify VM status. This means maintaining IAM credentials per operator per cloud account per environment.
- **Security teams worry about credential sprawl.** Every developer with direct cluster access holds credentials that could be misused or leaked. Every operator with console access adds to the blast radius.
- **Visibility is fragmented** across cloud provider consoles, terminal sessions, and monitoring tools. There is no single place to answer "what's running for this deployment?"

Cloud Ops solves this by routing all operational requests through Planton's secure tunnel infrastructure. Developers and operators interact with Cloud Ops through the web console or CLI. Cloud Ops resolves the correct credentials, routes the request to the appropriate [Planton Runner](/docs/runner) deployed in customer infrastructure, and the runner executes the operation using local credentials. Cloud credentials never leave the customer's environment.

## Two Ways to Access

Every Cloud Ops operation supports two access modes, designed for two different audiences.

### Developer Mode

Developers think in terms of their deployments: "show me the pods for my API." They specify their Infra Component — the deployment they care about — and Cloud Ops resolves everything else.

When a developer targets an Infra Component, the system:

1. Looks up the Infra Component in Infra Hub to find its cluster, namespace, and connection
2. Verifies the developer has access to that specific Infra Component
3. Routes the operation to the correct runner, scoped to the Infra Component's namespace

The developer never needs to know which cluster their deployment runs on, what namespace it lives in, or which credential connects to it. They also cannot access resources outside their deployment's scope — the namespace and connection are derived server-side and cannot be overridden.

**In the CLI**, developer mode uses the `--resource` flag:

```bash
planton kubectl get pods -r my-org/prod/KubernetesDeployment/payments-api
```

**In the web console**, developer mode is the default. Navigate to any Infra Component and open the Kubernetes tab to see its live state.

### Admin Mode

Infrastructure administrators think in terms of clusters and accounts: "show me everything in the production namespace" or "list all EC2 instances in us-east-1." They specify a provider connection (credential) and supply provider-specific parameters directly.

Connection resolution follows a fallback chain:

1. If a specific connection is provided, use it directly
2. If only an environment is provided, use the environment-level default connection
3. If neither is provided, use the organization-level default connection

For details on how default connections are configured, see [Connections > Default Connections](/docs/connections/default-connections).

**In the CLI**, admin mode uses the `--connection` flag:

```bash
planton kubectl get pods --connection k8s-prod-cluster -n kube-system
```

### How the Two Modes Compare

| Aspect | Developer Mode | Admin Mode |
|--------|---------------|------------|
| **Target audience** | Service developers | Infrastructure administrators |
| **Mental model** | "Show me pods for my deployment" | "Show me pods in kube-system" |
| **Scope** | Single deployment's resources | Any resources accessible via the connection |
| **Namespace/region** | Derived server-side (cannot override) | Supplied by the caller |
| **Authorization** | On the Infra Component | On the provider connection |

Both modes use the same operations, the same tunnel infrastructure, and the same security model. The access mode determines how Cloud Ops resolves the target — not what operations are available.

## Supported Providers

### Kubernetes (Full Operations)

Kubernetes has the deepest Cloud Ops integration, covering the full spectrum of day-2 operations:

- **Pod management** — Find pods by their managing resource (Deployment, StatefulSet, DaemonSet), view status, restart count, and container details
- **Log streaming** — Stream logs in real time with filtering by container, time range, tail lines, and content search
- **Container exec** — Open an interactive shell in any running container from the browser or CLI
- **Resource browsing** — List, describe, and view YAML for any Kubernetes resource in a namespace
- **Resource editing** — Modify Kubernetes resource YAML directly from the browser or CLI
- **Resource deletion** — Delete resources with confirmation
- **Namespace graph** — Visualize all resources in a namespace as a directed acyclic graph showing relationships

Both developer mode and admin mode are supported for all Kubernetes operations.

See [Kubernetes Operations](/docs/operations/kubernetes-operations) for the full reference.

### AWS, GCP, and Azure (Resource Browsing)

Cloud Ops extends beyond Kubernetes with resource listing operations across three major cloud providers:

| Provider | Operations |
|----------|-----------|
| **AWS** | List EC2 instances (with filters), list S3 buckets |
| **GCP** | List Compute Engine instances, list Cloud Storage buckets |
| **Azure** | List Virtual Machines, list Blob Storage containers |

These operations currently support admin mode only. Developer mode for non-Kubernetes providers will be available as the infra catalog expands to cover more resource kinds.

See [Resource Browser](/docs/operations/resource-browser) for the full reference.

## How It Works

Every Cloud Ops operation follows the same path, regardless of provider or access mode:

1. **Request arrives** — The web console or CLI sends the operation request to the Cloud Ops service
2. **Context resolution** — Cloud Ops determines the access mode and resolves the provider connection and any provider-specific parameters (namespace, region, project)
3. **Authorization** — Cloud Ops verifies the caller has access (to the Infra Component in developer mode, or to the provider connection in admin mode)
4. **Tunnel routing** — Cloud Ops routes the request through the secure tunnel to the [Planton Runner](/docs/runner) bound to the resolved connection
5. **Runner execution** — The runner executes the operation against the cloud provider API using local credentials
6. **Response** — Results flow back through the tunnel to the caller

<!-- SCREENSHOT: Cloud Ops architecture diagram
  Page: /docs/operations
  Action: Show the request flow from CLI/Console through CloudOps to Runner
  Focus: Full architecture diagram
  Alt: Diagram showing the Cloud Ops request flow: CLI or Web Console sends request to Cloud Ops service, which resolves credentials and routes through the secure tunnel to a Planton Runner in customer infrastructure
-->

This architecture means:

- **Credentials stay in customer infrastructure.** The runner handles all cloud API calls locally. Cloud Ops never sees, stores, or caches credentials.
- **No firewall rules needed.** The runner establishes an outbound connection to the tunnel server. No inbound ports need to be opened in customer infrastructure.
- **Low latency.** The tunnel adds approximately 1-5ms per request. For streaming operations (logs, exec), the overhead is negligible.

## Related Documentation

- [Kubernetes Operations](/docs/operations/kubernetes-operations) — Pod management, log streaming, container exec, resource browsing
- [Resource Browser](/docs/operations/resource-browser) — Multi-cloud listing of provider resources in AWS, GCP, and Azure
- [Runner](/docs/runner) — The secure execution agent that Cloud Ops routes operations through
- [Connections](/docs/connections) — Credential and integration management, including default connection resolution

# KubernetesAgentSandboxWarmPool

Declares a Kubernetes SIG Apps agent-sandbox `SandboxWarmPool`: a number of sandboxes from one template kept running and ready, so a claim gets one in the time it takes to hand it over. The spec mirrors the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxWarmPool field for field (pinned to agent-sandbox v1.0.7, the release `KubernetesAgentSandboxController` installs), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One SandboxWarmPool** in the target namespace, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The agent-sandbox controller (extensions on) then keeps the pool's members: Sandboxes named `<pool>-<random>`, each running the template's pod. A claim takes a ready member (which then belongs to the claim), and the controller starts a replacement. Deleting the pool deletes its unclaimed members; claimed sandboxes stay with their claims.

## How a Product Uses It

The pool's name and namespace are exactly what agent-sandbox's SDKs take: `create_sandbox(warmpool="<pool>", namespace="<namespace>")` creates a claim against it, waits for it to bind, and reaches the sandbox through the controller's router. Declare a `KubernetesAgentSandboxClaim` for a claim that should exist as long as its declaration does.

## Size, Autoscaling and Rollouts

- **`replicas`** is how many ready members the pool keeps. Unset, the cluster writes 1 when the pool is created and Planton never writes it again, so an autoscaler can own it.
- **The scale subresource** (`.spec.replicas`, `.status.replicas`, `.status.selector`) lets a HorizontalPodAutoscaler or KEDA scale the pool like a Deployment.
- **`updateStrategy.type`**: `OnReplenish` (default) replaces members that no longer match their template only as they are claimed; `Recreate` replaces them at once.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `namespace` | Namespace the pool, its template, its members and its claims share |
| `labels`, `annotations` | The SandboxWarmPool object's own metadata |
| `replicas` | How many ready members to keep (unset: an autoscaler owns it) |
| `sandboxTemplateRef.name` | The template (a KubernetesAgentSandboxTemplate reference or a literal name) |
| `updateStrategy.type` | `OnReplenish` or `Recreate` |

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_warm_pool_name` | Name of the created SandboxWarmPool (equals `metadata.name`), which claims and agent-sandbox's SDKs reference |
| `namespace` | Namespace the SandboxWarmPool was created in |

## Prerequisites

- `KubernetesAgentSandboxController` on the cluster, with extensions on.
- A `KubernetesAgentSandboxTemplate` in the same namespace.
- The target namespace (see `KubernetesNamespace`).

See [GUIDE.md](GUIDE.md) for the decisions behind a good pool and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

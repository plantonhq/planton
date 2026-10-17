# KubernetesAgentSandboxClaim

Declares a Kubernetes SIG Apps agent-sandbox `SandboxClaim`: a request for one sandbox from a warm pool, bound to a ready member when one exists, or to a fresh sandbox from the pool's template. The spec mirrors the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxClaim field for field (pinned to agent-sandbox v1.0.7, the release `KubernetesAgentSandboxController` installs), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One SandboxClaim** in the target namespace, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The agent-sandbox controller (extensions on) then binds the claim:

- **Warm**: a ready member of the pool (named `<pool>-<random>`) becomes the claim's sandbox, and the pool starts a replacement.
- **Cold**: when the claim sets environment variables or volume claim templates, or no member is ready, a sandbox named after the claim starts from the pool's template.

The sandbox's name is in the claim's status (`sandbox.name`). The sandbox belongs to the claim: deleting the claim deletes it, with its volumes.

## Declared Claims and Runtime Claims

Products usually create claims at runtime through agent-sandbox's SDKs. Declare one here when a sandbox from a pool should exist as long as the declaration does: a standing sandbox for a named user or team, a demo environment, a canary that proves the pool hands out working sandboxes.

## What a Claim May Ask For

| Field | Effect | Warm or cold |
|-------|--------|--------------|
| `additionalPodMetadata` | Labels (in the controller's allowed label domains) and annotations on the sandbox's pod | warm |
| `env` | Literal environment variables, needing the template's `envVarsInjectionPolicy` | cold |
| `volumeClaimTemplates` | Volumes the sandbox owns, needing the template's `volumeClaimTemplatesPolicy` | cold |
| `lifecycle` | An expiry, a time to live after finishing, and what happens to the claim then | either |

A claim the template does not allow reports `Ready=False` with reason `EnvVarsInjectionRejected`, `VolumeClaimTemplatesError` or `InvalidMetadata`.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `namespace` | Namespace the claim, its pool and its sandbox share |
| `labels`, `annotations` | The SandboxClaim object's own metadata |
| `warmPoolRef.name` | The pool (a KubernetesAgentSandboxWarmPool reference or a literal name) |
| `lifecycle` | `shutdownTime`, `ttlSecondsAfterFinished`, `shutdownPolicy` |
| `additionalPodMetadata` | Labels and annotations for the sandbox's pod |
| `env` | Environment variables added to the sandbox |
| `volumeClaimTemplates` | Volumes added to (or replacing the template's for) the sandbox |

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_claim_name` | Name of the created SandboxClaim (equals `metadata.name`) |
| `namespace` | Namespace the SandboxClaim was created in |

## Prerequisites

- `KubernetesAgentSandboxController` on the cluster, with extensions on.
- A `KubernetesAgentSandboxWarmPool` in the same namespace.
- The target namespace (see `KubernetesNamespace`).

See [GUIDE.md](GUIDE.md) for the decisions behind a good claim and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

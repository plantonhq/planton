# KubernetesAgentSandboxTemplate

Declares a Kubernetes SIG Apps agent-sandbox `SandboxTemplate`: the hardened blueprint of a sandbox, from which warm pools keep sandboxes ready and claims hand them out. The spec mirrors the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxTemplate field for field (pinned to agent-sandbox v1.0.7, the release `KubernetesAgentSandboxController` installs), Kubernetes' full pod specification included, after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One SandboxTemplate** in the target namespace, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

A template runs nothing. The agent-sandbox controller (extensions on) acts on it:

- **A NetworkPolicy** named `<template>-network-policy`, selecting every sandbox created from the template: the secure default, or the declared rules. None when the template is Unmanaged.
- **Sandboxes**, when a `KubernetesAgentSandboxWarmPool` references the template (its members) or a claim binds through that pool: each runs the template's pod and owns its volume claims.

## The Network Fence

| `networkPolicyManagement` | `networkPolicy` | What the sandboxes get |
|---------------------------|-----------------|-------------------------|
| `Managed` (default) | unset | The secure default: ingress only from the controller's router; egress only to the public internet (private ranges, link-local and metadata blocked, which blocks cluster DNS too) |
| `Managed` | set | Exactly the declared ingress and egress rules; an empty block denies everything both ways |
| `Unmanaged` | ignored | No NetworkPolicy; another system (Cilium, a mesh) owns the fence |

The policy is shared: a change updates the fence of every existing and future sandbox from the template.

## What a Claim May Change

| Policy | `Disallowed` (default) | `Allowed` | `Overrides` |
|--------|------------------------|-----------|-------------|
| `envVarsInjectionPolicy` | a claim setting `env` is refused | new names only | new names, and the template's values |
| `volumeClaimTemplatesPolicy` | a claim setting volume claim templates is refused | new templates only | new templates, and replacing the template's by name |

A claim that sets either always gets a fresh, cold sandbox: environment variables and volumes are baked into a pod before it starts.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `namespace` | Namespace the template, its pools, claims and sandboxes share |
| `labels`, `annotations` | The SandboxTemplate object's own metadata |
| `podTemplate` | The pod every sandbox runs: Kubernetes' full pod specification, plus the pod's labels and annotations |
| `volumeClaimTemplates` | Volumes every sandbox owns, one claim each |
| `service` | A headless Service per sandbox |
| `networkPolicy`, `networkPolicyManagement` | The sandboxes' network fence, and who owns it |
| `envVarsInjectionPolicy`, `volumeClaimTemplatesPolicy` | What a claim may add or override |

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_template_name` | Name of the created SandboxTemplate (equals `metadata.name`), which warm pools reference |
| `namespace` | Namespace the SandboxTemplate was created in |

## Prerequisites

- `KubernetesAgentSandboxController` on the cluster, with extensions on.
- The target namespace (see `KubernetesNamespace`).

See [GUIDE.md](GUIDE.md) for the decisions behind a good template and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

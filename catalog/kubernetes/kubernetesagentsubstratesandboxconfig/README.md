# KubernetesAgentSubstrateSandboxConfig

Declares an Agent Substrate `SandboxConfig`: the exact sandbox runtime an actor boots with, for one sandbox class (gVisor or microVM), in versions a team rolls forward and back. The spec mirrors the upstream `ate.dev/v1alpha1` SandboxConfig field for field (pinned to the definition of the Agent Substrate release `KubernetesAgentSubstrate` composes, v0.6.0), after a small Planton envelope: the object's name, labels and annotations. A SandboxConfig is cluster-scoped.

## What Gets Created

- **One SandboxConfig** named `name`, carrying the spec exactly as written. Each architecture's asset files are written as the bare map upstream defines.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The cluster's admission policy (installed by the control plane) checks every version carries the assets its class needs; the spec checks the same rules before the apply. Nothing runs until a worker pool resolves the config: the pool's actors boot with its default version, and the pool keeps it from deletion. Every node downloads its default version ahead of time.

## How a Product Uses It

- **A worker pool** runs the config through `sandbox_classes[].config_ref.name` (a reference to this kind), or through the class default.
- **An actor template** selects pools by their labels (`worker_selector`); its actors run in the class of the pool they land on, with the config that pool resolves. Templates name no config.

## Versions and the Class Default

- **`versions`** lists up to 16 runtime versions, each an exact set of files checked by SHA-256. A new actor boots `default_version`; a suspended actor restores its memory on the version its snapshot was taken with, while that version stays listed.
- **Rolling out** a new runtime is a second version and a move of `default_version`, keeping the previous version listed so suspended actors keep their memory and a rollback is one more move. `state: Disabled` marks a listed version that must not become the default again.
- **The class default** is Substrate's annotation `sandboxconfig.ate.dev/is-class-default: "true"` (set in `annotations`): the config a pool resolves when its `config_ref` is unset. The default created last wins, so annotating an older config later never takes over, and a control plane reinstall hands the default back to `gvisor-default`. The control plane ships `gvisor-default` as gVisor's.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `name` | The config's name in the cluster, which worker pools name |
| `labels`, `annotations` | The SandboxConfig object's own metadata (the class-default annotation among them) |
| `sandboxClass` | `gvisor` or `microvm` |
| `defaultVersion` | The version new sandboxes boot |
| `versions[]` | `name`, `state`, `assets` by architecture and asset name (`url`, `sha256`) |

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_config_name` | Name of the created SandboxConfig, which a worker pool's `config_ref.name` names |
| `sandbox_class` | The config's class, which the pool's `sandbox_classes` entry must equal |

## Prerequisites

- `KubernetesAgentSubstrate` on the cluster: it installs the SandboxConfig definition and the admission policy, and runs the API server and node agents that read it.
- Nodes that can download the asset URLs: the public `gs://gvisor` releases from anywhere with internet egress, your own bucket through the control plane's object-store plugin.

See [GUIDE.md](GUIDE.md) for the decisions behind a good config and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

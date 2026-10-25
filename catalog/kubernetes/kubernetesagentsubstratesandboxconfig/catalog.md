# Agent Substrate Sandbox Config

Declares an Agent Substrate SandboxConfig: the exact sandbox runtime an actor boots with — the gVisor release, or a microVM's hypervisor, kernel and root filesystem — as versions of files checked by SHA-256, rolled forward and back by moving the default. Worker pools resolve the config and keep it from deletion, and the actors on them boot its default version. The spec is faithful to the upstream `ate.dev/v1alpha1` SandboxConfig (pinned to Agent Substrate v0.6.0) — this is how a team pins, rolls out and rolls back the runtime its agents run in.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A SandboxConfig** — one cluster-scoped object naming its sandbox class, its default version and every version's files.
- **Kubernetes Labels** — your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The config costs nothing by itself. Worker pools that run it download its files onto every worker node and cache them by hash.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Substrate installed** — it installs the SandboxConfig definition and the admission policy that holds each class to its files.
- **The runtime files reachable** — the public gVisor releases need internet egress from the nodes; microVM files belong in a bucket the control plane's storage plugin serves.

## Deploy

### Console

Open the deployment store, find **Agent Substrate Sandbox Config**, and click **Deploy**. The creation wizard walks you through the class, the versions and their files, and which version is the default.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateSandboxConfig
metadata:
  name: gvisor-pinned
  org: acme-corp
  env: prod
spec:
  name: gvisor-pinned
  sandboxClass: gvisor
  defaultVersion: gvisor-2026-10-17
  versions:
    - name: gvisor-2026-10-17
      assets:
        amd64:
          entries:
            gvisor:
              url: gs://gvisor/releases/nightly/2026-10-17/x86_64/gvisor.tar.zstd
              sha256: fa72f0d45756c69a30beec21222c78241b0d330c36f4f2185648617cc23415c8
        arm64:
          entries:
            gvisor:
              url: gs://gvisor/releases/nightly/2026-10-17/aarch64/gvisor.tar.zstd
              sha256: 238634978d99f0dd3e631c4f88eae7b19fb36937065428467f0b6d6b10697689
```

```shell
planton apply -f sandbox-config.yaml
```

Every node downloads that gVisor release ahead of time, and a worker pool that names `gvisor-pinned` keeps it from deletion and boots its actors with it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, a worker pool references the config, so the config is created first:

```yaml
spec:
  sandboxClasses:
    - name: gvisor
      configRef:
        name:
          valueFrom:
            kind: KubernetesAgentSubstrateSandboxConfig
            name: gvisor-pinned
            fieldPath: status.outputs.sandbox_config_name
```

A product's actor templates select that pool by its labels; they name no config.

## Key Configuration

These are the most important decisions when configuring a sandbox config. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**A version is an exact runtime.** Every file is checked by SHA-256 on every node, so a version boots the same bytes everywhere, and a changed file is refused. Publish new bytes as a new version.

**Roll out by moving the default.** Add the new version, move `defaultVersion` to it, and keep the previous one listed: rolling back is moving it back. A new actor boots the new default; an actor suspended under the previous version restores its memory on it while it stays listed.

**Cover every architecture your workers run on.** A version with no files for a node's architecture cannot boot there; the default version must cover them all.

**The class default is an annotation.** `sandboxconfig.ate.dev/is-class-default: "true"` makes this config what a worker pool resolves when it names none; the default created last wins, so annotating an older config later never takes over. Pools that name a config explicitly never move on their own.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|-------------------|
| **Agent Substrate** | (prerequisite: the definition and admission policy) | — |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `sandbox_config_name` | Name of the created SandboxConfig | A worker pool's `configRef.name` |
| `sandbox_class` | The config's class (`gvisor` or `microvm`) | The worker pool's `sandboxClasses[].name` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Your own gVisor pin** — one validated gVisor release, independent of the control plane's. Start from the **Pinned gVisor Release** preset.

**A rollout with a way back** — a new release as the default, the previous one kept for rollback. Start from the **Staged gVisor Rollout** preset.

**MicroVM sandboxes** — cloud-hypervisor with Kata's guest kernel and root filesystem, from your bucket. Start from the **MicroVM Runtime** preset.

## Works With

- [**Agent Substrate Worker Pool**](/infra-catalog/kubernetes-agent-substrate-worker-pool) — resolves the config and keeps it from deletion while it runs.
- [**Agent Substrate**](/infra-catalog/kubernetes-agent-substrate) — the control plane: the definition, the admission policy and the shipped `gvisor-default`.
- [**Agent Substrate CSI Driver Config**](/infra-catalog/kubernetes-agent-substrate-csi-driver-config) — your storage for the actors' external volumes.

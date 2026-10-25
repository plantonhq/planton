# KubernetesAgentSubstrateSandboxConfig Guide

The judgment this guide carries: a SandboxConfig is a runtime release, pinned to the byte. Its default version is what every new actor on the pools that run it boots with, and the versions it keeps listed are what suspended actors restore their memory on. Treat it like a release you roll forward with a way back.

## When to declare one

The control plane already ships `gvisor-default`, gVisor's class default, pinned to the gVisor release the installed Substrate release was tested with. Declare your own config when you want any of these:

- **Your own pin.** A gVisor release you have validated, independent of what the next control plane release ships.
- **A controlled rollout.** A new gVisor release beside the old, moved to by `default_version`, with the old kept for rollback.
- **MicroVM sandboxes.** The microVM class has no shipped config: its hypervisor, guest kernel and root filesystem are files you stage in your own bucket.
- **A per-tenant runtime.** A worker pool per tenant can run its own config.

## How a config is used

1. **A worker pool resolves it.** The pool's `config_ref` names it (or the pool takes the class default). The pool's status records it, and the config cannot be deleted while a pool uses it.
2. **An actor lands on the pool.** Its template selects pools by their labels; the pool decides the actor's class and its config. Templates name no config.
3. **A new actor boots the default version**, as the config stands at that moment. A suspended actor restores its memory on the version its snapshot was taken with, if the config still lists it; otherwise it resumes from its files without its memory.
4. **Every node downloads the default version ahead of time**, for every config (microVM configs only on nodes with `/dev/kvm`), so an actor's boot does not wait for the download. Any other version downloads when it is first needed.

## Versions

- **Names** are DNS labels: name them after what they pin (`gvisor-2026-10-17`) rather than a counter, so the default reads as a fact.
- **`default_version`** must name an Enabled version. Moving it changes what every new actor boots; an actor running at the moment keeps its sandbox until it stops.
- **The other versions** are what actors suspended on them restore their memory on. Keep the previous default listed after a move: its suspended actors keep their memory, and a rollback is one field, moving `default_version` back. Its files stay cached on the nodes that downloaded them while it was the default.
- **`state: Disabled`** marks a listed version that must not become the default again (the cluster refuses a Disabled default). Disable a version you have found bad; delete it from the list when you no longer want it as a rollback target.
- **Assets** are keyed by Go's architecture names (`amd64`, `arm64`). A node uses its own architecture's block. Every version must cover every architecture your workers run on, or workers on the others cannot boot it.

## Assets and where they come from

- **gVisor** needs a `gvisor` asset: the release's `gvisor.tar.zstd` (runsc and its helpers). Upstream publishes them at `gs://gvisor/releases/...`, a public bucket the node agent reads anonymously from any cluster with internet egress. A bare `runsc` binary is refused. Substrate's workers need a gVisor release from 2026-10-17 on, which saves an actor's root filesystem and durable directories as separate checkpoint bundles and runs the sandbox with no pause container.
- **MicroVM** needs `cloud-hypervisor`, `virtiofsd`, `kata-kernel` and `kata-image` for every architecture. Stage them in a bucket the control plane's object-store plugin serves (gs:// on GCS, s3:// on S3), outside the snapshot location (the plugin refuses runtime files under it).
- **`sha256`** names the cached copy on every node and verifies every download. A changed file at the same URL is refused rather than booted: publish new bytes under a new URL and a new version.

## The class default

`sandboxconfig.ate.dev/is-class-default: "true"` in `annotations` makes a config its class's default, the one a pool without `config_ref` resolves. More than one config may carry it; the one created last wins, so a new default takes over the moment it is created and pools that follow the default move to it. Two traps follow from "created last":

- Adding the annotation to a config that already existed does not make it win over a default created after it.
- A control plane reinstall recreates `gvisor-default`, which then wins over every gVisor default you created before it.

For a deliberate rollout, name configs explicitly in pools (`config_ref`), and keep the annotation for the one config pools should otherwise follow.

## Deleting a config

A config a worker pool runs cannot be deleted until the pool moves off it: Substrate's controller holds it with a finalizer and the delete waits. A pool that references this kind is created after it and destroyed before it, so a combined destroy never waits.

## Parity accounting

Built for 100% parity with the upstream SandboxConfig (`ate.dev/v1alpha1`) at Agent Substrate v0.6.0, the install shape the control plane composes. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.name`, `metadata.labels`, `metadata.annotations` | `name`, `labels`, `annotations` (the envelope) |
| `spec.sandboxClass` | `sandboxClass` |
| `spec.defaultVersion` | `defaultVersion` |
| `spec.versions[].name`, `.state` | `versions[].name`, `.state` |
| `spec.versions[].assets.<arch>.<asset>.url`, `.sha256` | `versions[].assets.<arch>.entries.<asset>.url`, `.sha256` (the projection writes the bare map) |

### What the spec adds beyond upstream's schema

- **The admission policy's rules, at plan.** Upstream's `sandboxconfig-assets` policy (a gVisor asset per architecture; the four microVM assets per architecture) is checked by the spec, so a config the cluster would refuse never reaches it.
- **Unique version names**, which upstream's list keys enforce on the server.
- **The class is required.** The cluster would default it to gvisor; the class decides which assets every version needs, and a consumer reads it back as an output.

### Not modeled

- **Status** — a SandboxConfig has none. Which config a pool runs is on the pool's status.

## Pairs well with

- **[KubernetesAgentSubstrateWorkerPool](../kubernetesagentsubstrateworkerpool/README.md)** — resolves the config and keeps it from deletion while it runs.
- **[KubernetesAgentSubstrate](../kubernetesagentsubstrate/README.md)** — the control plane: the definition, the admission policy, and the shipped `gvisor-default`.
- **[KubernetesAgentSubstrateCsiDriverConfig](../kubernetesagentsubstratecsidriverconfig/README.md)** — external volumes for the actors a config boots.

# KubernetesAgentSubstrateCsiDriverConfig

Declares an Agent Substrate `CSIDriverConfig`: how Substrate reaches one CSI driver, so an actor's external volume is provisioned and mounted by your storage rather than kept in the actor's snapshots. The spec mirrors the upstream `ate.dev/v1alpha1` CSIDriverConfig field for field (pinned to the definition of the Agent Substrate release `KubernetesAgentSubstrate` composes, v0.6.0), after a small Planton envelope: the object's name, labels and annotations. A CSIDriverConfig is cluster-scoped.

## What Gets Created

- **One CSIDriverConfig** named after the driver, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

Nothing runs for the config itself. The driver must already run on the cluster; this config registers it with Substrate.

## How Substrate Uses It

1. An actor template declares an external volume of a StorageClass (`external_volume_template.storage_class_name`).
2. The StorageClass's `provisioner` is a CSI driver's name; Substrate looks up the CSIDriverConfig of that name.
3. The API server dials the driver's Controller service at `controllerEndpoint` (with TLS through Substrate's pod identity, when on) to create and attach the volume, and checks the driver reports the same name.
4. Each node agent reaches its node's Node service over the kubelet plugin socket (`unix:///var/lib/kubelet/plugins/<driverName>/csi.sock`, or `nodeSocketOverride`) to mount it into the actor.

Because the lookup is by object name, the spec requires `name` to equal `driverName`.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `name` | The config's name in the cluster, equal to `driverName` |
| `labels`, `annotations` | The CSIDriverConfig object's own metadata |
| `driverName` | The driver's CSI name, which its StorageClasses name as provisioner |
| `controllerEndpoint` | `tcp://host:port` or `dns:///service:port` of the driver's Controller service |
| `nodeSocketOverride` | The Node service socket when it is not the kubelet convention |
| `tls` | `enabled`, `usePodIdentity` (required with TLS) and `serverName` |

## Outputs

| Output | Description |
|--------|-------------|
| `driver_name` | The driver's name (also the config's name), which a StorageClass names as its `provisioner` |

## Prerequisites

- `KubernetesAgentSubstrate` on the cluster: it installs the CSIDriverConfig definition and runs the API server and node agents that read it.
- The CSI driver running on the cluster, its Controller service reachable at `controllerEndpoint` from the API server (usually through a TCP bridge in front of the driver's unix socket), and its node plugin bind-mounting `/var/lib/ate` with bidirectional propagation so mounts reach the node agent.

See [GUIDE.md](GUIDE.md) for the decisions behind a good config and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

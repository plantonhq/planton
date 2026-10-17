# KubernetesAgentSandbox

Declares a Kubernetes SIG Apps agent-sandbox `Sandbox`: one isolated, stateful pod with a stable name, its own volumes and an optional headless Service, run by the agent-sandbox controller. The spec mirrors the upstream `agents.x-k8s.io/v1beta1` Sandbox field for field (pinned to agent-sandbox v1.0.7, the release `KubernetesAgentSandboxController` installs), Kubernetes' full pod specification included, after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One Sandbox** in the target namespace, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The agent-sandbox controller then creates and owns the sandbox's children:

- **A pod** named after the sandbox, from `podTemplate`, recreated if it is deleted while the sandbox is Running.
- **One PersistentVolumeClaim per volume claim template**, named `<template>-<sandbox>` and mounted by the template's name.
- **A headless Service** named after the sandbox when `service` is true.

Deleting the Sandbox deletes all three.

## Declared Sandboxes and Claimed Ones

Declare a Sandbox for one that should exist as long as its declaration does: an always-on agent, a long-lived dev box, a service an agent hosts. Sandboxes a product hands out on demand come from a `KubernetesAgentSandboxWarmPool` through a `KubernetesAgentSandboxClaim`, usually created at runtime by agent-sandbox's SDKs. Both kinds of sandbox run the same pod specification; the difference is who owns the object.

## Isolation

A sandbox is as isolated as its pod's runtime:

- `podTemplate.spec.runtimeClassName: gvisor` on GKE Sandbox nodes (with the `sandbox.gke.io/runtime: gvisor` node selector and toleration).
- A Kata class (`kata-qemu`, `kata-fc`, AKS `kata-vm-isolation`) where Kata Containers are installed.
- No runtime class means an ordinary pod on the node's kernel.

A gVisor or Kata pod cannot be port-forwarded. Callers reach it through the controller's router, which looks the sandbox up by name (`X-Sandbox-ID`, `X-Sandbox-Namespace`), or through the headless Service at `<sandbox>.<namespace>.svc`.

## Lifecycle

- **`operatingMode: Suspended`** deletes the pod and keeps the Sandbox, its claims and its Service; `Running` starts a new pod on the same claims. The `Ready` and `Suspended` conditions report what the controller has reached.
- **`shutdownTime`** ends the sandbox at an absolute RFC 3339 time: the pod and Service are deleted, then `shutdownPolicy` keeps the object (`Retain`, marked `SandboxExpired`) or deletes it (`Delete`).
- **`volumeClaimTemplates` are fixed** once the sandbox exists; the API server refuses any change to the list.

## Spec Overview

| Field | Purpose |
|-------|---------|
| `namespace` | Namespace the Sandbox, its pod, claims and Service live in (a KubernetesNamespace reference or a literal) |
| `labels`, `annotations` | The Sandbox object's own metadata |
| `podTemplate.spec` | Kubernetes' full pod specification: containers, volumes, security, scheduling, runtime class |
| `podTemplate.metadata` | The labels and annotations the pod carries |
| `volumeClaimTemplates` | Volumes the sandbox owns, one claim each |
| `service` | Keep a headless Service named after the sandbox |
| `shutdownTime`, `shutdownPolicy` | When the sandbox expires, and whether the object stays |
| `operatingMode` | `Running` or `Suspended` |

Names that point at other objects in the pod (Secrets, ConfigMaps, the ServiceAccount, a PriorityClass, a PersistentVolumeClaim, a StorageClass) are references, so a chart deploys each before the sandbox that uses it.

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_name` | Name of the created Sandbox (equals `metadata.name`); its pod and headless Service share it |
| `namespace` | Namespace the Sandbox was created in |

## Prerequisites

- `KubernetesAgentSandboxController` on the cluster: it installs the Sandbox definition and runs the pod.
- The target namespace (see `KubernetesNamespace`).

See [GUIDE.md](GUIDE.md) for the decisions behind a good sandbox and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

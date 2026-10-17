# KubernetesAgentSandbox Guide

The judgment this guide carries: a sandbox is a pod with a promise. Its name is stable, its volumes outlive its pod, and it can sleep without losing anything. Whether that pod is actually isolated, reachable and safe to leave running is decided in the pod template and the lifecycle, so those are where the care goes.

## Declare it, or let a product claim it

A Sandbox declared here exists as long as its declaration does. That is right for an always-on agent, a long-lived dev box, a demo environment, a service an agent hosts. A product that hands out sandboxes per user or per task should not declare each one: it creates claims at runtime against a warm pool (`KubernetesAgentSandboxWarmPool`, `KubernetesAgentSandboxClaim`), and the pool's template carries the same pod specification this kind does. If you find yourself stamping out near-identical sandboxes from a chart, you want a template and a pool.

## Isolation is a runtime class, not a setting

The controller runs whatever pod you describe; nothing about a Sandbox isolates it by itself. For code you don't fully trust (model-generated code, a user's repository, a browser), give the pod a sandboxing runtime:

- **gVisor** — `runtimeClassName: gvisor` on GKE Sandbox node pools, with their `sandbox.gke.io/runtime: gvisor` node selector and toleration. A user-space kernel between the code and the node.
- **Kata Containers** — `kata-qemu`, `kata-fc` or AKS `kata-vm-isolation`. A lightweight virtual machine per pod.

Layer the pod's own hardening on top: non-root, no privilege escalation, every capability dropped, the runtime's seccomp profile, and `automountServiceAccountToken: false` so the code holds no Kubernetes credentials. A ServiceAccount with workload identity is the way to give an agent cloud access without a key in the pod.

A gVisor or Kata pod cannot be port-forwarded. Plan how callers reach it: the controller's router (by sandbox name, with TokenReview or scoped-token authorization), or the headless Service from inside the cluster.

## Volumes are the sandbox's memory, and they are fixed

Each volume claim template becomes a claim named `<template>-<sandbox>`, owned by the Sandbox. The pod mounts it by the template's name, so a `volumeMounts` entry naming the template is all it takes. The claim outlives every pod the sandbox ever runs: a restart, a reschedule and a suspension all keep the files. It goes when the Sandbox is deleted.

The API server refuses any change to `volumeClaimTemplates` once the sandbox exists. Size the workspace for the sandbox's life, and pick the storage class (an encrypted one, a faster one) up front. Changing a sandbox's volumes means replacing it, with its data.

An ephemeral volume (`volumes[].ephemeral`) is the other choice: a claim that lives and dies with each pod, for scratch space too large for `emptyDir`.

## Suspend, expire, delete

Three ways a sandbox stops running, each keeping a different amount:

| What you do | The pod | The claims | The Sandbox object |
|-------------|---------|------------|--------------------|
| `operatingMode: Suspended` | deleted | kept | kept; `Running` resumes it |
| `shutdownTime` reached, `shutdownPolicy: Retain` | deleted (and the Service) | kept | kept, `Ready=False` with reason `SandboxExpired` |
| `shutdownTime` reached, `shutdownPolicy: Delete` | deleted | deleted | deleted |
| Destroy this resource | deleted | deleted | deleted |

Suspending is the cheap way to idle: CPU and memory come back to the cluster, storage keeps costing. `shutdownTime` is absolute, so a manifest whose time has passed creates a sandbox that is torn down at once; extend a window by moving the time. Prefer `Retain` for a declared sandbox: `Delete` removes an object Planton still declares, and the next apply recreates it (already expired).

## What the API server checks that the spec cannot

- **The name is at most 63 characters when `service` is true**, because the Service is named after the sandbox. The rule reads the object's metadata, which no spec rule can see, so the API server refuses the apply.
- **Feature-gated pod fields** (pod-level resources, `hostnameOverride`, pod certificates, container restart rules and the others whose comments name a gate) are accepted on the Sandbox, but a cluster without the gate drops them silently from the pod it creates. Check the cluster's version and gates before relying on one.
- **Pod-level validation the definition doesn't carry** — the spec refuses the common mistakes before apply (two sources in one volume, two handlers on one probe, a wrong selector operator), but some of the Pod API's checks only run when the controller creates the pod. A sandbox whose pod is refused stays `Ready=False`; its conditions and the controller's events say why.

## Parity accounting

Built for 100% parity with the upstream Sandbox (`agents.x-k8s.io/v1beta1`) at agent-sandbox v1.0.7. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.namespace`, `metadata.labels`, `metadata.annotations` | `namespace`, `labels`, `annotations` (the envelope) |
| `spec.podTemplate.spec` | `podTemplate.spec`: Kubernetes 1.37's full PodSpec (`k8s.io/api` v0.37.1), every key of it, from `catalog/kubernetes/core_v1_api.proto` |
| `spec.podTemplate.metadata.labels`, `.annotations` | `podTemplate.metadata.labels`, `.annotations` |
| `spec.volumeClaimTemplates[]` (`metadata.name`, `.labels`, `.annotations`, `spec`) | `volumeClaimTemplates[]`, with the full PersistentVolumeClaimSpec |
| `spec.service` | `service` |
| `spec.shutdownTime` | `shutdownTime` (RFC 3339) |
| `spec.shutdownPolicy` | `shutdownPolicy` |
| `spec.operatingMode` | `operatingMode` |

### What the spec adds beyond upstream's schema

Each addition is a rule the cluster enforces later, moved earlier:

- **Closed sets** for every value the Pod API accepts from a fixed list (restart policy, DNS policy, protocols, pull policy, seccomp and AppArmor types, toleration operators and effects, and the rest). The served definition carries no enums, but the pod is refused for any other value.
- **Quantity grammar** on every resource and size field, and **number-or-name rules** on every port that takes either.
- **One source per volume, one handler per probe and hook, one origin per environment variable**, and the value counts each selector operator needs.
- **A name on every volume claim template**, as a DNS label: upstream marks it optional, but the controller names the pod's volume and the claim after it.
- **Label grammar** on the object's labels, the pod's labels and the claims' labels.
- **References** for the Secrets, ConfigMaps, ServiceAccount, PriorityClass, PersistentVolumeClaims and StorageClasses the pod names, so a chart deploys them first.

### Not modeled

- **The status block** (`serviceFQDN`, `service`, `conditions`, `selector`, `podIPs`, `nodeName`) — it is the controller's report, not configuration. The sandbox's name and namespace are the outputs a caller needs; the Service's name is the sandbox's.
- **The ephemeral volume template's `metadata`** is modeled, but empty: the definition serves it with no properties, so the API server prunes anything set there. Label the claim through the sandbox's own volume claim templates instead.

## Pairs well with

- **[KubernetesAgentSandboxController](../kubernetesagentsandboxcontroller/README.md)** — installs the definition, runs the pod, and routes callers to the sandbox by name.
- **[KubernetesServiceAccount](../kubernetesserviceaccount/README.md)** — the identity the sandbox's code runs as.
- **[KubernetesSecret](../kubernetessecret/README.md)** — credentials the sandbox reads through `secretKeyRef` or a secret volume, never as plain environment values.
- **[KubernetesNetworkPolicy](../kubernetesnetworkpolicy/README.md)** — fences a declared sandbox by its pod labels; a SandboxTemplate's managed policy does this only for the sandboxes it creates.

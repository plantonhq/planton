# Hardened gVisor Template

The blueprint for sandboxes that run code nobody has reviewed: gVisor isolation, no Kubernetes credentials, a locked-down container and the controller's secure-default network fence. Point a warm pool at it and every sandbox a product claims starts this way.

## When to Use

- Sandboxes run model-generated code, a user's repository or anything else you would not run on a node's kernel.
- You want one hardened definition that every pool and claim inherits, owned by the platform team.

## How It Works

Every sandbox created from the template runs one pod under the `gvisor` runtime class on GKE Sandbox nodes, as a non-root user with no privilege escalation, no capabilities, a read-only root filesystem and the runtime's seccomp profile. `/tmp` is a 1Gi scratch volume on the node. With `network_policy` left unset, the controller fences the template's sandboxes with its secure default: only the controller's router can reach them, and they can reach only the public internet (private ranges, the cloud metadata server and in-cluster DNS are blocked).

## Key Configuration Choices

- **Secure-default fence.** No `network_policy` means the controller's default, which blocks in-cluster DNS too. Declare rules when a sandbox must resolve or reach in-cluster services.
- **No credentials.** `automountServiceAccountToken: false`; the controller would set it anyway for template sandboxes, and saying it keeps the template honest.
- **Read-only root.** Write only to `/tmp` and to volumes; add a volume claim template for state that must outlive a sandbox's pod.
- **Claim policies left at `Disallowed`.** Claims cannot inject environment variables or volumes, so every sandbox starts warm from the pool.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on and its router (the only way into a gVisor sandbox).
- Nodes running GKE Sandbox (or swap in your cluster's Kata runtime class and node labels).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<template-name>` | Name of the template, which warm pools reference. |
| `<namespace>` | Namespace the template, its pools and their sandboxes live in. |
| `<sandbox-image>` | The image every sandbox runs, serving on port 8888. |

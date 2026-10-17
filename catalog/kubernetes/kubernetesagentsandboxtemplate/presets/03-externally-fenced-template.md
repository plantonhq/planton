# Externally Fenced Template

A template whose network fence belongs to another system: Cilium network policies with DNS-aware egress, a service mesh's authorization, or a platform-wide policy engine. The controller creates no NetworkPolicy for it, and the sandboxes carry the label that system selects them by.

## When to Use

- The cluster's network policy is owned by Cilium, Calico or a mesh, and a second, overlapping Kubernetes NetworkPolicy would fight it.
- You need egress control by host name (Cilium's `toFQDNs`), which a Kubernetes NetworkPolicy cannot express.

## How It Works

With `networkPolicyManagement: Unmanaged`, the controller skips the template's shared NetworkPolicy entirely, and any `networkPolicy` rules would be ignored. Every sandbox's pod carries `example.com/fence: agent-sandbox`, the label your own policy selects. Until that policy exists, the sandboxes are unfenced.

## Key Configuration Choices

- **Write the external policy first.** An Unmanaged template with no policy behind it leaves every sandbox open to the cluster network. Deploy the Cilium or mesh policy before the warm pool.
- **One label to select by.** Keep the fence label stable; renaming it in the template silently drops existing sandboxes out of the policy.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- The external policy system on the cluster, with a policy selecting the fence label.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<template-name>` | Name of the template, which warm pools reference. |
| `<namespace>` | Namespace the template, its pools and their sandboxes live in. |
| `<sandbox-image>` | The image every sandbox runs. |

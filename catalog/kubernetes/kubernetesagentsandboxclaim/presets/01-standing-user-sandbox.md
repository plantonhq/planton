# Standing User Sandbox

A sandbox from a pool that stays assigned to one person or team for as long as it is declared: a named workbench, a team's shared agent, a demo environment. It binds warm and carries a label naming its owner, so policies and dashboards can tell whose it is.

## When to Use

- Someone needs a sandbox of the pool's shape that outlives any single task.
- You want it declared with the rest of the platform rather than created by a product at runtime.

## How It Works

The claim takes a ready member of the pool, which becomes this claim's sandbox, and the pool starts a replacement. The sandbox's pod gains the label `sandbox.users.io/user`, set to the owner (`alice` in the preset). A warm member takes labels without a restart, so the claim still binds warm. The claimed sandbox's name, which a caller sends to the router, is in the claim's status (`sandbox.name`).

## Key Configuration Choices

- **Labels in an allowed domain.** The controller accepts pod labels only under its allowed label domains (`sandbox.users.io` by default; `extensions.allowed_label_domains` on the controller). Anything else is refused with reason `InvalidMetadata`.
- **No variables or volumes.** Either would force a cold start; this claim keeps the warm bind.
- **No expiry.** The sandbox lives until the claim is removed from the declaration.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- **Agent Sandbox Warm Pool** in the same namespace.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<user>` | The person or team the sandbox belongs to, in the claim's name; set the `sandbox.users.io/user` label to the same owner. |
| `<namespace>` | Namespace the claim, its pool and its sandbox share. |
| `<pool-name>` | Name of the warm pool to claim from (or reference a KubernetesAgentSandboxWarmPool with `valueFrom`). |

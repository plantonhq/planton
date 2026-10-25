# Dedicated Tenant Pool

One tenant's own worker pool: its own label (which only that tenant's actor templates select), its own namespace, and its own tainted nodes, running a SandboxConfig the tenant's team pins. Stamp one out per tenant, and a noisy or hostile tenant's actors never share a node, a worker or a runtime release with anyone else's.

## When to Use

- A product runs agents for several customers and each must be isolated down to the node.
- A tenant has its own capacity commitment, runtime pin or rollout schedule.

## How It Works

The pool carries `tenant: acme`, so only actor templates with `worker_selector.match_labels: {tenant: acme}` run on it. Its workers select the tenant's nodes and tolerate their taint, so nothing else schedules there (taint the nodes `agents.example.com/tenant=acme:NoSchedule` and label them the same). The pool resolves the SandboxConfig it names and holds it from deletion, and it is the runtime the tenant's actors boot.

## Key Configuration Choices

- **The pool's label is the boundary.** Actor templates select pools by the pool's own labels; the worker pod label here only helps dashboards and policies.
- **Taint and select the nodes.** The toleration lets workers onto the tenant's nodes; the node selector keeps them off everyone else's.
- **Two CPUs and two GiB per worker** are the budget the actors on one worker share; the memory request equals the limit, because a worker short of memory kills every actor it hosts.
- **A pinned runtime**, so the tenant's runtime moves only when its config does.

## Prerequisites

- **Agent Substrate** on the cluster (`KubernetesAgentSubstrate`).
- **A SandboxConfig** of class gvisor (`KubernetesAgentSubstrateSandboxConfig`).
- The tenant's namespace, with no Pod Security enforcement, and the tenant's nodes labeled and tainted.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, such as `acme-workers`. |
| `acme` (tenant) | The tenant's identifier, in the pool's label, the worker pod label, the node selector and the toleration; replace it everywhere. |
| `acme-agents` (namespace) | The tenant's namespace. |
| `<substrate-resource-name>` | Name of the KubernetesAgentSubstrate resource whose worker image the pool runs. |
| `<sandbox-config-resource-name>` | Name of the KubernetesAgentSubstrateSandboxConfig resource the pool runs. |

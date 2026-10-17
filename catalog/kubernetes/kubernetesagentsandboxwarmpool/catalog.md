# Agent Sandbox Warm Pool

Declares an agent-sandbox SandboxWarmPool: a number of sandboxes, built from one template, kept running and ready so a claim gets one in the time it takes to hand it over instead of the time it takes to start a pod. The agent-sandbox controller keeps the pool filled, gives a ready member to each claim, and replaces what it hands out. The spec is faithful to the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxWarmPool (pinned to agent-sandbox v1.0.7) -- this is what a product claims from, and the dial that sets how many idle sandboxes you pay for.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A SandboxWarmPool** -- one namespaced object naming its template, its size and how stale members are replaced.
- **Kubernetes Labels** -- your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The controller then keeps the pool's members: Sandboxes named `<pool>-<random>`, each running the template's pod and owning its volume claims, until a claim takes one.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Sandbox Controller installed, with extensions on** -- it installs the SandboxWarmPool definition and the controller that fills the pool.
- **An Agent Sandbox Template in the same namespace** -- the blueprint every member is built from.
- **Target namespace exists** -- reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Agent Sandbox Warm Pool**, and click **Deploy**. The creation wizard walks you through the namespace, the template, the pool's size and its update strategy.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxWarmPool
metadata:
  name: coding-agent-pool
  org: acme-corp
  env: prod
spec:
  namespace:
    value: agents
  replicas: 3
  sandboxTemplateRef:
    name:
      value: coding-agent
```

```shell
planton apply -f agent-sandbox-warm-pool.yaml
```

The controller starts three sandboxes from the `coding-agent` template. A product then claims from the pool by name and namespace (`create_sandbox(warmpool="coding-agent-pool", namespace="agents")` in agent-sandbox's SDKs) and gets a ready sandbox at once. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the template and the namespace to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: agents-namespace
      fieldPath: spec.name
  sandboxTemplateRef:
    name:
      valueFrom:
        kind: KubernetesAgentSandboxTemplate
        name: coding-agent
        fieldPath: status.outputs.sandbox_template_name
```

The InfraPipeline creates the namespace, then the template, then the pool. A claim (or a product's configuration) reads the pool's `sandbox_warm_pool_name` and `namespace` outputs.

## Key Configuration

These are the most important decisions when configuring a warm pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The size is the cost of instant sandboxes.** Every member reserves its template's resources while it waits. Size the pool for the claims that arrive within one pod start-up time; claims beyond that start cold.

**Leave `replicas` unset when an autoscaler owns it.** The cluster then starts the pool at one and Planton never writes the size again, so a HorizontalPodAutoscaler or KEDA can scale the pool through its scale endpoint without each apply resetting it.

**The update strategy decides how template changes roll out.** `OnReplenish` (the default) replaces stale members only as they are claimed, so the pool never dips. `Recreate` replaces them at once, so every new claim gets the current template.

**Claimed sandboxes leave the pool.** A claimed member belongs to its claim; the pool starts a replacement and never touches it again.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesAgentSandboxTemplate** | `sandboxTemplateRef.name` | `status.outputs.sandbox_template_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `sandbox_warm_pool_name` | Name of the created SandboxWarmPool (equals `metadata.name`) | A claim's `warmPoolRef.name`; the pool name a product passes to agent-sandbox's SDKs; an autoscaler's scale target |
| `namespace` | The namespace the pool was created in | Where a product creates its claims |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Always a few ready** -- three members, replaced as they are claimed, so claims always bind warm. Start from the **Steady Warm Pool** preset.

**Sized by demand** -- an autoscaler owns the pool's size through its scale endpoint. Start from the **Autoscaled Warm Pool** preset.

**Every claim on the newest template** -- stale members replaced at once on a template change. Start from the **Recreate-Rolled Warm Pool** preset.

## Works With

- [**Agent Sandbox Template**](/infra-catalog/kubernetes-agent-sandbox-template) -- the blueprint every member is built from.
- [**Agent Sandbox Claim**](/infra-catalog/kubernetes-agent-sandbox-claim) -- takes a ready member from the pool.
- [**Agent Sandbox Controller**](/infra-catalog/kubernetes-agent-sandbox-controller) -- fills the pool and binds claims to its members.
- [**Kubernetes Horizontal Pod Autoscaler**](/infra-catalog/kubernetes-horizontal-pod-autoscaler) -- scales the pool through its scale endpoint.

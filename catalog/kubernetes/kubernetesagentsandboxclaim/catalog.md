# Agent Sandbox Claim

Declares an agent-sandbox SandboxClaim: a request for one sandbox from a warm pool. The agent-sandbox controller hands the claim a ready member of the pool, or starts a fresh sandbox from the pool's template when none is ready or the claim asks for something a running pod cannot take, and the sandbox then belongs to the claim. The spec is faithful to the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxClaim (pinned to agent-sandbox v1.0.7) -- products usually create claims at runtime, and this is how you declare one that should exist as long as its declaration does.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A SandboxClaim** -- one namespaced object naming its pool, its lifecycle, and the labels, variables and volumes its sandbox gets.
- **Kubernetes Labels** -- your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The controller then binds the claim to a sandbox -- a warm pool member, or one started for it -- and records the sandbox's name in the claim's status.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Sandbox Controller installed, with extensions on** -- it installs the SandboxClaim definition and the controller that binds claims.
- **An Agent Sandbox Warm Pool in the same namespace** -- the pool the claim draws from.
- **Target namespace exists** -- reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Agent Sandbox Claim**, and click **Deploy**. The creation wizard walks you through the namespace, the pool, the sandbox's labels, variables and volumes, and the claim's lifecycle.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxClaim
metadata:
  name: alice-workbench
  org: acme-corp
  env: prod
spec:
  namespace:
    value: agents
  warmPoolRef:
    name:
      value: coding-agent-pool
  additionalPodMetadata:
    labels:
      sandbox.users.io/user: alice
```

```shell
planton apply -f agent-sandbox-claim.yaml
```

The controller hands the claim a ready member of `coding-agent-pool` and labels its pod for alice; the claim's status names the sandbox, which callers reach through the controller's router. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the pool and the namespace to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: agents-namespace
      fieldPath: spec.name
  warmPoolRef:
    name:
      valueFrom:
        kind: KubernetesAgentSandboxWarmPool
        name: coding-agent-pool
        fieldPath: status.outputs.sandbox_warm_pool_name
```

The InfraPipeline creates the namespace, the template and the pool before the claim.

## Key Configuration

These are the most important decisions when configuring a claim. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Ask for nothing extra, and it binds warm.** A claim that sets no environment variables and no volume claim templates takes a ready pool member at once. Pod labels and annotations do not affect this: a warm member takes them without a restart.

**Variables and volumes cost a cold start, and need the template's consent.** A running pod cannot take either, so the controller starts a sandbox for the claim. The pool's template must allow them (`envVarsInjectionPolicy`, `volumeClaimTemplatesPolicy`); otherwise the claim reports `Ready=False` with the reason.

**Pod labels must be in an allowed domain.** The controller accepts claim labels only under its allowed label domains (`sandbox.users.io` by default). Anything else is refused with reason `InvalidMetadata`.

**Expiry ends the sandbox; the policy decides the claim.** `lifecycle.shutdownTime` and `ttlSecondsAfterFinished` end the sandbox. `shutdownPolicy` keeps the claim (`Retain`) or deletes it (`Delete`, `DeleteForeground`), and a deleted claim is recreated by the next apply. Keep `Retain` for a declared claim.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesAgentSandboxWarmPool** | `warmPoolRef.name` | `status.outputs.sandbox_warm_pool_name` |
| **KubernetesStorageClass** | `volumeClaimTemplates[].spec.storageClassName` | `status.outputs.storage_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `sandbox_claim_name` | Name of the created SandboxClaim (equals `metadata.name`) | Reading the claim's status for the sandbox's name |
| `namespace` | The namespace the claim was created in | The `X-Sandbox-Namespace` a caller sends through the router |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**A sandbox that stays someone's** -- a warm claim with an owner label, for a named workbench or a team's shared agent. Start from the **Standing User Sandbox** preset.

**A sandbox shaped for one task** -- the task's ID, a larger workspace, and an end time. Start from the **Task Sandbox With Workspace** preset.

## Works With

- [**Agent Sandbox Warm Pool**](/infra-catalog/kubernetes-agent-sandbox-warm-pool) -- the pool the claim draws from.
- [**Agent Sandbox Template**](/infra-catalog/kubernetes-agent-sandbox-template) -- decides what the claimed sandbox runs and what the claim may change.
- [**Agent Sandbox Controller**](/infra-catalog/kubernetes-agent-sandbox-controller) -- binds the claim and routes callers to its sandbox.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target the claim, its pool and its sandbox share.

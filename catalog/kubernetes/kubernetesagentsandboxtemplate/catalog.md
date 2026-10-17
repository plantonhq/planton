# Agent Sandbox Template

Declares an agent-sandbox SandboxTemplate: the hardened blueprint of a sandbox -- its pod, its volumes, its network fence, and what a claim may change -- from which warm pools keep sandboxes ready and claims hand them out. A template runs nothing by itself; the agent-sandbox controller fences its sandboxes with a shared NetworkPolicy, and every sandbox a pool or a claim creates runs its pod. The spec is faithful to the upstream `extensions.agents.x-k8s.io/v1beta1` SandboxTemplate (pinned to agent-sandbox v1.0.7), Kubernetes' full pod specification included -- this is the node a platform team owns and hardens once for every sandbox a product hands out.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A SandboxTemplate** -- one namespaced object carrying the pod template, the volume claim templates, the network policy and the claim policies.
- **Kubernetes Labels** -- your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The controller then keeps one NetworkPolicy for the template, named `<template>-network-policy`, selecting every sandbox created from it -- unless the template is Unmanaged.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Sandbox Controller installed, with extensions on** -- it installs the SandboxTemplate definition and the controllers that act on it.
- **A NetworkPolicy-enforcing network** -- for the template's fence to take effect (most managed clusters enforce it; kind's default network does not).
- **Target namespace exists** -- reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Agent Sandbox Template**, and click **Deploy**. The creation wizard walks you through the namespace, the pod every sandbox runs, its volumes, its network fence, and what claims may change.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxTemplate
metadata:
  name: coding-agent
  org: acme-corp
  env: prod
spec:
  namespace:
    value: agents
  podTemplate:
    spec:
      runtimeClassName: gvisor
      nodeSelector:
        sandbox.gke.io/runtime: gvisor
      tolerations:
        - key: sandbox.gke.io/runtime
          operator: Equal
          value: gvisor
          effect: NoSchedule
      containers:
        - name: agent
          image: us-docker.pkg.dev/acme/agents/coding-agent:2.3.0
          ports:
            - containerPort: 8888
          resources:
            requests:
              cpu: "1"
              memory: 2Gi
```

```shell
planton apply -f agent-sandbox-template.yaml
```

The controller fences the template's future sandboxes with its secure default: only the router can reach them, and they reach only the public internet. A warm pool referencing `coding-agent` then keeps sandboxes of this shape ready. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace to a resource managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: agents-namespace
      fieldPath: spec.name
  podTemplate:
    spec:
      containers:
        - name: agent
          image: us-docker.pkg.dev/acme/agents/coding-agent:2.3.0
```

The InfraPipeline creates the namespace first, then the template. A KubernetesAgentSandboxWarmPool references the template's `sandbox_template_name` output, so the pool is created after it.

## Key Configuration

These are the most important decisions when configuring a template. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The template decides isolation for every sandbox.** Set the runtime class (gVisor on GKE Sandbox nodes, a Kata class elsewhere), the security context and the resources here, once. A warm pool rolls its members to a changed template by its update strategy.

**The fence is managed by default, and strict.** With no `networkPolicy`, the controller admits ingress only from its router and egress only to the public internet -- private ranges, the metadata server and cluster DNS are blocked. Declaring `networkPolicy` replaces that default with exactly your rules; `networkPolicyManagement: Unmanaged` hands the fence to Cilium or a mesh.

**Claim policies trade flexibility for warm starts.** `envVarsInjectionPolicy` and `volumeClaimTemplatesPolicy` decide whether a claim may add (or override) environment variables and volumes. Any claim that does gets a fresh, cold sandbox, because those are baked into a pod before it starts.

**Sandboxes hold no Kubernetes credentials by default.** When the pod template leaves `automountServiceAccountToken` unset, the controller sets it to false for every sandbox from a template.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSecret** | `podTemplate.spec` -- `secretKeyRef`, `secretRef`, secret volumes and projections, `imagePullSecrets` | `status.outputs.secret_name` |
| **KubernetesConfigMap** | `podTemplate.spec` -- `configMapKeyRef`, `configMapRef`, ConfigMap volumes and projections | `status.outputs.configmap_name` |
| **KubernetesServiceAccount** | `podTemplate.spec.serviceAccountName` | `status.outputs.service_account_name` |
| **KubernetesPriorityClass** | `podTemplate.spec.priorityClassName` | `status.outputs.priority_class_name` |
| **KubernetesPersistentVolumeClaim** | `podTemplate.spec.volumes[].persistentVolumeClaim.claimName` | `status.outputs.pvc_name` |
| **KubernetesStorageClass** | `volumeClaimTemplates[].spec.storageClassName` | `status.outputs.storage_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `sandbox_template_name` | Name of the created SandboxTemplate (equals `metadata.name`) | A warm pool's `sandboxTemplateRef.name` |
| `namespace` | The namespace the template was created in | Placing the warm pools and claims that use it |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Sandboxes for untrusted code** -- gVisor isolation, no credentials, a locked-down container and the controller's secure-default fence. Start from the **Hardened gVisor Template** preset.

**Coding agents with a workspace** -- a per-sandbox workspace, HTTPS egress outside private ranges, and claims allowed to bring their own settings. Start from the **Coding Agent Template** preset.

**A fence owned by Cilium or a mesh** -- the controller creates no policy and the sandboxes carry the label your policy selects. Start from the **Externally Fenced Template** preset.

## Works With

- [**Agent Sandbox Controller**](/infra-catalog/kubernetes-agent-sandbox-controller) -- installs the definition, fences the template's sandboxes, and routes callers to them.
- [**Agent Sandbox Warm Pool**](/infra-catalog/kubernetes-agent-sandbox-warm-pool) -- keeps sandboxes from this template ready to claim.
- [**Agent Sandbox**](/infra-catalog/kubernetes-agent-sandbox) -- the same pod blueprint, for one long-lived sandbox you declare directly.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target the template, its pools and their sandboxes share.

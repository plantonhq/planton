# Agent Sandbox

Declares an agent-sandbox Sandbox: one isolated, stateful pod with a stable name, its own volumes and an optional headless Service, run by the agent-sandbox controller for an AI agent, a notebook or a dev environment. The controller keeps the pod running, suspends it on request while keeping its files, and tears it down when it expires. The spec is faithful to the upstream `agents.x-k8s.io/v1beta1` Sandbox (pinned to agent-sandbox v1.0.7), Kubernetes' full pod specification included, after a Planton envelope that places the object and sets its own labels -- this is how you declare a sandbox that should exist as long as its declaration does.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A Sandbox** -- one namespaced object carrying the pod template, the volume claim templates, the lifecycle and the operating mode, which the agent-sandbox controller turns into a running sandbox.
- **Kubernetes Labels** -- your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The controller then creates, and owns, the sandbox's pod (named after it), one PersistentVolumeClaim per volume claim template, and with `service: true` a headless Service.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Sandbox Controller installed** -- it installs the Sandbox definition and runs every sandbox's pod.
- **A sandboxing runtime, for untrusted code** -- gVisor nodes (GKE Sandbox) or Kata Containers, selected through the pod's runtime class.
- **Target namespace exists** -- reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Agent Sandbox**, and click **Deploy**. The creation wizard walks you through the namespace, the pod (containers, isolation, security, scheduling), the sandbox's volumes, and its lifecycle.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandbox
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
      automountServiceAccountToken: false
      containers:
        - name: agent
          image: us-docker.pkg.dev/acme/agents/coding-agent:2.3.0
          ports:
            - containerPort: 8888
          volumeMounts:
            - name: workspace
              mountPath: /workspace
  volumeClaimTemplates:
    - metadata:
        name: workspace
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 20Gi
  service: true
```

```shell
planton apply -f agent-sandbox.yaml
```

The controller starts the agent's pod under gVisor with a 20Gi workspace, and the sandbox answers at `coding-agent.agents.svc` and through the controller's router by name. An Infra Job tracks the provisioning in real time.

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

The InfraPipeline creates the namespace first, then the sandbox. Secrets, ConfigMaps, the ServiceAccount, a PriorityClass and storage classes the pod names are references too, so each is deployed before the sandbox that uses it.

## Key Configuration

These are the most important decisions when configuring a sandbox. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Isolation is the pod's runtime class.** A sandbox is as isolated as the runtime its pod runs under: `runtimeClassName: gvisor` on GKE Sandbox nodes, a Kata class elsewhere. Without one, it is an ordinary pod sharing the node's kernel. Port-forwarding does not reach a gVisor or Kata pod; callers come through the controller's router or the headless Service.

**Volumes belong to the sandbox, and are fixed.** Each volume claim template becomes a claim named `<template>-<sandbox>` that outlives the pod, so restarts and suspensions keep the files. The list cannot change once the sandbox exists; changing a sandbox's volumes means replacing it.

**Suspend instead of delete.** `operatingMode: Suspended` deletes the pod and keeps the object, its claims and its Service, giving back CPU and memory while keeping state. `Running` brings it back where it left off.

**Expiry is absolute.** `shutdownTime` ends the sandbox at a moment in time, and `shutdownPolicy` decides whether the object stays (`Retain`, marked expired) or goes (`Delete`, which the next apply recreates). Keep `Retain` for a declared sandbox.

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
| `sandbox_name` | Name of the created Sandbox (equals `metadata.name`); its pod and headless Service share it | The `X-Sandbox-ID` a caller sends through the router |
| `namespace` | The namespace the sandbox was created in | The `X-Sandbox-Namespace` a caller sends through the router |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**An agent that is always there** -- a gVisor-isolated agent with a persistent workspace and a stable in-cluster name. Start from the **Always-On Agent Workbench** preset.

**A notebook that sleeps when idle** -- a Jupyter sandbox you suspend between sessions, keeping its files. Start from the **Suspendable Notebook** preset.

**A sandbox with a deadline** -- an isolated sandbox that ends itself at a set time and records that it did. Start from the **Time-Boxed Sandbox** preset.

## Works With

- [**Agent Sandbox Controller**](/infra-catalog/kubernetes-agent-sandbox-controller) -- installs the Sandbox definition, runs the sandbox's pod, and routes callers to it by name.
- [**Agent Sandbox Template**](/infra-catalog/kubernetes-agent-sandbox-template) -- the same pod blueprint, for sandboxes a warm pool and claims create on demand.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target the sandbox lives in.
- [**Kubernetes Service Account**](/infra-catalog/kubernetes-service-account) -- the identity the sandbox's code runs as, with workload identity for cloud access.

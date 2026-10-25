# Agent Substrate Worker Pool

Declares an Agent Substrate WorkerPool: the worker pods agents run in, many actors to a worker, sized, placed and isolated as one pool. Substrate places each actor a product creates or resumes on a worker with room for it, in the pool its template selects, and takes the actor's share back when it is suspended into storage, so a pool serves many more actors than it runs at once. The spec is faithful to the upstream `ate.dev/v1alpha1` WorkerPool (pinned to Agent Substrate v0.6.0) — this is the capacity and the boundary an agent runtime is built from, one pool per tenant if you want.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A WorkerPool** — one namespaced object naming its size, its worker image, its worker pods' placement and resources, and its sandbox class.
- **Kubernetes Labels** — your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

Substrate's controller then runs the pool: a Deployment of the same name with one pod per worker, and a NetworkPolicy that lets only Substrate's router reach the workers.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Substrate installed** — the definition, the controller that runs the pool, the node agent on every node and the API server the workers register with.
- **Target namespace exists, with no Pod Security enforcement** — workers run as root with the capabilities gVisor and microVM need.
- **KVM-capable nodes for a microVM pool** — nested virtualization on cloud VMs.

## Deploy

### Console

Open the deployment store, find **Agent Substrate Worker Pool**, and click **Deploy**. The creation wizard walks you through the namespace, the pool's labels and size, the sandbox class and its config, and where the workers run.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateWorkerPool
metadata:
  name: acme-workers
  org: acme-corp
  env: prod
spec:
  namespace:
    value: acme-agents
  labels:
    tenant: acme
  replicas: 3
  workerImage:
    valueFrom:
      kind: KubernetesAgentSubstrate
      name: substrate
      fieldPath: status.outputs.gvisor_worker_image
  template:
    resources:
      requests:
        cpu: 250m
        memory: 1Gi
      limits:
        cpu: "1"
        memory: 1Gi
  sandboxClasses:
    - name: gvisor
```

```shell
planton apply -f worker-pool.yaml
```

Three gVisor workers start and register, each with one CPU and one GiB for the actors it hosts. Actor templates with `worker_selector.match_labels: {tenant: acme}` run on them, as many at a time as fit within those limits. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the image, the runtime and the namespace to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: acme-agents
      fieldPath: spec.name
  workerImage:
    valueFrom:
      kind: KubernetesAgentSubstrate
      name: substrate
      fieldPath: status.outputs.gvisor_worker_image
  sandboxClasses:
    - name: gvisor
      configRef:
        name:
          valueFrom:
            kind: KubernetesAgentSubstrateSandboxConfig
            name: gvisor-pinned
            fieldPath: status.outputs.sandbox_config_name
```

The InfraPipeline creates the control plane, the namespace and the config, then the pool. Stamp out one pool per tenant from the same chart.

## Key Configuration

These are the most important decisions when configuring a worker pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Size for actors running at once.** A suspended actor costs a snapshot in storage, not a share of a worker. When no worker has room, the API refuses a create or resume for lack of capacity until an actor suspends or the pool grows.

**Worker limits are a shared budget.** A worker hosts many actors (up to 1000 at this release), and each one spends the limits its template declares from the worker's cpu and memory limits; an actor that fits on no worker is not placed. A worker without limits bounds nothing, so set them. Requests are what each worker reserves on its node.

**Actors select the pool by its labels.** The pool's own `labels` (not the worker pod labels) are what an actor template's `worker_selector` matches. Give each tenant's pool its own label.

**Reference the control plane's worker image.** The image must match the class and the installed release; the `gvisor_worker_image` and `microvm_worker_image` outputs always do, mirror included.

**Name the SandboxConfig the pool resolves.** Leave `configRef` unset to follow the class default, or name a config; the pool holds it from deletion while it uses it, and every actor on the pool boots it: templates name no config.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|-------------------|
| **Kubernetes Namespace** | `namespace` | `spec.name` |
| **Agent Substrate** | `workerImage` | `status.outputs.gvisor_worker_image` (or `microvm_worker_image`) |
| **Agent Substrate Sandbox Config** | `sandboxClasses[].configRef.name` | `status.outputs.sandbox_config_name` |
| **Kubernetes ServiceAccount** | `template.serviceAccountName` | `status.outputs.service_account_name` |
| **Kubernetes PriorityClass** | `template.priorityClassName` | `status.outputs.priority_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `worker_pool_name` | Name of the created WorkerPool (equals `metadata.name`) | Finding the pool and its worker Deployment; dashboards and alerts per pool |
| `namespace` | The namespace the pool and its workers run in | Policies and quotas scoped to the pool's namespace |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**A shared gVisor pool** — a few workers on any nodes, on the class default. Start from the **gVisor Worker Pool** preset.

**A pool per tenant** — the tenant's own labels, namespace and tainted nodes. Start from the **Dedicated Tenant Pool** preset.

**MicroVM isolation** — each actor in its own virtual machine, on KVM nodes. Start from the **MicroVM Worker Pool** preset.

## Works With

- [**Agent Substrate**](/infra-catalog/kubernetes-agent-substrate) — the control plane the workers register with, and the worker images.
- [**Agent Substrate Sandbox Config**](/infra-catalog/kubernetes-agent-substrate-sandbox-config) — the runtime the pool runs.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — a namespace per pool or per tenant.
- [**Kubernetes Priority Class**](/infra-catalog/kubernetes-priority-class) — the workers' scheduling priority.

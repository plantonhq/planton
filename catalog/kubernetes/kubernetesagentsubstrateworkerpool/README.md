# KubernetesAgentSubstrateWorkerPool

Declares an Agent Substrate `WorkerPool`: the worker pods actors run in, many actors to a worker, sized, placed and isolated as one pool. A suspended actor leaves its worker for storage and gives its share back, so a pool serves many more actors than it runs at once. The spec mirrors the upstream `ate.dev/v1alpha1` WorkerPool field for field (pinned to the definition of the Agent Substrate release `KubernetesAgentSubstrate` composes, v0.6.0), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One WorkerPool** in the target namespace, carrying the spec exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

Substrate's controller then renders a **NetworkPolicy** that admits only Substrate's router to the workers and, once the pool's sandbox class resolves to a SandboxConfig, a **Deployment of the same name** (one pod per worker). Each worker registers with the node agent on its node and then with the API server, which places actors on workers with room for them.

## How a Product Uses It

An actor template selects the pool by the **pool's own labels**: its `worker_selector.match_labels` must match `labels` here (for example `tenant: acme`). The template declares the resources each actor spends from a worker's limits; the pool decides its class and its SandboxConfig. The product then creates actors from the template through the API; each runs on a worker of the pool with room for it, and is reached through the router with `ate-target-actor: <atespace>/<actor>`.

## Size, Placement and Runtime

- **`replicas`** is how many workers the pool runs. It is required and written on every apply.
- **`template.resources`**: the limits are each worker's capacity, a budget the actors it hosts share (up to 1000 actors per worker at this release); the requests are what the scheduler reserves per worker.
- **`template`** places workers: a node selector, node affinity and tolerations for dedicated (tenant) nodes, a PriorityClass, and the ServiceAccount they run as (unset, `ateworker`, which the controller creates in the pool's namespace).
- **`workerImage`** is the worker built with the installed release: a reference to the control plane's `gvisor_worker_image` (or `microvm_worker_image`) output.
- **`sandboxClasses`**: one class, `gvisor` or `microvm`, with the SandboxConfig the pool resolves (`configRef`, or the class default).

## Spec Overview

| Field | Purpose |
|-------|---------|
| `namespace` | Namespace the pool and its workers run in |
| `labels`, `annotations` | The WorkerPool object's own metadata; the labels are what actor templates select |
| `replicas` | How many workers |
| `workerImage` | The worker image (a KubernetesAgentSubstrate output reference, or a literal) |
| `template` | Worker pods' labels, annotations, node selector, tolerations, PriorityClass, node affinity, resources, ServiceAccount |
| `sandboxClasses[]` | `name` (`gvisor` or `microvm`) and `configRef.name` (a KubernetesAgentSubstrateSandboxConfig reference, or a literal) |

## Outputs

| Output | Description |
|--------|-------------|
| `worker_pool_name` | Name of the created WorkerPool (equals `metadata.name`), which its worker Deployment shares |
| `namespace` | Namespace the WorkerPool and its workers run in |

## Prerequisites

- `KubernetesAgentSubstrate` on the cluster: the definition, the controller, the node agent on every node and the API server.
- The target namespace (see `KubernetesNamespace`), enforcing no Pod Security level: workers run as root with the capabilities gVisor and microVM need.
- For a microVM pool, nodes that expose KVM.

See [GUIDE.md](GUIDE.md) for the decisions behind a good pool and the parity accounting, and [catalog.md](catalog.md) for the catalog page.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

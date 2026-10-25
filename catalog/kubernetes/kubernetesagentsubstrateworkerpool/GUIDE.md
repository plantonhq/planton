# KubernetesAgentSubstrateWorkerPool Guide

The judgment this guide carries: a worker pool is capacity and a boundary at once. Its workers' limits are the budget the actors running together share, its size multiplies that budget, and its placement decides whose nodes an actor's code runs on. Everything else about a worker (its image, capabilities, certificates, host directory) is Substrate's.

## How a pool serves actors

Substrate's controller renders the pool as a Deployment of the same name. Each worker pod starts, registers with the node agent on its node, and the node agent registers it with the API server. When a product creates or resumes an actor, the API server places it on a worker, of a pool matching the actor template's `worker_selector`, that still has room for it; the actor runs sandboxed inside that worker until it is suspended into storage, which gives its share of the worker back.

A worker hosts many actors at once. Room is two budgets, and placement stops at whichever runs out first:

- **The resource budget.** A worker advertises its cpu and memory limits as its capacity, and every actor placed on it spends the limits its actor template declares. An actor fits on a worker only if what is left still covers its limits.
- **The actor limit.** A worker hosts at most 1000 actors at this release, a setting of the worker binary the pool does not expose.

A worker with no limit on a resource puts no bound on it for placement, so a pool without limits packs actors until the actor limit, and the node, run out. Suspended actors cost a snapshot in storage, not a share of a worker, so size the pool for the actors **running at the same moment**, not for the actors that exist.

## Sizing a worker

- **Limits are the shared budget.** Set them to what the actors a worker should host together need, at the limits their templates declare: four actors at one CPU each fit a worker limited to four CPUs. An actor larger than any worker's whole budget is never placed on the pool.
- **Requests are reservation.** The Kubernetes scheduler reserves the requests per worker, busy or not; lower requests pack more workers on a node, at the risk of contention when every worker is busy.
- **Count the runtime.** gVisor runs inside the worker pod, so its overhead is inside the limits; leave headroom above what the actors themselves declare.
- **Few large workers or many small ones.** Larger workers pack actors densely and waste less of a node on per-pod overhead; smaller workers bound how many actors one worker's restart touches, since a stopping worker drains every actor it hosts.

## Placing a pool

- **A pool per tenant** is a pool with its own labels (`tenant: acme`), its own namespace, and its own nodes: taint the nodes, tolerate the taint in `template.tolerations`, and select them with `template.node_selector` or `template.node_affinity`. Actor templates of that tenant select the pool by its labels.
- **Substrate's own taint.** The controller always adds the toleration `ate.dev/sandboxClass=<class>:NoSchedule`, so tainting nodes that way reserves them for workers of that class without any toleration of yours.
- **MicroVM** workers request the KVM device the node agent advertises, so they land only on nodes that have it, and mount the host's TUN device; no node selector is needed for KVM.
- **Pod Security.** Workers run as root with broad capabilities and a host directory. A namespace that enforces the baseline or restricted Pod Security level refuses them.
- **Network.** The controller gives every pool a NetworkPolicy that admits only Substrate's router to the workers.
- **Identity.** Left unset, workers run as `ateworker`, a ServiceAccount the controller creates in the pool's namespace. Name your own in `template.service_account_name` when the workers need a cloud identity of their own.

## Runtime: class, image and config

- **The class** is `gvisor` or `microvm`, one per pool at this release, and it cannot change once the pool exists: recreate the pool to switch.
- **The worker image** must be built for the class and the installed release. Reference the control plane's `gvisor_worker_image` or `microvm_worker_image` output, and the pool always runs the right build, with the control plane's registry mirror already applied.
- **The SandboxConfig** the pool resolves is the one `config_ref` names, or, left unset, the class default (the newest config annotated as its class's default; `gvisor-default` for gVisor). The pool's status names the config in use, the controller starts no worker until the class resolves, and the config cannot be deleted while a pool uses it.
- **The runtime an actor boots** is the default version of the config the pool resolves: templates name no config, so the deletion guard the pool holds covers the runtime its actors boot. A suspended actor restores its memory on the version its snapshot was taken with, while the config still lists it. The node agent downloads every config's default version ahead of time on every node that can run its class (microVM configs only where `/dev/kvm` exists), whatever pools run there; any other version downloads when an actor first needs it.

## Rollouts and resizing

A change to the pool rolls its workers at most 10% at a time and never starts extra pods. A stopping worker is given up to an hour to drain its actors first, so shrinking or rolling a large pool is deliberate, not instant.

`replicas` is required and every apply writes it. The pool serves the scale subresource, but upstream treats worker autoscaling as a demo at this release (an HPA on a Prometheus metric of busy workers), so an autoscaler that resizes the pool is undone by the next apply.

## Watching it

- `kubectl -n <namespace> get workerpool <name>` shows desired, current and ready workers, the SandboxConfig in use, and whether the class resolved.
- A class that does not resolve (a config that is missing, being deleted or of another class) keeps a new pool from rendering any worker; its condition's reason says which.
- The control plane exports a count of workers by state (`idle`, `partial`, `at_capacity`, `unschedulable`) on its metrics endpoint; when no worker has room, the API answers a create or resume with ResourceExhausted until an actor suspends or the pool grows.

## Parity accounting

Built for 100% parity with the upstream WorkerPool (`ate.dev/v1alpha1`) at Agent Substrate v0.6.0, the install shape the control plane composes. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.namespace`, `metadata.labels`, `metadata.annotations` | `namespace`, `labels`, `annotations` (the envelope) |
| `spec.replicas` | `replicas` |
| `spec.workerImage` | `workerImage`, a reference to KubernetesAgentSubstrate's worker image outputs |
| `spec.template.labels`, `.annotations`, `.nodeSelector`, `.priorityClassName`, `.serviceAccountName` | `template.labels`, `.annotations`, `.nodeSelector`, `.priorityClassName` (a KubernetesPriorityClass reference), `.serviceAccountName` (a KubernetesServiceAccount reference) |
| `spec.template.tolerations`, `.nodeAffinity`, `.resources` | the same, typed from the core/v1 library |
| `spec.sandboxClasses[].name`, `.configRef.name` | `sandboxClasses[].name`, `.configRef.name` (a KubernetesAgentSubstrateSandboxConfig reference) |

### What the spec adds beyond upstream's schema

- **Typed references**: the worker image from the control plane, the SandboxConfig, the ServiceAccount and the PriorityClass, so the resource graph draws and orders them.
- **The reserved ate.dev keys, the label grammar and the bounds** (64 labels and annotations, 16 tolerations, exactly one class) checked before the apply.

### Not modeled

- **The status block** (`replicas`, `readyReplicas`, `selector`, `sandboxClasses[].configRef`, `pendingConfigRef`, `conditions`) — the controller's report.
- **The class-name immutability rule** — it compares the change with the object as it was, which only the API server can see; it refuses a renamed class there.

## Pairs well with

- **[KubernetesAgentSubstrate](../kubernetesagentsubstrate/README.md)** — the control plane: the worker images, the controller, the node agent and the API server workers register with.
- **[KubernetesAgentSubstrateSandboxConfig](../kubernetesagentsubstratesandboxconfig/README.md)** — the runtime the pool resolves and its actors boot.
- **[KubernetesNamespace](../kubernetesnamespace/README.md)** — a namespace per pool or per tenant.
- **[KubernetesPriorityClass](../kubernetespriorityclass/README.md)** — the workers' scheduling priority.

# gVisor Worker Pool

Three gVisor workers on any nodes, each with four CPUs and eight GiB for the actors it hosts, running the control plane's gVisor worker on the class default runtime. The starting point for a shared agent runtime: twelve one-CPU, two-GiB actors run at once, and every other actor waits suspended in storage at no compute cost.

## When to Use

- One product, or a few trusted ones, share an agent runtime.
- Actors need up to one CPU and two GiB of memory each (smaller actors pack more densely).

## How It Works

Substrate's controller runs three worker pods and registers each with the API server. Actor templates whose `worker_selector.match_labels` is `{workload: agents}` run on the pool. Each actor spends the limits its template declares from a worker's four CPUs and eight GiB, so a worker hosts four actors of one CPU and two GiB; a suspended actor gives its share back. The pool resolves the class default runtime (`gvisor-default`, unless a newer gVisor config is annotated as the default), and its actors boot it.

## Key Configuration Choices

- **Three workers of four CPUs and eight GiB** are twelve CPUs and 24 GiB of actors running at once; raise `replicas` for more actors running at the same moment, not for more actors in total.
- **Memory requested in full, CPU in part.** A worker short of memory kills every actor it hosts, so its memory request equals its limit; CPU only slows down when short, so one CPU is requested and four are the ceiling, which packs more workers on a node.
- **The worker image by reference**, so the pool always runs the worker built with the installed release.
- **The class default runtime.** Name a SandboxConfig in `configRef` to pin the runtime instead.

## Prerequisites

- **Agent Substrate** on the cluster (`KubernetesAgentSubstrate`).
- The target namespace exists, with no Pod Security enforcement (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, which its worker Deployment shares. |
| `<namespace>` | Namespace the pool and its workers run in. |
| `<substrate-resource-name>` | Name of the KubernetesAgentSubstrate resource whose worker image the pool runs. |

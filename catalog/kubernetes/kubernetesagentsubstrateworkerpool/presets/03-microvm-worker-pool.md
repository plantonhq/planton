# MicroVM Worker Pool

Two microVM workers, each running every actor it hosts in that actor's own lightweight virtual machine (cloud-hypervisor with Kata's guest kernel), a hardware boundary rather than gVisor's syscall boundary. The workers land only on nodes whose node agent advertises KVM.

## When to Use

- Actors run code you trust least, and each deserves a hardware-virtualized boundary.
- Actors need a full Linux kernel, or snapshots of their root filesystem.

## How It Works

The pool runs the control plane's microVM worker image. Substrate's controller requests the KVM device the node agent advertises, so workers schedule only onto nodes with KVM (bare metal, or virtual machines with nested virtualization), and mounts the host's TUN device into each worker. The pool resolves the microVM class default SandboxConfig (a microVM config you declare, annotated as the class default); actor templates with `worker_selector.match_labels: {isolation: microvm}` run on them, in the microVM class.

## Key Configuration Choices

- **The microVM worker image**, the control plane's `microvm_worker_image` output: a gVisor image cannot run microVM actors.
- **Requests equal limits.** A virtual machine's memory is committed when it boots; reserving the worker's whole budget up front keeps a node from overcommitting.
- **Count each actor's guest kernel.** Every actor boots its own virtual machine inside the worker, so the worker's limits must cover the actors' declared limits plus a guest kernel's overhead each.

## Prerequisites

- **Agent Substrate** on the cluster (`KubernetesAgentSubstrate`).
- **A microVM SandboxConfig** annotated as its class default, or named in `configRef` (`KubernetesAgentSubstrateSandboxConfig`, the **MicroVM Runtime** preset).
- Nodes with KVM (nested virtualization enabled on cloud VMs) and `/dev/net/tun`.
- The target namespace exists, with no Pod Security enforcement.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, which its worker Deployment shares. |
| `<namespace>` | Namespace the pool and its workers run in. |
| `<substrate-resource-name>` | Name of the KubernetesAgentSubstrate resource whose microVM worker image the pool runs. |

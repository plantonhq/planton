# Always-On Agent Workbench

One long-lived, gVisor-isolated sandbox for an agent that should always be there: a coding agent with its repository checked out, an assistant that keeps its memory on disk, a service an agent hosts. It keeps a 20Gi workspace across restarts and answers at a stable in-cluster name.

## When to Use

- An agent runs continuously and owns state that must outlive its pod.
- You want the sandbox declared next to the rest of your platform, not created at runtime by a product.
- Untrusted or model-generated code runs inside, so the pod gets a sandboxing runtime instead of the node's kernel.

## How It Works

The agent-sandbox controller creates one pod named after the sandbox, under the `gvisor` runtime class on GKE Sandbox nodes, and a 20Gi claim named `workspace-<agent-name>` mounted at `/workspace`. `service: true` adds a headless Service, so the pod answers at `<agent-name>.<namespace>.svc`. The controller's router reaches it by name too, which is the only way in under gVisor (port-forwarding does not work there).

## Key Configuration Choices

- **Runtime class and node placement.** `runtimeClassName: gvisor` with the GKE Sandbox node selector and toleration. On other clusters use the class your sandboxing runtime installs (`kata-qemu`, `kata-fc`, AKS `kata-vm-isolation`) and its node labels.
- **No Kubernetes credentials.** `automountServiceAccountToken: false`, so code in the sandbox cannot call the Kubernetes API. Reference a KubernetesServiceAccount in `serviceAccountName` and turn it back on only for an agent that needs it.
- **Hardened container.** Non-root, no privilege escalation, every capability dropped, the runtime's seccomp profile.
- **Workspace size.** 20Gi on the cluster's default storage class; set `storageClassName` for a faster or encrypted class. The volume list is fixed once the sandbox exists.

## Prerequisites

- **Agent Sandbox Controller** on the cluster.
- Nodes running the sandboxing runtime (a GKE node pool with GKE Sandbox enabled, or Kata).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<agent-name>` | Name of the sandbox, its pod and its Service. |
| `<namespace>` | Namespace the sandbox lives in. |
| `<agent-image>` | The agent's container image, serving on port 8888. |

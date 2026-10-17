# Time-Boxed Sandbox

A gVisor-isolated sandbox that ends itself at a set time: an evaluation run, a demo environment, a contractor's box. At its expiry the controller deletes the pod and keeps the Sandbox object, marked expired, so nothing keeps running past the deadline and the record of it stays.

## When to Use

- A sandbox should exist for a known window, and forgetting to delete it must not leave it running.
- You want the expiry visible on the object (`Ready=False`, reason `SandboxExpired`) for whoever audits it later.

## How It Works

At `shutdownTime` the controller deletes the sandbox's pod and Service and, with `shutdownPolicy: Retain`, keeps the Sandbox object with its expiry recorded. The time is absolute: redeploying the same manifest after the deadline creates a sandbox that is torn down at once, so move the time forward to extend the window.

## Key Configuration Choices

- **`shutdownTime`.** An RFC 3339 time; the spec refuses anything else before it reaches the cluster.
- **`Retain` over `Delete`.** `Delete` removes the Sandbox object itself, which Planton still declares, so the next apply would create it again (already expired). Keep `Retain` for a declared sandbox; `Delete` suits sandboxes a product creates at runtime.
- **No volumes.** A time-boxed sandbox usually carries nothing worth keeping; add a volume claim template if it does, and remember the claim lives until the Sandbox object is deleted.

## Prerequisites

- **Agent Sandbox Controller** on the cluster.
- Nodes running the sandboxing runtime (a GKE node pool with GKE Sandbox enabled, or Kata).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<sandbox-name>` | Name of the sandbox and its pod. |
| `<namespace>` | Namespace the sandbox lives in. |
| `<workload-image>` | The container image the sandbox runs. |

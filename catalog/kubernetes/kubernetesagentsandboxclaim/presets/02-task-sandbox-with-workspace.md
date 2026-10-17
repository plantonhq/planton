# Task Sandbox With Workspace

A sandbox for one bounded task: it receives the task's ID, gets a 50Gi workspace instead of the template's, and ends itself at a deadline or an hour after its work finishes. Both customizations force a cold start, which is the price of a sandbox shaped for the task.

## When to Use

- A long-running task (a large build, a data job, an evaluation) needs more disk than the pool's template gives, and its own identity.
- The sandbox must not outlive the task.

## How It Works

Because the claim sets an environment variable and a volume claim template, the controller starts a fresh sandbox from the pool's template instead of handing over a warm member (a running pod cannot take either). The sandbox is named after the claim, its pod gets `TASK_ID`, and its `workspace` claim is 50Gi. At `shutdownTime`, or one hour after the sandbox's pod finishes (`ttlSecondsAfterFinished`), the controller deletes the sandbox and keeps the claim, marked expired.

## Key Configuration Choices

- **The template must allow it.** The pool's template needs `envVarsInjectionPolicy` `Allowed` (or `Overrides`) and `volumeClaimTemplatesPolicy` `Overrides` to replace its `workspace` template by name. Otherwise the claim is refused with reason `EnvVarsInjectionRejected` or `VolumeClaimTemplatesError`.
- **`Retain` for a declared claim.** `Delete` would remove an object Planton still declares, and the next apply would claim again.
- **Plain values only.** Claim variables land in the custom resource as written; keep secrets in the template's `secretKeyRef` variables.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- **Agent Sandbox Warm Pool** in the same namespace, whose template allows variables and volume overrides.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<task-name>` | Name of the claim (and of its sandbox), and the task ID the sandbox receives. |
| `<namespace>` | Namespace the claim, its pool and its sandbox share. |
| `<pool-name>` | Name of the warm pool to claim from. |

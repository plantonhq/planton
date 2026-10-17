# Suspendable Notebook

A Jupyter notebook sandbox with its own 10Gi work volume, meant to be suspended when idle. Suspending deletes the pod, so it gives back its CPU and memory, while the notebook's files stay on the claim for when it resumes.

## When to Use

- A data scientist or an agent works in a notebook in bursts, and paying for an idle pod makes no sense.
- The work must survive between sessions without a separate storage resource to manage.

## How It Works

The controller runs one Jupyter pod named after the sandbox and creates a claim named `work-<notebook-name>`, mounted at `/home/jovyan/work`. Changing `operatingMode` to `Suspended` and redeploying deletes the pod and keeps the Sandbox object and the claim. Changing it back to `Running` starts a fresh pod on the same claim, with every file where it was left.

## Key Configuration Choices

- **`operatingMode`.** The one setting you toggle. The Sandbox's `Ready` and `Suspended` conditions report when the controller has reached the state you declared.
- **No token on the notebook server.** `--IdentityProvider.token=` turns Jupyter's own login off, because callers reach the sandbox through the agent-sandbox router, whose authorization (TokenReview or scoped tokens) is the gate. Keep a token if anything else can reach the pod.
- **`fsGroup: 100`.** The Jupyter images' `users` group, so the notebook user can write to the claim.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with its router to reach the notebook.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<notebook-name>` | Name of the sandbox and its pod. |
| `<namespace>` | Namespace the notebook lives in. |

# Coding Agent Template

The blueprint for coding-agent sandboxes: gVisor isolation, a 10Gi workspace per sandbox, a fence that lets the agent resolve names and reach HTTPS endpoints (model APIs, package registries, git hosts) and nothing private, and claims allowed to bring their own settings and a larger workspace.

## When to Use

- Each task or user gets a sandbox with a repository checked out and room to build.
- The agent needs the internet over HTTPS, but must never reach the cluster's private network or the cloud metadata server.
- A product passes per-task settings (a task ID, a model choice) or a bigger disk when it claims a sandbox.

## How It Works

Each sandbox runs the agent under gVisor and owns a claim named `workspace-<sandbox>` mounted at `/workspace`. The template's NetworkPolicy admits only the controller's router on the `http` port. It lets sandboxes out to cluster DNS and to TCP 443 anywhere outside the private ranges and the link-local metadata range. `envVarsInjectionPolicy: Allowed` lets a claim add environment variables the template doesn't define. `volumeClaimTemplatesPolicy: Overrides` lets a claim replace the workspace template, for example with a larger size.

## Key Configuration Choices

- **Claims that customize start cold.** Variables and volumes are baked into a pod before it starts, so a claim that sets either gets a fresh sandbox instead of a warm one. Keep per-task settings few, and size the pool for the claims that set none.
- **Egress by port, not by host.** NetworkPolicy matches addresses, not names. For host-level egress control (only `api.example.com`), put an egress proxy or a mesh policy in front, and run the template `Unmanaged` if that system owns the fence.
- **`Allowed` over `Overrides` for variables.** Claims can add, never change, what the template sets, so the template's own settings stay authoritative.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on and its router.
- Nodes running GKE Sandbox (or swap in your cluster's Kata runtime class and node labels).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<template-name>` | Name of the template, which warm pools reference. |
| `<namespace>` | Namespace the template, its pools and their sandboxes live in. |
| `<agent-image>` | The coding agent's image, serving on port 8888. |

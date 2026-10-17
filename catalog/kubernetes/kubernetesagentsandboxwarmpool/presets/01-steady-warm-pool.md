# Steady Warm Pool

Three sandboxes from one template, always running and ready, so a product's claims bind in about the time it takes to update an object. When the template changes, members are replaced as they are claimed, so the pool never dips below three ready.

## When to Use

- A product hands out sandboxes at a steady rate and every claim should start warm.
- Template changes are routine (an image bump) and claims may get the previous shape for a while.

## How It Works

The controller keeps three Sandboxes named `<pool-name>-<random>` running the template's pod. A claim against the pool takes a ready member, which then belongs to the claim, and the controller immediately starts a replacement. With `OnReplenish`, a template change leaves existing members alone; every replacement is built from the new template, so the pool converges as it is used.

## Key Configuration Choices

- **Size for the gap between refills.** A replacement takes a pod's start-up time. Three members absorb three claims arriving within that time; a fourth claim starts cold.
- **`OnReplenish` over `Recreate`.** No dip in ready members during a rollout. Switch to `Recreate` when a template change must reach every new claim at once (a security fix).
- **Cost.** Each member reserves the template's resources while it waits. Three idle members cost three sandboxes.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- **Agent Sandbox Template** in the same namespace.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, which claims and agent-sandbox's SDKs reference. |
| `<namespace>` | Namespace the pool, its template, its members and its claims share. |
| `<template-name>` | Name of the template the members are built from (or reference a KubernetesAgentSandboxTemplate with `valueFrom`). |

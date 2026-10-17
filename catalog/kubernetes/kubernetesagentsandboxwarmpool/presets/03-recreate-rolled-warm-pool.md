# Recreate-Rolled Warm Pool

Five ready sandboxes, replaced at once whenever their template changes, so no claim ever gets a sandbox built from an old template. Right when a template change is a fix that must reach every new sandbox immediately.

## When to Use

- Template changes carry security or correctness fixes (a patched image, a tightened security context) that every new claim must get.
- A short dip in ready members during a rollout is acceptable.

## How It Works

The controller compares each unclaimed member's blueprint with the template's. With `Recreate`, a change to the template's pod, volume claim templates or `service` deletes every stale unclaimed member at once, and the controller starts replacements from the new template. Until they are ready, claims may start cold. Claimed sandboxes are never touched; they belong to their claims and keep what they started with.

## Key Configuration Choices

- **`Recreate` trades a dip for freshness.** Every new claim gets the current template, at the cost of a window with fewer ready members after each template change.
- **Size for the dip.** Five members refill in about one pod start-up time; claims that arrive in that window start cold.
- **Not every change is a rollout.** Labels, annotations and the template's network or claim policies never make members stale; network policy changes reach every sandbox at once anyway.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- **Agent Sandbox Template** in the same namespace.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, which claims and agent-sandbox's SDKs reference. |
| `<namespace>` | Namespace the pool, its template, its members and its claims share. |
| `<template-name>` | Name of the template the members are built from. |

# Autoscaled Warm Pool

A pool whose size an autoscaler owns. `replicas` is left unset, so the cluster starts the pool at one member and Planton never writes the size again. A HorizontalPodAutoscaler or a KEDA ScaledObject then grows and shrinks the pool through its scale endpoint, and a redeploy never resets it.

## When to Use

- Demand for sandboxes swings through the day, and idle members are a cost worth trimming.
- You want the pool at zero outside working hours (KEDA's cron trigger) or sized from a queue of pending tasks.

## How It Works

The pool serves the scale subresource: its size at `.spec.replicas`, its current members at `.status.replicas`, and the label selector of its members' pods at `.status.selector`. An autoscaler targets it by API version `extensions.agents.x-k8s.io/v1beta1`, kind `SandboxWarmPool` and the pool's name. Because the spec leaves `replicas` out, Planton's server-side apply does not own the field, so the autoscaler's changes stand.

## Key Configuration Choices

- **Pair it with an autoscaler.** A KubernetesHorizontalPodAutoscaler with `scaleTargetRef` set to the pool (its name from this pool's `sandbox_warm_pool_name` output) scales on the members' CPU or memory. KEDA scales on anything else, including a schedule.
- **Never set `replicas` later.** Setting it hands the field back to Planton, and every apply would reset the autoscaler's size.
- **Scale to zero carefully.** An empty pool means every claim starts cold. Keep a floor of one or two during the hours claims arrive.

## Prerequisites

- **Agent Sandbox Controller** on the cluster, with extensions on.
- **Agent Sandbox Template** in the same namespace.
- An autoscaler: the cluster's HorizontalPodAutoscaler with metrics-server, or KEDA.
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<pool-name>` | Name of the pool, which claims, agent-sandbox's SDKs and the autoscaler reference. |
| `<namespace>` | Namespace the pool, its template, its members and its claims share. |
| `<template-name>` | Name of the template the members are built from. |

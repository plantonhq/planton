---
title: "Tekton Operator with Pipelines, Triggers, and Dashboard"
description: "This preset deploys the Tekton Operator with all three core components enabled: Pipelines, Triggers, and Dashboard. The operator manages the lifecycle of Tekton components in fixed namespaces..."
type: "preset"
rank: "01"
presetSlug: "01-pipelines-and-dashboard"
componentSlug: "tekton-operator"
componentTitle: "Tekton Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Tekton Operator with Pipelines, Triggers, and Dashboard

This preset deploys the Tekton Operator with all three core components enabled: Pipelines, Triggers, and Dashboard. The operator manages the lifecycle of Tekton components in fixed namespaces (`tekton-operator` and `tekton-pipelines`) that cannot be customized.

## When to Use

- You need a complete CI/CD platform on Kubernetes using Tekton
- You want event-driven pipeline execution via Triggers (e.g., GitHub webhooks)
- You want a web UI for monitoring and managing pipelines via the Dashboard

## Key Configuration Choices

- **All components enabled** -- Pipelines (core CI/CD engine), Triggers (event-driven execution), and Dashboard (web UI) are all active
- **Operator version** (`v0.78.0`) -- pinned for reproducibility; check [Tekton releases](https://github.com/tektoncd/operator/releases) for updates
- **No namespace field** -- the Tekton Operator uses fixed namespaces managed by the operator itself (`tekton-operator`, `tekton-pipelines`)
- **No dashboard ingress** -- the Dashboard is deployed but not exposed externally; configure `dashboardIngress` separately if external access is needed
- **Resource requests** (`100m` CPU, `128Mi` memory) -- conservative baseline for the operator pod
- **Resource limits** (`500m` CPU, `512Mi` memory) -- matches proto recommended defaults

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

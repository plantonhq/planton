---
title: "Standard GitHub Actions Runner Scale Set Controller"
description: "This preset deploys the GitHub Actions Runner Scale Set Controller with recommended default resources. The controller manages AutoScalingRunnerSet and EphemeralRunner custom resources, enabling..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "gha-runner-scale-set-controller"
componentTitle: "GHA Runner Scale Set Controller"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard GitHub Actions Runner Scale Set Controller

This preset deploys the GitHub Actions Runner Scale Set Controller with recommended default resources. The controller manages AutoScalingRunnerSet and EphemeralRunner custom resources, enabling self-hosted GitHub Actions runners that scale dynamically based on workflow demand.

## When to Use

- You need self-hosted GitHub Actions runners on Kubernetes
- You want dynamic scaling of runner pods based on queued workflow jobs
- This is the controller only -- deploy `KubernetesGhaRunnerScaleSet` resources separately for actual runner pods

## Key Configuration Choices

- **Namespace** (`gha-runner-system`) -- dedicated namespace for the controller; runner pods can be deployed in different namespaces
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`100m` CPU, `128Mi` memory) -- conservative baseline for the controller pod
- **Resource limits** (`500m` CPU, `512Mi` memory) -- matches proto recommended defaults
- **Default flags** -- controller watches all namespaces; single replica with no leader election
- **No metrics** -- metrics endpoints are not configured; enable via the `metrics` field if needed

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

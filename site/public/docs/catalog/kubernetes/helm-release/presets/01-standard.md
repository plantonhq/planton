---
title: "Standard Helm Release"
description: "This preset deploys a Helm chart using the generic KubernetesHelmRelease component. Use this as an escape hatch for deploying any Helm chart that does not have a dedicated OpenMCF component."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "helm-release"
componentTitle: "Helm Release"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Helm Release

This preset deploys a Helm chart using the generic KubernetesHelmRelease component. Use this as an escape hatch for deploying any Helm chart that does not have a dedicated OpenMCF component.

## When to Use

- Deploying third-party Helm charts not covered by dedicated OpenMCF components
- Community or vendor charts that you want to manage through OpenMCF's KRM workflow
- Charts where custom `values` overrides are needed

## Key Configuration Choices

- **Create namespace** (`true`) -- the target namespace is created if it does not exist
- **No custom values** -- deploy with chart defaults; add `values` map entries for overrides

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the Helm release | Your namespace management |
| `<your-helm-repo-url>` | Helm chart repository URL (e.g., `https://charts.bitnami.com/bitnami`) | Chart documentation |
| `<your-chart-name>` | Chart name within the repository (e.g., `redis`) | Chart documentation |
| `<your-chart-version>` | Chart version to deploy (e.g., `18.6.1`) | `helm search repo` or chart documentation |

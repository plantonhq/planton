---
title: "Single Resource Manifest"
description: "This preset deploys a raw Kubernetes manifest using the generic KubernetesManifest component. Use this as an escape hatch for deploying any Kubernetes resource that does not have a dedicated OpenMCF..."
type: "preset"
rank: "01"
presetSlug: "01-single-resource"
componentSlug: "manifest"
componentTitle: "Manifest"
provider: "kubernetes"
icon: "package"
order: 1
---

# Single Resource Manifest

This preset deploys a raw Kubernetes manifest using the generic KubernetesManifest component. Use this as an escape hatch for deploying any Kubernetes resource that does not have a dedicated OpenMCF component.

## When to Use

- Deploying custom resources (CRDs) not covered by OpenMCF components
- One-off Kubernetes resources (ConfigMaps, ServiceAccounts, RBAC rules) that do not warrant a dedicated component
- Multi-document manifests (separate resources with `---` delimiters)

## Key Configuration Choices

- **Raw manifest YAML** -- the `manifestYaml` field accepts any valid Kubernetes YAML, including multi-document manifests separated by `---`
- **Namespace** -- resources without their own namespace metadata will be deployed to this namespace
- **Example** -- shows a ConfigMap; replace with any Kubernetes resource

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Default namespace for resources that do not specify their own | Your namespace management |

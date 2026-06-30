---
title: "Presets"
description: "Ready-to-deploy configuration presets for Helm Release"
type: "preset-list"
componentSlug: "helm-release"
componentTitle: "Helm Release"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Helm Release"
    excerpt: "This preset deploys a Helm chart using the generic KubernetesHelmRelease component. Use this as an escape hatch for deploying any Helm chart that does not have a dedicated Planton component."
---

# Helm Release Presets

Ready-to-deploy configuration presets for Helm Release. Each preset is a complete manifest you can copy, customize, and deploy.

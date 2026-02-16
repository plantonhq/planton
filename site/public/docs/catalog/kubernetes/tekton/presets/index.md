---
title: "Presets"
description: "Ready-to-deploy configuration presets for Tekton"
type: "preset-list"
componentSlug: "tekton"
componentTitle: "Tekton"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Tekton Pipelines"
    excerpt: "This preset deploys Tekton pipeline resources (PipelineRuns, TaskRuns) with default resources. Use this alongside `KubernetesTektonOperator` which manages the Tekton control plane."
---

# Tekton Presets

Ready-to-deploy configuration presets for Tekton. Each preset is a complete manifest you can copy, customize, and deploy.

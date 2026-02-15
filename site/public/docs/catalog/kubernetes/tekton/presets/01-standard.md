---
title: "Standard Tekton Pipelines"
description: "This preset deploys Tekton pipeline resources (PipelineRuns, TaskRuns) with default resources. Use this alongside `KubernetesTektonOperator` which manages the Tekton control plane."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "tekton"
componentTitle: "Tekton"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Tekton Pipelines

This preset deploys Tekton pipeline resources (PipelineRuns, TaskRuns) with default resources. Use this alongside `KubernetesTektonOperator` which manages the Tekton control plane.

## When to Use

- You need to deploy Tekton pipeline resources on a cluster that already has the Tekton Operator running
- Standard resource allocation is sufficient for pipeline execution

## Key Configuration Choices

- **Namespace** (`tekton-pipelines`) -- the standard namespace where Tekton components run
- **Default resources** -- proto recommended defaults; pipeline tasks themselves define their own resource requirements

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

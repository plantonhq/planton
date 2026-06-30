# Kubernetes Tekton Operator - Terraform Module

This directory contains the Terraform implementation for deploying the Tekton Operator on Kubernetes.

## Overview

The Terraform module installs the Tekton Operator using official release manifests. The operator manages the lifecycle of Tekton components (Pipelines, Triggers, Dashboard) via the TektonConfig CRD.

## ⚠️ Important: Fixed Namespace Architecture

**The Tekton Operator uses fixed namespaces** that are managed by the operator itself:

| Component | Namespace |
|-----------|-----------|
| Tekton Operator | `tekton-operator` |
| Tekton Pipelines | `tekton-pipelines` |
| Tekton Triggers | `tekton-pipelines` |
| Tekton Dashboard | `tekton-pipelines` |

**These namespaces are automatically created by the operator and cannot be customized.**

## Module Structure

```
iac/tf/
├── main.tf           # Main resource definitions
├── variables.tf      # Input variable definitions
├── locals.tf         # Local value computations
├── outputs.tf        # Output value definitions
├── provider.tf       # Provider configuration
└── README.md         # This file
```

## Prerequisites

- **Terraform**: Version 1.0 or later
- **kubectl provider**: For applying raw YAML manifests
- **Kubernetes Cluster**: Target cluster with kubectl access

## Quick Start

### 1. Create tfvars File

Create a `terraform.tfvars` file:

```hcl
metadata = {
  name = "tekton-operator"
  id   = "tekton-op-dev"
}

spec = {
  container = {
    resources = {
      requests = {
        cpu    = "100m"
        memory = "128Mi"
      }
      limits = {
        cpu    = "500m"
        memory = "512Mi"
      }
    }
  }
  components = {
    pipelines = true
    triggers  = true
    dashboard = true
  }
  operator_version = "v0.78.0"  # Optional, defaults to v0.78.0
}
```

### 2. Initialize and Apply

```bash
terraform init
terraform plan
terraform apply
```

### 3. Verify Deployment

```bash
kubectl get pods -n tekton-operator
kubectl get tektonconfig
kubectl get pods -n tekton-pipelines
```

## Input Variables

| Variable | Description | Type | Required | Default |
|----------|-------------|------|----------|---------|
| `metadata.name` | Resource name | string | Yes | - |
| `metadata.id` | Resource ID | string | No | - |
| `metadata.org` | Organization | string | No | - |
| `metadata.env` | Environment | string | No | - |
| `spec.container.resources` | Container resource allocation | object | Yes | - |
| `spec.components.pipelines` | Enable Tekton Pipelines | bool | Yes | - |
| `spec.components.triggers` | Enable Tekton Triggers | bool | Yes | - |
| `spec.components.dashboard` | Enable Tekton Dashboard | bool | Yes | - |
| `spec.operator_version` | Tekton Operator version | string | No | v0.78.0 |

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Components namespace (tekton-pipelines) |
| `operator_namespace` | Operator namespace (tekton-operator) |
| `tekton_config_name` | Name of TektonConfig resource |
| `tekton_profile` | Profile used (all, basic, lite) |
| `pipelines_controller_service` | Pipelines controller service name |
| `triggers_controller_service` | Triggers controller service name |
| `dashboard_service` | Dashboard service name |
| `dashboard_port_forward_command` | Port-forward command for dashboard |

## Component Profiles

The module selects a profile based on enabled components:

| Components Enabled | Profile |
|-------------------|---------|
| Pipelines + Triggers + Dashboard | all |
| Pipelines + Triggers | basic |
| Pipelines only | lite |

## Cleanup

Remove the Tekton Operator:

```bash
terraform destroy
```

> **Warning**: This will remove the operator and all Tekton components.

## References

- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest)
- [Tekton Documentation](https://tekton.dev/docs/)
- [Tekton Operator GitHub](https://github.com/tektoncd/operator)

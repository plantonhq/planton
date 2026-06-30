# KubernetesRookCephOperator Terraform Module

This Terraform module deploys the Rook Ceph Operator on a Kubernetes cluster.

## Overview

The module:
1. Optionally creates a namespace for the operator
2. Deploys the Rook Ceph Operator Helm chart
3. Configures CSI drivers based on spec

## Prerequisites

- Terraform 1.0+
- Kubernetes cluster access
- `hashicorp/kubernetes` provider
- `hashicorp/helm` provider

## Usage

### With Planton CLI

```bash
planton terraform apply \
  --manifest manifest.yaml \
  --stack org/project/env
```

### Standalone

```hcl
module "rook_ceph_operator" {
  source = "./path/to/module"

  metadata = {
    name = "rook-ceph-operator"
  }

  spec = {
    namespace        = "rook-ceph"
    create_namespace = true
    operator_version = "v1.16.6"
    crds_enabled     = true
    container = {
      resources = {
        limits = {
          cpu    = "500m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "200m"
          memory = "128Mi"
        }
      }
    }
    csi = {
      enable_rbd_driver    = true
      enable_cephfs_driver = true
    }
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| metadata | Resource metadata | object | - | yes |
| spec | Operator specification | object | - | yes |

### Metadata Object

| Attribute | Description | Type | Required |
|-----------|-------------|------|----------|
| name | Resource name | string | yes |
| id | Resource ID | string | no |
| org | Organization | string | no |
| env | Environment | string | no |

### Spec Object

| Attribute | Description | Type | Default |
|-----------|-------------|------|---------|
| namespace | Target namespace | string | "rook-ceph" |
| create_namespace | Create namespace | bool | false |
| operator_version | Helm chart version | string | "v1.16.6" |
| crds_enabled | Let Helm manage CRDs | bool | true |
| container | Container resources | object | - |
| csi | CSI configuration | object | - |

## Outputs

| Name | Description |
|------|-------------|
| namespace | Namespace where operator is deployed |
| helm_release_name | Name of the Helm release |
| webhook_service | Webhook service name |
| port_forward_command | Port-forward command for metrics |

## Resources Created

- **kubernetes_namespace_v1.rook_ceph_operator** (conditional): Namespace for operator
- **helm_release.rook_ceph_operator**: Helm release for operator chart

## Providers

| Name | Version |
|------|---------|
| kubernetes | >= 2.0.0 |
| helm | >= 2.0.0 |

## Example

```bash
# Initialize
terraform init

# Plan
terraform plan -var-file=terraform.tfvars

# Apply
terraform apply -var-file=terraform.tfvars

# Destroy
terraform destroy -var-file=terraform.tfvars
```

## Post-Deployment

After the operator is deployed, create CephCluster resources:

```bash
kubectl apply -f cephcluster.yaml
```

## Troubleshooting

### Operator Pod Not Starting

```bash
kubectl get pods -n rook-ceph -l app=rook-ceph-operator
kubectl logs -n rook-ceph -l app=rook-ceph-operator
```

### Helm Release Issues

```bash
terraform state show helm_release.rook_ceph_operator
helm list -n rook-ceph
```

## Additional Resources

- [Component README](../../README.md)
- [Examples](../../examples.md)
- [Rook Documentation](https://rook.io/docs/rook/latest/)

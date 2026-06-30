# Harbor Kubernetes - Terraform Module

This Terraform module deploys Harbor cloud-native registry on Kubernetes.

## Prerequisites

- Terraform >= 1.5
- Kubernetes cluster with configured kubectl access
- Helm 3.x

## Usage

```hcl
module "harbor" {
  source = "./path/to/module"

  harbor_kubernetes = {
    metadata = {
      name = "my-harbor"
    }
    spec = {
      database = {
        is_external = false
      }
      cache = {
        is_external = false
      }
      storage = {
        type = "filesystem"
        filesystem = {
          disk_size = "100Gi"
        }
      }
    }
  }
}
```

## Outputs

- `namespace`: Kubernetes namespace where Harbor is deployed
- `core_service`: Harbor Core service name
- `portal_service`: Harbor Portal service name
- `registry_service`: Harbor Registry service name
- `port_forward_command`: kubectl command for local access

## Note

This is a simplified Terraform structure. For full production deployment with Harbor Helm chart,
external databases, and ingress configuration, refer to the Pulumi module implementation or
extend this Terraform module with additional Helm release resources.

## References

- [Harbor Documentation](https://goharbor.io/docs/)
- [Harbor Helm Chart](https://github.com/goharbor/harbor-helm)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)


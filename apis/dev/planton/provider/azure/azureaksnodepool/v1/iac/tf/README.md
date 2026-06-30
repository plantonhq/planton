# Azure AKS Node Pool Terraform Module

This Terraform module deploys an Azure Kubernetes Service (AKS) node pool to an existing AKS cluster.

## Features

- **Flexible VM Sizing**: Support for all Azure VM sizes (D-series, E-series, F-series, GPU, etc.)
- **Autoscaling**: Configure cluster autoscaler with min/max node counts
- **High Availability**: Multi-AZ deployment across availability zones
- **Spot Instances**: Cost optimization with Azure Spot VMs
- **System/User Modes**: Support for both system and user node pools
- **Windows Support**: Deploy Windows node pools for legacy .NET apps

## Prerequisites

- Terraform >= 1.0
- Azure CLI authenticated
- Existing AKS cluster
- Azure subscription with appropriate permissions

## Usage

### Basic Example

```hcl
module "app_node_pool" {
  source = "./iac/tf"

  metadata = {
    name = "app-pool"
    env  = "production"
  }

  spec = {
    cluster_name       = "production-aks-cluster"
    vm_size            = "Standard_D8s_v3"
    initial_node_count = 2
    availability_zones = ["1", "2", "3"]
  }
}
```

### Production Example with Autoscaling

```hcl
module "production_pool" {
  source = "./iac/tf"

  metadata = {
    name = "production-pool"
    env  = "production"
    labels = {
      team        = "platform"
      cost-center = "engineering"
    }
  }

  spec = {
    cluster_name       = "production-aks-cluster"
    vm_size            = "Standard_D8s_v3"  # 8 vCPU, 32 GB RAM
    initial_node_count = 2

    autoscaling = {
      min_nodes = 2
      max_nodes = 10
    }

    availability_zones = ["1", "2", "3"]
    mode               = "USER"
    os_type            = "LINUX"
    spot_enabled       = false
  }
}
```

### Spot Instance Pool

```hcl
module "spot_pool" {
  source = "./iac/tf"

  metadata = {
    name = "spot-batch-pool"
  }

  spec = {
    cluster_name       = "production-aks-cluster"
    vm_size            = "Standard_D8s_v3"
    initial_node_count = 1

    autoscaling = {
      min_nodes = 0   # Allow scale to zero
      max_nodes = 20
    }

    availability_zones = ["1", "2", "3"]
    mode               = "USER"
    spot_enabled       = true  # Enable Spot for cost savings
  }
}
```

### System Node Pool

```hcl
module "system_pool" {
  source = "./iac/tf"

  metadata = {
    name = "system-secondary"
  }

  spec = {
    cluster_name       = "production-aks-cluster"
    vm_size            = "Standard_D4s_v5"
    initial_node_count = 3
    availability_zones = ["1", "2", "3"]
    mode               = "SYSTEM"  # System mode for cluster components
    os_type            = "LINUX"   # System pools must be Linux
    spot_enabled       = false     # Spot not supported for system pools
  }
}
```

### Windows Node Pool

```hcl
module "windows_pool" {
  source = "./iac/tf"

  metadata = {
    name = "windows-pool"
  }

  spec = {
    cluster_name       = "production-aks-cluster"
    vm_size            = "Standard_D4s_v3"
    initial_node_count = 2

    autoscaling = {
      min_nodes = 2
      max_nodes = 5
    }

    availability_zones = ["1", "2"]
    mode               = "USER"      # Windows must be User mode
    os_type            = "WINDOWS"
    spot_enabled       = false
  }
}
```

## Input Variables

### metadata

Object containing resource metadata:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| name | string | yes | - | Node pool name |
| env | string | no | "default" | Environment (dev, staging, prod) |
| labels | map(string) | no | {} | Additional labels |

### spec

Object containing node pool specification:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| cluster_name | string | yes | - | Parent AKS cluster name |
| vm_size | string | yes | - | Azure VM size (e.g., Standard_D8s_v3) |
| initial_node_count | number | yes | - | Initial number of nodes (>= 1) |
| autoscaling | object | no | null | Autoscaling configuration |
| autoscaling.min_nodes | number | yes* | - | Minimum nodes (>= 0) |
| autoscaling.max_nodes | number | yes* | - | Maximum nodes (> 0) |
| availability_zones | list(string) | no | [] | Availability zones (e.g., ["1", "2", "3"]) |
| os_type | string | no | "LINUX" | OS type: "LINUX" or "WINDOWS" |
| mode | string | no | "USER" | Pool mode: "SYSTEM" or "USER" |
| spot_enabled | bool | no | false | Enable Spot instances |

*Required if autoscaling object is specified.

## Outputs

| Name | Description |
|------|-------------|
| node_pool_name | Name of the created node pool |
| agent_pool_resource_id | Azure Resource Manager ID of the node pool |
| max_pods_per_node | Maximum pods per node |

## Resource Group

The module expects the AKS cluster to exist in a resource group named `rg-<cluster_name>`. This follows Planton's naming convention.

If your cluster uses a different resource group name, you'll need to modify `locals.tf`.

## VM Size Recommendations

### General Purpose
- **Standard_D2s_v3**: 2 vCPU, 8 GB - Dev/test
- **Standard_D4s_v3**: 4 vCPU, 16 GB - Small production workloads
- **Standard_D8s_v3**: 8 vCPU, 32 GB - Standard production workloads

### Compute Optimized
- **Standard_F8s_v2**: 8 vCPU, 16 GB - CPU-intensive workloads
- **Standard_F16s_v2**: 16 vCPU, 32 GB - High-performance computing

### Memory Optimized
- **Standard_E4s_v5**: 4 vCPU, 32 GB - Memory-intensive apps
- **Standard_E8s_v5**: 8 vCPU, 64 GB - In-memory databases, caching

### GPU
- **Standard_NC6s_v3**: 6 vCPU, 112 GB, 1x V100 - AI/ML training
- **Standard_NC4as_T4_v3**: 4 vCPU, 28 GB, 1x T4 - Cost-effective inference

## Availability Zones

- **Production**: Always use 3 zones: `["1", "2", "3"]`
- **Dev/Test**: Single zone acceptable: `["1"]`
- **No zones**: `[]` (uses regional defaults, no zone pinning)

## Autoscaling Best Practices

### User Pools
```hcl
autoscaling = {
  min_nodes = 2   # Baseline capacity
  max_nodes = 10  # Peak capacity
}
```

### Spot Pools (can scale to zero)
```hcl
autoscaling = {
  min_nodes = 0   # Scale to zero when idle
  max_nodes = 20  # Aggressive scaling for burst
}
```

### System Pools
System pools typically use fixed sizing (no autoscaling) for stability.

## Limitations

- **System pools cannot use Spot instances**
- **System pools cannot scale to zero**
- **Windows pools must be User mode**
- **Availability zones cannot be changed after creation** - requires new pool
- **VM size is immutable** - requires new pool to change

## Security Considerations

- Node pools inherit cluster's network configuration
- System pools are automatically tainted with `CriticalAddonsOnly=true:NoSchedule`
- Use separate pools to isolate workload types
- Apply pod security policies at cluster level

## Cost Optimization

1. **Use Spot instances** for fault-tolerant workloads (30-90% savings)
2. **Enable autoscaling** with appropriate min/max bounds
3. **Scale to zero** for dev/test and batch workloads
4. **Right-size VMs** - don't over-provision
5. **Use Azure Reservations** for predictable workloads (save up to 72%)

## Troubleshooting

### Pool Creation Fails

```shell
# Check if cluster exists
az aks show --name <cluster-name> --resource-group rg-<cluster-name>

# Check available VM sizes in region
az vm list-sizes --location <region>
```

### Nodes Not Joining Cluster

```shell
# Check node pool status
az aks nodepool show \
  --cluster-name <cluster-name> \
  --resource-group rg-<cluster-name> \
  --name <pool-name>

# Check Kubernetes nodes
kubectl get nodes -l agentpool=<pool-name>
```

### Autoscaler Not Scaling

```shell
# Check autoscaler logs
kubectl logs -n kube-system -l app=cluster-autoscaler

# Check for pod disruption budgets blocking scale-down
kubectl get pdb --all-namespaces
```

## Examples Directory

See `../examples.md` for more detailed examples including:
- GPU node pools
- Memory-optimized pools
- Multi-zone configurations
- Spot instance best practices

## Contributing

When modifying this module:
1. Update input variable documentation
2. Run `terraform fmt`
3. Run `terraform validate`
4. Test with a real cluster
5. Update this README

## License

See the main project LICENSE file.

## Support

For issues or questions:
- Check the [main component README](../../README.md)
- Review [research documentation](../../docs/README.md)
- Consult [Azure AKS documentation](https://docs.microsoft.com/en-us/azure/aks/)


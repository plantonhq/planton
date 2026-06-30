# Terraform Module: KubernetesNeo4j

This Terraform module deploys **Neo4j Community Edition** on Kubernetes using the official Neo4j Helm chart. It provides a simple, production-ready graph database deployment with persistent storage, optional external access, and configurable memory settings.

## Overview

Neo4j is a leading graph database that excels at storing and querying highly connected data. This module deploys Neo4j Community Edition (single-node) on Kubernetes with the following features:

- **Persistent Storage**: Automatic persistent volume provisioning for database files
- **Resource Management**: Configurable CPU and memory limits
- **External Access**: Optional LoadBalancer with external-DNS integration
- **Memory Tuning**: Configurable heap and page cache sizes
- **Secure by Default**: Auto-generated admin password stored in Kubernetes secret

## Prerequisites

- Kubernetes cluster (1.19+)
- Helm provider configured
- kubectl access to the cluster
- (Optional) external-dns for automatic DNS configuration

## Usage

### Basic Deployment

```hcl
module "neo4j_basic" {
  source = "./path/to/kubernetesneo4j/v1/iac/tf"

  metadata = {
    name = "my-graph-db"
  }

  spec = {
    container = {
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
        requests = {
          cpu    = "50m"
          memory = "100Mi"
        }
      }
      persistence_enabled = true
      disk_size          = "10Gi"
    }

    ingress = {
      enabled  = false
      hostname = ""
    }
  }
}
```

### Production Deployment with External Access

```hcl
module "neo4j_production" {
  source = "./path/to/kubernetesneo4j/v1/iac/tf"

  metadata = {
    name = "production-neo4j"
    org  = "my-organization"
    env  = "production"
  }

  spec = {
    container = {
      resources = {
        limits = {
          cpu    = "4000m"
          memory = "8Gi"
        }
        requests = {
          cpu    = "2000m"
          memory = "4Gi"
        }
      }
      persistence_enabled = true
      disk_size          = "100Gi"
    }

    memory_config = {
      heap_max   = "4Gi"
      page_cache = "2Gi"
    }

    ingress = {
      enabled  = true
      hostname = "neo4j.example.com"
    }
  }
}
```

## Variables

### metadata (required)

Resource metadata for identification and labeling.

| Field   | Type   | Required | Description                                    |
|---------|--------|----------|------------------------------------------------|
| name    | string | yes      | Name of the Neo4j instance                     |
| id      | string | no       | Custom resource ID (defaults to name)          |
| org     | string | no       | Organization label                             |
| env     | string | no       | Environment label (e.g., "production", "dev")  |
| labels  | map    | no       | Additional labels to apply                     |
| tags    | list   | no       | Tags for categorization                        |
| version | object | no       | Version information                            |

### spec (required)

Neo4j deployment specifications.

#### spec.container (required)

Container resource configuration.

| Field               | Type   | Required | Description                                    |
|---------------------|--------|----------|------------------------------------------------|
| resources.limits    | object | yes      | Maximum CPU and memory                         |
| resources.requests  | object | yes      | Requested CPU and memory                       |
| persistence_enabled | bool   | no       | Enable persistent storage (default: false)     |
| disk_size           | string | no       | Size of persistent volume (e.g., "10Gi")       |

#### spec.memory_config (optional)

Neo4j memory tuning parameters.

| Field      | Type   | Required | Description                                      |
|------------|--------|----------|--------------------------------------------------|
| heap_max   | string | no       | Maximum Java heap size (e.g., "1Gi", "512m")     |
| page_cache | string | no       | Page cache size for on-disk data (e.g., "512m") |

#### spec.ingress (required)

External access configuration.

| Field    | Type   | Required | Description                                           |
|----------|--------|----------|-------------------------------------------------------|
| enabled  | bool   | yes      | Enable external LoadBalancer service                  |
| hostname | string | no       | External hostname (required if enabled is true)       |

## Outputs

| Output                  | Description                                                  |
|-------------------------|--------------------------------------------------------------|
| namespace               | Kubernetes namespace where Neo4j is deployed                 |
| service                 | Service name for in-cluster connections                     |
| bolt_uri_kube_endpoint  | Bolt URI for database connections (internal)                |
| http_uri_kube_endpoint  | HTTP URL for Neo4j Browser (internal)                       |
| port_forward_command    | kubectl command to port-forward for local development       |
| username                | Default Neo4j username (always "neo4j")                      |
| password_secret_name    | Kubernetes secret name containing the password               |
| password_secret_key     | Key within the secret for the password                       |
| external_hostname       | External hostname if ingress is enabled                      |

## Connecting to Neo4j

### Retrieve the Password

The Neo4j admin password is auto-generated and stored in a Kubernetes secret:

```bash
kubectl get secret <password_secret_name> -n <namespace> \
  -o jsonpath='{.data.neo4j-password}' | base64 -d
```

Or using Terraform output:

```bash
kubectl get secret $(terraform output -raw password_secret_name) \
  -n $(terraform output -raw namespace) \
  -o jsonpath='{.data.neo4j-password}' | base64 -d
```

### From Within the Cluster

Use the Bolt URI output for database connections:

```bash
bolt://<service_fqdn>:7687
```

Username: `neo4j`  
Password: Retrieved from secret

### Using kubectl port-forward

For local development, use the port-forward command:

```bash
# Copy the command from Terraform output
terraform output -raw port_forward_command

# Then access Neo4j Browser at: http://localhost:7474
```

### External Access (if ingress enabled)

Access Neo4j Browser at:

```
http://<external_hostname>:7474
```

Connect via Bolt:

```
bolt://<external_hostname>:7687
```

## Memory Configuration Best Practices

Neo4j performance heavily depends on proper memory configuration:

- **Heap Memory**: Used for query execution and transactions
  - Recommended: 25-50% of available memory (up to 31GB)
  - Example: For 8GB pod, set `heap_max = "2Gi"`

- **Page Cache**: Used for caching graph data from disk
  - Recommended: 50-70% of available memory
  - Example: For 8GB pod, set `page_cache = "4Gi"`

- **OS Memory**: Reserve ~1-2GB for OS and Neo4j processes

### Example Memory Allocation for 8GB Pod:

```hcl
spec = {
  container = {
    resources = {
      limits = {
        memory = "8Gi"
      }
    }
  }
  memory_config = {
    heap_max   = "2Gi"   # 25% for heap
    page_cache = "4Gi"   # 50% for page cache
  }                      # ~2Gi reserved for OS
}
```

## Persistence

- Persistent storage is **highly recommended** for production deployments
- Data is stored in `/data` directory within the container
- Uses `defaultStorageClass` from your Kubernetes cluster
- Volume size is configurable via `disk_size`

**Note**: The disk size cannot be reduced after initial creation due to Kubernetes StatefulSet limitations.

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n <namespace>
kubectl logs -n <namespace> <pod-name>
```

### Verify Helm Release

```bash
helm list -n <namespace>
helm status <release-name> -n <namespace>
```

### Common Issues

1. **Pod fails to start**: Check resource limits and availability
2. **Connection refused**: Verify service is running and port-forward is active
3. **Out of memory**: Increase memory limits or tune heap/page cache settings
4. **Persistent volume issues**: Check StorageClass availability

## Version Compatibility

- **Neo4j Version**: Community Edition (chart version 2025.03.0)
- **Kubernetes**: 1.19+
- **Helm Chart**: Official Neo4j Helm chart
- **Terraform**: 1.0+

## Security Considerations

- Default password is auto-generated and stored in Kubernetes secret
- Change the default password after first login
- For production, consider using Neo4j Enterprise with authentication/authorization
- Enable network policies to restrict access
- Use TLS for external connections

## License

Neo4j Community Edition is licensed under GPLv3. For commercial use, consider Neo4j Enterprise Edition.

## Support

For issues specific to this Terraform module, refer to the Planton documentation.  
For Neo4j-specific questions, consult the [Neo4j Documentation](https://neo4j.com/docs/).


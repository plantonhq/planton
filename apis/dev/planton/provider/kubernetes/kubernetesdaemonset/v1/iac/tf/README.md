# KubernetesDaemonSet Terraform Module

This Terraform module deploys a Kubernetes DaemonSet with comprehensive configuration options for node-level operations.

## Features

- **DaemonSet Deployment**: Ensures pods run on all (or selected) nodes
- **Environment Variables**: Support for direct values and secrets
- **Secret Management**: Both direct string values and Kubernetes Secret references
- **Volume Mounts**: HostPath, ConfigMap, Secret, EmptyDir, and PVC support
- **Security Context**: Privileged mode, capabilities, run-as settings
- **RBAC**: ClusterRole and Role with bindings
- **Tolerations**: Schedule on tainted nodes (master, control-plane, etc.)
- **Node Selectors**: Target specific node types
- **Update Strategies**: RollingUpdate and OnDelete

## Usage

```hcl
module "daemonset" {
  source = "./path/to/module"

  metadata = {
    name = "fluentd"
  }

  spec = {
    namespace        = "logging"
    create_namespace = true

    container = {
      app = {
        image = {
          repo = "fluent/fluentd-kubernetes-daemonset"
          tag  = "v1.16-debian-elasticsearch8"
        }

        resources = {
          limits = {
            cpu    = "500m"
            memory = "512Mi"
          }
          requests = {
            cpu    = "100m"
            memory = "200Mi"
          }
        }

        env = {
          variables = {
            FLUENT_ELASTICSEARCH_HOST = "elasticsearch.logging.svc.cluster.local"
          }
          secrets = {
            API_KEY = {
              secret_ref = {
                name = "fluentd-secrets"
                key  = "api-key"
              }
            }
          }
        }

        volume_mounts = [
          {
            name       = "varlog"
            mount_path = "/var/log"
            read_only  = true
            host_path = {
              path = "/var/log"
              type = "Directory"
            }
          }
        ]
      }
    }

    tolerations = [
      {
        operator = "Exists"
      }
    ]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name | object | yes |
| spec | DaemonSet specification | object | yes |

See `variables.tf` for complete type definitions.

## Outputs

| Name | Description |
|------|-------------|
| namespace | The Kubernetes namespace |
| daemonset_name | The DaemonSet name |
| service_account_name | The ServiceAccount name (if created) |
| labels | Labels applied to resources |

## Secret Management

This module supports two ways to provide secrets:

### Direct String Values (Development)

```hcl
env = {
  secrets = {
    DEBUG_TOKEN = {
      value = "my-debug-token"
    }
  }
}
```

### Kubernetes Secret References (Production)

```hcl
env = {
  secrets = {
    API_KEY = {
      secret_ref = {
        name = "my-app-secrets"
        key  = "api-key"
      }
    }
  }
}
```

## Common Use Cases

1. **Log Collection**: Fluentd, Fluent Bit, Filebeat
2. **Node Monitoring**: Prometheus Node Exporter, Datadog Agent
3. **Network Plugins**: Calico, Cilium
4. **Storage Daemons**: Ceph, Longhorn
5. **Security Agents**: Falco, Sysdig

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| kubernetes | ~> 2.35 |


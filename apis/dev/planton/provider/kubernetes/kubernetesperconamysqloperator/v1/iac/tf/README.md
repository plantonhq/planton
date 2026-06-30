# KubernetesPerconaMysqlOperator Terraform Module

## Overview

This Terraform module deploys the Percona Operator for MySQL to a Kubernetes cluster. The operator enables the management of MySQL database clusters within Kubernetes using custom resources.

## Key Features

### Terraform-Based Deployment
- **HCL Configuration**: Uses standard Terraform HCL for infrastructure definition
- **State Management**: Leverages Terraform state for tracking operator deployment
- **Provider Integration**: Integrates with Kubernetes and Helm providers seamlessly

### Operator Management
- **Helm Chart Deployment**: Deploys using the official Percona Helm chart
- **CRD Installation**: Automatically installs MySQL Custom Resource Definitions
- **Resource Configuration**: Configurable CPU and memory resources for the operator pod
- **Namespace Isolation**: Creates dedicated namespace for the operator

### Infrastructure as Code
- **Declarative Configuration**: Define desired state using Terraform syntax
- **Version Control**: Track changes to operator configuration in Git
- **Reproducible Deployments**: Consistently deploy across environments

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster with appropriate access
- kubectl configured with cluster access
- Helm provider configured

## Module Structure

```
tf/
├── main.tf          # Main resources (namespace, Helm release)
├── provider.tf      # Provider configurations (Kubernetes, Helm)
├── variables.tf     # Input variables
└── README.md        # This file
```

## Resources Created

1. **Kubernetes Namespace**: Conditionally created based on `create_namespace` flag
2. **Helm Release**: Deploys the Percona Operator for MySQL
3. **CRDs**: PerconaServerMySQL and related custom resources (installed by Helm chart)

## Input Variables

### metadata
Metadata for the resource, including name and labels.

**Type**: `object`

**Fields**:
- `name` (string, required): Name of the operator deployment
- `id` (string, optional): Unique identifier
- `org` (string, optional): Organization name
- `env` (string, optional): Environment name
- `labels` (map(string), optional): Key-value labels
- `tags` (list(string), optional): List of tags
- `version` (object, optional): Version information

### spec
Specification for the operator deployment.

**Type**: `object`

**Fields**:
- `namespace` (string, required): Namespace to install the operator
- `create_namespace` (bool, required): Whether to create the namespace
  - `true`: Module creates the namespace before deploying the operator
  - `false`: Module expects the namespace to already exist
- `container` (object, required): Container specifications
  - `resources` (object, required): Resource allocations
    - `limits` (object): Maximum resources
      - `cpu` (string): CPU limit (e.g., "1000m")
      - `memory` (string): Memory limit (e.g., "1Gi")
    - `requests` (object): Guaranteed resources
      - `cpu` (string): CPU request (e.g., "100m")
      - `memory` (string): Memory request (e.g., "256Mi")

## Outputs

### namespace
The Kubernetes namespace where the operator is deployed.

**Type**: `string`

**Value**: The namespace name from the spec (regardless of whether it was created by this module or already existed)

## Usage

### Basic Example

```hcl
module "percona_mysql_operator" {
  source = "./tf"

  metadata = {
    name = "percona-mysql-operator-prod"
  }

  spec = {
    namespace        = "percona-mysql-operator"
    create_namespace = true
    
    container = {
      resources = {
        requests = {
          cpu    = "100m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
      }
    }
  }
}
```

### With Custom Resources

```hcl
module "percona_mysql_operator_large" {
  source = "./tf"

  metadata = {
    name = "percona-mysql-operator-large"
    labels = {
      environment = "production"
      team        = "data-platform"
    }
  }

  spec = {
    namespace        = "percona-mysql-operator-prod"
    create_namespace = true
    
    container = {
      resources = {
        requests = {
          cpu    = "200m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
  }
}
```

## Namespace Management

The module provides flexible namespace management through the `create_namespace` variable:

### Automatic Namespace Creation

Set `create_namespace = true` to have the module create the namespace automatically:

```hcl
spec = {
  namespace        = "percona-mysql-operator"
  create_namespace = true
  # ...
}
```

**Use cases:**
- Initial operator installation
- Development and testing environments
- Simplified deployment workflows

### Using Existing Namespace

Set `create_namespace = false` to use a pre-existing namespace:

```hcl
spec = {
  namespace        = "database-operators"
  create_namespace = false
  # ...
}
```

**Prerequisites:**
- Namespace must exist before running `terraform apply`
- Create with: `kubectl create namespace database-operators`

**Use cases:**
- Namespaces managed by platform teams
- Pre-configured namespace policies or quotas
- Shared namespaces across multiple components
- Environments with restricted namespace creation permissions

## Deployment

### Initialize Terraform

```bash
terraform init
```

### Plan Changes

```bash
terraform plan -var-file="operator.tfvars"
```

### Apply Configuration

```bash
terraform apply -var-file="operator.tfvars"
```

### Destroy Resources

```bash
terraform destroy -var-file="operator.tfvars"
```

## Verification

After deployment, verify the operator is running:

```bash
# Check operator pod
kubectl get pods -n percona-mysql-operator

# Check operator logs
kubectl logs -n percona-mysql-operator -l app.kubernetes.io/name=kubernetes-percona-mysql-operator

# Verify CRDs
kubectl get crds | grep percona
```

## Configuration

### Helm Chart Configuration

The module configures the Helm chart with the following defaults:

- **Chart Repository**: `https://percona.github.io/percona-helm-charts/`
- **Chart Name**: `ps-operator`
- **Chart Version**: `0.8.0`
- **Timeout**: 300 seconds
- **Atomic**: true (rollback on failure)
- **Cleanup on Fail**: true

### Resource Defaults

Default resource allocations:
- **CPU Request**: 100m
- **Memory Request**: 256Mi
- **CPU Limit**: 1000m
- **Memory Limit**: 1Gi

## Troubleshooting

### Helm Release Failed

Check Helm release status:
```bash
helm list -n percona-mysql-operator
helm status -n percona-mysql-operator ps-operator
```

### CRDs Not Installed

Verify CRD installation:
```bash
kubectl get crds | grep percona
```

The Percona Helm chart automatically installs CRDs.

### Resource Limits Too Low

If the operator is resource-constrained, increase the limits in your variables file:

```hcl
spec = {
  container = {
    resources = {
      limits = {
        cpu    = "2000m"
        memory = "2Gi"
      }
    }
  }
}
```

## Integration with Other Modules

After deploying the operator, you can deploy MySQL clusters using:
- Terraform MySQL module
- PerconaServerMySQL CRDs directly
- MySQLKubernetes Planton resource

## State Management

### Local Backend

For development, use local state:

```hcl
terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}
```

### Remote Backend (Recommended for Production)

For production, use remote state:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "percona-mysql-operator/terraform.tfstate"
    region = "us-west-2"
  }
}
```

## Contributing

Contributions are welcome! Please ensure:
- All variables are properly documented
- Examples are tested and working
- README is updated with new features

## License

This project is licensed under the [MIT License](LICENSE).


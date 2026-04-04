# Confluent Cloud Kafka Terraform Module

## Overview

The **Confluent Cloud Kafka Terraform Module** provides Infrastructure as Code (IaC) capabilities for deploying and managing Confluent Cloud Kafka clusters across AWS, GCP, and Azure. This module integrates with OpenMCF's unified API framework, enabling developers to provision production-ready Kafka infrastructure using declarative Terraform configurations.

This module translates the `ConfluentKafka` resource specification into Confluent Cloud infrastructure, supporting all cluster types (Basic, Standard, Enterprise, Dedicated) with capabilities ranging from simple development clusters to enterprise-grade deployments with private networking and provisioned capacity.

## Key Features

### Module Capabilities

- **Multi-Cloud Support**: Deploy Kafka clusters on AWS, GCP, or Azure with consistent configuration
- **All Cluster Types**: Supports Basic, Standard, Enterprise, and Dedicated cluster types
- **High Availability**: Configurable single-zone or multi-zone deployments with 99.99% SLA support
- **Private Networking**: Integration with AWS PrivateLink, Azure Private Link, and GCP Private Service Connect
- **Provisioned Capacity**: Dedicated clusters with configurable CKU (Confluent Kafka Units)
- **Elastic Scaling**: Standard and Enterprise clusters with automatic capacity management
- **Validated Configuration**: Built-in validation for cluster type combinations and network requirements

### Terraform Best Practices

- **Type Safety**: Strongly-typed variables with validation rules
- **Dynamic Blocks**: Efficient resource configuration based on cluster type
- **Lifecycle Management**: Preconditions to prevent invalid configurations
- **Sensitive Data**: Proper handling of API keys and secrets
- **Comprehensive Outputs**: All essential cluster information exported

## Prerequisites

- **Terraform**: Version 1.0 or higher
- **Confluent Cloud Account**: Active account with appropriate permissions
- **Confluent Cloud API Key**: Cloud API key with cluster creation permissions
- **Confluent Cloud Environment**: Pre-created environment ID
- **Network Resource** (Optional): For Enterprise/Dedicated clusters with private networking

## Installation

This module is part of the OpenMCF monorepo. To use it:

```bash
# Clone the repository
git clone https://github.com/plantonhq/openmcf.git
cd openmcf/apis/org/openmcf/provider/confluent/confluentkafka/v1/iac/tf
```

## Usage

### Basic Example

```hcl
module "dev_kafka" {
  source = "./path/to/module"

  metadata = {
    name = "dev-kafka-cluster"
    labels = {
      environment = "development"
      team        = "platform"
    }
  }

  spec = {
    cloud          = "AWS"
    region         = "us-east-2"
    availability   = "SINGLE_ZONE"
    environment_id = "env-dev-abc123"
    cluster_type   = "BASIC"
  }

  confluent_api_key    = var.confluent_api_key
  confluent_api_secret = var.confluent_api_secret
}
```

### Production Standard Cluster

```hcl
module "prod_kafka" {
  source = "./path/to/module"

  metadata = {
    name = "prod-orders-kafka"
    labels = {
      environment  = "production"
      team         = "platform"
      criticality  = "high"
    }
  }

  spec = {
    cloud          = "GCP"
    region         = "us-central1"
    availability   = "MULTI_ZONE"
    environment_id = "env-prod-xyz789"
    cluster_type   = "STANDARD"
    display_name   = "Production Orders Kafka Cluster"
  }

  confluent_api_key    = var.confluent_api_key
  confluent_api_secret = var.confluent_api_secret
}
```

### Enterprise Cluster with Private Networking

```hcl
module "enterprise_kafka" {
  source = "./path/to/module"

  metadata = {
    name = "enterprise-secure-kafka"
    labels = {
      environment    = "production"
      security-zone  = "restricted"
    }
  }

  spec = {
    cloud          = "AWS"
    region         = "us-west-2"
    availability   = "MULTI_ZONE"
    environment_id = "env-prod-enterprise"
    cluster_type   = "ENTERPRISE"
    display_name   = "Enterprise Secure Kafka Cluster"
    
    network_config = {
      network_id = "n-abc123"  # Pre-created network resource
    }
  }

  confluent_api_key    = var.confluent_api_key
  confluent_api_secret = var.confluent_api_secret
}
```

### Dedicated Cluster with Provisioned Capacity

```hcl
module "dedicated_kafka" {
  source = "./path/to/module"

  metadata = {
    name = "dedicated-high-throughput"
    labels = {
      environment  = "production"
      workload     = "high-throughput"
      criticality  = "tier-1"
    }
  }

  spec = {
    cloud          = "AZURE"
    region         = "eastus"
    availability   = "MULTI_ZONE"
    environment_id = "env-prod-critical"
    cluster_type   = "DEDICATED"
    display_name   = "High-Throughput Dedicated Kafka"
    
    dedicated_config = {
      cku = 4  # 4 Confluent Kafka Units
    }
    
    network_config = {
      network_id = "n-azure-private-123"
    }
  }

  confluent_api_key    = var.confluent_api_key
  confluent_api_secret = var.confluent_api_secret
}
```

## Module Inputs

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata including name and labels |
| `spec` | object | Kafka cluster specification (see spec fields below) |
| `confluent_api_key` | string | Confluent Cloud API Key (sensitive) |
| `confluent_api_secret` | string | Confluent Cloud API Secret (sensitive) |

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cloud` | string | Yes | Cloud provider: `AWS`, `AZURE`, or `GCP` |
| `region` | string | Yes | Cloud-specific region (e.g., `us-east-2`, `us-central1`, `eastus`) |
| `availability` | string | Yes | Availability: `SINGLE_ZONE`, `MULTI_ZONE`, `LOW`, or `HIGH` |
| `environment_id` | string | Yes | Confluent Cloud environment ID |
| `cluster_type` | string | No | Cluster type: `BASIC`, `STANDARD`, `ENTERPRISE`, `DEDICATED` (default: `STANDARD`) |
| `dedicated_config` | object | Conditional | Required when `cluster_type` is `DEDICATED` |
| `dedicated_config.cku` | number | Conditional | Number of Confluent Kafka Units (minimum: 1) |
| `network_config` | object | No | Network configuration for private networking |
| `network_config.network_id` | string | Conditional | ID of pre-created network resource |
| `display_name` | string | No | Display name in UI (defaults to `metadata.name`) |

## Module Outputs

| Output | Description |
|--------|-------------|
| `id` | Unique cluster ID assigned by Confluent Cloud |
| `bootstrap_endpoint` | Bootstrap endpoint for Kafka clients (e.g., `SASL_SSL://pkc-xxxxx.us-central1.gcp.confluent.cloud:9092`) |
| `crn` | Confluent Resource Name (CRN) of the cluster |
| `rest_endpoint` | REST API endpoint for cluster management (e.g., `https://pkc-xxxxx.us-central1.gcp.confluent.cloud:443`) |
| `api_version` | API version of the cluster |
| `kind` | Resource kind |

## Cluster Type Selection Guide

| Cluster Type | Use Case | Tenancy | Pricing Model | Private Networking | Availability |
|--------------|----------|---------|---------------|-------------------|--------------|
| **BASIC** | Development, Testing | Multi-tenant | Elastic (pay-as-you-go) | ❌ | Single-zone only |
| **STANDARD** | Production (general) | Multi-tenant | Elastic | ❌ | Single or Multi-zone |
| **ENTERPRISE** | Production (secure) | Multi-tenant | Elastic | ✅ | Multi-zone |
| **DEDICATED** | Production (critical) | Single-tenant | Provisioned (CKU) | ✅ | Single or Multi-zone |

### Recommendations

- **Development/Testing**: Use `BASIC` with `SINGLE_ZONE`
- **Production (public internet)**: Use `STANDARD` with `MULTI_ZONE`
- **Production (private networking)**: Use `ENTERPRISE` or `DEDICATED` with `MULTI_ZONE` and `network_config`
- **High-throughput workloads**: Use `DEDICATED` with appropriate CKU count

## Validation Rules

The module includes built-in validations:

1. **Dedicated Config Validation**: `dedicated_config` is required when `cluster_type` is `DEDICATED`
2. **Network Config Restriction**: `network_config` is only supported for `ENTERPRISE` and `DEDICATED` cluster types
3. **CKU Minimum**: When specified, CKU must be at least 1

## Security Best Practices

### Credential Management

**DO NOT** hardcode credentials in your Terraform files. Use one of these secure methods:

#### Environment Variables

```bash
export TF_VAR_confluent_api_key="your-api-key"
export TF_VAR_confluent_api_secret="your-api-secret"
terraform apply
```

#### Terraform Cloud/Enterprise

Store sensitive variables in Terraform Cloud workspace variables marked as sensitive.

#### Vault Integration

```hcl
data "vault_generic_secret" "confluent_creds" {
  path = "secret/confluent/prod"
}

module "kafka" {
  source = "./path/to/module"
  # ... other config ...
  confluent_api_key    = data.vault_generic_secret.confluent_creds.data["api_key"]
  confluent_api_secret = data.vault_generic_secret.confluent_creds.data["api_secret"]
}
```

### Private Networking

For production workloads, always use private networking:

1. Pre-create a Confluent Cloud network resource
2. Configure AWS PrivateLink, Azure Private Link, or GCP Private Service Connect
3. Reference the network ID in `spec.network_config.network_id`

## Troubleshooting

### Common Issues

**Error: "dedicated_config is required when cluster_type is DEDICATED"**

Solution: Add the `dedicated_config` block with CKU specification:

```hcl
spec = {
  # ... other fields ...
  cluster_type = "DEDICATED"
  dedicated_config = {
    cku = 2
  }
}
```

**Error: "network_config is only supported for ENTERPRISE and DEDICATED cluster types"**

Solution: Change `cluster_type` to `ENTERPRISE` or `DEDICATED`, or remove `network_config`.

**Error: "Environment not found"**

Solution: Ensure the `environment_id` exists in your Confluent Cloud account and the API key has access to it.

## Examples

For more comprehensive examples, see [examples.md](../../examples.md) which includes:

- Basic development cluster
- Standard production cluster
- Enterprise cluster with private networking
- Dedicated cluster with provisioned capacity
- Multi-region disaster recovery setup

## Terraform Examples

For Terraform-specific examples with complete variable definitions and output usage, see [examples.md](./examples.md).

## Contributing

Contributions are welcome! Please see the [project contribution guidelines](../../../../../../../../../CONTRIBUTING.md).

## Support

For support:
- **Issues**: Open an issue on [GitHub](https://github.com/plantonhq/openmcf/issues)
- **Email**: [support@planton.ai](mailto:support@planton.ai)
- **Documentation**: [docs.confluent.io](https://docs.confluent.io/cloud/)

## References

- [Confluent Terraform Provider Documentation](https://registry.terraform.io/providers/confluentinc/confluent/latest/docs)
- [Confluent Cloud Documentation](https://docs.confluent.io/cloud/)
- [OpenMCF Documentation](https://github.com/plantonhq/openmcf)
- [Terraform Best Practices](https://www.terraform.io/docs/cloud/guides/recommended-practices/index.html)

## License

This module is part of OpenMCF and is licensed under the MIT License. See [LICENSE](../../../../../../../../../LICENSE) for details.


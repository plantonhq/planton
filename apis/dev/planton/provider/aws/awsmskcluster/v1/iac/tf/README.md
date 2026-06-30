# AwsMskCluster — Terraform IaC Module

Terraform module for provisioning AWS MSK (Managed Streaming for Apache Kafka) clusters using the Planton `AwsMskClusterSpec`.

## Overview

This module creates:
- An MSK Cluster (`aws_msk_cluster`) with configurable brokers, encryption, authentication, logging, and monitoring.
- A managed Security Group (`aws_security_group`) with ingress rules on Kafka ports (9092-9098) and ZooKeeper ports (2181-2182) — conditional on `security_group_ids` or `allowed_cidr_blocks` being provided.
- An inline MSK Configuration (`aws_msk_configuration`) from `server_properties` — conditional on the map being non-empty.

## Usage

```hcl
module "msk" {
  source = "./path/to/this/module"

  provider_config = {
    region = "us-east-1"
  }

  metadata = {
    id   = "prod-events"
    name = "prod-events"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    kafka_version          = "3.6.0"
    number_of_broker_nodes = 3
    instance_type          = "kafka.m5.large"
    subnet_ids             = ["subnet-aaa", "subnet-bbb", "subnet-ccc"]

    authentication = {
      sasl_iam_enabled = true
    }

    server_properties = {
      "auto.create.topics.enable"  = "false"
      "default.replication.factor" = "3"
      "min.insync.replicas"        = "2"
    }
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `provider_config` | object | yes | AWS region and optional credentials |
| `metadata` | object | yes | Resource ID, name, org, env |
| `spec` | object | yes | `AwsMskClusterSpec` — see `variables.tf` for full type |

See `variables.tf` for the complete type definition of `spec`, including all optional fields and their defaults.

## Outputs

| Output | Description |
|--------|-------------|
| `cluster_arn` | ARN of the MSK cluster |
| `cluster_name` | Cluster name |
| `cluster_uuid` | UUID extracted from ARN |
| `current_version` | Cluster version (for updates) |
| `bootstrap_brokers` | Plaintext broker endpoints (port 9092) |
| `bootstrap_brokers_tls` | TLS broker endpoints (port 9094) |
| `bootstrap_brokers_sasl_iam` | SASL/IAM broker endpoints (port 9098) |
| `bootstrap_brokers_sasl_scram` | SASL/SCRAM broker endpoints (port 9096) |
| `bootstrap_brokers_public_tls` | Public TLS broker endpoints |
| `bootstrap_brokers_public_sasl_iam` | Public SASL/IAM broker endpoints |
| `bootstrap_brokers_public_sasl_scram` | Public SASL/SCRAM broker endpoints |
| `zookeeper_connect_string` | ZooKeeper plaintext endpoints |
| `zookeeper_connect_string_tls` | ZooKeeper TLS endpoints |
| `security_group_id` | Managed SG ID (empty if not created) |
| `configuration_arn` | Inline config ARN (empty if not created) |

## File Structure

| File | Purpose |
|------|---------|
| `provider.tf` | AWS provider configuration (hashicorp/aws ~> 5.0) |
| `variables.tf` | Input variable definitions with full type constraints |
| `locals.tf` | Tags, ingress condition, server_properties serialization |
| `main.tf` | All resource definitions (SG, configuration, cluster) |
| `outputs.tf` | 15 output definitions |

## Prerequisites

- Terraform 1.5+
- AWS provider ~> 5.0
- AWS credentials (via provider config or ambient)

## Related

- [Spec reference](../../README.md)
- [Examples](../../examples.md)

# AwsMemcachedElasticache Terraform Module

This directory contains the Terraform IaC module for provisioning AWS ElastiCache Memcached clusters.

## Structure

```
.
├── main.tf        # Primary resources: subnet group, parameter group, cluster
├── locals.tf      # Computed values and StringValueOrRef resolution
├── outputs.tf     # Output values matching stack_outputs.proto
├── variables.tf   # Input variables (metadata + spec)
└── provider.tf    # AWS provider configuration
```

## Local Validation

```bash
terraform init && terraform validate
```

## Resources Created

- `aws_elasticache_subnet_group` (conditional) — when `subnetIds` provided
- `aws_elasticache_parameter_group` (conditional) — when `parameters` + `parameterGroupFamily` provided
- `aws_elasticache_cluster` (always) — Memcached cluster with engine="memcached"

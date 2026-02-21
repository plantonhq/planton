# AlicloudPolardbCluster -- Terraform Module

## Overview

This Terraform module deploys an Alibaba Cloud PolarDB cluster with bundled databases, accounts, and account privileges.

## Module Structure

| File | Purpose |
|------|---------|
| `main.tf` | PolarDB cluster resource |
| `databases.tf` | Database resources (for_each) |
| `accounts.tf` | Account and privilege resources (for_each) |
| `variables.tf` | Input variables from proto spec |
| `outputs.tf` | Output values matching stack_outputs.proto |
| `locals.tf` | Tag computation, defaults, map flattening |
| `provider.tf` | Alibaba Cloud provider configuration |

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Provider

Uses `aliyun/alicloud` provider version `~> 1.200`.

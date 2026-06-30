# AwsSagemakerDomain — Terraform IaC Module

Terraform module for provisioning Amazon SageMaker Domains using the Planton `AwsSagemakerDomainSpec`.

## Overview

This module creates:
- A SageMaker Domain (`aws_sagemaker_domain`) with configurable authentication, VPC networking, user settings, Docker access, and encryption.
- Default user settings including JupyterLab configuration, KernelGateway configuration, idle timeout, sharing settings, and space storage.

## Usage

```hcl
module "sagemaker_domain" {
  source = "./path/to/this/module"

  provider_config = {
    region = "us-east-1"
  }

  metadata = {
    id   = "ml-workspace"
    name = "ml-workspace"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    auth_mode = "IAM"
    vpc_id    = "vpc-0abc123def456789"
    subnet_ids = ["subnet-aaa", "subnet-bbb"]

    default_user_settings = {
      execution_role_arn = "arn:aws:iam::111122223333:role/SageMakerExecutionRole"

      jupyter_lab_app_settings = {
        default_resource_spec = {
          instance_type = "ml.t3.medium"
        }
        idle_settings = {
          lifecycle_management     = "ENABLED"
          idle_timeout_in_minutes  = 120
        }
      }
    }
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `provider_config` | object | yes | AWS region and optional credentials |
| `metadata` | object | yes | Resource ID, name, org, env |
| `spec` | object | yes | `AwsSagemakerDomainSpec` — see `variables.tf` for full type |

See `variables.tf` for the complete type definition of `spec`, including all optional fields and their defaults.

## Outputs

| Output | Description |
|--------|-------------|
| `domain_id` | Unique identifier of the SageMaker Domain |
| `domain_arn` | ARN of the SageMaker Domain |
| `domain_url` | HTTPS URL for SageMaker Studio web interface |
| `home_efs_file_system_id` | ID of the auto-created EFS file system |
| `security_group_id_for_domain_boundary` | AWS-managed security group for domain boundary |
| `single_sign_on_application_arn` | IAM Identity Center application ARN (SSO mode only) |

## File Structure

| File | Purpose |
|------|---------|
| `provider.tf` | AWS provider configuration (hashicorp/aws ~> 5.0) |
| `variables.tf` | Input variable definitions with full type constraints |
| `locals.tf` | Tags, computed values |
| `main.tf` | SageMaker Domain resource definition |
| `outputs.tf` | 6 output definitions |

## Prerequisites

- Terraform 1.5+
- AWS provider ~> 5.0
- AWS credentials (via provider config or ambient)

## Validate

```bash
cd iac/tf
terraform init
terraform validate
```

## Related

- [Spec reference](../../README.md)
- [Examples](../../examples.md)

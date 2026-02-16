# Terraform Module — AwsServerlessElasticache

This directory contains the Terraform IaC module for provisioning AWS ElastiCache
Serverless caches.

## Structure

- `main.tf` — ElastiCache Serverless cache resource with dynamic scaling limits
- `locals.tf` — Variable transformations, tag construction, limit computation
- `outputs.tf` — Stack outputs matching `AwsServerlessElasticacheStackOutputs`
- `variables.tf` — Input variables from stack input
- `provider.tf` — AWS provider configuration

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply
```

## Validate

```bash
terraform init
terraform validate
```

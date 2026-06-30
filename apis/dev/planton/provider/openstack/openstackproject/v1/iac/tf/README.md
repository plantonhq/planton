# OpenStackProject Terraform Module

This Terraform module provisions an OpenStack Identity (Keystone) project.

## Resources Created

- `openstack_identity_project_v3` -- Keystone project

## Usage

This module is invoked by the Planton CLI. For local development:

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, labels, etc.) |
| `spec` | OpenStackProjectSpec configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `project_id` | UUID of the created project |
| `name` | Project name |
| `domain_id` | Keystone domain UUID |
| `enabled` | Whether the project is active |
| `region` | OpenStack region |

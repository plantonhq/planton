# OpenStackServerGroup Terraform Module

Provisions an OpenStack Compute server group using the Terraform OpenStack provider.

## Resources Created

- `openstack_compute_servergroup_v2` -- A server group with the specified placement policy

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata.name` | Server group name |
| `spec.policy` | Placement policy (affinity, anti-affinity, soft-affinity, soft-anti-affinity) |
| `spec.region` | Optional region override |

## Outputs

| Output | Description |
|--------|-------------|
| `server_group_id` | UUID of the server group |
| `name` | Name of the server group |
| `members` | List of member instance UUIDs |
| `region` | Region where created |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

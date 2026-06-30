# OpenStackLoadBalancerMember Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Octavia pool members.

## Structure

```
iac/tf/
|-- variables.tf   # Input variables (metadata + spec)
|-- locals.tf      # Derived values and FK resolution
|-- main.tf        # Member resource definition
|-- outputs.tf     # Stack outputs
|-- provider.tf    # OpenStack provider configuration
+-- README.md      # This file
```

## Usage

This module is invoked by the Planton CLI, not directly. The CLI:

1. Translates the YAML manifest into Terraform variable values
2. Sets OpenStack credentials via `OS_*` environment variables
3. Runs `terraform plan` / `terraform apply`

## Variables

### `metadata`

Standard Planton metadata object with `name`, `id`, `org`, `env`, `labels`.

### `spec`

OpenStackLoadBalancerMemberSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `pool_id` | `object({value=string})` | Yes | - | Pool ID (StringValueOrRef) |
| `address` | `string` | Yes | - | Backend server IP address |
| `protocol_port` | `number` | Yes | - | Backend server port (1-65535) |
| `subnet_id` | `object({value=string})` | No | `null` | Subnet ID (StringValueOrRef) |
| `weight` | `number` | No | `null` | Member weight (0-256) |
| `admin_state_up` | `bool` | No | `true` | Administrative state |
| `tags` | `list(string)` | No | `[]` | Tags |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `member_id` | UUID of the member |
| `name` | Member name |
| `address` | Backend IP address |
| `protocol_port` | Backend port |
| `weight` | Member weight |
| `region` | OpenStack region |

## Foreign Key Resolution

The `pool_id` and `subnet_id` variables use the `StringValueOrRef` pattern. In Terraform:
- The CLI resolves `value_from` references before invoking Terraform
- The resolved value is passed as `spec.pool_id.value` / `spec.subnet_id.value`
- `locals.tf` extracts them

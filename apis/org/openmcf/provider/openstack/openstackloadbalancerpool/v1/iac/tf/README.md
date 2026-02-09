# OpenStackLoadBalancerPool Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Octavia backend pools.

## Structure

```
iac/tf/
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Pool resource definition
├── outputs.tf     # Stack outputs
├── provider.tf    # OpenStack provider configuration
└── README.md      # This file
```

## Usage

This module is invoked by the OpenMCF CLI, not directly. The CLI:

1. Translates the YAML manifest into Terraform variable values
2. Sets OpenStack credentials via `OS_*` environment variables
3. Runs `terraform plan` / `terraform apply`

## Variables

### `metadata`

Standard OpenMCF metadata object with `name`, `id`, `org`, `env`, `labels`.

### `spec`

OpenStackLoadBalancerPoolSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `listener_id` | `object({value=string})` | Yes | - | Listener ID (StringValueOrRef) |
| `protocol` | `string` | Yes | - | Backend protocol (HTTP, HTTPS, TCP, UDP, PROXY) |
| `lb_method` | `string` | Yes | - | Load-balancing algorithm |
| `persistence` | `object` | No | `null` | Session persistence config |
| `description` | `string` | No | `""` | Description |
| `admin_state_up` | `bool` | No | `true` | Administrative state |
| `tags` | `list(string)` | No | `[]` | Tags |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `pool_id` | UUID of the pool |
| `name` | Pool name |
| `protocol` | Backend protocol |
| `lb_method` | Load-balancing algorithm |
| `region` | OpenStack region |

## Foreign Key Resolution

The `listener_id` variable uses the `StringValueOrRef` pattern. In Terraform:
- The CLI resolves `value_from` references before invoking Terraform
- The resolved value is passed as `spec.listener_id.value`
- `locals.tf` extracts it: `local.listener_id = var.spec.listener_id.value`

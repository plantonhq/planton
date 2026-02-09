# OpenStackLoadBalancer Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Octavia load balancers.

## Structure

```
iac/tf/
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Load balancer resource definition
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

OpenStackLoadBalancerSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `vip_subnet_id` | `object({value=string})` | Yes | - | VIP subnet ID (StringValueOrRef FK to OpenStackSubnet) |
| `vip_address` | `string` | No | `""` | Specific VIP address |
| `description` | `string` | No | `""` | Description |
| `admin_state_up` | `bool` | No | `true` | Administrative state |
| `flavor_id` | `string` | No | `""` | Octavia flavor ID |
| `tags` | `list(string)` | No | `[]` | Tags |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `loadbalancer_id` | UUID of the load balancer |
| `name` | Load balancer name |
| `vip_address` | Virtual IP address |
| `vip_port_id` | Neutron port ID of the VIP |
| `vip_subnet_id` | Subnet where the VIP was allocated |
| `region` | OpenStack region |

## Foreign Key Resolution

The `vip_subnet_id` variable uses the `StringValueOrRef` pattern. In Terraform:
- The CLI resolves `value_from` references before invoking Terraform
- The resolved value is passed as `spec.vip_subnet_id.value`
- `locals.tf` extracts it: `local.vip_subnet_id = var.spec.vip_subnet_id.value`

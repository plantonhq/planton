# OpenStackSubnet Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Neutron subnets.

## Structure

```
iac/tf/
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Subnet resource definition
├── outputs.tf     # Stack outputs
├── provider.tf    # OpenStack provider configuration
└── README.md      # This file
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

OpenStackSubnetSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `network_id` | `object({value=string})` | Yes | - | Parent network ID (StringValueOrRef) |
| `cidr` | `string` | Yes | - | CIDR notation |
| `ip_version` | `number` | No | `4` | IPv4 or IPv6 |
| `gateway_ip` | `string` | No | `""` | Gateway IP |
| `no_gateway` | `bool` | No | `false` | Disable gateway |
| `enable_dhcp` | `bool` | No | `true` | Enable DHCP |
| `dns_nameservers` | `list(string)` | No | `[]` | DNS servers |
| `allocation_pools` | `list(object)` | No | `[]` | IP pools |
| `description` | `string` | No | `""` | Description |
| `tags` | `list(string)` | No | `[]` | Tags |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `subnet_id` | UUID of the subnet |
| `name` | Subnet name |
| `cidr` | CIDR block |
| `gateway_ip` | Gateway IP address |
| `network_id` | Parent network ID |
| `region` | OpenStack region |

## Foreign Key Resolution

The `network_id` variable uses the `StringValueOrRef` pattern. In Terraform:
- The CLI resolves `value_from` references before invoking Terraform
- The resolved value is passed as `spec.network_id.value`
- `locals.tf` extracts it: `local.network_id = var.spec.network_id.value`

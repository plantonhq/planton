# OpenStackLoadBalancerListener Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Octavia listeners.

## Structure

```
iac/tf/
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Listener resource definition
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

OpenStackLoadBalancerListenerSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `loadbalancer_id` | `object({value=string})` | Yes | - | Load balancer ID (StringValueOrRef FK to OpenStackLoadBalancer) |
| `protocol` | `string` | Yes | - | Protocol (HTTP, HTTPS, TCP, UDP, TERMINATED_HTTPS) |
| `protocol_port` | `number` | Yes | - | Port number (1-65535) |
| `description` | `string` | No | `""` | Description |
| `connection_limit` | `number` | No | `null` | Max connections (-1 = unlimited) |
| `default_tls_container_ref` | `string` | No | `""` | Barbican TLS secret URI |
| `insert_headers` | `map(string)` | No | `{}` | HTTP headers to insert |
| `allowed_cidrs` | `list(string)` | No | `[]` | CIDRs allowed to access the listener |
| `admin_state_up` | `bool` | No | `true` | Administrative state |
| `tags` | `list(string)` | No | `[]` | Tags |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `listener_id` | UUID of the listener |
| `name` | Listener name |
| `protocol` | Protocol the listener accepts |
| `protocol_port` | Port the listener accepts traffic on |
| `region` | OpenStack region |

## Foreign Key Resolution

The `loadbalancer_id` variable uses the `StringValueOrRef` pattern. In Terraform:
- The CLI resolves `value_from` references before invoking Terraform
- The resolved value is passed as `spec.loadbalancer_id.value`
- `locals.tf` extracts it: `local.loadbalancer_id = var.spec.loadbalancer_id.value`

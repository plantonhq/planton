# OpenStackLoadBalancerMonitor Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Octavia health monitors.

## Structure

```
iac/tf/
|-- variables.tf   # Input variables (metadata + spec)
|-- locals.tf      # Derived values and FK resolution
|-- main.tf        # Monitor resource definition
|-- outputs.tf     # Stack outputs
|-- provider.tf    # OpenStack provider configuration
+-- README.md      # This file
```

## Usage

This module is invoked by the Planton CLI, not directly.

## Important Note

**Health monitors do NOT support tags** in the Terraform OpenStack provider.
This is a provider limitation, not an Planton design choice.

## Variables

### `metadata`

Standard Planton metadata object with `name`, `id`, `org`, `env`, `labels`.

### `spec`

OpenStackLoadBalancerMonitorSpec fields:

| Variable | Type | Required | Default | Description |
|---|---|---|---|---|
| `pool_id` | `object({value=string})` | Yes | - | Pool ID (StringValueOrRef) |
| `type` | `string` | Yes | - | Health check type |
| `delay` | `number` | Yes | - | Interval between checks (seconds) |
| `timeout` | `number` | Yes | - | Timeout per check (seconds) |
| `max_retries` | `number` | Yes | - | Consecutive successes for healthy (1-10) |
| `max_retries_down` | `number` | No | `null` | Consecutive failures for unhealthy (1-10) |
| `url_path` | `string` | No | `""` | URL path (HTTP/HTTPS only) |
| `http_method` | `string` | No | `""` | HTTP method (HTTP/HTTPS only) |
| `expected_codes` | `string` | No | `""` | Expected HTTP codes (HTTP/HTTPS only) |
| `admin_state_up` | `bool` | No | `true` | Administrative state |
| `region` | `string` | No | `""` | Region override |

## Outputs

| Output | Description |
|---|---|
| `monitor_id` | UUID of the health monitor |
| `name` | Monitor name |
| `type` | Health check type |
| `pool_id` | Monitored pool ID |
| `region` | OpenStack region |

## Foreign Key Resolution

The `pool_id` variable uses the `StringValueOrRef` pattern.

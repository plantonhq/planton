# OpenStackRouter Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Neutron routers.

## Structure

```
iac/tf/
├── provider.tf    # OpenStack provider configuration
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Router resource definition
├── outputs.tf     # Stack outputs
└── README.md      # This file
```

## Usage

The module is invoked by the Planton Terraform runner, which passes the `metadata` and `spec` variables from the resource manifest.

## Foreign Key Resolution

The `external_network_id` field uses the `StringValueOrRef` pattern:
- At the Terraform level, it's always a literal `{value: "uuid"}` object (FK resolution happens before TF runs)
- When absent (`null`), no external gateway is configured
- `locals.tf` extracts the value with a null-safe expression

## Outputs

| Output | Description |
|--------|-------------|
| `router_id` | UUID of the created router |
| `name` | Router name |
| `external_network_id` | External network UUID (empty if internal-only) |
| `external_gateway_ip` | Primary external IP address (empty if internal-only) |
| `region` | OpenStack region |

The `external_gateway_ip` output extracts the first IP from the computed `external_fixed_ip` list, handling the empty case (no external gateway) with a conditional expression.

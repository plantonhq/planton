# OpenStackRouterInterface Terraform Module

This directory contains the Terraform HCL module for attaching OpenStack Neutron routers to subnets.

## Structure

```
iac/tf/
├── provider.tf    # OpenStack provider configuration
├── variables.tf   # Input variables (metadata + spec)
├── locals.tf      # Derived values and FK resolution
├── main.tf        # Router interface resource definition
├── outputs.tf     # Stack outputs
└── README.md      # This file
```

## Usage

The module is invoked by the Planton Terraform runner, which passes the `metadata` and `spec` variables from the resource manifest.

## Foreign Key Resolution

Both `router_id` and `subnet_id` use the `StringValueOrRef` pattern:
- At the Terraform level, they are always literal `{value: "uuid"}` objects (FK resolution happens before TF runs)
- `locals.tf` extracts the values with simple `.value` access

## Outputs

| Output | Description |
|--------|-------------|
| `port_id` | UUID of the auto-created port (also the TF resource ID) |
| `router_id` | UUID of the router |
| `subnet_id` | UUID of the subnet |
| `region` | OpenStack region |

The `port_id` output comes from `openstack_networking_router_interface_v2.main.id` because the Terraform provider sets the resource ID to the port UUID that OpenStack creates when attaching the interface.

# OpenStackNetwork Terraform Module

Terraform (HCL) IaC module for provisioning OpenStack Neutron networks.

## Structure

```
.
├── provider.tf     # OpenStack provider configuration
├── variables.tf    # Input variables (mirrors spec.proto)
├── locals.tf       # Computed local values
├── main.tf         # Network resource definition
└── outputs.tf      # Output values (mirrors stack_outputs.proto)
```

## Provider Configuration

The OpenStack provider is configured via `OS_*` environment variables, which are set by the Planton providerenvvars layer from the `OpenStackProviderConfig` proto.

## Resources Created

- `openstack_networking_network_v2.main` — The Neutron network

## Outputs

| Name | Description |
|------|-------------|
| `network_id` | The unique identifier (UUID) of the network |
| `name` | The network name |
| `region` | Region where the network was created |

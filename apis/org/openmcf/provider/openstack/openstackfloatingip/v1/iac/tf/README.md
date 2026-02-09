# OpenStackFloatingIp Terraform Module

This directory contains the Terraform HCL module for provisioning OpenStack Neutron floating IPs.

## Structure

```
iac/tf/
├── variables.tf    # Input variables (metadata + spec)
├── locals.tf       # FK extraction and label computation
├── main.tf         # Floating IP resource
├── outputs.tf      # Stack outputs
├── provider.tf     # OpenStack provider config
└── README.md       # This file
```

## Resource Mapping

| Terraform Resource | Count | Description |
|---|---|---|
| `openstack_networking_floatingip_v2` | 1 | Floating IP allocation + optional association |

## Key Design Notes

- **`pool` mapping**: The TF field `pool` accepts a network name or UUID. Our `floating_network_id` FK resolves to a UUID, which is passed to `pool`.
- **Single resource**: No separate `floatingip_associate_v2`. The `port_id` field on the floating IP resource handles built-in association natively.
- **Optional FK**: `port_id` uses conditional extraction in `locals.tf` -- null when not provided.
- **ForceNew fields**: `address`, `pool` (floating_network_id), and `region` are ForceNew. Changing them recreates the resource.

## Usage

This module is invoked by the OpenMCF CLI's Terraform runner. It is not intended for standalone use.

```bash
# Variables are passed as a JSON file by the runner
terraform apply -var-file=terraform.tfvars.json
```

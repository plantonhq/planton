# HetznerCloudPlacementGroup Terraform Module

Terraform (HCL) IaC module for creating placement groups in Hetzner Cloud.

## Structure

```
.
├── provider.tf     # HetznerCloud provider configuration (hcloud ~> 1.60)
├── variables.tf    # Input variables (metadata + spec + hcloud_token)
├── locals.tf       # Computed values: placement_group_name, placement_group_type, standard_labels
├── main.tf         # Placement group resource definition
└── outputs.tf      # Output values (mirrors stack_outputs.proto)
```

## Provider Configuration

The Hetzner Cloud provider is configured via the `hcloud_token` variable, which is set by the OpenMCF `providerenvvars` layer from the `HetznerCloudProviderConfig` proto. Requires Terraform >= 1.5.

## Resources Created

- `hcloud_placement_group.this` — Placement group with the specified strategy (defaults to `spread`)

## Outputs

| Name | Description |
|------|-------------|
| `placement_group_id` | Hetzner Cloud numeric ID of the created placement group |

## Usage

```bash
# Initialize
openmcf tofu init --manifest manifest.yaml

# Plan
openmcf tofu plan --manifest manifest.yaml

# Apply
openmcf tofu apply --manifest manifest.yaml --auto-approve
```

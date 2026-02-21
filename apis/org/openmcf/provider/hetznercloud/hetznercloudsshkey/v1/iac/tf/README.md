# HetznerCloudSshKey Terraform Module

Terraform (HCL) IaC module for registering SSH public keys in Hetzner Cloud.

## Structure

```
.
├── provider.tf     # HetznerCloud provider configuration (hcloud ~> 1.60)
├── variables.tf    # Input variables (metadata + spec + hcloud_token)
├── locals.tf       # Computed values: ssh_key_name, public_key, standard_labels
├── main.tf         # SSH key resource definition
└── outputs.tf      # Output values (mirrors stack_outputs.proto)
```

## Provider Configuration

The Hetzner Cloud provider is configured via the `hcloud_token` variable, which is set by the OpenMCF `providerenvvars` layer from the `HetznerCloudProviderConfig` proto. Requires Terraform >= 1.5.

## Resources Created

- `hcloud_ssh_key.this` — SSH public key registered in Hetzner Cloud

## Outputs

| Name | Description |
|------|-------------|
| `ssh_key_id` | Hetzner Cloud numeric ID of the created SSH key |
| `fingerprint` | MD5 fingerprint of the SSH public key |

## Usage

```bash
# Initialize
openmcf tofu init --manifest manifest.yaml

# Plan
openmcf tofu plan --manifest manifest.yaml

# Apply
openmcf tofu apply --manifest manifest.yaml --auto-approve
```

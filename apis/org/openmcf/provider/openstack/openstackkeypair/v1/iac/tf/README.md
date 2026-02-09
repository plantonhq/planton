# OpenStackKeypair Terraform Module

Terraform (HCL) IaC module for provisioning OpenStack compute keypairs.

## Structure

```
.
├── provider.tf     # OpenStack provider configuration
├── variables.tf    # Input variables (mirrors spec.proto)
├── locals.tf       # Computed local values
├── main.tf         # Keypair resource definition
└── outputs.tf      # Output values (mirrors stack_outputs.proto)
```

## Provider Configuration

The OpenStack provider is configured via `OS_*` environment variables, which are set by the OpenMCF providerenvvars layer from the `OpenStackProviderConfig` proto.

## Resources Created

- `openstack_compute_keypair_v2.main` — The SSH keypair

## Outputs

| Name | Description |
|------|-------------|
| `name` | The keypair name |
| `fingerprint` | MD5 fingerprint of the public key |
| `public_key` | The SSH public key |
| `region` | Region where the keypair was created |
| `private_key` | Generated private key (sensitive, only when no public_key provided) |

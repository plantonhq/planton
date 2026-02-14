---
title: "Compute Keypair — Research Documentation"
description: "Compute Keypair — Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackkeypair"
---

# OpenStack Compute Keypair — Research Documentation

## OpenStack Compute Keypair Overview

OpenStack Nova (Compute) provides a keypair management service that stores SSH public keys and associates them with compute instances at launch time. When an instance boots, cloud-init (or a metadata service equivalent) injects the public key into the default user's `~/.ssh/authorized_keys`, enabling passwordless SSH access.

### Key Concepts

1. **Keypair Name**: A unique identifier within a user's scope. Used to reference the key when launching instances.
2. **Public Key**: The SSH public key in OpenSSH `authorized_keys` format.
3. **Fingerprint**: An MD5 hash of the public key, used for quick identification.
4. **Private Key**: Only available when OpenStack generates the keypair. Returned exactly once at creation time.

### API Details

- **Nova API Version**: v2 (v2.1 with microversions)
- **Terraform Resource**: `openstack_compute_keypair_v2` (terraform-provider-openstack)
- **Pulumi Resource**: `openstack.compute.Keypair` (pulumi-openstack)

### Authentication

OpenStack supports multiple authentication methods via Keystone (Identity service):

| Method | Use Case | Required Fields |
|--------|----------|-----------------|
| Password | Interactive / development | `auth_url`, `user_name`, `password`, domain/project context |
| Application Credential | CI/CD / automation (recommended) | `auth_url`, credential `id` or `name`, `secret` |
| Token | Short-lived / delegated access | `auth_url`, `token` |

### Resource Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Unique keypair name. Changing it forces recreation. |
| `public_key` | No | OpenSSH public key to import. If omitted, a keypair is generated. |
| `region` | No | Override the provider's region. |
| `user_id` | No | Admin-only: manage keypairs for other users (microversion 2.10+). |
| `value_specs` | No | Vendor-specific request body parameters. |

### Resource Attributes (Outputs)

| Attribute | Description |
|-----------|-------------|
| `name` | The keypair name |
| `public_key` | The SSH public key |
| `fingerprint` | MD5 fingerprint of the public key |
| `private_key` | Generated private key (only when no `public_key` is provided) |
| `region` | The region where the keypair exists |

### Design Decisions

#### Fields Included (80/20 Principle)

- **`public_key`**: Core functionality — import vs. generate decision
- **`region`**: Common override for multi-region deployments

#### Fields Excluded

- **`user_id`**: Admin-only feature requiring Nova microversion 2.10+. Very niche use case.
- **`value_specs`**: Vendor-specific escape hatch. Adds complexity without clear benefit for the vast majority of users.

#### Keypair Name from metadata.name

The keypair name is derived from `metadata.name` rather than having a separate `spec.name` field. This simplifies the API — the OpenMCF resource name IS the keypair name in OpenStack. This mirrors how most users think about keypairs: a named key that they reference by name when launching instances.

#### Private Key Not in Stack Outputs

The `private_key` is intentionally excluded from `stack_outputs.proto` following the platform principle that secrets belong in external secret managers, not in platform-tracked outputs. The private key IS available through the IaC engine's native secret handling:

- **Pulumi**: Exported as a secret output, retrievable via `pulumi stack output private_key --show-secrets`
- **Terraform**: Marked as `sensitive = true`, retrievable via `terraform output -raw private_key`

### Security Considerations

1. **Key Generation**: Generated private keys are stored unencrypted in IaC state. For production, always import externally generated keys.
2. **Key Rotation**: Keypairs cannot be updated in place — changing the public key forces recreation. Plan for key rotation workflows.
3. **Key Algorithm**: OpenStack accepts RSA, ECDSA, and ED25519 keys. ED25519 is recommended for its security and performance characteristics.

### Related OpenStack Services

- **Nova (Compute)**: Uses keypairs for instance SSH access
- **Keystone (Identity)**: Provides authentication for the API calls
- **Glance (Image)**: Images must support cloud-init for key injection

### References

- [Terraform openstack_compute_keypair_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/compute_keypair_v2)
- [Pulumi openstack.compute.Keypair](https://www.pulumi.com/registry/packages/openstack/api-docs/compute/keypair/)
- [OpenStack Nova Keypairs API](https://docs.openstack.org/api-ref/compute/#keypairs-os-keypairs)
- [OpenStack Application Credentials](https://docs.openstack.org/keystone/latest/user/application_credentials.html)

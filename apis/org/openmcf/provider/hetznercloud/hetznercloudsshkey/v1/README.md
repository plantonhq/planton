# HetznerCloudSshKey

The **HetznerCloudSshKey** resource registers an SSH public key in a Hetzner Cloud account. Servers reference the key's numeric ID at creation time, injecting the public key into `authorized_keys` for password-less SSH access.

## What It Represents

A [Hetzner Cloud SSH Key](https://docs.hetzner.cloud/#ssh-keys) is a named public key stored at the account level. It supports RSA (>= 1024 bits), ED25519, and ECDSA key types in OpenSSH `authorized_keys` format.

## Bundled Resources

| Terraform Resource | Created When | Purpose |
|---|---|---|
| `hcloud_ssh_key` | Always | Registers the SSH public key in Hetzner Cloud |

This is a single-resource component — no optional or conditional sub-resources.

## Key Features

### Single-Field Spec

The only user-specified field is `publicKey`. The SSH key name is derived from `metadata.name`, and labels are computed from metadata. This eliminates naming inconsistencies across environments.

### Force Replacement on Key Change

Hetzner Cloud does not support in-place updates to key material. Changing `publicKey` triggers resource replacement (delete + create), producing a new `ssh_key_id`. Servers referencing the old ID are unaffected until redeployed.

### Automatic Labeling

Standard labels (`resource`, `resource_name`, `resource_kind`, `org`, `env`, `resource_id`) are applied to the Hetzner Cloud SSH key resource from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence.

### Key Type Support

- **ED25519** — recommended; shorter, faster, strong security
- **RSA** (>= 1024 bits) — broad compatibility with legacy systems
- **ECDSA** — supported but no practical advantage over ED25519

## Upstream Dependencies (What This Resource Needs)

None. `HetznerCloudSshKey` is a root resource with no foreign key dependencies.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec.sshKeyIds` | Inject SSH key into server at boot |

## Stack Outputs

| Output | Description |
|---|---|
| `ssh_key_id` | Hetzner Cloud numeric ID of the created SSH key (as string) |
| `fingerprint` | MD5 fingerprint of the SSH public key (e.g., `"aa:bb:cc:dd:..."`) |

## References

- [Hetzner Cloud SSH Keys Documentation](https://docs.hetzner.cloud/#ssh-keys)
- [Terraform hcloud_ssh_key Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/ssh_key)
- [Pulumi hcloud.SshKey Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/sshkey/)

# DigitalOcean SSH Key

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_ssh_key` resource at the pinned provider version.

## What this component models

An SSH public key registered on the DigitalOcean account, ready to be injected into droplets (and droplet autoscale pools) at create time. The account stores only the PUBLIC half; the private key never leaves your machine.

The component covers the provider's full argument surface:

- `key_name` -- the display name (renames apply in place)
- `public_key` -- the OpenSSH single-line public key material (create-only: any in-line change REPLACES the key)

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSshKey
metadata:
  name: ci-deploy-key
spec:
  keyName: ci-deploy-key
  publicKey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB5K1XlWQxr9nMytXvvFyzYZaFVNSTAmTUYbSGXPqIQd ci@example.com
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `ssh_key_id` | Numeric id of the key, as a string (the API identity and the ONLY import id) |
| `fingerprint` | MD5 colon-hex fingerprint DigitalOcean computed from the material |

## Behavior worth knowing

- **The key material is create-only, and comparison trims only outer whitespace.** A changed comment, a re-encoded body, or different casing REPLACES the key -- rotating the numeric id and the fingerprint droplets reference.
- **Deleting a key never touches droplets created with it.** They keep the key in their `authorized_keys`; deletion only stops NEW droplets from selecting it.
- **Imports take the numeric id only.** A fingerprint does not work as an import id even though droplets accept fingerprints as key references.
- **Malformed keys fail at apply, not validation.** DigitalOcean validates the material server-side; the spec enforces only presence.
- **One key body per account.** DigitalOcean deduplicates on the material, not the name -- a second resource embedding the same public key fails with `SSH Key is already in use on your account`. Register shared material once and reference it.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

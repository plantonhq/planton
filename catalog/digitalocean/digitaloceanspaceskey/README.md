# DigitalOcean Spaces Key

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_spaces_key` resource at the pinned provider version.

## What this kind models

An access-key pair for Spaces, DigitalOcean's S3-compatible object storage: the credential workloads actually sign requests with, optionally scoped to specific buckets through per-bucket grants.

- `key_name` -- the key's name in the control panel (renames apply in place)
- `grants[]` -- per-bucket access rows: a `bucket` reference plus a `permission` (`read`, `readwrite`, or `fullaccess`); a fullaccess grant names NO bucket and covers the whole account, and a key with no grants authorizes nothing

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSpacesKey
metadata:
  name: ci-uploads-key
spec:
  keyName: ci-uploads
  grants:
    - bucket:
        valueFrom:
          kind: DigitalOceanBucket
          name: app-assets
          fieldPath: status.outputs.bucket_id
      permission: readwrite
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `access_key` | The access key ID (the resource's API identity) |
| `secret_key` | The secret access key -- a SECRET, returned ONLY at creation and never retrievable again |

## Behavior worth knowing

- **The secret exists exactly once.** DigitalOcean returns `secret_key` only in the create response; there is no API to read it back. Both provisioners mark it sensitive in state -- capture it from the output, never expect to re-fetch it.
- **Rotation is destroy-and-recreate.** The key material is immutable; name and grants update in place (the grant list is replaced wholesale on every update).
- **The permission wall lives here.** The provider accepts ANY permission string and silently turns unknown values into an empty grant that authorizes nothing -- this spec rejects everything outside `read` / `readwrite` / `fullaccess` at validation.
- **No import.** The resource has no upstream importer, and the write-once secret could never round-trip anyway.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

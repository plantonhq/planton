# GcpPrivateCaPool — Pulumi Implementation

This directory contains the Pulumi implementation for a Certificate Authority Service CA pool from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.certificateauthority.CaPool`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `caPool` |
| `module/locals.go` | The pool ID defaulted from `metadata.name`, the merged labels |
| `module/ca_pool.go` | Enables the API; maps the issuance policy, publishing options, and encryption key; exports the outputs |
| `module/x509.go` | The baseline X.509 values: key usage, presence-based CA options, policy IDs, OCSP servers, extensions, name constraints |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.ca_pool_id`, defaulting to `metadata.name`.
- **Issuance policy** -- each lever sent only when set; RSA modulus bounds as decimal strings, 0 unset.
- **Baseline values** -- `CaOptions` and `KeyUsage` (with both usage groups) always sent when baseline values are, empty when the spec omits them.
- **X.509 `CaOptions`** -- presence-based: `is_ca: false` goes with `NonCa`, a path length of 0 with `ZeroMaxIssuerPathLength`.
- **`EncryptionSpec`** -- sent only when `kms_key_name` is set.
- **`name` output** -- the resource ID (the full resource name), the shape the Terraform module's `id` exports.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

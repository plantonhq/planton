# GcpPrivateCaCertificate — Terraform Implementation

This directory contains the Terraform implementation for a Certificate Authority Service certificate from the Planton spec: one `google_privateca_certificate`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare pool and authority IDs trimmed from their references, the certificate ID defaulted from `metadata.name`, null-for-empty optionals, the merged labels |
| `main.tf` | `google_privateca_certificate` from a CSR or a config block (subject, subject key ID, X.509, public key) |
| `outputs.tf` | `name`, `certificate_id`, `pem_certificate`, `pem_certificate_chain`, `issuer_certificate_authority` |

## Send Posture

- **`pool`, `certificate_authority`** -- the last path segment of each reference, so full names and bare IDs both work; an empty authority lets the pool choose.
- **`name`** -- `spec.certificate_id`, defaulting to `metadata.name`.
- **`config`** -- emitted only for the structured request; `key_usage` is sent with both groups (empty when the spec omits them), `public_key.format` defaults to `PEM`.
- **X.509 `ca_options`** -- presence-based: an unset `is_ca` is left out; `false` goes with the provider's `non_ca`, and a path length of 0 with `zero_max_issuer_path_length` (the provider reads a plain 0 as unset).
- **Labels** -- the spec's labels merged under the platform attribution labels.
- **`deletion_policy`** -- sent only when set, leaving the provider's `DELETE` (revoke) default otherwise.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GcpPrivateCaPool — Terraform Implementation

This directory contains the Terraform implementation for a Certificate Authority Service CA pool from the Planton spec: one `google_project_service` (API enablement) and one `google_privateca_ca_pool`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the pool ID defaulted from `metadata.name`, null-for-empty optionals, the merged labels |
| `main.tf` | `google_project_service`, `google_privateca_ca_pool` with its issuance policy (baseline X.509 values included), publishing options, and encryption block |
| `outputs.tf` | `name`, `ca_pool_id`, `location` |

## Send Posture

- **`name`** -- `spec.ca_pool_id`, defaulting to `metadata.name`.
- **Issuance policy** -- each lever sent only when set; a set message sends all of its flags. RSA modulus bounds go as the decimal strings the provider takes, and 0 stays unset.
- **Baseline values** -- the provider requires `ca_options` and `key_usage` (with both usage groups) whenever baseline values are sent; an omitted one is sent empty, which states nothing.
- **X.509 `ca_options`** -- presence-based: an unset `is_ca` or `max_issuer_path_length` is left out of certificates; `is_ca: false` is sent with the provider's `non_ca`, and a path length of 0 with `zero_max_issuer_path_length` (the provider reads a plain 0 as unset).
- **`encryption_spec`** -- emitted only when `kms_key_name` is set.
- **Labels** -- the spec's labels merged under the platform attribution labels.
- **`deletion_policy`** -- sent only when set, leaving the provider's `DELETE` default otherwise.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

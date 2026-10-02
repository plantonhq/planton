# GcpPrivateCaCertificateAuthority — Terraform Implementation

This directory contains the Terraform implementation for a Certificate Authority Service certificate authority from the Planton spec: one `google_privateca_certificate_authority`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare pool ID trimmed from the reference, the authority ID defaulted from `metadata.name`, the deletion guard defaulted to true, null-for-empty optionals, the merged labels |
| `main.tf` | `google_privateca_certificate_authority` with its config (subject, subject key ID, X.509), key spec, subordinate config, and access URLs |
| `outputs.tf` | `name`, `certificate_authority_id`, `state`, `pem_ca_certificate`, `pem_ca_certificates`, `ca_certificate_access_url`, `crl_access_urls` |

## Send Posture

- **`pool`** -- the last path segment of the spec's pool, so a reference's full name and a bare ID both work.
- **`certificate_authority_id`** -- defaulting to `metadata.name`.
- **`deletion_protection`** -- always sent; unset in the spec means true.
- **`skip_grace_period`, `ignore_active_certificates_on_deletion`** -- always sent (false by default).
- **X.509 `ca_options`** -- `is_ca` always sent (the spec requires it); `false` goes with the provider's `non_ca`, and a path length of 0 with `zero_max_issuer_path_length` (the provider reads a plain 0 as unset). `key_usage` is sent with both groups, empty when the spec omits them.
- **`subordinate_config`** -- the parent by reference, or the PEM issuer chain wrapped in the provider's `pem_issuer_chain` block.
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

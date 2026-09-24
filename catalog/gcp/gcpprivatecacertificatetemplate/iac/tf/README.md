# GcpPrivateCaCertificateTemplate — Terraform Implementation

This directory contains the Terraform implementation for a Certificate Authority Service certificate template from the Planton spec: one `google_project_service` (API enablement) and one `google_privateca_certificate_template`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the template ID defaulted from `metadata.name`, null-for-empty optionals, the merged labels |
| `main.tf` | `google_project_service`, `google_privateca_certificate_template` with its predefined values, identity constraints, and passthrough extensions |
| `outputs.tf` | `name`, `template_id` |

## Send Posture

- **`name`** -- `spec.template_id`, defaulting to `metadata.name`.
- **Predefined values** -- every block sent only when set; key usage groups too.
- **X.509 `ca_options`** -- presence-based: an unset `is_ca` is sent as the provider's `null_ca` (this resource would otherwise send false), a set one as itself; a path length of 0 goes through `zero_max_issuer_path_length` (the provider reads a plain 0 as unset).
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

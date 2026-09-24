# GcpDatastreamConnectionProfile — Terraform Implementation

This directory contains the Terraform implementation for a Datastream connection profile from the Planton spec: one `google_project_service` (API enablement) and one `google_datastream_connection_profile`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id and display name defaulted from `metadata.name`, null-for-empty optionals, the merged labels |
| `main.tf` | `google_project_service`, `google_datastream_connection_profile` with one dynamic block per profile type and connectivity option |
| `outputs.tf` | `name`, `connection_profile_id` |

## Send Posture

- **`connection_profile_id`, `display_name`** -- defaulting to `metadata.name`.
- **`bigquery_profile`, `srv_connection_format`** -- the spec's bools emit Google's empty marker blocks.
- **`private_connection`** -- the lifted leaf; the `private_connectivity` block is sent only when it is set.
- **Ports** -- zero is sent as null, so the provider's engine default applies (3306, 5432, 1521, 1433, 22).
- **`ssl_config`** -- declared (even empty) means sent.
- **Passwords, keys, certificates** -- sensitive in the provider; marked secret in Pulumi (`ToSecret`).
- **Empty strings, empty maps, false flags** -- sent as null.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

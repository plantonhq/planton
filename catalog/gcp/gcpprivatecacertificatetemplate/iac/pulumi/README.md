# GcpPrivateCaCertificateTemplate — Pulumi Implementation

This directory contains the Pulumi implementation for a Certificate Authority Service certificate template from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.certificateauthority.CertificateTemplate`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `certificateTemplate` |
| `module/locals.go` | The template ID defaulted from `metadata.name`, the merged labels |
| `module/certificate_template.go` | Enables the API; maps the predefined values, identity constraints, and passthrough extensions; exports the outputs |
| `module/x509.go` | The predefined X.509 values: key usage, presence-based CA options (`NullCa` for unset), policy IDs, OCSP servers, extensions, name constraints |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.template_id`, defaulting to `metadata.name`.
- **Predefined values** -- every block sent only when set.
- **X.509 `CaOptions`** -- an unset `is_ca` is sent as `NullCa`; a path length of 0 as `ZeroMaxIssuerPathLength`.
- **`name` output** -- the resource ID (the full resource name), the shape the Terraform module's `id` exports.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

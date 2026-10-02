# GcpPrivateCaCertificate — Pulumi Implementation

This directory contains the Pulumi implementation for a Certificate Authority Service certificate from the Planton spec: one `gcp.certificateauthority.Certificate`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `certificate` |
| `module/locals.go` | The bare pool and authority IDs, the certificate ID defaulted from `metadata.name`, the merged labels |
| `module/certificate.go` | Maps the CSR or the structured config (public key format defaulting to PEM); exports the outputs |
| `module/x509.go` | The certificate's X.509 fields and subject: key usage, presence-based CA options, policy IDs, OCSP servers, extensions, name constraints, and the subject builder |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Pool`, `CertificateAuthority`** -- the last path segment of each reference; an empty authority is left unset.
- **`Config`** -- sent only for the structured request; `KeyUsage` with both groups, `PublicKey.Format` defaulting to `PEM`.
- **X.509 `CaOptions`** -- `is_ca: false` goes with `NonCa`, a path length of 0 with `ZeroMaxIssuerPathLength`.
- **`name` output** -- the resource ID (the full resource name), the shape the Terraform module's `id` exports.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

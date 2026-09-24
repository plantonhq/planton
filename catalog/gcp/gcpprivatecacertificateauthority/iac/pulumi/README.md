# GcpPrivateCaCertificateAuthority — Pulumi Implementation

This directory contains the Pulumi implementation for a Certificate Authority Service certificate authority from the Planton spec: one `gcp.certificateauthority.Authority`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `certificateAuthority` |
| `module/locals.go` | The bare pool ID, the authority ID defaulted from `metadata.name`, the merged labels |
| `module/certificate_authority.go` | Maps the config, key spec, subordinate config, access URLs, and teardown guards; exports the outputs |
| `module/x509.go` | The CA certificate's X.509 fields and subject: key usage, presence-based CA options, policy IDs, OCSP servers, extensions, name constraints, and the subject builder |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Pool`** -- the last path segment of the spec's pool.
- **`DeletionProtection`** -- always sent; unset means true. **`SkipGracePeriod`**, **`IgnoreActiveCertificatesOnDeletion`** -- always sent.
- **X.509 `CaOptions`** -- `IsCa` always sent; false goes with `NonCa`, a path length of 0 with `ZeroMaxIssuerPathLength`. `KeyUsage` sent with both groups.
- **`pem_ca_certificate` output** -- the first entry of the provider's `pem_ca_certificates`; the access URLs come from its first `access_urls` entry, the same as the Terraform module.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

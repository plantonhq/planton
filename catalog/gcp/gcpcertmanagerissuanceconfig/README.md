# GcpCertManagerIssuanceConfig

## Overview

`GcpCertManagerIssuanceConfig` manages one Certificate Manager certificate issuance config — the recipe Google follows to issue Google-managed certificates from YOUR private CA (a Certificate Authority Service pool) instead of a public CA. A `GcpCertManagerCert` opts in by naming the config in `managed.issuance_config`, and many certificates share one config.

## Purpose

Private PKI usually means hand-rolled issuance and renewal. With an issuance config, Google requests every certificate from your pool with the key algorithm and lifetime set here, and renews it automatically once the rotation window is reached. Internal load balancers and service-to-service TLS get certificates that chain to your own root, with public-CA convenience.

## Before It Works

Google documents two prerequisites this kind does not create:

1. **An enabled certificate authority in the pool** — a `GcpPrivateCaCertificateAuthority` in the `GcpPrivateCaPool` this config names.
2. **The Certificate Manager service agent** (`service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com`) holding `roles/privateca.certificateRequester` on the pool.

The config itself creates without either; certificates that name it fail to issue until both are in place.

## Key Features

- One pool reference (`caPool`, a `GcpPrivateCaPool`'s `name` output) — the provider's two nested wrappers lifted away
- `RSA_2048` or `ECDSA_P256` keys, lifetimes from 21 to 30 days, and a rotation window validated against Google's seven-day rule
- Global by default; regional for regional certificates
- User labels merged beneath the platform's attribution labels, identically on both engines
- `deletionPolicy` (DELETE / PREVENT / ABANDON) controls what a destroy does
- Enables the Certificate Manager API on the target project (never disabled on destroy)

## Example Usage

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerIssuanceConfig
metadata:
  name: internal-tls-issuance
spec:
  caPool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: internal-pool
      fieldPath: status.outputs.name
  keyAlgorithm: ECDSA_P256
  lifetime: 2592000s
  rotationWindowPercentage: 66
```

A certificate then names it:

```yaml
spec:
  managed:
    domains:
      - api.internal.example.com
    issuanceConfig:
      valueFrom:
        kind: GcpCertManagerIssuanceConfig
        name: internal-tls-issuance
        fieldPath: status.outputs.issuance_config_id
```

## Outputs

| Output | Description |
|--------|-------------|
| `issuance_config_id` | Full resource name (`projects/{project}/locations/{location}/certificateIssuanceConfigs/{name}`) — what `managed.issuance_config` takes |
| `issuance_config_name` | The issuance config's name in GCP |
| `location` | The Certificate Manager location (`global` unless set) |

## Best Practices

1. **One config per policy, many certificates** — share a config across every certificate with the same key algorithm and lifetime.
2. **Pick the rotation window from the lifetime** — Google requires renewal at least 7 days after issuance and 7 days before expiry: 34–66 for 21 days, 24–76 for 30 days.
3. **Treat every field but labels as immutable** — changing the pool, key algorithm, lifetime, or window replaces the config; certificates already issued keep serving until their own renewal.
4. **Set `deletionPolicy: PREVENT`** while certificates depend on the config; Google also refuses deletion while a certificate references it.

## Related Kinds

- **GcpCertManagerCert** — names this config in `managed.issuance_config`
- **GcpPrivateCaPool** — the pool certificates are issued from
- **GcpPrivateCaCertificateAuthority** — the enabled authority the pool needs

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GCP Private CA Certificate Authority

A certificate authority in a Certificate Authority Service CA pool: a self-signed root, or a subordinate signed by another authority (by reference, activated on create) or by an outside CA. The pool issues through every enabled authority in it, so rotation is adding a new authority, enabling it, and retiring the old one -- relying parties that trust the pool never notice.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Certificate authority** -- a `privateca_certificate_authority` in the named pool, with its own CA certificate (subject, X.509 fields, lifetime) and signing key (a Google-managed HSM key or your own Cloud KMS key version), activated and enabled on create

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service admin permissions (`roles/privateca.caManager`) on the project; a subordinate signed by reference also needs `roles/privateca.certificateManager` on the parent's pool.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpPrivateCaPool`** -- the pool the authority lives in (`pool`).
- Optional: a parent **`GcpPrivateCaCertificateAuthority`** (`subordinateConfig.certificateAuthority`), a **`GcpKmsKey`** version (`keySpec.cloudKmsKeyVersion`, Enterprise pools), a **`GcpGcsBucket`** to publish to (`gcsBucket`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateAuthority
metadata:
  name: root-ca
spec:
  location: us-central1
  pool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: root-pool
  config:
    subjectConfig:
      subject:
        commonName: Example Root CA
        organization: Example
    x509Config:
      caOptions:
        isCa: true
      keyUsage:
        baseKeyUsage:
          certSign: true
          crlSign: true
  keySpec:
    algorithm: EC_P384_SHA384
```

```shell
planton apply -f root-ca.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The pool's region. Immutable. |
| `pool` | `StringValueOrRef` | A `GcpPrivateCaPool` reference, its full name, or its bare ID. Immutable. |
| `config` | `object` | The CA certificate: `subjectConfig` (subject, optional SANs), optional `subjectKeyId`, and `x509Config` with `caOptions.isCa` set. Immutable. |
| `keySpec` | `object` | Exactly one of `algorithm` (Google-managed HSM key) or `cloudKmsKeyVersion` (your own key). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The pool's project. |
| `certificateAuthorityId` | `string` | `metadata.name` | Immutable. |
| `type` | `string` | `SELF_SIGNED` | `SELF_SIGNED` or `SUBORDINATE`. Immutable. |
| `lifetime` | `string` | `315360000s` | The CA certificate's validity. Immutable. |
| `subordinateConfig` | `object` | -- | Exactly one of a parent `certificateAuthority` or a `pemIssuerChain`. |
| `pemCaCertificate` | `string` | -- | An outside CA's signed certificate, with `subordinateConfig.pemIssuerChain`. |
| `gcsBucket` | `StringValueOrRef` | Google-managed bucket | Where the CA certificate and CRLs are published. Immutable. |
| `userDefinedAccessUrls` | `object` | -- | URLs advertised instead of the bucket's. |
| `desiredState` | `string` | `ENABLED` | `ENABLED`, `STAGED` (only at create), or `DISABLED` (only after create). |
| `deletionProtection` | `bool` | `true` | A destroy fails until this is false. |
| `skipGracePeriod` | `bool` | `false` | Delete immediately instead of after 30 days. |
| `ignoreActiveCertificatesOnDeletion` | `bool` | `false` | Allow destroy while issued certificates are still valid. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `config.x509Config.caOptions.isCa` must be set.
- `keySpec` is exactly one of `algorithm` or `cloudKmsKeyVersion`; `subordinateConfig` is exactly one of `certificateAuthority` or `pemIssuerChain`.
- `subordinateConfig` and `pemCaCertificate` apply only to `type: SUBORDINATE`; `pemCaCertificate` needs `subordinateConfig.pemIssuerChain`.
- The subject has a `commonName`; a `subjectAltName` block lists at least one name.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/caPools/{pool}/certificateAuthorities/{id}` |
| `certificate_authority_id` | `string` | The authority's ID |
| `state` | `string` | `ENABLED`, `DISABLED`, `STAGED`, or `AWAITING_USER_ACTIVATION` |
| `pem_ca_certificate` | `string` | The authority's own CA certificate -- a root's trust anchor |
| `pem_ca_certificates` | `list(string)` | The chain, own certificate first |
| `ca_certificate_access_url` | `string` | Where Google publishes the CA certificate |
| `crl_access_urls` | `list(string)` | Where Google publishes the CRLs |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy is guarded twice.** `deletionProtection` defaults to true; with it false, destroy disables the authority and schedules deletion after 30 days (restorable until then) unless `skipGracePeriod`. The pool cannot be deleted while the authority exists, even in that window.
- **The certificate is forever.** `config`, `keySpec`, `type`, `lifetime`, and `gcsBucket` are immutable; changing one replaces the authority -- plan it as a rotation.
- **A subordinate signed by reference is activated on create**; one signed by an outside CA needs its signed certificate and issuer chain (read the CSR from Google after the first apply); with neither, it waits in `AWAITING_USER_ACTIVATION`.
- **Each authority bills a monthly fee** at its pool's tier (see `cost.yaml`).

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpPrivateCaPool** -- the pool the authority lives in
- **GcpPrivateCaCertificate** -- certificates it signs
- **GcpKmsKey** -- your own signing key (Enterprise pools)
- **GcpGcsBucket** -- where it publishes its certificate and CRLs

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

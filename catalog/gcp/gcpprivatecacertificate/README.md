# GCP Private CA Certificate

One X.509 certificate issued from a Certificate Authority Service CA pool, for a key its owner holds -- described by a certificate signing request or by structured fields (subject, SANs, key usage, and the public key). Everything but labels is immutable; destroying the resource revokes the certificate. The pool must be `ENTERPRISE`: Google cannot describe or revoke certificates in a `DEVOPS` pool, so neither engine could track one.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Certificate** -- a `privateca_certificate` signed by the pool (or the named authority in it), optionally through a template

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with certificate permissions (`roles/privateca.certificateManager`) on the pool, plus `roles/privateca.templateUser` on the template when one is named.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpPrivateCaPool`** (Enterprise tier) with at least one enabled **`GcpPrivateCaCertificateAuthority`**.
- Optional: a **`GcpPrivateCaCertificateTemplate`** (`certificateTemplate`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificate
metadata:
  name: api-server
spec:
  location: us-central1
  pool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: internal-servers
  lifetime: 2592000s
  pemCsr: |
    -----BEGIN CERTIFICATE REQUEST-----
    ...
    -----END CERTIFICATE REQUEST-----
```

```shell
planton apply -f certificate.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The pool's region. Immutable. |
| `pool` | `StringValueOrRef` | A `GcpPrivateCaPool` reference, its full name, or its bare ID. Immutable. |
| one request | -- | Exactly one of `pemCsr` or `config`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The pool's project. |
| `certificateId` | `string` | `metadata.name` | Immutable; not reusable in the pool after revocation. |
| `certificateAuthority` | `StringValueOrRef` | pool chooses | The signing authority (reference, full name, or bare ID). Immutable. |
| `certificateTemplate` | `StringValueOrRef` | none | A `GcpPrivateCaCertificateTemplate` in the same region. Immutable. |
| `lifetime` | `string` | `315360000s` | Cut short by the pool's and template's caps and the authority's expiry. Immutable. |
| `pemCsr` | `string` | -- | A PEM certificate signing request. |
| `config` | `object` | -- | `subjectConfig`, optional `subjectKeyId`, `x509Config`, and `publicKey` (`key`, base64 of the PEM; `format` `PEM`). |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (revokes), `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one of `pemCsr` or `config`.
- `config` needs a subject with a `commonName`, the X.509 fields, and a public key; a `subjectAltName` block lists at least one name.
- `publicKey.format` is `PEM` (the only value Google accepts).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/caPools/{pool}/certificates/{id}` |
| `certificate_id` | `string` | The certificate's ID |
| `pem_certificate` | `string` | The signed certificate |
| `pem_certificate_chain` | `list(string)` | The chain that verifies it, issuer first |
| `issuer_certificate_authority` | `string` | The authority that signed it |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy revokes.** Google keeps the revoked record, and the certificate ID cannot be reused in the pool.
- **The private key never reaches Google or the manifest.** The workload generates it and supplies the CSR or public key; the outputs are the certificate and its chain.
- **Enterprise pools only**, and each issue bills at the pool's tier (see `cost.yaml`).
- **Renewal is a new certificate**: change the ID (or the request) to issue a successor before the current one expires.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpPrivateCaPool** -- the pool that issues it
- **GcpPrivateCaCertificateAuthority** -- the authority that signs it
- **GcpPrivateCaCertificateTemplate** -- the shape it is issued with

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

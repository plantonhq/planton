# GCP Private CA Pool

A Certificate Authority Service CA pool: the trust anchor your services trust and the policy every certificate issued from it follows. Certificate authorities (`GcpPrivateCaCertificateAuthority`) live inside the pool and rotate in and out behind it; certificates (`GcpPrivateCaCertificate`) are issued from it. Choose the tier once: `ENTERPRISE` for long-lived certificates Google stores and can revoke, `DEVOPS` for high-volume, short-lived workload certificates it does not track.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `privateca.googleapis.com` on the project (never disabled on destroy)
- **CA pool** -- a `privateca_ca_pool` with its tier, issuance policy, publishing options, and optional at-rest encryption key

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service admin permissions (`roles/privateca.caManager` or `roles/privateca.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- None. **`GcpKmsKey`** is optional, for `kmsKeyName` (encryption of stored certificates at rest; Enterprise pools).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaPool
metadata:
  name: internal-tls
spec:
  location: us-central1
  tier: DEVOPS
  issuancePolicy:
    maximumLifetime: 2592000s
    allowedKeyTypes:
      - ellipticCurveSignatureAlgorithm: ECDSA_P256
```

```shell
planton apply -f private-ca-pool.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The CA Service region, e.g. `us-central1`. Immutable. |
| `tier` | `string` | `ENTERPRISE` or `DEVOPS`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `caPoolId` | `string` | `metadata.name` | 1-63 letters, digits, `-`, `_`. Immutable. |
| `issuancePolicy` | `object` | no policy | Allowed key types, maximum lifetime, backdating (at most 48h), request forms, identity constraints (with a CEL expression), and X.509 baseline values stamped onto every certificate. |
| `publishingOptions` | `object` | nothing published | `publishCaCert`, `publishCrl`, and `encodingFormat` (`PEM` or `DER`). |
| `kmsKeyName` | `StringValueOrRef` | Google encryption | A `GcpKmsKey` that encrypts stored certificates at rest. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `tier` is `ENTERPRISE` or `DEVOPS`.
- Each allowed key type is exactly one of `rsa` (a modulus range, minimum not above maximum) or `ellipticCurveSignatureAlgorithm` (`ECDSA_P256`, `ECDSA_P384`, `EDDSA_25519`).
- Durations are seconds with an `s` suffix; `backdateDuration` is at most `172800s`.
- The identity constraints' CEL `expression` is required when `celExpression` is set.
- OIDs have at least one non-negative arc; custom extensions need an OID and a value; `maxIssuerPathLength` is not negative.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/caPools/{id}` -- what authorities, certificates, and TLS consumers reference |
| `ca_pool_id` | `string` | The pool's ID |
| `location` | `string` | The pool's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The tier is forever.** DevOps pools issue about 3.5 times faster per authority but store no certificates: nothing can be listed, described, or revoked, and authorities use Google-managed keys only. A `GcpPrivateCaCertificate` needs an Enterprise pool.
- **Baseline values win over requests, and conflict-fail against templates.** A request's conflicting value is overwritten; a template's conflicting predefined value fails the request.
- **A pool cannot be deleted while it holds an authority** -- even one in its 30-day soft delete. Destroy authorities with `skipGracePeriod` first.
- **The pool itself bills nothing.** Authorities bill a monthly fee each; certificates bill per issue (see `cost.yaml`).

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpPrivateCaCertificateAuthority** -- the authorities inside the pool
- **GcpPrivateCaCertificate** -- certificates issued from the pool
- **GcpPrivateCaCertificateTemplate** -- reusable certificate shapes requests use
- **GcpManagedKafkaCluster**, **GcpRedisCluster** -- TLS consumers that reference the pool
- **GcpKmsKey** -- at-rest encryption of stored certificates

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GCP Secret Manager Secret

Creates a Secret Manager secret — the container for versioned secret payloads (API keys, passwords, certificates) — with replication or regional residency, rotation reminders, expiry, CMEK, an optional first version, and secret-scoped IAM grants. One manifest takes a workload from nothing to a readable secret: container, payload, and the `secretAccessor` grant to the consuming service account.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Secret** -- `secretmanager.Secret` (global, with replication control) or `secretmanager.RegionalSecret` (payloads never leave `spec.region`)
- **Secret Version** (optional) -- version 1 seeded from `initialVersion.data`
- **IAM Members** (optional) -- one additive grant per `iamMembers` entry, scoped to this secret
- **Secret Manager API enablement** -- `secretmanager.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the secret is created (directly or via a GcpProject reference).
- **Deployer permissions**: the least-privilege permission set the IaC runner's principal needs lives in [`iac/permissions.yaml`](iac/permissions.yaml).
- **For CMEK**: a KMS key (reference a GcpKmsKey resource) with the Secret Manager service agent granted `roles/cloudkms.cryptoKeyEncrypterDecrypter`.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSecretManagerSecret
metadata:
  name: db-password
spec:
  initialVersion:
    data:
      value: super-secret-value
  iamMembers:
    - role: roles/secretmanager.secretAccessor
      member:
        value: serviceAccount:app@my-project.iam.gserviceaccount.com
```

```shell
planton apply -f secret.yaml
```

## Configuration Reference

### Optional Fields (every field has a working default)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `secretId` | `string` | `metadata.name` | The secret ID (letters/digits/underscores/hyphens, ≤255). ForceNew. |
| `region` | `string` | global | Set for a REGIONAL secret (data residency); payloads never leave the region. ForceNew. |
| `replication` | `message` | automatic | GLOBAL only: omit for automatic placement; `auto` (with CMEK) or `userManaged` replicas. ForceNew. |
| `customerManagedEncryption` | `message` | Google-managed | REGIONAL only: CMEK key reference (same region as the secret). |
| `initialVersion` | `message` | none | Seeds version 1: `data` (managed secret; a literal or a valueFrom to any kind's output), `enabled`, `isBase64`, version-level `deletionPolicy` (DELETE/DISABLE/ABANDON). |
| `iamMembers` | `list` | `[]` | Additive secret-scoped grants: role + member (reference a GcpServiceAccount's `member` output) + optional IAM condition. |
| `expireTime` / `ttl` | `string` | none | Auto-delete the whole secret (RFC3339 timestamp XOR seconds duration). |
| `versionDestroyTtl` | `string` | immediate | Delayed version destruction (≥86400s): destroy first disables, restore window applies. |
| `rotation` + `topics` | — | none | Rotation REMINDERS via Pub/Sub (GCP rotates nothing itself); rotation requires topics, period requires nextRotationTime. |
| `versionAliases` | `map` | `{}` | Alias → version number (e.g. `prod: "3"`); re-point without touching consumers. |
| `labels` / `annotations` / `tags` | `map` | `{}` | Metadata surfaces; `tags` is ForceNew (org-policy binding). |
| `deletionProtection` | `bool` | `false` | Engine-side plan blocker, evaluated before deletionPolicy. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (all versions, unrecoverable), `PREVENT`, `ABANDON`. |

### Validation Rules

- **Replication is global-only**; **customerManagedEncryption is regional-only** — the two scopes' API truth, mirrored.
- **`expireTime` XOR `ttl`**; **rotation requires topics**; **rotationPeriod requires nextRotationTime**.
- **`versionDestroyTtl` ≥ 24h**; **at most 10 topics**; replication carries exactly one arm.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `secret_name` | `string` | Full resource name — the handle consumers mount (e.g. Cloud Run `valueFromSecret`) |
| `secret_id` | `string` | The short secret ID |
| `latest_version_name` | `string` | `…/versions/1` when `initialVersion` was configured; empty otherwise |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **The payload is a version, not the secret**: `initialVersion.data` seeds version 1 and is immutable — rotations add versions via GCP tooling or pipelines, and `versionAliases` re-points consumers.
- **Two destroy guards compose**: `deletionProtection` blocks the plan engine-side; `deletionPolicy: PREVENT` blocks at the API. Production credentials deserve both.
- **Whole-secret deletion destroys every version regardless of `versionDestroyTtl`** — the TTL protects individual version destruction only.
- **In charts, feed `initialVersion.data` via valueFrom** from a producing resource's sensitive output instead of a literal.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the workload identity granted `secretAccessor`
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — CMEK for payload encryption
- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — rotation-reminder delivery
- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) — mounts the secret via `valueFromSecret`

## Additional Resources

- [Secret Manager Overview](https://cloud.google.com/secret-manager/docs/overview)
- [Secret Rotation](https://cloud.google.com/secret-manager/docs/secret-rotation)
- [Regional Secrets](https://cloud.google.com/secret-manager/docs/regional-secrets-overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

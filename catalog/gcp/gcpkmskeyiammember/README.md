# GCP KMS Key IAM Member

Deploys a single ADDITIVE IAM grant ON a KMS crypto key (`google_kms_crypto_key_iam_member`) — one role, to one member, on one key. This is the least-privilege unit of customer-managed encryption (CMEK) access control: the grant merges into the key's IAM policy without touching any other member's bindings, and removal subtracts only this exact pair.

The canonical use is CMEK: every service that encrypts with a customer-managed key needs its service agent granted `roles/cloudkms.cryptoKeyEncrypterDecrypter` on that key. Granting it here — on the key — beats a project-wide grant on both least privilege and orchestration: the agent can use exactly this key, and the grant is a real dependency edge so the encrypted resource deploys after the permission it needs exists.

## What Gets Created

When you deploy a GcpKmsKeyIamMember resource, Planton provisions:

- **Crypto Key IAM Member** — one (role, member[, condition]) entry merged into the target key's IAM policy

Nothing else in the policy is read as owned or modified — grants made by other charts, teams, or tools are never clobbered.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing crypto key** — referenced via `cryptoKeyId`
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **The member must exist** — a service agent email, or reference a GcpServiceAccount

## Quick Start

Create a file `cmek-grant.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyIamMember
metadata:
  name: gcs-state-key-grant
spec:
  cryptoKeyId:
    value: projects/my-gcp-project-123/locations/us-central1/keyRings/app-ring/cryptoKeys/state-key
  role:
    value: roles/cloudkms.cryptoKeyEncrypterDecrypter
  member:
    value: serviceAccount:service-123456789@gs-project-accounts.iam.gserviceaccount.com
```

Deploy:

```shell
planton apply -f cmek-grant.yaml
```

Or compose with the key as a first-class node:

```yaml
spec:
  cryptoKeyId:
    valueFrom:
      kind: GcpKmsKey
      name: state-key
      fieldPath: status.outputs.key_id
  role:
    value: roles/cloudkms.cryptoKeyEncrypterDecrypter
  member:
    value: serviceAccount:service-123456789@gs-project-accounts.iam.gserviceaccount.com
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `cryptoKeyId` | `StringValueOrRef` | The key whose IAM policy receives the grant, as its fully-qualified resource path (`projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>`). References a GcpKmsKey's `key_id` output by default. Immutable. |
| `role` | `StringValueOrRef` | The role to grant ON the key: a predefined role (typically `roles/cloudkms.cryptoKeyEncrypterDecrypter`) or a custom role's fully-qualified name. References a GcpIamCustomRole's `name` output by default. Immutable. |
| `member` | `StringValueOrRef` | The identity receiving the grant, in IAM member format. References a GcpServiceAccount's `member` output by default. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `condition` | `object` | — | IAM Condition (`title`, `expression`, optional `description`) restricting when the grant applies. Part of the grant's identity. Immutable. |

There is no project or location field: both are embedded in the key's resource path.

### Member Formats

| Format | Grants to |
|--------|-----------|
| `serviceAccount:<email>` | A service account or Google service agent (the most common for CMEK) |
| `user:<email>` | A Google account |
| `group:<email>` | A Google group |
| `domain:<domain>` | Everyone in a Workspace/Cloud Identity domain |
| `principal://...` / `principalSet://...` | Workload identity federation principals |

Grants to deleted principals (`deleted:...`) are rejected at deploy time.

## Key-Scoped vs Ring-Scoped vs Project-Scoped

IAM on KMS flows down the resource hierarchy: a project-level grant covers every ring, a ring-level grant covers every key in the ring, and a key-level grant — this component — covers exactly one key. Use key-scoped grants when different keys in a ring serve different consumers (the common case once a ring hosts state, database, and artifact keys side by side). The key-scoped grant is also what gives a first CMEK deploy a real ordering edge: the encrypted bucket, dataset, or disk can depend on the grant instead of racing project-wide IAM propagation.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `crypto_key_id` | `string` | The key whose policy received the grant (after reference resolution) |
| `role` | `string` | The granted role (after reference resolution) |
| `member` | `string` | The granted member (after reference resolution) |
| `etag` | `string` | The key IAM policy etag after the grant — useful for audit correlation |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable**: IAM grants have no update — changing key, role, member, or condition replaces the grant atomically (destroy the old pair, create the new one), which mirrors the underlying API exactly.
- **Conditions are part of the grant's identity**: the same role granted with and without a condition are two independent grants that do not interfere.
- **Only additive grants are modeled**: authoritative per-role bindings and whole-policy writes clobber every grant they do not list and are deliberately not modeled.
- **Finding service agent emails**: each service documents its agent format (Cloud Storage: `service-<project_number>@gs-project-accounts.iam.gserviceaccount.com`; BigQuery: `bq-<project_number>@bigquery-encryption.iam.gserviceaccount.com`; and so on). Some agents are created lazily on the service's first use in the project.

## Related Kinds

- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — the key being granted on (its `key_id` output feeds this component)
- [GcpKmsKeyRing](/docs/catalog/gcp/gcpkmskeyring) — the key's parent ring (ring-level IAM flows down to every key)
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — a grantable identity (its `member` output feeds this component)
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — the project-scoped counterpart

## Additional Resources

- [CMEK Overview](https://cloud.google.com/kms/docs/cmek)
- [KMS Permissions and Roles](https://cloud.google.com/kms/docs/reference/permissions-and-roles)
- [IAM Conditions](https://cloud.google.com/iam/docs/conditions-overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

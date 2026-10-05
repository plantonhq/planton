# GCP Service Account

Deploys a Google Cloud service account (`google_service_account`) — the identity that workloads (GKE pods, Cloud Run services, Cloud Functions, Compute instances, CI/CD pipelines) authenticate as — with optional key creation and additive project- or organization-level role grants.

## What Gets Created

When you deploy a GcpServiceAccount resource, Planton provisions:

- **Service Account** — a `google_service_account` in the specified project
- **User-managed key** (optional, off by default) — a `google_service_account_key` when the `userManagedKey` block is present, shaped by its algorithm/format/rotation fields
- **Project role grants** (optional) — one additive `google_project_iam_member` per role in `projectIamRoles`
- **Organization role grants** (optional) — one additive `google_organization_iam_member` per role in `orgIamRoles`

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** — the deploying principal's permissions (per feature: account, optional key, project-role grants, org-role grants) are listed in [`iac/permissions.yaml`](iac/permissions.yaml)

## Quick Start

Create a file `service-account.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceAccount
metadata:
  name: my-app-identity
spec:
  serviceAccountId: my-app-prod
  projectId:
    value: my-gcp-project-123
  displayName: My App (production)
  description: Runtime identity for the my-app Cloud Run service
  projectIamRoles:
    - roles/logging.logWriter
    - roles/monitoring.metricWriter
```

Deploy:

```shell
planton apply -f service-account.yaml
```

This creates `my-app-prod@my-gcp-project-123.iam.gserviceaccount.com` with two additive role grants and no key (keyless is the default).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `serviceAccountId` | `string` | Short account ID forming the email `<id>@<project>.iam.gserviceaccount.com`. Immutable — changing it recreates the account and invalidates every reference to the old email. | 6-30 chars; lowercase letters, digits, hyphens; starts with a letter; cannot end with a hyphen |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project in which the account is created. Can reference a GcpProject. Immutable. |
| `displayName` | `string` | metadata name | Human-readable identity shown in the GCP console. Mutable. |
| `description` | `string` | `""` | What this identity is for (max 256 bytes). Mutable. |
| `disabled` | `bool` | `false` | Disabled accounts keep their IAM bindings but cannot authenticate — a kill switch for incident response or staged decommissioning. Mutable. |
| `userManagedKey` | `object` | absent (keyless) | Presence creates a user-managed key; see the key table below. Prefer keyless patterns. |
| `projectIamRoles` | `list(string)` | `[]` | Roles granted at the project scope (additive member grants). |
| `orgId` | `string` | `""` | Numeric organization ID; required when `orgIamRoles` is set. |
| `orgIamRoles` | `list(string)` | `[]` | Roles granted at the organization scope (additive; affects every project under the org — grant sparingly). |
| `createIgnoreAlreadyExists` | `bool` | `false` | Adopt an existing account with the same email instead of failing — idempotent bootstrap flows. |
| `deletionPolicy` | `string` | `DELETE` | `PREVENT` fails destroys — a guard rail for identities whose deletion would invalidate IAM bindings fleet-wide. |

### User-managed key (`userManagedKey`)

An empty block (`userManagedKey: {}`) creates the classic 2048-bit RSA JSON key. Two mutually exclusive flows: GENERATE (GCP creates the key pair; the private key exports once as `key_base64`) and UPLOAD (`publicKeyData` supplies your own public key; GCP never sees a private key).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `algorithm` | `string` | `KEY_ALG_RSA_2048` | Key algorithm (`KEY_ALG_RSA_2048` / `KEY_ALG_RSA_1024`). Create-time only. |
| `privateKeyType` | `string` | `TYPE_GOOGLE_CREDENTIALS_FILE` | Private key format; `TYPE_PKCS12_FILE` for legacy p12 tooling. Generate flow only. |
| `publicKeyType` | `string` | `TYPE_X509_PEM_FILE` | Public key output format (`TYPE_RAW_PUBLIC_KEY`, `TYPE_NONE`). Generate flow only. |
| `publicKeyData` | `string` | `""` | Base64 X.509 PEM public key — the upload flow; conflicts with the two `*KeyType` fields. |
| `keepers` | `map(string)` | `{}` | Rotation trigger: changing any value replaces the key. |
| `deletionPolicy` | `string` | `DELETE` | `PREVENT` fails destroys while the key exists. |

### Role lists vs first-class grants

The role lists are a convenience for the common "identity plus its obvious roles" case. For grants that deserve to be visible, independently-owned nodes in the resource graph — custom roles, conditional grants, grants owned by a different chart — use [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember), which references this account's `member` output directly.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `email` | `string` | The service account email — the handle workload configs attach identity by |
| `member` | `string` | Ready-made IAM member string (`serviceAccount:<email>`) — feed directly into IAM grants |
| `unique_id` | `string` | Stable numeric ID, never reused across delete/recreate |
| `name` | `string` | Fully-qualified resource name (`projects/<project>/serviceAccounts/<email>`) |
| `key_base64` | `string` | Base64-encoded private key (generate flow only — empty for keyless accounts and the `publicKeyData` upload flow; sensitive) |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Keyless by default**: no user-managed key is created unless the `userManagedKey` block is present. Prefer Workload Identity (GKE), attached service accounts (Cloud Run/Compute), or Workload Identity Federation (external CI/CD) — a key is a long-lived credential that must be rotated and protected. When a key is unavoidable, `keepers` gives you declarative rotation and `publicKeyData` keeps the private key in your custody.
- **Immutability**: `serviceAccountId` and `projectId` are ForceNew. `displayName`, `description`, and `disabled` update in place (GCP flips disabled state via dedicated Enable/Disable API calls, handled transparently).
- **Additive grants**: all role grants use member-level (additive) semantics — they never clobber other members' bindings on the same role.
- **Email reuse caution**: deleting and recreating an account produces the same email but a different `unique_id`; IAM bindings referencing the old account do not transfer.

## Related Kinds

- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — first-class additive grant referencing this account's `member` output
- [GcpIamCustomRole](/docs/catalog/gcp/gcpiamcustomrole) — least-privilege role definitions to grant to this account
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the project the account lives in
- [GcpGkeWorkloadIdentityBinding](/docs/catalog/gcp/gcpgkeworkloadidentitybinding) — binds a Kubernetes service account to this identity for keyless GKE workloads

## Additional Resources

- [Service Accounts Overview](https://cloud.google.com/iam/docs/service-account-overview)
- [Best Practices for Using Service Accounts](https://cloud.google.com/iam/docs/best-practices-service-accounts)
- [Migrating Away from Service Account Keys](https://cloud.google.com/iam/docs/migrate-from-service-account-keys)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

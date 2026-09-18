# GCP Cloud SQL User

Deploys a database user (`google_sql_user`) on a Cloud SQL instance. Users are first-class nodes: one per application/service with its own credential — classic username/password (`BUILT_IN`) or passwordless IAM authentication (`CLOUD_IAM_USER` / `CLOUD_IAM_SERVICE_ACCOUNT` / `CLOUD_IAM_GROUP`).

## What Gets Created

When you deploy a GcpCloudSqlUser resource, Planton provisions:

- **Cloud SQL user** — a `google_sql_user` on the referenced instance

No API enablement is needed: the instance the user lives on cannot exist without `sqladmin.googleapis.com` already enabled.

## Prerequisites

- **An existing Cloud SQL instance** — referenced via `instance` (a [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) resource or a literal instance name)
- **For IAM users on PostgreSQL** — the instance must set the database flag `cloudsql.iam_authentication = "on"`
- **For service-account users** — the account needs `roles/cloudsql.instanceUser` (and `roles/cloudsql.client` to connect) on the project, granted through [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember)
- **GCP credentials** — [`iac/permissions.yaml`](iac/permissions.yaml) lists the exact least-privilege permissions

## Quick Start

Create a file `user.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlUser
metadata:
  name: orders-app-user
spec:
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: my-postgres
      fieldPath: status.outputs.instance_name
  userName: orders-app
  password: a-strong-generated-password
```

A passwordless service-account user, wired by reference (the instance must have `cloudsql.iam_authentication = "on"`):

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlUser
metadata:
  name: orders-runner-user
spec:
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: my-postgres
      fieldPath: status.outputs.instance_name
  type: CLOUD_IAM_SERVICE_ACCOUNT
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: orders-runner
      fieldPath: status.outputs.email
```

Deploy:

```shell
planton apply -f user.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `instance` | `StringValueOrRef` | The hosting Cloud SQL instance. Immutable. | Ref → GcpCloudSql `instance_name` |
| `userName` | `string` | Login name (BUILT_IN) or IAM principal (IAM types; for a service account, the email with `.gserviceaccount.com` dropped — both engines normalize a pasted full email). Immutable. Exactly one of `userName` or `serviceAccount`. | 1–128 chars |
| `serviceAccount` | `StringValueOrRef` | For `CLOUD_IAM_SERVICE_ACCOUNT` users: the service account wired by reference (a GcpServiceAccount's `email` output) or as a literal email; both engines derive the database username. Immutable. Exactly one of `userName` or `serviceAccount`. | Ref → GcpServiceAccount `email`; requires `type: CLOUD_IAM_SERVICE_ACCOUNT` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. |
| `password` | `string` (secret) | — | BUILT_IN credential. Mutable — updating it rotates in place. Never set for IAM types. |
| `type` | `string` | `BUILT_IN` | `BUILT_IN`, `CLOUD_IAM_USER`, `CLOUD_IAM_SERVICE_ACCOUNT`, or `CLOUD_IAM_GROUP`. Immutable. |
| `host` | `string` | — | MySQL only: `user@host` scoping (e.g. `%`). Immutable. |
| `passwordPolicy` | object | — | Per-user hardening: `allowedFailedAttempts`, `passwordExpirationDuration`, `enableFailedAttemptsCheck`, `enablePasswordVerification` (MySQL). BUILT_IN only. |

### Validation Rules (enforced pre-deploy)

- IAM-authenticated types must not set a `password`.
- Exactly one of `userName` or `serviceAccount` is set; `serviceAccount` requires `type: CLOUD_IAM_SERVICE_ACCOUNT`.
- `passwordPolicy` applies to `BUILT_IN` users only.
- `passwordExpirationDuration` must be a seconds duration string (e.g. `2592000s`).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `user_name` | `string` | The user name as stored by Cloud SQL (IAM users on MySQL are stored truncated before the `@`) |
| `instance_name` | `string` | Name of the Cloud SQL instance this user belongs to |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **The password is secret-by-default** — encrypted in IaC state, never exported in outputs.
- **Rotation is in-place** — updating `password` updates the credential without recreating the user.
- **Database privileges are schema territory** — GRANTs inside the database are applied by migrations/application tooling, not by this resource (the API manages the login, not its grants).
- **PostgreSQL user deletion** — a user that owns objects cannot be deleted until ownership is reassigned; plan teardown accordingly.

### Deliberately not modeled (recorded reasons)

Everything else on `google_sql_user` at the pinned provider is representable — including `databaseRoles` (roles granted at creation on MySQL 8+ / PostgreSQL) and `deletionPolicy`, whose `ABANDON` mode is the documented answer for PostgreSQL users that still own database objects. The one recorded exclusion:

| Excluded Feature | Why |
|---|---|
| `password_wo` / `password_wo_version` | Write-only variants of the modeled `password` — same capability through engine-side ergonomics; the spec field is secret-annotated and encrypted in state on both engines. |

## Related Components

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance this user lives on
- [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) — pair each user with its application database
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity behind `CLOUD_IAM_SERVICE_ACCOUNT` users

## Additional Resources

- [Creating and managing users](https://cloud.google.com/sql/docs/postgres/create-manage-users)
- [IAM database authentication](https://cloud.google.com/sql/docs/postgres/iam-authentication)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

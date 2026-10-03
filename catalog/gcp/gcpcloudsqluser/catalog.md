# GCP Cloud SQL User

Creates a database user on an existing Google Cloud SQL instance — a classic password (BUILT_IN) user, or a passwordless IAM-authenticated user, service account, or group. Users as first-class resources give every application its own principal on shared infrastructure: created, rotated, and revoked independently, never sharing the admin credential.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud SQL User** -- a `google_sql_user` on the referenced instance: BUILT_IN (username + password, with an optional per-user password policy) or one of the three IAM types (CLOUD_IAM_USER, CLOUD_IAM_SERVICE_ACCOUNT, CLOUD_IAM_GROUP — passwordless, authenticated through IAM)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Org Secret** (BUILT_IN users) -- store the password as an org secret and reference it as `$secret/<slug>`; the runner resolves it just-in-time and the platform rejects plaintext.

### GCP Project

- **A Cloud SQL instance** -- the [GcpCloudSql](/infra-catalog/gcp-cloud-sql) instance the user authenticates against. Reference it via ValueFromRef so the pipeline deploys the instance first.
- **IAM authentication flag** (IAM-typed users on PostgreSQL) -- the instance must carry the database flag `cloudsql.iam_authentication: "on"` before IAM users can be created.

## Deploy

### Console

Open the deployment store, find **GCP Cloud SQL User**, and click **Deploy**. The creation wizard walks two decisions: which instance the user lives on, then who the user is and how it authenticates. Start from the **Application User (Built-in)** or **IAM Service Account User (Passwordless)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlUser
metadata:
  name: orders-app
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: orders-db-prod
      fieldPath: status.outputs.instance_name
  userName: orders-app
  password: $secret/orders-app-db-password
```

```shell
planton apply -f user.yaml
```

This creates a BUILT_IN user named `orders-app` on the referenced instance, its password resolved from the org secret at deploy — plaintext never enters the manifest. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the hosting instance by reference so the pipeline orders the graph:

```yaml
spec:
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: orders-db-prod
      fieldPath: status.outputs.instance_name
  userName: orders-app
  password: $secret/orders-app-db-password
```

The InfraPipeline deploys the Cloud SQL instance first, then creates the user on it.

## Key Configuration

These are the most important decisions when configuring a Cloud SQL user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication type** -- Immutable. `BUILT_IN` (the default) is a classic password user. The IAM types authenticate through Google IAM instead: nothing to store, leak, or rotate, and access is revoked by revoking the IAM role — the strongest posture for workloads that already run as a service account (GKE Workload Identity, Cloud Run).

**User name** -- Immutable, max 128 characters. For IAM types it is the principal's full email; on MySQL, GCP stores IAM user names truncated before the `@` (the `user_name` output reflects the stored name).

**Password** -- BUILT_IN only, and MUTABLE: updating the value rotates the credential in place, making this resource the rotation lever. Always a `$secret/<slug>` reference. Subject to the instance's password validation policy.

**Host scope** -- MySQL only: classic `user@host` scoping (e.g. `10.0.0.%`). Immutable; leave empty for PostgreSQL/SQL Server.

**Per-user password policy** -- BUILT_IN only: failed-attempts lockout, password expiration, and (MySQL) current-password verification — layered on the instance-wide policy.

**Deletion policy** -- `DELETE` (the default) drops the user on destroy; `PREVENT` fails any plan that would drop it. `ABANDON` removes it from IaC management while leaving it on the instance — the documented answer for PostgreSQL users that cannot be dropped while they still own database objects.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCloudSql** | `instance` | `status.outputs.instance_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_name` | The stored user name (IAM names may be truncated on MySQL) | Application connection strings |
| `instance_name` | The hosting instance's name | Cross-checking attachment |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One user per application** -- each service gets its own BUILT_IN principal with its password in an org secret; rotation is updating the secret and redeploying (the `password` field is the one mutable knob), and revocation is destroying one resource — the admin credential is never shared. Start from the **Application User (Built-in)** preset.

**Passwordless workload identity** -- a `CLOUD_IAM_SERVICE_ACCOUNT` user for workloads that already run as a service account (Cloud Run, GKE Workload Identity): nothing to store, leak, or rotate, and access is revoked at the IAM layer. On PostgreSQL, the instance must first set `cloudsql.iam_authentication: "on"`. Start from the **IAM Service Account User (Passwordless)** preset.

**Roles at creation** -- `databaseRoles` (MySQL 8+ / PostgreSQL) grants predefined or pre-created roles as the user is born, so the grant travels with the resource instead of living in an out-of-band SQL script.

## Works With

- [**GCP Cloud SQL**](/infra-catalog/gcp-cloud-sql) -- the instance this user authenticates against
- [**GCP Cloud SQL Database**](/infra-catalog/gcp-cloud-sql-database) -- the database the application connects to with these credentials
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the workload identity behind a CLOUD_IAM_SERVICE_ACCOUNT user

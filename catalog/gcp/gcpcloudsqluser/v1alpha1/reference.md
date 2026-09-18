# GcpCloudSqlUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudSqlUserSpec defines a database user (`google_sql_user`) on a Cloud
SQL instance.

Users are first-class nodes: create one per application/service with its
own credential instead of sharing the instance's admin user. Two auth
families exist:
  - BUILT_IN — classic username + password, enforced by the instance's
    password validation policy.
  - IAM types (CLOUD_IAM_USER / CLOUD_IAM_SERVICE_ACCOUNT /
    CLOUD_IAM_GROUP) — passwordless database authentication through IAM.
    On PostgreSQL, the instance must first set the database flag
    "cloudsql.iam_authentication" = "on".

## Example

```yaml
# Exercises the user surface offline: a BUILT_IN app user with a hardening
# policy, referencing its instance by literal name.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlUser
metadata:
  name: hack-orders-app
spec:
  # project_id omitted — falls back to the provider's default project.
  instance:
    value: hack-postgres
  userName: orders-app
  password: HackAppPassword123!  # replace before applying anywhere real
  type: BUILT_IN
  passwordPolicy:
    allowedFailedAttempts: 5
    enableFailedAttemptsCheck: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instance` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.instance_name`) |
| `spec.userName` | `string` | yes |  |  |
| `spec.password` | `string` (sensitive) | yes |  |  |
| `spec.type` | `string` |  | `BUILT_IN` |  |
| `spec.host` | `string` |  |  |  |
| `spec.passwordPolicy` | `GcpCloudSqlUserPasswordPolicy` |  |  |  |
| `spec.passwordPolicy.allowedFailedAttempts` | `int32` |  |  |  |
| `spec.passwordPolicy.passwordExpirationDuration` | `string` |  |  |  |
| `spec.passwordPolicy.enableFailedAttemptsCheck` | `bool` |  |  |  |
| `spec.passwordPolicy.enablePasswordVerification` | `bool` |  |  |  |
| `spec.databaseRoles` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the Cloud SQL instance.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instance

`string | valueFrom` · required

The Cloud SQL instance this user lives on. Accepts the instance name or
a reference to a GcpCloudSql resource. Immutable — a user cannot move
between instances.

- references: GcpCloudSql (`status.outputs.instance_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.userName

`string` · required

The user name. Immutable. For BUILT_IN users this is the login name
("orders-app"). For IAM types it is the IAM principal: the user's
full email for CLOUD_IAM_USER, the group email for CLOUD_IAM_GROUP,
and for CLOUD_IAM_SERVICE_ACCOUNT the service account's email with
the ".gserviceaccount.com" suffix dropped
("ci-runner@my-project.iam") — the form Cloud SQL stores on
PostgreSQL; MySQL keeps only the part before "@". Both engines apply
that normalization for CLOUD_IAM_SERVICE_ACCOUNT, so a pasted full
email also works. Prefer service_account for the service-account
case: it wires the identity by reference instead of by hand. Exactly
one of user_name or service_account is set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"128"}}

### spec.password

`string` · required · sensitive

Login password for a BUILT_IN user. Mutable — updating it rotates the
credential in place. Subject to the instance's password validation
policy. Never set for IAM-authenticated types.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1"}}

### spec.type

`string` · optional (explicit presence)

Authentication type. BUILT_IN (default) uses username + password; the
CLOUD_IAM_* types authenticate through IAM without a stored password.
Immutable.

- default: `BUILT_IN`
- rule: type must be BUILT_IN, CLOUD_IAM_USER, CLOUD_IAM_SERVICE_ACCOUNT, or CLOUD_IAM_GROUP

### spec.host

`string`

MySQL only: the host the user may connect from (classic MySQL
user@host semantics, e.g. "%" for any host or "10.0.0.0/8"). Leave
empty for PostgreSQL and SQL Server. Immutable.

### spec.passwordPolicy

`GcpCloudSqlUserPasswordPolicy`

Per-user password policy (BUILT_IN users only), layered on top of the
instance-level password validation policy.

### spec.passwordPolicy.allowedFailedAttempts

`int32` · optional (explicit presence)

Number of failed login attempts after which the user is locked
(requires enable_failed_attempts_check).

- rule: {"int32":{"gte":1}}

### spec.passwordPolicy.passwordExpirationDuration

`string`

Password lifetime as a seconds duration string, e.g. "2592000s" (30
days). After expiry the user must change the password to log in.

- rule: password_expiration_duration must be a seconds duration string such as 2592000s

### spec.passwordPolicy.enableFailedAttemptsCheck

`bool`

Whether failed login attempts are counted toward
allowed_failed_attempts.

### spec.passwordPolicy.enablePasswordVerification

`bool`

MySQL only: require the current password when changing the password.

### spec.databaseRoles

`[]string`

MySQL 8+ / PostgreSQL only: database roles granted to the user at
creation — predefined Cloud SQL roles (e.g. "cloudsqlsuperuser") or
custom roles already created in the database.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.deletionPolicy

`string`

Engine-side teardown behavior. "DELETE" (default) drops the user;
"PREVENT" fails any plan that would drop it; "ABANDON" removes it
from IaC management while leaving it on the instance. ABANDON is the
documented answer for PostgreSQL users that cannot be dropped while
they still own database objects.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.serviceAccount

`string | valueFrom`

The service account this IAM database user represents, wired by
reference (a GcpServiceAccount's email) or as a literal email. Both
engines derive the database username Cloud SQL expects — the email
with ".gserviceaccount.com" dropped (Cloud SQL stores exactly that on
PostgreSQL, and only the part before "@" on MySQL) — so a chart never
hand-builds the form. Requires type CLOUD_IAM_SERVICE_ACCOUNT and the
instance's "cloudsql.iam_authentication" flag; the account still
needs roles/cloudsql.instanceUser (and roles/cloudsql.client to
connect) on the project, granted separately. Alternative to
user_name: exactly one of the two is set. Immutable.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

## Validation Rules

- `iam_user_must_not_set_password`: IAM-authenticated users (CLOUD_IAM_USER, CLOUD_IAM_SERVICE_ACCOUNT, CLOUD_IAM_GROUP) must not set a password — authentication goes through IAM
- `password_policy_requires_built_in`: password_policy applies to BUILT_IN users only
- `user_name_xor_service_account`: set exactly one of user_name (a login name or IAM principal) or service_account (a GcpServiceAccount whose IAM database username is derived)
- `service_account_requires_iam_sa_type`: service_account derives an IAM service-account database user -- set type to CLOUD_IAM_SERVICE_ACCOUNT

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudSqlUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_name` | `string` | The user name as stored by Cloud SQL. For IAM users on MySQL this is the truncated form (email without the "@domain" suffix) — the name clients actually authenticate with. |
| `status.outputs.instance_name` | `string` | Name of the Cloud SQL instance this user belongs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.instance` | GcpCloudSql | `status.outputs.instance_name` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)

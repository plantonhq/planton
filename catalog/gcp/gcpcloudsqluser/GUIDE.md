# GcpCloudSqlUser Guide

The judgment this guide protects: users are per-service identities, not
shared logins. The design decision that matters is the auth family —
password or IAM — because it decides where credential risk lives and it
is immutable per user.

## IAM-typed users move the credential problem to IAM

`CLOUD_IAM_USER` / `CLOUD_IAM_SERVICE_ACCOUNT` / `CLOUD_IAM_GROUP` users
carry NO stored password — authentication rides IAM tokens through the
Auth Proxy or connectors, rotation stops being your problem, and access
review becomes an IAM review. On PostgreSQL the instance must first set
`cloudsql.iam_authentication = "on"` in `databaseFlags`; the group type
grants a whole team through one node. Prefer IAM types everywhere the
client library stack supports them; BUILT_IN is for engines and tools
that genuinely need a password.

## Service-account users are wired by reference

For `CLOUD_IAM_SERVICE_ACCOUNT` users, set `serviceAccount` to the
`GcpServiceAccount` (a `valueFrom` on its `email` output, or a literal
email) instead of `userName`. Both engines derive the database username
Cloud SQL expects — the email with `.gserviceaccount.com` dropped, which
PostgreSQL stores exactly and MySQL truncates before the `@` — so a chart
never hand-builds `sa-id@project.iam` from parameters, and a wrong project
id or account id cannot silently produce a user nobody can log in as.
The account still needs `roles/cloudsql.instanceUser` (and
`roles/cloudsql.client` to connect) on the project, granted through
`GcpProjectIamMember`, and the instance needs
`cloudsql.iam_authentication = "on"`. Exactly one of `userName` or
`serviceAccount` is set.

## Passwords rotate in place; roles land at creation

For BUILT_IN users, updating `password` rotates the credential without
recreating the user — pair it with the per-user `passwordPolicy` (and
the instance-level validation policy) rather than inventing rotation
tooling. `databaseRoles` grants roles AT CREATION on MySQL 8+ and
PostgreSQL — predefined ones like `cloudsqlsuperuser`, or custom roles
that must already exist in the database. In-database GRANTs beyond that
remain migration territory: the API manages the login, not its schema
privileges.

## Destroy stance

PostgreSQL will not drop a user that still owns objects — reassign
ownership first, or the destroy hangs on the API's refusal.
`deletionPolicy: ABANDON` is the documented workaround when the
teardown must proceed anyway: the login stays, IaC forgets it, ownership
gets fixed on your schedule. MySQL users scoped with `host` are
identities per user@host pair — dropping the node drops exactly that
pair.

## On the diagram

A leaf under its `GcpCloudSql` instance (referenced by `instance`),
typically paired one-to-one with a `GcpCloudSqlDatabase`. IAM
service-account users point back at the `GcpServiceAccount` whose email
they carry — the diagram shows the workload identity chain end to end.

## Pairs well with

- `GcpCloudSql` — the instance; its password validation policy governs
  BUILT_IN users here.
- `GcpCloudSqlDatabase` — the application database this user works in.
- `GcpServiceAccount` — the identity behind CLOUD_IAM_SERVICE_ACCOUNT
  users.

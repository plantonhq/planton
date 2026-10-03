# AwsRedshiftServerlessNamespace

An Amazon Redshift Serverless namespace -- the DATA plane of the serverless warehouse: the database, its admin credentials, the KMS key that encrypts stored data, the IAM roles the engine assumes, and audit log exports.

A namespace stores; it never computes. Compute lives on `AwsRedshiftServerlessWorkgroup` nodes that attach to this namespace by name -- many workgroups can serve one namespace (a capped dev workgroup and an autoscaling production workgroup over the same data), and each is created and destroyed without touching the data. That split is AWS's own resource model, mirrored here as two composable nodes. KMS keys and IAM roles compose by reference.

## Spec highlights

- **Credentials** -- `manageAdminPassword` (recommended) keeps the admin password in Secrets Manager, generated and rotated by AWS, with the secret's ARN exported; or supply `adminUserPassword` directly (sensitive). A namespace does not hard-require admin credentials at all -- IAM identities can use temporary credentials without one.
- **First database** -- `dbName` (create-time only; empty keeps the AWS default `dev`). Additional databases are created with SQL.
- **Data encryption** -- `kmsKeyId` by reference for a customer-managed key; empty keeps the AWS-owned Redshift service key.
- **Data movement** -- `iamRoles` + `defaultIamRoleArn` (by reference) for COPY/UNLOAD/Spectrum.
- **Observability** -- `logExports` streams connection/user-activity/user audit logs to CloudWatch Logs.

## Outputs

`namespace_name` (the join key workgroups attach with), `namespace_id`, `arn`, `db_name`, `admin_password_secret_arn`.

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsRedshiftServerlessNamespaceIacInput` (provider credentials + IaC info).

## References

- Redshift Serverless overview: https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-whatis.html
- Namespaces and workgroups: https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-workgroups-and-namespaces.html
- Managed admin passwords: https://docs.aws.amazon.com/redshift/latest/mgmt/redshift-secrets-manager-integration.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

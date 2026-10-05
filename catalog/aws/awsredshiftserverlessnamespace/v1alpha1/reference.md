# AwsRedshiftServerlessNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRedshiftServerlessNamespaceSpec defines an Amazon Redshift
Serverless namespace -- the DATA plane of the serverless warehouse: the
database, its admin credentials, the KMS key that encrypts stored
data, the IAM roles the engine assumes for COPY/UNLOAD/Spectrum, and
audit log exports. A namespace stores; it never computes.

Compute lives on AwsRedshiftServerlessWorkgroup nodes that attach to
this namespace by name -- many workgroups can serve one namespace,
each with its own capacity, VPC placement, and endpoint, and each can
be created and destroyed without touching the data. That split is
AWS's own resource model, mirrored here as two composable nodes.

The namespace name comes from metadata.name (create-time immutable in
AWS). KMS keys and IAM roles compose by reference -- this namespace
never creates or mutates resources that deserve to be their own nodes.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedshiftServerlessNamespace
metadata:
  name: awsredshiftserverlessnamespace-demo
spec:
  region: us-west-2
  dbName: analytics
  adminUsername: hackadmin
  manageAdminPassword: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.dbName` | `string` |  |  |  |
| `spec.adminUsername` | `string` |  |  |  |
| `spec.manageAdminPassword` | `bool` |  | `true` |  |
| `spec.adminUserPassword` | `string` (sensitive) |  |  |  |
| `spec.adminPasswordSecretKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.iamRoles` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.defaultIamRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.logExports` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the namespace is created in. Workgroups that attach
to this namespace must live in the same region.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.dbName

`string`

The name of the first database created in the namespace. Empty
keeps the AWS default ("dev"). Create-time only -- changing it
replaces the namespace (and the data in it); additional databases
are created with SQL, not here.

### spec.adminUsername

`string`

The admin username for the first database. Empty keeps the AWS
default ("admin"). Unlike the provisioned Redshift cluster, a
serverless namespace does not hard-require an admin user at create
time -- IAM identities can use temporary credentials
(GetCredentials) without one.

### spec.manageAdminPassword

`bool`

Let AWS manage the admin password in Secrets Manager: AWS generates
it, stores it, rotates it on schedule, and no secret ever touches
this manifest or the IaC state. The managed secret's ARN is
exported as the admin_password_secret_arn output. Mutually
exclusive with admin_user_password -- and the recommended posture.

- default: `true`

### spec.adminUserPassword

`string` · sensitive

The admin password, supplied directly (8-64 chars with at least one
uppercase letter, one lowercase letter, and one digit). Stored in
IaC state -- prefer manage_admin_password, which keeps the secret
in Secrets Manager entirely. Mutually exclusive with
manage_admin_password.

### spec.adminPasswordSecretKmsKeyId

`string | valueFrom`

The KMS key that encrypts the Secrets Manager secret holding the
managed admin password. Empty uses the AWS-managed
aws/secretsmanager key. Only meaningful with
manage_admin_password. Reference an AwsKmsKey key_arn output or
pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

The KMS key that encrypts the namespace's stored data. Empty uses
the AWS-owned Redshift service key. Reference an AwsKmsKey key_arn
output or pass a literal key ARN. Switching keys on a live
namespace is an in-place but long-running re-encryption.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.iamRoles

`[]string | valueFrom`

IAM roles the serverless engine assumes to access other AWS
services during COPY, UNLOAD, CREATE EXTERNAL FUNCTION, and
Redshift Spectrum queries (S3, DynamoDB, Glue, Lambda, ...).
Reference AwsIamRole role_arn outputs or pass literal role ARNs.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.defaultIamRoleArn

`string | valueFrom`

The IAM role assumed when a SQL command does not name one
explicitly (e.g. COPY ... IAM_ROLE default). Must also be present
in iam_roles -- AWS rejects a default role it has not been given.
Reference an AwsIamRole role_arn output or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.logExports

`[]string`

Which audit log types the namespace exports to CloudWatch Logs:
"connectionlog" (connection attempts), "useractivitylog" (every
executed query), "userlog" (user create/alter/drop events). Empty
exports nothing.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["connectionlog","useractivitylog","userlog"]}}}}

## Validation Rules

- `password_xor_managed`: admin_user_password cannot be set when manage_admin_password is true -- pick one password strategy
- `secret_kms_requires_managed`: admin_password_secret_kms_key_id is only meaningful with manage_admin_password -- it encrypts the AWS-managed secret, which exists only in managed mode

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRedshiftServerlessNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_name` | `string` | The namespace name. Exported because it is the join key workgroups attach with -- downstream references resolve against outputs, so the name must surface here even though it equals metadata.name. |
| `status.outputs.namespace_id` | `string` | The unique identifier AWS assigns to the namespace. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the namespace, for IAM policies, usage limits, and resource policies. |
| `status.outputs.db_name` | `string` | The name of the first database in the namespace. |
| `status.outputs.admin_password_secret_arn` | `string` | The ARN of the AWS-managed admin-password secret in Secrets Manager. Populated only when manage_admin_password is true -- the handle applications use to fetch credentials at runtime. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.adminPasswordSecretKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.iamRoles` | AwsIamRole | `status.outputs.role_arn` |
| `spec.defaultIamRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsRedshiftServerlessWorkgroup | `spec.namespaceName` | `status.outputs.namespace_name` |

## See Also

- [Overview](../README.md)

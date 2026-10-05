# AWS Redshift Serverless Namespace

Deploys an Amazon Redshift Serverless namespace — the data plane of the serverless warehouse: the first database, its admin credentials, the KMS key encrypting stored data, the IAM roles the engine assumes for COPY/UNLOAD/Spectrum, and the audit-log exports. A namespace stores; it never computes. Compute lives on [AwsRedshiftServerlessWorkgroup](/infra-catalog/aws-redshift-serverless-workgroup) nodes that attach to this namespace by name — many workgroups can serve one namespace, each created and destroyed without touching the data. The namespace defaults to the AWS-managed admin password strategy, so no secret ever touches the manifest or the IaC state.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Redshift Serverless Namespace** -- the single provider resource, and the data plane of the warehouse. Its name is the resource name (create-time immutable), and workgroups attach to it by that name. The namespace carries the first database (blank keeps the AWS default `dev`; additional databases are created with SQL), the admin credential posture (the AWS-managed password in Secrets Manager, or a supplied secret-reference password), the encryption binding (the AWS-owned Redshift service key or a referenced customer-managed KMS key), the engine IAM role associations (the roles COPY, UNLOAD, Spectrum, and external functions run under, plus the optional default role), and the audit-log exports to CloudWatch Logs
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Managed secret for a supplied password** -- only when you opt out of the AWS-managed strategy: create the password as an org secret first; the spec carries a `$secret/<slug>` reference and the runner resolves it just-in-time at deploy. The value must be 8-64 characters with an uppercase letter, a lowercase letter, and a digit.

### AWS Account

- **Engine IAM roles** -- roles referenced in `iamRoles` need a trust policy allowing `redshift.amazonaws.com` to assume them; the default role must also appear in the associated list (AWS rejects a default it has not been given).
- **Region choice is permanent** -- every workgroup that serves this data, and the subnets and security groups those workgroups reference, must live in the namespace's region.

## Deploy

### Console

Open the deployment store, find **AWS Redshift Serverless Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Managed-Password Namespace** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedshiftServerlessNamespace
metadata:
  name: analytics-data
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  dbName: analytics
  adminUsername: admin
  manageAdminPassword: true
  logExports:
    - connectionlog
```

```shell
planton apply -f redshift-serverless-namespace.yaml
```

This creates the data plane with an AWS-managed admin password and connection auditing. An Infra Job tracks the provisioning in real time.

### InfraChart

When the namespace deploys alongside its encryption key and engine roles in one chart, wire its own references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  dbName: analytics
  manageAdminPassword: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: warehouse-key
      fieldPath: status.outputs.key_arn
  iamRoles:
    - valueFrom:
        kind: AwsIamRole
        name: redshift-engine-role
        fieldPath: status.outputs.role_arn
  defaultIamRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: redshift-engine-role
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key and IAM role first, then provisions the namespace -- and the workgroups that serve it attach afterwards by referencing the `namespace_name` output.

## Key Configuration

These are the most important decisions when configuring a serverless namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The data/compute split** -- this resource is the half that stores. RPU capacity, VPC placement, endpoints, and query configuration all live on workgroups — so a capped dev workgroup and an autoscaling production workgroup can serve this same data, and destroying either never touches what is stored here.

**Managed password first** -- with `manageAdminPassword: true` (the recommended default), AWS generates the password, stores it in Secrets Manager, and rotates it on schedule; applications fetch credentials at runtime through the `admin_password_secret_arn` output. A supplied password is stored in IaC state — reserve it for external systems that must own the credential lifecycle. The two strategies are mutually exclusive, and unlike the provisioned cluster, a namespace may also run adminless: IAM identities can use temporary credentials (GetCredentials) without an admin user.

**Encryption ownership** -- empty `kmsKeyId` uses the AWS-owned Redshift service key; referencing an [AwsKmsKey](/infra-catalog/aws-kms-key) adds your rotation policy, access audit, and revocation kill switch. Switching keys on a live namespace is an in-place but long-running re-encryption.

**Engine roles and the default** -- COPY from S3, UNLOAD, Spectrum queries, and CREATE EXTERNAL FUNCTION run under a role from `iamRoles`. The `defaultIamRoleArn` answers SQL that says `IAM_ROLE default` and must also be one of the associated roles — the console offers it only from the declared list, so the rule cannot be violated there.

**Audit exports** -- `connectionlog` (authentication attempts), `useractivitylog` (every executed query), and `userlog` (user create/alter/drop). The user activity log is a two-halves pairing: the namespace exports the trail, but data is produced only while a serving workgroup sets `enable_user_activity_logging` to `true` in its config parameters.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId`, `adminPasswordSecretKmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `iamRoles`, `defaultIamRoleArn` | `status.outputs.role_arn` |

All references are optional -- the namespace works with none of them set.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace_name` | The namespace name (the resource name) | AwsRedshiftServerlessWorkgroup attachment (`namespaceName`) |
| `namespace_id` | The unique identifier AWS assigns | Account-level automation and audits |
| `arn` | Amazon Resource Name of the namespace | IAM policies, usage limits, resource policies |
| `db_name` | The first database's resolved name | Application connection strings |
| `admin_password_secret_arn` | The managed admin secret's ARN (managed mode only) | Applications fetching credentials at runtime |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed-password namespace** -- the starting point for nearly every serverless warehouse: a named first database and the AWS-managed admin password. Start from the **Managed-Password Namespace** preset.

**Production namespace with customer KMS and engine roles** -- stored data encrypted with your KMS key, COPY/Spectrum roles composed from the resource graph, and all three audit trails exported. Start from the **Production Namespace with Customer KMS and Engine Roles** preset.

## Works With

- [**AWS Redshift Serverless Workgroup**](/infra-catalog/aws-redshift-serverless-workgroup) -- the compute plane that serves this data (references `namespace_name`)
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed encryption for stored data and the managed admin secret
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the engine's COPY/UNLOAD/Spectrum identities
- [**AWS Redshift Cluster**](/infra-catalog/aws-redshift-cluster) -- the provisioned alternative when steady, predictable load makes reserved capacity cheaper

# Pulumi Module to Deploy AwsRedshiftServerlessNamespace

This Pulumi Go program deploys an Amazon Redshift Serverless namespace -- the
data plane of the serverless warehouse -- using the Planton API and module.
KMS keys and IAM roles attach by reference; compute attaches separately as
`AwsRedshiftServerlessWorkgroup` nodes. The module never creates a resource
that deserves to be its own node.

## Requirements

- Planton CLI built locally
- Valid AWS credential provided via the CLI IaC input (not in `spec`)

## CLI commands

Preview:

```shell
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Update (apply):

```shell
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

Destroy:

```shell
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Resources Created

1. **`redshiftserverless.Namespace`** — the data plane: first database, admin
   credential strategy, data-encryption key, engine IAM roles, audit log
   exports

## Outputs

| Key | Description |
|-----|-------------|
| `namespace_name` | The join key workgroups attach with |
| `namespace_id` | Unique identifier AWS assigns to the namespace |
| `arn` | ARN of the namespace |
| `db_name` | First database name |
| `admin_password_secret_arn` | Secrets Manager secret ARN (managed password only) |

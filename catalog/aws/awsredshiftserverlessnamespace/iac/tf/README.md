# Terraform Module to Deploy AwsRedshiftServerlessNamespace

This module provisions an Amazon Redshift Serverless namespace -- the data
plane of the serverless warehouse -- aligned with the Planton API. KMS keys
and IAM roles attach by reference; compute attaches separately as
`AwsRedshiftServerlessWorkgroup` nodes. The module never creates a resource
that deserves to be its own node.

## CLI (local backend)

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are passed via the IaC input through the CLI, not in `spec`.

## Files

- `variables.tf` (generated; do not edit)
- `provider.tf` — provider setup (`hashicorp/aws >= 6.0.0`)
- `locals.tf` — naming basis and identity tags
- `namespace.tf` — the `aws_redshiftserverless_namespace` resource
- `outputs.tf` — outputs matching `AwsRedshiftServerlessNamespaceOutputs`

## Outputs

| Name | Description |
|------|-------------|
| namespace_name | The join key workgroups attach with |
| namespace_id | Unique identifier AWS assigns to the namespace |
| arn | ARN of the namespace |
| db_name | First database name |
| admin_password_secret_arn | Secrets Manager secret ARN (managed password only) |

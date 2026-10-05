# Terraform Module to Deploy AwsElasticacheUserGroup

This module provisions an AWS ElastiCache RBAC user group — a membership
set that attaches to replication groups and serverless caches — aligned
with the Planton API.

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
- `provider.tf` — provider setup
- `locals.tf` — computed locals; `metadata.name` is the AWS user group id
- `main.tf` — `aws_elasticache_user_group`
- `outputs.tf` — outputs matching `AwsElasticacheUserGroupOutputs`

## Examples

See `../../presets/` for example manifests.

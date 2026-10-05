# Terraform Module to Deploy AwsElasticacheUser

This module provisions an AWS ElastiCache RBAC user — one Redis/Valkey
identity with an ACL access string and an authentication mode — aligned
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
- `locals.tf` — computed locals; `metadata.name` is the AWS user id
- `main.tf` — `aws_elasticache_user`
- `outputs.tf` — outputs matching `AwsElasticacheUserOutputs`

## Examples

See `../../presets/` for example manifests.

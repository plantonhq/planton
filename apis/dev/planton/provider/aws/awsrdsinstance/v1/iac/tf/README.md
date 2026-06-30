# Terraform Module to Deploy AwsRdsInstance

This module provisions a single AWS RDS DB instance aligned with the Planton API.

## CLI (local backend)

```shell
planton tofu init --manifest ../hack/manifest.yaml
planton tofu plan --manifest ../hack/manifest.yaml
planton tofu apply --manifest ../hack/manifest.yaml --auto-approve
planton tofu destroy --manifest ../hack/manifest.yaml --auto-approve
```

Credentials are passed via the stack input through the CLI, not in `spec`.

## Files
- `variables.tf` (generated; do not edit)
- `provider.tf` — provider setup
- `locals.tf` — computed locals and flags
- `subnet_group.tf` — DB subnet group when subnet IDs provided
- `instance.tf` — main DB instance resource
- `outputs.tf` — outputs matching `AwsRdsInstanceStackOutputs`

## Examples
See `../../examples.md` for example manifests.

# Terraform Module to Deploy AwsDocumentDb

This module provisions an Amazon DocumentDB (MongoDB-compatible) cluster and its instances aligned with the Planton API.

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
- `locals.tf` — computed locals and flags
- `subnet_group.tf` — DB subnet group when subnet IDs provided
- `cluster_param_group.tf` — managed cluster parameter group for inline parameters
- `docdb_cluster.tf` — main cluster resource
- `cluster_instances.tf` — per-name folded instances
- `outputs.tf` — outputs matching `AwsDocumentDbOutputs`

## Presets
See `../../presets/` for ready-to-adapt manifests.

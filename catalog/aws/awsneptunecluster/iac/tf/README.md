# Terraform Module to Deploy AwsNeptuneCluster

This module provisions an Amazon Neptune graph database cluster and its instances aligned with the Planton API.

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
- `subnet_group.tf` — Neptune subnet group when subnet IDs provided
- `cluster_param_group.tf` — managed cluster parameter group for inline
  parameters, plus the instance-level twin for inline instance_parameters
- `neptune_cluster.tf` — main cluster resource
- `cluster_instances.tf` — per-name folded instances (pinned to the
  cluster's port; apply_immediately and skip_final_snapshot forwarded)
- `custom_endpoints.tf` — per-name folded custom cluster endpoints
- `outputs.tf` — outputs matching `AwsNeptuneClusterOutputs`

## Presets
See `../../presets/` for ready-to-adapt manifests.

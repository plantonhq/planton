# Terraform Module to Deploy AwsRdsCluster

This module provisions an AWS RDS Cluster (Aurora MySQL/PostgreSQL or Multi-AZ DB Cluster) aligned with the Planton API.

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
- `security_group.tf` — optional managed SG and rules
- `subnet_group.tf` — DB subnet group when subnet IDs provided
- `cluster_param_group.tf` — optional cluster parameter group
- `rds_cluster.tf` — main cluster resource
- `outputs.tf` — outputs matching `AwsRdsClusterOutputs`

## Examples
See `../../e2e/manifest.yaml` for example manifests.



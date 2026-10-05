# Terraform Module to Deploy AwsRedshiftCluster

This module provisions an Amazon Redshift provisioned cluster aligned with the
Planton API: the cluster itself plus its folded satellites (subnet group,
parameter group, audit logging, cross-region snapshot copy). Security groups,
IAM roles, KMS keys, and Elastic IPs attach by reference -- the module never
creates them.

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
- `provider.tf` — provider setup (`hashicorp/aws >= 6.0.0`; audit logging and
  snapshot copy are standalone resources on the v6 line)
- `locals.tf` — computed locals and folding flags
- `redshift_cluster.tf` — the `aws_redshift_cluster` resource
- `subnet_group.tf` — Redshift subnet group when `subnet_ids` are provided
- `cluster_param_group.tf` — parameter group from inline `parameters`
- `logging.tf` — audit logging when `spec.logging` is set
- `snapshot_copy.tf` — cross-region snapshot copy when `spec.snapshot_copy` is set
- `usage_limits.tf` — per-feature consumption caps from `spec.usage_limits`
- `scheduled_actions.tf` — cron/at pause, resume, and resize actions
- `endpoint_access.tf` — managed VPC endpoints from `spec.endpoint_accesses`
- `endpoint_authorization.tf` — grantor-side cross-account access grants
- `snapshot_schedule.tf` — the singular snapshot-schedule association
- `outputs.tf` — outputs matching `AwsRedshiftClusterOutputs`

## Conditional Resources

- **Subnet group** — created from `spec.subnet_ids` (≥ 2 entries); or bring an
  existing group via `spec.cluster_subnet_group_name`
- **Parameter group** — created from inline `spec.parameters`; or bring an
  existing group via `spec.cluster_parameter_group_name`
- **Logging** — created when `spec.logging` is provided
- **Snapshot copy** — created when `spec.snapshot_copy` is provided

## Outputs

| Name | Description |
|------|-------------|
| cluster_identifier | Unique identifier of the Redshift cluster |
| cluster_arn | ARN of the cluster |
| cluster_namespace_arn | Namespace ARN for data sharing and the Data API |
| endpoint | Connection endpoint (address:port) |
| dns_name | DNS hostname (without port) |
| database_name | First database name (empty when the spec omitted it; AWS's default is "dev") |
| port | TCP port for connections |
| subnet_group_name | Subnet group name in use (managed or referenced) |
| parameter_group_name | Parameter group name in use (managed or referenced) |
| master_password_secret_arn | Secrets Manager secret ARN (managed password only) |
| endpoint_access_addresses | Managed VPC endpoint addresses, keyed by endpoint name |
| usage_limit_ids | AWS-generated usage-limit IDs, keyed by feature_type/limit_type/period |

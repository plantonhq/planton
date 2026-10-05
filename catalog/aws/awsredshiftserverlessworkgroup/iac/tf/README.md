# Terraform Module to Deploy AwsRedshiftServerlessWorkgroup

This module provisions an Amazon Redshift Serverless workgroup -- the compute
plane of the serverless warehouse -- aligned with the Planton API. The
namespace it serves, the subnets it places compute in, and the security groups
on its endpoint all attach by reference. The module never creates a resource
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
- `provider.tf` — provider setup (catalog-wide `hashicorp/aws` pin, plus
  `hashicorp/time` for the destroy-side settle)
- `locals.tf` — naming basis and identity tags
- `workgroup.tf` — the `aws_redshiftserverless_workgroup` resource
- `endpoint_access.tf` — managed VPC endpoints into consuming subnets
- `usage_limits.tf` — per-period consumption caps (RPU-hours / TB)
- `custom_domain.tf` — the one-per-workgroup custom domain association
- `satellite_settle.tf` — the destroy-side `time_sleep` between the
  usage-limit deletes and the endpoint delete (AWS serializes workgroup
  operations; the full live-probed contract is documented there)
- `outputs.tf` — outputs matching `AwsRedshiftServerlessWorkgroupOutputs`

## Outputs

| Name | Description |
|------|-------------|
| workgroup_name | The handle the serverless APIs and credentials API address |
| workgroup_id | Unique identifier AWS assigns to the workgroup |
| arn | ARN of the workgroup |
| endpoint_address | DNS hostname SQL clients connect to |
| port | TCP port for connections |
| endpoint_access_addresses | Private endpoint DNS addresses, keyed by endpoint name |
| usage_limit_ids | AWS-generated usage-limit IDs, keyed by usage_type/period |
| custom_domain_certificate_expiry_time | ACM certificate expiry (empty without a custom domain) |

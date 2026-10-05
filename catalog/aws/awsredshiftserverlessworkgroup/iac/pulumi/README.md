# Pulumi Module to Deploy AwsRedshiftServerlessWorkgroup

This Pulumi Go program deploys an Amazon Redshift Serverless workgroup -- the
compute plane of the serverless warehouse -- using the Planton API and module.
The namespace it serves, the subnets it places compute in, and the security
groups on its endpoint all attach by reference. The module never creates a
resource that deserves to be its own node.

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

1. **`redshiftserverless.Workgroup`** — the compute plane: RPU capacity (fixed
   baseline or price-performance target), VPC placement, reachability, config
   parameters, release track
2. **`redshiftserverless.CustomDomainAssociation`** — the one-per-workgroup
   branded TLS domain (rendered only when the spec sets it)
3. **`redshiftserverless.EndpointAccess`** — one managed VPC endpoint per
   `endpoint_accesses` entry, created straight after the workgroup (AWS
   serializes workgroup operations and rejects this create against a busy
   workgroup; on destroy each endpoint rides the workgroup's cascading
   delete via `DeletedWith` — the contract lives in
   `module/endpoint_access.go`)
4. **`redshiftserverless.UsageLimit`** — one per-period consumption cap per
   `usage_limits` entry, applied last (conflict-immune)

## Outputs

| Key | Description |
|-----|-------------|
| `workgroup_name` | The handle the serverless APIs and credentials API address |
| `workgroup_id` | Unique identifier AWS assigns to the workgroup |
| `arn` | ARN of the workgroup |
| `endpoint_address` | DNS hostname SQL clients connect to |
| `port` | TCP port for connections |
| `endpoint_access_addresses` | Private endpoint DNS addresses, keyed by endpoint name |
| `usage_limit_ids` | AWS-generated usage-limit IDs, keyed by usage_type/period |
| `custom_domain_certificate_expiry_time` | ACM certificate expiry (empty without a custom domain) |

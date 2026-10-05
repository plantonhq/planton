# Terraform Module to Deploy AwsRoute53HealthCheck

This module provisions an Amazon Route 53 health check — the availability probe DNS
records reference to keep unhealthy endpoints out of DNS answers.

## Features

- **All eight check types**: HTTP, HTTPS, HTTP_STR_MATCH, HTTPS_STR_MATCH, TCP,
  CALCULATED, CLOUDWATCH_METRIC, RECOVERY_CONTROL
- **Probe tuning**: request interval (10/30s), failure threshold, checker regions,
  latency graphs, SNI
- **State shaping**: inversion and administrative disable (the maintenance-window
  switch)
- **Calculated aggregation**: child health checks with a healthy-count threshold
- **CloudWatch mirroring**: the private-resource health-check pattern
- The spec's CEL rules enforce the per-type contract at authoring time, so this
  module maps fields 1:1 without re-validating

Generated `variables.tf` reflects the proto schema for `AwsRoute53HealthCheck`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For examples, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

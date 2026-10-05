# Terraform Module to Deploy AwsRoute53Zone

This module provisions an AWS Route53 hosted zone with support for multiple DNS record types and comprehensive domain management.
It includes configurable DNS records, TTL settings, and scalable DNS resolution for internet applications and internal services.

Generated `variables.tf` reflects the proto schema for `AwsRoute53Zone`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For more examples, see `../../e2e/manifest.yaml` and [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

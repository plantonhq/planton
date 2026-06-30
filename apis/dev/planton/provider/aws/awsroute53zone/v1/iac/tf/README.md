# Terraform Module to Deploy AwsRoute53Zone

This module provisions an AWS Route53 hosted zone with support for multiple DNS record types and comprehensive domain management.
It includes configurable DNS records, TTL settings, and scalable DNS resolution for internet applications and internal services.

Generated `variables.tf` reflects the proto schema for `AwsRoute53Zone`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).

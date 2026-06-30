# Terraform Module to Deploy AwsVpc

This module provisions AWS Virtual Private Clouds (VPCs) with support for multi-AZ subnet configuration, NAT gateways, internet gateways, and comprehensive DNS management.
It includes configurable CIDR blocks, availability zones, subnet sizing, and network infrastructure for secure and scalable AWS environments.

Generated `variables.tf` reflects the proto schema for `AwsVpc`.

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

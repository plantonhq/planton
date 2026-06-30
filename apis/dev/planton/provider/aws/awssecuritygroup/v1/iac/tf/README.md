# Terraform Module to Deploy AwsSecurityGroup

This module provisions AWS EC2 Security Groups with support for fine-grained ingress and egress rule management.
It includes configurable VPC integration, IPv4/IPv6 CIDR support, security group references, and comprehensive network security controls.

Generated `variables.tf` reflects the proto schema for `AwsSecurityGroup`.

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

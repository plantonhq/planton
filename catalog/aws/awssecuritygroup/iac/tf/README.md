# Terraform Module to Deploy AwsSecurityGroup

This module provisions AWS EC2 Security Groups with support for fine-grained ingress and egress rule management.
It includes configurable VPC integration, IPv4/IPv6 CIDR support, security group references, and comprehensive network security controls.

Generated `variables.tf` reflects the proto schema for `AwsSecurityGroup`.

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

# Terraform Module to Deploy AwsEc2Instance

This module provisions a single AWS EC2 virtual machine instance with networking, IAM, and access configuration.
It supports multiple connection methods including SSM, SSH via bastion, and EC2 Instance Connect.

Generated `variables.tf` reflects the proto schema for `AwsEc2Instance`.

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



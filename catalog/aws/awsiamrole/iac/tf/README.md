# Terraform Module to Deploy AwsIamRole

This module provisions an AWS IAM role with support for trust policies, managed policies, and inline policies.
It includes configurable role paths, descriptions, and comprehensive policy management capabilities.

Generated `variables.tf` reflects the proto schema for `AwsIamRole`.

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


# Terraform Module to Deploy AwsIamUser

This module provisions an AWS IAM user with support for managed policies, inline policies, and optional access key creation.
It includes configurable user names, policy attachments, and secure credential management for CI/CD and application use cases.

Generated `variables.tf` reflects the proto schema for `AwsIamUser`.

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


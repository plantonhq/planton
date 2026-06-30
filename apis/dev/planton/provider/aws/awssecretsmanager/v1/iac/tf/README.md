# Terraform Module to Deploy AwsSecretsManager

This module provisions AWS Secrets Manager resources with support for bulk secret creation and secure secret management.
It includes configurable secret names, placeholder values, and comprehensive secrets management capabilities for sensitive data storage.

Generated `variables.tf` reflects the proto schema for `AwsSecretsManager`.

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

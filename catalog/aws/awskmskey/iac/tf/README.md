# Terraform Module to Deploy AwsKmsKey

This module provisions an AWS KMS (Key Management Service) key with support for symmetric and asymmetric encryption, key rotation, and optional aliases.
It includes configurable key types, deletion windows, and comprehensive encryption management capabilities.

Generated `variables.tf` reflects the proto schema for `AwsKmsKey`.

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


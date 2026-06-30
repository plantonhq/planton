# Terraform Module to Deploy AwsEcrRepo

This module provisions an AWS ECR (Elastic Container Registry) repository for storing and managing Docker images.
It supports encryption, image tag mutability, automatic image scanning, and lifecycle policies with best practice defaults.

## Features

- **Encryption**: AES256 (default) or KMS encryption
- **Image Tag Mutability**: Configurable MUTABLE/IMMUTABLE tags
- **Image Scanning**: Automatic vulnerability scanning on push
- **Lifecycle Policies**: Automatic cleanup of untagged images
- **Force Delete**: Optional protection against accidental deletion
- **Best Practices**: Secure defaults aligned with production standards

Generated `variables.tf` reflects the proto schema for `AwsEcrRepo`.

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


# Terraform Module to Deploy AwsEcrRepo

This module provisions an AWS ECR (Elastic Container Registry) repository at the full
repository surface, plus its two folded 1:1 satellites: the lifecycle policy and the
repository policy.

## Features

- **Tag Mutability**: MUTABLE, IMMUTABLE, and the exclusion-filtered pair
  (IMMUTABLE_WITH_EXCLUSION / MUTABLE_WITH_EXCLUSION) with wildcard tag filters
- **Encryption**: AES256 (default), customer-managed KMS, or dual-layer KMS_DSSE
  (the whole configuration is create-time — changing it replaces the repository)
- **Image Scanning**: Automatic vulnerability scanning on push (default on)
- **Lifecycle Rules**: The full structured rule model (priorities, tag
  prefix/pattern selectors, count-by-age or count-by-number) rebuilt into the exact
  policy JSON AWS expects
- **Repository Policy**: Folded resource-based access policy (cross-account pulls,
  service principals)
- **Force Delete**: Optional protection against deleting a repository that still
  holds images

Generated `variables.tf` reflects the proto schema for `AwsEcrRepo`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a full-surface example, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

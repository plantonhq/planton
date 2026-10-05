# Pulumi Module: AWS ECR Repository

Provisions an Amazon ECR (Elastic Container Registry) repository using
Pulumi (Go), plus its two folded 1:1 satellites: the lifecycle policy and
the repository policy.

## Resources Created

- `ecr.Repository` — the repository itself: tag mutability (including the
  exclusion-filtered pair), encryption (AES256, customer KMS, or
  dual-layer KMS_DSSE — create-time-immutable), push-time vulnerability
  scanning, and force-delete protection.
- `ecr.LifecyclePolicy` — rendered from the spec's structured
  `lifecycle_rules` into the exact policy JSON AWS expects (priorities,
  tag prefix/pattern selectors, count-by-age or count-by-number). Created
  only when rules are declared.
- `ecr.RepositoryPolicy` — the folded resource-based access policy
  (cross-account pulls, service principals). Created only when a policy
  document is declared.

## How It Works

The module receives an `AwsEcrRepoIacInput` (the manifest plus provider
credentials), builds the AWS provider through the shared builder, and
renders the repository from the spec. Send conditions match the Terraform
module argument-for-argument: optional scalars pass through only when
set, so provider defaults (scan-on-push, AES256 encryption) apply
exactly as AWS documents them. The encryption configuration is
create-time — changing it replaces the repository.

Both satellites hang off the repository's own name, so they follow the
repository's lifecycle and never outlive it.

## Outputs

| Name | Description |
|------|-------------|
| `repository_name` | Name of the repository |
| `repository_url` | Full URL used in `docker push`/`pull` |
| `repository_arn` | ARN of the repository |
| `registry_id` | The account ID of the owning registry |

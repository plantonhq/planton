# Pulumi Module: AWS Secrets Manager Secret

Provisions an AWS Secrets Manager secret using Pulumi (Go).

## Resources Created

- `secretsmanager.Secret` — The named, KMS-encrypted container:
  description, key choice, cross-region replicas, deletion recovery
  window, and the managed-external-secret partner type.
- `secretsmanager.SecretPolicy` — The resource policy (only when
  `policy` is declared), rendered through the standalone resource so
  `BlockPublicPolicy` (default on) rejects policies granting anonymous
  access.
- `secretsmanager.SecretVersion` — The managed version (only when a
  value arm is set). Custom staging labels always ride WITH `AWSCURRENT`.
- `secretsmanager.SecretRotation` — Rotation (only when `rotation` is
  declared), ordered after the version via an explicit dependency so the
  immediate first rotation has a value to read.

## How It Works

The module receives an `AwsSecretsManagerSecretIacInput` (the manifest
plus provider credentials), builds the AWS provider through the shared
builder, and renders the secret and its satellites from the spec. Send
conditions match the Terraform module argument-for-argument: the policy
Struct marshals to the same JSON, `force_overwrite_replica_secret` and
the materialized defaults (`recovery_window_in_days`,
`block_public_policy`, `rotate_immediately`) are always sent explicitly,
and the `type` partner identifier is omitted entirely when empty (AWS
treats absent and empty differently in error paths).

## Outputs

| Name | Description |
|------|-------------|
| `secret_arn` | ARN of the secret (carries AWS's random 6-character suffix — never derivable from the name) |
| `secret_name` | Name of the secret (matches metadata.name) |
| `version_id` | Version ID of the managed version (empty for a shell secret) |

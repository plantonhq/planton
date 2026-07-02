# S3 Read-Only Access

This preset creates a managed policy granting read-only access to a single S3
bucket -- the most common shared permission set in a typical AWS estate. Attach
it to any role or user through their `managedPolicyArns` field (via `valueFrom`
referencing this policy's `policy_arn` output).

## When to Use

- Application roles that consume data from a bucket without writing to it
- CI or analytics principals that need to read build artifacts or datasets
- Any case where several principals share the same read-only grant

## Key Configuration Choices

- **Object + bucket permissions** -- `s3:GetObject`/`s3:GetObjectVersion` on the
  objects and `s3:ListBucket` on the bucket itself; listing requires the
  bucket-level resource, which is why both Resource entries are present
- **No write or delete actions** -- attach a separate, deliberately-scoped
  policy for writers rather than widening this one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<bucket-name>` | Name of the S3 bucket to grant read access to | S3 console or `AwsS3Bucket` status outputs |

## Related Presets

- **02-permissions-boundary** -- a ceiling policy for delegated principal creation

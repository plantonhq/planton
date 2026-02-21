---
title: "Scoped OSS Bucket Access"
description: "This preset creates a custom RAM policy that grants read/write access to a single OSS bucket and its objects. System policies like `AliyunOSSFullAccess` grant access to every bucket in the account --..."
type: "preset"
rank: "01"
presetSlug: "01-scoped-oss-access"
componentSlug: "ram-policy"
componentTitle: "RAM Policy"
provider: "alicloud"
icon: "package"
order: 1
---

# Scoped OSS Bucket Access

This preset creates a custom RAM policy that grants read/write access to a single OSS bucket and its objects. System policies like `AliyunOSSFullAccess` grant access to every bucket in the account -- this preset scopes permissions down to exactly one bucket, following the principle of least privilege.

## When to Use

- Application roles that need to read/write objects in a specific OSS bucket without access to other buckets
- Production environments where system-wide OSS policies are too permissive
- Microservice architectures where each service has its own data bucket and should not access other services' storage
- Attaching to an `AliCloudRamRole` via `policyAttachments` with `policyType: Custom`

## Key Configuration Choices

- **Bucket-scoped resource ARNs** (`acs:oss:*:*:<bucket>` and `acs:oss:*:*:<bucket>/*`) -- Both the bucket itself and all objects within it are included. The bucket-level ARN is needed for `ListObjects` and `GetBucket`; the object-level wildcard ARN is needed for `GetObject`, `PutObject`, and `DeleteObject`.
- **Multipart upload actions** (`ListMultipartUploads`, `AbortMultipartUpload`) -- Included because large file uploads use multipart transfers. Without these actions, incomplete uploads cannot be listed or cleaned up, leading to storage cost leaks.
- **Automatic version rotation** (`rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded`) -- Bucket-scoped policies tend to be updated as new actions or path restrictions are added. Automatic rotation prevents hitting the 5-version limit and failing on update.
- **No ListBuckets action** -- Intentionally omitted at the account level. The role can interact with its designated bucket but cannot enumerate other buckets in the account.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-policy-name>` | RAM policy name, unique per account (1-128 chars: letters, digits, hyphens) | Choose a name following your naming convention (e.g., `app-data-bucket-rw`) |
| `<your-bucket-name>` | The exact OSS bucket name to grant access to | OSS console or `AliCloudStorageBucket` stack outputs |

## Related Presets

- **02-cicd-deploy-pipeline** -- Use instead when you need cross-service permissions for a CI/CD workflow rather than single-service bucket access

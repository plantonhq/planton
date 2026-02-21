# AliCloudRamPolicy

Manages an Alibaba Cloud Resource Access Management (RAM) custom policy.

## Overview

A RAM policy is a JSON document that defines a set of permissions -- which actions are allowed or denied on which resources, optionally under which conditions. Alibaba Cloud provides hundreds of system-managed policies for common scenarios, but when you need fine-grained access control beyond what system policies offer, you create a custom policy.

### What Gets Created

- **RAM Policy** -- a custom IAM policy with the JSON document you provide

### Important: RAM is a Global Service

RAM policies are account-global and not region-scoped. The `region` field configures the provider endpoint only.

### Policy Versioning

Alibaba Cloud maintains up to 5 versions per policy. Each update creates a new version and sets it as the default. When the limit is reached, the `rotateStrategy` field controls whether the oldest non-default version is automatically deleted or the update fails.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider endpoint (e.g., `cn-hangzhou`) |
| `policyName` | string | Policy name, unique per account (1-128 chars) |
| `policyDocument` | string | JSON IAM policy document (max 6144 bytes) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description (max 1024 chars) |
| `rotateStrategy` | string | `"None"` | Version rotation: `None` or `DeleteOldestNonDefaultVersionWhenLimitExceeded` |
| `tags` | map | `{}` | Key-value tags for the policy |
| `force` | bool | `false` | Force-detach from all entities before deletion |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `policy_name` | The policy name as created |
| `policy_type` | Always `Custom` for user-created policies |

## Policy Document Structure

Alibaba Cloud RAM policy documents follow this structure:

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["<service>:<action>"],
      "Resource": ["<resource-arn>"],
      "Condition": {}
    }
  ]
}
```

- **Version**: Always `"1"` for Alibaba Cloud RAM policies.
- **Effect**: `"Allow"` or `"Deny"`.
- **Action**: Service-specific actions like `oss:GetObject`, `ecs:DescribeInstances`, `rds:*`.
- **Resource**: Alibaba Cloud Resource Names (ARNs) in format `acs:<service>:<region>:<account-id>:<resource>`.
- **Condition** (optional): Context-based restrictions (IP ranges, time windows, etc.).

## Related Components

- **AliCloudRamRole** -- attach this policy to a role via `policyAttachments`
- **AliCloudStorageBucket** -- common target for fine-grained OSS access policies
- **AliCloudRdsInstance** -- common target for database access policies

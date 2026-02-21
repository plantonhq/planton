# AliCloudRamRole

Manages an Alibaba Cloud Resource Access Management (RAM) role with bundled policy attachments.

## Overview

A RAM role is a virtual identity without permanent credentials. Trusted entities -- services, accounts, or federated identities -- assume the role via STS (Security Token Service) to obtain temporary security tokens. This is the standard mechanism for granting Alibaba Cloud services the permissions they need to operate on your behalf.

### What Gets Created

- **RAM Role** -- the identity with a trust policy defining who can assume it
- **Policy Attachments** (optional) -- system-managed or custom policies granting the role permissions

### Important: RAM is a Global Service

RAM roles are account-global and not region-scoped. The `region` field configures the provider endpoint only.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider endpoint (e.g., `cn-hangzhou`) |
| `roleName` | string | RAM role name, unique per account (1-64 chars) |
| `assumeRolePolicyDocument` | string | JSON trust policy defining who can assume this role |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description of the role |
| `maxSessionDuration` | int | `3600` | Maximum STS session duration in seconds (3600-43200) |
| `tags` | map | `{}` | Key-value tags for the role |
| `force` | bool | `false` | Force-detach policies before deletion |
| `policyAttachments` | list | `[]` | Policies to attach to the role |

### Policy Attachment Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `policyName` | string | (required) | Policy name (e.g., `AliyunECSFullAccess`) |
| `policyType` | string | `System` | `System` for managed policies, `Custom` for user-created |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `role_id` | The RAM role ID |
| `role_name` | The RAM role name |
| `arn` | The role ARN (`acs:ram::<account-id>:role/<role-name>`) |

## Common Trust Policy Patterns

### Allow an Alibaba Cloud Service

```json
{
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {"Service": ["ecs.aliyuncs.com"]}
  }],
  "Version": "1"
}
```

### Allow Cross-Account Assumption

```json
{
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {"RAM": ["acs:ram::OTHER_ACCOUNT_ID:root"]}
  }],
  "Version": "1"
}
```

## Related Components

- **AliCloudRamPolicy** -- create custom policies to attach to this role
- **AliCloudAckManagedCluster** -- references this role for cluster service authentication
- **AliCloudFcFunction** -- references this role for function execution permissions
- **AliCloudEcsInstance** -- references this role for instance profile

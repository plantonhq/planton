---
title: "RAM Policy"
description: "RAM Policy deployment documentation"
icon: "package"
order: 100
componentName: "alicloudrampolicy"
---

# AliCloud RAM Policy

Deploys an Alibaba Cloud RAM custom policy with a JSON permission document, optional version rotation strategy, and tag management. Custom policies define fine-grained permissions beyond what system-managed policies provide and can be attached to RAM roles via AliCloudRamRole.

## What Gets Created

When you deploy an AliCloudRamPolicy resource, OpenMCF provisions:

- **RAM Policy** — an `alicloud_ram_policy` resource with the specified JSON policy document, version rotation strategy, and tags

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An Alibaba Cloud account** with RAM service enabled
- **A valid JSON policy document** conforming to the Alibaba Cloud RAM policy structure (Version, Statement, Effect, Action, Resource)

## Quick Start

Create a file `ram-policy.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: my-oss-reader
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudRamPolicy.my-oss-reader
spec:
  region: cn-hangzhou
  policyName: oss-read-only
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": ["oss:GetObject", "oss:ListObjects"],
          "Resource": ["acs:oss:*:*:my-bucket/*"]
        }
      ]
    }
```

Deploy:

```shell
openmcf apply -f ram-policy.yaml
```

This creates a custom RAM policy granting read-only access to a specific OSS bucket.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider endpoint configuration. RAM is a global service, but the provider requires a region for API routing (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `policyName` | `string` | RAM policy name, unique within the Alibaba Cloud account. English letters, digits, and hyphens only. | Required; 1-128 characters |
| `policyDocument` | `string` | JSON IAM policy document defining permissions. Must conform to the Alibaba Cloud RAM policy structure with Version, Statement, Effect, Action, and Resource fields. | Required; non-empty; max 6144 bytes |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of what the policy allows or denies. Maximum 1024 characters. |
| `rotateStrategy` | `string` | `"None"` | Strategy for handling the 5-version limit. `None`: update fails at the limit. `DeleteOldestNonDefaultVersionWhenLimitExceeded`: auto-deletes the oldest non-default version. |
| `tags` | `map<string, string>` | `{}` | Tags applied to the policy. Merged with standard OpenMCF tags (`resource_name`, `resource_kind`, `organization`, `environment`). User tags take precedence on conflict. |
| `force` | `bool` | `false` | Force-delete the policy even if attached to roles, users, or groups. When `true`, detaches from all entities and deletes all non-default versions before deletion. |

## Examples

### Read-Only OSS Access

A minimal policy granting read-only access to all OSS buckets:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: oss-reader
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudRamPolicy.oss-reader
spec:
  region: cn-hangzhou
  policyName: oss-read-only
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "oss:GetObject",
            "oss:GetBucket",
            "oss:ListObjects",
            "oss:ListBuckets"
          ],
          "Resource": ["acs:oss:*:*:*"]
        }
      ]
    }
```

### Scoped Bucket Access with Version Rotation

A policy granting full access to a specific OSS bucket with automatic version rotation for frequent updates:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: app-data-access
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudRamPolicy.app-data-access
spec:
  region: cn-shanghai
  policyName: app-data-bucket-full-access
  description: Grants full access to the application data bucket and its objects
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": ["oss:*"],
          "Resource": [
            "acs:oss:*:*:app-data-prod",
            "acs:oss:*:*:app-data-prod/*"
          ]
        }
      ]
    }
  rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded
  tags:
    team: platform
    costCenter: infrastructure
```

### Multi-Service CI/CD Pipeline Policy

A cross-service policy for CI/CD pipelines with force delete enabled for clean teardowns:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: cicd-deploy-policy
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudRamPolicy.cicd-deploy-policy
spec:
  region: cn-hangzhou
  policyName: cicd-pipeline-deploy-policy
  description: Permissions for CI/CD pipeline to build images, deploy to ACK, and manage logs
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "cr:GetRepository",
            "cr:PushRepository",
            "cr:PullRepository"
          ],
          "Resource": ["acs:cr:*:*:repository/my-org/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "cs:DescribeClusterDetail",
            "cs:GetClusterKubeconfig",
            "cs:DescribeClusterNodes"
          ],
          "Resource": ["acs:cs:*:*:cluster/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "log:PostLogStoreLogs",
            "log:GetLogStore"
          ],
          "Resource": ["acs:log:*:*:project/cicd-logs/*"]
        }
      ]
    }
  rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded
  force: true
  tags:
    purpose: cicd
    managedBy: platform-team
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `policy_name` | `string` | The RAM policy name as created |
| `policy_type` | `string` | Always `Custom` for user-created policies. Used by AliCloudRamRole `policyAttachments` which require both policy name and type for attachment. |

## Related Components

- [AliCloudRamRole](/docs/catalog/alicloud/ram-role) — attach this policy to a role via `policyAttachments` with `policyType: Custom`
- [AliCloudStorageBucket](/docs/catalog/alicloud/oss-bucket) — common target for fine-grained OSS access policies
- [AliCloudRdsInstance](/docs/catalog/alicloud/rds-instance) — common target for database access policies

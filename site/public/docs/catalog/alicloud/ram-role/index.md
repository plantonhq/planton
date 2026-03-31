---
title: "RAM Role"
description: "RAM Role deployment documentation"
icon: "package"
order: 100
componentName: "alicloudramrole"
---

# AliCloud RAM Role

Deploys an Alibaba Cloud RAM role with bundled policy attachments and a configurable trust policy document. The component provisions the role and its policy attachments as a single atomic unit, ensuring the role is always created with its intended permissions.

## What Gets Created

When you deploy an AliCloudRamRole resource, OpenMCF provisions:

- **RAM Role** — an `alicloud_ram_role` resource with the specified trust policy, session duration, and tags
- **Policy Attachments** — one `alicloud_ram_role_policy_attachment` per entry in `policyAttachments`, granting the role permissions defined by system-managed or custom policies

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An Alibaba Cloud account** with RAM service enabled
- **Custom policies** must already exist before referencing them with `policyType: Custom` — create them with AliCloudRamPolicy

## Quick Start

Create a file `ram-role.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: my-ecs-role
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudRamRole.my-ecs-role
spec:
  region: cn-hangzhou
  roleName: my-ecs-service-role
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["ecs.aliyuncs.com"]}
      }],
      "Version": "1"
    }
```

Deploy:

```shell
openmcf apply -f ram-role.yaml
```

This creates a RAM role that ECS instances can assume via STS. No policies are attached in this minimal configuration.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider endpoint configuration. RAM is a global service, but the provider requires a region for API routing (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `roleName` | `string` | RAM role name, unique within the Alibaba Cloud account. Letters, digits, periods, hyphens, and underscores only. | Required; 1-64 characters |
| `assumeRolePolicyDocument` | `string` | JSON trust policy document defining which principals can assume this role via STS. | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the role's purpose. |
| `maxSessionDuration` | `int` | `3600` | Maximum STS session duration in seconds when assuming this role. Range: 3600-43200 (1 hour to 12 hours). |
| `tags` | `map<string, string>` | `{}` | Tags applied to the RAM role. Merged with standard OpenMCF tags (`resource_name`, `resource_kind`, `organization`, `environment`). User tags take precedence on conflict. |
| `force` | `bool` | `false` | Force-detach all attached policies before deleting the role. When `false`, deletion fails if policies are still attached. |
| `policyAttachments` | `list` | `[]` | Policies to attach to this role. Each entry creates a policy attachment resource. |
| `policyAttachments[].policyName` | `string` | — | Policy name to attach (e.g., `AliyunECSFullAccess`, `AliyunOSSReadOnlyAccess`). Required per attachment. |
| `policyAttachments[].policyType` | `string` | `System` | `System` for Alibaba Cloud managed policies, `Custom` for user-created policies. |

## Examples

### ECS Service Role with System Policies

A role for ECS instances that need access to OSS and log services, with a 2-hour session duration:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: ecs-worker-role
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudRamRole.ecs-worker-role
spec:
  region: cn-shanghai
  roleName: ecs-worker-role
  description: Role for ECS worker instances accessing OSS and monitoring
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["ecs.aliyuncs.com"]}
      }],
      "Version": "1"
    }
  maxSessionDuration: 7200
  tags:
    team: platform
    costCenter: infrastructure
  policyAttachments:
    - policyName: AliyunOSSFullAccess
    - policyName: AliyunCloudMonitorFullAccess
    - policyName: AliyunLogFullAccess
```

### Cross-Account Audit Role

A role that another Alibaba Cloud account can assume for read-only audit access, with both system and custom policies:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: cross-account-audit
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudRamRole.cross-account-audit
spec:
  region: cn-hangzhou
  roleName: cross-account-audit-role
  description: Allows the security audit account to read billing and logs
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"RAM": ["acs:ram::1234567890123456:root"]}
      }],
      "Version": "1"
    }
  maxSessionDuration: 43200
  force: true
  tags:
    purpose: security-audit
  policyAttachments:
    - policyName: AliyunBSSReadOnlyAccess
    - policyName: AliyunLogReadOnlyAccess
    - policyName: audit-log-reader-policy
      policyType: Custom
```

### Function Compute Execution Role

A role for Function Compute functions that need VPC access and log service integration:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: fc-execution-role
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AliCloudRamRole.fc-execution-role
spec:
  region: cn-hangzhou
  roleName: fc-execution-role
  description: Execution role for Function Compute with VPC and logging access
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["fc.aliyuncs.com"]}
      }],
      "Version": "1"
    }
  policyAttachments:
    - policyName: AliyunVPCFullAccess
    - policyName: AliyunECSNetworkInterfaceManagement
    - policyName: AliyunLogFullAccess
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `role_id` | `string` | The RAM role ID assigned by Alibaba Cloud |
| `role_name` | `string` | The RAM role name as created |
| `arn` | `string` | The Alibaba Cloud Resource Name for the role (format: `acs:ram::<account-id>:role/<role-name>`) |

## Related Components

- [AliCloudRamPolicy](/docs/catalog/alicloud/ram-policy) — create custom policies to attach to this role
- [AliCloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) — references this role for cluster service authentication
- [AliCloudFcFunction](/docs/catalog/alicloud/alicloudfcfunction) — references the role ARN for function execution permissions
- [AliCloudEcsInstance](/docs/catalog/alicloud/ecsinstance) — references this role for instance profile attachment

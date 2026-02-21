---
title: "ECS Service Role"
description: "This preset creates a RAM role that ECS instances can assume to access common Alibaba Cloud services: OSS for object storage, CloudMonitor for metrics and alerting, and Log Service (SLS) for..."
type: "preset"
rank: "01"
presetSlug: "01-ecs-service-role"
componentSlug: "ram-role"
componentTitle: "RAM Role"
provider: "alicloud"
icon: "package"
order: 1
---

# ECS Service Role

This preset creates a RAM role that ECS instances can assume to access common Alibaba Cloud services: OSS for object storage, CloudMonitor for metrics and alerting, and Log Service (SLS) for centralized logging. Since ACK worker nodes are ECS instances, this preset also serves as a starting point for Kubernetes node roles -- add container registry and cluster-specific policies as needed.

## When to Use

- ECS instances that need to read/write objects in OSS buckets
- Workloads reporting custom metrics or requiring CloudMonitor integration
- Applications writing structured logs to SLS via the Logtail agent
- Starting point for ACK worker node roles (add `AliyunCRReadOnlyAccess` for image pulling)

## Key Configuration Choices

- **ECS service principal** (`ecs.aliyuncs.com`) -- Only ECS instances can assume this role via instance RAM role binding
- **OSS full access** (`AliyunOSSFullAccess`) -- Grants read/write to all buckets; scope down to a custom policy with bucket-level restrictions for production
- **CloudMonitor full access** (`AliyunCloudMonitorFullAccess`) -- Enables custom metric reporting and alarm management
- **Log Service full access** (`AliyunLogFullAccess`) -- Allows the Logtail agent and application SDKs to write logs to any SLS project

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-ecs-role-name>` | RAM role name, unique per Alibaba Cloud account (1-64 chars: letters, digits, `.`, `-`, `_`) | Choose a name following your organization's naming convention |

## Related Presets

- **02-fc-execution-role** -- Use instead for Function Compute execution roles
- **03-cross-account-audit** -- Use instead for cross-account security audit access

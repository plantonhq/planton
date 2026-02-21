---
title: "Event Handler Function"
description: "This preset creates a lightweight Python 3.12 function for event-driven workloads. The function code is deployed from an OSS bucket, sized for modest compute needs (0.5 vCPU, 256 MB), and logs..."
type: "preset"
rank: "01"
presetSlug: "01-event-handler"
componentSlug: "function"
componentTitle: "Function"
provider: "alicloud"
icon: "package"
order: 1
---

# Event Handler Function

This preset creates a lightweight Python 3.12 function for event-driven workloads. The function code is deployed from an OSS bucket, sized for modest compute needs (0.5 vCPU, 256 MB), and logs invocation metrics to SLS. This is the simplest starting point for serverless event processing on Function Compute v3.

## When to Use

- Event-driven processing triggered by OSS uploads, message queues, or API Gateway events
- Lightweight data transformations, notifications, or webhook handlers
- Functions that do not need access to VPC-internal resources (databases, caches)
- Getting started with Function Compute v3

## Key Configuration Choices

- **Python 3.12** (`runtime: python3.12`) -- Latest stable Python runtime on FC v3. Change to `nodejs20`, `java11`, or `go1` for other language preferences.
- **Minimal compute** (`cpu: 0.5`, `memorySize: 256`) -- Appropriate for event handlers that process small payloads (JSON events, notifications). Scale up for CPU-intensive or memory-heavy workloads.
- **60-second timeout** (`timeout: 60`) -- Generous for most event processing. Reduce for latency-sensitive triggers or increase (up to 86400s) for long-running batch operations.
- **OSS code deployment** (`code.ossBucketName`, `code.ossObjectName`) -- The function code is a ZIP package stored in OSS. This is the standard deployment model for built-in runtimes. Update the object key to deploy new versions.
- **SLS logging with request metrics** (`logConfig`) -- Captures function invocation logs and per-request latency/status metrics in SLS. Instance-level metrics are omitted to reduce noise for lightweight functions.
- **No VPC** -- Event handlers typically process data without needing private network access. Add `vpcConfig` if the function needs to reach databases or caches inside a VPC.
- **No execution role** -- Functions that only process events and write logs may not need a RAM role. Add `role` if the function needs to call other Alibaba Cloud APIs (e.g., writing to OSS, publishing to MNS).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region |
| `<your-function-name>` | Function name (1-128 chars, unique per region) | Your naming convention |
| `<your-code-bucket>` | OSS bucket containing the function code ZIP | `AliCloudStorageBucket` stack outputs |
| `<your-code-object-key>` | Object key for the ZIP package (e.g., `functions/handler-v1.zip`) | Your CI/CD pipeline |
| `<your-log-project-name>` | SLS project for function logs | `AliCloudLogProject` stack outputs |
| `<your-logstore-name>` | SLS logstore within the project | Your SLS project configuration |

## Related Presets

- **02-vpc-api-function** -- Use when the function needs VPC access to private resources and higher compute for API traffic
- **03-custom-container** -- Use when the function runs a custom Docker image instead of a built-in runtime

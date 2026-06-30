# AliCloudFunction

Manages an Alibaba Cloud Function Compute v3 function.

## Overview

Function Compute (FC) is a fully managed, event-driven serverless compute service. FC v3 uses a service-less model where functions are top-level resources -- there is no wrapping service layer. VPC access, logging, IAM role, and all other configuration is set directly on the function.

This component wraps a single `alicloud_fcv3_function` resource. Triggers, aliases, versions, and concurrency configs have independent lifecycles and are managed separately.

### What Gets Created

- **Function** -- an FC v3 function in the specified region with configurable runtime, compute sizing, networking, logging, and storage mounts

### Supported Runtimes

| Family | Runtimes |
|--------|----------|
| Python | `python3.12`, `python3.10`, `python3.9`, `python3` |
| Node.js | `nodejs20`, `nodejs18`, `nodejs16`, `nodejs14` |
| Java | `java11`, `java8` |
| Go | `go1` |
| PHP | `php7.2` |
| .NET | `dotnetcore3.1` |
| Custom | `custom`, `custom.debian10`, `custom.debian11`, `custom.debian12` |
| Container | `custom-container` (requires `customContainerConfig`) |

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`) |
| `functionName` | string | Function name, 1-128 characters. Immutable after creation |
| `handler` | string | Entry point (e.g., `index.handler`, `main`) |
| `runtime` | string | Runtime environment (see table above) |

### Compute Sizing (optional, provider computes defaults)

| Field | Type | Description |
|-------|------|-------------|
| `cpu` | double | vCPU allocation (0.05-16) |
| `memorySize` | int32 | Memory in MB (64-32768) |
| `timeout` | int32 | Max execution time in seconds (1-86400) |
| `diskSize` | int32 | Temp disk in MB (min 512) |
| `instanceConcurrency` | int32 | Concurrent requests per instance (1-200) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description |
| `role` | StringValueOrRef | | RAM role ARN for execution (references AliCloudRamRole) |
| `internetAccess` | bool | | Whether function can access the public internet |
| `layers` | list | `[]` | Layer ARNs (max 5) |
| `environmentVariables` | map | `{}` | Environment variables |
| `tags` | map | `{}` | Tags merged with standard Planton tags |
| `resourceGroupId` | string | `""` | Resource group ID (per DD05) |

### Nested Configurations

- **code** -- OSS bucket/object or base64 ZIP for function code
- **vpcConfig** -- VPC, VSwitch, and security group for private network access
- **logConfig** -- SLS project and logstore for invocation logging
- **customContainerConfig** -- Container image, entrypoint, health check
- **customRuntimeConfig** -- Bootstrap command, args, port, health check
- **instanceLifecycleConfig** -- Initializer and pre-stop hooks
- **nasConfig** -- NAS file system mount points
- **gpuConfig** -- GPU type and memory for AI/ML workloads

## Stack Outputs

| Output | Description |
|--------|-------------|
| `function_id` | The FC function ID assigned by Alibaba Cloud |
| `function_name` | The function name |
| `function_arn` | The function ARN for use in IAM policies and trigger configurations |

## Related Components

- **AliCloudRamRole** -- provides the execution role for the function
- **AliCloudLogProject** -- provides the SLS project for function logging
- **AliCloudVpc** / **AliCloudVswitch** / **AliCloudSecurityGroup** -- VPC networking for private resource access
- **AliCloudNasFileSystem** -- NAS mount target for shared file storage
- **AliCloudStorageBucket** -- OSS bucket for function code packages

# AliCloud Function

Deploys an Alibaba Cloud Function Compute v3 function. The component provisions a single `alicloud_fcv3_function` resource with configurable runtime, compute sizing, VPC networking, SLS logging, custom container/runtime settings, lifecycle hooks, NAS mounts, and GPU acceleration. FC v3 uses a service-less model where functions are top-level resources — VPC access, logging, IAM role, and all other configuration is set directly on the function. Triggers, aliases, versions, and concurrency configs have independent lifecycles and are managed by separate components.

## What Gets Created

When you deploy an AliCloudFunction resource, OpenMCF provisions:

- **FC v3 Function** — a single function in the specified region with the configured runtime, handler, compute sizing, and optional networking/logging/storage settings
- **Standard tags** — `resource`, `resource_name`, `resource_kind`, plus optional `organization` and `environment` tags, merged with user-provided `tags`

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`) or OpenMCF provider config
- **A code package** — either an OSS bucket/object containing a ZIP file, an inline base64-encoded ZIP, or (for `custom-container` runtime) a container image accessible from FC
- **RAM role** (if the function accesses other Alibaba Cloud services) — the role must trust the FC service principal (`fc.aliyuncs.com`)
- **VPC resources** (if the function needs private network access) — VPC, VSwitch(es), and security group must exist
- **SLS project and logstore** (if logging is configured) — must exist in the same region

## Quick Start

Create a file `function.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: hello-world
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudFunction.hello-world
spec:
  region: cn-hangzhou
  functionName: hello-world
  handler: index.handler
  runtime: python3.12
  code:
    ossBucketName: my-code-bucket
    ossObjectName: functions/hello-world.zip
```

Deploy:

```shell
openmcf apply -f function.yaml
```

This creates an FC v3 function named `hello-world` in `cn-hangzhou` running `python3.12` with code from the specified OSS location.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region where the function will be created (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `functionName` | `string` | Function name. Immutable after creation (ForceNew). | Required; 1-128 characters |
| `handler` | `string` | Entry point for invocation (e.g., `index.handler`, `com.example.Main::handleRequest`, `main`). | Required; non-empty |
| `runtime` | `string` | Runtime environment. One of: `python3.12`, `python3.10`, `python3.9`, `python3`, `nodejs20`, `nodejs18`, `nodejs16`, `nodejs14`, `java11`, `java8`, `go1`, `php7.2`, `dotnetcore3.1`, `custom`, `custom.debian10`, `custom.debian11`, `custom.debian12`, `custom-container`. | Required; must be in allowed set |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the function. |
| `cpu` | `double` | Provider default | vCPU allocation per instance. Range: 0.05-16. |
| `memorySize` | `int32` | Provider default | Memory in MB per instance. Range: 64-32768. |
| `timeout` | `int32` | Provider default | Max execution time in seconds. Range: 1-86400. |
| `diskSize` | `int32` | Provider default | Temp disk in MB. Minimum: 512. |
| `instanceConcurrency` | `int32` | Provider default | Concurrent requests per instance. Range: 1-200. |
| `code` | `object` | — | Code package. Fields: `ossBucketName`, `ossObjectName`, `zipFile`, `checksum`. Not required for `custom-container`. |
| `role` | `StringValueOrRef` | — | RAM role ARN for execution. References `AliCloudRamRole`. |
| `internetAccess` | `bool` | — | Whether the function can access the public internet. |
| `vpcConfig` | `object` | — | VPC attachment. Fields: `vpcId` (ref: AliCloudVpc), `vswitchIds`, `securityGroupId` (ref: AliCloudSecurityGroup). |
| `logConfig` | `object` | — | SLS logging. Fields: `project` (ref: AliCloudLogProject), `logstore`, `logBeginRule`, `enableInstanceMetrics`, `enableRequestMetrics`. |
| `customContainerConfig` | `object` | — | Container runtime config. Fields: `image`, `entrypoint`, `command`, `port`, `healthCheckConfig`. |
| `customRuntimeConfig` | `object` | — | Custom runtime config. Fields: `command`, `args`, `port`, `healthCheckConfig`. |
| `instanceLifecycleConfig` | `object` | — | Lifecycle hooks. Fields: `initializer` (handler, timeout, command), `preStop` (handler, timeout). |
| `nasConfig` | `object` | — | NAS mounts. Fields: `userId`, `groupId`, `mountPoints[]` (serverAddr, mountDir, enableTls). Requires `vpcConfig`. |
| `gpuConfig` | `object` | — | GPU acceleration. Fields: `gpuMemorySize` (int, >0), `gpuType` (one of: `fc.gpu.tesla.1`, `fc.gpu.ampere.1`, `fc.gpu.ada.1`, `g1`). |
| `layers` | `list<string>` | `[]` | Layer ARNs to attach (max 5). |
| `environmentVariables` | `map<string, string>` | `{}` | Environment variables passed to the function at runtime. |
| `tags` | `map<string, string>` | `{}` | Tags merged with standard OpenMCF tags. User tags take precedence on conflicts. |
| `resourceGroupId` | `string` | `""` | Alibaba Cloud resource group ID for organizational grouping. |

## Examples

### Minimal Python Function

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: hello-world
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudFunction.hello-world
spec:
  region: cn-hangzhou
  functionName: hello-world
  handler: index.handler
  runtime: python3.12
  code:
    ossBucketName: my-code-bucket
    ossObjectName: functions/hello-world.zip
```

### VPC-Connected API Function with Logging

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: api-handler
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AliCloudFunction.api-handler
spec:
  region: cn-shanghai
  functionName: api-handler
  handler: index.handler
  runtime: nodejs20
  description: API handler with VPC database access
  cpu: 1.0
  memorySize: 2048
  timeout: 30
  instanceConcurrency: 10
  internetAccess: true
  code:
    ossBucketName: staging-code-bucket
    ossObjectName: functions/api-handler-v1.2.0.zip
  role:
    value: acs:ram::123456789:role/fc-api-execution-role
  vpcConfig:
    vpcId:
      value: vpc-abc123
    vswitchIds:
      - value: vsw-shanghai-a
      - value: vsw-shanghai-b
    securityGroupId:
      value: sg-xyz789
  logConfig:
    project:
      value: staging-logs
    logstore: api-function-logs
    logBeginRule: DefaultRegex
    enableInstanceMetrics: true
    enableRequestMetrics: true
  environmentVariables:
    DB_HOST: rm-abc123.mysql.rds.aliyuncs.com
    DB_PORT: "3306"
  tags:
    team: backend
    service: api
```

### GPU-Accelerated Container Function

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: ml-inference
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudFunction.ml-inference
spec:
  region: cn-hangzhou
  functionName: ml-inference
  handler: not-applicable
  runtime: custom-container
  cpu: 4.0
  memorySize: 16384
  timeout: 300
  diskSize: 10240
  customContainerConfig:
    image: registry.cn-hangzhou.aliyuncs.com/ml-team/inference:v2.1.0
    entrypoint:
      - /app/serve.sh
    port: 8080
    healthCheckConfig:
      httpGetUrl: /healthz
      initialDelaySeconds: 10
      periodSeconds: 15
      timeoutSeconds: 2
      failureThreshold: 3
      successThreshold: 1
  instanceLifecycleConfig:
    initializer:
      handler: index.warmup
      timeout: 120
    preStop:
      handler: index.cleanup
      timeout: 30
  gpuConfig:
    gpuMemorySize: 8192
    gpuType: fc.gpu.ampere.1
  vpcConfig:
    vpcId:
      value: vpc-ml-prod
    vswitchIds:
      - value: vsw-gpu-zone-a
    securityGroupId:
      value: sg-ml-inference
  nasConfig:
    userId: 0
    groupId: 0
    mountPoints:
      - serverAddr: 0f2a1b2c3d-abc12.cn-hangzhou.nas.aliyuncs.com:/models
        mountDir: /mnt/models
        enableTls: true
  logConfig:
    project:
      value: production-logs
    logstore: ml-inference-logs
    enableInstanceMetrics: true
    enableRequestMetrics: true
  role:
    value: acs:ram::123456789:role/fc-ml-inference-role
  tags:
    workload: ml-inference

```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `function_id` | `string` | The FC function ID assigned by Alibaba Cloud |
| `function_name` | `string` | The function name (mirrors the spec input) |
| `function_arn` | `string` | The function ARN (`acs:fc:{region}:{account-id}:functions/{function-name}`). Used in RAM policies and trigger configurations. |

## Related Components

- [AliCloudRamRole](/docs/catalog/alicloud/alicloudramrole) — provides the execution role for the function
- [AliCloudLogProject](/docs/catalog/alicloud/alicloudlogproject) — provides the SLS project for function logging
- [AliCloudVpc](/docs/catalog/alicloud/alicloudvpc) — provides the VPC for private network access
- [AliCloudSecurityGroup](/docs/catalog/alicloud/alicloudsecuritygroup) — provides the security group for VPC-attached functions
- [AliCloudNasFileSystem](/docs/catalog/alicloud/alicloudnasfilesystem) — provides NAS mount targets for shared file storage
- [AliCloudStorageBucket](/docs/catalog/alicloud/alicloudstoragebucket) — provides OSS buckets for function code packages

# Alibaba Cloud Function Compute v3: From Console Clicks to Control Planes

## Introduction

Alibaba Cloud Function Compute (FC) is the platform's fully managed, event-driven compute service. Users package code or a container image, upload it, and the platform handles provisioning, scaling, and reclamation of compute instances. FC v3 — the current generation — replaces the service-scoped model of FC v2 with a flat, service-less architecture where each function is a top-level resource. VPC access, logging, IAM roles, lifecycle hooks, and storage mounts are configured directly on the function rather than inherited from a parent service.

The core unit of work is the **function**. A function has:

- A **runtime** — a built-in language environment (Python, Node.js, Java, Go, PHP, .NET), a custom Debian-based environment, or a custom container image.
- A **handler** — the entry point that FC invokes for each event.
- A **code package** — a ZIP file on OSS, a base64-encoded inline payload, or (for custom containers) a registry image URI.
- **Compute sizing** — vCPU, memory, disk, and concurrency limits that control the shape of each instance.
- **Networking** — optional VPC attachment that enables private access to databases, caches, NAS, and other VPC-internal resources.
- **Logging** — optional integration with Simple Log Service (SLS) for structured invocation logs and metrics.
- **Lifecycle hooks** — initializer and pre-stop handlers for warm-up and cleanup logic.

Despite this apparent simplicity, deploying functions correctly and consistently across environments is a multi-step process with many failure modes. Misconfigured VPC settings strand functions without network access. Missing IAM roles cause silent permission failures at invocation time. Forgetting to wire SLS log config means invocation logs are lost. These are not complexity problems — they are coordination problems. The same fields exist in every tool; the challenge is getting them right every time.

This document examines the full lifecycle of FC v3 function deployment — from manual console workflows to infrastructure-as-code to control-plane automation — and explains how OpenMCF captures the common 80% of function configuration in a single, validated manifest.

## Deployment Landscape

### Level 0: Alibaba Cloud Console

The Alibaba Cloud console provides a guided workflow for creating functions:

1. Navigate to **Function Compute** in the console
2. Click **Create Function**, select the creation method (from scratch, template, or container)
3. Choose a runtime, enter a function name, set the handler
4. Optionally configure compute resources, VPC, logging, and environment variables
5. Upload code or specify an OSS location / container image

The console handles provider defaults well — omitted fields like `memorySize` and `timeout` receive reasonable values. However, the workflow is inherently manual and non-reproducible:

**Common Mistakes**:

1. **VPC half-configuration** — Selecting a VPC but forgetting to select VSwitches in multiple availability zones. The function deploys but has limited resilience if one AZ goes down.

2. **Missing execution role** — Creating a function that needs to access other Alibaba Cloud services (OSS, RDS, SLS) without attaching a RAM role. The function deploys successfully but fails at invocation time with opaque permission errors.

3. **Logging not configured** — The console does not require SLS log configuration. Functions created without it produce no invocation logs, making debugging impossible until the operator manually adds log config and redeploys.

4. **Incorrect handler format** — Different runtimes use different handler conventions. `index.handler` works for Python and Node.js, but Java expects `com.example.Main::handleRequest` and Go expects `main`. The console does not validate the handler format against the selected runtime.

5. **Forgetting lifecycle hooks** — For functions that need warm-up (loading ML models, opening database connection pools), omitting the initializer hook causes cold-start latency on every new instance, not just the first invocation.

**Verdict**: Useful for experimentation and one-off test functions. Not viable for production environments requiring reproducibility, version control, and multi-environment promotion.

### Level 1: Alibaba Cloud CLI (`aliyun`)

The `aliyun` CLI provides direct access to the FC v3 API:

```bash
# Create a function
aliyun fc POST /2023-03-30/functions --body '{
  "functionName": "hello-world",
  "handler": "index.handler",
  "runtime": "python3.12",
  "code": {
    "ossBucketName": "my-code-bucket",
    "ossObjectName": "functions/hello.zip"
  }
}'

# Update compute sizing
aliyun fc PUT /2023-03-30/functions/hello-world --body '{
  "cpu": 1.0,
  "memorySize": 2048,
  "timeout": 30
}'

# Delete function
aliyun fc DELETE /2023-03-30/functions/hello-world
```

The CLI brings scriptability but not idempotency. Creating a function that already exists returns an error. Updating requires knowing which fields have changed. There is no state tracking — determining whether the live function matches the desired configuration requires a GET and field-by-field comparison.

**The VPC wiring problem**: Attaching a function to a VPC requires three IDs (VPC, VSwitch, security group) that must be obtained from other API calls or hardcoded. Scripts that reference these IDs by value break when the VPC is recreated. Scripts that look them up by name require additional API calls and error handling.

**The code deployment problem**: Uploading code requires first pushing a ZIP to OSS, then referencing the bucket/object in the function creation call. This two-step process is not atomic — if the OSS upload succeeds but the function update fails, the code artifact exists but isn't deployed.

**Verdict**: Suitable for debugging, one-off invocations, and CI/CD pipeline steps where a higher-level tool manages idempotency. Not suitable as the primary provisioning method for production functions.

### Level 2: Infrastructure as Code (Terraform / OpenTofu)

Terraform's `alicloud` provider exposes the `alicloud_fcv3_function` resource:

```hcl
resource "alicloud_fcv3_function" "main" {
  function_name = "hello-world"
  handler       = "index.handler"
  runtime       = "python3.12"
  memory_size   = 2048
  timeout       = 30

  code {
    oss_bucket_name = "my-code-bucket"
    oss_object_name = "functions/hello.zip"
  }

  vpc_config {
    vpc_id            = alicloud_vpc.main.id
    vswitch_ids       = [alicloud_vswitch.a.id, alicloud_vswitch.b.id]
    security_group_id = alicloud_security_group.fc.id
  }

  log_config {
    project  = alicloud_log_project.main.name
    logstore = alicloud_log_store.fc_logs.name
  }

  role = alicloud_ram_role.fc_execution.arn
}
```

Terraform solves idempotency, state management, and cross-resource references. The VPC IDs, SLS project name, and RAM role ARN are computed references rather than hardcoded strings. Plan/apply cycles make changes predictable.

**What Terraform does not solve**: Validation that the role trusts the FC service principal, that the SLS project and logstore actually exist in the correct region, that the VSwitch IDs belong to the specified VPC, or that the handler format matches the runtime. These are runtime errors that surface only during `terraform apply`.

**The module sprawl problem**: A production function typically depends on a RAM role, a VPC with VSwitches and security groups, an SLS project with a log store, and an OSS bucket for code. Declaring all of these in a single Terraform configuration creates a monolithic root module. Splitting them into modules requires output-variable wiring that grows with each dependency.

### Level 3: Pulumi (Go SDK)

Pulumi's `pulumi-alicloud` SDK provides the `fc.V3Function` resource with type-safe arguments:

```go
function, err := fc.NewV3Function(ctx, "hello-world", &fc.V3FunctionArgs{
    FunctionName: pulumi.String("hello-world"),
    Handler:      pulumi.String("index.handler"),
    Runtime:      pulumi.String("python3.12"),
    MemorySize:   pulumi.Int(2048),
    Timeout:      pulumi.Int(30),
    Code: &fc.V3FunctionCodeArgs{
        OssBucketName: pulumi.String("my-code-bucket"),
        OssObjectName: pulumi.String("functions/hello.zip"),
    },
})
```

The Go type system catches some misconfiguration at compile time (e.g., passing an `int` where a `string` is expected), but it does not validate semantic constraints like runtime-handler compatibility or VPC wiring correctness. The operational benefits over Terraform are real (conditionals, loops, and error handling in a real language) but the coordination problem remains.

### Level 4: Control-Plane Automation (OpenMCF)

OpenMCF operates one level above Terraform and Pulumi. Instead of declaring low-level resources, the operator writes a manifest that describes the desired function:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudFunction
metadata:
  name: hello-world
spec:
  region: cn-hangzhou
  functionName: hello-world
  handler: index.handler
  runtime: python3.12
  code:
    ossBucketName: my-code-bucket
    ossObjectName: functions/hello.zip
```

OpenMCF validates the manifest against a protobuf schema before any infrastructure is touched. Invalid runtimes, out-of-range memory sizes, and missing required fields are caught at admission time — not during a 3-minute Terraform apply cycle. The manifest is IaC-engine agnostic: the same YAML drives either the Pulumi or Terraform module.

Cross-component references use the `StringValueOrRef` pattern: instead of hardcoding a VPC ID, the manifest can reference an `AlicloudVpc` resource by name, and OpenMCF resolves the output value at deployment time.

## Production Architecture: FC v3 in Depth

### The Service-Less Model

FC v2 organized functions under **services** — a service was a container for related functions, and VPC/logging/role configuration was set at the service level. This meant:

- Changing a VPC binding required updating the service, affecting all functions in it.
- Functions in the same service shared the same execution role, even if they needed different permissions.
- Moving a function to a different VPC meant moving it to a different service.

FC v3 eliminates the service layer. Every function is a top-level resource. VPC config, logging, role, and all other settings are per-function. This makes function configuration more verbose (each function must declare its own VPC config) but also more flexible and predictable.

The provider resource in Terraform is `alicloud_fcv3_function`. The Pulumi SDK exposes `fc.V3Function`. Both map to the same FC v3 API.

### Runtime Families

FC v3 supports three runtime families, each with different deployment models:

**Built-in runtimes** — `python3.12`, `python3.10`, `python3.9`, `python3`, `nodejs20`, `nodejs18`, `nodejs16`, `nodejs14`, `java11`, `java8`, `go1`, `php7.2`, `dotnetcore3.1`. The platform provides the runtime; the user provides a code package (ZIP). The handler entry point follows runtime-specific conventions.

**Custom runtimes** — `custom`, `custom.debian10`, `custom.debian11`, `custom.debian12`. The user provides a code package that includes an HTTP server binary. FC starts the server and forwards invocations to it over HTTP on a configurable port. The `customRuntimeConfig` block specifies the bootstrap command, arguments, and health check. This is the preferred model for languages not in the built-in list (Rust, Ruby, Elixir) or for frameworks that need full HTTP server control.

**Custom container** — `custom-container`. The user provides a container image (ACR or any registry accessible to FC). The `customContainerConfig` block specifies the image URI, entrypoint, command, port, and health check. This is the heaviest deployment model (slower cold starts due to image pull) but offers maximum flexibility: any language, any library, any system dependency. The `handler` field is still required by the provider (set it to a placeholder like `not-applicable`).

### Compute Sizing

Each function instance runs with allocated compute resources:

| Parameter | Range | Default Behavior |
|-----------|-------|-----------------|
| `cpu` | 0.05-16 vCPU | Computed from `memorySize` |
| `memorySize` | 64-32768 MB | Provider default |
| `timeout` | 1-86400 seconds (24h) | Provider default |
| `diskSize` | 512+ MB | Provider default |
| `instanceConcurrency` | 1-200 | 1 (single-request per instance) |

The `instanceConcurrency` setting is critical for throughput. At the default of 1, each instance handles one request at a time; scaling is purely horizontal. At higher values, a single instance handles multiple concurrent requests, reducing the total instance count needed. However, this requires the function code to be safe for concurrent execution (no shared mutable state, no global locks).

### VPC Networking

Functions that need to access VPC-internal resources (RDS, Redis, NAS, ECS instances) must be attached to a VPC:

```
vpcConfig:
  vpcId: <VPC ID>
  vswitchIds: [<VSwitch ID in AZ-a>, <VSwitch ID in AZ-b>]
  securityGroupId: <Security Group ID>
```

When VPC config is set:

1. FC provisions Elastic Network Interfaces (ENIs) in the specified VSwitches
2. Function instances route traffic through these ENIs
3. Security group rules control inbound/outbound access
4. The function can reach private IP addresses within the VPC

**Multi-AZ placement**: Specifying VSwitches in multiple availability zones improves function availability. If one AZ experiences issues, FC can place instances in the other AZ.

**Internet access**: When `internetAccess` is `true`, functions can reach the public internet even when attached to a VPC (traffic routes through the ENI). When `false`, all outbound internet traffic is blocked.

### Logging Integration

FC v3 integrates with Simple Log Service (SLS) for invocation logging:

```
logConfig:
  project: <SLS Project Name>
  logstore: <SLS Logstore Name>
  logBeginRule: DefaultRegex
  enableInstanceMetrics: true
  enableRequestMetrics: true
```

Without log config, function stdout/stderr output is not persisted anywhere. This is the most common production misconfiguration — the function runs correctly but operators cannot debug failures because there are no logs.

The `enableInstanceMetrics` flag activates per-instance CPU and memory usage reporting. The `enableRequestMetrics` flag activates per-request latency and status code metrics. Both are disabled by default but are essential for production observability.

### Lifecycle Hooks

FC v3 supports two lifecycle hooks:

**Initializer** — Runs once when a new instance is created, before it receives any invocations. Use for warm-up tasks:

- Loading ML models into memory
- Opening database connection pools
- Downloading configuration files from OSS

The initializer has its own handler entry point and timeout (up to 600 seconds). If the initializer fails, the instance is not used for invocations.

**Pre-stop** — Runs before an idle instance is reclaimed. Use for cleanup tasks:

- Flushing write buffers
- Closing database connections
- Sending final metric batches

The pre-stop hook has a timeout of up to 900 seconds.

Both hooks can be defined as handler references (e.g., `index.initializer`) or as command arrays.

### NAS Storage Mounts

Functions can mount NAS (Network Attached Storage) file systems for shared, persistent storage:

```
nasConfig:
  userId: 0
  groupId: 0
  mountPoints:
    - serverAddr: "{file-system-id}-{mount-target-id}.{region}.nas.aliyuncs.com:/{path}"
      mountDir: /mnt/data
      enableTls: true
```

NAS mounts require VPC config (NAS mount targets are VPC-internal). The `userId` and `groupId` control POSIX file ownership for read/write access.

Use cases:
- Sharing model weights across multiple function instances
- Persistent scratch storage larger than the ephemeral disk
- Cross-function data sharing without going through OSS

### GPU Acceleration

FC v3 supports GPU-accelerated instances for AI/ML inference:

```
gpuConfig:
  gpuMemorySize: 8192
  gpuType: fc.gpu.ampere.1
```

Available GPU types: `fc.gpu.tesla.1`, `fc.gpu.ampere.1`, `fc.gpu.ada.1`, `g1`. GPU functions typically pair with NAS mounts (for model storage) and VPC config (for accessing internal model registries).

### Triggers, Aliases, and Versions

FC v3 supports triggers (HTTP, timer, OSS events, MNS, Kafka), function aliases, versions, and provisioned concurrency configurations. These resources have independent lifecycles — an alias points to a function version, a trigger invokes a function or alias, and provisioned concurrency is configured per function/qualifier.

OpenMCF intentionally excludes triggers, aliases, versions, and concurrency configs from the `AlicloudFunction` component. Each of these is a separate resource with its own create/update/delete lifecycle. Bundling them into the function component would create a monolithic resource where changing a trigger forces a full function redeployment plan. Separate components (when implemented) will reference the function by name or ARN.

## Best Practices

| Area | Recommendation | Rationale |
|------|---------------|-----------|
| Naming | Use descriptive, environment-scoped function names (`{env}-{service}-{action}`) | Function names are immutable after creation (ForceNew) |
| Compute | Set explicit `cpu` and `memorySize` in production | Provider defaults may not match workload requirements |
| Concurrency | Set `instanceConcurrency` > 1 for I/O-bound functions | Reduces instance count and cold starts |
| VPC | Use VSwitches in 2+ AZs | Improves availability during AZ outages |
| Logging | Always configure `logConfig` in production | Without it, invocation logs are lost |
| Metrics | Enable both `enableInstanceMetrics` and `enableRequestMetrics` | Essential for capacity planning and SLA monitoring |
| IAM | Use least-privilege RAM roles; one role per function | Shared roles create a blast radius for permission escalation |
| Lifecycle | Add initializer hooks for functions with heavy cold starts | Initializer runs once per instance, not per request |
| Code | Use OSS for code deployment, not inline `zipFile` | `zipFile` has a size limit and clutters manifests |
| Layers | Extract shared dependencies into layers | Reduces code package size and deployment time |
| Tags | Use `tags` for cost allocation and operational grouping | FC tags propagate to billing reports |

## What OpenMCF Supports

### The 80/20 Design

OpenMCF's `AlicloudFunction` component wraps a single `alicloud_fcv3_function` resource. It exposes the 80% of configuration that covers the vast majority of production deployments:

**Included**:
- Function identity: name, handler, runtime, description
- Compute sizing: CPU, memory, timeout, disk, instance concurrency
- Code deployment: OSS bucket/object, inline ZIP, checksum
- IAM: execution role via `StringValueOrRef` (value or cross-resource reference)
- Networking: VPC ID, VSwitch IDs, security group ID — all via `StringValueOrRef`
- Logging: SLS project (via `StringValueOrRef`), logstore, log begin rule, instance/request metrics
- Custom container: image, entrypoint, command, port, health check
- Custom runtime: bootstrap command, args, port, health check
- Lifecycle hooks: initializer and pre-stop with handler, timeout, and command
- NAS mounts: user/group ID, mount points with TLS
- GPU acceleration: memory size and GPU type
- Layers, environment variables, tags, resource group ID

**Excluded (separate lifecycle)**:
- Triggers (HTTP, timer, OSS, MNS, Kafka)
- Function aliases
- Function versions
- Provisioned concurrency configurations
- Custom domains
- Async invocation configurations
- On-demand configurations

### Foreign Keys

The component uses `StringValueOrRef` for cross-resource references:

| Field | Default Kind | Default Output Path |
|-------|-------------|-------------------|
| `role` | `AlicloudRamRole` | `status.outputs.arn` |
| `vpcConfig.vpcId` | `AlicloudVpc` | `status.outputs.vpc_id` |
| `vpcConfig.securityGroupId` | `AlicloudSecurityGroup` | `status.outputs.security_group_id` |
| `logConfig.project` | `AlicloudLogProject` | `status.outputs.project_name` |

Each `StringValueOrRef` field accepts either a direct `value` (a string literal) or a `ref` (a reference to another OpenMCF resource's output). This allows both standalone deployment (hardcoded IDs) and orchestrated deployment (cross-resource wiring).

### Implementation Landscape

Both the Pulumi module and the Terraform module implement the same specification:

**Pulumi module** (`iac/pulumi/module/`):

| File | Role |
|------|------|
| `main.go` | Creates the alicloud provider and `fc.V3Function` with all optional config blocks |
| `locals.go` | Initializes locals struct, merges standard + user tags, provides optional-field helpers |
| `outputs.go` | Defines output key constants: `function_id`, `function_name`, `function_arn` |

The Pulumi module creates a single `fc.V3Function` resource. Each optional config block (`vpcConfig`, `logConfig`, `customContainerConfig`, `customRuntimeConfig`, `instanceLifecycleConfig`, `nasConfig`, `gpuConfig`) is conditionally included only when the corresponding spec field is non-nil. This prevents sending empty/zero-valued blocks to the provider, which would override server-side defaults.

**Terraform module** (`iac/tf/`):

| File | Role |
|------|------|
| `main.tf` | `alicloud_fcv3_function` resource with dynamic blocks for all optional configs |
| `locals.tf` | Tag merging: base tags + org/env tags + user tags |
| `outputs.tf` | `function_id`, `function_name`, `function_arn` |
| `variables.tf` | Typed variable definitions mirroring the protobuf spec |
| `provider.tf` | Alicloud provider with region from spec |

The Terraform module uses `dynamic` blocks extensively. Each optional configuration section (`code`, `vpc_config`, `log_config`, `custom_container_config`, `custom_runtime_config`, `instance_lifecycle_config`, `nas_config`, `gpu_config`) is wrapped in a `dynamic` block that only renders when the corresponding variable is non-null.

### Tag Management

Both modules inject standard OpenMCF tags and merge them with user-provided tags:

| Tag Key | Source | Description |
|---------|--------|-------------|
| `resource` | Hardcoded `"true"` | Identifies managed resources |
| `resource_name` | `metadata.name` | Manifest resource name |
| `resource_kind` | Hardcoded `"alicloud_function"` | Component type |
| `resource_id` | `metadata.id` (if set) | Optional unique ID |
| `organization` | `metadata.org` (if set) | Organization scope |
| `environment` | `metadata.env` (if set) | Environment scope |

User-provided `spec.tags` are merged after standard tags. On key conflicts, user tags take precedence.

### Validation

The protobuf schema enforces constraints at admission time:

| Field | Constraint |
|-------|-----------|
| `region` | Required, non-empty |
| `functionName` | Required, 1-128 characters |
| `handler` | Required, non-empty |
| `runtime` | Required, must be in the allowed set |
| `cpu` | 0.05-16 |
| `memorySize` | 64-32768 |
| `timeout` | 1-86400 |
| `diskSize` | >= 512 |
| `instanceConcurrency` | 1-200 |
| `gpuConfig.gpuMemorySize` | > 0 |
| `gpuConfig.gpuType` | Must be in `[fc.gpu.tesla.1, fc.gpu.ampere.1, fc.gpu.ada.1, g1]` |
| `logConfig.logBeginRule` | Must be `None` or `DefaultRegex` |
| `healthCheckConfig.initialDelaySeconds` | 0-120 |
| `healthCheckConfig.timeoutSeconds` | 0-3 |
| `healthCheckConfig.periodSeconds` | 0-120 |
| `healthCheckConfig.failureThreshold` | 1-120 |
| `healthCheckConfig.successThreshold` | 1-120 |
| `nasMountPoint.serverAddr` | Non-empty |
| `nasMountPoint.mountDir` | Non-empty |
| `customContainerConfig.image` | Non-empty when config is set |
| `customRuntimeConfig.port` | 0-65535 |

These validations run before any infrastructure is provisioned. An invalid manifest is rejected with a descriptive error message, not a provider-level error during apply.

### Stack Outputs

After deployment, three outputs are available:

| Output | Type | Description |
|--------|------|-------------|
| `function_id` | `string` | The FC function ID assigned by Alibaba Cloud |
| `function_name` | `string` | The function name (mirrors spec input) |
| `function_arn` | `string` | The Alibaba Cloud Resource Name — `acs:fc:{region}:{account-id}:functions/{function-name}` |

The `function_arn` is the primary value used by downstream resources: trigger configurations, RAM policies, and event source mappings all reference functions by ARN.

## Conclusion

Deploying an FC v3 function involves a moderate number of configuration fields, but the real complexity lies in coordinating those fields correctly: matching handler formats to runtimes, wiring VPC/VSwitch/security-group IDs, pairing NAS mounts with VPC config, and ensuring SLS logging is always present in production.

OpenMCF's `AlicloudFunction` component addresses this coordination problem by:

1. **Schema validation** — Protobuf-level constraints catch invalid configurations before any infrastructure is touched.
2. **Foreign keys** — `StringValueOrRef` fields enable type-safe cross-resource references without hardcoding IDs.
3. **Conditional blocks** — Both the Pulumi and Terraform modules only emit configuration blocks that are actually set, avoiding provider-level defaults being overridden by zero values.
4. **Tag management** — Standard OpenMCF tags are automatically injected for operational grouping and cost attribution.
5. **Single resource focus** — The component manages one function. Triggers, aliases, and versions are separate lifecycle concerns managed by separate components.

The result is a manifest that captures the intent ("I want a Python function in cn-hangzhou that can access my VPC and logs to SLS") without requiring the operator to understand the specific provider arguments, dynamic block syntax, or conditional nil-handling that the underlying IaC modules manage.

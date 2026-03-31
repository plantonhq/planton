# AliCloudFunction Pulumi Examples

Apply any example below using the OpenMCF CLI:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal Python Function

A Python function with code deployed from an OSS bucket. Uses provider defaults
for compute sizing (CPU, memory, timeout).

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: hello-world
spec:
  region: cn-hangzhou
  functionName: hello-world
  handler: index.handler
  runtime: python3.12
  code:
    ossBucketName: my-code-bucket
    ossObjectName: functions/hello-world.zip
```

**Key Points**:
- Only the four required fields (`region`, `functionName`, `handler`, `runtime`) plus `code` are set
- Compute sizing defers to FC provider defaults
- No VPC attachment — the function runs in the FC shared network
- No logging — invocation logs are not persisted (suitable for testing only)

---

## VPC-Connected Node.js Function with Logging

A Node.js function that accesses a database in a VPC, logs to SLS, and runs
with a dedicated execution role. Compute sizing is tuned for API workloads.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: api-handler
  org: my-org
  env: staging
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
    NODE_ENV: staging
  tags:
    team: backend
    service: api
```

**Key Points**:
- VPC config uses VSwitches in two AZs for resilience
- SLS logging with instance and request metrics enabled
- `instanceConcurrency: 10` allows each instance to handle 10 concurrent requests
- `internetAccess: true` allows outbound internet even though the function is in a VPC
- Environment variables pass database connection info

---

## Production Container Function with Lifecycle Hooks

A containerized function running a custom Docker image with initializer and
pre-stop hooks. Includes NAS mount for shared model data and GPU acceleration.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudFunction
metadata:
  name: ml-inference
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  functionName: ml-inference
  handler: not-applicable
  runtime: custom-container
  description: ML inference function with GPU and NAS
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
      - value: vsw-gpu-zone-b
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
    gpu: "true"
```

**Key Points**:
- `runtime: custom-container` with a full `customContainerConfig` block
- `handler: not-applicable` — the provider requires this field but it is ignored for containers
- GPU acceleration via `gpuConfig` with Ampere GPU
- NAS mount at `/mnt/models` for shared model weights (requires VPC config)
- Initializer hook (120s timeout) for model loading warm-up
- Pre-stop hook (30s timeout) for graceful shutdown
- Health check on `/healthz` verifies the container is ready before receiving traffic

---

## Next Steps

- Customize compute sizing based on profiling results
- Add environment-specific manifests for dev/staging/production
- Configure triggers (HTTP, timer, OSS events) using separate OpenMCF components
- Set up function aliases for blue-green deployments

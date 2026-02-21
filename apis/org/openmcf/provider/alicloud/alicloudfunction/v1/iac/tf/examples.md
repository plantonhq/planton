# AliCloudFunction Terraform Examples

Apply any example below using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --auto-approve
```

---

## Minimal Python Function

A Python function with code deployed from an OSS bucket. Uses provider defaults
for compute sizing.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

Resources created:
- `alicloud_fcv3_function.main`

No VPC, logging, or lifecycle configuration — the function uses FC defaults.

---

## VPC-Connected API Function with Logging

A Node.js function that accesses VPC-internal resources and logs invocations to
SLS. Compute sizing is explicitly configured for API workloads.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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
  description: API handler with database access
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
```

Resources created:
- `alicloud_fcv3_function.main` — with `vpc_config`, `log_config`, `code`, and
  `environment_variables` dynamic blocks populated

---

## Production Container Function with GPU and NAS

A containerized ML inference function with GPU acceleration, NAS-mounted model
storage, and lifecycle hooks for warm-up and cleanup.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

Resources created:
- `alicloud_fcv3_function.main` — with `custom_container_config`,
  `instance_lifecycle_config`, `gpu_config`, `nas_config`, `vpc_config`, and
  `log_config` dynamic blocks populated

---

## After Deploying

Confirm the function exists using the Alibaba Cloud CLI:

```shell
aliyun fc GET /2023-03-30/functions/<function-name>
```

Invoke the function:

```shell
aliyun fc POST /2023-03-30/functions/<function-name>/invocations \
  --body '{"key": "value"}'
```

List all functions in the region:

```shell
aliyun fc GET /2023-03-30/functions
```

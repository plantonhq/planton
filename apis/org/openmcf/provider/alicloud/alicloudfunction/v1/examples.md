# Examples

## Minimal Python Function

A simple Python function with code deployed from an OSS bucket. Uses default compute sizing.

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
    ossObjectName: functions/hello-world.zip
```

## Production API Function with VPC and Logging

A Node.js function that accesses a database in a VPC, logs to SLS, and runs with a dedicated execution role. Compute sizing is tuned for API workloads.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudFunction
metadata:
  name: api-handler
  org: my-org
  env: production
spec:
  region: cn-shanghai
  functionName: api-handler
  handler: index.handler
  runtime: nodejs20
  description: Production API handler with database access
  cpu: 1.0
  memorySize: 2048
  timeout: 30
  instanceConcurrency: 10
  internetAccess: true
  code:
    ossBucketName: prod-code-bucket
    ossObjectName: functions/api-handler-v2.3.1.zip
  role:
    value: acs:ram::123456789:role/fc-api-execution-role
  vpcConfig:
    vpcId:
      value: vpc-abc123
    vswitchIds:
      - value: vsw-hangzhou-a
      - value: vsw-hangzhou-b
    securityGroupId:
      value: sg-xyz789
  logConfig:
    project:
      value: production-logs
    logstore: api-function-logs
    logBeginRule: DefaultRegex
    enableInstanceMetrics: true
    enableRequestMetrics: true
  environmentVariables:
    DB_HOST: rm-abc123.mysql.rds.aliyuncs.com
    DB_PORT: "3306"
    NODE_ENV: production
  tags:
    team: backend
    service: api
```

## Custom Container Function

A containerized function running a custom Docker image. Suitable for runtimes not natively supported or complex dependency chains.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudFunction
metadata:
  name: container-func
spec:
  region: cn-hangzhou
  functionName: container-service
  handler: not-applicable
  runtime: custom-container
  cpu: 2.0
  memorySize: 4096
  timeout: 120
  customContainerConfig:
    image: registry.cn-hangzhou.aliyuncs.com/my-namespace/my-service:v1.0.0
    entrypoint:
      - /app/entrypoint.sh
    command:
      - serve
    port: 8080
    healthCheckConfig:
      httpGetUrl: /healthz
      initialDelaySeconds: 5
      periodSeconds: 10
      timeoutSeconds: 2
      failureThreshold: 3
      successThreshold: 1
  instanceLifecycleConfig:
    initializer:
      handler: index.initializer
      timeout: 30
    preStop:
      handler: index.cleanup
      timeout: 15
```

## GPU-Accelerated AI Inference Function

A function configured with GPU acceleration for machine learning inference workloads.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudFunction
metadata:
  name: inference-func
spec:
  region: cn-hangzhou
  functionName: ml-inference
  handler: index.handler
  runtime: custom.debian12
  cpu: 4.0
  memorySize: 16384
  timeout: 300
  diskSize: 10240
  gpuConfig:
    gpuMemorySize: 8192
    gpuType: fc.gpu.ampere.1
  nasConfig:
    userId: 0
    groupId: 0
    mountPoints:
      - serverAddr: 0f2a1b2c3d-abc12.cn-hangzhou.nas.aliyuncs.com:/models
        mountDir: /mnt/models
        enableTls: true
  vpcConfig:
    vpcId:
      value: vpc-model-serving
    vswitchIds:
      - value: vsw-gpu-zone
    securityGroupId:
      value: sg-inference
  code:
    ossBucketName: ml-artifacts
    ossObjectName: inference/handler-v3.zip
  tags:
    workload: ml-inference
```

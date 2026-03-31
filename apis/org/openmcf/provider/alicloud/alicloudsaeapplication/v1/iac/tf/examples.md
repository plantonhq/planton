# AliCloudSaeApplication Terraform Examples

Below are several examples demonstrating how to deploy SAE applications with the OpenMCF Terraform module.

After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Minimal Image Deployment

A container image deployment with the smallest resource tier.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudSaeApplication
metadata:
  name: hello-sae
spec:
  region: cn-hangzhou
  appName: hello-sae
  packageType: Image
  replicas: 1
  cpu: 500
  memory: 1024
  imageUrl: registry.cn-hangzhou.aliyuncs.com/my-ns/hello:latest
  envs:
    PORT: "8080"
  tags:
    team: dev
```

This example:
- Uses the `Image` package type with the smallest CPU/memory tier (500m / 1 GB)
- Single replica for development use
- Environment variable `PORT` is converted to SAE's JSON array format by `locals.tf`
- No VPC — runs in SAE-managed networking

---

## Java FatJar with VPC and Health Checks

A production-grade Java application with VPC networking, health probes, and a rolling update strategy.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudSaeApplication
metadata:
  name: order-service
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  appName: order-service
  appDescription: Order management REST API
  packageType: FatJar
  replicas: 3
  cpu: 4000
  memory: 8192
  packageUrl: https://deploy-bucket.oss-cn-shanghai.aliyuncs.com/order-service-2.1.0.jar
  packageVersion: "2.1.0"
  jdk: Open JDK 17
  jarStartOptions: "-Xms1g -Xmx4g -Dspring.profiles.active=prod"
  programmingLanguage: java
  timezone: Asia/Shanghai
  terminationGracePeriodSeconds: 30
  minReadyInstances: 2
  vpcId:
    value: vpc-prod-001
  vswitchId:
    value: vsw-prod-a
  securityGroupId:
    value: sg-prod-app
  namespaceId: cn-shanghai:production
  envs:
    SPRING_DATASOURCE_URL: jdbc:mysql://rm-xxx.mysql.rds.aliyuncs.com:3306/orders
    REDIS_HOST: r-xxx.redis.rds.aliyuncs.com
  liveness:
    httpGet:
      path: /actuator/health/liveness
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 30
    timeoutSeconds: 5
    failureThreshold: 3
  readiness:
    httpGet:
      path: /actuator/health/readiness
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 3
  updateStrategy:
    type: BatchUpdate
    batchUpdate:
      batch: 2
      batchWaitTime: 10
      releaseType: auto
  tags:
    team: platform
    costCenter: eng-123
```

This example:
- FatJar package with Open JDK 17 and explicit JVM heap settings
- 3 replicas with `minReadyInstances: 2` for availability during deploys
- VPC deployment for private access to RDS and Redis
- HTTP GET liveness/readiness probes against Spring Boot Actuator endpoints
- BatchUpdate strategy with 2 batches and automatic progression

---

## Production Container with Canary Releases

A high-traffic service with ACR EE, custom host aliases, SLS logging, and canary-style deployments.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudSaeApplication
metadata:
  name: payment-gateway
  org: fintech-corp
  env: production
spec:
  region: cn-hangzhou
  appName: payment-gateway
  packageType: Image
  replicas: 4
  cpu: 8000
  memory: 16384
  imageUrl: payment-registry.cn-hangzhou.cr.aliyuncs.com/services/gateway:v3.2.1
  acrInstanceId: cri-abc123def456
  vpcId:
    value: vpc-finance-prod
  vswitchId:
    value: vsw-finance-a
  securityGroupId:
    value: sg-finance-strict
  namespaceId: cn-hangzhou:finance-prod
  terminationGracePeriodSeconds: 45
  minReadyInstances: 3
  command: /app/gateway
  commandArgs:
    - "--config"
    - "/etc/gateway/config.yaml"
  envs:
    LOG_LEVEL: warn
    TLS_ENABLED: "true"
  customHostAliases:
    - hostName: hsm.internal
      ip: "10.0.5.100"
    - hostName: audit-log.internal
      ip: "10.0.5.200"
  liveness:
    httpGet:
      path: /health
      port: 8443
    initialDelaySeconds: 20
    periodSeconds: 30
    failureThreshold: 3
  readiness:
    httpGet:
      path: /ready
      port: 8443
    initialDelaySeconds: 10
    periodSeconds: 10
  slsConfigs: '[{"logDir":"/var/log/gateway","logType":"file_log"}]'
  updateStrategy:
    type: GrayBatchUpdate
    batchUpdate:
      batch: 4
      batchWaitTime: 30
      releaseType: manual
  tags:
    compliance: pci-dss
    team: payments
```

This configuration:
- ACR Enterprise Edition (`acrInstanceId`) for private registry authentication
- 4 replicas with `minReadyInstances: 3` to maintain 75% capacity during deployments
- Custom ENTRYPOINT and arguments override image defaults
- Custom host aliases inject internal hostnames into `/etc/hosts`
- GrayBatchUpdate with manual approval between 4 batches for canary releases
- SLS file log collection from `/var/log/gateway`
- 45-second graceful shutdown for in-flight transaction draining

---

## After Deploying

Verify the application using the Alibaba Cloud CLI:

```bash
# List SAE applications in a namespace
aliyun sae ListApplications --NamespaceId cn-hangzhou:finance-prod

# Get application details
aliyun sae DescribeApplicationConfig --AppId <app-id>

# Check instance status
aliyun sae DescribeApplicationInstances --AppId <app-id>

# View recent deployment history
aliyun sae ListAppEvents --AppId <app-id>
```

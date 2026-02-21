# AliCloudSaeApplication Pulumi Examples

## CLI

```bash
# Deploy using the OpenMCF CLI
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Preview changes
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

---

## Minimal Image Deployment

A container image deployment with the smallest resource tier. Suitable for development and testing.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

**Key Points:**
- Uses the `Image` package type with a public ACR registry URL
- Smallest CPU/memory tier (500m / 1 GB) for cost-effective development
- Single replica — no high-availability guarantees
- Environment variable `PORT` injected at runtime
- No VPC, health checks, or update strategy — all use provider defaults

---

## Java FatJar with VPC and Health Checks

A production-grade Java application deployed as a FatJar inside a VPC with liveness and readiness probes.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

**Key Points:**
- FatJar package type with Open JDK 17 and explicit JVM heap settings
- 3 replicas with `minReadyInstances: 2` to maintain availability during deploys
- VPC deployment for access to RDS and Redis instances
- HTTP GET liveness and readiness probes against Spring Boot Actuator endpoints
- 30-second initial delay on liveness to accommodate JVM startup time
- BatchUpdate strategy with 2 batches and 10-second pause between them
- `org` and `env` metadata populate the `organization` and `environment` tags automatically

---

## Production Container with ACR EE and Canary Releases

A high-traffic service pulling images from ACR Enterprise Edition with canary-style releases, custom host aliases, and SLS log collection.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

**Key Points:**
- ACR Enterprise Edition (`acrInstanceId`) for private registry access
- 4 replicas with `minReadyInstances: 3` — at least 75% capacity during deployments
- Custom ENTRYPOINT (`command`) and arguments (`commandArgs`) override the image defaults
- Custom host aliases inject internal service hostnames into `/etc/hosts`
- GrayBatchUpdate with manual approval between 4 batches for controlled canary releases
- 45-second graceful shutdown to drain in-flight payment transactions
- SLS file log collection from `/var/log/gateway`

---

**Next Steps:**

- See [README.md](./README.md) for CLI flows and debugging instructions
- See [overview.md](./overview.md) for module architecture details

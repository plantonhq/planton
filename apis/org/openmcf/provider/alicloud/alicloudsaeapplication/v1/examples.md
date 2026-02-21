# Examples

## Minimal Image Deployment

Deploy a container image with the smallest resource tier. Suitable for development and testing.

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

## Java FatJar with VPC and Health Checks

Production-grade Java application deployed as a FatJar in a VPC with liveness and readiness probes.

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

## Python Microservice

Lightweight Python ZIP application for a stateless API.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudSaeApplication
metadata:
  name: data-api
spec:
  region: cn-hangzhou
  appName: data-api
  packageType: PythonZip
  replicas: 2
  cpu: 1000
  memory: 2048
  packageUrl: https://deploy-bucket.oss-cn-hangzhou.aliyuncs.com/data-api-1.0.zip
  packageVersion: "1.0"
  programmingLanguage: other
  command: python
  commandArgs:
    - "-m"
    - "uvicorn"
    - "main:app"
    - "--host"
    - "0.0.0.0"
    - "--port"
    - "8080"
  envs:
    DATABASE_URL: postgresql://user:pass@db.internal:5432/mydb
  readiness:
    tcpSocket:
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 10
```

## Container Image with ACR Enterprise Edition

Production deployment pulling images from a private ACR EE registry with custom host aliases and full observability.

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

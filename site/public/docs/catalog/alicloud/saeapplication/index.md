---
title: "SaeApplication"
description: "SaeApplication deployment documentation"
icon: "package"
order: 100
componentName: "alicloudsaeapplication"
---

# AliCloud SaeApplication

Deploys an Alibaba Cloud SAE application. The component provisions a container-based serverless application with configurable compute tiers, VPC networking, health checks, rolling update strategy, environment variables, custom host aliases, and SLS log collection. Supports five package types: container images, Java JAR/WAR archives, and Python/PHP ZIP packages.

## What Gets Created

When you deploy an AliCloudSaeApplication resource, Planton provisions:

- **SAE Application** -- an `alicloud_sae_application` resource with the specified compute tier, replica count, deployment source, health probes, update strategy, and metadata tags

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or Planton provider config
- **VPC, VSwitch, and Security Group** if the application needs private network access to databases, caches, or other VPC-resident services
- **Container image** accessible from the deployment region (ACR Personal/Enterprise Edition or any Docker-compatible registry) when using `Image` package type
- **OSS bucket or HTTP endpoint** hosting the deployment package when using `FatJar`, `War`, `PythonZip`, or `PhpZip` package types
- **SAE namespace** pre-created if deploying into a specific namespace (the component does not create namespaces)

## Quick Start

Create a file `sae-app.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSaeApplication
metadata:
  name: my-sae-app
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudSaeApplication.my-sae-app
spec:
  region: cn-hangzhou
  appName: my-sae-app
  packageType: Image
  replicas: 1
  cpu: 1000
  memory: 2048
  imageUrl: registry.cn-hangzhou.aliyuncs.com/my-ns/my-app:latest
```

Deploy:

```shell
planton apply -f sae-app.yaml
```

This creates a single-instance SAE application running a container image with 1 vCPU and 2 GB memory.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `appName` | `string` | Application name. Immutable after creation. | Required; 1-36 chars; starts with letter |
| `packageType` | `string` | Deployment package format. Immutable after creation. | Required; one of: `Image`, `FatJar`, `War`, `PythonZip`, `PhpZip` |
| `replicas` | `int32` | Number of application instances. | Required; >= 1 |
| `cpu` | `int32` | CPU per instance in millicores. | Required; one of: 500, 1000, 2000, 4000, 8000, 16000, 32000 |
| `memory` | `int32` | Memory per instance in MB. | Required; one of: 1024, 2048, 4096, 8192, 12288, 16384, 24576, 32768, 65536, 131072 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `appDescription` | `string` | `""` | Human-readable description (max 1024 chars). |
| `vpcId` | `StringValueOrRef` | | VPC ID for network isolation. Immutable. |
| `vswitchId` | `StringValueOrRef` | | VSwitch ID for subnet placement. |
| `securityGroupId` | `StringValueOrRef` | | Security group for traffic rules. |
| `namespaceId` | `string` | `""` | SAE namespace (`{region}:{short_id}`). Immutable. |
| `imageUrl` | `string` | `""` | Container image URL. Required when `packageType` is `Image`. |
| `packageUrl` | `string` | `""` | Package URL (OSS/HTTP). Required for non-Image types. |
| `packageVersion` | `string` | `""` | Package version identifier. |
| `command` | `string` | `""` | Container ENTRYPOINT override. |
| `commandArgs` | `list<string>` | `[]` | Container CMD arguments. |
| `envs` | `map<string, string>` | `{}` | Environment variables. |
| `jdk` | `string` | `""` | JDK version for Java apps (e.g., `Open JDK 17`). |
| `jarStartOptions` | `string` | `""` | JVM startup options for FatJar. |
| `jarStartArgs` | `string` | `""` | Application arguments for FatJar. |
| `programmingLanguage` | `string` | `""` | One of: `java`, `php`, `other`. Immutable. |
| `timezone` | `string` | `""` | Application timezone (e.g., `Asia/Shanghai`). |
| `terminationGracePeriodSeconds` | `int32` | Provider default (30) | Graceful shutdown timeout (1-60s). |
| `minReadyInstances` | `int32` | | Min available instances during deploys. |
| `acrInstanceId` | `string` | `""` | ACR Enterprise Edition instance ID for private registry. |
| `liveness` | `object` | | Liveness probe (see health check fields below). |
| `readiness` | `object` | | Readiness probe (see health check fields below). |
| `customHostAliases` | `list<object>` | `[]` | Custom `/etc/hosts` entries: `{hostName, ip}`. |
| `updateStrategy` | `object` | | Rolling update configuration (see below). |
| `slsConfigs` | `string` | `""` | SLS log collection JSON config. |
| `tags` | `map<string, string>` | `{}` | Resource tags (merged with Planton metadata tags). |

**Health Check Fields** (for `liveness` and `readiness`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `httpGet.path` | `string` | | HTTP GET path (e.g., `/healthz`). |
| `httpGet.port` | `int32` | | HTTP GET port. |
| `tcpSocket.port` | `int32` | | TCP connection port. |
| `exec.command` | `string` | | Command to execute. |
| `initialDelaySeconds` | `int32` | | Delay before first check. |
| `periodSeconds` | `int32` | | Interval between checks. |
| `timeoutSeconds` | `int32` | | Timeout per check. |
| `failureThreshold` | `int32` | | Consecutive failures before action. |
| `successThreshold` | `int32` | | Consecutive successes to recover. |

**Update Strategy Fields** (for `updateStrategy`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `string` | | `BatchUpdate` or `GrayBatchUpdate`. |
| `batchUpdate.batch` | `int32` | | Number of release batches. |
| `batchUpdate.batchWaitTime` | `int32` | | Seconds between batches. |
| `batchUpdate.releaseType` | `string` | | `auto` or `manual`. |

## Examples

### Minimal Image Deployment

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSaeApplication
metadata:
  name: hello-sae
  labels:
    planton.dev/provisioner: pulumi
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
```

### Java FatJar with VPC and Health Checks

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSaeApplication
metadata:
  name: order-service
  labels:
    planton.dev/provisioner: pulumi
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  appName: order-service
  packageType: FatJar
  replicas: 3
  cpu: 4000
  memory: 8192
  packageUrl: https://deploy-bucket.oss-cn-shanghai.aliyuncs.com/order-service.jar
  packageVersion: "2.1.0"
  jdk: Open JDK 17
  jarStartOptions: "-Xms1g -Xmx4g"
  vpcId:
    value: vpc-prod-001
  vswitchId:
    value: vsw-prod-a
  securityGroupId:
    value: sg-prod-app
  namespaceId: cn-shanghai:production
  liveness:
    httpGet:
      path: /actuator/health/liveness
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 30
    failureThreshold: 3
  readiness:
    httpGet:
      path: /actuator/health/readiness
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 10
  minReadyInstances: 2
  updateStrategy:
    type: BatchUpdate
    batchUpdate:
      batch: 2
      releaseType: auto
```

### Production Container with Canary Releases

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSaeApplication
metadata:
  name: payment-gateway
  labels:
    planton.dev/provisioner: pulumi
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
  terminationGracePeriodSeconds: 45
  minReadyInstances: 3
  liveness:
    httpGet:
      path: /health
      port: 8443
    initialDelaySeconds: 20
    failureThreshold: 3
  readiness:
    httpGet:
      path: /ready
      port: 8443
    initialDelaySeconds: 10
  customHostAliases:
    - hostName: hsm.internal
      ip: "10.0.5.100"
  slsConfigs: '[{"logDir":"/var/log/gateway","logType":"file_log"}]'
  updateStrategy:
    type: GrayBatchUpdate
    batchUpdate:
      batch: 4
      batchWaitTime: 30
      releaseType: manual
  tags:
    compliance: pci-dss
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `app_id` | `string` | The SAE application ID assigned by Alibaba Cloud |
| `app_name` | `string` | The application name (mirrors the spec input) |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) -- VPC for network isolation
- [AliCloudVswitch](/docs/catalog/alicloud/vswitch) -- subnet for VPC-based deployment
- [AliCloudSecurityGroup](/docs/catalog/alicloud/security-group) -- network access control
- [AliCloudContainerRegistry](/docs/catalog/alicloud/containerregistry) -- private container image registry
- [AliCloudFunction](/docs/catalog/alicloud/function) -- event-driven serverless compute (alternative)
- [AliCloudLogProject](/docs/catalog/alicloud/log-project) -- SLS project for centralized log collection

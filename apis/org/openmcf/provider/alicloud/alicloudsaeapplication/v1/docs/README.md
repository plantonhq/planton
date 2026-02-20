# SAE Application Deployment: From Manual Console Workflows to Declarative Infrastructure

## Introduction

Alibaba Cloud Serverless App Engine (SAE) occupies a middle ground in the compute spectrum: it offers the container flexibility of Kubernetes without requiring cluster management, and the operational simplicity of a PaaS without sacrificing deployment model choice. An SAE application runs as one or more instances of a container image, JAR, WAR, or ZIP package, with CPU, memory, replica count, and networking configured per application. SAE handles instance scheduling, health monitoring, rolling deployments, and log collection.

The surface area of a single SAE application resource is deceptively large. Beyond the obvious fields (image URL, CPU, memory), a production-grade SAE deployment requires decisions about VPC placement, namespace isolation, health check configuration, update strategy, graceful shutdown behavior, and log shipping. Each of these decisions has operational consequences that compound when managing multiple applications across environments.

This document examines the SAE application deployment landscape, traces the operational risks at each level of automation, and explains the specific subset of SAE functionality that OpenMCF exposes through the `AlicloudSaeApplication` component.

---

## The SAE Deployment Landscape

SAE application management spans a spectrum from manual console interactions to fully declarative infrastructure pipelines. Each level introduces capabilities that address the shortcomings of the previous one.

### Level 0: Manual Provisioning via Alibaba Cloud Console

The SAE console provides a guided workflow for creating applications. The user selects a region, namespace, package type, resource tier, and deployment source, then clicks through networking and advanced configuration pages.

**Common Mistakes**:

1. **Resource Tier Mismatch**: The console allows any combination of CPU and memory, but SAE enforces specific tiers. Selecting 500 millicores with 8192 MB memory will fail at deploy time with a cryptic error. The valid CPU tiers are 500, 1000, 2000, 4000, 8000, 16000, and 32000 millicores. Memory tiers range from 1024 to 131072 MB, but not all combinations are valid — the provider documentation lists the permitted pairs.

2. **Missing VPC Configuration**: SAE applications can run in SAE-managed networking or inside a user-provided VPC. The console allows creating an application without VPC settings, which works for simple cases but breaks when the application needs to access RDS, Redis, or other VPC-resident services. Switching from managed networking to VPC networking after creation requires recreating the application (the `vpc_id` field is immutable).

3. **Namespace Confusion**: SAE namespaces are region-scoped and format-specific (`{region}:{short_id}`). Creating an application in the wrong namespace isolates it from configuration items and service discovery mechanisms shared by other applications. The namespace ID is immutable after creation.

4. **Health Check Omission**: The console does not enforce health check configuration. Applications deployed without liveness or readiness probes will appear healthy even when the application process has crashed (but the container is still running) or when the application is still initializing (causing request failures during rolling updates).

5. **Package Type Lock-in**: The `package_type` field (Image, FatJar, War, PythonZip, PhpZip) is immutable after creation. Selecting FatJar for an application that will later be containerized requires destroying and recreating the application, losing the application ID and any associated configuration.

**Console Fragility**: The SAE console orchestrates multiple API calls internally. Partial failures during application creation can leave orphaned resources (namespaces, SLB instances) that are not cleaned up automatically.

**Verdict**: Acceptable for exploration and one-off experiments. Not suitable for managing applications that need to be reproducible across environments.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun sae` CLI provides imperative commands for application lifecycle management:

```bash
# Create an application
aliyun sae CreateApplication \
  --AppName my-service \
  --PackageType Image \
  --ImageUrl registry.cn-hangzhou.aliyuncs.com/ns/app:v1 \
  --Replicas 2 \
  --Cpu 1000 \
  --Memory 2048 \
  --VpcId vpc-xxx \
  --VSwitchId vsw-xxx \
  --SecurityGroupId sg-xxx

# Deploy a new version
aliyun sae DeployApplication \
  --AppId app-xxx \
  --ImageUrl registry.cn-hangzhou.aliyuncs.com/ns/app:v2

# Scale replicas
aliyun sae RescaleApplication \
  --AppId app-xxx \
  --Replicas 4
```

**The Field Explosion Problem**: The `CreateApplication` API accepts over 60 parameters. Health checks, update strategies, environment variables, JVM options, and log configurations are all passed as JSON strings within the CLI arguments. Building these JSON payloads by hand is error-prone and difficult to review.

**The State Problem**: The CLI creates and modifies resources but does not track desired state. There is no built-in mechanism to compare the current application configuration against a desired specification. Configuration drift — where the running application diverges from the intended configuration — is invisible without manual inspection.

**The Environment Problem**: Promoting an application from staging to production requires translating CLI commands between environments, changing VPC IDs, namespace IDs, image tags, and resource tiers. This translation is manual and error-prone.

**Key Advantage**: Scriptable and CI/CD-friendly. CLI commands can be embedded in shell scripts and version-controlled.

**Verdict**: Suitable for automated deployments in CI/CD pipelines where state tracking is handled externally. Not ideal for managing the full lifecycle of applications across environments.

### Level 2: Infrastructure as Code (Terraform, Pulumi)

IaC tools represent the standard approach for managing SAE applications as code. They provide:

- **Declarative state**: Define the desired application configuration; the tool calculates and applies the diff
- **State tracking**: Know exactly what was provisioned and detect configuration drift
- **Dependency management**: Handle the creation order of VPC, VSwitch, security group, and application
- **Plan/preview**: See exactly what will change before applying

#### Terraform / OpenTofu

The `alicloud_sae_application` resource wraps the SAE application API:

```hcl
resource "alicloud_sae_application" "main" {
  app_name     = "order-service"
  package_type = "FatJar"
  replicas     = 3
  cpu          = 4000
  memory       = 8192

  package_url     = "https://bucket.oss-cn-hangzhou.aliyuncs.com/order-service.jar"
  package_version = "2.1.0"
  jdk             = "Open JDK 17"

  vpc_id            = alicloud_vpc.main.id
  vswitch_id        = alicloud_vswitch.main.id
  security_group_id = alicloud_security_group.main.id

  liveness_v2 {
    http_get {
      path = "/actuator/health/liveness"
      port = 8080
    }
    initial_delay_seconds = 30
    period_seconds        = 30
  }
}
```

**Immutability Constraints**: Several fields trigger application recreation when changed: `app_name`, `package_type`, `vpc_id`, `namespace_id`, and `programming_language`. IaC tools surface these as "force new" in their plan output, making the consequence visible before it happens.

**Environment Variables Format**: The SAE API expects environment variables as a JSON array (`[{"name":"K","value":"V"},...]`), not a native map. Both the Terraform and Pulumi modules handle this conversion internally, accepting a simple map and serializing it to the required format.

**Health Check Versioning**: The SAE provider has two generations of health check arguments. The current version uses `liveness_v2` and `readiness_v2` (structured blocks with `http_get`, `tcp_socket`, and `exec` sub-blocks). The legacy `liveness` and `readiness` fields accept raw JSON strings. The OpenMCF modules use the v2 variants exclusively.

**Tags**: The SAE provider supports tags on applications. The OpenMCF modules merge user-provided tags with standard metadata tags (resource name, kind, organization, environment) so that every application is identifiable through tag-based queries.

#### Pulumi

The Pulumi Go SDK wraps the same provider:

```go
app, err := sae.NewApplication(ctx, "order-service", &sae.ApplicationArgs{
    AppName:     pulumi.String("order-service"),
    PackageType: pulumi.String("FatJar"),
    Replicas:    pulumi.Int(3),
    Cpu:         pulumi.IntPtr(4000),
    Memory:      pulumi.IntPtr(8192),
    // ...
})
```

The Pulumi module provides the same functionality as the Terraform module but with Go type safety and the ability to compose with other Pulumi resources programmatically.

**Verdict**: The correct choice for managing SAE applications in production. Both Terraform and Pulumi provide the state management, plan/preview, and dependency resolution that SAE applications require.

### Level 3: Declarative Control Plane (OpenMCF)

OpenMCF wraps the IaC layer with a Kubernetes-like resource model:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudSaeApplication
metadata:
  name: order-service
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  appName: order-service
  packageType: FatJar
  replicas: 3
  cpu: 4000
  memory: 8192
  # ...
```

The manifest is applied through `openmcf apply -f`, which handles IaC execution, state management, and output capture. The same manifest format works with both the Pulumi and Terraform backends.

---

## Production Architecture Considerations

### Namespace Isolation

SAE namespaces provide logical isolation within a region. Applications in the same namespace share ConfigMaps and can discover each other through SAE's built-in service registration. Applications in different namespaces are isolated.

**Namespace Strategy**:
- One namespace per environment (dev, staging, production) prevents configuration leakage between environments
- The namespace ID format (`{region}:{short_id}`) ties the namespace to a specific region
- Namespace assignment is immutable after application creation

**OpenMCF Approach**: The `namespaceId` field in the spec maps directly to the provider's `namespace_id`. If omitted, the application is created in the default namespace. OpenMCF does not create namespaces — they must exist before the application is deployed.

### Package Type Selection

The `packageType` field determines the entire deployment pipeline for the application. This is the most consequential immutable decision:

| Package Type | Source | Runtime Requirements | Use Case |
|---|---|---|---|
| `Image` | Container registry URL | None (self-contained) | Containerized applications, polyglot services |
| `FatJar` | OSS or HTTP URL to JAR | JDK version selection | Spring Boot, Quarkus, Micronaut applications |
| `War` | OSS or HTTP URL to WAR | JDK + web container | Legacy Java web applications |
| `PythonZip` | OSS or HTTP URL to ZIP | Python runtime | Python Flask, FastAPI, Django applications |
| `PhpZip` | OSS or HTTP URL to ZIP | PHP runtime | PHP applications |

**Recommendation**: Choose `Image` for new applications. Container images are the most flexible packaging format — they work across all languages, include all dependencies, and can be built and tested identically in CI and production. The other package types exist for teams migrating existing applications that are not yet containerized.

### VPC Networking

SAE applications can run in two networking modes:

1. **SAE-managed networking**: The application runs in SAE's internal network. Suitable for applications that only need internet access (outbound) and do not need to communicate with VPC-resident resources.

2. **VPC-based networking**: The application runs inside a user-provided VPC, VSwitch, and security group. Required for applications that access RDS, Redis, MongoDB, NAS, or other services deployed in a VPC.

**Critical**: The `vpcId` field is immutable after creation. An application created without VPC configuration cannot be moved into a VPC later — it must be recreated.

**OpenMCF Approach**: The `vpcId`, `vswitchId`, and `securityGroupId` fields support both direct values and foreign key references to other OpenMCF components:

```yaml
vpcId:
  value: vpc-abc123           # Direct value

vswitchId:
  ref:
    kind: AlicloudVswitch
    name: my-vswitch           # Reference to another OpenMCF resource
```

### Health Check Configuration

SAE supports three types of health checks, each available for both liveness and readiness probes:

| Check Type | Mechanism | Best For |
|---|---|---|
| HTTP GET | Sends GET request, expects 2xx/3xx | Web services with health endpoints |
| TCP Socket | Attempts TCP connection | Services without HTTP (gRPC, raw TCP) |
| Exec | Runs command in container, expects exit 0 | Custom validation logic |

**Liveness vs. Readiness**:
- **Liveness**: If the check fails repeatedly (exceeding `failureThreshold`), SAE restarts the instance. Use for detecting deadlocked or crashed processes.
- **Readiness**: Until the check passes, SAE does not route traffic to the instance. Use for detecting initialization completion and dependency availability.

**Timing Parameters**:

| Parameter | Purpose | Typical Value |
|---|---|---|
| `initialDelaySeconds` | Wait before first check after container start | 10-60s (depends on startup time) |
| `periodSeconds` | Interval between checks | 10-30s |
| `timeoutSeconds` | Max wait for a check response | 3-10s |
| `failureThreshold` | Consecutive failures before action | 3-5 |
| `successThreshold` | Consecutive successes to recover (readiness only) | 1-2 |

**Common Pitfall**: Setting `initialDelaySeconds` too low for Java applications that have long startup times (30-90 seconds for Spring Boot with database migrations). The liveness probe kills the instance before it finishes starting, creating a restart loop.

### Update Strategies

SAE supports two rolling update strategies:

**BatchUpdate**: Releases new instances in sequential batches. The `batch` parameter controls how many batches the release is split into, and `batchWaitTime` controls the pause between batches.

**GrayBatchUpdate**: Canary-style release where a small percentage of instances are updated first, allowing observation before proceeding to the remaining instances.

| Parameter | Description |
|---|---|
| `type` | `BatchUpdate` or `GrayBatchUpdate` |
| `batch` | Number of batches (e.g., 2 for 50/50, 4 for 25/25/25/25) |
| `batchWaitTime` | Seconds to wait between batches |
| `releaseType` | `auto` (proceed automatically) or `manual` (pause for approval) |

**Interaction with `minReadyInstances`**: The `minReadyInstances` field sets a floor on available instances during deployments. If `replicas` is 4 and `minReadyInstances` is 3, SAE ensures at least 3 instances are serving traffic at all times during the rollout.

### Logging and Observability

SAE integrates with Alibaba Cloud Simple Log Service (SLS) for log collection. The `slsConfigs` field accepts a JSON array specifying which log directories to collect:

```json
[{"logDir":"/var/log/app","logType":"file_log"}]
```

**Log Types**:
- `stdout`: Captures stdout/stderr output
- `file_log`: Captures log files from specified directories

**OpenMCF Approach**: The `slsConfigs` field is passed through as a raw JSON string. OpenMCF does not parse or validate the SLS configuration — it is forwarded directly to the SAE API. This preserves compatibility with all SLS configuration options without requiring the proto schema to track SLS API changes.

---

## Best Practices

| Area | Recommendation | Rationale |
|---|---|---|
| Package type | Use `Image` for new applications | Most flexible, language-agnostic, consistent CI/CD |
| VPC | Always deploy in a VPC for production | Required for accessing VPC-resident databases and caches |
| Health checks | Configure both liveness and readiness | Prevents traffic to unready instances, restarts crashed ones |
| Replicas | Minimum 2 for production | Maintains availability during rolling updates and AZ failures |
| `minReadyInstances` | Set to `replicas - 1` or higher | Guarantees capacity during deployments |
| Update strategy | Use `BatchUpdate` with `batch: 2` | Balances deployment speed with safety |
| Graceful shutdown | Set `terminationGracePeriodSeconds` to 30 | Allows in-flight requests to complete |
| Environment variables | Use `envs` map, not hardcoded in images | Enables per-environment configuration without rebuilding |
| Tags | Apply team, cost-center, and environment tags | Enables cost allocation and resource filtering |
| Namespace | Use one namespace per environment | Isolates configuration and service discovery |
| JDK selection | Use Open JDK 17 or Dragonwell 17 for new Java apps | LTS versions with active security patches |
| Memory sizing | Match JVM heap (-Xmx) to ~75% of instance memory | Reserves memory for JVM overhead and OS |
| Image registry | Use ACR within the same region | Minimizes image pull latency and avoids cross-region transfer costs |
| Log collection | Configure `slsConfigs` for file-based logs | Stdout alone loses structured log context |

---

## What OpenMCF Supports

### The 80/20 Design

The `AlicloudSaeApplicationSpec` proto schema exposes the fields that cover the vast majority of production SAE deployments. The SAE API has over 60 parameters; the OpenMCF component exposes 31 fields organized into logical groups:

| Group | Fields | Coverage |
|---|---|---|
| Identity | `region`, `appName`, `appDescription`, `packageType` | Application identity and packaging |
| Compute | `replicas`, `cpu`, `memory` | Resource allocation |
| Networking | `vpcId`, `vswitchId`, `securityGroupId` | VPC placement |
| Deployment | `namespaceId`, `imageUrl`, `packageUrl`, `packageVersion`, `command`, `commandArgs`, `envs` | Application code and configuration |
| Runtime | `jdk`, `jarStartOptions`, `jarStartArgs`, `programmingLanguage`, `timezone`, `terminationGracePeriodSeconds`, `minReadyInstances`, `acrInstanceId` | Language-specific and lifecycle settings |
| Health | `liveness`, `readiness` | Health check probes |
| Hosts | `customHostAliases` | Custom DNS overrides |
| Update | `updateStrategy` | Rolling deployment configuration |
| Logging | `slsConfigs` | SLS log collection |
| Tags | `tags` | Resource tagging |

**What is NOT Exposed**:
- Auto-scaling rules (SAE scaling is managed separately through scaling rules)
- Custom SLB (Server Load Balancer) binding
- Microservice governance (service mesh, circuit breakers)
- Configuration management (SAE ConfigMaps via the console/API)
- PHP and Tomcat-specific configuration fields
- OSS mount configuration
- Pre-stop and post-start lifecycle hooks (beyond graceful shutdown)

These features are either rarely used, better managed through separate resources, or outside the scope of the single-application deployment model.

### Foreign Key References

Three fields support OpenMCF's foreign key mechanism:

| Field | Default Kind | Default Field Path |
|---|---|---|
| `vpcId` | `AlicloudVpc` | `status.outputs.vpc_id` |
| `vswitchId` | `AlicloudVswitch` | `status.outputs.vswitch_id` |
| `securityGroupId` | `AlicloudSecurityGroup` | `status.outputs.security_group_id` |

Foreign keys allow referencing other OpenMCF resources by name instead of hardcoding IDs:

```yaml
vpcId:
  ref:
    kind: AlicloudVpc
    name: production-vpc
```

When the manifest is applied, OpenMCF resolves the reference to the actual VPC ID from the referenced resource's stack outputs.

### Implementation Landscape

The component is implemented as two parallel IaC modules:

**Pulumi Module** (`iac/pulumi/`):
- Written in Go using `pulumi-alicloud` SDK v3
- Single `sae.Application` resource
- Converts `envs` map to JSON array format via `envsToJSON` helper
- Handles health check type mapping (liveness_v2, readiness_v2)
- Exports `app_id` and `app_name` as stack outputs

**Terraform Module** (`iac/tf/`):
- Uses `alicloud` provider `~> 1.200`
- Single `alicloud_sae_application` resource
- Dynamic blocks for liveness, readiness, custom host aliases, and update strategy
- `locals.tf` handles tag merging and environment variable JSON serialization
- Outputs `app_id` and `app_name`

Both modules accept the same proto-derived input schema and produce the same outputs, ensuring consistent behavior regardless of which IaC backend is used.

### Tag Management

Both modules automatically apply metadata tags to the SAE application:

| Tag | Source | Purpose |
|---|---|---|
| `resource` | `"true"` | Identifies OpenMCF-managed resources |
| `resource_name` | `metadata.name` | Resource name for filtering |
| `resource_kind` | `"alicloud_sae_application"` | Resource type for filtering |
| `resource_id` | `metadata.id` (if set) | Unique resource identifier |
| `organization` | `metadata.org` (if set) | Organization ownership |
| `environment` | `metadata.env` (if set) | Environment classification |

User-provided tags in `spec.tags` are merged with these base tags. In case of key conflicts, user-provided tags take precedence.

### Resource Lifecycle

The SAE application resource has several immutable fields that trigger recreation if changed:

| Immutable Field | Consequence of Change |
|---|---|
| `appName` | New application created, old one deleted |
| `packageType` | New application created — cannot switch between Image and JAR |
| `vpcId` | Cannot move between VPC and non-VPC mode |
| `namespaceId` | Cannot move between namespaces |
| `programmingLanguage` | Cannot change after creation |

All other fields (replicas, CPU, memory, image URL, environment variables, health checks, update strategy) can be updated in-place without recreation.

---

## CPU and Memory Tiers

SAE does not allow arbitrary CPU and memory values. Resources are allocated in discrete tiers:

**CPU (millicores)**:

| Tier | vCPU Equivalent | Typical Use Case |
|---|---|---|
| 500 | 0.5 vCPU | Minimal footprint, development |
| 1000 | 1 vCPU | Lightweight services |
| 2000 | 2 vCPU | Standard web applications |
| 4000 | 4 vCPU | Compute-intensive services |
| 8000 | 8 vCPU | Data processing, ML inference |
| 16000 | 16 vCPU | Heavy batch processing |
| 32000 | 32 vCPU | Maximum single-instance compute |

**Memory (MB)**:

| Tier | GB Equivalent |
|---|---|
| 1024 | 1 GB |
| 2048 | 2 GB |
| 4096 | 4 GB |
| 8192 | 8 GB |
| 12288 | 12 GB |
| 16384 | 16 GB |
| 24576 | 24 GB |
| 32768 | 32 GB |
| 65536 | 64 GB |
| 131072 | 128 GB |

Not all CPU-memory combinations are valid. The provider enforces specific pairings — consult the Alibaba Cloud SAE pricing documentation for the supported combinations in your region.

---

## Environment Variable Handling

The SAE API expects environment variables in a specific JSON format:

```json
[{"name":"PORT","value":"8080"},{"name":"LOG_LEVEL","value":"info"}]
```

OpenMCF accepts environment variables as a native map:

```yaml
envs:
  PORT: "8080"
  LOG_LEVEL: info
```

Both the Pulumi and Terraform modules convert this map to the required JSON array format internally. The Pulumi module uses the `envsToJSON` function in `locals.go`; the Terraform module uses a `jsonencode` expression in `locals.tf`.

---

## Comparison with Other Serverless Compute Options

| Feature | SAE Application | Function Compute (FC) | ECS Instance |
|---|---|---|---|
| Container support | Yes (primary model) | Yes (custom container) | Yes (manual) |
| Package formats | Image, JAR, WAR, ZIP | Image, ZIP, code | N/A |
| Long-running processes | Yes | Yes (up to 86400s) | Yes |
| Persistent connections | Yes (WebSocket, gRPC) | Limited | Yes |
| Built-in load balancing | Yes | Yes (via HTTP trigger) | No (requires SLB) |
| Auto-scaling | Yes (separate config) | Yes (built-in) | No (requires ESS) |
| VPC networking | Optional | Optional | Required |
| Minimum cost | Per-instance-second | Per-invocation | Per-hour |
| Health checks | Liveness + Readiness | Health check config | Manual |
| Rolling updates | Built-in | N/A (stateless) | Manual |

SAE is the appropriate choice when you need a long-running application with multiple replicas, health checks, rolling updates, and container-level control — without the operational overhead of managing a Kubernetes cluster.

---

## Conclusion

The `AlicloudSaeApplication` component reduces SAE application deployment from a 60-parameter API surface to a focused 31-field specification that covers the operational requirements of production deployments. The immutable fields (`appName`, `packageType`, `vpcId`, `namespaceId`, `programmingLanguage`) are called out in the proto schema documentation, the IaC plan output surfaces recreation risks before they happen, and the tag management system ensures every application is traceable through the standard OpenMCF metadata.

The two parallel IaC implementations (Pulumi and Terraform) provide identical functionality with different execution models, allowing teams to choose the tool that fits their existing workflows. Both modules handle the SAE API's quirks — JSON-formatted environment variables, discrete CPU/memory tiers, v2 health check blocks — so that the manifest author works with a clean, map-based interface.

For applications that need event-driven execution rather than long-running instances, see the `AlicloudFunction` component. For applications that require full Kubernetes control (custom operators, service mesh, node-level configuration), see the `AlicloudKubernetesCluster` and `AlicloudKubernetesNodePool` components.

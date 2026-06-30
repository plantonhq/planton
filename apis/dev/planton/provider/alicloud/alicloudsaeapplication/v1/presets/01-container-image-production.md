# Production Container Image Application

This preset creates a production SAE application deployed as a container image inside a VPC. Three replicas provide horizontal redundancy, with liveness and readiness HTTP probes ensuring traffic is only routed to healthy instances. Deployments use a 2-batch rolling update strategy to maintain availability during releases. A minimum of 2 ready instances is enforced at all times.

## When to Use

- Production microservices and web applications deployed as Docker images
- Workloads that need VPC connectivity to reach databases, caches, or internal services
- Applications requiring health check-driven traffic management and rolling deployments
- Teams that want container-based serverless without managing Kubernetes

## Key Configuration Choices

- **3 replicas** (`replicas: 3`) -- Provides N-1 redundancy: the application remains available even if one instance fails or is being restarted. Adjust based on traffic volume and latency requirements.
- **4 vCPU, 8 GB** (`cpu: 4000`, `memory: 8192`) -- Mid-tier sizing suitable for most production API and web workloads. SAE enforces discrete tiers; scale up to `cpu: 8000` / `memory: 16384` for heavier workloads.
- **VPC deployment** (`vpcId`, `vswitchId`, `securityGroupId`) -- Application instances run inside the specified VPC with security group isolation. Required for accessing private endpoints (databases, caches, internal services).
- **Liveness probe** (`liveness`) -- HTTP GET to the liveness path every 30 seconds. Three consecutive failures trigger an instance restart. The 30-second initial delay allows time for application startup.
- **Readiness probe** (`readiness`) -- HTTP GET to the readiness path every 10 seconds. Traffic is only routed to instances that pass this check. The 10-second initial delay is shorter than liveness to start accepting traffic quickly.
- **2-batch rolling update** (`updateStrategy`) -- Splits the 3 replicas into 2 deployment batches with 10 seconds between them. The `auto` release type proceeds automatically without manual approval. This ensures at least 1-2 instances are always serving traffic during deployments.
- **Minimum ready instances** (`minReadyInstances: 2`) -- SAE maintains at least 2 healthy instances at all times, including during deployments and scaling events.
- **30-second graceful shutdown** (`terminationGracePeriodSeconds: 30`) -- Instances receive SIGTERM and have 30 seconds to drain connections and clean up before SIGKILL.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<your-app-name>` | Application name (1-36 chars, alphanumeric + dashes) | Your naming convention |
| `<your-app-description>` | Human-readable description (max 1024 chars) | Your service documentation |
| `<your-container-image-url>` | Full image URL (e.g., `registry.cn-hangzhou.aliyuncs.com/ns/app:v1`) | Your container registry (ACR) |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID | `AliCloudVswitch` stack outputs |
| `<your-security-group-id>` | Security group ID | `AliCloudSecurityGroup` stack outputs |
| `<your-liveness-path>` | Liveness endpoint (e.g., `/health`, `/actuator/health/liveness`) | Your application's health check API |
| `<your-readiness-path>` | Readiness endpoint (e.g., `/ready`, `/actuator/health/readiness`) | Your application's health check API |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-service-name>` | Logical service name | Your service catalog |

## Related Presets

- **02-java-fatjar-production** -- Use for Java applications deployed as FatJar packages with JVM tuning
- **03-container-image-development** -- Use for development and testing with minimal resources and no VPC

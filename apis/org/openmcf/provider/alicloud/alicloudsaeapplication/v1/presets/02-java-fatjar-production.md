# Production Java FatJar Application

This preset creates a production SAE application deployed as a Java FatJar package with JVM tuning, Spring Boot Actuator health checks, and VPC connectivity. Three replicas with rolling updates ensure zero-downtime deployments. The JVM is configured with G1GC and a 4 GB heap within the 8 GB instance memory, leaving headroom for off-heap allocations, metaspace, and the OS.

## When to Use

- Spring Boot or other Java framework applications packaged as executable JARs
- Java microservices that need VPC access to databases, caches, or message brokers
- Production workloads using Spring Actuator health endpoints for lifecycle management
- Teams that prefer JAR deployment over container images for simpler CI/CD pipelines

## Key Configuration Choices

- **FatJar package type** (`packageType: FatJar`) -- SAE downloads and runs the JAR directly. No Dockerfile or container registry required. The JAR URL can point to an OSS bucket or any HTTP endpoint.
- **Open JDK 17** (`jdk: Open JDK 17`) -- LTS release with modern language features (records, sealed classes, pattern matching). Change to `Dragonwell 17` for Alibaba Cloud's optimized JDK with better GC performance and diagnostic tools.
- **JVM tuning** (`jarStartOptions: "-Xms1g -Xmx4g -XX:+UseG1GC"`) -- 1 GB initial heap, 4 GB max heap with G1 garbage collector. The 4 GB max leaves 4 GB of the 8 GB instance memory for metaspace, thread stacks, native memory, and OS overhead. Adjust based on profiling.
- **60-second liveness initial delay** (`liveness.initialDelaySeconds: 60`) -- Java applications have longer startup times than containerized or interpreted runtimes. The 60-second delay prevents SAE from restarting instances that are still initializing the Spring context.
- **30-second readiness initial delay** (`readiness.initialDelaySeconds: 30`) -- Shorter than liveness because Spring Boot starts the readiness endpoint before the full application context is loaded.
- **Spring Actuator endpoints** (`/actuator/health/liveness`, `/actuator/health/readiness`) -- Standard Spring Boot health check paths that integrate with Kubernetes-style liveness and readiness probes. Requires `spring-boot-starter-actuator` in the application dependencies.
- **Asia/Shanghai timezone** (`timezone: Asia/Shanghai`) -- Ensures log timestamps and scheduled tasks use China Standard Time. Change for other deployment regions.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<your-app-name>` | Application name (1-36 chars) | Your naming convention |
| `<your-app-description>` | Human-readable description | Your service documentation |
| `<your-jar-url>` | OSS or HTTP URL to the FatJar (e.g., `https://bucket.oss-cn-hangzhou.aliyuncs.com/app-2.0.jar`) | Your CI/CD pipeline |
| `<your-jar-version>` | Version identifier (e.g., `2.0.0`) | Your release process |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID | `AliCloudVswitch` stack outputs |
| `<your-security-group-id>` | Security group ID | `AliCloudSecurityGroup` stack outputs |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-service-name>` | Logical service name | Your service catalog |

## Related Presets

- **01-container-image-production** -- Use for production workloads deployed as container images instead of JARs
- **03-container-image-development** -- Use for development and testing with minimal resources

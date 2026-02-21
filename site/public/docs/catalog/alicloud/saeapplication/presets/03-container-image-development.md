---
title: "Development Container Image Application"
description: "This preset creates a minimal SAE application for development and testing. A single replica with the smallest compute tier (0.5 vCPU, 1 GB) keeps costs low. SAE-managed networking is used instead of..."
type: "preset"
rank: "03"
presetSlug: "03-container-image-development"
componentSlug: "saeapplication"
componentTitle: "SaeApplication"
provider: "alicloud"
icon: "package"
order: 3
---

# Development Container Image Application

This preset creates a minimal SAE application for development and testing. A single replica with the smallest compute tier (0.5 vCPU, 1 GB) keeps costs low. SAE-managed networking is used instead of a dedicated VPC, and only a TCP socket readiness probe is configured for basic health checking. No liveness probe or update strategy is needed for a single-instance development deployment.

## When to Use

- Development and testing environments
- Proof-of-concept deployments and rapid iteration
- Applications that do not need VPC connectivity or complex health check logic
- Cost-sensitive environments where the smallest compute tier is sufficient

## Key Configuration Choices

- **Single replica** (`replicas: 1`) -- Minimum for a running application. No redundancy, but sufficient for development where availability is not critical.
- **Smallest compute tier** (`cpu: 500`, `memory: 1024`) -- 0.5 vCPU and 1 GB RAM. The lowest SAE tier, minimizing cost. Upgrade to `cpu: 1000` / `memory: 2048` if the application needs more resources.
- **SAE-managed networking** (no `vpcId`, `vswitchId`, `securityGroupId`) -- SAE provides default networking without requiring a dedicated VPC. Simpler setup, but the application cannot access private VPC resources. Add VPC fields if you need database or cache access.
- **TCP socket readiness** (`readiness.tcpSocket`) -- Checks if the application is listening on port 8080 via a TCP connection. Simpler than HTTP probes and works with any application framework. No liveness probe avoids unnecessary restarts during debugging sessions.
- **Debug logging** (`LOG_LEVEL: debug`) -- Development default. Provides verbose output for troubleshooting.
- **No update strategy** -- With a single replica, SAE replaces the instance in-place during deployments. No batch or canary strategy is needed.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-app-name>` | Application name (1-36 chars) | Your naming convention |
| `<your-container-image-url>` | Full image URL (e.g., `registry.cn-hangzhou.aliyuncs.com/ns/app:dev`) | Your container registry (ACR) |

## Related Presets

- **01-container-image-production** -- Use for production with 3 replicas, VPC, health checks, and rolling updates
- **02-java-fatjar-production** -- Use for Java applications deployed as FatJar packages

---
title: "Public Cloud Run Service"
description: "This preset deploys a publicly accessible Cloud Run service with unauthenticated access, scale-to-zero, and Gen 2 execution environment. It uses all recommended defaults from the spec and is the..."
type: "preset"
rank: "01"
presetSlug: "01-public-service"
componentSlug: "cloud-run"
componentTitle: "Cloud Run"
provider: "gcp"
icon: "package"
order: 1
---

# Public Cloud Run Service

This preset deploys a publicly accessible Cloud Run service with unauthenticated access, scale-to-zero, and Gen 2 execution environment. It uses all recommended defaults from the spec and is the fastest way to get an HTTP service running on Cloud Run.

## When to Use

- Public-facing web applications, APIs, or webhooks
- Services that should scale to zero when idle to minimize cost
- Any HTTP service that doesn't require authentication at the infrastructure level

## Key Configuration Choices

- **Public access** (`ingress: INGRESS_TRAFFIC_ALL`, `allowUnauthenticated: true`) -- reachable from the public internet without IAM auth
- **Gen 2 execution** (`executionEnvironment: EXECUTION_ENVIRONMENT_GEN2`) -- full Linux compatibility, network filesystem support
- **Scale-to-zero** (`replicas.min: 0`) -- no cost when idle; cold starts apply
- **1 vCPU / 512 MiB** -- cost-effective starting point; increase for compute-heavy workloads
- **80 max concurrency** -- Cloud Run handles up to 80 concurrent requests per instance before scaling out
- **5-minute timeout** -- sufficient for most HTTP requests; increase for long-running operations

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<container-image-repo>` | Container image repository (e.g., `us-docker.pkg.dev/project/repo/app`) | `GcpArtifactRegistryRepo` outputs or container registry |
| `<image-tag>` | Image tag (e.g., `1.0.0`, `latest`) | Your CI/CD pipeline |

## Related Presets

- **02-private-vpc-connected** -- Use for internal services that need VPC access and should not be publicly reachable

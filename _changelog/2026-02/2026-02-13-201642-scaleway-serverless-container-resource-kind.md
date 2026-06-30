# Scaleway Serverless Container Resource Kind (R18)

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Protobuf Schemas, Pulumi CLI Integration, Provider Framework

## Summary

Implemented ScalewayServerlessContainer (R18), the eighteenth Scaleway resource kind and second in the serverless tier. This composite resource bundles a container namespace, the container itself, and optional cron triggers into a single declarable unit. Introduces the **structured image message** pattern with `StringValueOrRef` on the registry endpoint, enabling infra-chart DAG edges between ScalewayContainerRegistry and serverless containers.

## Problem Statement / Motivation

Scaleway Serverless Containers provides a Container-as-a-Service (CaaS) platform for deploying pre-built Docker images without managing servers. While ScalewayServerlessFunction (R17) handles code-based serverless workloads, many production services are distributed as container images and need the CaaS deployment model -- configurable ports, health checks, protocol selection, and fine-grained scaling.

### Pain Points

- No Planton kind existed for deploying container images to Scaleway's serverless platform
- Container image references needed to compose with ScalewayContainerRegistry for infra-chart DAG edges
- gRPC services and HTTP/2 backends required h2c protocol support (not available in ServerlessFunction)
- Production containers needed health check and autoscaling threshold configuration

## Solution / What's New

### Composite Resource (3 Terraform Resource Types)

```
ScalewayServerlessContainer
├── scaleway_container_namespace (1x) -- grouping, region/project scope
├── scaleway_container (1x) -- image, port, scaling, health checks, env
└── scaleway_container_cron (0..Nx) -- optional scheduled triggers
```

### Structured Image Message (New Pattern)

The key design innovation in R18 is the `ScalewayServerlessContainerImage` message that splits the container image reference into three fields:

- `registry_endpoint` (StringValueOrRef) -- enables infra-chart DAG edges to ScalewayContainerRegistry
- `name` (string) -- image name within the registry
- `tag` (string) -- image tag

The IaC modules compose these into the full URL: `{registry_endpoint}/{name}:{tag}`.

This works with any OCI registry (Scaleway, Docker Hub, GHCR) -- Scaleway registries use `valueFrom`, external registries use plain `value`.

### Container-Specific Features (Not Available in R17 Functions)

- **Port exposure** -- configurable listening port (default 8080)
- **Protocol selection** -- HTTP/1.1 or h2c (HTTP/2 cleartext for gRPC)
- **Health checks** -- HTTP path probing with failure threshold and interval
- **Scaling options** -- concurrent requests, CPU usage, and memory usage thresholds
- **CPU limit** -- explicit vCPU allocation in milliCPU
- **Command/args override** -- override CMD and ENTRYPOINT without image rebuild
- **Local storage** -- configurable ephemeral storage limit

## Implementation Details

### Proto Schemas (4 files)

- `spec.proto` -- 23 spec fields + 3 enums + 7 nested messages (Image, Env, EnvVar, CronTrigger, HealthCheck, ScalingOption)
- `stack_outputs.proto` -- 3 outputs (container_id, namespace_id, domain_name)
- `api.proto` -- KRM wrapper with api_version `scaleway.planton.dev/v1`
- `stack_input.proto` -- target + provider config

### Pulumi Go Module (6 files)

Uses `containers.NewNamespace`, `containers.NewContainer`, `containers.NewCron` from the `scaleway/containers` subpackage (pulumiverse SDK v1.43.0). Image URL composed in Go with `fmt.Sprintf`.

### Terraform HCL Module (5 files)

Image URL composed in locals: `"${var.spec.image.registry_endpoint}/${var.spec.image.name}:${var.spec.image.tag}"`. Health check and scaling option as dynamic blocks. Cron triggers via `for_each`.

### Patterns Reused from R17 (ServerlessFunction)

- K8s-style environment variables (repeated name-value messages, not maps)
- Namespace-per-resource isolation
- Cron trigger bundling
- Privacy and HTTP option enums
- Tag building and export patterns

### New Patterns Introduced in R18

- **Structured image message with StringValueOrRef** -- splits registry endpoint from image name/tag
- **Health check configuration** -- HTTP path probing with failure threshold
- **Scaling option thresholds** -- replaces deprecated `max_concurrency`
- **Protocol enum** -- http1/h2c selection for gRPC support
- **Command/args override** -- CMD and ENTRYPOINT overrides

### Discoveries During Build

- Terraform provider uses `command` (singular) not `commands` (plural) -- the Pulumi SDK uses `Commands` (plural). Fixed during `terraform validate`.
- `local_storage_limit_mb` local was missing from locals.tf -- added during build verification.

## Benefits

- **18 of 19 Scaleway resource kinds complete** (95% of resource tier)
- **CaaS support** -- first Planton kind for container image deployment on Scaleway
- **Infra-chart composable** -- StringValueOrRef on registry endpoint creates DAG edges
- **gRPC-ready** -- h2c protocol support for modern service architectures
- **Production-grade** -- health checks and scaling thresholds for reliable operations

## Impact

### Users
- Can deploy any Docker/OCI container image to Scaleway Serverless with a single YAML manifest
- Health checks and scaling options provide production-grade container lifecycle management
- gRPC services supported via h2c protocol

### Infra Charts
- `image.registry_endpoint` creates DAG edge: ContainerRegistry -> ServerlessContainer
- `private_network_id` creates DAG edge: PrivateNetwork -> ServerlessContainer
- `domain_name` output enables downstream ScalewayDnsRecord CNAME records
- Ready for the `scaleway/serverless-environment` infra chart (IC02)

### Developers
- Structured image message is extensible (future: digest pinning, pull policy)
- Component-local messages avoid cross-component coupling with R17

## Related Work

- **R17: ScalewayServerlessFunction** -- Sibling resource in the serverless tier (code-based vs image-based)
- **R14: ScalewayContainerRegistry** -- Upstream dependency for the `image.registry_endpoint` StringValueOrRef
- **R16: ScalewayDnsRecord** -- Downstream consumer of `domain_name` output
- **IC02: scaleway/serverless-environment** -- Future infra chart that will compose R17 and R18

---

**Status**: Production Ready
**Files Created**: 25 new files (4 proto, 6 Pulumi Go, 5 Terraform HCL, 2 docs, plus generated stubs and BUILD.bazel)
**Files Modified**: `pkg/crkreflect/kind_map_gen.go` (auto-generated kind map registration)

# Scaleway Serverless Container

## Overview

The **ScalewayServerlessContainer** resource kind provides a declarative interface for deploying and managing serverless containers on Scaleway. This is a **composite** resource that bundles a container namespace, the container itself, and optional cron triggers into a single declarable unit.

Unlike ScalewayServerlessFunction (which deploys source code with a runtime), this resource deploys **pre-built Docker images** from any OCI-compatible registry -- Scaleway Container Registry, Docker Hub, GHCR, ECR, or any other registry.

## Key Features

- **Container image deployment** -- Deploy any Docker/OCI container image with structured image references enabling infra-chart composability
- **Composite resource** -- Automatically manages the container namespace, container, and scheduled triggers as a single lifecycle unit
- **Kubernetes-style environment variables** -- Variables and secrets defined as ordered name-value lists (not maps), preserving sort order and enabling future `valueFrom` extension
- **VPC connectivity** -- Optional Private Network attachment for secure access to databases, Redis, and other VPC resources
- **Auto-scaling** -- Configurable min/max scale with scale-to-zero support, plus fine-grained scaling thresholds (concurrent requests, CPU, memory)
- **Health checks** -- HTTP health check support for production-grade container lifecycle management
- **Protocol selection** -- HTTP/1.1 or h2c (HTTP/2 cleartext) for gRPC services
- **Command/args override** -- Override container CMD and ENTRYPOINT without rebuilding the image
- **Scheduled triggers** -- Inline cron triggers for recurring container invocations
- **Dual IaC backend** -- Deploy using either Pulumi (Go) or Terraform with identical specifications

## Scaleway Terraform Resource Mapping

| Planton Kind | Terraform Resources | Relationship |
|---|---|---|
| ScalewayServerlessContainer | `scaleway_container_namespace` + `scaleway_container` + `scaleway_container_cron` | 1:1:N (composite) |

## Architecture

```
ScalewayServerlessContainer
├── scaleway_container_namespace (1x, auto-created)
│   └── Groups the container, holds region/project scope
├── scaleway_container (1x)
│   └── Image, port, scaling, health checks, env vars, networking
└── scaleway_container_cron (0..Nx, optional)
    └── Scheduled triggers with cron expressions and JSON args
```

## Spec Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | Yes | -- | Scaleway region (e.g., "fr-par", "nl-ams", "pl-waw") |
| `image` | message | Yes | -- | Container image (registry_endpoint + name + tag) |
| `registry_sha256` | string | No | -- | Deployment trigger (any string, changes trigger redeploy) |
| `port` | uint32 | No | 8080 | Listening port exposed by the container |
| `privacy` | enum | Yes | -- | `public` (no auth) or `private` (token required) |
| `description` | string | No | "" | Human-readable description |
| `memory_limit_mb` | uint32 | No | 256 | Memory allocation in MB |
| `cpu_limit` | uint32 | No | auto | vCPU in milliCPU (0 = auto from memory) |
| `min_scale` | uint32 | No | 0 | Minimum always-running instances (0 = scale-to-zero) |
| `max_scale` | uint32 | No | 20 | Maximum concurrent instances |
| `timeout_seconds` | uint32 | No | 300 | Max request time per invocation (seconds) |
| `http_option` | enum | No | enabled | `enabled` (HTTP+HTTPS) or `redirected` (HTTP->HTTPS) |
| `protocol` | enum | No | http1 | `http1` (HTTP/1.1) or `h2c` (HTTP/2 cleartext) |
| `commands` | repeated string | No | -- | CMD override |
| `args` | repeated string | No | -- | ENTRYPOINT args override |
| `env` | message | No | -- | Environment variables and secrets (see below) |
| `private_network_id` | StringValueOrRef | No | -- | Private Network for VPC connectivity |
| `sandbox` | string | No | "" | Execution environment (e.g., "v1", "v2") |
| `health_check` | message | No | -- | HTTP health check configuration |
| `scaling_option` | message | No | -- | Autoscaling thresholds |
| `local_storage_limit_mb` | uint32 | No | 0 | Ephemeral local storage in MB |
| `deploy` | bool | No | true | Whether to start the container after provisioning |
| `cron_triggers` | repeated | No | [] | Scheduled cron triggers |

### Image Message

The `image` field is a structured message with `StringValueOrRef` on the registry endpoint for infra-chart composability:

```yaml
image:
  registry_endpoint:
    valueFrom:
      kind: ScalewayContainerRegistry
      name: my-registry
      fieldPath: status.outputs.endpoint
  name: my-app
  tag: v1.2.3
```

The IaC module composes these into the full URL: `rg.fr-par.scw.cloud/my-registry/my-app:v1.2.3`

For non-Scaleway registries, provide a plain value:

```yaml
image:
  registry_endpoint:
    value: ghcr.io/my-org
  name: my-service
  tag: latest
```

### Environment Variables

Environment variables use Kubernetes-style repeated name-value messages grouped in an `env` message:

```yaml
env:
  variables:
    - name: NODE_ENV
      value: production
  secrets:
    - name: DATABASE_URL
      value: postgresql://10.0.1.5:5432/mydb
```

### Health Check Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `path` | string | Yes | HTTP path to probe (e.g., "/health") |
| `failure_threshold` | uint32 | No | Consecutive failures before unhealthy (default: 3) |
| `interval_seconds` | uint32 | No | Seconds between probes (default: 30) |

### Scaling Option Fields

| Field | Type | Description |
|---|---|---|
| `concurrent_requests_threshold` | uint32 | Scale up when concurrent requests per instance exceed this |
| `cpu_usage_threshold` | uint32 | Scale up when CPU usage percentage exceeds this |
| `memory_usage_threshold` | uint32 | Scale up when memory usage percentage exceeds this |

## Stack Outputs

| Output | Description |
|---|---|
| `container_id` | Scaleway container UUID |
| `namespace_id` | Container namespace UUID |
| `domain_name` | Native Scaleway invoke domain (for custom domain CNAME) |

## Dependencies

**Upstream:**
- **ScalewayContainerRegistry** -- `image.registry_endpoint` references `status.outputs.endpoint` for registry DAG edges
- **ScalewayPrivateNetwork** -- `private_network_id` references `status.outputs.private_network_id` for VPC connectivity

**Downstream:**
- **ScalewayDnsRecord** -- Create CNAME records pointing to `domain_name` for custom domains

## Comparison with ScalewayServerlessFunction

| Aspect | ServerlessFunction (R17) | ServerlessContainer (R18) |
|---|---|---|
| Source | Runtime + handler + optional zip | Container image from any OCI registry |
| Port | Not applicable | Configurable (default 8080) |
| Protocol | Not applicable | HTTP/1.1 or h2c |
| Health checks | Not supported | HTTP health checks |
| CPU control | Memory-proportional | Explicit vCPU limit |
| Scaling triggers | Simple min/max | Concurrent requests, CPU, memory thresholds |
| Best for | Event-driven functions, webhooks | Long-running HTTP services, APIs, gRPC services |

## Important Constraints

### Namespace Lifecycle
The namespace is an implementation detail -- users interact with the container as a single resource. One namespace is created per ScalewayServerlessContainer for clean isolation.

### Name Immutability
Changing the container `name` (from `metadata.name`) or namespace name triggers resource recreation. Plan accordingly.

### Secret Lifecycle
Secret environment variables are ignored in Terraform's change detection (`lifecycle.ignore_changes`) to prevent unnecessary updates when secrets are managed externally.

### Image Composition
The full image URL is composed by the IaC module from the three `image` fields. The registry endpoint, image name, and tag are combined as `{registry_endpoint}/{name}:{tag}`.

### No Built-in Custom Domains
Custom domain binding is handled by creating ScalewayDnsRecord CNAME records pointing to the container's `domain_name` output. The `scaleway_container_domain` Terraform resource exists but is not bundled to avoid overlapping with the DNS tier.

## Scaleway Documentation

- [Scaleway Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/)
- [Container Limitations](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/)
- [Terraform: scaleway_container](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/container)
- [Terraform: scaleway_container_namespace](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/container_namespace)
- [Terraform: scaleway_container_cron](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/container_cron)
- [Pulumi: scaleway.containers.Container](https://www.pulumi.com/registry/packages/scaleway/api-docs/containers/container/)

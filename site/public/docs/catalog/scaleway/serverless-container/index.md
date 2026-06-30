---
title: "Serverless Container"
description: "Serverless Container deployment documentation"
icon: "package"
order: 100
componentName: "scalewayserverlesscontainer"
---

# Scaleway Serverless Container

Deploys a Scaleway serverless container as a composite resource, provisioning a container namespace, the container itself, and optional cron triggers in a single declarable unit. The container runs a pre-built OCI image from any compatible registry and auto-scales based on incoming traffic.

## What Gets Created

When you deploy a ScalewayServerlessContainer resource, Planton provisions:

- **Container Namespace** — a `containers.Namespace` resource that groups the container for lifecycle isolation (one namespace per container)
- **Serverless Container** — a `containers.Container` resource running the specified OCI image with configurable CPU, memory, scaling, health checks, and networking
- **Cron Triggers** (0..N) — optional `containers.Cron` resources that invoke the container on UNIX CRON schedules with JSON arguments

The namespace is an internal implementation detail. Users interact with the container as a single resource. Environment variables and secrets are set directly on the container.

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A pre-built container image** pushed to an OCI-compatible registry (Scaleway Container Registry, Docker Hub, GHCR, or any other)
- **A valid Scaleway region** — serverless containers are available in `fr-par`, `nl-ams`, and `pl-waw`

## Quick Start

Create a file `serverless-container.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessContainer
metadata:
  name: my-api
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayServerlessContainer.my-api
spec:
  region: fr-par
  image:
    registryEndpoint:
      value: rg.fr-par.scw.cloud/my-registry
    name: my-api
    tag: latest
  privacy: public
  port: 8080
```

Deploy:

```shell
planton apply -f serverless-container.yaml
```

This creates a publicly accessible serverless container in the Paris region, listening on port 8080. After deployment, the container is reachable at the `domainName` shown in stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region where the container namespace and container are deployed. Valid values: `"fr-par"`, `"nl-ams"`, `"pl-waw"`. | Required, non-empty |
| `image` | `object` | Container image to deploy. See sub-fields below. | Required |
| `image.registryEndpoint` | `StringValueOrRef` | Base URL of the container registry. Use `value` for a literal string or `valueFrom` to reference a ScalewayContainerRegistry output. | Required |
| `image.name` | `string` | Image name within the registry (e.g., `"my-app"`, `"backend/api"`). | Required, non-empty |
| `image.tag` | `string` | Image tag to deploy (e.g., `"v1.2.3"`, `"latest"`, `"sha-abc1234"`). | Required, non-empty |
| `privacy` | `enum` | Endpoint authentication mode. `"public"` — no authentication required. `"private"` — requires a valid token. | Required, must be specified |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | `uint32` | `8080` | Listening port exposed by the container. Scaleway routes HTTP traffic to this port. |
| `description` | `string` | `""` | Human-readable description. Displayed in the Scaleway console. |
| `memoryLimitMb` | `uint32` | `256` | Memory allocated per container instance in megabytes. Common values: 128, 256, 512, 1024, 2048, 4096. |
| `cpuLimit` | `uint32` | auto | vCPU allocated per instance in milliCPU units (1000 = 1 vCPU). When omitted, Scaleway auto-allocates proportional to memory. Common values: 70, 140, 280, 560, 1120. |
| `minScale` | `uint32` | `0` | Minimum always-running instances. Set to 0 for scale-to-zero (no compute charges when idle). Set to 1+ to eliminate cold starts. |
| `maxScale` | `uint32` | `20` | Maximum concurrent container instances. The container auto-scales up to this limit. |
| `timeoutSeconds` | `uint32` | `300` | Maximum request processing time in seconds before Scaleway terminates the request. |
| `httpOption` | `enum` | `"enabled"` | HTTP/HTTPS behavior. `"enabled"` — both HTTP and HTTPS accepted. `"redirected"` — HTTP redirects to HTTPS. |
| `protocol` | `enum` | `"http1"` | Communication protocol between the Scaleway gateway and the container. `"http1"` — HTTP/1.1. `"h2c"` — HTTP/2 cleartext (required for gRPC). |
| `commands` | `string[]` | `[]` | Overrides the image CMD. Example: `["node", "server.js"]`. |
| `args` | `string[]` | `[]` | Overrides the image ENTRYPOINT arguments. Example: `["--port", "8080"]`. |
| `env.variables` | `list` | `[]` | Non-secret environment variables. Each entry has `name` and `value` fields. Visible in logs and the Scaleway console. |
| `env.secrets` | `list` | `[]` | Encrypted environment variables. Each entry has `name` and `value` fields. Encrypted at rest and masked in the Scaleway console. |
| `privateNetworkId` | `StringValueOrRef` | unset | Connects the container to a Scaleway Private Network for VPC-internal communication. Use `valueFrom` to reference a ScalewayPrivateNetwork output. |
| `sandbox` | `string` | platform default | Execution environment. Common values: `"v1"` (standard), `"v2"` (enhanced security). |
| `healthCheck` | `object` | unset | HTTP health check configuration. Sub-fields: `path` (probe path), `failureThreshold` (consecutive failures before unhealthy), `intervalSeconds` (probe period). |
| `scalingOption` | `object` | unset | Autoscaling thresholds. Sub-fields: `concurrentRequestsThreshold`, `cpuUsageThreshold`, `memoryUsageThreshold`. Scaleway scales up when any threshold is breached. |
| `localStorageLimitMb` | `uint32` | platform default | Ephemeral storage per instance in megabytes. Lost when the instance stops. |
| `deploy` | `bool` | `true` | When true, the container is deployed immediately after provisioning. Set to false to pre-provision without starting. |
| `registrySha256` | `string` | `""` | Deployment trigger string. When this value changes, the container is redeployed. Typically an image digest, CI build number, or git SHA. |
| `cronTriggers` | `list` | `[]` | Scheduled triggers. Each entry has `name` (optional identifier), `schedule` (CRON expression), and `args` (JSON string passed to the container). |

## Examples

### Public HTTP API with Environment Variables

A public-facing REST API deployed from a Scaleway Container Registry image, configured with environment variables and secrets:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessContainer
metadata:
  name: order-api
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayServerlessContainer.order-api
spec:
  region: fr-par
  image:
    registryEndpoint:
      value: rg.fr-par.scw.cloud/my-registry
    name: order-api
    tag: v1.4.0
  privacy: public
  port: 8080
  memoryLimitMb: 512
  maxScale: 10
  timeoutSeconds: 60
  httpOption: redirected
  env:
    variables:
      - name: LOG_LEVEL
        value: info
      - name: SERVICE_NAME
        value: order-api
    secrets:
      - name: DATABASE_URL
        value: postgres://user:pass@db.example.com:5432/orders
      - name: API_KEY
        value: sk-abc123secret
```

### gRPC Service with Health Checks and Scaling

A private gRPC service using HTTP/2 cleartext protocol, custom health checks, request-based autoscaling, and a Private Network connection:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessContainer
metadata:
  name: inference-grpc
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayServerlessContainer.inference-grpc
spec:
  region: nl-ams
  image:
    registryEndpoint:
      value: ghcr.io/my-org
    name: inference-service
    tag: sha-7f3a2b1
  privacy: private
  port: 50051
  protocol: h2c
  memoryLimitMb: 2048
  cpuLimit: 1120
  minScale: 2
  maxScale: 20
  timeoutSeconds: 120
  deploy: true
  privateNetworkId:
    value: fr-par/11111111-1111-1111-1111-111111111111
  healthCheck:
    path: /grpc.health.v1.Health/Check
    failureThreshold: 3
    intervalSeconds: 15
  scalingOption:
    concurrentRequestsThreshold: 50
    cpuUsageThreshold: 70
  env:
    variables:
      - name: MODEL_CACHE_DIR
        value: /tmp/models
    secrets:
      - name: MODEL_API_TOKEN
        value: tok-inference-secret
```

### Scheduled Batch Job with Cron Triggers

A container that runs scheduled data processing tasks via cron triggers, with a command override and ephemeral local storage for temporary files:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayServerlessContainer
metadata:
  name: data-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayServerlessContainer.data-pipeline
spec:
  region: pl-waw
  image:
    registryEndpoint:
      value: docker.io/mycompany
    name: data-pipeline
    tag: v2.1.0
  privacy: private
  port: 8080
  memoryLimitMb: 4096
  cpuLimit: 1120
  minScale: 0
  maxScale: 5
  timeoutSeconds: 600
  localStorageLimitMb: 1024
  commands:
    - python
    - -m
    - pipeline.main
  args:
    - --workers
    - "4"
  registrySha256: sha256:abcdef1234567890
  env:
    variables:
      - name: PIPELINE_MODE
        value: batch
    secrets:
      - name: S3_ACCESS_KEY
        value: scw-access-key
      - name: S3_SECRET_KEY
        value: scw-secret-key
  cronTriggers:
    - name: hourly-sync
      schedule: "0 * * * *"
      args: '{"task": "sync", "source": "s3://input-bucket"}'
    - name: nightly-report
      schedule: "0 2 * * *"
      args: '{"task": "report", "format": "pdf", "recipients": ["team@example.com"]}'
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `containerId` | `string` | Unique identifier (UUID) of the deployed serverless container. Used for Scaleway API operations, CLI commands, and Terraform import. |
| `namespaceId` | `string` | Unique identifier (UUID) of the container namespace. Useful for managing additional resources that reference the namespace. |
| `domainName` | `string` | Native Scaleway HTTPS domain for invoking the container. Format: `"<name>-<id>.containers.fnc.<region>.scw.cloud"`. Use with ScalewayDnsRecord to create custom domain CNAME records. |

## Related Components

- [ScalewayContainerRegistry](/docs/catalog/scaleway/container-registry) — provides a managed OCI registry whose endpoint can be referenced via `image.registryEndpoint.valueFrom`
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — connects the container to a VPC for internal communication with databases, caches, and other services
- [ScalewayServerlessFunction](/docs/catalog/scaleway/serverless-function) — deploys source code with a runtime instead of pre-built container images, suited for lightweight event-driven functions
- [ScalewayDnsRecord](/docs/catalog/scaleway/dns-record) — creates custom domain CNAME records pointing to the container's `domainName` output

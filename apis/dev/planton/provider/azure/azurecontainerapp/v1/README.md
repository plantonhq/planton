# AzureContainerApp

An Azure Container App is a serverless container workload running inside an Azure Container Apps Managed Environment. It is the Azure equivalent of a Kubernetes Deployment or an ECS Service -- it defines one or more containers, their resource allocations, scaling rules, ingress configuration, and runtime secrets.

## Overview

The `AzureContainerApp` component provisions an `azurerm_container_app` resource, deploying a containerized workload into an existing Container App Environment. It is the **workload resource** for the `container-apps-environment` infra chart -- the layer that defines what actually runs inside the environment.

Every Container App runs inside an `AzureContainerAppEnvironment`. The Environment provides the shared networking boundary, logging, Dapr infrastructure, and compute capacity. The Container App defines the workload: which containers to run, how to scale them, how to expose them, and what secrets they need.

## Key Features

- **Full container support**: Main containers and init containers with CPU/memory allocation, environment variables, command/args overrides, and volume mounts
- **Comprehensive scaling**: HTTP, TCP, Azure Queue, and custom KEDA scale rules with configurable polling and cooldown intervals; scale-to-zero support (`min_replicas=0`)
- **Health probes**: Liveness, readiness, and startup probes with TCP, HTTP, or HTTPS transport
- **Secrets management**: Plain-text secrets and Azure Key Vault references with managed identity access
- **Private registries**: ACR, Docker Hub, GHCR authentication via username/password or managed identity
- **Ingress**: External or internal HTTP/TCP access with traffic splitting, IP security restrictions, CORS policy, and client certificate mode (mTLS)
- **Dapr sidecar**: Service-to-service invocation, pub/sub, state management via Dapr with HTTP or gRPC protocol
- **Managed identity**: SystemAssigned, UserAssigned, or both -- for Key Vault, ACR, Storage, and other Azure service authentication
- **Volumes**: EmptyDir (ephemeral) and AzureFile (persistent) volumes with container mount support
- **Revision model**: Single revision (default, simplest) or Multiple revision mode for blue-green and canary deployments
- **StringValueOrRef composability**: `resource_group`, `container_app_environment_id`, and `identity_ids` all support `valueFrom` for infra-chart wiring

## When to Use

- **Microservices**: Deploy individual services into a shared Container App Environment
- **Web APIs**: HTTP services with auto-scaling, health probes, and ingress
- **Background workers**: Queue-processing jobs with scale-to-zero (no cost when idle)
- **Event-driven apps**: KEDA-powered scaling on Kafka, Prometheus, Redis, Service Bus, cron, etc.
- **Blue-green / canary deployments**: Traffic splitting across revisions in Multiple mode
- **Dapr-enabled services**: Distributed applications using Dapr building blocks

## Relationship to AzureContainerAppEnvironment

```
AzureContainerAppEnvironment (Layer 1)
├── Networking, logging, compute capacity
└── AzureContainerApp (Layer 2) ← this resource
    ├── Containers, scaling, ingress
    └── Secrets, registries, identity
```

The environment_id from `AzureContainerAppEnvironment` is a required input to every `AzureContainerApp`. No region field is needed -- the Container App inherits its location from the environment.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) |
| `name` | string | Yes | - | App name (lowercase, hyphens, 2-32 chars) |
| `container_app_environment_id` | StringValueOrRef | Yes | - | Environment ARM ID (literal or AzureContainerAppEnvironment ref) |
| `revision_mode` | string | No | `"Single"` | `"Single"` or `"Multiple"` |
| `workload_profile_name` | string | No | - | Named profile from the environment |
| `max_inactive_revisions` | int | No | 100 | Max retained inactive revisions (0-100) |
| `containers` | repeated | Yes | - | Main containers (at least one) |
| `init_containers` | repeated | No | - | Init containers (run before main) |
| `volumes` | repeated | No | - | Volumes (EmptyDir or AzureFile) |
| `min_replicas` | int | No | `0` | Minimum replicas (0 = scale-to-zero) |
| `max_replicas` | int | No | `10` | Maximum replicas (1-300) |
| `cooldown_period_in_seconds` | int | No | `300` | Scale cooldown period |
| `polling_interval_in_seconds` | int | No | `30` | KEDA polling interval |
| `revision_suffix` | string | No | - | Manual revision suffix |
| `termination_grace_period_seconds` | int | No | `0` | Graceful shutdown wait (0-600) |
| `http_scale_rules` | repeated | No | - | HTTP concurrent request scaling |
| `tcp_scale_rules` | repeated | No | - | TCP concurrent connection scaling |
| `azure_queue_scale_rules` | repeated | No | - | Azure Storage Queue scaling |
| `custom_scale_rules` | repeated | No | - | Custom KEDA scalers |
| `secrets` | repeated | No | - | Plain-text or Key Vault secrets |
| `registries` | repeated | No | - | Private container registry credentials |
| `ingress` | message | No | - | HTTP/TCP ingress configuration |
| `dapr` | message | No | - | Dapr sidecar configuration |
| `identity` | message | No | - | Managed identity configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `container_app_id` | ARM resource ID of the Container App |
| `latest_revision_name` | Name of the latest active revision |
| `latest_revision_fqdn` | FQDN of the latest revision (direct access) |
| `outbound_ip_addresses` | Egress IPs for firewall allowlists |
| `ingress_fqdn` | App FQDN (only when ingress is configured) |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerApp
metadata:
  name: my-api
spec:
  resource_group: production-rg
  name: my-api
  container_app_environment_id:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: prod-env
      fieldPath: status.outputs.environment_id
  containers:
    - name: api
      image: myregistry.azurecr.io/my-api:v1.0.0
      cpu: 0.5
      memory: "1Gi"
  min_replicas: 1
  max_replicas: 10
  http_scale_rules:
    - name: http-requests
      concurrent_requests: "100"
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## Downstream Usage

This is a **leaf workload resource** -- no downstream Planton resources reference its outputs. The outputs are consumed directly by users, CI/CD pipelines, and external systems (DNS records, firewall rules, monitoring).

## Full Feature Set (No Omissions)

This component exposes the **complete** Azure Container App feature surface with 21 message types. Unlike some Planton components that apply 80/20 scoping, AzureContainerApp maps 1:1 to the Terraform/Pulumi resource. Every field in `azurerm_container_app` that is relevant to workload definition is represented.

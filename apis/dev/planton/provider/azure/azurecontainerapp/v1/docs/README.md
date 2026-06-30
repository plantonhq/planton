# AzureContainerApp: Research & Design Documentation

## Overview

Azure Container Apps is Microsoft's fully managed serverless container platform, built on Kubernetes, Envoy, KEDA, and Dapr. A **Container App** (`Microsoft.App/containerApps`) is the individual workload resource that runs inside a Container App Environment -- it defines what containers to run, how they scale, how they receive traffic, and what secrets they need.

This document captures the research, design decisions, and scoping rationale for the Planton `AzureContainerApp` component.

## Azure Container Apps Deployment Landscape

### What is a Container App?

A Container App is:
- A serverless container workload running inside a Managed Environment
- The equivalent of a Kubernetes Deployment + Service + Ingress in a single resource
- A revision-based deployment model where template changes create new revisions
- A KEDA-integrated auto-scaler that can scale from 0 to 300 replicas

Container Apps abstract away the Kubernetes control plane entirely. Users never interact with pods, deployments, services, or ingress controllers. Instead, they declare containers, scale rules, and ingress -- and the platform handles orchestration.

### Container App vs Container App Environment

| Concept | Container App Environment | Container App |
|---------|--------------------------|---------------|
| Azure resource | `Microsoft.App/managedEnvironments` | `Microsoft.App/containerApps` |
| Analogy | Kubernetes cluster / ECS cluster | Kubernetes Deployment / ECS Service |
| Scope | Networking, logging, compute | Containers, scaling, ingress, secrets |
| Cardinality | One per boundary | Many per environment |
| Planton kind | `AzureContainerAppEnvironment` | `AzureContainerApp` (this resource) |

The relationship is strictly 1:N. Every Container App must reference exactly one Environment. The Environment provides:
- VNet connectivity (shared by all apps)
- Log Analytics integration (centralized logging)
- Workload profiles (shared compute capacity)
- Dapr infrastructure (shared sidecar mesh)

The Container App provides:
- Container images and resource allocation
- Scaling rules (per-app KEDA configuration)
- Ingress (per-app HTTP/TCP routing)
- Secrets and registry credentials (per-app)
- Managed identity (per-app Azure service access)

### Comparison with Other Azure Compute

| Feature | Container Apps | AKS | App Service | Azure Functions |
|---------|---------------|-----|-------------|-----------------|
| Managed K8s | Yes (hidden) | Yes (visible) | No | No |
| Scale to zero | Yes | No | No (except Y1) | Yes (Y1/Flex) |
| GPU support | Yes (dedicated profiles) | Yes | No | No |
| VNet injection | Via environment | Required | Optional | Optional |
| Dapr built-in | Yes | Manual | No | No |
| Custom domains | Yes | Via Ingress | Yes | Yes |
| Container images | Yes | Yes | Limited | Limited |
| Init containers | Yes | Yes | No | No |
| Sidecar containers | Yes | Yes | Yes (preview) | No |
| Blue-green deploy | Yes (revisions) | Manual | Yes (slots) | Yes (slots) |
| Max scale | 300 replicas | Node limits | 30 instances | 200 instances |
| Pricing | Per-vCPU-second | Per-node | Per-instance | Per-execution |

### When Container Apps > AKS

- Microservice teams that don't need Kubernetes API access
- Event-driven workloads that benefit from KEDA auto-scaling with scale-to-zero
- Applications needing Dapr for state management, pub/sub, service invocation
- Teams wanting serverless containers without cluster management overhead
- Cost-sensitive workloads that benefit from scale-to-zero (pay only when running)

### When AKS > Container Apps

- Workloads needing direct Kubernetes API access (CRDs, operators)
- Complex networking requirements (CNI, network policies, service mesh)
- Helm chart-based deployments with existing charts
- Stateful workloads needing persistent volumes (Container Apps has limited PV support)
- Very large scale (>300 replicas per workload)

### When Container Apps > App Service

- Multi-container workloads (sidecar, init containers)
- KEDA-based auto-scaling (queue depth, custom metrics)
- Scale-to-zero requirement
- Dapr-enabled distributed applications
- Container-native workflows (Dockerfile, CI/CD image builds)

### When Container Apps > Azure Functions

- Long-running processes (Functions has execution time limits)
- Full container control (custom runtimes, system dependencies)
- sidecar patterns (auth proxy, log shipper)
- gRPC services
- Workloads that don't fit the Functions trigger model

## Scaling Architecture

### KEDA Integration

Container Apps uses KEDA (Kubernetes Event-Driven Autoscaling) under the hood. Scale rules map directly to KEDA scalers:

| Scale Rule Type | KEDA Scaler | Typical Use Case |
|----------------|-------------|-----------------|
| `http_scale_rules` | Built-in HTTP | Web services, REST APIs |
| `tcp_scale_rules` | Built-in TCP | TCP servers, WebSocket backends |
| `azure_queue_scale_rules` | `azure-queue` | Azure Storage Queue consumers |
| `custom_scale_rules` | Any KEDA scaler | Kafka, Service Bus, Prometheus, cron, etc. |

### Scale-to-Zero

Setting `min_replicas: 0` enables scale-to-zero. When all scale rule metrics are below their thresholds, the app scales down to zero replicas. This means:
- **No cost** when the app is idle
- **Cold start latency** when the first request arrives (typically 2-10 seconds)
- **Queue-based workers** are the ideal use case (scale from 0 when messages appear)

The `cooldown_period_in_seconds` (default: 300s) controls how long the scaler waits after the last scale event before evaluating rules again. For aggressive scaling, reduce this. For cost optimization with bursty traffic, increase it.

The `polling_interval_in_seconds` (default: 30s) controls how often KEDA evaluates scale rules. Lower values mean faster scale reactions but more API calls.

### Scale Ranges

- **Consumption plan**: 0-300 replicas
- **Dedicated workload profiles**: 0-300 replicas (shared with environment limits)
- Scale rules are evaluated independently -- the highest desired replica count wins

## Networking Modes

### External Ingress

The most common mode. The app gets a public FQDN and is accessible from the internet.

```
Internet → Azure Front Door/CDN (optional) → Container App FQDN → Envoy → Container
```

### Internal Ingress (Environment-Only)

When `external_enabled: false`, the app is only accessible from within the environment. Other Container Apps in the same environment can reach it via the app name.

```
Container App A → Container App B (name-based discovery, no public FQDN)
```

### No Ingress

When the `ingress` block is omitted entirely, the app has no HTTP/TCP endpoint. It's purely a background worker or event processor. Other apps in the environment can still invoke it via Dapr if configured.

### VNet-Injected Environment

When the environment is VNet-injected, Container Apps can access private resources (databases, storage, etc.) within the VNet. The app inherits this connectivity automatically.

## Secret Management Patterns

### Plain-Text Secrets

The simplest approach. Secret values are stored in the spec and passed to the Terraform/Pulumi resource. Suitable for development and non-sensitive configuration.

```yaml
secrets:
  - name: db-connection
    value: "Server=mydb;Database=mydb;..."
```

### Key Vault References

Production pattern. Secrets are stored in Azure Key Vault and referenced by URI. Requires a managed identity with Key Vault read access.

```yaml
secrets:
  - name: db-connection
    key_vault_secret_id: https://my-kv.vault.azure.net/secrets/db-connection
    identity: /subscriptions/.../userAssignedIdentities/app-identity
```

Key Vault references support versioned and versionless URIs:
- Versionless: `https://kv.vault.azure.net/secrets/name` (auto-rotates)
- Versioned: `https://kv.vault.azure.net/secrets/name/version` (pinned)

### Secret Reference Patterns

Secrets are referenced by name in three places:
1. **Container env vars**: `env: [{name: DB_URL, secret_name: db-connection}]`
2. **Registry credentials**: `registries: [{server: acr.io, password_secret_name: acr-pass}]`
3. **Scale rule auth**: `authentication: [{secret_name: queue-conn, trigger_parameter: connection}]`

## Identity Patterns

### SystemAssigned

Azure creates and manages an identity tied to the Container App. Simplest option. The identity is automatically deleted when the app is deleted.

```yaml
identity:
  type: SystemAssigned
```

### UserAssigned

References pre-created identities. Can be shared across multiple apps. Has an independent lifecycle -- not deleted when the app is deleted.

```yaml
identity:
  type: UserAssigned
  identity_ids:
    - /subscriptions/.../userAssignedIdentities/shared-identity
```

### Combined

Both SystemAssigned and UserAssigned simultaneously. Use when the app needs both a unique identity and access to shared identities.

```yaml
identity:
  type: "SystemAssigned,UserAssigned"
  identity_ids:
    - /subscriptions/.../userAssignedIdentities/shared-identity
```

### Common Identity Uses

| Azure Service | Identity Used For |
|--------------|-------------------|
| Azure Key Vault | Reading secrets via `key_vault_secret_id` |
| Azure Container Registry | Pulling images via registry `identity` |
| Azure Storage | Accessing blobs, queues, tables |
| Azure SQL | Token-based authentication |
| Azure Service Bus | KEDA scale rule authentication |

## Revision Model

### Single Revision Mode (Default)

Only one revision is active at a time. Each deployment creates a new revision that automatically replaces the old one. This is the simplest model and covers most use cases.

```yaml
revision_mode: Single  # or omit (default)
```

### Multiple Revision Mode

Multiple revisions can be active simultaneously. Traffic is distributed according to `traffic_weight` rules. This enables:

**Blue-Green Deployment**: Route 100% to the new revision after testing, then deactivate the old one.

**Canary Deployment**: Gradually shift traffic from the old revision to the new one (e.g., 90/10 → 80/20 → 50/50 → 0/100).

**A/B Testing**: Split traffic between two feature versions using labeled FQDNs.

```yaml
revision_mode: Multiple
ingress:
  traffic_weight:
    - revision_suffix: v2-stable
      percentage: 80
      label: stable
    - revision_suffix: v3-canary
      percentage: 20
      label: canary
```

### Revision Suffix

When `revision_suffix` is set, the revision name becomes `{app-name}--{suffix}`. When omitted, Azure auto-generates a random suffix. Explicit suffixes are useful for:
- Referencing specific revisions in traffic weights
- CD pipeline verification ("deploy succeeded if revision X is active")
- Human-readable revision names

## Health Probe Best Practices

### Liveness Probe

Detects when a container has entered an unrecoverable state (deadlock, memory leak). Failed probes trigger container restart.

- **Transport**: HTTP for web services, TCP for non-HTTP workloads
- **Path**: `/healthz` (convention) -- should verify the app process is running
- **initial_delay_in_seconds**: Set high enough for the app to start (5-30s)
- **Do NOT** check downstream dependencies (database, cache) -- that belongs in readiness

### Readiness Probe

Detects when a container is ready to receive traffic. Failed probes remove the container from load balancing (but don't restart it).

- **Path**: `/ready` -- should verify the app can serve requests (DB connected, cache warmed)
- **interval_seconds**: 5s (faster recovery from temporary issues)
- **success_count_threshold**: 2-3 (avoid flapping)

### Startup Probe

Disables liveness/readiness probes during startup. Use for slow-starting applications.

- **Path**: Same as liveness (`/healthz`)
- **failure_count_threshold**: High (e.g., 30) to allow long startup times
- **interval_seconds**: 3-5s (check frequently during startup)

## Full Feature Set (No 80/20 Scoping)

Unlike `AzureContainerAppEnvironment`, which applies 80/20 scoping to exclude niche fields, the `AzureContainerApp` component exposes the **complete** workload feature surface. This decision was made because:

1. **Container Apps are workload resources** -- every field matters to someone's use case
2. **21 message types** cover the full Terraform/Pulumi schema without gaps
3. **No write-only or redundant fields** -- every field is user-actionable
4. **No computed/auto-derived fields** -- unlike the environment (where `logs_destination` is auto-derived), the Container App has no fields that should be hidden from users
5. **Leaf resource** -- no downstream resources need a simplified interface

The full feature set includes:
- Containers (main + init) with full probe support
- All four scale rule types (HTTP, TCP, Azure Queue, Custom KEDA)
- Complete ingress configuration (transport, traffic weights, IP restrictions, CORS, mTLS)
- Secret management (plain-text + Key Vault references)
- Registry authentication (username/password + managed identity)
- Dapr sidecar configuration
- Managed identity (SystemAssigned + UserAssigned)
- Volumes (EmptyDir + AzureFile)
- Revision model (Single + Multiple with traffic splitting)

## Cost Considerations

### Consumption Plan Pricing

- **vCPU**: ~$0.000024/second per active vCPU
- **Memory**: ~$0.000003/second per GiB
- **Scale-to-zero**: No charge when replicas = 0
- **Free grant**: 180,000 vCPU-seconds and 360,000 GiB-seconds per subscription per month

### Dedicated Workload Profile Pricing

- Fixed cost per node (regardless of utilization)
- Better for predictable, always-on workloads
- Required for GPU workloads

### Cost Optimization Tips

1. **Scale-to-zero** for dev/staging and background workers (`min_replicas: 0`)
2. **Right-size containers** -- start with 0.25 vCPU / 0.5 GiB, increase only if needed
3. **Use Consumption plan** for bursty or low-utilization workloads
4. **Use dedicated profiles** only for high-utilization or GPU workloads
5. **Tune polling_interval** -- don't over-poll for cost-sensitive workloads (30s default is fine)
6. **Tune cooldown_period** -- longer cooldown prevents rapid scale-up/down cycles

## Terraform/Pulumi Method Comparison

### Terraform: `azurerm_container_app`

The Terraform azurerm provider (v4.x) exposes the full Container App surface:

**Key fields**: name, resource_group_name, container_app_environment_id, revision_mode, workload_profile_name, template (containers, init_containers, volumes, scale, revision_suffix), ingress, secret, registry, dapr, identity, tags.

**Computed outputs**: custom_revision_suffix, latest_revision_fqdn, latest_revision_name, outbound_ip_addresses.

### Pulumi: `containerapp.App`

The Pulumi Azure classic provider (v6) mirrors the Terraform schema 1:1 since it's generated from the same Terraform provider.

### Azure CLI: `az containerapp create`

```bash
az containerapp create \
  --name my-api \
  --resource-group my-rg \
  --environment my-env \
  --image myregistry.azurecr.io/my-api:v1 \
  --cpu 0.5 --memory 1Gi \
  --min-replicas 1 --max-replicas 10 \
  --ingress external --target-port 8080 \
  --registry-server myregistry.azurecr.io \
  --registry-username myuser \
  --registry-password $ACR_PASSWORD
```

## Infra Chart Integration

The `AzureContainerApp` is a Layer 2 resource in the `container-apps-environment` infra chart:

```
Layer 0: AzureResourceGroup
Layer 0: AzureVpc → AzureSubnet
Layer 0: AzureLogAnalyticsWorkspace
Layer 0: AzureUserAssignedIdentity (optional)
Layer 1: AzureContainerAppEnvironment
Layer 2: AzureContainerApp (this resource) ← one or more
```

All upstream dependencies are connected via `StringValueOrRef` fields (`resource_group`, `container_app_environment_id`, `identity_ids`), enabling the infra chart DAG to resolve deployment order automatically.

## Related Resources

- **AzureContainerAppEnvironment**: The hosting environment (required parent)
- **AzureResourceGroup**: Required container for all Azure resources
- **AzureUserAssignedIdentity**: Optional identity for Key Vault and ACR access
- **AzureKeyVault**: Source for Key Vault-referenced secrets

---

**Status**: Production Ready
**API Version**: azure.planton.dev/v1
**Terraform Resource**: `azurerm_container_app`
**Pulumi Resource**: `containerapp.App`

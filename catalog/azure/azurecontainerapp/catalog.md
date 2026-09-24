# Azure Container App

Deploys a serverless container workload inside an Azure Container Apps Managed Environment with configurable containers, KEDA-based auto-scaling, revision-based traffic splitting, Dapr sidecar integration, and managed identity support. Scale-to-zero is the default posture: with no minimum replica count the app costs nothing while idle and cold-starts on the first request.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Container App** -- a containerized workload running inside the specified Container App Environment, with configurable revision mode (Single or Multiple), container definitions, and scaling rules
- **Container Template** -- one or more main containers with CPU/memory allocation, environment variables (literal or secret-backed), health probes (liveness, readiness, startup), and volume mounts
- **Init Containers** -- created only when `initContainers` entries are configured; run-to-completion containers for database migrations, configuration generation, or asset downloads
- **Scale Rules** -- created only when scale rule entries are configured; KEDA-compatible rules for HTTP concurrent requests, TCP connections, Azure Queue depth, or custom scalers (Kafka, Prometheus, cron, etc.)
- **Ingress Configuration** -- created only when `ingress` is configured; HTTP/TCP ingress with traffic splitting across revisions, IP security restrictions, and CORS policy
- **Dapr Sidecar** -- created only when `dapr` is configured; enables service invocation, pub/sub messaging, state management, and bindings
- **Managed Identity** -- created only when `identity` is configured; SystemAssigned, UserAssigned, or both for credential-free access to Azure services
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Container App Environment** where the app will run. The environment provides the shared networking boundary, logging, and compute capacity. Provide the environment ID directly or reference an AzureContainerAppEnvironment Cloud Resource via ValueFromRef.
- **An Azure Resource Group** where the Container App will be created. Provide the name directly or reference an AzureResourceGroup Cloud Resource via ValueFromRef.
- **A User Assigned Identity** (optional) for credential-free access to Key Vault secrets and ACR registries. Provide the identity resource ID directly or reference an AzureUserAssignedIdentity Cloud Resource.

## Deploy

### Console

Open the deployment store, find **Azure Container App**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Service** preset in the [Presets](#presets) tab to pre-populate a working configuration for a publicly accessible HTTP service.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerApp
metadata:
  name: my-web-app
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  containerAppName: my-web-app
  containerAppEnvironmentId:
    valueFrom:
      name: prod-env
  containers:
    - name: app
      image: mcr.microsoft.com/k8se/quickstart:latest
      cpu: 0.5
      memory: "1Gi"
```

```shell
planton apply -f container-app.yaml
```

This creates a Container App in Single revision mode with one container, scale-to-zero enabled, and no ingress -- add an `ingress` block to expose the app externally. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Container App to its environment and resource group:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: prod-env
      fieldPath: status.outputs.environment_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and environment first, then provisions the Container App with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Container App. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Revision mode** -- `revisionMode` controls how deployments work. `SINGLE` (the unspecified default) replaces the previous revision on each update. `MULTIPLE` keeps multiple revisions active simultaneously with traffic splitting via `ingress.trafficWeight`, enabling blue-green and canary deployment patterns.

**Scaling strategy** -- `minReplicas` and `maxReplicas` control the replica range (0-300). Set `minReplicas: 0` for scale-to-zero (no cost when idle) or `minReplicas: 1+` to avoid cold starts. Add HTTP, TCP, Azure Queue, or custom KEDA scale rules to control when the app scales up.

**Ingress configuration** -- Without ingress, the app is only accessible within the environment via service discovery. Set `ingress.externalEnabled: true` for public internet access. Use `ingress.transport` to choose between `AUTO`, `HTTP`, `HTTP2` (for gRPC), or `TCP` (paired with `exposedPort`).

**Secrets and registries** -- Secrets can be plain-text values or Key Vault references (requiring managed identity). Private container registries authenticate via username/password or managed identity. ACR with managed identity eliminates the need for registry passwords.

**Workload profile** -- Omit `workloadProfileName` to use the default Consumption (serverless) profile. Set it to a named profile defined in the environment for dedicated compute, GPU access, or guaranteed resources.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| **AzureContainerAppEnvironmentStorage** (optional) | `volumes[].storageName` | `status.outputs.storage_name` |
| **AzureContainerRegistry** (optional) | `registries[].server` | `status.outputs.login_server` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `container_app_id` | Azure Resource Manager ID of the Container App | Azure Policy assignments, diagnostic settings |
| `container_app_name` | Name of the Container App | Scripting, az CLI commands |
| `latest_revision_name` | Name of the latest active revision | CD pipelines to verify which revision is deployed |
| `latest_revision_fqdn` | FQDN pointing directly to the latest revision | Debugging, bypassing traffic splitting |
| `outbound_ip_addresses` | Egress IP addresses used by the Container App | Firewall allowlists on external services |
| `ingress_fqdn` | Primary user-facing FQDN (when ingress is configured) | DNS CNAME records, API gateway backends |
| `custom_domain_verification_id` | Value for the `asuid.{domain}` TXT record | Proving domain ownership when binding a per-app custom domain |
| `identity_principal_id` | Object ID of the system-assigned identity (when enabled) | RBAC grants (AcrPull, Key Vault Secrets User) for keyless access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web service** -- Publicly accessible HTTP service with 1 minimum replica (no cold starts), HTTP auto-scaling at 100 concurrent requests, and liveness/readiness probes. Suitable for web applications and REST APIs. Start from the **Web Service** preset.

**Background worker** -- Queue-processing worker with no ingress, scale-to-zero when idle, and custom KEDA scale rules for Azure Service Bus or other message sources. No cost when the queue is empty. Start from the **Background Worker** preset.

**Enterprise API** -- Production API with User Assigned managed identity, Key Vault secrets, ACR authentication via identity, IP security restrictions, all three probe types, and graceful shutdown. Meets enterprise security requirements. Start from the **Enterprise API** preset.

## Works With

- [**Azure Resource Group**](/cloud-catalog/azure-resource-group) -- provides the resource group where the Container App is created
- [**Azure Container App Environment**](/cloud-catalog/azure-container-app-environment) -- provides the hosting environment with networking, logging, and compute capacity
- [**Azure User Assigned Identity**](/cloud-catalog/azure-user-assigned-identity) -- provides credential-free access to Key Vault, ACR, and other Azure services
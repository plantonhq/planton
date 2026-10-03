# Azure Container App Environment

Deploys a Container Apps Managed Environment that serves as the hosting boundary for Azure Container Apps, with configurable VNet injection, internal load balancing, zone redundancy, Log Analytics integration, and dedicated workload profiles for GPU or guaranteed compute. The network shape is the decision to get right first: the subnet, internal-load-balancer mode, and zone redundancy are all set at creation, and changing any of them destroys and recreates the environment -- along with every app running in it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Container App Environment** -- a managed execution boundary in the specified Azure region and resource group, providing shared networking, Dapr infrastructure, and logging for all Container Apps running inside it
- **VNet Integration** -- created only when `infrastructureSubnetId` is provided; injects the environment into a customer-managed VNet for private connectivity to databases, storage, and other VNet resources
- **Log Analytics Integration** -- created only when `logAnalyticsWorkspaceId` is provided; configures log-analytics as the logging destination for container app logs, enabling KQL querying and alerting
- **Workload Profiles** -- created only when `workloadProfiles` entries are configured; dedicated compute pools (D4, D8, E4, NC24-A100, etc.) alongside the default Consumption profile
- **Custom DNS Suffix Binding** -- created only when `customDomain` is configured; replaces the environment's generated default domain with your own suffix, backed by the wildcard PFX certificate you provide
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the environment will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A VNet Subnet** (optional, for VNet-injected environments) with a /21 or larger address space (minimum 2048 IPs). Provide the subnet resource ID directly or reference an AzureSubnet Infra Component via ValueFromRef.
- **A Log Analytics Workspace** (optional) for centralized log collection. Provide the workspace ID directly or reference an AzureLogAnalyticsWorkspace Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Container App Environment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Consumption Plan Environment** preset in the [Presets](#presets) tab for a minimal serverless environment, or the **Workload Profiles with VNet Integration** preset for production with dedicated compute and private networking.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironment
metadata:
  name: dev-env
  org: acme-corp
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: "acme-dev-rg"
  environmentName: dev-environment
```

```shell
planton apply -f container-app-env.yaml
```

This creates a Consumption-plan environment with Azure-managed networking, external access (apps can receive public internet traffic), and streaming-only logs -- no VNet injection, no workload profiles, no Log Analytics. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the environment to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  infrastructureSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: container-apps-subnet
      fieldPath: status.outputs.subnet_id
  logAnalyticsWorkspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: central-logs
      fieldPath: status.outputs.workspace_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, and Log Analytics workspace first, then provisions the environment with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Container App Environment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Networking mode** -- Without `infrastructureSubnetId`, Azure manages networking and apps get public endpoints. With a subnet, the environment is VNet-injected, enabling private connectivity. Set `internalLoadBalancerEnabled: true` to restrict all apps to VNet-only access (no public internet). The subnet must have a /21 or larger address space.

**Zone redundancy** -- Set `zoneRedundancyEnabled: true` to distribute the environment across availability zones for higher resilience. Requires VNet injection (`infrastructureSubnetId` must be set). Recommended for production workloads.

**Workload profiles** -- By default, all environments include the Consumption profile (serverless, pay-per-use). Add dedicated profiles (D4, D8, E4, NC24-A100) for workloads needing guaranteed CPU/memory, GPU access, or predictable performance with no cold starts. Container Apps select their profile by name.

**Logging** -- `logsDestination` selects the pipe: `LOG_ANALYTICS` pairs with `logAnalyticsWorkspaceId` (KQL querying, alerting, dashboards); `AZURE_MONITOR` routes through diagnostic settings configured separately (no workspace here). Leaving it unspecified follows the workspace: linked means log-analytics, unlinked means streaming-only (logs are never stored).

**Platform access and east-west TLS** -- `publicNetworkAccess: DISABLED` refuses public traffic at the platform level (pair with a private endpoint); the Azure default derives it from the network shape. `mutualTlsEnabled: true` encrypts and peer-authenticates all app-to-app traffic with Azure-managed per-app certificates, at some latency cost.

**Custom DNS suffix** -- the optional `customDomain` block replaces the generated `*.{region}.azurecontainerapps.io` default domain with your own suffix (apps become `{app}.{dnsSuffix}`), backed by a wildcard PFX certificate + password (both secret-referenced). One suffix per environment; per-app custom domains bind on the app itself.

**Managed identity** -- the optional `identity` block gives the ENVIRONMENT its own Entra identity for platform-level operations (pulling job images, Key Vault reads for the custom domain). Apps configure their own identity separately.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (optional) | `infrastructureSubnetId` | `status.outputs.subnet_id` |
| **AzureLogAnalyticsWorkspace** (optional) | `logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `environment_id` | Azure Resource Manager ID of the environment | AzureContainerApp / AzureContainerAppJob / environment satellites' `containerAppEnvironmentId` field |
| `environment_name` | Name of the environment | Scripting, az CLI commands |
| `default_domain` | Default domain for apps in this environment | DNS CNAME records, verifying connectivity |
| `static_ip_address` | Static IP (public or private depending on mode) | DNS A records, firewall allowlists |
| `platform_reserved_cidr` | CIDR reserved for environment infrastructure (VNet-injected only) | Network planning, avoiding address conflicts |
| `platform_reserved_dns_ip_address` | Internal DNS server IP for service discovery (VNet-injected only) | Custom DNS configuration |
| `docker_bridge_cidr` | Docker bridge network CIDR inside the infrastructure (VNet-injected only) | Diagnosing address-space overlaps |
| `custom_domain_verification_id` | Value for the `asuid.{domain}` TXT record | Proving domain ownership for the custom DNS suffix and per-app custom domains |
| `identity_principal_id` | Object ID of the system-assigned identity (when enabled) | RBAC grants (AcrPull, Key Vault Secrets User) for keyless platform operations |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Consumption environment** -- Minimal serverless environment with Azure-managed networking, no VNet injection, and Log Analytics for log collection. Suitable for development, staging, and cost-sensitive workloads that benefit from scale-to-zero. Start from the **Consumption Plan Environment** preset.

**Production VNet environment** -- VNet-injected environment with internal load balancer, zone redundancy, dedicated D4 workload profiles, and Log Analytics. Apps have private connectivity to databases and storage with no public internet exposure. Start from the **Workload Profiles with VNet Integration** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the environment is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the VNet subnet for VNet-injected environments
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- provides centralized log collection for container app logs
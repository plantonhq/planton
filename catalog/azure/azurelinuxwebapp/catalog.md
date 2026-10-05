# Azure Linux Web App

Deploys a Linux Web App on Azure App Service with configurable application stacks (.NET, Node.js, Python, PHP, Java, or Docker containers), VNet integration, managed identity, IP restrictions, CORS, logging, and storage mounts. The App Service Plan it binds to sets the feature ceiling -- always-on needs Basic or higher, VNet integration needs Standard or higher -- so the plan choice decides what the app can do before a single spec field is set.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Azure Linux Web App** -- a managed web application hosted on the referenced App Service Plan, with the configured runtime stack, app settings, and site configuration
- **Application Stack** -- the selected runtime (Node.js, Python, .NET, PHP, Java with Tomcat/JBoss, or a custom Docker container image) configured within the site
- **Managed Identity** -- created only when the `identity` block is provided; enables credential-free authentication to Azure services (Key Vault, Storage, ACR)
- **VNet Integration** -- configured only when `virtualNetworkSubnetId` is provided; routes outbound traffic through the specified subnet for private resource access
- **Connection Strings** -- created only when `connectionStrings` entries are provided; named database and service connections exposed as environment variables
- **Storage Mounts** -- created only when `storageMounts` entries are provided; Azure File Shares or Blob containers mounted as directories
- **IP Restrictions** -- created only when `ipRestrictions` entries are provided in `siteConfig`; controls which IPs, service tags, or subnets can access the app
- **CORS Configuration** -- created only when the `cors` block is provided in `siteConfig`; controls cross-origin request policies
- **Logging** -- created only when the `logs` block is provided; application logs, HTTP logs, failed request tracing, and detailed error messages
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the web app for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Web App will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **An Azure Service Plan** that provides compute resources. The plan's OS type must be Linux and its SKU tier determines available features (always-on requires Basic+, VNet integration requires Standard+). Provide the plan ID directly or reference an AzureServicePlan Infra Component via ValueFromRef.
- **A VNet subnet** (optional) delegated to `Microsoft.Web/serverFarms` for VNet integration. Required for private resource access.
- **Application Insights** (optional) for APM telemetry. Provide the connection string directly or reference an AzureApplicationInsights Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Linux Web App**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Node.js Web API** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLinuxWebApp
metadata:
  name: platform-api
  org: acme-corp
  env: prod
spec:
  region: "eastus"
  resourceGroup:
    value: "acme-prod-rg"
  webAppName: "acme-platform-api"
  servicePlanId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Web/serverFarms/acme-prod-plan"
  siteConfig:
    applicationStack:
      nodeVersion: "22-lts"
    alwaysOn: true
    healthCheckPath: "/health"
```

```shell
planton apply -f linux-web-app.yaml
```

This creates a Node.js 22 LTS web app with always-on and health check monitoring -- VNet integration, managed identity, IP restrictions, CORS, and logging are not configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Web App to upstream dependencies deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: production-plan
      fieldPath: status.outputs.service_plan_id
  applicationInsightsConnectionString:
    valueFrom:
      kind: AzureApplicationInsights
      name: production-insights
      fieldPath: status.outputs.connection_string
  virtualNetworkSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: webapp-subnet
      fieldPath: status.outputs.subnet_id
  keyVaultReferenceIdentityId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: webapp-identity
      fieldPath: status.outputs.identity_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, service plan, Application Insights, subnet, and identity first, then provisions the Web App with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Linux Web App. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Application stack** -- Exactly one runtime must be selected: Node.js (`nodeVersion`), Python (`pythonVersion`), .NET (`dotnetVersion`), PHP (`phpVersion`), Java (`javaVersion` + `javaServer` + `javaServerVersion`), or Docker (`docker` block with registry URL, image name, and tag). For custom or multi-language applications, use the Docker stack.

**Always-on and health checks** -- Enable `alwaysOn` on Standard+ plans to prevent cold starts. Set `healthCheckPath` to an endpoint returning 200-299; Azure removes unhealthy instances from the load balancer after `healthCheckEvictionTimeInMin` minutes of continuous failure.

**VNet integration** -- Set `virtualNetworkSubnetId` to route outbound traffic through a VNet subnet, enabling access to private databases, Redis, and other VNet-connected services. Enable `vnetRouteAllEnabled` in `siteConfig` to route all outbound traffic (including public internet) through the VNet. Not supported on Free and Shared tiers.

**Public vs. private access** -- By default, the Web App is publicly accessible. Set `publicNetworkAccessEnabled: false` for Private Endpoint-only access. Use `ipRestrictions` with `ipRestrictionDefaultAction: Deny` for allow-list-based access control.

**Managed identity and Key Vault references** -- Configure a system-assigned identity to authenticate with Azure services without credentials. Use `@Microsoft.KeyVault(SecretUri=...)` syntax in `appSettings` to fetch secrets from Key Vault at runtime using the managed identity or a dedicated `keyVaultReferenceIdentityId`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureServicePlan** | `servicePlanId` | `status.outputs.service_plan_id` |
| **AzureApplicationInsights** (optional) | `applicationInsightsConnectionString` | `status.outputs.connection_string` |
| **AzureSubnet** (optional) | `virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| **AzureUserAssignedIdentity** (optional) | `keyVaultReferenceIdentityId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `web_app_id` | Azure Resource Manager ID of the Web App | Azure Policy assignments, diagnostic settings |
| `default_hostname` | Default FQDN (`{name}.azurewebsites.net`) | DNS CNAME records, application configuration |
| `outbound_ip_addresses` | Outbound IPs currently used by the Web App | Database firewall rules, third-party API whitelists |
| `possible_outbound_ip_addresses` | Every outbound IP the app may use across scale operations | The COMPLETE set for firewall rules (the current set can grow into it) |
| `identity_principal_id` | Principal ID of the system-assigned managed identity | Azure RBAC role assignments (Key Vault, Storage, ACR) |
| `identity_tenant_id` | Tenant ID of the system-assigned managed identity | Azure RBAC configuration |
| `custom_domain_verification_id` | Domain verification ID for custom domain binding | DNS TXT record at `asuid.{custom-domain}` |
| `site_credential_name` | Basic-auth publishing username (secret -- masked in the console) | Web Deploy/FTP publishing; inert when both basic-auth toggles are off |
| `site_credential_password` | Basic-auth publishing password (secret -- masked in the console) | Web Deploy/FTP publishing; inert when both basic-auth toggles are off |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Node.js Web API** -- Node.js 22 LTS with health check monitoring, HTTP/2, CORS, and Application Insights telemetry. The standard starting point for REST APIs and web services. Start from the **Node.js Web API** preset.

**Docker container** -- Custom container image from Azure Container Registry with managed identity for credential-free image pulls and always-on mode. Use for custom runtimes or pre-built images from CI/CD pipelines. Start from the **Docker Container** preset.

**Enterprise private Web App** -- Python runtime on a Premium plan with VNet integration, IP restrictions (default deny), Key Vault secret references, diagnostic logging, and full security hardening. The standard pattern for compliance-sensitive production applications. Start from the **Enterprise Private Web App** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Web App is created
- [**Azure Service Plan**](/infra-catalog/azure-service-plan) -- provides the compute tier hosting the Web App
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- provides the APM telemetry connection string
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the VNet subnet for outbound traffic routing
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- provides the managed identity for Key Vault reference authentication
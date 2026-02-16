# Azure Linux Web App

Deploys an Azure Linux Web App -- a managed web hosting platform for running long-lived web applications, APIs, and containerized services on Azure App Service. Supports .NET, Node.js, Python, PHP, Java (with Tomcat, JBoss EAP, or embedded SE), and Docker containers with configurable managed identity, VNet integration, Application Insights telemetry, logging, IP restrictions, CORS, and connection strings.

## What Gets Created

When you deploy an AzureLinuxWebApp resource, OpenMCF provisions:

- **Linux Web App** -- an `appservice.LinuxWebApp` resource in the specified region and resource group, configured with the chosen application stack, operational settings, logging, and security configuration
- **Managed Identity** -- created only when `identity` is configured, provides credential-free authentication to Azure services
- **VNet Integration** -- created only when `virtualNetworkSubnetId` is set, routes outbound traffic through a VNet subnet
- **Azure Tags** -- resource metadata tags applied to the web app for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the web app will be created (can reference an AzureResourceGroup resource)
- **An Azure Service Plan** providing compute resources -- Basic (`B1`-`B3`) for dedicated compute, Standard (`S1`-`S3`) for autoscale and deployment slots, or Premium (`P1v3`-`P3v3`) for enhanced performance and zone redundancy
- **A globally unique app name** -- the name becomes the hostname `{name}.azurewebsites.net`

## Quick Start

Create a file `webapp.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: my-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureLinuxWebApp.my-web
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-web-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/serverFarms/my-plan
  siteConfig:
    applicationStack:
      nodeVersion: "20-lts"
```

Deploy:

```shell
openmcf apply -f webapp.yaml
```

This creates a Node.js 20 LTS Web App with HTTPS-only access, TLS 1.2, and 64-bit worker processes.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the web app. **ForceNew**. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. **ForceNew**. | Required |
| `name` | `string` | Globally unique app name. Becomes `{name}.azurewebsites.net`. **ForceNew**. | Required, 2-60 characters, pattern `^[a-zA-Z0-9][a-zA-Z0-9-]{0,58}[a-zA-Z0-9]$` |
| `servicePlanId` | `StringValueOrRef` | Service Plan providing compute resources. Can reference an AzureServicePlan resource via `valueFrom`. | Required |
| `siteConfig` | `object` | Site configuration containing the application stack. | Required |
| `siteConfig.applicationStack` | `object` | Runtime selection. Exactly one runtime: `dotnetVersion`, `nodeVersion`, `pythonVersion`, `phpVersion`, `javaVersion` (with `javaServer` + `javaServerVersion`), or `docker`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `httpsOnly` | `bool` | `true` | Redirect all HTTP to HTTPS. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. |
| `enabled` | `bool` | `true` | Enable/disable the web app without deleting it. |
| `clientAffinityEnabled` | `bool` | `false` | Enable ARR session affinity cookies. Use for stateful apps only. |
| `applicationInsightsConnectionString` | `StringValueOrRef` | -- | App Insights connection string. Can reference an AzureApplicationInsights resource via `valueFrom`. |
| `virtualNetworkSubnetId` | `StringValueOrRef` | -- | Subnet ID for VNet integration. Can reference an AzureSubnet resource via `valueFrom`. |
| `identity.type` | `string` | -- | Managed identity: `SystemAssigned`, `UserAssigned`, or `SystemAssigned,UserAssigned`. |
| `appSettings` | `map<string, string>` | `{}` | Application environment variables. |
| `connectionStrings` | `list` | `[]` | Named connection strings with `name`, `type`, and `value`. |
| `siteConfig.alwaysOn` | `bool` | -- | Keep app loaded in memory. Critical for Standard/Premium plans. |
| `siteConfig.healthCheckPath` | `string` | -- | Health check endpoint (e.g., `/health`). |
| `siteConfig.healthCheckEvictionTimeInMin` | `int` | -- | Minutes before unhealthy instance eviction (2-10). |
| `siteConfig.cors.allowedOrigins` | `string[]` | -- | CORS allowed origins. |
| `siteConfig.ipRestrictions` | `list` | `[]` | IP-based access restriction rules. |
| `logs.applicationLogs.fileSystemLevel` | `string` | `"Error"` | Log level: `Off`, `Error`, `Warning`, `Information`, `Verbose`. |
| `logs.httpLogs.retentionInMb` | `int` | `35` | HTTP log file size limit (25-100 MB). |
| `logs.failedRequestTracing` | `bool` | `false` | Capture detailed traces for failed requests. |
| `logs.detailedErrorMessages` | `bool` | `false` | Return detailed error pages. Disable in production. |

## Examples

### Node.js Web API

A Node.js 20 LTS Web App with Application Insights and health checks:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: node-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLinuxWebApp.node-api
spec:
  region: eastus
  resourceGroup: prod-rg
  name: node-api-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/serverFarms/prod-plan
  applicationInsightsConnectionString: "InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://eastus-0.in.applicationinsights.azure.com/"
  siteConfig:
    applicationStack:
      nodeVersion: "20-lts"
    alwaysOn: true
    healthCheckPath: /health
    http2Enabled: true
  appSettings:
    NODE_ENV: production
    DATABASE_URL: "postgresql://..."
  logs:
    applicationLogs:
      fileSystemLevel: Information
    httpLogs:
      retentionInMb: 50
      retentionInDays: 7
```

### Docker Container Web App

A containerized Web App with VNet integration and managed identity:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: docker-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLinuxWebApp.docker-web
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: docker-web-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/serverFarms/premium-plan
  virtualNetworkSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/webapp
  identity:
    type: SystemAssigned
  siteConfig:
    applicationStack:
      docker:
        registryUrl: https://myregistry.azurecr.io
        imageName: myorg/my-web-app
        imageTag: v2.0.0
    containerRegistryUseManagedIdentity: true
    alwaysOn: true
    healthCheckPath: /healthz
    vnetRouteAllEnabled: true
```

### Enterprise Private Web App

A Premium-tier Web App with private-only access, client certificate authentication, and comprehensive logging:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: private-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLinuxWebApp.private-web
spec:
  region: eastus
  resourceGroup: prod-rg
  name: private-web-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/serverFarms/premium-plan
  publicNetworkAccessEnabled: false
  clientCertificateEnabled: true
  clientCertificateMode: Required
  identity:
    type: SystemAssigned
  siteConfig:
    applicationStack:
      dotnetVersion: "8.0"
    alwaysOn: true
    healthCheckPath: /api/health
    ipRestrictionDefaultAction: Deny
  logs:
    applicationLogs:
      fileSystemLevel: Warning
    httpLogs:
      retentionInMb: 100
      retentionInDays: 30
    failedRequestTracing: true
```

### Using Foreign Key References

Reference OpenMCF-managed resources:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: ref-web
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLinuxWebApp.ref-web
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-web-app
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: my-plan
      field: status.outputs.plan_id
  applicationInsightsConnectionString:
    valueFrom:
      kind: AzureApplicationInsights
      name: my-insights
      field: status.outputs.connection_string
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
    alwaysOn: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `web_app_id` | `string` | Azure Resource Manager ID of the Web App |
| `default_hostname` | `string` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | `string[]` | Outbound IP addresses for firewall allowlisting |
| `identity_principal_id` | `string` | System-assigned identity principal ID (when identity is configured) |
| `identity_tenant_id` | `string` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | `string` | TXT record value for custom domain verification |
| `kind` | `string` | Resource kind (e.g., `app,linux`) |

## Related Components

- [AzureServicePlan](/docs/catalog/azure/azureserviceplan) -- provides the compute tier for the Web App
- [AzureApplicationInsights](/docs/catalog/azure/azureapplicationinsights) -- provides APM telemetry collection
- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for app placement
- [AzureSubnet](/docs/catalog/azure/azuresubnet) -- provides VNet integration for outbound connectivity
- [AzureFrontDoorProfile](/docs/catalog/azure/azurefrontdoorprofile) -- global CDN and load balancing for the Web App

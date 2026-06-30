# Azure Front Door Profile

Deploys an Azure Front Door profile with endpoints, origin groups, origins, and routes for global HTTP load balancing, SSL offloading, caching, and application acceleration. Front Door is a global resource deployed across all Microsoft edge locations worldwide. The component bundles all five resource types (profile, endpoints, origin groups, origins, routes) because they form a single coherent routing unit.

## What Gets Created

When you deploy an AzureFrontDoorProfile resource, Planton provisions:

- **Front Door Profile** -- a `cdn.FrontDoorProfile` resource in the specified resource group (global, no region), configured with the chosen SKU tier and response timeout
- **Endpoints** -- a `cdn.FrontDoorEndpoint` for each entry in `endpoints`, each assigned a public hostname (`*.azurefd.net`) for client traffic
- **Origin Groups** -- a `cdn.FrontDoorOriginGroup` for each entry in `originGroups`, configured with load balancing settings and optional health probes
- **Origins** -- a `cdn.FrontDoorOrigin` for each origin within an origin group, representing a backend server with priority, weight, and optional Private Link connectivity
- **Routes** -- a `cdn.FrontDoorRoute` for each entry in `routes`, connecting endpoints to origin groups via URL pattern matching with optional caching and HTTPS redirect
- **Azure Tags** -- resource metadata tags applied to the profile for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** for ARM organization (can reference an AzureResourceGroup resource)
- **Backend origins** -- one or more backend servers with public hostnames (App Service, Container Apps, Storage Account, custom server, etc.)
- **Premium SKU** if using Private Link to origins -- Standard SKU does not support private connectivity

## Quick Start

Create a file `frontdoor.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureFrontDoorProfile
metadata:
  name: my-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureFrontDoorProfile.my-cdn
spec:
  resourceGroup: my-rg
  name: my-cdn
  endpoints:
    - name: web
  originGroups:
    - name: web-backends
      origins:
        - name: primary
          hostName: myapp.azurewebsites.net
  routes:
    - name: default
      endpointName: web
      originGroupName: web-backends
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
```

Deploy:

```shell
planton apply -f frontdoor.yaml
```

This creates a Standard-tier Front Door profile with one endpoint, one origin group pointing to an App Service, and a catch-all route.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique profile name. **ForceNew**. | Required, 2-46 characters, pattern `^[a-zA-Z0-9][a-zA-Z0-9-]{0,44}[a-zA-Z0-9]$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sku` | `string` | `"Standard_AzureFrontDoor"` | SKU tier. `Standard_AzureFrontDoor` (global LB, SSL, caching) or `Premium_AzureFrontDoor` (adds Private Link to origins, enhanced WAF). **ForceNew**. |
| `responseTimeoutSeconds` | `int` | `120` | Origin response timeout. Range: 16-240 seconds. |
| `endpoints` | `list` | `[]` | Endpoints (entry points). Each has `name` (required, 2-46 chars) and optional `enabled` (default `true`). |
| `originGroups` | `list` | `[]` | Origin groups with load balancing. See origin group fields below. |
| `routes` | `list` | `[]` | Routes connecting endpoints to origin groups. See route fields below. |

**Origin group fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Origin group name (required) |
| `sessionAffinityEnabled` | `bool` | `true` | Enable sticky sessions via cookies |
| `loadBalancing.sampleSize` | `int` | `4` | Recent health probe samples to evaluate (0-255) |
| `loadBalancing.successfulSamplesRequired` | `int` | `3` | Successful samples for healthy status (0-255) |
| `loadBalancing.additionalLatencyInMilliseconds` | `int` | `50` | Latency tolerance for origin selection (0-1000ms) |
| `healthProbe.protocol` | `string` | -- | Probe protocol: `Http`, `Https` (required if probe is set) |
| `healthProbe.path` | `string` | `"/"` | Probe URL path |
| `healthProbe.requestType` | `string` | `"HEAD"` | Probe method: `HEAD`, `GET` |
| `healthProbe.intervalInSeconds` | `int` | -- | Probe interval (required, 1-255) |
| `origins` | `list` | `[]` | Backend origins in the group. See origin fields below. |

**Origin fields** (each entry in `origins`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Origin name (required). **ForceNew**. |
| `hostName` | `string` | -- | Backend hostname (required). |
| `certificateNameCheckEnabled` | `bool` | `true` | Validate origin SSL certificate hostname. |
| `originHostHeader` | `string` | -- | Host header override for multi-tenant backends. |
| `httpPort` | `int` | `80` | HTTP port (1-65535). |
| `httpsPort` | `int` | `443` | HTTPS port (1-65535). |
| `priority` | `int` | `1` | Failover priority (1-5, lower = preferred). |
| `weight` | `int` | `500` | Traffic weight within same priority (1-1000). |
| `enabled` | `bool` | `true` | Whether origin receives traffic. |
| `privateLink` | `object` | -- | Private Link config (Premium only). Has `location`, `privateLinkTargetId`, `requestMessage`, `targetType`. |

**Route fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Route name (required). **ForceNew**. |
| `endpointName` | `string` | -- | Target endpoint name (required). |
| `originGroupName` | `string` | -- | Target origin group name (required). |
| `patternsToMatch` | `string[]` | -- | URL patterns (e.g., `["/*"]`, `["/api/*"]`). |
| `supportedProtocols` | `string[]` | -- | Accepted protocols: `Http`, `Https`. |
| `forwardingProtocol` | `string` | `"MatchRequest"` | Origin protocol: `MatchRequest`, `HttpOnly`, `HttpsOnly`. |
| `httpsRedirectEnabled` | `bool` | `true` | Auto-redirect HTTP to HTTPS. |
| `linkToDefaultDomain` | `bool` | `true` | Associate with endpoint's *.azurefd.net hostname. |
| `enabled` | `bool` | `true` | Whether route processes traffic. |
| `cache` | `object` | -- | Cache config. Has `queryStringCachingBehavior`, `queryStrings`, `compressionEnabled`, `contentTypesToCompress`. |

## Examples

### Web Acceleration with Caching

A Standard-tier profile accelerating a web application with compression and caching enabled:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureFrontDoorProfile
metadata:
  name: web-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureFrontDoorProfile.web-cdn
spec:
  resourceGroup: prod-rg
  name: web-cdn
  endpoints:
    - name: web
  originGroups:
    - name: app-backends
      healthProbe:
        protocol: Https
        path: /healthz
        intervalInSeconds: 30
      origins:
        - name: primary
          hostName: myapp.azurewebsites.net
          originHostHeader: myapp.azurewebsites.net
  routes:
    - name: default
      endpointName: web
      originGroupName: app-backends
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
      cache:
        queryStringCachingBehavior: UseQueryString
        compressionEnabled: true
        contentTypesToCompress:
          - text/html
          - text/css
          - application/javascript
          - application/json
          - image/svg+xml
```

### Multi-Region API Gateway

A profile with multiple origins across regions for active-passive failover:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureFrontDoorProfile
metadata:
  name: api-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureFrontDoorProfile.api-gateway
spec:
  resourceGroup: prod-rg
  name: api-gateway
  responseTimeoutSeconds: 60
  endpoints:
    - name: api
  originGroups:
    - name: api-backends
      sessionAffinityEnabled: false
      loadBalancing:
        sampleSize: 4
        successfulSamplesRequired: 3
        additionalLatencyInMilliseconds: 50
      healthProbe:
        protocol: Https
        path: /api/health
        requestType: GET
        intervalInSeconds: 15
      origins:
        - name: eastus
          hostName: api-eastus.azurewebsites.net
          originHostHeader: api-eastus.azurewebsites.net
          priority: 1
          weight: 500
        - name: westeurope
          hostName: api-westeurope.azurewebsites.net
          originHostHeader: api-westeurope.azurewebsites.net
          priority: 1
          weight: 500
        - name: southeastasia-dr
          hostName: api-sea.azurewebsites.net
          originHostHeader: api-sea.azurewebsites.net
          priority: 2
          weight: 500
  routes:
    - name: api-route
      endpointName: api
      originGroupName: api-backends
      patternsToMatch:
        - "/api/*"
      supportedProtocols:
        - Https
      forwardingProtocol: HttpsOnly
      httpsRedirectEnabled: false
```

### Premium with Private Link

A Premium-tier profile connecting privately to an App Service backend:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureFrontDoorProfile
metadata:
  name: enterprise-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureFrontDoorProfile.enterprise-cdn
spec:
  resourceGroup: prod-rg
  name: enterprise-cdn
  sku: Premium_AzureFrontDoor
  endpoints:
    - name: secure-web
  originGroups:
    - name: private-backends
      healthProbe:
        protocol: Https
        path: /health
        intervalInSeconds: 30
      origins:
        - name: webapp
          hostName: myapp.azurewebsites.net
          originHostHeader: myapp.azurewebsites.net
          privateLink:
            location: eastus
            privateLinkTargetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/sites/myapp
            targetType: sites
  routes:
    - name: default
      endpointName: secure-web
      originGroupName: private-backends
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
```

### Using Foreign Key References

Reference an Planton-managed resource group:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureFrontDoorProfile
metadata:
  name: ref-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureFrontDoorProfile.ref-cdn
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-cdn
  endpoints:
    - name: web
  originGroups:
    - name: backends
      origins:
        - name: app
          hostName: myapp.azurewebsites.net
  routes:
    - name: default
      endpointName: web
      originGroupName: backends
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `profile_id` | `string` | Azure Resource Manager ID of the Front Door profile |
| `profile_name` | `string` | Name of the Front Door profile |
| `resource_guid` | `string` | Unique GUID assigned by Azure's Front Door service |
| `endpoint_ids` | `map<string, string>` | Map of endpoint names to their Azure Resource Manager IDs |
| `endpoint_hostnames` | `map<string, string>` | Map of endpoint names to their generated hostnames (e.g., `my-endpoint-abc123.z01.azurefd.net`). Use as CNAME targets for custom domains. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for profile placement
- [AzureDnsRecord](/docs/catalog/azure/azurednsrecord) -- creates CNAME records pointing to endpoint hostnames
- [AzureLinuxWebApp](/docs/catalog/azure/azurelinuxwebapp) -- common origin backend for web applications
- [AzureFunctionApp](/docs/catalog/azure/azurefunctionapp) -- common origin backend for serverless APIs
- [AzureStorageAccount](/docs/catalog/azure/azurestorageaccount) -- common origin backend for static content

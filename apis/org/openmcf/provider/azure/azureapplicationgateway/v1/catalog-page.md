# Azure Application Gateway

Deploys an Azure Application Gateway -- a Layer 7 (HTTP/HTTPS) load balancer and reverse proxy that provides SSL termination, host-based routing, cookie-based session affinity, custom health probes, and optional Web Application Firewall (WAF) protection. The component bundles the gateway with all structural sub-resources (frontend configuration, backend pools, listeners, routing rules, probes, and SSL certificates) into a single deployable unit.

## What Gets Created

When you deploy an AzureApplicationGateway resource, OpenMCF provisions:

- **Application Gateway** -- a `network.ApplicationGateway` resource in the specified region and resource group, configured with the chosen SKU tier, capacity or autoscale settings, and HTTP/2 support
- **Gateway IP Configuration** -- auto-derived from the resource name (as `{name}-gw-ip-config`), binding the gateway to the dedicated subnet
- **Frontend IP Configuration** -- auto-derived from the resource name (as `{name}-frontend-ip-config`), binding the gateway to the provided public IP address
- **Frontend Ports** -- auto-derived from listener definitions (as `{listener_name}-port`), one per HTTP listener
- **Backend Address Pools** -- one or more pools of backend targets identified by FQDN and/or IP address
- **Backend HTTP Settings** -- protocol, port, session affinity, timeout, and optional health probe configuration for backend communication
- **HTTP Listeners** -- entry points for incoming traffic, each bound to a port and protocol with optional host-based routing
- **Request Routing Rules** -- Basic-type rules connecting listeners to backend pools via backend HTTP settings
- **Health Probes** -- optional custom health checks for backend servers, referenced by backend HTTP settings
- **SSL Certificates** -- optional Key Vault-sourced certificates for HTTPS listeners
- **Identity** -- optional user-assigned managed identity block for Key Vault certificate access
- **WAF Configuration** -- optional Web Application Firewall using OWASP 3.2 rule set (requires WAF_v2 SKU)
- **Azure Tags** -- resource metadata tags applied to the gateway for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the gateway will be created (can reference an AzureResourceGroup resource)
- **A dedicated subnet** with no other resources deployed; Application Gateway v2 requires its own subnet. A /24 CIDR block is recommended for production (supports up to 125 instances plus 5 Azure-reserved addresses)
- **A Standard SKU public IP** with static allocation for the frontend (can reference an AzurePublicIp resource, which enforces Standard SKU and Static allocation)
- **A user-assigned managed identity** with GET permission on Key Vault certificates, if using SSL certificates sourced from Key Vault

## Quick Start

Create a file `appgateway.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: my-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureApplicationGateway.my-appgw
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-appgw
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/appgw-subnet
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/publicIPAddresses/my-appgw-pip
  sku: Standard_v2
  backendAddressPools:
    - name: default-pool
      ipAddresses:
        - "10.0.1.4"
  backendHttpSettings:
    - name: http-settings
      port: 80
      protocol: Http
  httpListeners:
    - name: http-listener
      port: 80
      protocol: Http
  requestRoutingRules:
    - name: http-rule
      httpListenerName: http-listener
      backendAddressPoolName: default-pool
      backendHttpSettingsName: http-settings
      priority: 100
```

Deploy:

```shell
openmcf apply -f appgateway.yaml
```

This creates a Standard_v2 Application Gateway with a single HTTP listener on port 80, routing traffic to a backend pool at 10.0.1.4. Capacity defaults to 2 instances.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Application Gateway (e.g., `eastus`, `westeurope`). Must match the region of the subnet, public IP, and backend resources. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Application Gateway. Must be unique within the resource group. | Required, 1-80 characters |
| `subnetId` | `StringValueOrRef` | ID of a dedicated subnet for the Application Gateway. App Gateway v2 requires its own subnet with no other resources. Can reference an AzureSubnet resource via `valueFrom`. | Required |
| `publicIpId` | `StringValueOrRef` | ID of a Standard SKU public IP with static allocation. Can reference an AzurePublicIp resource via `valueFrom`. | Required |
| `sku` | `string` | SKU tier. Values: `Standard_v2` (general L7 load balancing), `WAF_v2` (adds Web Application Firewall). | Required, must be `Standard_v2` or `WAF_v2` |
| `backendAddressPools` | `object[]` | Backend address pools that receive routed traffic. Each pool has a `name` and optional `fqdns` and `ipAddresses` arrays. | At least 1 required |
| `backendHttpSettings` | `object[]` | Backend HTTP settings defining port, protocol, affinity, and timeout. Each entry has required fields `name`, `port`, `protocol`. | At least 1 required |
| `httpListeners` | `object[]` | HTTP listeners defining entry points for traffic. Each entry has required fields `name`, `port`, `protocol`. | At least 1 required |
| `requestRoutingRules` | `object[]` | Routing rules connecting listeners to backend pools. Each entry has required fields `name`, `httpListenerName`, `backendAddressPoolName`, `backendHttpSettingsName`, `priority`. | At least 1 required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `capacity` | `int` | `2` | Fixed instance count. Range: 1-125. Mutually exclusive with `autoscale`. |
| `autoscale` | `object` | -- | Autoscale configuration with `minCapacity` (0-100) and optional `maxCapacity` (2-125). Mutually exclusive with `capacity`. |
| `probes` | `object[]` | `[]` | Custom health probes. Each has required fields `name`, `protocol`, `path` and optional `host`, `interval` (default 30s), `timeout` (default 30s), `unhealthyThreshold` (default 3). |
| `sslCertificates` | `object[]` | `[]` | SSL certificates sourced from Key Vault. Each has required fields `name` and `keyVaultSecretId`. Required when any listener uses protocol `Https`. |
| `identityIds` | `StringValueOrRef[]` | `[]` | User-assigned managed identity IDs. Required when `sslCertificates` reference Key Vault secrets. Can reference AzureUserAssignedIdentity resources via `valueFrom`. |
| `wafEnabled` | `bool` | `false` | Enable Web Application Firewall. Only valid when `sku` is `WAF_v2`. |
| `wafMode` | `string` | `Prevention` | WAF firewall mode. Values: `Detection` (log only), `Prevention` (block and log). Only used when `wafEnabled` is `true`. |
| `enableHttp2` | `bool` | `false` | Enable HTTP/2 for client-to-gateway connections. Backend connections always use HTTP/1.1. |
| `backendHttpSettings[].cookieBasedAffinity` | `string` | `Disabled` | Cookie-based session affinity. Values: `Enabled`, `Disabled`. |
| `backendHttpSettings[].requestTimeout` | `int` | `30` | Backend request timeout in seconds. Range: 1-86400. |
| `backendHttpSettings[].probeName` | `string` | -- | Name of a custom health probe. Must match a probe in `probes`. |
| `backendHttpSettings[].hostName` | `string` | -- | Override the Host header sent to backends. Mutually exclusive with `pickHostNameFromBackendAddress`. |
| `backendHttpSettings[].pickHostNameFromBackendAddress` | `bool` | `false` | Automatically set Host header to the backend hostname. Mutually exclusive with `hostName`. |
| `httpListeners[].hostName` | `string` | -- | Host name for host-based routing. Enables virtual hosting on the same port. |
| `httpListeners[].sslCertificateName` | `string` | -- | SSL certificate name for HTTPS listeners. Must match a certificate in `sslCertificates`. Required when protocol is `Https`. |

## Examples

### Minimal HTTP Gateway

A basic HTTP-only Application Gateway with a single listener, backend pool, and routing rule:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: basic-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureApplicationGateway.basic-appgw
spec:
  region: eastus
  resourceGroup: dev-rg
  name: basic-appgw
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet/subnets/appgw-subnet
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/publicIPAddresses/appgw-pip
  sku: Standard_v2
  capacity: 1
  backendAddressPools:
    - name: web-pool
      ipAddresses:
        - "10.0.1.4"
        - "10.0.1.5"
  backendHttpSettings:
    - name: http-settings
      port: 80
      protocol: Http
  httpListeners:
    - name: http-listener
      port: 80
      protocol: Http
  requestRoutingRules:
    - name: http-rule
      httpListenerName: http-listener
      backendAddressPoolName: web-pool
      backendHttpSettingsName: http-settings
      priority: 100
```

### HTTPS with SSL Termination

An HTTPS Application Gateway using a Key Vault certificate for SSL termination, with a user-assigned managed identity for Key Vault access:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: https-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationGateway.https-appgw
spec:
  region: eastus
  resourceGroup: prod-rg
  name: https-appgw
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/appgw-subnet
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/appgw-pip
  sku: Standard_v2
  capacity: 2
  identityIds:
    - /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/appgw-identity
  sslCertificates:
    - name: wildcard-cert
      keyVaultSecretId: https://prod-keyvault.vault.azure.net/secrets/wildcard-cert
  backendAddressPools:
    - name: api-pool
      fqdns:
        - "api.internal.contoso.com"
  backendHttpSettings:
    - name: https-backend
      port: 443
      protocol: Https
      pickHostNameFromBackendAddress: true
      requestTimeout: 60
  httpListeners:
    - name: https-listener
      port: 443
      protocol: Https
      hostName: api.contoso.com
      sslCertificateName: wildcard-cert
  requestRoutingRules:
    - name: https-rule
      httpListenerName: https-listener
      backendAddressPoolName: api-pool
      backendHttpSettingsName: https-backend
      priority: 100
```

### WAF-Enabled Gateway with Autoscale

A WAF_v2 Application Gateway with autoscaling, a custom health probe, and Prevention mode enabled:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: waf-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationGateway.waf-appgw
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: waf-appgw
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/appgw-subnet
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/waf-appgw-pip
  sku: WAF_v2
  wafEnabled: true
  wafMode: Prevention
  enableHttp2: true
  autoscale:
    minCapacity: 2
    maxCapacity: 10
  probes:
    - name: api-health
      protocol: Http
      path: /healthz
      interval: 15
      timeout: 10
      unhealthyThreshold: 3
  backendAddressPools:
    - name: api-pool
      fqdns:
        - "api-1.internal.contoso.com"
        - "api-2.internal.contoso.com"
  backendHttpSettings:
    - name: http-settings
      port: 8080
      protocol: Http
      cookieBasedAffinity: Enabled
      requestTimeout: 120
      probeName: api-health
  httpListeners:
    - name: http-listener
      port: 80
      protocol: Http
  requestRoutingRules:
    - name: api-rule
      httpListenerName: http-listener
      backendAddressPoolName: api-pool
      backendHttpSettingsName: http-settings
      priority: 100
```

### Host-Based Routing with Multiple Listeners

An Application Gateway with multiple listeners on the same port routing to different backend pools based on host name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: multi-host-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationGateway.multi-host-appgw
spec:
  region: eastus
  resourceGroup: prod-rg
  name: multi-host-appgw
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/appgw-subnet
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/appgw-pip
  sku: Standard_v2
  capacity: 3
  backendAddressPools:
    - name: api-pool
      fqdns:
        - "api.internal.contoso.com"
    - name: web-pool
      ipAddresses:
        - "10.0.2.10"
        - "10.0.2.11"
  backendHttpSettings:
    - name: api-settings
      port: 8080
      protocol: Http
      requestTimeout: 60
    - name: web-settings
      port: 80
      protocol: Http
  httpListeners:
    - name: api-listener
      port: 80
      protocol: Http
      hostName: api.contoso.com
    - name: web-listener
      port: 80
      protocol: Http
      hostName: www.contoso.com
  requestRoutingRules:
    - name: api-route
      httpListenerName: api-listener
      backendAddressPoolName: api-pool
      backendHttpSettingsName: api-settings
      priority: 100
    - name: web-route
      httpListenerName: web-listener
      backendAddressPoolName: web-pool
      backendHttpSettingsName: web-settings
      priority: 200
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding Azure resource IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: ref-appgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationGateway.ref-appgw
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-appgw
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: appgw-subnet
      field: status.outputs.subnet_id
  publicIpId:
    valueFrom:
      kind: AzurePublicIp
      name: appgw-pip
      field: status.outputs.public_ip_id
  sku: Standard_v2
  identityIds:
    - valueFrom:
        kind: AzureUserAssignedIdentity
        name: appgw-identity
        field: status.outputs.identity_id
  sslCertificates:
    - name: wildcard-cert
      keyVaultSecretId: https://my-keyvault.vault.azure.net/secrets/wildcard-cert
  backendAddressPools:
    - name: app-pool
      fqdns:
        - "app.internal.contoso.com"
  backendHttpSettings:
    - name: http-settings
      port: 80
      protocol: Http
  httpListeners:
    - name: https-listener
      port: 443
      protocol: Https
      sslCertificateName: wildcard-cert
  requestRoutingRules:
    - name: https-rule
      httpListenerName: https-listener
      backendAddressPoolName: app-pool
      backendHttpSettingsName: http-settings
      priority: 100
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `appGatewayId` | `string` | Azure Resource Manager ID of the Application Gateway |
| `appGatewayName` | `string` | Name of the Application Gateway |

The public frontend IP address is not exported here because it comes from the AzurePublicIp resource referenced via `publicIpId`. DNS records should reference the AzurePublicIp's `ipAddress` output directly.

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for gateway placement
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the virtual network containing the dedicated Application Gateway subnet
- [AzureSubnet](/docs/catalog/azure/azuresubnet) -- provides the dedicated subnet required by Application Gateway v2
- [AzurePublicIp](/docs/catalog/azure/azurepublicip) -- provides the Standard SKU static public IP for the frontend
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- stores SSL/TLS certificates referenced by HTTPS listeners
- [AzureUserAssignedIdentity](/docs/catalog/azure/azureuserassignedidentity) -- managed identity for Key Vault certificate access
- [AzureDnsRecord](/docs/catalog/azure/azurednsrecord) -- DNS entries pointing to the gateway's public IP

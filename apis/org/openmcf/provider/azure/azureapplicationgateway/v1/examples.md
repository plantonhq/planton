# Azure Application Gateway Examples

This document provides comprehensive examples for deploying Azure Application Gateways using the AzureApplicationGateway API resource.

## Table of Contents

- [Minimal HTTP Application Gateway](#minimal-http-application-gateway)
- [HTTPS with Key Vault SSL Certificate](#https-with-key-vault-ssl-certificate)
- [WAF-Enabled Application Gateway](#waf-enabled-application-gateway)
- [Host-Based Routing](#host-based-routing)
- [Autoscale Configuration](#autoscale-configuration)
- [Multi-Backend with Health Probes](#multi-backend-with-health-probes)
- [Infra Chart Reference (valueFrom)](#infra-chart-reference-valuefrom)

---

## Minimal HTTP Application Gateway

The simplest possible configuration: a single HTTP listener routing all traffic to one backend pool.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: minimal-agw
spec:
  # Azure region -- must match subnet and public IP region
  region: eastus
  resourceGroup:
    value: rg-networking
  name: minimal-agw

  # Dedicated subnet (no other resources allowed)
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet

  # Standard SKU public IP with static allocation
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  # Standard_v2 for general L7 load balancing (no WAF)
  sku: Standard_v2
  # capacity defaults to 2 instances when neither capacity nor autoscale is set

  # Single backend pool targeting an FQDN
  backendAddressPools:
    - name: web-backend
      fqdns:
        - "app.internal.contoso.com"

  # Backend HTTP settings: plain HTTP on port 80
  backendHttpSettings:
    - name: http-settings
      port: 80
      protocol: Http
      # cookieBasedAffinity defaults to Disabled
      # requestTimeout defaults to 30 seconds

  # HTTP listener on port 80
  httpListeners:
    - name: http-listener
      port: 80
      protocol: Http

  # Route all HTTP traffic to the web-backend pool
  requestRoutingRules:
    - name: http-rule
      httpListenerName: http-listener
      backendAddressPoolName: web-backend
      backendHttpSettingsName: http-settings
      priority: 100
```

**Deploy:**
```shell
planton apply -f minimal-agw.yaml
```

**What you get:**
- Standard_v2 Application Gateway with 2 fixed instances
- Single HTTP listener on port 80
- Traffic routed to backend FQDN via plain HTTP
- Default Azure health probes (GET / on port 80)
- Frontend port names auto-derived from listener definitions

---

## HTTPS with Key Vault SSL Certificate

SSL termination using a certificate stored in Azure Key Vault. Requires a user-assigned managed identity with Key Vault access.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: https-agw
spec:
  region: eastus
  resourceGroup:
    value: rg-networking
  name: https-agw

  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  sku: Standard_v2
  capacity: 3

  # User-assigned identity with GET permission on Key Vault certificates
  identityIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/agw-identity

  # SSL certificate from Key Vault (no PFX data in the manifest)
  sslCertificates:
    - name: wildcard-contoso
      keyVaultSecretId: https://my-keyvault.vault.azure.net/secrets/wildcard-contoso-cert

  backendAddressPools:
    - name: api-backend
      fqdns:
        - "api.internal.contoso.com"

  backendHttpSettings:
    - name: https-backend-settings
      port: 443
      protocol: Https
      requestTimeout: 60
      # Pick host from backend FQDN for multi-tenant backends (e.g., App Service)
      pickHostNameFromBackendAddress: true

  # HTTPS listener on port 443 with the Key Vault certificate
  httpListeners:
    - name: https-listener
      port: 443
      protocol: Https
      sslCertificateName: wildcard-contoso

  requestRoutingRules:
    - name: https-rule
      httpListenerName: https-listener
      backendAddressPoolName: api-backend
      backendHttpSettingsName: https-backend-settings
      priority: 100
```

**Deploy:**
```shell
planton apply -f https-agw.yaml
```

**What you get:**
- End-to-end encryption: HTTPS from client to gateway, HTTPS from gateway to backend
- Certificate managed in Key Vault (auto-renewal supported)
- User-assigned identity for secure Key Vault access
- 3 fixed instances for predictable capacity

**Prerequisites:**
1. Create the user-assigned identity (AzureUserAssignedIdentity)
2. Create the Key Vault and upload the certificate (AzureKeyVault)
3. Grant the identity GET permission on Key Vault certificates
4. Then deploy the Application Gateway

---

## WAF-Enabled Application Gateway

Web Application Firewall in Prevention mode to block common web exploits.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: waf-agw
  labels:
    environment: production
spec:
  region: eastus
  resourceGroup:
    value: rg-networking
  name: waf-agw

  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  # WAF_v2 SKU required for Web Application Firewall
  sku: WAF_v2
  capacity: 3

  # Enable WAF in Prevention mode (blocks attacks, not just logs)
  wafEnabled: true
  wafMode: Prevention

  # Enable HTTP/2 for improved client-to-gateway performance
  enableHttp2: true

  identityIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/agw-identity

  sslCertificates:
    - name: wildcard-contoso
      keyVaultSecretId: https://my-keyvault.vault.azure.net/secrets/wildcard-contoso-cert

  backendAddressPools:
    - name: web-backend
      fqdns:
        - "web.internal.contoso.com"

  backendHttpSettings:
    - name: http-settings
      port: 80
      protocol: Http
      requestTimeout: 30

  httpListeners:
    - name: https-listener
      port: 443
      protocol: Https
      sslCertificateName: wildcard-contoso

  requestRoutingRules:
    - name: https-rule
      httpListenerName: https-listener
      backendAddressPoolName: web-backend
      backendHttpSettingsName: http-settings
      priority: 100
```

**Deploy:**
```shell
planton apply -f waf-agw.yaml
```

**What you get:**
- WAF_v2 with OWASP 3.2 rule set in Prevention mode
- Blocks SQL injection, XSS, and other OWASP Top 10 attacks
- SSL termination at the gateway, HTTP to backend (common pattern)
- HTTP/2 enabled for client connections

**WAF mode guidance:**
- Use **Prevention** for production (blocks and logs)
- Use **Detection** during initial deployment to assess false positives without blocking traffic, then switch to Prevention

---

## Host-Based Routing

Multiple listeners on the same port routing to different backends based on domain name.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: multi-host-agw
spec:
  region: eastus
  resourceGroup:
    value: rg-networking
  name: multi-host-agw

  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  sku: Standard_v2
  capacity: 2

  # Two distinct backend pools for different services
  backendAddressPools:
    - name: api-backend
      fqdns:
        - "api.internal.contoso.com"
    - name: web-backend
      fqdns:
        - "web.internal.contoso.com"

  # Different HTTP settings per backend
  backendHttpSettings:
    - name: api-settings
      port: 8080
      protocol: Http
      requestTimeout: 60
      cookieBasedAffinity: Disabled
    - name: web-settings
      port: 80
      protocol: Http
      requestTimeout: 30
      cookieBasedAffinity: Enabled  # sticky sessions for web app

  # Host-based listeners: same port (80), different host names
  httpListeners:
    - name: api-listener
      port: 80
      protocol: Http
      hostName: api.contoso.com       # matches api.contoso.com requests
    - name: web-listener
      port: 80
      protocol: Http
      hostName: www.contoso.com       # matches www.contoso.com requests

  # Route each listener to its backend
  requestRoutingRules:
    - name: api-rule
      httpListenerName: api-listener
      backendAddressPoolName: api-backend
      backendHttpSettingsName: api-settings
      priority: 100
    - name: web-rule
      httpListenerName: web-listener
      backendAddressPoolName: web-backend
      backendHttpSettingsName: web-settings
      priority: 200
```

**Deploy:**
```shell
planton apply -f multi-host-agw.yaml
```

**What you get:**
- Single public IP serving two domains
- api.contoso.com routes to API backend on port 8080
- www.contoso.com routes to web backend on port 80 with sticky sessions
- Each backend gets independent HTTP settings (timeout, affinity, etc.)

**DNS setup:**
Both `api.contoso.com` and `www.contoso.com` should have A records pointing to the Application Gateway's public IP.

---

## Autoscale Configuration

Dynamic scaling based on traffic load instead of fixed instance count.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: autoscale-agw
spec:
  region: westus2
  resourceGroup:
    value: rg-networking
  name: autoscale-agw

  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  sku: Standard_v2

  # Autoscale replaces fixed capacity -- they are mutually exclusive
  # Min 2 instances for HA, max 10 to cap costs
  autoscale:
    minCapacity: 2
    maxCapacity: 10

  backendAddressPools:
    - name: app-backend
      ipAddresses:
        - "10.0.1.4"
        - "10.0.1.5"
        - "10.0.1.6"

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
      backendAddressPoolName: app-backend
      backendHttpSettingsName: http-settings
      priority: 100
```

**Deploy:**
```shell
planton apply -f autoscale-agw.yaml
```

**What you get:**
- Scales from 2 to 10 instances based on traffic
- Cost-efficient: pay only for instances in use
- Backends addressed by IP (useful for VMs or containers with static IPs)

**Autoscale guidance:**
- Set `minCapacity` >= 2 for production (HA across zones)
- Set `maxCapacity` based on expected peak traffic and budget
- Use `minCapacity: 0` only for non-production (scale-to-zero saves cost but adds cold-start latency)

---

## Multi-Backend with Health Probes

Custom health probes for accurate backend health monitoring.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: probed-agw
spec:
  region: eastus
  resourceGroup:
    value: rg-networking
  name: probed-agw

  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/agw-subnet
  publicIpId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-networking/providers/Microsoft.Network/publicIPAddresses/agw-pip

  sku: Standard_v2
  capacity: 3

  # Custom health probes for each backend service
  probes:
    - name: api-health
      protocol: Http
      path: /api/healthz        # Kubernetes-style health endpoint
      interval: 15               # Check every 15 seconds
      timeout: 10                # Timeout after 10 seconds
      unhealthyThreshold: 3      # 3 failures = unhealthy
    - name: web-health
      protocol: Http
      path: /health
      host: www.contoso.com      # Override Host header for the probe
      interval: 30
      timeout: 30
      unhealthyThreshold: 2

  backendAddressPools:
    - name: api-backend
      fqdns:
        - "api-1.internal.contoso.com"
        - "api-2.internal.contoso.com"
    - name: web-backend
      ipAddresses:
        - "10.0.2.10"
        - "10.0.2.11"

  # Each backend HTTP settings references its custom probe
  backendHttpSettings:
    - name: api-settings
      port: 8080
      protocol: Http
      requestTimeout: 60
      probeName: api-health       # Link to the api-health probe
    - name: web-settings
      port: 80
      protocol: Http
      requestTimeout: 30
      cookieBasedAffinity: Enabled
      probeName: web-health       # Link to the web-health probe

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
    - name: api-rule
      httpListenerName: api-listener
      backendAddressPoolName: api-backend
      backendHttpSettingsName: api-settings
      priority: 100
    - name: web-rule
      httpListenerName: web-listener
      backendAddressPoolName: web-backend
      backendHttpSettingsName: web-settings
      priority: 200
```

**Deploy:**
```shell
planton apply -f probed-agw.yaml
```

**What you get:**
- Custom health probes with application-specific endpoints
- API backend checked every 15s with fast failure detection
- Web backend with Host header override for accurate health checks
- Unhealthy backends automatically removed from rotation

**Health probe best practices:**
- Use a dedicated health endpoint (not `/`) that checks downstream dependencies
- Set `interval` lower for critical services (15s) and higher for less critical ones (30s)
- Set `unhealthyThreshold` to 2-3 to avoid flapping on transient failures
- Match the probe `protocol` to the backend HTTP settings protocol

---

## Infra Chart Reference (valueFrom)

Using `ref` foreign keys to wire resources from the enterprise-network-foundation infra chart. This eliminates hard-coded Azure resource IDs.

```yaml
# Prerequisite resources (deployed separately or in the same infra chart)
---
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: rg-networking
spec:
  region: eastus
  name: rg-networking

---
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: agw-subnet
spec:
  # ... subnet configuration ...

---
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: agw-pip
spec:
  # ... public IP configuration ...

---
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: agw-identity
spec:
  # ... identity configuration ...

---
# Application Gateway using refs to resolve IDs at deploy time
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationGateway
metadata:
  name: enterprise-agw
spec:
  region: eastus

  # Reference the resource group name from AzureResourceGroup output
  resourceGroup:
    ref:
      kind: AzureResourceGroup
      name: rg-networking
      path: status.outputs.resource_group_name

  name: enterprise-agw

  # Reference the subnet ID from AzureSubnet output
  subnetId:
    ref:
      kind: AzureSubnet
      name: agw-subnet
      path: status.outputs.subnet_id

  # Reference the public IP ID from AzurePublicIp output
  publicIpId:
    ref:
      kind: AzurePublicIp
      name: agw-pip
      path: status.outputs.public_ip_id

  sku: WAF_v2
  wafEnabled: true
  wafMode: Prevention
  enableHttp2: true

  autoscale:
    minCapacity: 2
    maxCapacity: 10

  # Reference the identity from AzureUserAssignedIdentity output
  identityIds:
    - ref:
        kind: AzureUserAssignedIdentity
        name: agw-identity
        path: status.outputs.identity_id

  sslCertificates:
    - name: wildcard-contoso
      keyVaultSecretId: https://enterprise-kv.vault.azure.net/secrets/wildcard-contoso-cert

  backendAddressPools:
    - name: default
      fqdns:
        - "app.internal.contoso.com"

  backendHttpSettings:
    - name: https-settings
      port: 443
      protocol: Https
      requestTimeout: 60
      pickHostNameFromBackendAddress: true

  httpListeners:
    - name: https-listener
      port: 443
      protocol: Https
      sslCertificateName: wildcard-contoso

  requestRoutingRules:
    - name: default-rule
      httpListenerName: https-listener
      backendAddressPoolName: default
      backendHttpSettingsName: https-settings
      priority: 100
```

**Deploy:**
```shell
# Deploy the infra chart (all resources in dependency order)
planton apply -f enterprise-network-foundation.yaml
```

**What you get:**
- Zero hard-coded Azure resource IDs
- Type-safe foreign key references with default field paths
- Automatic dependency resolution between resources
- Infra chart deploys resources in the correct order
- Changes to upstream resources (e.g., subnet rebuild) automatically propagate

**Foreign key defaults:**
The `ref` shorthand uses default field paths defined in the proto schema:
- `AzureResourceGroup` -> `status.outputs.resource_group_name`
- `AzureSubnet` -> `status.outputs.subnet_id`
- `AzurePublicIp` -> `status.outputs.public_ip_id`
- `AzureUserAssignedIdentity` -> `status.outputs.identity_id`

---

## Deployment Tips

### Verify Outputs

After deployment, check the Application Gateway status:

```shell
az network application-gateway show \
  --resource-group rg-networking \
  --name enterprise-agw \
  --query "{id:id, state:operationalState, sku:sku}" \
  --output table
```

### Common Configurations

**Production Checklist:**
- ✅ `sku`: `WAF_v2` (or `Standard_v2` if WAF not needed)
- ✅ `wafEnabled`: `true` with `wafMode`: `Prevention`
- ✅ `autoscale`: configured with min >= 2
- ✅ SSL certificates from Key Vault (not PFX in manifests)
- ✅ Custom health probes for each backend service
- ✅ `enableHttp2`: `true` for performance
- ✅ Dedicated /24 subnet with no other resources
- ✅ Standard SKU public IP with static allocation

**Development Configuration:**
- ⚠️ `sku`: `Standard_v2` (skip WAF cost)
- ⚠️ `capacity`: `1` (single instance, no HA)
- ⚠️ HTTP only (skip SSL setup complexity)
- ⚠️ Default health probes (skip custom probe configuration)

## Support

For issues or questions:
- Check the [main README](./README.md) for component overview
- Review the [research documentation](./docs/README.md) for architecture details
- Consult the [Azure Application Gateway Documentation](https://learn.microsoft.com/en-us/azure/application-gateway/)

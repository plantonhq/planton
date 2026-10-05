# Azure Application Gateway

Deploys an Azure Application Gateway -- the Layer 7 (HTTP/HTTPS) load balancer and reverse proxy that routes by host name and URI path, terminates TLS (including mutual TLS) with Key Vault certificates that renew in place, rewrites requests and responses in flight, proxies raw TCP/TLS at layer 4, and enforces a Web Application Firewall policy on the WAF_v2 SKU. The gateway bundles its sub-objects -- frontends, ports, listeners, backend pools, backend settings, routing rules, path maps, probes, certificates, SSL profiles, redirects, and rewrites -- because Azure configures them as one atomic ARM resource: none has a life outside its gateway, and they wire to each other BY NAME within the spec. What other resources need to reach is exported as name-keyed map outputs, so pool membership composes from the member side without splitting the gateway apart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions one Application Gateway carrying every declared sub-object:

- **The gateway** -- BASIC, STANDARD_V2, or WAF_V2 SKU with fixed capacity or autoscale bounds, availability zones, HTTP/2, FIPS mode, and request/response buffering posture
- **Frontend IP Configurations** -- public (a referenced Standard AzurePublicIp) or private (an address in the gateway's own dedicated subnet), at least one; optionally exposed over Private Link to other networks
- **Frontend Ports** -- declared once and shared by name (one "https" 443 serves every HTTPS listener)
- **Backend Address Pools** -- FQDN and/or IP targets, or empty pools that NICs and scale sets join member-side after deploy
- **Backend HTTP Settings** -- port, protocol, cookie affinity, timeouts, probe binding, host-header handling, connection draining, and backend-TLS trust per backend group
- **HTTP Listeners** -- frontend + port + protocol entry points with wildcard host names, TLS termination, SNI, per-listener WAF overrides, and custom error pages
- **Request Routing Rules and URL Path Maps** -- basic listener-to-backend (or redirect) wiring, or per-path routing with first-match-wins path rules
- **Health Probes** -- HTTP(S) probes with host pairing and match criteria, or TCP/TLS probes for layer-4 backends
- **SSL Certificates, Trusted Root/Client Certificates, and SSL Profiles** -- Key Vault or inline sources, backend-CA trust, and named mutual-TLS postures
- **Redirect Configurations and Rewrite Rule Sets** -- the HTTP-to-HTTPS pattern, header edits, and conditional URL rewrites
- **Layer-4 Listeners, Backend Settings, and Routing Rules** -- raw TCP/TLS proxying through the same gateway
- **Managed Identity association** -- system- and/or user-assigned; the user-assigned identity is how Key Vault certificate fetches authenticate
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`

Pool membership is NOT created here -- each AzureNetworkInterface or scale set references `status.outputs.backend_address_pool_ids.<pool-name>` to join. The WAF policy is its own first-class resource (AzureWebApplicationFirewallPolicy), referenced at up to three levels: gateway, listener, and path rule -- most specific wins.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the gateway will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A DEDICATED subnet** -- Azure allows no other resource type in the gateway's subnet. /24 is recommended for production (up to 125 v2 instances plus Azure-reserved addresses). Reference an AzureSubnet Infra Component.
- **A Standard SKU public IP** (for public frontends) -- reference an AzurePublicIp Infra Component; DNS records point at ITS address output.
- **For Key Vault certificates** -- a user-assigned identity (AzureUserAssignedIdentity) with GET on the vault's secrets, granted BEFORE the gateway deploys.
- **Time** -- applies run 15-25 minutes; Azure's slowest networking resource. Plan pipelines accordingly.

## Deploy

### Console

Open the deployment store, find **Azure Application Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the gateway's sub-objects in dependency order -- every name reference is a dropdown over what you have declared, so a dangling reference cannot be typed. Start from the **Standard HTTPS Gateway with Key Vault TLS** preset for the production TLS baseline in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationGateway
metadata:
  name: web-gateway
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "web-rg"
  name: web-gateway
  subnetId:
    valueFrom:
      name: appgw-subnet
  sku: STANDARD_V2
  autoscale:
    minCapacity: 2
    maxCapacity: 10
  zones: ["1", "2", "3"]
  frontendIpConfigurations:
    - name: public
      publicIpAddressId:
        valueFrom:
          name: gw-pip
  frontendPorts:
    - name: http
      port: 80
  backendAddressPools:
    # Membership joins from the member side: a NIC ip_configuration or a
    # scale set references status.outputs.backend_address_pool_ids.web.
    - name: web
  backendHttpSettings:
    - name: http-settings
      port: 8080
      protocol: HTTP
  httpListeners:
    - name: http
      frontendIpConfigurationName: public
      frontendPortName: http
      protocol: HTTP
  requestRoutingRules:
    - name: default
      ruleType: BASIC_ROUTING
      httpListenerName: http
      priority: 100
      backendAddressPoolName: web
      backendHttpSettingsName: http-settings
```

```shell
planton apply -f azure-application-gateway.yaml
```

This creates a zone-redundant autoscaling Standard v2 gateway with one public frontend, an HTTP listener, and a basic rule to an empty pool that members join after deploy. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the gateway to its resource group, subnet, and public IP -- and wire each member NIC to the pool:

```yaml
# On the AzureApplicationGateway:
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: web-rg
      fieldPath: status.outputs.resource_group_name
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: appgw-subnet
      fieldPath: status.outputs.subnet_id
  frontendIpConfigurations:
    - name: public
      publicIpAddressId:
        valueFrom:
          kind: AzurePublicIp
          name: gw-pip
          fieldPath: status.outputs.public_ip_id

# On each AzureNetworkInterface that should join the pool:
spec:
  ipConfigurations:
    - applicationGatewayBackendPoolIds:
        - valueFrom:
            kind: AzureApplicationGateway
            name: web-gateway
            fieldPath: status.outputs.backend_address_pool_ids.web
```

The InfraPipeline resolves the dependency graph, deploys the group, subnet, and public IP first, then the gateway, then the interfaces that join its pools.

## Key Configuration

These are the most important decisions when configuring an Application Gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU** -- fixed at creation and feature-gating everything else: `STANDARD_V2` is the general-purpose platform (autoscale, zones, rewrites, mTLS, Private Link); `WAF_V2` adds Web Application Firewall enforcement via a referenced `firewallPolicyId`; `BASIC` caps capacity at 2 and drops autoscale, mTLS, URL rewriting, zones, and WAF. Azure's retired v1 SKUs are not modeled.

**Scale** -- exactly one of `capacity` (1-125 fixed; BASIC: 1-2) and `autoscale` (min 0-100, max 2-125). Autoscale 2-10 with all three `zones` is the production posture; a minimum of 0 allows scale-to-zero billing at idle at the cost of cold-request latency.

**The traffic path** -- a listener (frontend + port + protocol + optional wildcard host names) feeds a routing rule, which targets a backend pool via backend HTTP settings XOR a redirect (BASIC_ROUTING), or consults a URL path map for per-path destinations (PATH_BASED_ROUTING). `priority` is required and unique across ALL rules (1-20000, lower wins) -- a specific-host listener's rule must out-rank a wildcard's.

**TLS** -- certificates source from Key Vault (the production grain: reference an AzureKeyVaultCertificate's `versionless_secret_id` so renewals propagate without a redeploy) XOR inline PFX data. Key Vault fetches authenticate through the gateway's USER-ASSIGNED identity, which needs GET on the vault's secrets before deploy. SSL profiles add mutual TLS (client-CA bundles, issuer-DN verification, OCSP) and per-listener policy overrides; the gateway-wide `sslPolicy` sets the baseline (AppGwSslPolicy20220101S is Microsoft's strict recommendation).

**Probes** -- backends without a custom probe get Azure's default (GET / on the backend port, healthy on 200-399) -- an app whose root path redirects with auth or serves only /api/* is silently dropped from rotation. Declare a /healthz probe; HTTP(S) probes need exactly one host source (explicit XOR pick-from-settings), TCP/TLS probes serve layer-4 backends.

**Layer 4** -- the `listeners`/`backends`/`routingRules` trio proxies raw TCP or TLS (databases, MQTT, custom protocols) through the same gateway, sharing the declared frontends, ports, pools, and certificates. The spec requires at least one listener, one backend-settings entry, and one routing rule across EITHER the L7 or the L4 side.

**Provisioning time** -- 15-25 minutes per apply. Changes to the atomic gateway re-run the deploy; plan pipelines accordingly.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `subnetId` (and per private frontend / Private Link allocation) | `status.outputs.subnet_id` |
| **AzurePublicIp** (per public frontend) | `frontendIpConfigurations[].publicIpAddressId` | `status.outputs.public_ip_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |
| **AzureKeyVaultCertificate** (per Key Vault certificate) | `sslCertificates[].keyVaultSecretId` | `status.outputs.versionless_secret_id` |
| **AzureWebApplicationFirewallPolicy** (WAF_V2; three levels) | `firewallPolicyId`, `httpListeners[].firewallPolicyId`, path rules' `firewallPolicyId` | `status.outputs.policy_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef. The name-keyed maps are the composition seams:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `application_gateway_id` | Azure Resource Manager ID of the gateway | Diagnostics settings, RBAC scopes |
| `application_gateway_name` | Name of the gateway | Operational tooling |
| `backend_address_pool_ids.<name>` | Each pool's ARM ID, keyed by name | THE membership seam: NIC ipConfigurations and scale-set network profiles reference it to join |
| `frontend_ip_configuration_ids.<name>` | Each frontend's ARM ID, keyed by name | Chaining the frontend into other resources |
| `private_ip_address` | The first private frontend's address | The address internal DNS records point at |
| `private_ip_addresses` | All private frontends' addresses, in declaration order | Multi-frontend internal DNS |

A public frontend's ADDRESS is not exported here -- it lives on the referenced AzurePublicIp (its `ip_address` output), which DNS records point at.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production HTTPS baseline** -- a zone-redundant autoscaling Standard v2 gateway terminating TLS with a Key Vault certificate, the universal HTTP-to-HTTPS 301 redirect, a /healthz probe, and the Microsoft strict TLS policy. Start from the **Standard HTTPS Gateway with Key Vault TLS** preset.

**WAF with path-based routing** -- a WAF v2 gateway enforcing a referenced firewall policy, with a URL path map splitting /api/* and /static/* across pools. Start from the **WAF Gateway with Path-Based Routing** preset.

**Internal gateway** -- a private frontend in the dedicated subnet serving east-west traffic with no public exposure. Start from the **Internal (Private-Only) Gateway** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the gateway is created in
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) and [**Azure Subnet**](/infra-catalog/azure-subnet) -- host the gateway's dedicated subnet and private frontends
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the address public frontends receive traffic on (and what DNS points at)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- authenticates Key Vault certificate fetches
- [**Azure Key Vault Certificate**](/infra-catalog/azure-key-vault-certificate) -- the renewing TLS certificate source
- [**Azure Web Application Firewall Policy**](/infra-catalog/azure-web-application-firewall-policy) -- the WAF rules a WAF_v2 gateway enforces
- [**Azure Network Interface**](/infra-catalog/azure-network-interface) -- joins backend pools from the member side
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the workload the pools front (via its network interfaces)

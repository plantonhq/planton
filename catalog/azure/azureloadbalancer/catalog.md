# Azure Load Balancer

Deploys an Azure Load Balancer -- the Layer 4 (TCP/UDP) traffic distributor, complete with its frontend IP configurations, backend address pools, health probes, load-balancing rules, inbound NAT rules, and outbound (SNAT) rules. The load balancer and its sub-resources are configured as one unit because none of them has a life outside it; what IS independent -- pool membership -- is expressed from the member side, matching Azure's own attachment model (a network interface or scale set references the pool's exported ID). Public vs internal is decided per frontend, not per load balancer: one resource can carry public ingress and internal east-west traffic side by side.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Load Balancer** -- a Standard SKU (the default -- Basic was retired in September 2025) or Gateway SKU resource, Regional or Global tier, with optional edge-zone placement
- **Frontend IP Configurations** -- the addresses that receive traffic: public (a referenced AzurePublicIp or AzurePublicIpPrefix) or internal (a subnet with an optional pinned static address, address family, and availability zones), at least one, mixable on one load balancer
- **Backend Address Pools** -- named containers rules route to; NIC-based members join from the member side after deploy, IP-based members (and a Global tier's regional-frontend members) are declared inline
- **Health Probes** -- TCP, HTTP, or HTTPS checks with configurable interval, failure count, and success threshold (the flap dampener)
- **Load-Balancing Rules** -- frontend port/protocol to pool/port mappings with session persistence, idle timeout, TCP reset, floating IP (Direct Server Return), and HA-ports support
- **Inbound NAT Rules** -- single-target port forwarding (completed by a member-side NIC association) or pool-style port ranges giving every member its own frontend port
- **Outbound Rules** -- explicit SNAT: which public frontends carry a pool's egress and how many SNAT ports each instance gets
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`

Pool membership is NOT created here -- each AzureNetworkInterface or scale set references `status.outputs.backend_pool_ids.<pool-name>` to join.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the load balancer will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A Standard SKU public IP** (for public frontends) -- reference an AzurePublicIp Infra Component; the IP resource carries the zone posture. A public IP PREFIX (AzurePublicIpPrefix) serves egress-heavy estates that allowlist one CIDR.
- **A subnet** (for internal frontends) within a VNet -- the frontend takes a private address there. All internal frontends of one load balancer live in the same virtual network.
- **Region alignment** -- the load balancer only serves backends in its own region (the Global tier fronts REGIONAL load balancers instead).
- **Gateway SKU only** -- the subscription needs the Microsoft.Network/AllowGatewayLoadBalancer feature registered (via an Azure support ticket), and every backend pool must declare tunnel interfaces.

## Deploy

### Console

Open the deployment store, find **Azure Load Balancer**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public Load Balancer** preset for an internet-facing load balancer or the **Internal Load Balancer** preset for private VNet traffic in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLoadBalancer
metadata:
  name: web-lb
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "web-rg"
  name: web-lb
  # STANDARD SKU and REGIONAL tier apply by default -- the production shape.
  frontendIpConfigurations:
    - name: public
      publicIpAddressId:
        value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/web-rg/providers/Microsoft.Network/publicIPAddresses/web-pip"
  backendPools:
    # Membership joins from the member side: a NIC ip_configuration or a
    # VM scale set references status.outputs.backend_pool_ids.web.
    - name: web
  healthProbes:
    - name: http-health
      protocol: PROBE_HTTP
      port: 80
      requestPath: /healthz
  rules:
    - name: http
      protocol: TCP
      frontendPort: 80
      backendPort: 8080
      backendPoolNames: [web]
      probeName: http-health
      tcpResetEnabled: true
```

```shell
planton apply -f azure-load-balancer.yaml
```

This creates a public Standard load balancer with one pool, an HTTP probe, and a TCP 80-to-8080 rule. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the load balancer to its resource group and public IP -- and wire each member NIC to the pool:

```yaml
# On the AzureLoadBalancer:
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: web-rg
      fieldPath: status.outputs.resource_group_name
  frontendIpConfigurations:
    - name: public
      publicIpAddressId:
        valueFrom:
          kind: AzurePublicIp
          name: web-pip
          fieldPath: status.outputs.public_ip_id

# On each AzureNetworkInterface that should join the pool:
spec:
  ipConfigurations:
    - loadBalancerBackendPoolIds:
        - valueFrom:
            kind: AzureLoadBalancer
            name: web-lb
            fieldPath: status.outputs.backend_pool_ids.web
```

The InfraPipeline resolves the dependency graph, deploys the group and public IP first, then the load balancer, then the interfaces that join its pools.

## Key Configuration

These are the most important decisions when configuring a load balancer. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU and tier** -- `sku` unset applies STANDARD (zone redundancy, SLA, outbound rules, HA ports); GATEWAY chains network virtual appliances (every pool then declares `tunnelInterfaces`). `skuTier` unset applies REGIONAL; GLOBAL creates a cross-region load balancer whose pool members are regional load balancers' frontends (each address's `loadBalancerFrontendIpConfigurationId`) -- Standard SKU only. Both fixed at creation.

**Frontends** -- each frontend sets exactly one address source: `subnetId` (internal, with optional `privateIpAddress` pinning, `privateIpAddressVersion`, and `zones` for zone redundancy), `publicIpAddressId`, or `publicIpPrefixId`. Rules target frontends by `name` -- optional only while exactly one frontend exists. Azure does not allow removing ALL frontends from an existing load balancer.

**Probes and rules** -- HTTP/HTTPS probes require `requestPath` (TCP forbids it); rules referencing pools, probes, and frontends must name declared entries (the spec validates every reference). Protocol ALL creates an HA-ports rule -- every port, both protocols, ports fixed at 0 -- for appliance patterns on internal Standard frontends. Two pools per rule are the Gateway dual-tunnel pattern only.

**Egress** -- implicit rule-level SNAT has a small, exhaustion-prone port budget. Production pools that egress declare an outbound rule (set `disableOutboundSnat` on the same pool's load-balancing rules -- Azure requires it to combine both) or attach a NAT gateway to the subnet. `allocatedOutboundPorts: 0` divides the frontend budget (64,000 ports per IP) evenly but reallocates as the pool scales; explicit sizing (budget / max instances, multiples of 8) is the production choice.

**NAT rules** -- a single-target rule (`frontendPort`) is half the attachment: the target NIC's NAT-rule association references `status.outputs.nat_rule_ids.<name>` to complete it. A pool-style rule (`backendPoolName` + `frontendPortStart`/`frontendPortEnd`) gives every pool member its own frontend port -- the modern replacement for the legacy NAT pool mechanism.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (per internal frontend) | `frontendIpConfigurations[].subnetId` | `status.outputs.subnet_id` |
| **AzurePublicIp** (per public frontend) | `frontendIpConfigurations[].publicIpAddressId` | `status.outputs.public_ip_id` |
| **AzurePublicIpPrefix** (per prefix frontend) | `frontendIpConfigurations[].publicIpPrefixId` | `status.outputs.public_ip_prefix_id` |
| **AzureVirtualNetwork** (per IP-member pool) | `backendPools[].virtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef. The name-keyed maps are the composition seams:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `load_balancer_id` | Azure Resource Manager ID of the load balancer | Diagnostics settings, RBAC scopes |
| `private_ip_address` | The first internal frontend's private IP | The address internal DNS records point at |
| `private_ip_addresses` | All internal frontends' private IPs, in declaration order | Multi-frontend internal DNS |
| `frontend_ip_configuration_ids.<name>` | Each frontend's ARM ID, keyed by name | Chaining behind a Gateway LB; registering a regional frontend in a GLOBAL-tier pool |
| `backend_pool_ids.<name>` | Each pool's ARM ID, keyed by name | THE membership seam: NIC ip_configurations and scale-set network profiles reference it to join |
| `probe_ids.<name>` | Each probe's ARM ID, keyed by name | A scale set's rolling-upgrade health probe |
| `nat_rule_ids.<name>` | Each inbound NAT rule's ARM ID, keyed by name | The NIC-side association completing a single-target NAT attachment |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public web tier** -- A public frontend on a referenced Standard public IP, a `web` pool, an HTTP health probe, and TCP rules with TCP reset: the internet-facing production shape. Start from the **Public Load Balancer** preset.

**Internal east-west tier** -- A zone-redundant internal frontend with a pinned static address in your subnet, an `app` pool, and a raised idle timeout for long-lived connections. Start from the **Internal Load Balancer** preset.

**Full traffic story** -- Inbound load balancing, explicit outbound SNAT (with implicit SNAT disabled on the rule), a single-target admin SSH NAT rule, and a pool-style per-instance SSH range -- all on one public load balancer. Start from the **Outbound SNAT + NAT Port Forwarding** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the load balancer is created in
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the addresses public frontends receive traffic on
- [**Azure Public IP Prefix**](/infra-catalog/azure-public-ip-prefix) -- reserved ranges for prefix frontends and scalable SNAT
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- hosts internal frontends' private addresses
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- scopes pools with IP-based members
- [**Azure Network Interface**](/infra-catalog/azure-network-interface) -- joins pools and completes single-target NAT attachments from the member side
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the workload the pools front (via its network interfaces)

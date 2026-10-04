# AzureVirtualNetworkGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVirtualNetworkGatewaySpec** defines a virtual network gateway --
the managed gateway appliance Azure deploys into a virtual network to
terminate hybrid connectivity: site-to-site IPsec VPN tunnels,
point-to-site client VPN, VNet-to-VNet tunnels, and ExpressRoute
private circuits.

**Two gateway types**, selected by `type`:
- **VPN** (the default): terminates IPsec tunnels. Pairs with an
  AzureLocalNetworkGateway (the on-premises side) and an
  AzureVirtualNetworkGatewayConnection (the tunnel) for site-to-site,
  or carries a vpn_client_configuration for point-to-site.
- **EXPRESS_ROUTE**: terminates an ExpressRoute circuit's private
  peering into the VNet. ExpressRoute gateways get Azure-managed
  addressing -- ip_configurations must NOT carry public IPs.

**The GatewaySubnet contract**: every gateway lives in a dedicated
subnet of its virtual network whose ARM name is EXACTLY
"GatewaySubnet" -- ARM rejects any other name. Microsoft recommends
/27 or larger; the subnet carries no other workloads, no NSG, and no
route table pointing 0.0.0.0/0 away from Azure infrastructure.

**Cost and time**: gateways bill hourly per SKU from the moment they
provision, and provisioning is SLOW -- 25-45 minutes to create,
10-20 minutes to delete. The ForceNew surface (name, region,
resource_group, type, vpn_type, generation, edge_zone,
private_ip_address_enabled, every ip_configuration) is therefore
expensive; design changes to avoid replacement.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualNetworkGateway
metadata:
  name: test-virtual-network-gateway
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hub-vpn-gateway
  # VPN (site-to-site/point-to-site) is the default type; RouteBased the
  # default routing model. VPN_GW_1_AZ is the production entry point --
  # Azure retired new non-AZ VpnGw creates, and the AZ SKUs deploy in
  # every region (regional where zones are absent).
  sku: VPN_GW_1_AZ
  ipConfigurations:
    # The subnet's ARM name must be EXACTLY "GatewaySubnet" (/27 or
    # larger recommended); VPN gateways require a Standard static public
    # IP on every configuration, and the AZ SKUs require that address to
    # carry zones (give the AzurePublicIp zones ["1","2","3"]).
    - subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/GatewaySubnet
      publicIpAddressId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/publicIPAddresses/vpn-gw-pip
  bgpEnabled: true
  bgpSettings:
    asn: 65515
  # A NAT rule translating overlapping on-premises space; connections opt
  # in via their egress/ingress NAT rule id lists (the rule's ARM id
  # surfaces in the nat_rule_ids output under its name).
  natRules:
    - name: egress-overlap
      externalMappings:
        - addressSpace: "100.64.1.0/24"
      internalMappings:
        - addressSpace: "10.0.1.0/24"
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `enum` |  |  |  |
| `spec.vpnType` | `enum` |  |  |  |
| `spec.sku` | `enum` |  |  |  |
| `spec.generation` | `enum` |  |  |  |
| `spec.ipConfigurations` | `[]AzureVirtualNetworkGatewayIpConfiguration` | yes |  |  |
| `spec.ipConfigurations[].name` | `string` |  |  |  |
| `spec.ipConfigurations[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.ipConfigurations[].publicIpAddressId` | `string \| valueFrom` |  |  | AzurePublicIp (`status.outputs.public_ip_id`) |
| `spec.ipConfigurations[].privateIpAddressAllocation` | `enum` |  |  |  |
| `spec.activeActive` | `bool` |  |  |  |
| `spec.privateIpAddressEnabled` | `bool` |  |  |  |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.bgpEnabled` | `bool` |  |  |  |
| `spec.bgpSettings` | `AzureVirtualNetworkGatewayBgpSettings` |  |  |  |
| `spec.bgpSettings.asn` | `int64` |  |  |  |
| `spec.bgpSettings.peerWeight` | `int32` |  |  |  |
| `spec.bgpSettings.peeringAddresses` | `[]AzureVirtualNetworkGatewayBgpPeeringAddress` |  |  |  |
| `spec.bgpSettings.peeringAddresses[].ipConfigurationName` | `string` |  |  |  |
| `spec.bgpSettings.peeringAddresses[].apipaAddresses` | `[]string` | yes |  |  |
| `spec.customRouteAddressPrefixes` | `[]string` |  |  |  |
| `spec.defaultLocalNetworkGatewayId` | `string \| valueFrom` |  |  | AzureLocalNetworkGateway (`status.outputs.local_network_gateway_id`) |
| `spec.vpnClientConfiguration` | `AzureVirtualNetworkGatewayVpnClientConfiguration` |  |  |  |
| `spec.vpnClientConfiguration.addressSpaces` | `[]string` | yes |  |  |
| `spec.vpnClientConfiguration.aadTenant` | `string` |  |  |  |
| `spec.vpnClientConfiguration.aadAudience` | `string` |  |  |  |
| `spec.vpnClientConfiguration.aadIssuer` | `string` |  |  |  |
| `spec.vpnClientConfiguration.rootCertificates` | `[]AzureVirtualNetworkGatewayVpnClientRootCertificate` |  |  |  |
| `spec.vpnClientConfiguration.rootCertificates[].name` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.rootCertificates[].publicCertData` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.revokedCertificates` | `[]AzureVirtualNetworkGatewayVpnClientRevokedCertificate` |  |  |  |
| `spec.vpnClientConfiguration.revokedCertificates[].name` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.revokedCertificates[].thumbprint` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.radiusServerAddress` | `string` |  |  |  |
| `spec.vpnClientConfiguration.radiusServerSecret` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.vpnClientConfiguration.radiusServers` | `[]AzureVirtualNetworkGatewayVpnClientRadiusServer` |  |  |  |
| `spec.vpnClientConfiguration.radiusServers[].address` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.radiusServers[].secret` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.vpnClientConfiguration.radiusServers[].score` | `int32` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy` | `AzureVirtualNetworkGatewayVpnClientIpsecPolicy` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.dhGroup` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.ikeEncryption` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.ikeIntegrity` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.ipsecEncryption` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.ipsecIntegrity` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.pfsGroup` | `string` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.saLifetimeSeconds` | `int32` |  |  |  |
| `spec.vpnClientConfiguration.ipsecPolicy.saDataSizeKilobytes` | `int32` |  |  |  |
| `spec.vpnClientConfiguration.vpnClientProtocols` | `[]string` |  |  |  |
| `spec.vpnClientConfiguration.vpnAuthTypes` | `[]string` |  |  |  |
| `spec.vpnClientConfiguration.clientConnections` | `[]AzureVirtualNetworkGatewayClientConnection` |  |  |  |
| `spec.vpnClientConfiguration.clientConnections[].name` | `string` | yes |  |  |
| `spec.vpnClientConfiguration.clientConnections[].policyGroupNames` | `[]string` | yes |  |  |
| `spec.vpnClientConfiguration.clientConnections[].addressPrefixes` | `[]string` | yes |  |  |
| `spec.policyGroups` | `[]AzureVirtualNetworkGatewayPolicyGroup` |  |  |  |
| `spec.policyGroups[].name` | `string` | yes |  |  |
| `spec.policyGroups[].policyMembers` | `[]AzureVirtualNetworkGatewayPolicyMember` | yes |  |  |
| `spec.policyGroups[].policyMembers[].name` | `string` | yes |  |  |
| `spec.policyGroups[].policyMembers[].type` | `string` |  |  |  |
| `spec.policyGroups[].policyMembers[].value` | `string` | yes |  |  |
| `spec.policyGroups[].isDefault` | `bool` |  |  |  |
| `spec.policyGroups[].priority` | `int32` |  |  |  |
| `spec.bgpRouteTranslationForNatEnabled` | `bool` |  |  |  |
| `spec.dnsForwardingEnabled` | `bool` |  |  |  |
| `spec.ipSecReplayProtectionEnabled` | `bool` |  | `true` |  |
| `spec.minimumScaleUnit` | `int32` |  |  |  |
| `spec.maximumScaleUnit` | `int32` |  |  |  |
| `spec.remoteVnetTrafficEnabled` | `bool` |  |  |  |
| `spec.virtualWanTrafficEnabled` | `bool` |  |  |  |
| `spec.natRules` | `[]AzureVirtualNetworkGatewayNatRule` |  |  |  |
| `spec.natRules[].name` | `string` | yes |  |  |
| `spec.natRules[].mode` | `enum` |  |  |  |
| `spec.natRules[].type` | `enum` |  |  |  |
| `spec.natRules[].externalMappings` | `[]AzureVirtualNetworkGatewayNatRuleMapping` | yes |  |  |
| `spec.natRules[].externalMappings[].addressSpace` | `string` | yes |  |  |
| `spec.natRules[].externalMappings[].portRange` | `string` |  |  |  |
| `spec.natRules[].internalMappings` | `[]AzureVirtualNetworkGatewayNatRuleMapping` | yes |  |  |
| `spec.natRules[].internalMappings[].addressSpace` | `string` | yes |  |  |
| `spec.natRules[].internalMappings[].portRange` | `string` |  |  |  |
| `spec.natRules[].ipConfigurationId` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the gateway lives in, e.g. "eastus". Must match the
virtual network it deploys into. Changing the region replaces the
gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the gateway is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the gateway.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gateway's name, unique within the resource group. 1-80
characters; must begin with a letter or number, end with a letter,
number, or underscore, and may contain only letters, numbers,
underscores, periods, or hyphens. Changing the name replaces the
gateway.

- rule: Gateway names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.type

`enum`

What the gateway terminates: VPN (IPsec tunnels -- the default) or
EXPRESS_ROUTE (an ExpressRoute circuit's private peering). Fixed at
creation. Unspecified deploys VPN, the site-to-site shape.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_type_unspecified` -- Not specified -- deploys VPN, the site-to-site/point-to-site shape.
- `VPN` -- Terminate IPsec VPN tunnels (site-to-site, VNet-to-VNet) and point-to-site client VPN.
- `EXPRESS_ROUTE` -- Terminate an ExpressRoute circuit's private peering into the VNet.

### spec.vpnType

`enum`

The VPN routing model: ROUTE_BASED (any-to-any IKEv2 tunnels, BGP,
P2S -- the default and what virtually every modern deployment uses)
or POLICY_BASED (legacy IKEv1 with static traffic selectors; Basic
SKU only, no BGP, no P2S). Fixed at creation. Only meaningful on VPN
gateways.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_vpn_type_unspecified` -- Not specified -- deploys ROUTE_BASED, the modern any-to-any model.
- `ROUTE_BASED` -- Any-to-any IKEv2 tunnels with dynamic routing, BGP, and point-to-site support. The default and the right choice for anything new.
- `POLICY_BASED` -- Legacy IKEv1 with static traffic selectors. Basic SKU only; no BGP, no point-to-site. Exists for old on-premises devices that cannot do route-based.

### spec.sku

`enum`

The gateway SKU -- sizes throughput, tunnel/connection counts, and
hourly cost. VPN gateways use BASIC or VPN_GW_1_AZ..5_AZ: Azure
stopped accepting NEW non-AZ VpnGw1-5 creates on 2025-11-01 (ARM
rejects them with NonAzSkusNotAllowedForVPNGateway; the legacy
Standard/HighPerformance VPN tiers retired 2025-09-30), and the AZ
SKUs deploy in EVERY region -- where the region has no availability
zones the gateway is simply regional. ExpressRoute gateways use
STANDARD/HIGH_PERFORMANCE/ULTRA_PERFORMANCE or ER_GW_*_AZ;
ER_GW_SCALE autoscales between minimum_scale_unit and
maximum_scale_unit. BASIC is dev/test only (no zone redundancy, no
IKEv2 P2S, requires a Standard public IP) and cannot resize to other
SKUs in place. Generation2 VPN SKUs start at VPN_GW_2_AZ.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_sku_unspecified` -- Not specified -- invalid: the SKU choice is explicit (see the sku_required contract).
- `BASIC` -- Dev/test VPN tier: no zone redundancy, no IKEv2 point-to-site, no in-place resize to other SKUs. Not retiring, but new BASIC gateways must use a Standard public IP.
- `STANDARD` -- Legacy tier (predates the VpnGw/ErGw families). ExpressRoute gateways only: Azure retired its VPN use on 2025-09-30 (the vpn_sku_vocabulary contract rejects it).
- `HIGH_PERFORMANCE` -- Legacy tier (predates the VpnGw/ErGw families). ExpressRoute gateways only: Azure retired its VPN use on 2025-09-30 (the vpn_sku_vocabulary contract rejects it).
- `ULTRA_PERFORMANCE` -- ExpressRoute-only legacy top tier.
- `VPN_GW_1_AZ` -- VPN tier 1 (Generation1 only, ~650 Mbps aggregate) -- the production entry point. Zone-redundant where the region has availability zones, regional elsewhere.
- `VPN_GW_2_AZ` -- VPN tier 2 (Generation1 or Generation2). Zone-redundant where the region has availability zones.
- `VPN_GW_3_AZ` -- VPN tier 3 (Generation1 or Generation2). Zone-redundant where the region has availability zones.
- `VPN_GW_4_AZ` -- VPN tier 4 (Generation2 only). Zone-redundant where the region has availability zones.
- `VPN_GW_5_AZ` -- VPN tier 5 (Generation2 only). Zone-redundant where the region has availability zones.
- `ER_GW_1_AZ` -- Zone-redundant ExpressRoute gateway, size 1.
- `ER_GW_2_AZ` -- Zone-redundant ExpressRoute gateway, size 2.
- `ER_GW_3_AZ` -- Zone-redundant ExpressRoute gateway, size 3.
- `ER_GW_SCALE` -- Autoscaling ExpressRoute gateway -- requires minimum_scale_unit and maximum_scale_unit (Terraform engine only; see the spec's parity note).

### spec.generation

`enum`

The VPN gateway generation. GENERATION2 doubles throughput ceilings
and requires VPN_GW_2_AZ or higher. Fixed at creation. Unspecified
lets Azure pick the default generation for the SKU. Not applicable
to ExpressRoute gateways (leave unset or NONE).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_generation_unspecified` -- Not specified -- Azure picks the default generation for the SKU.
- `GENERATION1` -- First generation: supports BASIC and VPN_GW_1_AZ..3_AZ.
- `GENERATION2` -- Second generation: doubled throughput ceilings; requires VPN_GW_2_AZ or higher.
- `NONE` -- No generation -- ExpressRoute gateways.

### spec.ipConfigurations

`[]AzureVirtualNetworkGatewayIpConfiguration` · required

The gateway's IP configurations, each binding a public IP on the
dedicated "GatewaySubnet". One configuration is the norm; ACTIVE-ACTIVE
VPN gateways need two (each with its own public IP); three supports
active-active with a separate P2S configuration. FIXED AT CREATION --
any change replaces the gateway. VPN gateways require a public IP on
every configuration; ExpressRoute gateways must NOT set public IPs
(Azure manages their addressing).

- rule: {"repeated":{"minItems":"1","maxItems":"3"}}

### spec.ipConfigurations[].name

`string`

The configuration's name, unique on the gateway. Defaults to
"vnetGatewayConfig" (the name the Azure portal uses) when empty.

### spec.ipConfigurations[].subnetId

`string | valueFrom` · required

The gateway's subnet -- references an AzureSubnet's ARM id. ARM
requires the subnet's name to be EXACTLY "GatewaySubnet" (/27 or
larger recommended); the subnet carries nothing but gateways. Fixed
at creation.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.ipConfigurations[].publicIpAddressId

`string | valueFrom`

The public IP tunnels terminate on -- references an AzurePublicIp's
ARM id (Standard SKU, static). REQUIRED on VPN gateways (every
configuration); FORBIDDEN on ExpressRoute gateways (Azure manages
their addressing). The AZ VPN SKUs (the only creatable VpnGw tiers)
require the address to carry ZONES -- ARM rejects a no-zone
Standard IP with VmssVpnGatewayPublicIpsMustHaveZonesConfigured, so
give the AzurePublicIp zones ["1","2","3"] (zone-redundant) unless
you are deliberately pinning a zone. A cross-resource apply-time
contract: the zones live on the referenced address, so it cannot be
validated here.

- references: AzurePublicIp (`status.outputs.public_ip_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIp, name: <that resource's name>, fieldPath: status.outputs.public_ip_id}} -- a bare string does not parse

### spec.ipConfigurations[].privateIpAddressAllocation

`enum`

How the gateway's PRIVATE IP on the GatewaySubnet is assigned.
Unspecified uses DYNAMIC, Azure's default.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_ip_allocation_unspecified` -- Not specified -- uses DYNAMIC, Azure's default.
- `DYNAMIC` -- Azure assigns the private IP.
- `STATIC` -- The private IP is fixed for the gateway's lifetime.

### spec.activeActive

`bool`

Run the VPN gateway as an ACTIVE-ACTIVE pair: two gateway instances,
each with its own ip_configuration and public IP, both terminating
tunnels simultaneously (higher availability, no failover gap).
Requires at least two ip_configurations. VPN gateways only.

### spec.privateIpAddressEnabled

`bool`

Also assign the gateway a PRIVATE IP from the GatewaySubnet
(required for private-peering scenarios like ExpressRoute private
connectivity to the gateway). Fixed at creation.

### spec.edgeZone

`string`

The extended-zone (Edge Zone) to deploy the gateway in. Rarely used;
leave empty for regional deployment. Fixed at creation.

### spec.bgpEnabled

`bool`

Enable BGP on the gateway -- required for route-based tunnels that
exchange routes dynamically (and for ExpressRoute route exchange,
where Azure enables it implicitly). Tune the speaker via
bgp_settings.

### spec.bgpSettings

`AzureVirtualNetworkGatewayBgpSettings`

The gateway's BGP speaker settings: its ASN, route weight, and
per-ip-configuration APIPA peering addresses (for tunnels whose
peers require link-local BGP endpoints, e.g. AWS site-to-site VPN).

### spec.bgpSettings.asn

`int64`

The gateway's Autonomous System Number. Azure defaults to 65515;
private ASNs 64512-65514 and 65521-65534 are safe custom choices
(65515-65520 are Azure-reserved).

- rule: {"int64":{"gte":"0"}}

### spec.bgpSettings.peerWeight

`int32`

The weight added to routes learned over BGP, 0-100. Higher wins in
route selection on the gateway.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.bgpSettings.peeringAddresses

`[]AzureVirtualNetworkGatewayBgpPeeringAddress`

Per-ip-configuration BGP peering addresses. One entry unless the
gateway is active-active (then one per ip_configuration, each named
explicitly).

- rule: {"repeated":{"maxItems":"2"}}

### spec.bgpSettings.peeringAddresses[].ipConfigurationName

`string`

The ip_configuration this entry applies to. May be omitted when the
gateway has exactly one ip_configuration; REQUIRED (by ARM) when it
has several.

### spec.bgpSettings.peeringAddresses[].apipaAddresses

`[]string` · required

Azure custom APIPA addresses assigned to this BGP peer -- the
link-local endpoints some peers (e.g. AWS site-to-site VPN) require.
Azure public regions accept 169.254.21.0 through 169.254.22.255.

- rule: {"repeated":{"minItems":"1","items":{"string":{"ipv4":true}}}}

### spec.customRouteAddressPrefixes

`[]string`

Custom address prefixes the gateway advertises to ALL connected
clients and tunnels beyond the VNet's own space (the gateway's
"custom routes"). CIDR notation.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultLocalNetworkGatewayId

`string | valueFrom`

FORCED TUNNELING: the local network gateway whose address space
becomes the gateway's default route -- on-premises egress inspection
for all VNet traffic. References an AzureLocalNetworkGateway's ARM
id. Leave unset for normal split routing.

- references: AzureLocalNetworkGateway (`status.outputs.local_network_gateway_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLocalNetworkGateway, name: <that resource's name>, fieldPath: status.outputs.local_network_gateway_id}} -- a bare string does not parse

### spec.vpnClientConfiguration

`AzureVirtualNetworkGatewayVpnClientConfiguration`

POINT-TO-SITE: the client VPN configuration -- address pool,
authentication (Entra ID, certificate, or RADIUS), tunnel protocols,
and per-group routing. VPN gateways only; requires a route-based
gateway on VPN_GW_1_AZ or higher for IKEv2/OpenVPN.

- rule: Entra ID authentication needs all three of aad_tenant, aad_audience, and aad_issuer
- rule: radius_server_address and radius_server_secret are set together

### spec.vpnClientConfiguration.addressSpaces

`[]string` · required

The address pool VPN clients draw from, in CIDR notation. Must not
overlap the VNet or any connected network.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.vpnClientConfiguration.aadTenant

`string`

Entra ID (Azure AD) authentication -- the tenant URL, e.g.
"https://login.microsoftonline.com/{tenant-id}". Set with
aad_audience and aad_issuer (the three travel together).

### spec.vpnClientConfiguration.aadAudience

`string`

Entra ID authentication -- the Azure VPN application's client id.

### spec.vpnClientConfiguration.aadIssuer

`string`

Entra ID authentication -- the STS issuer URL, e.g.
"https://sts.windows.net/{tenant-id}/".

### spec.vpnClientConfiguration.rootCertificates

`[]AzureVirtualNetworkGatewayVpnClientRootCertificate`

Certificate authentication -- root certificates whose chains sign
client certificates.

### spec.vpnClientConfiguration.rootCertificates[].name

`string` · required

The certificate's name on the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnClientConfiguration.rootCertificates[].publicCertData

`string` · required

The base64-encoded public certificate data (the .cer body, without
the PEM header/footer lines).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnClientConfiguration.revokedCertificates

`[]AzureVirtualNetworkGatewayVpnClientRevokedCertificate`

Certificate authentication -- individually revoked client
certificates (by thumbprint).

### spec.vpnClientConfiguration.revokedCertificates[].name

`string` · required

The revocation entry's name on the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnClientConfiguration.revokedCertificates[].thumbprint

`string` · required

The revoked client certificate's SHA-1 thumbprint.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnClientConfiguration.radiusServerAddress

`string`

RADIUS authentication -- the single-server form. Set with
radius_server_secret; use radius_servers for multiple servers.

### spec.vpnClientConfiguration.radiusServerSecret

`string | valueFrom` · sensitive

The RADIUS shared secret for radius_server_address. Reference a
secret rather than embedding the literal in manifests.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vpnClientConfiguration.radiusServers

`[]AzureVirtualNetworkGatewayVpnClientRadiusServer`

RADIUS authentication -- the multi-server form (each with its own
secret and score).

### spec.vpnClientConfiguration.radiusServers[].address

`string` · required

The RADIUS server's IPv4 address.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.vpnClientConfiguration.radiusServers[].secret

`string | valueFrom` · required · sensitive

The RADIUS shared secret (1-128 characters). Reference a secret
rather than embedding the literal in manifests. Azure never returns
it on reads.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vpnClientConfiguration.radiusServers[].score

`int32`

The server's priority score, 1-30 (lower is preferred).

- rule: {"int32":{"lte":30,"gte":1}}

### spec.vpnClientConfiguration.ipsecPolicy

`AzureVirtualNetworkGatewayVpnClientIpsecPolicy`

A custom IPsec policy for point-to-site IKEv2/OpenVPN connections.
Leave unset to use Azure's defaults.

### spec.vpnClientConfiguration.ipsecPolicy.dhGroup

`string`

The IKE Phase 1 Diffie-Hellman group.

- rule: {"string":{"in":["DHGroup1","DHGroup14","DHGroup2","DHGroup2048","DHGroup24","ECP256","ECP384","None"]}}

### spec.vpnClientConfiguration.ipsecPolicy.ikeEncryption

`string`

The IKE encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES256"]}}

### spec.vpnClientConfiguration.ipsecPolicy.ikeIntegrity

`string`

The IKE integrity algorithm.

- rule: {"string":{"in":["GCMAES128","GCMAES256","MD5","SHA1","SHA256","SHA384"]}}

### spec.vpnClientConfiguration.ipsecPolicy.ipsecEncryption

`string`

The IPsec (Phase 2) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES192","GCMAES256","None"]}}

### spec.vpnClientConfiguration.ipsecPolicy.ipsecIntegrity

`string`

The IPsec (Phase 2) integrity algorithm.

- rule: {"string":{"in":["GCMAES128","GCMAES192","GCMAES256","MD5","SHA1","SHA256"]}}

### spec.vpnClientConfiguration.ipsecPolicy.pfsGroup

`string`

The Perfect Forward Secrecy group.

- rule: {"string":{"in":["ECP256","ECP384","None","PFS1","PFS14","PFS2","PFS2048","PFS24","PFSMM"]}}

### spec.vpnClientConfiguration.ipsecPolicy.saLifetimeSeconds

`int32`

The security association lifetime in seconds, 300-172799.

- rule: {"int32":{"lte":172799,"gte":300}}

### spec.vpnClientConfiguration.ipsecPolicy.saDataSizeKilobytes

`int32`

The security association size limit in kilobytes, at least 1024.

- rule: {"int32":{"gte":1024}}

### spec.vpnClientConfiguration.vpnClientProtocols

`[]string`

The tunnel protocols clients may use: "OpenVPN" (recommended),
"IkeV2", and/or "SSTP". Unset lets Azure pick its default set.

- rule: {"repeated":{"items":{"string":{"in":["IkeV2","OpenVPN","SSTP"]}}}}

### spec.vpnClientConfiguration.vpnAuthTypes

`[]string`

The enabled authentication types: "Certificate", "AAD" (Entra ID),
and/or "Radius". REQUIRED by Azure when combining multiple types;
unset lets Azure infer the single configured type.

- rule: {"repeated":{"maxItems":"3","items":{"string":{"in":["Certificate","AAD","Radius"]}}}}

### spec.vpnClientConfiguration.clientConnections

`[]AzureVirtualNetworkGatewayClientConnection`

Per-policy-group client connection pools: map policy groups (defined
in the spec's policy_groups) to dedicated address prefixes.

### spec.vpnClientConfiguration.clientConnections[].name

`string` · required

The connection configuration's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnClientConfiguration.clientConnections[].policyGroupNames

`[]string` · required

The policy groups (by name, from the spec's policy_groups) whose
members receive addresses from this pool.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.vpnClientConfiguration.clientConnections[].addressPrefixes

`[]string` · required

The address prefixes this group's clients draw from, in CIDR
notation (carved from the vpn_client_configuration address pool).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.policyGroups

`[]AzureVirtualNetworkGatewayPolicyGroup`

Policy groups for point-to-site connection segmentation: members are
matched by Entra ID group, certificate CN, or RADIUS attribute, and
vpn_client_configuration.client_connections maps each group to its
own address pool.

### spec.policyGroups[].name

`string` · required

The policy group's name (referenced by
vpn_client_configuration.client_connections).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policyGroups[].policyMembers

`[]AzureVirtualNetworkGatewayPolicyMember` · required

The membership rules -- a client matching ANY member joins the
group.

- rule: {"repeated":{"minItems":"1"}}

### spec.policyGroups[].policyMembers[].name

`string` · required

The member rule's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policyGroups[].policyMembers[].type

`string`

The attribute matched: "AADGroupId" (Entra ID group object id),
"CertificateGroupId" (certificate CN), or "RadiusAzureGroupId"
(RADIUS-provided group).

- rule: {"string":{"in":["AADGroupId","CertificateGroupId","RadiusAzureGroupId"]}}

### spec.policyGroups[].policyMembers[].value

`string` · required

The attribute value to match (e.g. the Entra ID group's object id).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.policyGroups[].isDefault

`bool`

Whether this is the gateway's default policy group (for clients
matching no other group).

### spec.policyGroups[].priority

`int32`

The group's evaluation priority (lower evaluates first). Azure
defaults to 0.

- rule: {"int32":{"gte":0}}

### spec.bgpRouteTranslationForNatEnabled

`bool`

Translate BGP-learned routes through the gateway's NAT rules --
advertise post-NAT prefixes instead of the raw on-premises ones.
Only meaningful when nat_rules are configured.

### spec.dnsForwardingEnabled

`bool`

Let point-to-site clients resolve names through Azure DNS via the
gateway. Sent only when enabled -- ARM rejects the parameter on
SKUs/types that do not support DNS forwarding.

### spec.ipSecReplayProtectionEnabled

`bool` · optional (explicit presence)

IPsec replay protection (anti-replay windows on tunnels). Azure and
the provider default this ON; disable only for peers whose replay
windows misbehave.

- default: `true`

### spec.minimumScaleUnit

`int32` · optional (explicit presence)

ER_GW_SCALE autoscale floor, 1-40 scale units. REQUIRED (with
maximum_scale_unit) when sku is ER_GW_SCALE; invalid on any other
SKU.

PARITY-EXCEPTION: the Pulumi engine's classic SDK does not expose
the autoscale bounds, so ER_GW_SCALE gateways deploy via the
Terraform engine only -- the Pulumi module fails loudly when these
are set.

- rule: {"int32":{"lte":40,"gte":1}}

### spec.maximumScaleUnit

`int32` · optional (explicit presence)

ER_GW_SCALE autoscale ceiling, 1-40 scale units; must be >=
minimum_scale_unit. See minimum_scale_unit for the engine parity
note.

- rule: {"int32":{"lte":40,"gte":1}}

### spec.remoteVnetTrafficEnabled

`bool`

Allow traffic from remote (peered) virtual networks to transit this
gateway -- the hub side of hub-spoke gateway transit toward
non-Virtual-WAN remotes.

### spec.virtualWanTrafficEnabled

`bool`

Allow traffic from Virtual WAN networks to transit this gateway.

### spec.natRules

`[]AzureVirtualNetworkGatewayNatRule`

NAT rules applied on the gateway -- translate overlapping address
space between on-premises sites and the VNet. Connections opt into
specific rules via their egress/ingress NAT rule id lists; the
gateway publishes each rule's ARM id in the nat_rule_ids output.

### spec.natRules[].name

`string` · required

The rule's name, unique on the gateway. The rule's ARM id surfaces
in the gateway's nat_rule_ids output under this name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].mode

`enum`

The translation direction: EGRESS_SNAT (translate the VNet-side
source -- the default) or INGRESS_SNAT (translate the
on-premises-side source).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_nat_rule_mode_unspecified` -- Not specified -- uses EGRESS_SNAT, the default direction.
- `EGRESS_SNAT` -- Translate the VNet-side source address space.
- `INGRESS_SNAT` -- Translate the on-premises-side source address space.

### spec.natRules[].type

`enum`

The translation type: STATIC_NAT (one-to-one, no port translation
-- the default) or DYNAMIC_NAT (many-to-one with port translation).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_nat_rule_type_unspecified` -- Not specified -- uses STATIC_NAT, one-to-one translation.
- `STATIC_NAT` -- One-to-one address translation without ports (wire value "Static").
- `DYNAMIC_NAT` -- Many-to-one translation with port translation (wire value "Dynamic").

### spec.natRules[].externalMappings

`[]AzureVirtualNetworkGatewayNatRuleMapping` · required

The external (post-translation, as seen by the remote side)
mappings.

- rule: {"repeated":{"minItems":"1"}}

### spec.natRules[].externalMappings[].addressSpace

`string` · required

The address space in CIDR notation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].externalMappings[].portRange

`string`

The port range (e.g. "100-200"). Dynamic rules only.

### spec.natRules[].internalMappings

`[]AzureVirtualNetworkGatewayNatRuleMapping` · required

The internal (pre-translation, VNet-side) mappings.

- rule: {"repeated":{"minItems":"1"}}

### spec.natRules[].internalMappings[].addressSpace

`string` · required

The address space in CIDR notation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].internalMappings[].portRange

`string`

The port range (e.g. "100-200"). Dynamic rules only.

### spec.natRules[].ipConfigurationId

`string`

Pin the rule to one gateway ip_configuration by its ARM id (of the
form {gateway-id}/ipConfigurations/{name}). Rarely needed -- leave
empty to apply on all configurations.

### spec.tags

`map<string, string>`

Free-form tags applied to the gateway, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `sku_required`: Choose the gateway SKU explicitly -- it sizes throughput, cost, and zone redundancy (VPN_GW_1_AZ is the common production entry point for VPN gateways)
- `vpn_sku_vocabulary`: VPN gateways use BASIC or the VPN_GW_1_AZ..5_AZ SKUs (Azure retired new non-AZ VpnGw and legacy Standard/HighPerformance VPN creates; AZ SKUs deploy in every region) -- the ER_GW/ULTRA_PERFORMANCE SKUs are ExpressRoute-only
- `express_route_sku_vocabulary`: ExpressRoute gateways use STANDARD, HIGH_PERFORMANCE, ULTRA_PERFORMANCE, ER_GW_1_AZ..3_AZ, or ER_GW_SCALE
- `policy_based_requires_basic_sku`: Policy-based VPN gateways support only the BASIC SKU (legacy IKEv1 -- prefer route-based for anything new)
- `generation1_sku_vocabulary`: Generation1 VPN gateways use BASIC or VPN_GW_1_AZ..3_AZ
- `generation2_sku_vocabulary`: Generation2 VPN gateways start at VPN_GW_2_AZ: use VPN_GW_2_AZ..5_AZ
- `generation_is_vpn_only`: The generation knob applies to VPN gateways only -- leave it unset (or NONE) on ExpressRoute gateways
- `express_route_gateway_has_no_public_ips`: ExpressRoute gateways get Azure-managed addressing -- remove public_ip_address_id from every ip_configuration
- `vpn_gateway_requires_public_ips`: VPN gateways require a public IP on every ip_configuration (tunnels terminate on it)
- `active_active_requires_two_ip_configurations`: An active-active gateway is a two-instance pair -- give it (at least) two ip_configurations, each with its own public IP
- `active_active_is_vpn_only`: Active-active mode applies to VPN gateways only
- `vpn_client_configuration_is_vpn_only`: Point-to-site (vpn_client_configuration) runs on VPN gateways only
- `scale_units_require_each_other`: minimum_scale_unit and maximum_scale_unit are set together (both bound the ER_GW_SCALE autoscaler)
- `scale_units_only_on_ergwscale`: Autoscale bounds apply only to the ER_GW_SCALE SKU
- `ergwscale_requires_scale_units`: The ER_GW_SCALE SKU requires minimum_scale_unit and maximum_scale_unit (Azure's autoscale contract)
- `scale_unit_floor_not_above_ceiling`: minimum_scale_unit cannot exceed maximum_scale_unit
- `nat_rule_names_unique`: NAT rule names must be unique on the gateway

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualNetworkGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_network_gateway_id` | `string` | The Azure Resource Manager ID of the gateway -- what connections reference as virtual_network_gateway_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworkGateways/{name} |
| `status.outputs.virtual_network_gateway_name` | `string` | The name of the gateway resource. |
| `status.outputs.nat_rule_ids` | `map<string, string>` | The ARM ids of the gateway's NAT rules, keyed by rule name -- connections opt into rules via their egress/ingress NAT rule id lists. Empty when the spec defines no nat_rules. (The gateway's PUBLIC address is not an output here: it belongs to the referenced AzurePublicIp resource and surfaces through that kind's outputs.) |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.ipConfigurations[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.ipConfigurations[].publicIpAddressId` | AzurePublicIp | `status.outputs.public_ip_id` |
| `spec.defaultLocalNetworkGatewayId` | AzureLocalNetworkGateway | `status.outputs.local_network_gateway_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualNetworkGatewayConnection | `spec.virtualNetworkGatewayId` | `status.outputs.virtual_network_gateway_id` |
| AzureVirtualNetworkGatewayConnection | `spec.peerVirtualNetworkGatewayId` | `status.outputs.virtual_network_gateway_id` |

## See Also

- [Overview](../README.md)

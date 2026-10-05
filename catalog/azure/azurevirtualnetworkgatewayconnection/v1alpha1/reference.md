# AzureVirtualNetworkGatewayConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVirtualNetworkGatewayConnectionSpec** defines a connection on a
virtual network gateway -- the tunnel object that joins the gateway to
its far side. Three connection types, selected by `type`:

- **IPSEC** (site-to-site): the gateway to an on-premises VPN device,
  described by an AzureLocalNetworkGateway. The classic
  datacenter-to-Azure tunnel. Requires local_network_gateway_id.
- **VNET_TO_VNET**: the gateway to another virtual network gateway --
  an encrypted tunnel between two VNets (typically across regions;
  same-region pairs usually prefer VNet peering, which is cheaper and
  faster). Requires peer_virtual_network_gateway_id, and a mirror
  connection on the far gateway with the same shared_key.
- **EXPRESS_ROUTE**: the gateway to an ExpressRoute circuit's private
  peering. Requires express_route_circuit_id.

**Provisioned is not connected**: ARM provisions the connection object
as soon as the parameters are valid; the tunnel itself only reaches
its Connected state when the far side (the on-premises device or peer
gateway) negotiates successfully. A Succeeded deployment with a
tunnel stuck in Connecting means the far side, the shared key, or the
IPsec parameters disagree -- not that the deployment failed.

**ForceNew fields**: `name`, `region`, `resource_group`, `type`,
`virtual_network_gateway_id`, `peer_virtual_network_gateway_id`,
`express_route_circuit_id`, `dpd_timeout_seconds`,
`local_azure_ip_address_enabled`, `connection_protocol`, and
`connection_mode` -- changing any of them replaces the connection
(cheap: seconds, no gateway rebuild).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualNetworkGatewayConnection
metadata:
  name: test-gateway-connection
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hq-to-azure
  # IPSEC (site-to-site) joins the gateway to an on-premises device
  # described by an AzureLocalNetworkGateway.
  type: IPSEC
  virtualNetworkGatewayId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworkGateways/hub-vpn-gateway
  localNetworkGatewayId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/localNetworkGateways/hq-datacenter
  # The pre-shared key both tunnel ends must agree on. Reference a secret
  # in real manifests; omit to let Azure generate one.
  sharedKey:
    value: replace-with-a-strong-pre-shared-key
  dpdTimeoutSeconds: 45
  # A pinned IPsec/IKE proposal for devices that need exact algorithms;
  # omit to use Azure's default proposal set.
  ipsecPolicy:
    dhGroup: DHGroup14
    ikeEncryption: AES256
    ikeIntegrity: SHA256
    ipsecEncryption: AES256
    ipsecIntegrity: SHA256
    pfsGroup: PFS2048
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
| `spec.virtualNetworkGatewayId` | `string \| valueFrom` | yes |  | AzureVirtualNetworkGateway (`status.outputs.virtual_network_gateway_id`) |
| `spec.localNetworkGatewayId` | `string \| valueFrom` |  |  | AzureLocalNetworkGateway (`status.outputs.local_network_gateway_id`) |
| `spec.peerVirtualNetworkGatewayId` | `string \| valueFrom` |  |  | AzureVirtualNetworkGateway (`status.outputs.virtual_network_gateway_id`) |
| `spec.expressRouteCircuitId` | `string \| valueFrom` |  |  | AzureExpressRouteCircuit (`status.outputs.express_route_circuit_id`) |
| `spec.sharedKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.authorizationKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.bgpEnabled` | `bool` |  |  |  |
| `spec.customBgpAddresses` | `AzureVirtualNetworkGatewayConnectionCustomBgpAddresses` |  |  |  |
| `spec.customBgpAddresses.primary` | `string` | yes |  |  |
| `spec.customBgpAddresses.secondary` | `string` |  |  |  |
| `spec.dpdTimeoutSeconds` | `int32` |  |  |  |
| `spec.connectionProtocol` | `enum` |  |  |  |
| `spec.connectionMode` | `enum` |  |  |  |
| `spec.routingWeight` | `int32` |  |  |  |
| `spec.egressNatRuleIds` | `[]string \| valueFrom` |  |  |  |
| `spec.ingressNatRuleIds` | `[]string \| valueFrom` |  |  |  |
| `spec.usePolicyBasedTrafficSelectors` | `bool` |  |  |  |
| `spec.expressRouteGatewayBypass` | `bool` |  |  |  |
| `spec.privateLinkFastPathEnabled` | `bool` |  |  |  |
| `spec.localAzureIpAddressEnabled` | `bool` |  |  |  |
| `spec.trafficSelectorPolicies` | `[]AzureVirtualNetworkGatewayConnectionTrafficSelectorPolicy` |  |  |  |
| `spec.trafficSelectorPolicies[].localAddressCidrs` | `[]string` | yes |  |  |
| `spec.trafficSelectorPolicies[].remoteAddressCidrs` | `[]string` | yes |  |  |
| `spec.ipsecPolicy` | `AzureVirtualNetworkGatewayConnectionIpsecPolicy` |  |  |  |
| `spec.ipsecPolicy.dhGroup` | `string` |  |  |  |
| `spec.ipsecPolicy.ikeEncryption` | `string` |  |  |  |
| `spec.ipsecPolicy.ikeIntegrity` | `string` |  |  |  |
| `spec.ipsecPolicy.ipsecEncryption` | `string` |  |  |  |
| `spec.ipsecPolicy.ipsecIntegrity` | `string` |  |  |  |
| `spec.ipsecPolicy.pfsGroup` | `string` |  |  |  |
| `spec.ipsecPolicy.saDatasize` | `int32` |  |  |  |
| `spec.ipsecPolicy.saLifetime` | `int32` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the connection lives in, e.g. "eastus". Must match
the gateway's region. Changing it replaces the connection.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the connection is created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The connection's name, unique within the resource group. Changing
the name replaces the connection.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.type

`enum`

What the connection joins the gateway to: IPSEC (an on-premises
device), VNET_TO_VNET (another gateway), or EXPRESS_ROUTE (a
circuit). Fixed at creation, and it decides which far-side
reference is required.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_connection_type_unspecified` -- Not specified -- invalid: the connection type decides the required far side (see the type_required contract).
- `IPSEC` -- Site-to-site: to an on-premises VPN device described by an AzureLocalNetworkGateway.
- `VNET_TO_VNET` -- To another virtual network gateway -- an encrypted VNet-to-VNet tunnel.
- `EXPRESS_ROUTE` -- To an ExpressRoute circuit's private peering.

### spec.virtualNetworkGatewayId

`string | valueFrom` · required

The virtual network gateway this connection belongs to -- references
an AzureVirtualNetworkGateway's ARM id. Fixed at creation.

- references: AzureVirtualNetworkGateway (`status.outputs.virtual_network_gateway_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetworkGateway, name: <that resource's name>, fieldPath: status.outputs.virtual_network_gateway_id}} -- a bare string does not parse

### spec.localNetworkGatewayId

`string | valueFrom`

IPSEC connections: the on-premises side -- references an
AzureLocalNetworkGateway's ARM id (the object carrying the device's
public IP/FQDN and reachable prefixes).

- references: AzureLocalNetworkGateway (`status.outputs.local_network_gateway_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLocalNetworkGateway, name: <that resource's name>, fieldPath: status.outputs.local_network_gateway_id}} -- a bare string does not parse

### spec.peerVirtualNetworkGatewayId

`string | valueFrom`

VNET_TO_VNET connections: the far-side virtual network gateway --
references an AzureVirtualNetworkGateway's ARM id. The far side
needs its own mirror connection back to this gateway carrying the
same shared_key. Fixed at creation.

- references: AzureVirtualNetworkGateway (`status.outputs.virtual_network_gateway_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetworkGateway, name: <that resource's name>, fieldPath: status.outputs.virtual_network_gateway_id}} -- a bare string does not parse

### spec.expressRouteCircuitId

`string | valueFrom`

EXPRESS_ROUTE connections: the circuit's ARM id. References an
AzureExpressRouteCircuit's ARM-id output (or a literal ARM ID for a
circuit outside the catalog). Fixed at creation.

- references: AzureExpressRouteCircuit (`status.outputs.express_route_circuit_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureExpressRouteCircuit, name: <that resource's name>, fieldPath: status.outputs.express_route_circuit_id}} -- a bare string does not parse

### spec.sharedKey

`string | valueFrom` · sensitive

The IPsec pre-shared key both tunnel ends must agree on (IPSEC and
VNET_TO_VNET connections). Reference a secret rather than embedding
the literal in manifests. Omit to let Azure generate one (readable
back from the connection's shared-key API).

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.authorizationKey

`string | valueFrom` · sensitive

EXPRESS_ROUTE connections to a circuit in ANOTHER subscription: the
circuit authorization key its owner issued. Reference a secret
rather than embedding the literal.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.bgpEnabled

`bool`

Exchange routes over BGP instead of static address spaces. Both
tunnel ends must speak BGP (the gateway needs bgp_enabled and the
local network gateway its bgp_settings).

### spec.customBgpAddresses

`AzureVirtualNetworkGatewayConnectionCustomBgpAddresses`

Custom APIPA BGP endpoints for this connection (IPSEC only) --
pick which of the gateway's configured APIPA addresses this tunnel
peers from. Requires bgp_enabled and an active-active gateway with
APIPA peering addresses configured.

### spec.customBgpAddresses.primary

`string` · required

The primary APIPA address -- one of the gateway's first
ip_configuration's configured APIPA peering addresses.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.customBgpAddresses.secondary

`string`

The secondary APIPA address (active-active gateways) -- one of the
second ip_configuration's APIPA peering addresses.

- rule: secondary must be an IPv4 address

### spec.dpdTimeoutSeconds

`int32` · optional (explicit presence)

Dead Peer Detection timeout in seconds. Azure's default is 45.
Fixed at creation.

- rule: {"int32":{"gte":9}}

### spec.connectionProtocol

`enum`

The IKE protocol version (route-based IPSEC connections). Azure
defaults to IKEv2; IKEv1 exists for legacy devices. Fixed at
creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_connection_protocol_unspecified` -- Not specified -- Azure applies its default (IKEv2).
- `IKE_V1` -- IKE version 1 -- legacy devices only.
- `IKE_V2` -- IKE version 2 -- the modern default.

### spec.connectionMode

`enum`

Which side initiates the tunnel: DEFAULT (either), INITIATOR_ONLY,
or RESPONDER_ONLY. Fixed at creation. Unspecified deploys DEFAULT.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_network_gateway_connection_mode_unspecified` -- Not specified -- deploys DEFAULT (either side initiates).
- `DEFAULT` -- Either side may initiate.
- `INITIATOR_ONLY` -- Only this side initiates.
- `RESPONDER_ONLY` -- Only the far side initiates.

### spec.routingWeight

`int32` · optional (explicit presence)

The routing weight for this connection (0-32000; higher is
preferred among parallel connections).

- rule: {"int32":{"lte":32000,"gte":0}}

### spec.egressNatRuleIds

`[]string | valueFrom`

Gateway NAT rules applied to this connection's EGRESS (VNet ->
remote) traffic, by ARM id. The owning gateway publishes its rules'
ids in its nat_rule_ids output (a map keyed by rule name) -- supply
the ids as literals or explicit references.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ingressNatRuleIds

`[]string | valueFrom`

Gateway NAT rules applied to this connection's INGRESS (remote ->
VNet) traffic, by ARM id.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.usePolicyBasedTrafficSelectors

`bool`

Use policy-based traffic selectors on this route-based connection
(compatibility with policy-based on-premises devices). Requires a
custom ipsec_policy -- Azure rejects the flag without one.

### spec.expressRouteGatewayBypass

`bool`

EXPRESS_ROUTE: bypass the gateway for data-path traffic (FastPath)
-- lower latency, higher throughput; the gateway stays for route
exchange only.

### spec.privateLinkFastPathEnabled

`bool`

EXPRESS_ROUTE FastPath for Private Link traffic. Requires
express_route_gateway_bypass.

### spec.localAzureIpAddressEnabled

`bool`

Terminate the tunnel on the gateway's PRIVATE IP instead of its
public one (for tunnels that ride ExpressRoute private peering).
The gateway needs private_ip_address_enabled. Fixed at creation.

### spec.trafficSelectorPolicies

`[]AzureVirtualNetworkGatewayConnectionTrafficSelectorPolicy`

Restrict the tunnel to specific address pairs: each policy names
local and remote CIDR sets. Most connections leave this empty (the
address spaces come from the gateways).

PARITY-EXCEPTION: the Pulumi engine's classic SDK models exactly ONE
traffic selector policy -- manifests carrying more than one deploy
via the Terraform engine only (the Pulumi module fails loudly).

### spec.trafficSelectorPolicies[].localAddressCidrs

`[]string` · required

The local (Azure-side) address ranges in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.trafficSelectorPolicies[].remoteAddressCidrs

`[]string` · required

The remote (far-side) address ranges in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.ipsecPolicy

`AzureVirtualNetworkGatewayConnectionIpsecPolicy`

A custom IPsec/IKE policy for this connection. Leave unset to use
Azure's default proposal set; set it when the on-premises device
needs pinned algorithms.

### spec.ipsecPolicy.dhGroup

`string`

The IKE Phase 1 Diffie-Hellman group.

- rule: {"string":{"in":["DHGroup1","DHGroup14","DHGroup2","DHGroup2048","DHGroup24","ECP256","ECP384","None"]}}

### spec.ipsecPolicy.ikeEncryption

`string`

The IKE encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES256"]}}

### spec.ipsecPolicy.ikeIntegrity

`string`

The IKE integrity algorithm. With GCM IKE encryption, Azure requires
the MATCHING GCM integrity value.

- rule: {"string":{"in":["GCMAES128","GCMAES256","MD5","SHA1","SHA256","SHA384"]}}

### spec.ipsecPolicy.ipsecEncryption

`string`

The IPsec (Phase 2) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES192","GCMAES256","None"]}}

### spec.ipsecPolicy.ipsecIntegrity

`string`

The IPsec (Phase 2) integrity algorithm.

- rule: {"string":{"in":["GCMAES128","GCMAES192","GCMAES256","MD5","SHA1","SHA256"]}}

### spec.ipsecPolicy.pfsGroup

`string`

The Perfect Forward Secrecy group.

- rule: {"string":{"in":["ECP256","ECP384","None","PFS1","PFS14","PFS2","PFS2048","PFS24","PFSMM"]}}

### spec.ipsecPolicy.saDatasize

`int32` · optional (explicit presence)

The security association size limit in kilobytes. Omit for Azure's
default; when set, at least 1024.

- rule: {"int32":{"gte":1024}}

### spec.ipsecPolicy.saLifetime

`int32` · optional (explicit presence)

The security association lifetime in seconds. Omit for Azure's
default; when set, at least 300.

- rule: {"int32":{"gte":300}}

### spec.tags

`map<string, string>`

Free-form tags applied to the connection, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `type_required`: Choose the connection type explicitly: IPSEC (to an on-premises device), VNET_TO_VNET (to another gateway), or EXPRESS_ROUTE (to a circuit)
- `ipsec_requires_local_network_gateway`: An IPSEC (site-to-site) connection needs local_network_gateway_id -- the AzureLocalNetworkGateway describing the on-premises side
- `vnet_to_vnet_requires_peer_gateway`: A VNET_TO_VNET connection needs peer_virtual_network_gateway_id (and the far gateway needs a mirror connection with the same shared_key)
- `express_route_requires_circuit`: An EXPRESS_ROUTE connection needs express_route_circuit_id
- `custom_bgp_is_ipsec_only`: custom_bgp_addresses applies to IPSEC connections only
- `custom_bgp_requires_bgp`: custom_bgp_addresses requires bgp_enabled
- `private_link_fast_path_requires_bypass`: private_link_fast_path_enabled requires express_route_gateway_bypass (Azure's FastPath contract)
- `policy_based_selectors_require_ipsec_policy`: use_policy_based_traffic_selectors requires a custom ipsec_policy (Azure rejects the flag without one)
- `shared_key_not_for_express_route`: ExpressRoute connections carry no pre-shared key -- remove shared_key (use authorization_key for cross-subscription circuits)
- `authorization_key_is_express_route_only`: authorization_key authorizes an ExpressRoute circuit in another subscription -- it has no meaning on IPSEC or VNET_TO_VNET connections

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualNetworkGatewayConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.connection_id` | `string` | The Azure Resource Manager ID of the connection. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/connections/{name} |
| `status.outputs.connection_name` | `string` | The name of the connection resource. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualNetworkGatewayId` | AzureVirtualNetworkGateway | `status.outputs.virtual_network_gateway_id` |
| `spec.localNetworkGatewayId` | AzureLocalNetworkGateway | `status.outputs.local_network_gateway_id` |
| `spec.peerVirtualNetworkGatewayId` | AzureVirtualNetworkGateway | `status.outputs.virtual_network_gateway_id` |
| `spec.expressRouteCircuitId` | AzureExpressRouteCircuit | `status.outputs.express_route_circuit_id` |

## See Also

- [Overview](../README.md)

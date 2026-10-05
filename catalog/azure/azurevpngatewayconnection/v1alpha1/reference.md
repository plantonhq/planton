# AzureVpnGatewayConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVpnGatewayConnectionSpec** defines a VPN gateway connection --
the tunnel bundle joining ONE branch (an AzureVpnSite) to a hub's
VPN gateway. ARM addresses it as a child of the gateway.

**One vpn_link per site link**: each entry in vpn_links pins to one
of the site's links by ARM ID (the site publishes them in its
name-keyed link_ids output) and carries that tunnel's own IPsec,
BGP, and NAT choices. A two-link site gets a two-link connection --
that is Virtual WAN's active-active branch pattern.

**Provisioned is not connected**: ARM provisions the connection as
soon as the parameters are valid; each tunnel reaches its Connected
state only when the branch device negotiates successfully. A
Succeeded deployment with a tunnel stuck in Connecting means the
device, the shared key, or the IPsec parameters disagree -- not that
the deployment failed.

**ForceNew fields**: `name`, `vpn_gateway_id`, `remote_vpn_site_id`,
and each link's `vpn_site_link_id` and `bgp_enabled` -- everything
else updates in place.

The connection carries no region, resource group, or tags of its
own: ARM derives all three through the gateway it is a child of,
and the provider's schema has no tags argument.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnGatewayConnection
metadata:
  name: test-vpn-connection
spec:
  name: branch-london
  vpnGatewayId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnGateways/hub-vpn-gateway
  remoteVpnSiteId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnSites/branch-london
  internetSecurityEnabled: false
  routing:
    associatedRouteTableId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus/hubRouteTables/defaultRouteTable
    propagatedRouteTable:
      routeTableIds:
        - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus/hubRouteTables/defaultRouteTable
      labels:
        - default
  vpnLinks:
    - name: primary-isp
      vpnSiteLinkId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnSites/branch-london/vpnSiteLinks/primary-isp
      bandwidthMbps: 50
      dpdTimeoutSeconds: 45
      sharedKey:
        value: test-pre-shared-key
      ipsecPolicies:
        - saLifetimeSec: 3600
          saDataSizeKb: 102400000
          encryptionAlgorithm: AES256
          integrityAlgorithm: SHA256
          ikeEncryptionAlgorithm: AES256
          ikeIntegrityAlgorithm: SHA256
          dhGroup: DHGroup14
          pfsGroup: PFS2048
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.vpnGatewayId` | `string \| valueFrom` | yes |  | AzureVpnGateway (`status.outputs.vpn_gateway_id`) |
| `spec.remoteVpnSiteId` | `string \| valueFrom` | yes |  | AzureVpnSite (`status.outputs.vpn_site_id`) |
| `spec.internetSecurityEnabled` | `bool` |  |  |  |
| `spec.routing` | `AzureVpnGatewayConnectionRouting` |  |  |  |
| `spec.routing.associatedRouteTableId` | `string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.routing.inboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.routing.outboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.routing.propagatedRouteTable` | `AzureVpnGatewayConnectionPropagatedRouteTable` |  |  |  |
| `spec.routing.propagatedRouteTable.routeTableIds` | `[]string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.routing.propagatedRouteTable.labels` | `[]string` |  |  |  |
| `spec.vpnLinks` | `[]AzureVpnGatewayConnectionLink` | yes |  |  |
| `spec.vpnLinks[].name` | `string` | yes |  |  |
| `spec.vpnLinks[].vpnSiteLinkId` | `string \| valueFrom` | yes |  | AzureVpnSite (`status.outputs.link_ids`) |
| `spec.vpnLinks[].bandwidthMbps` | `int32` |  | `10` |  |
| `spec.vpnLinks[].protocol` | `enum` |  |  |  |
| `spec.vpnLinks[].connectionMode` | `enum` |  |  |  |
| `spec.vpnLinks[].routeWeight` | `int32` |  |  |  |
| `spec.vpnLinks[].dpdTimeoutSeconds` | `int32` |  |  |  |
| `spec.vpnLinks[].sharedKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.vpnLinks[].bgpEnabled` | `bool` |  |  |  |
| `spec.vpnLinks[].ratelimitEnabled` | `bool` |  |  |  |
| `spec.vpnLinks[].localAzureIpAddressEnabled` | `bool` |  |  |  |
| `spec.vpnLinks[].policyBasedTrafficSelectorEnabled` | `bool` |  |  |  |
| `spec.vpnLinks[].egressNatRuleIds` | `[]string \| valueFrom` |  |  | AzureVpnGateway (`status.outputs.nat_rule_ids`) |
| `spec.vpnLinks[].ingressNatRuleIds` | `[]string \| valueFrom` |  |  | AzureVpnGateway (`status.outputs.nat_rule_ids`) |
| `spec.vpnLinks[].ipsecPolicies` | `[]AzureVpnGatewayConnectionIpsecPolicy` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].saLifetimeSec` | `int32` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].saDataSizeKb` | `int32` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].encryptionAlgorithm` | `string` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].integrityAlgorithm` | `string` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].ikeEncryptionAlgorithm` | `string` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].ikeIntegrityAlgorithm` | `string` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].dhGroup` | `string` |  |  |  |
| `spec.vpnLinks[].ipsecPolicies[].pfsGroup` | `string` |  |  |  |
| `spec.vpnLinks[].customBgpAddresses` | `[]AzureVpnGatewayConnectionCustomBgpAddress` |  |  |  |
| `spec.vpnLinks[].customBgpAddresses[].ipAddress` | `string` | yes |  |  |
| `spec.vpnLinks[].customBgpAddresses[].ipConfigurationId` | `string` |  |  |  |
| `spec.trafficSelectorPolicies` | `[]AzureVpnGatewayConnectionTrafficSelectorPolicy` |  |  |  |
| `spec.trafficSelectorPolicies[].localAddressCidrs` | `[]string` | yes |  |  |
| `spec.trafficSelectorPolicies[].remoteAddressCidrs` | `[]string` | yes |  |  |

## Field Details

### spec.name

`string` · required

The connection's name, unique on the gateway. Name it after the
branch it connects ("branch-london"). Changing the name replaces
the connection (cheap: the tunnels renegotiate, the gateway
stays).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnGatewayId

`string | valueFrom` · required

The Virtual WAN VPN gateway this connection belongs to --
references an AzureVpnGateway's ARM ID. Fixed at creation.

- references: AzureVpnGateway (`status.outputs.vpn_gateway_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnGateway, name: <that resource's name>, fieldPath: status.outputs.vpn_gateway_id}} -- a bare string does not parse

### spec.remoteVpnSiteId

`string | valueFrom` · required

The branch being connected -- references an AzureVpnSite's ARM ID
(the address-book entry carrying the branch's links and reachable
prefixes). Fixed at creation.

- references: AzureVpnSite (`status.outputs.vpn_site_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnSite, name: <that resource's name>, fieldPath: status.outputs.vpn_site_id}} -- a bare string does not parse

### spec.internetSecurityEnabled

`bool`

Enable "internet security": the hub advertises a default route
(0.0.0.0/0) to this branch, so the branch's internet-bound traffic
rides the tunnel into Azure (typically into a hub firewall via
routing intent). Off by default (ARM's default) -- the branch
keeps its own internet egress.

### spec.routing

`AzureVpnGatewayConnectionRouting`

The connection's routing configuration. Leave unset for ARM's
default behavior: associate with and propagate to the hub's
built-in default route table (any-to-any reachability).

### spec.routing.associatedRouteTableId

`string | valueFrom` · required

The ONE hub route table this connection's traffic is routed by --
references the hub's default table (the hub's
default_route_table_id output, the default reference) or a custom
table via the hub's name-keyed map output, e.g. valueFrom
fieldPath "status.outputs.route_table_ids.branches".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.routing.inboundRouteMapId

`string | valueFrom`

A route map applied to routes ARRIVING from the branch --
references the hub's name-keyed route_map_ids output, e.g.
valueFrom fieldPath "status.outputs.route_map_ids.strip-communities".

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.routing.outboundRouteMapId

`string | valueFrom`

A route map applied to routes ADVERTISED to the branch -- same
referencing shape as inbound_route_map_id.

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.routing.propagatedRouteTable

`AzureVpnGatewayConnectionPropagatedRouteTable`

The hub route tables this connection's routes are PROPAGATED to
(where other networks learn the branch's prefixes from). Unset
lets the service default propagation.

### spec.routing.propagatedRouteTable.routeTableIds

`[]string | valueFrom` · required

Propagate to these specific hub route tables -- reference the
hub's default_route_table_id output (the default reference) or its
name-keyed route_table_ids map, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.branches". At least one (the
provider's contract on a configured block).

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.routing.propagatedRouteTable.labels

`[]string`

Also propagate to every hub route table carrying these labels
(labels are declared on the hub's route_tables; the built-in label
"default" targets the hub's default table).

### spec.vpnLinks

`[]AzureVpnGatewayConnectionLink` · required

The tunnels of this connection -- one per site link being
connected, each pinned to its link by ARM ID and carrying that
tunnel's own parameters. At least one.

- rule: {"repeated":{"minItems":"1"}}
- rule: policy_based_traffic_selector_enabled requires a custom ipsec_policies entry (Azure rejects the flag without one)

### spec.vpnLinks[].name

`string` · required

The tunnel's name, unique on the connection. By convention the
name of the site link it connects.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vpnLinks[].vpnSiteLinkId

`string | valueFrom` · required

The site link this tunnel connects, by ARM ID -- references the
owning AzureVpnSite's name-keyed link_ids output, e.g. valueFrom
fieldPath "status.outputs.link_ids.primary-isp". Fixed at
creation (ARM cannot repoint a tunnel at a different link).

- references: AzureVpnSite (`status.outputs.link_ids`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnSite, name: <that resource's name>, fieldPath: status.outputs.link_ids}} -- a bare string does not parse

### spec.vpnLinks[].bandwidthMbps

`int32` · optional (explicit presence)

The tunnel's expected bandwidth in Mbps. Unset applies the
provider's default of 10.

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.vpnLinks[].protocol

`enum`

The IKE protocol version. Unspecified deploys IKE_V2 (ARM's
default); IKE_V1 exists for legacy devices.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_connection_protocol_unspecified` -- Not specified -- Azure applies its default (IKEv2).
- `IKE_V1` -- IKE version 1 -- legacy devices only (wire value "IKEv1").
- `IKE_V2` -- IKE version 2 -- the modern default (wire value "IKEv2").

### spec.vpnLinks[].connectionMode

`enum`

Which side initiates the tunnel: DEFAULT (either),
INITIATOR_ONLY, or RESPONDER_ONLY. Unspecified deploys DEFAULT.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_connection_mode_unspecified` -- Not specified -- deploys DEFAULT (either side initiates).
- `DEFAULT` -- Either side may initiate (wire value "Default").
- `INITIATOR_ONLY` -- Only the gateway initiates (wire value "InitiatorOnly").
- `RESPONDER_ONLY` -- Only the branch initiates (wire value "ResponderOnly").

### spec.vpnLinks[].routeWeight

`int32`

The routing weight among this connection's tunnels (higher is
preferred). 0 is the provider's default.

- rule: {"int32":{"gte":0}}

### spec.vpnLinks[].dpdTimeoutSeconds

`int32` · optional (explicit presence)

Dead Peer Detection timeout in seconds (9-3600). Omit for ARM's
default of 45.

- rule: {"int32":{"lte":3600,"gte":9}}

### spec.vpnLinks[].sharedKey

`string | valueFrom` · sensitive

The IPsec pre-shared key both tunnel ends must agree on.
Reference a secret rather than embedding the literal in
manifests. Omit to let Azure generate one -- with a caveat
proven live: the generated key is NOT readable back through
this resource (reads return it empty), so when the branch
device must be configured with the key, set it explicitly via
a secret reference instead of relying on retrieval.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vpnLinks[].bgpEnabled

`bool`

Exchange routes over BGP instead of the site's static
address_cidrs. The site link must carry its bgp block. Fixed at
creation.

### spec.vpnLinks[].ratelimitEnabled

`bool`

Rate-limit the tunnel to its bandwidth_mbps. Off is ARM's
default.

### spec.vpnLinks[].localAzureIpAddressEnabled

`bool`

Terminate the tunnel on the gateway's PRIVATE IP instead of its
public one (for tunnels that ride ExpressRoute private peering).

### spec.vpnLinks[].policyBasedTrafficSelectorEnabled

`bool`

Use policy-based traffic selectors on this route-based tunnel
(compatibility with policy-based branch devices). Requires a
custom ipsec_policies entry -- Azure rejects the flag without one.

### spec.vpnLinks[].egressNatRuleIds

`[]string | valueFrom`

Gateway NAT rules applied to this tunnel's EGRESS (Azure ->
branch) traffic, by ARM id -- references the owning gateway's
name-keyed nat_rule_ids output, e.g. valueFrom fieldPath
"status.outputs.nat_rule_ids.branch-overlap".

- references: AzureVpnGateway (`status.outputs.nat_rule_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnGateway, name: <that resource's name>, fieldPath: status.outputs.nat_rule_ids}} -- a bare string does not parse

### spec.vpnLinks[].ingressNatRuleIds

`[]string | valueFrom`

Gateway NAT rules applied to this tunnel's INGRESS (branch ->
Azure) traffic, by ARM id -- same referencing shape as
egress_nat_rule_ids.

- references: AzureVpnGateway (`status.outputs.nat_rule_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnGateway, name: <that resource's name>, fieldPath: status.outputs.nat_rule_ids}} -- a bare string does not parse

### spec.vpnLinks[].ipsecPolicies

`[]AzureVpnGatewayConnectionIpsecPolicy`

Custom IPsec/IKE proposals for this tunnel. Leave empty to use
Azure's default proposal set; set when the branch device needs
pinned algorithms.

### spec.vpnLinks[].ipsecPolicies[].saLifetimeSec

`int32`

The security association lifetime in seconds (300-172799).

- rule: {"int32":{"lte":172799,"gte":300}}

### spec.vpnLinks[].ipsecPolicies[].saDataSizeKb

`int32`

The security association size limit in kilobytes (0 disables the
limit).

- rule: {"int32":{"gte":0}}

### spec.vpnLinks[].ipsecPolicies[].encryptionAlgorithm

`string`

The IPsec (Phase 2) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES192","GCMAES256","None"]}}

### spec.vpnLinks[].ipsecPolicies[].integrityAlgorithm

`string`

The IPsec (Phase 2) integrity algorithm.

- rule: {"string":{"in":["GCMAES128","GCMAES192","GCMAES256","MD5","SHA1","SHA256"]}}

### spec.vpnLinks[].ipsecPolicies[].ikeEncryptionAlgorithm

`string`

The IKE (Phase 1) encryption algorithm.

- rule: {"string":{"in":["AES128","AES192","AES256","DES","DES3","GCMAES128","GCMAES256"]}}

### spec.vpnLinks[].ipsecPolicies[].ikeIntegrityAlgorithm

`string`

The IKE (Phase 1) integrity algorithm. With GCM IKE encryption,
Azure requires the MATCHING GCM integrity value.

- rule: {"string":{"in":["GCMAES128","GCMAES256","MD5","SHA1","SHA256","SHA384"]}}

### spec.vpnLinks[].ipsecPolicies[].dhGroup

`string`

The IKE Phase 1 Diffie-Hellman group.

- rule: {"string":{"in":["DHGroup1","DHGroup14","DHGroup2","DHGroup2048","DHGroup24","ECP256","ECP384","None"]}}

### spec.vpnLinks[].ipsecPolicies[].pfsGroup

`string`

The Perfect Forward Secrecy group.

- rule: {"string":{"in":["ECP256","ECP384","None","PFS1","PFS14","PFS2","PFS2048","PFS24","PFSMM"]}}

### spec.vpnLinks[].customBgpAddresses

`[]AzureVpnGatewayConnectionCustomBgpAddress`

Custom APIPA BGP endpoints for this tunnel -- pick which of the
gateway's configured custom_ips each instance peers from
(ip_configuration_id "Instance0" or "Instance1").

### spec.vpnLinks[].customBgpAddresses[].ipAddress

`string` · required

The APIPA address this tunnel peers from -- one of the gateway's
configured custom_ips for the chosen instance.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.vpnLinks[].customBgpAddresses[].ipConfigurationId

`string`

Which gateway instance the address belongs to: "Instance0" or
"Instance1" (ARM's identifiers, matching the gateway's
instance_0/instance_1 custom_ips).

- rule: {"string":{"in":["Instance0","Instance1"]}}

### spec.trafficSelectorPolicies

`[]AzureVpnGatewayConnectionTrafficSelectorPolicy`

Restrict the tunnels to specific address pairs: each policy names
local and remote CIDR sets. Most connections leave this empty
(routing comes from the site's prefixes or BGP).

### spec.trafficSelectorPolicies[].localAddressCidrs

`[]string` · required

The local (Azure-side) address ranges in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.trafficSelectorPolicies[].remoteAddressCidrs

`[]string` · required

The remote (branch-side) address ranges in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

## Validation Rules

- `vpn_link_names_unique`: vpn_links names must be unique on the connection

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVpnGatewayConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.connection_id` | `string` | The Azure Resource Manager ID of the connection (a child of the gateway). Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/vpnGateways/{gateway}/vpnConnections/{name} |
| `status.outputs.connection_name` | `string` | The name of the connection. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpnGatewayId` | AzureVpnGateway | `status.outputs.vpn_gateway_id` |
| `spec.remoteVpnSiteId` | AzureVpnSite | `status.outputs.vpn_site_id` |
| `spec.routing.associatedRouteTableId` | AzureVirtualHub | `status.outputs.default_route_table_id` |
| `spec.routing.inboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.routing.outboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.routing.propagatedRouteTable.routeTableIds` | AzureVirtualHub | `status.outputs.default_route_table_id` |
| `spec.vpnLinks[].vpnSiteLinkId` | AzureVpnSite | `status.outputs.link_ids` |
| `spec.vpnLinks[].egressNatRuleIds` | AzureVpnGateway | `status.outputs.nat_rule_ids` |
| `spec.vpnLinks[].ingressNatRuleIds` | AzureVpnGateway | `status.outputs.nat_rule_ids` |

## See Also

- [Overview](../README.md)

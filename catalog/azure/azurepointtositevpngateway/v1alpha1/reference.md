# AzurePointToSiteVpnGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzurePointToSiteVpnGatewaySpec** defines a point-to-site VPN
gateway -- the managed receiver INSIDE a virtual hub that individual
devices (laptops, phones) dial into from anywhere (ARM allows one
per hub, a slot separate from the hub's site-to-site VPN gateway).
HOW users authenticate lives on the AzureVpnServerConfiguration the
gateway is born pointing at; WHAT addresses connected clients get
comes from this gateway's connection configurations. The
classic-world sibling is an AzureVirtualNetworkGateway carrying a
vpn_client_configuration.

**Capacity is scale units**: each scale unit buys 500 connection
slots and aggregate throughput across the managed instance pair.
The gateway bills from creation and provisions in tens of minutes;
plan lifecycle around both.

**ForceNew fields**: `name`, `region`, `resource_group`,
`virtual_hub_id`, `vpn_server_configuration_id`, and
`routing_preference_internet_enabled` -- changing any of them
replaces the gateway (a 30-45 minute create plus a delete, and
every connected client drops). `scale_unit`,
`connection_configurations`, `dns_servers`, and tags update in
place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePointToSiteVpnGateway
metadata:
  name: test-p2s-vpn-gateway
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: remote-users-gw
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
  vpnServerConfigurationId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnServerConfigurations/remote-workforce
  # The unset optional scale_unit applies 1 (the provider requires an
  # explicit value; the module renders the default).
  connectionConfigurations:
    - name: default-clients
      addressPrefixes:
        - "172.16.201.0/24"
      route:
        associatedRouteTableId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus/hubRouteTables/defaultRouteTable
        propagatedRouteTable:
          routeTableIds:
            - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus/hubRouteTables/defaultRouteTable
          labels:
            - default
  dnsServers:
    - "10.0.0.4"
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.virtualHubId` | `string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.virtual_hub_id`) |
| `spec.vpnServerConfigurationId` | `string \| valueFrom` | yes |  | AzureVpnServerConfiguration (`status.outputs.vpn_server_configuration_id`) |
| `spec.connectionConfigurations` | `[]AzurePointToSiteVpnGatewayConnectionConfiguration` | yes |  |  |
| `spec.connectionConfigurations[].name` | `string` | yes |  |  |
| `spec.connectionConfigurations[].addressPrefixes` | `[]string` | yes |  |  |
| `spec.connectionConfigurations[].route` | `AzurePointToSiteVpnGatewayRoute` |  |  |  |
| `spec.connectionConfigurations[].route.associatedRouteTableId` | `string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.connectionConfigurations[].route.inboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.connectionConfigurations[].route.outboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.connectionConfigurations[].route.propagatedRouteTable` | `AzurePointToSiteVpnGatewayPropagatedRouteTable` |  |  |  |
| `spec.connectionConfigurations[].route.propagatedRouteTable.routeTableIds` | `[]string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.connectionConfigurations[].route.propagatedRouteTable.labels` | `[]string` |  |  |  |
| `spec.connectionConfigurations[].internetSecurityEnabled` | `bool` |  |  |  |
| `spec.scaleUnit` | `int32` |  | `1` |  |
| `spec.routingPreferenceInternetEnabled` | `bool` |  |  |  |
| `spec.dnsServers` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the gateway lives in. Must match the hub's
region (the gateway deploys into the hub). Changing it replaces
the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the gateway is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gateway's name, unique within the resource group. Changing
the name replaces the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.virtualHubId

`string | valueFrom` · required

The virtual hub the gateway deploys into -- references an
AzureVirtualHub's ARM ID. One point-to-site gateway per hub
(ARM's rule). Fixed at creation.

- references: AzureVirtualHub (`status.outputs.virtual_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.virtual_hub_id}} -- a bare string does not parse

### spec.vpnServerConfigurationId

`string | valueFrom` · required

The VPN server configuration defining how users authenticate --
references an AzureVpnServerConfiguration's ARM ID. Fixed at
creation (pointing the gateway at a different policy replaces
it); changes WITHIN the referenced configuration apply in place.

- references: AzureVpnServerConfiguration (`status.outputs.vpn_server_configuration_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVpnServerConfiguration, name: <that resource's name>, fieldPath: status.outputs.vpn_server_configuration_id}} -- a bare string does not parse

### spec.connectionConfigurations

`[]AzurePointToSiteVpnGatewayConnectionConfiguration` · required

The gateway's connection configurations -- each names a client
address pool (and optional hub-routing choices) for connected
devices. Most gateways carry exactly one; multiple
configurations require the server configuration to offer OpenVPN
and are matched to users via its policy groups. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.connectionConfigurations[].name

`string` · required

The configuration's name, unique on the gateway (e.g.
"default-clients").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.connectionConfigurations[].addressPrefixes

`[]string` · required

The address pool connected clients draw their tunnel addresses
from, in CIDR notation (e.g. "172.16.201.0/24"). Size it for the
expected concurrent connections; it must not overlap the hub or
any connected network. At least one prefix.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.connectionConfigurations[].route

`AzurePointToSiteVpnGatewayRoute`

How client traffic is routed in the hub. Leave unset for ARM's
default behavior: associate with and propagate to the hub's
built-in default route table (any-to-any reachability).

### spec.connectionConfigurations[].route.associatedRouteTableId

`string | valueFrom` · required

The ONE hub route table this configuration's traffic is routed by
-- references the hub's default table (the hub's
default_route_table_id output, the default reference) or a custom
table via the hub's name-keyed map output, e.g. valueFrom
fieldPath "status.outputs.route_table_ids.remote-users".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.connectionConfigurations[].route.inboundRouteMapId

`string | valueFrom`

A route map applied to routes ARRIVING from connected clients --
references the hub's name-keyed route_map_ids output, e.g.
valueFrom fieldPath "status.outputs.route_map_ids.strip-communities".

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.connectionConfigurations[].route.outboundRouteMapId

`string | valueFrom`

A route map applied to routes ADVERTISED to connected clients --
same referencing shape as inbound_route_map_id.

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.connectionConfigurations[].route.propagatedRouteTable

`AzurePointToSiteVpnGatewayPropagatedRouteTable`

The hub route tables client routes are PROPAGATED to (where other
networks learn the client pool's prefixes from). Unset lets the
service default propagation.

### spec.connectionConfigurations[].route.propagatedRouteTable.routeTableIds

`[]string | valueFrom` · required

Propagate to these specific hub route tables -- reference the
hub's default_route_table_id output (the default reference) or
its name-keyed route_table_ids map, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.remote-users". At least one (the
provider's contract on a configured block).

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.connectionConfigurations[].route.propagatedRouteTable.labels

`[]string`

Also propagate to every hub route table carrying these labels
(labels are declared on the hub's route_tables; the built-in
label "default" targets the hub's default table).

### spec.connectionConfigurations[].internetSecurityEnabled

`bool`

Enable "internet security": the hub advertises a default route
(0.0.0.0/0) to connected clients, so their internet-bound
traffic rides the tunnel into Azure (typically into a hub
firewall via routing intent). Off is ARM's default -- clients
keep their local internet egress.

### spec.scaleUnit

`int32` · optional (explicit presence)

The gateway's aggregate capacity in scale units (500 concurrent
connections each, across the managed instance pair). The
provider requires an explicit value; unset applies 1 (the
smallest gateway). Updates in place.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.routingPreferenceInternetEnabled

`bool`

Route internet-bound traffic from clients out via the public
internet close to the gateway (hot-potato) instead of riding
Microsoft's backbone. Off is ARM's default. Fixed at creation.

### spec.dnsServers

`[]string`

Custom DNS servers pushed to connecting clients (IPv4). Leave
empty for Azure-provided resolution. NOTE: once set, clearing
this list does NOT remove the servers from the deployed gateway
(the provider skips empty-list updates) -- replace the gateway to
remove them.

- rule: {"repeated":{"items":{"string":{"ipv4":true}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the gateway, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Validation Rules

- `connection_configuration_names_unique`: connection_configurations names must be unique on the gateway

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePointToSiteVpnGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.point_to_site_vpn_gateway_id` | `string` | The Azure Resource Manager ID of the point-to-site VPN gateway. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/p2sVpnGateways/{name} |
| `status.outputs.point_to_site_vpn_gateway_name` | `string` | The name of the point-to-site VPN gateway. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualHubId` | AzureVirtualHub | `status.outputs.virtual_hub_id` |
| `spec.vpnServerConfigurationId` | AzureVpnServerConfiguration | `status.outputs.vpn_server_configuration_id` |
| `spec.connectionConfigurations[].route.associatedRouteTableId` | AzureVirtualHub | `status.outputs.default_route_table_id` |
| `spec.connectionConfigurations[].route.inboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.connectionConfigurations[].route.outboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.connectionConfigurations[].route.propagatedRouteTable.routeTableIds` | AzureVirtualHub | `status.outputs.default_route_table_id` |

## See Also

- [Overview](../README.md)

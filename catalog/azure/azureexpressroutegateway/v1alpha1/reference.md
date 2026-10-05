# AzureExpressRouteGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureExpressRouteGatewaySpec** defines an ExpressRoute Gateway --
the on-ramp that lets ExpressRoute circuits reach a Virtual WAN hub.
This is the Virtual WAN counterpart of the classic VNet-resident
ExpressRoute gateway: it lives IN a hub (virtual_hub_id is required
by ARM and fixed at creation), and each of its connections joins one
ExpressRoute circuit PEERING to the hub, so branches and spokes across
the WAN reach the private circuit.

**Cost and time**: the gateway bills hourly per scale unit from
creation (every unit added scales the bill linearly), and ARM takes
roughly 30 minutes to provision one. Each scale unit adds ~2 Gbps of
circuit-to-WAN throughput.

**Connections need a PROVISIONED circuit**: an ExpressRoute
connection references a circuit peering, and ARM accepts it only when
the circuit's provider side is provisioned (a live carrier or an
ExpressRoute Direct port behind it).

**ForceNew fields**: `name`, `region`, `resource_group`, and
`virtual_hub_id` are fixed at creation -- changing any of them
replaces the gateway.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteGateway
metadata:
  name: test-express-route-gateway
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hub-er-gateway
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
  scaleUnits: 1
  # No connections in the canonical manifest: a connection requires a
  # provider-PROVISIONED circuit peering (a live carrier behind it).
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
| `spec.scaleUnits` | `int32` | yes |  |  |
| `spec.allowNonVirtualWanTraffic` | `bool` |  |  |  |
| `spec.connections` | `[]AzureExpressRouteGatewayConnection` |  |  |  |
| `spec.connections[].name` | `string` | yes |  |  |
| `spec.connections[].expressRouteCircuitPeeringId` | `string \| valueFrom` | yes |  | AzureExpressRouteCircuitPeering (`status.outputs.express_route_circuit_peering_id`) |
| `spec.connections[].authorizationKey` | `string` (sensitive) |  |  |  |
| `spec.connections[].internetSecurityEnabled` | `bool` |  |  |  |
| `spec.connections[].expressRouteGatewayBypassEnabled` | `bool` |  |  |  |
| `spec.connections[].routing` | `AzureExpressRouteGatewayConnectionRouting` |  |  |  |
| `spec.connections[].routing.associatedRouteTableId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.connections[].routing.inboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.connections[].routing.outboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.connections[].routing.propagatedRouteTable` | `AzureExpressRouteGatewayConnectionPropagatedRouteTable` |  |  |  |
| `spec.connections[].routing.propagatedRouteTable.labels` | `[]string` |  |  |  |
| `spec.connections[].routing.propagatedRouteTable.routeTableIds` | `[]string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.connections[].routingWeight` | `int32` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the gateway lives in -- must match its hub's
region. Changing the region replaces the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the gateway is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output. Changing it replaces the gateway.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gateway's name, unique within the resource group. Changing the
name replaces the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.virtualHubId

`string | valueFrom` · required

The Virtual WAN hub the gateway is deployed into -- references an
AzureVirtualHub's ARM ID. A hub holds at most one ExpressRoute
gateway. Fixed at creation.

- references: AzureVirtualHub (`status.outputs.virtual_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.virtual_hub_id}} -- a bare string does not parse

### spec.scaleUnits

`int32` · required

The gateway's MINIMUM scale units (1-10). Each unit carries
~2 Gbps of circuit-to-WAN throughput and bills hourly; ARM
auto-scales above this floor under load. Updatable in place.

- rule: {"required":true,"int32":{"lte":10,"gte":1}}

### spec.allowNonVirtualWanTraffic

`bool`

Allow traffic from networks OUTSIDE the Virtual WAN (classic
VNets connected to the same circuit) to flow through this gateway.
Off by default (ARM's default).

### spec.connections

`[]AzureExpressRouteGatewayConnection`

The gateway's connections -- each joins one ExpressRoute circuit
PEERING to the hub. Each connection's ARM ID surfaces in the
connection_ids output map under its name.

### spec.connections[].name

`string` · required

The connection's name, unique on the gateway: 1-80 characters,
starting with a letter or number, containing only letters, numbers,
underscores, periods, and hyphens, and ending in a letter, number,
or underscore (ARM's rule). The ID's lookup name in the
connection_ids output map.

- rule: {"required":true,"string":{"pattern":"^[0-9a-zA-Z]([0-9a-zA-Z_.-]{0,78}[0-9a-zA-Z_])?$"}}

### spec.connections[].expressRouteCircuitPeeringId

`string | valueFrom` · required

The circuit's PRIVATE PEERING being connected -- references an
AzureExpressRouteCircuitPeering's ARM ID. ARM accepts the
connection only when the circuit's provider side is provisioned.
Fixed at creation.

- references: AzureExpressRouteCircuitPeering (`status.outputs.express_route_circuit_peering_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureExpressRouteCircuitPeering, name: <that resource's name>, fieldPath: status.outputs.express_route_circuit_peering_id}} -- a bare string does not parse

### spec.connections[].authorizationKey

`string` · sensitive

The authorization key for connecting to a circuit in ANOTHER
subscription -- redeem one generated by the circuit owner's
authorization. Leave empty when the circuit is in this
subscription. The key is a UUID (8-4-4-4-12 hex) -- taught here
rather than enforced by a validation rule, because sensitive fields
hold a managed-secret reference on consuming platforms and a
value-shape rule would reject every reference (the sibling
connection kinds' authorization/shared keys follow the same
contract).

### spec.connections[].internetSecurityEnabled

`bool`

Enable "internet security": the hub advertises a default route
(0.0.0.0/0) over this connection toward the circuit. Off by
default (ARM's default).

### spec.connections[].expressRouteGatewayBypassEnabled

`bool`

Bypass the gateway for on-premises-to-Private-Link traffic
(FastPath). Off by default (ARM's default); requires a circuit SKU
that supports FastPath.

### spec.connections[].routing

`AzureExpressRouteGatewayConnectionRouting`

The connection's routing configuration. Leave unset for ARM's
default behavior: associate with and propagate to the hub's
built-in default route table.

- rule: Give the routing block something to do: an associated route table or a propagated_route_table -- or omit routing entirely for ARM's defaults

### spec.connections[].routing.associatedRouteTableId

`string | valueFrom`

The ONE hub route table this connection's traffic is routed by --
references the hub's default table (the hub's
default_route_table_id output, the default reference) or a custom
table via the hub's name-keyed map output, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.on-prem".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.connections[].routing.inboundRouteMapId

`string | valueFrom`

A route map applied to routes ARRIVING from the circuit --
references the hub's name-keyed route_map_ids output, e.g.
valueFrom fieldPath "status.outputs.route_map_ids.strip-communities".

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.connections[].routing.outboundRouteMapId

`string | valueFrom`

A route map applied to routes ADVERTISED to the circuit -- same
referencing shape as inbound_route_map_id.

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.connections[].routing.propagatedRouteTable

`AzureExpressRouteGatewayConnectionPropagatedRouteTable`

The hub route tables this connection's routes are PROPAGATED to.
Unset propagates to the hub's default table.

- rule: Give propagation a target: labels, route_table_ids, or both

### spec.connections[].routing.propagatedRouteTable.labels

`[]string`

Propagate to every hub route table carrying these labels.

### spec.connections[].routing.propagatedRouteTable.routeTableIds

`[]string | valueFrom`

Propagate to these specific hub route tables -- reference the hub's
default_route_table_id output (the default reference) or its
name-keyed route_table_ids map, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.on-prem".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.connections[].routingWeight

`int32`

The connection's routing weight (0-32000): when the same prefix is
reachable over multiple connections, higher weight wins.
Unspecified applies 0 (ARM's default).

- rule: {"int32":{"lte":32000,"gte":0}}

### spec.tags

`map<string, string>`

Free-form tags applied to the gateway, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `connection_names_unique`: Connection names must be unique on the gateway

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureExpressRouteGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.express_route_gateway_id` | `string` | The Azure Resource Manager ID of the gateway. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/expressRouteGateways/{name} |
| `status.outputs.express_route_gateway_name` | `string` | The name of the gateway. |
| `status.outputs.connection_ids` | `map<string, string>` | The ARM ID of each connection on the gateway, keyed by the connection's name from the spec. Example valueFrom fieldPath: status.outputs.connection_ids.dc-primary |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualHubId` | AzureVirtualHub | `status.outputs.virtual_hub_id` |
| `spec.connections[].expressRouteCircuitPeeringId` | AzureExpressRouteCircuitPeering | `status.outputs.express_route_circuit_peering_id` |
| `spec.connections[].routing.associatedRouteTableId` | AzureVirtualHub | `status.outputs.default_route_table_id` |
| `spec.connections[].routing.inboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.connections[].routing.outboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.connections[].routing.propagatedRouteTable.routeTableIds` | AzureVirtualHub | `status.outputs.default_route_table_id` |

## See Also

- [Overview](../README.md)

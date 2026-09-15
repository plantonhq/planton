# AzureVirtualHubConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureVirtualHubConnectionSpec** defines a Virtual Hub Connection --
the attachment that joins ONE virtual network (a spoke) to a Virtual
WAN hub. Once connected, the spoke reaches everything the hub's
routing lets it reach: other spokes, branches behind the hub's
gateways, and other hubs' networks through the WAN mesh.

**Routing is where topologies are built**: by default a connection
associates with the hub's built-in route table and propagates to it
(any-to-any). Spoke isolation, shared-services patterns, and
firewall-steered egress are all expressed through the routing block --
associate with a custom table, propagate to labels, inject static
routes toward an appliance in the spoke.

**ForceNew fields**: `name`, `virtual_hub_id`, and
`remote_virtual_network_id` are fixed at creation (the routing block
updates in place; its static_vnet_local_route_override_criteria is the
one ARM-enforced exception -- fixed once set).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualHubConnection
metadata:
  name: test-virtual-hub-connection
spec:
  name: spoke-app
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
  remoteVirtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/spoke-app
  # Unset routing applies ARM's default: associate with and propagate
  # to the hub's built-in default route table (any-to-any). The block
  # below exercises the propagation shape instead.
  routing:
    propagatedRouteTable:
      labels:
        - default
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.virtualHubId` | `string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.virtual_hub_id`) |
| `spec.remoteVirtualNetworkId` | `string \| valueFrom` | yes |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.internetSecurityEnabled` | `bool` |  |  |  |
| `spec.routing` | `AzureVirtualHubConnectionRouting` |  |  |  |
| `spec.routing.associatedRouteTableId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.routing.inboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.routing.outboundRouteMapId` | `string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.route_map_ids`) |
| `spec.routing.propagatedRouteTable` | `AzureVirtualHubConnectionPropagatedRouteTable` |  |  |  |
| `spec.routing.propagatedRouteTable.labels` | `[]string` |  |  |  |
| `spec.routing.propagatedRouteTable.routeTableIds` | `[]string \| valueFrom` |  |  | AzureVirtualHub (`status.outputs.default_route_table_id`) |
| `spec.routing.staticVnetRoutes` | `[]AzureVirtualHubConnectionStaticVnetRoute` |  |  |  |
| `spec.routing.staticVnetRoutes[].name` | `string` | yes |  |  |
| `spec.routing.staticVnetRoutes[].addressPrefixes` | `[]string` | yes |  |  |
| `spec.routing.staticVnetRoutes[].nextHopIpAddress` | `string` | yes |  |  |
| `spec.routing.staticVnetLocalRouteOverrideCriteria` | `enum` |  | `CONTAINS` |  |
| `spec.routing.staticVnetPropagateStaticRoutesEnabled` | `bool` |  | `true` |  |

## Field Details

### spec.name

`string` · required

The connection's name, unique on the hub: 2-80 characters, starting
with a letter or number, containing only letters, numbers,
underscores, periods, and hyphens, and ending in a letter, number,
or underscore. Mirrors the provider's own validation regex exactly
(which enforces a 2-character minimum, despite its error text
saying 1). Changing the name replaces the connection.

- rule: {"required":true,"string":{"pattern":"^[0-9a-zA-Z][-_.0-9a-zA-Z]{0,78}[_0-9a-zA-Z]$"}}

### spec.virtualHubId

`string | valueFrom` · required

The Virtual WAN hub the network attaches to -- references an
AzureVirtualHub's ARM ID. Fixed at creation.

- references: AzureVirtualHub (`status.outputs.virtual_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.virtual_hub_id}} -- a bare string does not parse

### spec.remoteVirtualNetworkId

`string | valueFrom` · required

The virtual network being attached (the spoke) -- references an
AzureVirtualNetwork's ARM ID. The VNet's address space must not
overlap the hub's or any other connected network's. Fixed at
creation. Containment-exempt: the connection is an ARM child of the
hub and merely attaches the spoke, so on a diagram it stands in the
hub with a line to the spoke network, never inside the network.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.internetSecurityEnabled

`bool`

Enable "internet security": the hub advertises a default route
(0.0.0.0/0) to this connection, so the spoke's internet-bound
traffic flows through the hub (typically into a hub firewall via
routing intent). Off by default (ARM's default) -- the spoke keeps
its own internet egress.

### spec.routing

`AzureVirtualHubConnectionRouting`

The connection's routing configuration. Leave unset for ARM's
default behavior: associate with and propagate to the hub's
built-in default route table (any-to-any reachability).

- rule: Give the routing block something to do: an associated route table, a propagated_route_table, or static_vnet_routes -- or omit routing entirely for ARM's defaults

### spec.routing.associatedRouteTableId

`string | valueFrom`

The ONE hub route table this connection's traffic is routed by --
references the hub's default table (the hub's
default_route_table_id output, the default reference) or a custom
table via the hub's name-keyed map output, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.prod-isolated".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.routing.inboundRouteMapId

`string | valueFrom`

A route map applied to routes ARRIVING from the spoke -- references
the hub's name-keyed route_map_ids output, e.g. valueFrom fieldPath
"status.outputs.route_map_ids.strip-communities".

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.routing.outboundRouteMapId

`string | valueFrom`

A route map applied to routes ADVERTISED to the spoke -- same
referencing shape as inbound_route_map_id.

- references: AzureVirtualHub (`status.outputs.route_map_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.route_map_ids}} -- a bare string does not parse

### spec.routing.propagatedRouteTable

`AzureVirtualHubConnectionPropagatedRouteTable`

The hub route tables this connection's routes are PROPAGATED to
(where other networks learn the spoke's prefixes from). Unset
propagates to the hub's default table.

- rule: Give propagation a target: labels, route_table_ids, or both

### spec.routing.propagatedRouteTable.labels

`[]string`

Propagate to every hub route table carrying these labels (labels
are declared on the hub's route_tables; the built-in label
"default" targets the hub's default table).

### spec.routing.propagatedRouteTable.routeTableIds

`[]string | valueFrom`

Propagate to these specific hub route tables -- reference the hub's
default_route_table_id output (the default reference) or its
name-keyed route_table_ids map, e.g. valueFrom fieldPath
"status.outputs.route_table_ids.shared-services".

- references: AzureVirtualHub (`status.outputs.default_route_table_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.default_route_table_id}} -- a bare string does not parse

### spec.routing.staticVnetRoutes

`[]AzureVirtualHubConnectionStaticVnetRoute`

Static routes injected into the association toward the spoke --
the service-chaining shape: steer prefixes at an NVA's IP inside
this spoke so other networks reach services through it.

### spec.routing.staticVnetRoutes[].name

`string` · required

The route's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routing.staticVnetRoutes[].addressPrefixes

`[]string` · required

The destination prefixes this route matches, in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$"}}}}

### spec.routing.staticVnetRoutes[].nextHopIpAddress

`string` · required

The IPv4 address inside the spoke traffic is forwarded to
(typically an NVA's address).

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}$"}}

### spec.routing.staticVnetLocalRouteOverrideCriteria

`enum` · optional (explicit presence)

How a static route interacts with the spoke's own address space:
CONTAINS (ARM's default -- the static route wins only when its
prefix is contained in the VNet's space) or EQUAL (wins only on an
exact prefix match). ARM-enforced: fixed once the connection is
created.

- default: `CONTAINS`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_connection_static_vnet_local_route_override_criteria_unspecified` -- Not specified -- CONTAINS (ARM's default) applies.
- `CONTAINS` -- The static route overrides local routing when its prefix is CONTAINED in the VNet's address space (ARM's default).
- `EQUAL` -- The static route overrides local routing only on an EXACT prefix match.

### spec.routing.staticVnetPropagateStaticRoutesEnabled

`bool` · optional (explicit presence)

Propagate the connection's static routes to the route tables the
connection propagates to. Unspecified applies true (ARM's
default); set false to keep static routes local to the
association.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualHubConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_hub_connection_id` | `string` | The Azure Resource Manager ID of the connection -- what a hub BGP peering references as its virtual_network_connection_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualHubs/{hub}/hubVirtualNetworkConnections/{name} |
| `status.outputs.virtual_hub_connection_name` | `string` | The name of the connection. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.virtualHubId` | AzureVirtualHub | `status.outputs.virtual_hub_id` |
| `spec.remoteVirtualNetworkId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.routing.associatedRouteTableId` | AzureVirtualHub | `status.outputs.default_route_table_id` |
| `spec.routing.inboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.routing.outboundRouteMapId` | AzureVirtualHub | `status.outputs.route_map_ids` |
| `spec.routing.propagatedRouteTable.routeTableIds` | AzureVirtualHub | `status.outputs.default_route_table_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualHub | `spec.bgpConnections[].virtualNetworkConnectionId` | `status.outputs.virtual_hub_connection_id` |

## See Also

- [Overview](../README.md)

# AzureVirtualHub

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVirtualHubSpec** defines a Virtual Hub -- the managed regional
router at the center of Azure's Virtual WAN. Every network that joins
the WAN in a region attaches to that region's hub: spoke VNets through
hub connections, branches through VPN/ExpressRoute gateways created on
the hub, and other hubs through the WAN's automatic hub-to-hub mesh.
The hub runs a Microsoft-managed router (a virtual router with its own
ASN, always 65515) that handles all transit routing.

**This kind models the Virtual WAN hub** -- `virtual_wan_id` is
required. Azure can also create a "standalone" hub (no WAN), which is
the legacy construction of Azure Route Server; that product has its
own dedicated ARM surface and is deliberately not modeled here.

**Routing customization lives on the hub**: custom route tables (spoke
isolation, service-chain steering), route maps (BGP route
transformation on ingress/egress), BGP peerings with NVAs running in
connected spokes, and routing intent (send Internet/private traffic
through a hub firewall). Connections and gateways reference the hub's
route tables and route maps by ID -- this component surfaces them as
name-keyed output maps for exactly that wiring.

**Cost and time**: a Standard hub bills hourly from
creation, and ARM takes 15-30 minutes to bring the hub's router to a
Provisioned routing state. Deleting a hub requires its connections and
gateways to be gone first.

**ForceNew fields**: `name`, `region`, `resource_group`,
`virtual_wan_id`, `address_prefix`, and `sku` are fixed at creation --
changing any of them replaces the hub (and everything attached to it).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualHub
metadata:
  name: test-virtual-hub
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hub-eastus
  virtualWanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualWans/global-wan
  addressPrefix: "10.100.0.0/23"
  # The unset optional fields apply ARM's defaults: Standard tier,
  # ExpressRoute routing preference, branch-to-branch off, router
  # capacity floor 2.
  routeTables:
    - name: prod-isolated
      labels:
        - prod
      routes:
        - name: to-firewall
          destinationsType: CIDR
          destinations:
            - "0.0.0.0/0"
          nextHop:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/azureFirewalls/hub-fw
  routeMaps:
    - name: ingress-policy
      rules:
        - name: tag-on-prem
          matchCriteria:
            - matchCondition: CONTAINS
              routePrefix:
                - "10.0.0.0/8"
          actions:
            - type: ADD
              parameters:
                - community:
                    - "65001:100"
          nextStepIfMatched: CONTINUE
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.virtualWanId` | `string \| valueFrom` | yes |  | AzureVirtualWan (`status.outputs.virtual_wan_id`) |
| `spec.addressPrefix` | `string` | yes |  |  |
| `spec.sku` | `enum` |  | `STANDARD` |  |
| `spec.hubRoutingPreference` | `enum` |  | `EXPRESS_ROUTE` |  |
| `spec.branchToBranchTrafficEnabled` | `bool` |  |  |  |
| `spec.virtualRouterAutoScaleMinCapacity` | `int32` |  | `2` |  |
| `spec.routes` | `[]AzureVirtualHubRoute` |  |  |  |
| `spec.routes[].addressPrefixes` | `[]string` | yes |  |  |
| `spec.routes[].nextHopIpAddress` | `string` | yes |  |  |
| `spec.routeTables` | `[]AzureVirtualHubRouteTable` |  |  |  |
| `spec.routeTables[].name` | `string` | yes |  |  |
| `spec.routeTables[].labels` | `[]string` |  |  |  |
| `spec.routeTables[].routes` | `[]AzureVirtualHubRouteTableRoute` |  |  |  |
| `spec.routeTables[].routes[].name` | `string` | yes |  |  |
| `spec.routeTables[].routes[].destinationsType` | `enum` |  |  |  |
| `spec.routeTables[].routes[].destinations` | `[]string` | yes |  |  |
| `spec.routeTables[].routes[].nextHop` | `string \| valueFrom` | yes |  | AzureFirewall (`status.outputs.firewall_id`) |
| `spec.routeMaps` | `[]AzureVirtualHubRouteMap` |  |  |  |
| `spec.routeMaps[].name` | `string` | yes |  |  |
| `spec.routeMaps[].rules` | `[]AzureVirtualHubRouteMapRule` |  |  |  |
| `spec.routeMaps[].rules[].name` | `string` | yes |  |  |
| `spec.routeMaps[].rules[].matchCriteria` | `[]AzureVirtualHubRouteMapMatchCriterion` |  |  |  |
| `spec.routeMaps[].rules[].matchCriteria[].matchCondition` | `enum` |  |  |  |
| `spec.routeMaps[].rules[].matchCriteria[].asPath` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].matchCriteria[].community` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].matchCriteria[].routePrefix` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].actions` | `[]AzureVirtualHubRouteMapAction` |  |  |  |
| `spec.routeMaps[].rules[].actions[].type` | `enum` |  |  |  |
| `spec.routeMaps[].rules[].actions[].parameters` | `[]AzureVirtualHubRouteMapActionParameter` |  |  |  |
| `spec.routeMaps[].rules[].actions[].parameters[].asPath` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].actions[].parameters[].community` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].actions[].parameters[].routePrefix` | `[]string` |  |  |  |
| `spec.routeMaps[].rules[].nextStepIfMatched` | `enum` |  |  |  |
| `spec.bgpConnections` | `[]AzureVirtualHubBgpConnection` |  |  |  |
| `spec.bgpConnections[].name` | `string` | yes |  |  |
| `spec.bgpConnections[].peerAsn` | `int64` |  |  |  |
| `spec.bgpConnections[].peerIp` | `string` | yes |  |  |
| `spec.bgpConnections[].virtualNetworkConnectionId` | `string \| valueFrom` |  |  | AzureVirtualHubConnection (`status.outputs.virtual_hub_connection_id`) |
| `spec.routingIntent` | `AzureVirtualHubRoutingIntent` |  |  |  |
| `spec.routingIntent.name` | `string` | yes |  |  |
| `spec.routingIntent.routingPolicies` | `[]AzureVirtualHubRoutingPolicy` | yes |  |  |
| `spec.routingIntent.routingPolicies[].name` | `string` | yes |  |  |
| `spec.routingIntent.routingPolicies[].destinations` | `[]enum` | yes |  |  |
| `spec.routingIntent.routingPolicies[].nextHop` | `string \| valueFrom` | yes |  | AzureFirewall (`status.outputs.firewall_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the hub routes traffic in, e.g. "eastus". A WAN
has at most one hub per region. Changing the region replaces the
hub.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the hub is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the hub.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The hub's name, unique within the resource group (1-256 characters,
ARM's rule). Connections, gateways, and route tables address the hub
by its ARM ID (the virtual_hub_id output). Changing the name
replaces the hub.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.virtualWanId

`string | valueFrom` · required

The Virtual WAN this hub belongs to -- references an
AzureVirtualWan's ARM ID. Required: a hub is a regional router OF a
WAN (standalone hubs are the legacy Route Server construction,
deliberately not modeled here). Fixed at creation.

- references: AzureVirtualWan (`status.outputs.virtual_wan_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualWan, name: <that resource's name>, fieldPath: status.outputs.virtual_wan_id}} -- a bare string does not parse

### spec.addressPrefix

`string` · required

The hub's private address space in CIDR notation (e.g.
"10.100.0.0/23"). The hub's router, gateways, and firewall draw
their addresses from this range, so it must not overlap any
connected VNet or branch. Microsoft requires at least /24 and
recommends /23. Fixed at creation. Required here although the
provider marks it optional: the provider's schema also serves the
standalone Route Server construction (not modeled by this kind),
and ARM rejects a WAN hub without an address prefix.

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$"}}

### spec.sku

`enum` · optional (explicit presence)

The hub tier. Unspecified applies STANDARD (ARM's default and the
full-mesh tier: ExpressRoute, S2S/P2S VPN, hub-to-hub transit,
firewall integration). BASIC is the constrained legacy tier
(site-to-site VPN only, inside a Basic WAN). Fixed at creation.

- default: `STANDARD`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_sku_unspecified` -- Not specified -- STANDARD (ARM's default) applies.
- `BASIC` -- The constrained legacy tier: site-to-site VPN only, inside a Basic WAN. ARM upgrades Basic-to-Standard in place but never downgrades.
- `STANDARD` -- The full-mesh tier: ExpressRoute, S2S/P2S VPN, hub-to-hub transit, firewall integration. The right choice for almost everyone.

### spec.hubRoutingPreference

`enum` · optional (explicit presence)

Which path the hub router prefers when the same prefix is learned
over both ExpressRoute and VPN. Unspecified applies EXPRESS_ROUTE
(ARM's default). AS_PATH picks the shortest BGP AS path regardless
of the tunnel type.

- default: `EXPRESS_ROUTE`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_routing_preference_unspecified` -- Not specified -- EXPRESS_ROUTE (ARM's default) applies.
- `EXPRESS_ROUTE` -- Prefer routes learned over ExpressRoute (ARM's default).
- `VPN_GATEWAY` -- Prefer routes learned over VPN tunnels.
- `AS_PATH` -- Prefer the shortest BGP AS path, regardless of tunnel type.

### spec.branchToBranchTrafficEnabled

`bool`

Allow branches (VPN sites) connected to this hub's gateways to
reach branches on other hubs through the WAN. Off by default (ARM's
default); the WAN's own allow_branch_to_branch_traffic must also be
enabled for branch-to-branch transit to flow.

### spec.virtualRouterAutoScaleMinCapacity

`int32` · optional (explicit presence)

The minimum capacity (routing infrastructure units) the hub's
virtual router auto-scales from. Unspecified applies 2 (ARM's
default and floor: 2 units ≈ 3 Gbps aggregate, supporting ~2,000
attached VMs); ARM scales up automatically and bills per unit.

- default: `2`
- rule: {"int32":{"gte":2}}

### spec.routes

`[]AzureVirtualHubRoute`

Static routes on the hub's DEFAULT route table, in the hub
resource's classic inline form (address prefixes forwarded to a
next-hop IP). For the modern per-table form -- named routes over
named next-hop RESOURCES -- use route_tables instead; this inline
set exists for parity with the provider's hub-level route block.

### spec.routes[].addressPrefixes

`[]string` · required

The destination prefixes this route matches, in CIDR notation.

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$"}}}}

### spec.routes[].nextHopIpAddress

`string` · required

The IPv4 address traffic matching the prefixes is forwarded to.

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}$"}}

### spec.routeTables

`[]AzureVirtualHubRouteTable`

Custom hub route tables. Connections associate with ONE route table
and propagate to MANY (that is how spoke isolation and shared-
services topologies are built); each table's ARM ID surfaces in the
route_table_ids output map under its name. The hub always has a
built-in defaultRouteTable (its ID is the default_route_table_id
output) -- these entries are ADDITIONAL tables.

- rule: Route names must be unique within the route table

### spec.routeTables[].name

`string` · required

The table's name, unique on the hub -- the ID's lookup name in the
route_table_ids output map. ARM forbids the characters <>%&:?/+ in
the name.

- rule: {"required":true,"string":{"pattern":"^[^<>%&:?/+]+$"}}

### spec.routeTables[].labels

`[]string`

Labels grouping this table for bulk propagation: a connection that
propagates to a label propagates to every table carrying it
(e.g. label all production-spoke tables "prod" and propagate once).

### spec.routeTables[].routes

`[]AzureVirtualHubRouteTableRoute`

The table's static routes. Routes steer matched destinations to a
next-hop RESOURCE (an Azure Firewall or a hub connection) -- this is
how service chaining and spoke-egress-through-firewall are built.

- rule: Choose how destinations are interpreted: CIDR (prefixes, the common choice), RESOURCE_ID, or SERVICE

### spec.routeTables[].routes[].name

`string` · required

The route's name, unique within its table.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routeTables[].routes[].destinationsType

`enum`

How the destinations list is interpreted: CIDR prefixes, ARM
resource IDs, or service names. CIDR is the common choice.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_route_destinations_type_unspecified` -- Not specified -- invalid: the interpretation is explicit.
- `CIDR` -- Destinations are CIDR prefixes (the common choice).
- `RESOURCE_ID` -- Destinations are ARM resource IDs.
- `SERVICE` -- Destinations are Azure service names.

### spec.routeTables[].routes[].destinations

`[]string` · required

The destinations this route matches, interpreted per
destinations_type.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.routeTables[].routes[].nextHop

`string | valueFrom` · required

The next-hop RESOURCE traffic is steered to -- the ARM ID of an
Azure Firewall in this hub (the classic service-chaining shape) or
of a hub connection. References an AzureFirewall's ARM id by
default. (ARM's next_hop_type has exactly one value, "ResourceId" --
the modules emit it; there is nothing to configure.)

- references: AzureFirewall (`status.outputs.firewall_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFirewall, name: <that resource's name>, fieldPath: status.outputs.firewall_id}} -- a bare string does not parse

### spec.routeMaps

`[]AzureVirtualHubRouteMap`

Route maps -- ordered rule lists that match BGP routes (by prefix,
AS path, or community) and transform or drop them as they enter or
leave hub connections. Each map's ARM ID surfaces in the
route_map_ids output map under its name; connections reference maps
as inbound_route_map_id / outbound_route_map_id.

- rule: Rule names must be unique within the route map

### spec.routeMaps[].name

`string` · required

The map's name, unique on the hub -- the ID's lookup name in the
route_map_ids output map. 1-80 characters, starting alphanumeric,
containing only letters, numbers, underscores, periods, and
hyphens, ending in a letter, number, or underscore (ARM's rule).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,78}[a-zA-Z_0-9])?$"}}

### spec.routeMaps[].rules

`[]AzureVirtualHubRouteMapRule`

The map's rules, evaluated in order.

### spec.routeMaps[].rules[].name

`string` · required

The rule's name, unique within its map.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routeMaps[].rules[].matchCriteria

`[]AzureVirtualHubRouteMapMatchCriterion`

The criteria a route must match for this rule to apply. A rule
with no criteria matches every route.

- rule: Choose how the criterion compares: CONTAINS, EQUALS, NOT_CONTAINS, or NOT_EQUALS
- rule: Give the criterion something to match: as_path, community, or route_prefix

### spec.routeMaps[].rules[].matchCriteria[].matchCondition

`enum`

How the criterion's lists compare against the route's attributes.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_route_map_match_condition_unspecified` -- Not specified -- invalid: the comparison is explicit.
- `CONTAINS` -- The route's attribute contains the listed values.
- `EQUALS` -- The route's attribute equals the listed values exactly.
- `NOT_CONTAINS` -- The route's attribute does not contain the listed values.
- `NOT_EQUALS` -- The route's attribute does not equal the listed values.

### spec.routeMaps[].rules[].matchCriteria[].asPath

`[]string`

Match against the route's BGP AS path (AS numbers as strings).

### spec.routeMaps[].rules[].matchCriteria[].community

`[]string`

Match against the route's BGP community values (e.g. "65001:100").

### spec.routeMaps[].rules[].matchCriteria[].routePrefix

`[]string`

Match against the route's prefix (CIDR notation).

### spec.routeMaps[].rules[].actions

`[]AzureVirtualHubRouteMapAction`

The actions applied to matching routes, in order.

- rule: Choose the action: ADD, REMOVE, or REPLACE (transform attributes), or DROP (discard the route)
- rule: ADD, REMOVE, and REPLACE actions need at least one parameter (the attribute values to transform); only DROP takes none

### spec.routeMaps[].rules[].actions[].type

`enum`

What the action does to the matching route: ADD/REMOVE/REPLACE
transform the attributes named in parameters; DROP discards the
route (and takes no parameters).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_route_map_action_type_unspecified` -- Not specified -- invalid: the action is explicit.
- `ADD` -- Add the parameter values to the route's attributes.
- `DROP` -- Discard the route. Takes no parameters.
- `REMOVE` -- Remove the parameter values from the route's attributes.
- `REPLACE` -- Replace the route's attributes with the parameter values.

### spec.routeMaps[].rules[].actions[].parameters

`[]AzureVirtualHubRouteMapActionParameter`

The attribute values the action adds, removes, or replaces. ARM
requires at least one parameter on every action except DROP.

### spec.routeMaps[].rules[].actions[].parameters[].asPath

`[]string`

BGP AS path values (AS numbers as strings).

### spec.routeMaps[].rules[].actions[].parameters[].community

`[]string`

BGP community values (e.g. "65001:100").

### spec.routeMaps[].rules[].actions[].parameters[].routePrefix

`[]string`

Route prefixes (CIDR notation).

### spec.routeMaps[].rules[].nextStepIfMatched

`enum` · optional (explicit presence)

What happens after a route matches this rule. Unspecified leaves
the decision to ARM (its "Unknown" value): evaluation stops.
CONTINUE evaluates subsequent rules too; TERMINATE stops here
explicitly.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_route_map_next_step_unspecified` -- Not specified -- ARM's "Unknown" applies: evaluation stops after the match.
- `CONTINUE` -- Continue evaluating subsequent rules against the (transformed) route.
- `TERMINATE` -- Stop evaluating -- this rule's outcome is final.

### spec.bgpConnections

`[]AzureVirtualHubBgpConnection`

BGP peerings between the hub's router and network virtual
appliances (NVAs) running in CONNECTED spoke VNets -- how an NVA
exchanges dynamic routes with the hub. The peer must be reachable
through a hub connection (set virtual_network_connection_id to the
connection carrying the NVA's VNet).

### spec.bgpConnections[].name

`string` · required

The peering's name, unique on the hub -- the ID's lookup name in
the bgp_connection_ids output map.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.bgpConnections[].peerAsn

`int64`

The NVA's BGP autonomous system number. Must differ from the hub
router's own ASN (always 65515).

- rule: {"int64":{"gte":"0"}}

### spec.bgpConnections[].peerIp

`string` · required

The NVA's IPv4 address inside the connected spoke VNet.

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}$"}}

### spec.bgpConnections[].virtualNetworkConnectionId

`string | valueFrom`

The hub connection carrying the spoke VNet the NVA lives in --
references an AzureVirtualHubConnection's ARM id. ARM accepts a
peering without it, but routes only flow once the peer is reachable
through a connection.

- references: AzureVirtualHubConnection (`status.outputs.virtual_hub_connection_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHubConnection, name: <that resource's name>, fieldPath: status.outputs.virtual_hub_connection_id}} -- a bare string does not parse

### spec.routingIntent

`AzureVirtualHubRoutingIntent`

Routing intent: steer categories of traffic (Internet-bound,
private) through a next-hop security appliance -- almost always an
Azure Firewall deployed in this hub. A hub has at most ONE routing
intent; setting it takes over the hub's routing policy, so
per-connection route-table customization and routing intent are
mutually exclusive on ARM's side.

- rule: Routing policy names must be unique within the routing intent

### spec.routingIntent.name

`string` · required

The routing intent's name (ARM convention: "hubRoutingIntent").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingIntent.routingPolicies

`[]AzureVirtualHubRoutingPolicy` · required

The policies -- in practice one per traffic category (an Internet
policy and/or a PrivateTraffic policy).

- rule: {"repeated":{"minItems":"1"}}

### spec.routingIntent.routingPolicies[].name

`string` · required

The policy's name (e.g. "InternetTraffic", "PrivateTrafficPolicy").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingIntent.routingPolicies[].destinations

`[]enum` · required

The traffic categories this policy steers (ARM models a list; the
common shape is exactly one category per policy).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_virtual_hub_routing_policy_destination_unspecified` -- Not specified -- invalid: the category is explicit.
- `INTERNET` -- Internet-bound traffic (0.0.0.0/0).
- `PRIVATE_TRAFFIC` -- Private traffic (RFC1918 ranges and connected network prefixes).

### spec.routingIntent.routingPolicies[].nextHop

`string | valueFrom` · required

The next-hop security appliance -- the ARM ID of an Azure Firewall
deployed in THIS hub (the supported appliance for routing intent;
referencing a firewall in another hub or VNet fails at deploy).

- references: AzureFirewall (`status.outputs.firewall_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFirewall, name: <that resource's name>, fieldPath: status.outputs.firewall_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the hub, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins.

## Validation Rules

- `route_table_names_unique`: Route table names must be unique on the hub
- `route_map_names_unique`: Route map names must be unique on the hub
- `bgp_connection_names_unique`: BGP connection names must be unique on the hub

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualHub, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_hub_id` | `string` | The Azure Resource Manager ID of the hub -- what connections, gateways, and firewalls reference as their virtual_hub_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualHubs/{name} |
| `status.outputs.virtual_hub_name` | `string` | The name of the hub. |
| `status.outputs.default_route_table_id` | `string` | The ARM ID of the hub's built-in default route table -- what a connection's routing associates with when no custom table is chosen. Example valueFrom fieldPath: status.outputs.default_route_table_id |
| `status.outputs.virtual_router_asn` | `int64` | The hub router's BGP autonomous system number (always 65515) -- configure it as the remote ASN when peering NVAs with the hub. |
| `status.outputs.virtual_router_ips` | `[]string` | The hub router's peering IPv4 addresses (a pair) -- the BGP neighbor addresses an NVA in a connected spoke peers with. |
| `status.outputs.route_table_ids` | `map<string, string>` | The ARM ID of each custom route table on the hub, keyed by the table's name from the spec -- what connections reference as their associated/propagated route tables. Example valueFrom fieldPath: status.outputs.route_table_ids.prod-isolated |
| `status.outputs.route_map_ids` | `map<string, string>` | The ARM ID of each route map on the hub, keyed by the map's name from the spec -- what connections reference as their inbound/outbound route maps. Example valueFrom fieldPath: status.outputs.route_map_ids.strip-private-communities |
| `status.outputs.bgp_connection_ids` | `map<string, string>` | The ARM ID of each BGP connection on the hub, keyed by the peering's name from the spec. |
| `status.outputs.routing_intent_id` | `string` | The ARM ID of the hub's routing intent -- empty when no routing intent is configured. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualWanId` | AzureVirtualWan | `status.outputs.virtual_wan_id` |
| `spec.routeTables[].routes[].nextHop` | AzureFirewall | `status.outputs.firewall_id` |
| `spec.bgpConnections[].virtualNetworkConnectionId` | AzureVirtualHubConnection | `status.outputs.virtual_hub_connection_id` |
| `spec.routingIntent.routingPolicies[].nextHop` | AzureFirewall | `status.outputs.firewall_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureExpressRouteGateway | `spec.virtualHubId` | `status.outputs.virtual_hub_id` |
| AzureExpressRouteGateway | `spec.connections[].routing.associatedRouteTableId` | `status.outputs.default_route_table_id` |
| AzureExpressRouteGateway | `spec.connections[].routing.inboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureExpressRouteGateway | `spec.connections[].routing.outboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureExpressRouteGateway | `spec.connections[].routing.propagatedRouteTable.routeTableIds` | `status.outputs.default_route_table_id` |
| AzurePointToSiteVpnGateway | `spec.virtualHubId` | `status.outputs.virtual_hub_id` |
| AzurePointToSiteVpnGateway | `spec.connectionConfigurations[].route.associatedRouteTableId` | `status.outputs.default_route_table_id` |
| AzurePointToSiteVpnGateway | `spec.connectionConfigurations[].route.inboundRouteMapId` | `status.outputs.route_map_ids` |
| AzurePointToSiteVpnGateway | `spec.connectionConfigurations[].route.outboundRouteMapId` | `status.outputs.route_map_ids` |
| AzurePointToSiteVpnGateway | `spec.connectionConfigurations[].route.propagatedRouteTable.routeTableIds` | `status.outputs.default_route_table_id` |
| AzureVirtualHubConnection | `spec.virtualHubId` | `status.outputs.virtual_hub_id` |
| AzureVirtualHubConnection | `spec.routing.associatedRouteTableId` | `status.outputs.default_route_table_id` |
| AzureVirtualHubConnection | `spec.routing.inboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureVirtualHubConnection | `spec.routing.outboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureVirtualHubConnection | `spec.routing.propagatedRouteTable.routeTableIds` | `status.outputs.default_route_table_id` |
| AzureVpnGateway | `spec.virtualHubId` | `status.outputs.virtual_hub_id` |
| AzureVpnGatewayConnection | `spec.routing.associatedRouteTableId` | `status.outputs.default_route_table_id` |
| AzureVpnGatewayConnection | `spec.routing.inboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureVpnGatewayConnection | `spec.routing.outboundRouteMapId` | `status.outputs.route_map_ids` |
| AzureVpnGatewayConnection | `spec.routing.propagatedRouteTable.routeTableIds` | `status.outputs.default_route_table_id` |

## See Also

- [Overview](../README.md)

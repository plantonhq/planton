# AzureExpressRouteCircuit

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureExpressRouteCircuitSpec** defines an ExpressRoute circuit -- the
dedicated PRIVATE connection between your infrastructure (on-premises
or colocation) and Microsoft, bought either through a connectivity
provider or on your own ExpressRoute Direct port. The circuit is the
billing and identity object: creating it issues a service key the
provider uses to provision the physical cross-connect, and routing
only starts once peerings (AzureExpressRouteCircuitPeering) are
configured on it.

**Exactly one provisioning mode**:
- **Service provider** (the common shape): name the provider, its
  peering location, and the bandwidth in Mbps -- the trio travels
  together. After creation the circuit sits in provisioning state
  "NotProvisioned" until you hand the service key to the provider and
  they complete the cross-connect.
- **ExpressRoute Direct**: reference the ExpressRoute Port and set the
  bandwidth in Gbps -- the pair travels together. No third-party
  provider is involved.

**Billing starts at creation**: Azure meters the circuit from the
moment the service key is issued, even while the provider side is
unprovisioned. Deprovision (delete) circuits you are not going to
hook up.

**ForceNew fields**: `name`, `region`, `resource_group`,
`service_provider_name`, `peering_location`, and
`express_route_port_id`. `bandwidth_in_mbps` may be INCREASED in
place; decreasing it replaces the circuit (ARM cannot shrink a
provisioned circuit).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteCircuit
metadata:
  name: test-express-route-circuit
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hq-circuit
  # STANDARD reaches the geopolitical area; METERED_DATA bills outbound
  # per GB (the common production pairing).
  skuTier: STANDARD
  skuFamily: METERED_DATA
  # Service-provider mode: the trio travels together. Billing starts at
  # creation; the circuit sits NotProvisioned until the provider
  # completes the cross-connect with the service key.
  serviceProviderName: "Equinix"
  peeringLocation: "Washington DC"
  bandwidthInMbps: 50
  # One issued authorization: its ARM-generated key (sensitive) surfaces
  # in the authorization_keys output under this name.
  authorizations:
    - name: partner-team
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.skuTier` | `enum` |  |  |  |
| `spec.skuFamily` | `enum` |  |  |  |
| `spec.serviceProviderName` | `string` |  |  |  |
| `spec.peeringLocation` | `string` |  |  |  |
| `spec.bandwidthInMbps` | `int32` |  |  |  |
| `spec.expressRoutePortId` | `string \| valueFrom` |  |  | AzureExpressRoutePort (`status.outputs.express_route_port_id`) |
| `spec.bandwidthInGbps` | `double` |  |  |  |
| `spec.rateLimitingEnabled` | `bool` |  |  |  |
| `spec.allowClassicOperations` | `bool` |  |  |  |
| `spec.authorizationKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.authorizations` | `[]AzureExpressRouteCircuitAuthorization` |  |  |  |
| `spec.authorizations[].name` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the circuit object lives in, e.g. "eastus". This is
the ARM metadata location -- the physical connectivity location is
peering_location. Changing the region replaces the circuit.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the circuit is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the circuit.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The circuit's name, unique within the resource group. Peerings
reference the circuit by this name. Changing the name replaces the
circuit.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.skuTier

`enum`

The SKU tier -- what the circuit can reach. LOCAL peers only within
the circuit's metro (no egress fees), STANDARD reaches every region
in the geopolitical area, PREMIUM adds global reach across
geopolitical areas plus higher route limits. BASIC is a legacy
constrained tier. Together with sku_family this forms Azure's SKU
name (e.g. "Standard_MeteredData").

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_circuit_sku_tier_unspecified` -- Not specified -- invalid: the tier choice is explicit (see the sku_tier_required contract).
- `BASIC` -- Legacy constrained tier (limited routes, no global reach). Prefer LOCAL or STANDARD for anything new.
- `LOCAL` -- Connectivity within the circuit's own metro area only -- no egress fees, priced below STANDARD.
- `STANDARD` -- Connectivity to every Azure region in the circuit's geopolitical area. The common production choice.
- `PREMIUM` -- STANDARD plus global reach across geopolitical areas, more route prefixes, and more VNet links.

### spec.skuFamily

`enum`

The billing family: METERED_DATA bills outbound data per GB;
UNLIMITED_DATA is a flat rate with no egress metering (economical
above roughly two-thirds sustained utilization).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_circuit_sku_family_unspecified` -- Not specified -- invalid: the family choice is explicit (see the sku_family_required contract).
- `METERED_DATA` -- Outbound data billed per GB on top of the circuit fee.
- `UNLIMITED_DATA` -- Flat circuit fee with unlimited data transfer.

### spec.serviceProviderName

`string`

SERVICE-PROVIDER MODE: the connectivity provider's name exactly as
Azure lists it (e.g. "Equinix", "Megaport" -- `az network
express-route list-service-providers` shows the vocabulary).
Requires peering_location and bandwidth_in_mbps. Fixed at creation.

### spec.peeringLocation

`string`

SERVICE-PROVIDER MODE: the provider's peering location -- the
physical site of the cross-connect (e.g. "Washington DC", "Silicon
Valley"), NOT an Azure region name. Fixed at creation.

### spec.bandwidthInMbps

`int32`

SERVICE-PROVIDER MODE: the circuit bandwidth in Mbps, from the
provider's offered steps (50, 100, 200, 500, 1000, 2000, 5000,
10000). May be INCREASED in place; decreasing replaces the circuit.

- rule: {"int32":{"gte":0}}

### spec.expressRoutePortId

`string | valueFrom`

EXPRESSROUTE DIRECT MODE: the ExpressRoute Port the circuit rides
on -- defaults to referencing an AzureExpressRoutePort's
express_route_port_id output; pass the ARM id as a literal for a
port managed outside Planton. Requires bandwidth_in_gbps. Fixed at
creation.

- references: AzureExpressRoutePort (`status.outputs.express_route_port_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureExpressRoutePort, name: <that resource's name>, fieldPath: status.outputs.express_route_port_id}} -- a bare string does not parse

### spec.bandwidthInGbps

`double`

EXPRESSROUTE DIRECT MODE: the circuit bandwidth in Gbps carved from
the port (fractional steps like 1, 2, 5, 10 up to the port size).

- rule: {"double":{"gte":0}}

### spec.rateLimitingEnabled

`bool`

Rate-limit the circuit to its configured bandwidth on ExpressRoute
Direct (maps to ARM's EnableDirectPortRateLimit). Off by default;
meaningful on Direct circuits only.

### spec.allowClassicOperations

`bool`

Allow the classic (ASM) deployment model to use this circuit.
Legacy interop only -- leave off for anything new.

### spec.authorizationKey

`string | valueFrom` · sensitive

The authorization key this circuit REDEEMS when it is built on
capacity someone else owns (e.g. an ExpressRoute Port authorization
from another subscription). Reference a secret rather than embedding
the literal in manifests. ARM never returns it on reads. Not to be
confused with `authorizations` below, which ISSUES keys to others.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.authorizations

`[]AzureExpressRouteCircuitAuthorization`

Authorizations ISSUED by this circuit: each entry creates a named
authorization whose generated key lets a virtual network gateway in
ANOTHER subscription connect to this circuit. The generated keys
surface (marked sensitive) in the circuit's authorization_keys
output, keyed by name. Deleting an entry revokes the authorization.

### spec.authorizations[].name

`string` · required

The authorization's name, unique on the circuit -- the key's lookup
name in the authorization_keys output. Renaming revokes the old
authorization and issues a new key.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tags

`map<string, string>`

Free-form tags applied to the circuit, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `sku_tier_required`: Choose the SKU tier explicitly -- LOCAL (metro only, no egress fees), STANDARD (geopolitical area), or PREMIUM (global reach)
- `sku_family_required`: Choose the billing family explicitly -- METERED_DATA (pay per outbound GB) or UNLIMITED_DATA (flat rate)
- `exactly_one_provisioning_mode`: Provision the circuit exactly one way: through a service provider (service_provider_name + peering_location + bandwidth_in_mbps) or on an ExpressRoute Direct port (express_route_port_id + bandwidth_in_gbps)
- `service_provider_trio_travels_together`: Service-provider mode needs all three of service_provider_name, peering_location, and bandwidth_in_mbps
- `provider_fields_only_in_provider_mode`: peering_location and bandwidth_in_mbps belong to service-provider mode -- remove them from an ExpressRoute Direct circuit
- `direct_pair_travels_together`: ExpressRoute Direct mode needs both express_route_port_id and bandwidth_in_gbps
- `gbps_only_in_direct_mode`: bandwidth_in_gbps belongs to ExpressRoute Direct mode -- service-provider circuits size bandwidth_in_mbps instead
- `authorization_names_unique`: Authorization names must be unique on the circuit

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureExpressRouteCircuit, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.express_route_circuit_id` | `string` | The Azure Resource Manager ID of the circuit. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/expressRouteCircuits/{name} |
| `status.outputs.express_route_circuit_name` | `string` | The name of the circuit -- what peerings (AzureExpressRouteCircuitPeering) reference as express_route_circuit_name. |
| `status.outputs.service_key` | `string` (sensitive) | The SERVICE KEY -- the circuit's provisioning credential, handed to the connectivity provider to complete the physical cross-connect. Marked sensitive in both engines; treat it like a password. |
| `status.outputs.service_provider_provisioning_state` | `string` | The provider side's provisioning state: "NotProvisioned" (fresh circuit, waiting on the provider), "Provisioning", "Provisioned" (peerings can be configured), or "Deprovisioning". |
| `status.outputs.authorization_keys` | `map<string, string>` (sensitive) | The generated key of each authorization ISSUED by this circuit, keyed by the authorization's name from the spec. A virtual network gateway in another subscription redeems one to connect. Marked sensitive in both engines. Example valueFrom fieldPath: status.outputs.authorization_keys.partner-team |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.expressRoutePortId` | AzureExpressRoutePort | `status.outputs.express_route_port_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureExpressRouteCircuitPeering | `spec.expressRouteCircuitName` | `status.outputs.express_route_circuit_name` |
| AzureVirtualNetworkGatewayConnection | `spec.expressRouteCircuitId` | `status.outputs.express_route_circuit_id` |

## See Also

- [Overview](../README.md)

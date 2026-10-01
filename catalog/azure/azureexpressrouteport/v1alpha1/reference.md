# AzureExpressRoutePort

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureExpressRoutePortSpec** defines an ExpressRoute Port -- your own
pair of physical ports on a Microsoft Enterprise Edge router at a
peering location (ExpressRoute Direct). Where an ordinary ExpressRoute
circuit rents capacity through a connectivity provider, a port IS the
capacity: you order 10 or 100 Gbps of dual physical links, receive the
facility facts (router, interface, patch panel, rack) in this
component's outputs, arrange the cross-connects with the facility, and
then carve circuits (AzureExpressRouteCircuit in Direct mode) from the
port's bandwidth.

**Billing and enrollment**: the port bills its full monthly rate from
creation, whether or not any cross-connect exists -- this is one of the
most expensive single objects in Azure networking, so create ports
deliberately. Subscriptions may also need Microsoft enrollment for
ExpressRoute Direct before ARM accepts the create.

**Links come as a fixed pair**: ARM always creates exactly two links
(the physical port pair) -- link1 and link2 here configure the existing
pair (admin state, MACsec); they never add or remove links. Links
start administratively DISABLED; enable them when the facility
completes the cross-connect.

**ForceNew fields**: `name`, `region`, `resource_group`,
`peering_location`, `bandwidth_in_gbps`, and `encapsulation` are all
fixed at creation -- changing any of them replaces the port (and every
circuit riding on it).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRoutePort
metadata:
  name: test-express-route-port
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hq-port
  # The ExpressRoute DIRECT location vocabulary (narrower than circuit
  # peering locations): `az network express-route port location list`.
  peeringLocation: "Equinix-Ashburn-DC2"
  # 10 or 100 at real locations. The port bills its full monthly rate
  # from creation -- cross-connects or not.
  bandwidthInGbps: 10
  # DOT1Q: one VLAN tag per circuit (the common choice). Fixed at
  # creation, together with location and bandwidth.
  encapsulation: DOT1Q
  # Links always exist as a pair; these blocks manipulate them. They
  # start administratively disabled -- enable once the facility
  # completes the cross-connects.
  link1:
    adminEnabled: true
  link2:
    adminEnabled: true
  # One issued authorization: its ARM-generated key (sensitive)
  # surfaces in the authorization_keys output under this name.
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
| `spec.peeringLocation` | `string` | yes |  |  |
| `spec.bandwidthInGbps` | `int32` | yes |  |  |
| `spec.encapsulation` | `enum` |  |  |  |
| `spec.billingType` | `enum` |  | `METERED_DATA` |  |
| `spec.identity` | `AzureExpressRoutePortIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.link1` | `AzureExpressRoutePortLink` |  |  |  |
| `spec.link1.adminEnabled` | `bool` |  |  |  |
| `spec.link1.macsecCipher` | `enum` |  | `GCM_AES_128` |  |
| `spec.link1.macsecCknKeyvaultSecretId` | `string` |  |  |  |
| `spec.link1.macsecCakKeyvaultSecretId` | `string` |  |  |  |
| `spec.link1.macsecSciEnabled` | `bool` |  |  |  |
| `spec.link2` | `AzureExpressRoutePortLink` |  |  |  |
| `spec.link2.adminEnabled` | `bool` |  |  |  |
| `spec.link2.macsecCipher` | `enum` |  | `GCM_AES_128` |  |
| `spec.link2.macsecCknKeyvaultSecretId` | `string` |  |  |  |
| `spec.link2.macsecCakKeyvaultSecretId` | `string` |  |  |  |
| `spec.link2.macsecSciEnabled` | `bool` |  |  |  |
| `spec.authorizations` | `[]AzureExpressRoutePortAuthorization` |  |  |  |
| `spec.authorizations[].name` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the port object lives in, e.g. "eastus". This is
the ARM metadata location -- the physical site is peering_location.
Changing the region replaces the port.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the port is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the port.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The port's name, unique within the resource group: 1-80 characters,
starting with a letter or number, containing only letters, numbers,
underscores, periods, and hyphens, and not ending in a period or
hyphen (ARM's rule, enforced here so a bad name fails in seconds
rather than at deploy). Changing the name replaces the port.

- rule: {"required":true,"string":{"pattern":"^[^\\W_]([\\w.-]{0,78}[\\w])?$"}}

### spec.peeringLocation

`string` · required

The ExpressRoute Direct peering location -- the colocation facility
where the physical port pair is provisioned (e.g.
"Equinix-Ashburn-DC2"). This is the ExpressRoute DIRECT vocabulary
(`az network express-route port location list`), which is narrower
than the provider-circuit peering locations. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.bandwidthInGbps

`int32` · required

The port pair's total bandwidth in Gbps. ARM offers exactly 10 or
100 at real peering locations (the API accepts any positive value,
so the vocabulary is documented rather than validated -- new sizes
appear without a provider release). Circuits carved from the port
may oversubscribe it (up to 2x on aggregate provisioned bandwidth).
Fixed at creation.

- rule: {"required":true,"int32":{"gte":1}}

### spec.encapsulation

`enum`

The Ethernet encapsulation of the physical links. DOT1Q carries one
VLAN tag per circuit (the common choice); QINQ stacks an outer
Azure-managed S-Tag over your C-Tags, letting overlapping customer
VLAN ranges share the port. Fixed at creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_port_encapsulation_unspecified` -- Not specified -- invalid: the encapsulation choice is explicit (see the encapsulation_required contract).
- `DOT1Q` -- One VLAN tag per circuit (802.1Q). The common choice when your VLAN ranges do not overlap.
- `QINQ` -- Stacked tags (802.1ad): Azure manages an outer S-Tag per circuit so overlapping customer VLAN ranges can share the port.

### spec.billingType

`enum` · optional (explicit presence)

How the port is billed. Unspecified applies METERED_DATA (ARM's
default): the port fee plus per-GB outbound charges on its circuits.
UNLIMITED_DATA is a higher flat fee with no egress metering --
economical at sustained high utilization.

- default: `METERED_DATA`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_port_billing_type_unspecified` -- Not specified -- METERED_DATA (ARM's default) applies.
- `METERED_DATA` -- Port fee plus per-GB outbound data charges on the circuits riding it.
- `UNLIMITED_DATA` -- Higher flat port fee with unlimited data transfer on its circuits.

### spec.identity

`AzureExpressRoutePortIdentity`

The port's managed identity. Required for MACsec (the identity reads
the CAK/CKN secrets from Key Vault at provisioning time -- grant it
Key Vault secret GET before enabling MACsec). Leave unset for ports
that need no identity.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the port; USER_ASSIGNED brings identities you manage (what
MACsec requires -- the identity must be grantable Key Vault access
BEFORE the port provisions its links); SYSTEM_AND_USER_ASSIGNED
carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_express_route_port_identity_type_unspecified` -- Not specified: the port has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the port.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids). What MACsec requires.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the port, by ARM ID. Reference
AzureUserAssignedIdentity resources so Key Vault grants (MACsec CAK/
CKN secret GET) can be composed before the port is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.link1

`AzureExpressRoutePortLink`

Configuration of the FIRST physical link of the pair. Links always
exist (ARM creates the pair with the port); this block only
manipulates the existing link. Leave unset to keep the link
administratively disabled with no MACsec.

- rule: MACsec needs both keys or neither -- set macsec_ckn_keyvault_secret_id and macsec_cak_keyvault_secret_id together

### spec.link1.adminEnabled

`bool`

Administratively enable the link. Links start DISABLED; enable them
once the facility completes the physical cross-connect. ARM applies
link configuration in a second call after the port exists, so a
fresh port briefly reports links disabled even when this is true.

### spec.link1.macsecCipher

`enum` · optional (explicit presence)

The MACsec cipher suite used when MACsec is configured. Unspecified
applies GCM_AES_128 (ARM's default). The XPN variants (extended
packet numbering) are for sustained rates above ~40 Gbps where the
32-bit packet number would wrap too quickly.

- default: `GCM_AES_128`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_port_macsec_cipher_unspecified` -- Not specified -- GCM_AES_128 (ARM's default) applies.
- `GCM_AES_128` -- AES-GCM with a 128-bit key.
- `GCM_AES_256` -- AES-GCM with a 256-bit key.
- `GCM_AES_XPN_128` -- AES-GCM 128-bit with extended (64-bit) packet numbering -- for sustained rates where a 32-bit packet number wraps too quickly.
- `GCM_AES_XPN_256` -- AES-GCM 256-bit with extended packet numbering.

### spec.link1.macsecCknKeyvaultSecretId

`string`

The Key Vault SECRET IDENTIFIER (a versioned https URL, not the
secret itself) holding the MACsec CKN -- the connectivity key name,
a hex string shared with the facility side. Travels together with
the CAK; the port's user-assigned identity must be able to GET both
secrets.

### spec.link1.macsecCakKeyvaultSecretId

`string`

The Key Vault SECRET IDENTIFIER (a versioned https URL, not the
secret itself) holding the MACsec CAK -- the connectivity
association key, the actual encryption key material. Travels
together with the CKN.

### spec.link1.macsecSciEnabled

`bool`

Enable MACsec SCI (secure channel identifier) tagging. Both sides
of the link must agree on this setting; leave it off unless the
facility side requires SCI.

### spec.link2

`AzureExpressRoutePortLink`

Configuration of the SECOND physical link of the pair -- same
semantics as link1.

- rule: MACsec needs both keys or neither -- set macsec_ckn_keyvault_secret_id and macsec_cak_keyvault_secret_id together

### spec.link2.adminEnabled

`bool`

Administratively enable the link. Links start DISABLED; enable them
once the facility completes the physical cross-connect. ARM applies
link configuration in a second call after the port exists, so a
fresh port briefly reports links disabled even when this is true.

### spec.link2.macsecCipher

`enum` · optional (explicit presence)

The MACsec cipher suite used when MACsec is configured. Unspecified
applies GCM_AES_128 (ARM's default). The XPN variants (extended
packet numbering) are for sustained rates above ~40 Gbps where the
32-bit packet number would wrap too quickly.

- default: `GCM_AES_128`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_port_macsec_cipher_unspecified` -- Not specified -- GCM_AES_128 (ARM's default) applies.
- `GCM_AES_128` -- AES-GCM with a 128-bit key.
- `GCM_AES_256` -- AES-GCM with a 256-bit key.
- `GCM_AES_XPN_128` -- AES-GCM 128-bit with extended (64-bit) packet numbering -- for sustained rates where a 32-bit packet number wraps too quickly.
- `GCM_AES_XPN_256` -- AES-GCM 256-bit with extended packet numbering.

### spec.link2.macsecCknKeyvaultSecretId

`string`

The Key Vault SECRET IDENTIFIER (a versioned https URL, not the
secret itself) holding the MACsec CKN -- the connectivity key name,
a hex string shared with the facility side. Travels together with
the CAK; the port's user-assigned identity must be able to GET both
secrets.

### spec.link2.macsecCakKeyvaultSecretId

`string`

The Key Vault SECRET IDENTIFIER (a versioned https URL, not the
secret itself) holding the MACsec CAK -- the connectivity
association key, the actual encryption key material. Travels
together with the CKN.

### spec.link2.macsecSciEnabled

`bool`

Enable MACsec SCI (secure channel identifier) tagging. Both sides
of the link must agree on this setting; leave it off unless the
facility side requires SCI.

### spec.authorizations

`[]AzureExpressRoutePortAuthorization`

Authorizations ISSUED by this port: each entry creates a named
authorization whose generated key lets a circuit in ANOTHER
subscription be built on this port's capacity. The generated keys
surface (marked sensitive) in the port's authorization_keys output,
keyed by name. Deleting an entry revokes the authorization.

### spec.authorizations[].name

`string` · required

The authorization's name, unique on the port -- the key's lookup
name in the authorization_keys output. Renaming revokes the old
authorization and issues a new key.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tags

`map<string, string>`

Free-form tags applied to the port, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins.

## Validation Rules

- `encapsulation_required`: Choose the link encapsulation explicitly -- DOT1Q (one VLAN tag per circuit, the common choice) or QINQ (stacked tags for overlapping VLAN ranges)
- `macsec_requires_user_assigned_identity`: MACsec needs a USER_ASSIGNED (or SYSTEM_AND_USER_ASSIGNED) identity on the port -- the identity reads the CAK/CKN secrets from Key Vault
- `authorization_names_unique`: Authorization names must be unique on the port

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureExpressRoutePort, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.express_route_port_id` | `string` | The Azure Resource Manager ID of the port -- what an ExpressRoute Direct circuit references as express_route_port_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/expressRoutePorts/{name} |
| `status.outputs.express_route_port_name` | `string` | The name of the port -- what port authorizations reference cloud-side. |
| `status.outputs.guid` | `string` | The port's globally unique resource GUID -- what some provider-side ordering systems ask for to identify the port. |
| `status.outputs.ethertype` | `string` | The link ethertype (e.g. "0x8100" for Dot1Q) -- fixed by the encapsulation choice. |
| `status.outputs.mtu` | `string` | The maximum transmission unit of the links (e.g. "1500"). |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the port's system-assigned identity, when the identity type includes SYSTEM_ASSIGNED (empty otherwise). |
| `status.outputs.link1_id` | `string` | The ARM ID of the first physical link. |
| `status.outputs.link1_router_name` | `string` | The Microsoft edge router of the first link -- part of the letter of authorization handed to the facility. |
| `status.outputs.link1_interface_name` | `string` | The router interface of the first link. |
| `status.outputs.link1_patch_panel_id` | `string` | The patch panel the first link lands on -- what the facility's cross-connect order references. |
| `status.outputs.link1_rack_id` | `string` | The rack of the first link's patch panel. |
| `status.outputs.link1_connector_type` | `string` | The physical connector type of the first link (e.g. "LC"). |
| `status.outputs.link2_id` | `string` | The ARM ID of the second physical link. |
| `status.outputs.link2_router_name` | `string` | The Microsoft edge router of the second link. |
| `status.outputs.link2_interface_name` | `string` | The router interface of the second link. |
| `status.outputs.link2_patch_panel_id` | `string` | The patch panel the second link lands on. |
| `status.outputs.link2_rack_id` | `string` | The rack of the second link's patch panel. |
| `status.outputs.link2_connector_type` | `string` | The physical connector type of the second link. |
| `status.outputs.authorization_keys` | `map<string, string>` (sensitive) | The generated key of each authorization ISSUED by this port, keyed by the authorization's name from the spec. A circuit in another subscription redeems one to be built on this port. Marked sensitive in both engines. Example valueFrom fieldPath: status.outputs.authorization_keys.partner-team |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureExpressRouteCircuit | `spec.expressRoutePortId` | `status.outputs.express_route_port_id` |

## See Also

- [Overview](../README.md)

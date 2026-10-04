# AzureVirtualWan

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVirtualWanSpec** defines a Virtual WAN -- the top-level umbrella
of Azure's managed hub-and-spoke networking. The WAN object itself is
a lightweight, free container of policy: the actual regional routers
(virtual hubs) and the gateways attached to them are separate
resources that reference this WAN. Create the WAN first; everything
else in the Virtual WAN world hangs off it.

**Basic vs Standard**: a Standard WAN (the default and the right
choice for almost everyone) supports the full mesh -- ExpressRoute,
site-to-site and point-to-site VPN, hub-to-hub and any-to-any transit
routing. Basic is a constrained legacy tier limited to site-to-site
VPN in a Basic hub, and a WAN can be UPGRADED Basic-to-Standard but
never downgraded.

**ForceNew fields**: `name`, `region`, and `resource_group` are fixed
at creation -- changing any of them replaces the WAN (and orphans
nothing by itself, but hubs must be deleted first; ARM refuses to
delete a WAN that still has hubs).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualWan
metadata:
  name: test-virtual-wan
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: global-wan
  # The unset optional fields apply ARM's defaults: type "Standard"
  # (the full-mesh tier), branch-to-branch transit on, VPN encryption
  # on, no Office 365 local breakout.
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.disableVpnEncryption` | `bool` |  |  |  |
| `spec.allowBranchToBranchTraffic` | `bool` |  | `true` |  |
| `spec.office365LocalBreakoutCategory` | `enum` |  | `NONE` |  |
| `spec.type` | `string` |  | `Standard` |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the WAN object lives in, e.g. "eastus". This is
ARM metadata placement only -- hubs choose their own regions, and a
WAN spans all of them. Changing the region replaces the WAN.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the WAN is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the WAN.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The WAN's name, unique within the resource group. Hubs reference
the WAN by its ARM ID (the virtual_wan_id output). Changing the
name replaces the WAN.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.disableVpnEncryption

`bool`

Disable IPsec encryption for VPN traffic transiting the WAN.
Off by default (traffic is encrypted); enabling this is a niche
performance trade for private-circuit-only topologies.

### spec.allowBranchToBranchTraffic

`bool` · optional (explicit presence)

Allow branches (VPN sites) connected to different hubs to reach
each other through the WAN. Unspecified applies true (ARM's
default) -- branch-to-branch transit is most of the point of a
Virtual WAN; set false only to force hub-and-spoke-only reachability
where branches must not see each other.

- default: `true`

### spec.office365LocalBreakoutCategory

`enum` · optional (explicit presence)

Route Office 365 traffic out of local branch internet breakouts
instead of hauling it through the WAN. Unspecified applies NONE
(ARM's default -- no local breakout). OPTIMIZE covers the
latency-sensitive endpoints Microsoft marks Optimize;
OPTIMIZE_AND_ALLOW adds the Allow category; ALL breaks out every
Office 365 endpoint.

- default: `NONE`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_virtual_wan_office365_breakout_category_unspecified` -- Not specified -- NONE (ARM's default) applies.
- `NONE` -- No local breakout: Office 365 traffic transits the WAN like everything else.
- `ALL` -- Break out every Office 365 endpoint locally.
- `OPTIMIZE` -- Break out only the endpoints Microsoft marks Optimize (the latency-sensitive core: Exchange, SharePoint, Teams media).
- `OPTIMIZE_AND_ALLOW` -- Break out the Optimize and Allow categories.

### spec.type

`string` · optional (explicit presence)

The WAN type. Unspecified applies "Standard" (ARM's default and the
full-mesh tier: ExpressRoute, S2S/P2S VPN, hub-to-hub transit).
"Basic" is a constrained legacy tier (site-to-site VPN only, Basic
hubs); ARM upgrades Basic-to-Standard in place but never
downgrades. The value is a free string on ARM's side -- these two
are the only published values.

- default: `Standard`

### spec.tags

`map<string, string>`

Free-form tags applied to the WAN, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualWan, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_wan_id` | `string` | The Azure Resource Manager ID of the WAN -- what virtual hubs reference as their virtual_wan_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualWans/{name} |
| `status.outputs.virtual_wan_name` | `string` | The name of the WAN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualHub | `spec.virtualWanId` | `status.outputs.virtual_wan_id` |
| AzureVpnSite | `spec.virtualWanId` | `status.outputs.virtual_wan_id` |

## See Also

- [Overview](../README.md)

# AzureVpnSite

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureVpnSiteSpec** defines a VPN site -- the Virtual WAN world's
DESCRIPTION of one branch location: its internet links (each with a
public endpoint and optional BGP speaker), the address space
reachable behind it, and the device that terminates the tunnels. It
deploys nothing at the branch and costs nothing to keep; it is the
address-book entry an AzureVpnGatewayConnection points at. The
classic-world sibling (without a Virtual WAN) is
AzureLocalNetworkGateway.

**Links are the connectable unit**: a connection's vpn_links each pin
to ONE of this site's links by its ARM ID (published in the site's
name-keyed link_ids output). A site with two links (e.g. two ISPs)
gets a two-link connection -- that is how Virtual WAN does
active-active branch connectivity.

**ForceNew fields**: `name`, `region`, `resource_group`, and
`virtual_wan_id` -- every other field updates in place (a connection
pins each link by ID, so renaming or removing a CONNECTED link is a
far-side change, not a site-only edit).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnSite
metadata:
  name: test-vpn-site
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: branch-london
  virtualWanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualWans/global-wan
  addressCidrs:
    - "192.168.10.0/24"
  deviceVendor: Cisco
  deviceModel: ISR4331
  links:
    - name: primary-isp
      providerName: ExampleCarrier
      speedInMbps: 200
      ipAddress: "203.0.113.10"
    - name: backup-isp
      speedInMbps: 100
      ipAddress: "198.51.100.7"
      bgp:
        asn: 65010
        peeringAddress: "169.254.21.2"
  o365Policy:
    trafficCategory:
      optimizeEndpointEnabled: true
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
| `spec.addressCidrs` | `[]string` |  |  |  |
| `spec.deviceVendor` | `string` |  |  |  |
| `spec.deviceModel` | `string` |  |  |  |
| `spec.links` | `[]AzureVpnSiteLink` |  |  |  |
| `spec.links[].name` | `string` | yes |  |  |
| `spec.links[].providerName` | `string` |  |  |  |
| `spec.links[].speedInMbps` | `int32` |  |  |  |
| `spec.links[].ipAddress` | `string` |  |  |  |
| `spec.links[].fqdn` | `string` |  |  |  |
| `spec.links[].bgp` | `AzureVpnSiteLinkBgp` |  |  |  |
| `spec.links[].bgp.asn` | `int64` |  |  |  |
| `spec.links[].bgp.peeringAddress` | `string` | yes |  |  |
| `spec.o365Policy` | `AzureVpnSiteO365Policy` |  |  |  |
| `spec.o365Policy.trafficCategory` | `AzureVpnSiteO365TrafficCategory` |  |  |  |
| `spec.o365Policy.trafficCategory.allowEndpointEnabled` | `bool` |  | `false` |  |
| `spec.o365Policy.trafficCategory.defaultEndpointEnabled` | `bool` |  | `false` |  |
| `spec.o365Policy.trafficCategory.optimizeEndpointEnabled` | `bool` |  | `false` |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the VPN site object lives in, e.g. "eastus". By
convention the region of the virtual hub whose gateway connects to
it. Changing the region replaces the object.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the VPN site is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The VPN site's name, unique within the resource group. Name it
after the branch it describes ("branch-london", "factory-pune") --
a connection references it per site. Must not contain any of
' < > % & : ? / + (the provider's own name rule). Changing the
name replaces the object.

- rule: {"required":true,"string":{"pattern":"^[^'<>%&:?/+]+$"}}

### spec.virtualWanId

`string | valueFrom` · required

The Virtual WAN this site belongs to -- references an
AzureVirtualWan's ARM ID. Sites are WAN-scoped so any hub's VPN
gateway in the WAN can connect to them. Fixed at creation.

- references: AzureVirtualWan (`status.outputs.virtual_wan_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualWan, name: <that resource's name>, fieldPath: status.outputs.virtual_wan_id}} -- a bare string does not parse

### spec.addressCidrs

`[]string`

The address space reachable behind the branch, in CIDR notation --
the prefixes Azure routes into the tunnels. Leave empty only when
every link speaks BGP (learned routes then carry the routing, the
AzureLocalNetworkGateway convention).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.deviceVendor

`string`

The branch VPN device's vendor (e.g. "Cisco") -- informational
metadata Azure surfaces in the portal and to partners.

### spec.deviceModel

`string`

The branch VPN device's model (e.g. "ISR4331") -- informational,
like device_vendor.

### spec.links

`[]AzureVpnSiteLink`

The branch's internet links -- the connectable endpoints of the
site. Each link's ARM ID surfaces in the site's name-keyed
link_ids output; a connection's vpn_links pin to them. Two links
(two ISPs) enable active-active branch connectivity.

- rule: Give the link a public endpoint: ip_address (static public IP) or fqdn (re-resolved name)

### spec.links[].name

`string` · required

The link's name, unique on the site (e.g. "primary-isp"). The
link's ARM ID surfaces in the site's link_ids output under this
name -- what a connection's vpn_site_link_id references.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.links[].providerName

`string`

The ISP behind this link (e.g. "Airtel") -- informational metadata
Azure surfaces to SD-WAN partners.

### spec.links[].speedInMbps

`int32`

The link's bandwidth in Mbps -- informational (Azure uses it for
partner automation and portal display, not for rate limiting; the
connection's bandwidth_mbps is separate). 0 (the provider's
default) means unspecified.

- rule: {"int32":{"gte":0}}

### spec.links[].ipAddress

`string`

The link's PUBLIC IP address -- where the hub's VPN gateway sends
tunnel traffic. One of ip_address or fqdn (ARM rejects a link with
neither; when both are set, the IP wins).

- rule: ip_address must be an IP address

### spec.links[].fqdn

`string`

The link's public FQDN (for branches whose public IP changes --
Azure re-resolves it). One of ip_address or fqdn.

### spec.links[].bgp

`AzureVpnSiteLinkBgp`

The BGP speaker behind this link -- set it to exchange routes
dynamically instead of (or alongside) the site's static
address_cidrs. The connection's matching vpn_link must enable BGP
too.

### spec.links[].bgp.asn

`int64`

The branch BGP speaker's Autonomous System Number. Must differ
from the hub router's 65515; 65515-65520 are Azure-reserved.

- rule: {"int64":{"gte":"1"}}

### spec.links[].bgp.peeringAddress

`string` · required

The branch BGP speaker's peering address -- an IP INSIDE the
tunnel (typically on the device's tunnel interface), not the
link's public address.

- rule: peering_address must be an IP address
- rule: {"required":true}

### spec.o365Policy

`AzureVpnSiteO365Policy`

Office 365 breakout policy for this branch (SD-WAN partners read
it): which O365 traffic categories exit the branch's local
internet instead of riding the tunnel. Leave unset for no
breakout.

### spec.o365Policy.trafficCategory

`AzureVpnSiteO365TrafficCategory`

The per-category breakout switches.

### spec.o365Policy.trafficCategory.allowEndpointEnabled

`bool` · optional (explicit presence)

Break out the "Allow" category (required O365 endpoints tolerant
of local egress) at the branch. Unset keeps it on the tunnel.

- default: `false`

### spec.o365Policy.trafficCategory.defaultEndpointEnabled

`bool` · optional (explicit presence)

Break out the "Default" category (everything else O365) at the
branch. Unset keeps it on the tunnel.

- default: `false`

### spec.o365Policy.trafficCategory.optimizeEndpointEnabled

`bool` · optional (explicit presence)

Break out the "Optimize" category (the latency-critical endpoints:
Teams media, Exchange) at the branch. Unset keeps it on the tunnel.

- default: `false`

### spec.tags

`map<string, string>`

Free-form tags applied to the object, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `link_names_unique`: Link names must be unique on the site -- each is the key connections and the link_ids output use

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVpnSite, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpn_site_id` | `string` | The Azure Resource Manager ID of the VPN site -- what a connection references as its remote_vpn_site_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/vpnSites/{name} |
| `status.outputs.vpn_site_name` | `string` | The name of the VPN site. |
| `status.outputs.link_ids` | `map<string, string>` | The ARM ID of each link on the site, keyed by the link's name from the spec -- what a connection's vpn_links reference as their vpn_site_link_id. Example valueFrom fieldPath: status.outputs.link_ids.primary-isp |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualWanId` | AzureVirtualWan | `status.outputs.virtual_wan_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVpnGatewayConnection | `spec.remoteVpnSiteId` | `status.outputs.vpn_site_id` |
| AzureVpnGatewayConnection | `spec.vpnLinks[].vpnSiteLinkId` | `status.outputs.link_ids` |

## See Also

- [Overview](../README.md)

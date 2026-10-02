# GcpHaVpnGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpHaVpnGatewaySpec creates one HA VPN GATEWAY and the Cloud Router its
tunnels speak BGP through. The gateway is the Google Cloud end of every
IPsec VPN from this VPC: it has two interfaces, each with its own public
IP (the two `interface_*_ip_address` outputs are what the on-premises
device is configured to reach), and Google's 99.99% availability SLA
holds when both interfaces carry a tunnel to redundant peers.

Declare the gateway ONCE per VPC and region, then declare each site or
peer cloud it connects to as a GcpHaVpnConnection that references this
gateway: the tunnels, their pre-shared keys, and their BGP sessions live
there and can be added or removed without touching the gateway. Two
Google Cloud VPCs peer over VPN by each declaring a gateway and a
connection pointing at the other's gateway.

Nearly everything here is IMMUTABLE: the network, region, IP version,
stack type, and interface pinning recreate the gateway -- and a recreated
gateway has NEW public IPs, so every on-premises device must be
reconfigured. Labels and the router's description and BGP advertisement
change in place; the router's ASN does not.

Cost: the gateway itself is free; Google bills per tunnel-hour and for
the traffic that crosses the tunnels, both of which are declared on the
GcpHaVpnConnection.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnGateway
metadata:
  name: planton-oss-e2e-gcpvpngw
  id: planton-oss-e2e-gcpvpngw
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcphavpngateway
  annotations:
    planton.dev/e2e: "true"
  tags:
    - planton-e2e
spec:
  # Gateway and router names are project+region-scoped with no soft-delete
  # reservation, so they stay fixed; gateway_name defaults to metadata.name.
  region: us-central1

  # The prerequisite GcpVpcNetwork, by reference.
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: planton-oss-e2e-gcpvpcnetwork-prereq
      fieldPath: status.outputs.network_self_link

  description: Planton E2E HA VPN gateway; safe to delete

  # Two public IPv4 interfaces (the defaults), IPv4-only tunnels.
  labels:
    purpose: e2e

  # The Cloud Router the tunnels' BGP sessions run on: a 16-bit private ASN,
  # default advertisement (every subnet).
  router:
    description: Planton E2E HA VPN router; safe to delete
    bgp:
      asn: 64514

  # DELETE (default; refused while tunnels reference the gateway), PREVENT,
  # ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.gatewayName` | `string` |  |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.description` | `string` |  |  |  |
| `spec.gatewayIpVersion` | `string` |  |  |  |
| `spec.stackType` | `string` |  |  |  |
| `spec.vpnInterfaces` | `[]GcpHaVpnGatewayVpnInterface` |  |  |  |
| `spec.vpnInterfaces[].id` | `int32` |  |  |  |
| `spec.vpnInterfaces[].interconnectAttachment` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.router` | `GcpHaVpnGatewayRouter` | yes |  |  |
| `spec.router.name` | `string` |  |  |  |
| `spec.router.description` | `string` |  |  |  |
| `spec.router.bgp` | `GcpHaVpnGatewayRouterBgp` | yes |  |  |
| `spec.router.bgp.asn` | `uint32` | yes |  |  |
| `spec.router.bgp.advertiseMode` | `string` |  |  |  |
| `spec.router.bgp.advertisedGroups` | `[]string` |  |  |  |
| `spec.router.bgp.advertisedIpRanges` | `[]GcpHaVpnGatewayRouterBgpAdvertisedIpRange` |  |  |  |
| `spec.router.bgp.advertisedIpRanges[].range` | `string` | yes |  |  |
| `spec.router.bgp.advertisedIpRanges[].description` | `string` |  |  |  |
| `spec.router.bgp.keepaliveInterval` | `int32` |  |  |  |
| `spec.router.bgp.identifierRange` | `string` |  |  |  |
| `spec.router.encryptedInterconnectRouter` | `bool` |  |  |  |
| `spec.router.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the gateway and router are created in: a reference to
a GcpProject or the project ID as a literal. Empty means the provider's
default project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.gatewayName

`string`

Name of the VPN gateway in GCP. Defaults to metadata.name when empty.
1-63 characters, lowercase letters, digits, and hyphens, starting with
a letter. Immutable.

- rule: gateway_name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.region

`string` · required

The region the gateway and router live in (e.g. "us-central1"). Every
tunnel of every connection on this gateway is in this region. Immutable.

- rule: {"required":true}

### spec.network

`string | valueFrom` · required

The VPC network the gateway attaches to: a reference to a GcpVpcNetwork
(its network_self_link output) or the network's self link as a literal.
The tunnels carry traffic for THIS network's subnets (and, with custom
routes, whatever the router advertises). Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.description

`string`

Human-readable description of the gateway. Immutable (Google keeps it
create-time on this resource).

### spec.gatewayIpVersion

`string`

The IP version of the gateway's two public interface addresses:
  ""     -- same as IPV4 (provider default)
  "IPV4" -- two public IPv4 addresses; what every on-premises device
            and every peer cloud supports
  "IPV6" -- two public IPv6 addresses; the peer must reach the gateway
            over IPv6 (Google-to-Google over IPv6, or an IPv6-capable
            device). The traffic INSIDE the tunnels is governed by
            stack_type, not by this.
Immutable.

- rule: gateway_ip_version must be IPV4 or IPV6

### spec.stackType

`string`

Which IP versions the TUNNELS carry between the networks:
  ""          -- same as IPV4_ONLY (provider default)
  "IPV4_ONLY" -- IPv4 traffic only
  "IPV4_IPV6" -- dual-stack: the VPC must have an internal IPv6 range
                 and each BGP session enables IPv6 to exchange IPv6
                 routes
  "IPV6_ONLY" -- IPv6 traffic only
Immutable.

- rule: stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY

### spec.vpnInterfaces

`[]GcpHaVpnGatewayVpnInterface`

Pin the gateway's interfaces to Cloud Interconnect VLAN attachments for
HA VPN over Interconnect. Empty (the normal case) means an
internet-facing gateway whose interfaces get public IPs. At most two
entries, one per interface. Requires
router.encrypted_interconnect_router. Immutable.

- rule: {"repeated":{"maxItems":"2"}}

### spec.vpnInterfaces[].id

`int32`

Which of the gateway's two interfaces this entry configures: 0 or 1.
Each interface may appear once.

- rule: {"int32":{"lte":1,"gte":0}}

### spec.vpnInterfaces[].interconnectAttachment

`string` · required

The Interconnect attachment this interface rides, as a self link or
`projects/{p}/regions/{r}/interconnectAttachments/{name}`. The
attachment must be an encrypted attachment (`encryption: IPSEC`) in the
same region, and the gateway's router must be an
encrypted_interconnect_router. Immutable.

- rule: {"required":true}

### spec.labels

`map<string, string>`

User labels on the gateway, merged with Planton's platform labels
(which win on key conflicts). The one mutable surface on the gateway
resource itself.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the gateway for org-policy and IAM
conditions. Keys `tagKeys/{id}`, values `tagValues/{id}`. Create-time
only: a change replaces the gateway (and its public IPs).

### spec.router

`GcpHaVpnGatewayRouter` · required

The Cloud Router the gateway's tunnels terminate BGP on -- its name,
ASN, and default advertisement. Required: an HA VPN without a BGP
router routes nothing.

- rule: {"required":true}

### spec.router.name

`string`

Name of the Cloud Router. Defaults to the gateway's name when empty --
a router and a gateway are different resource types, so the same name
never collides. 1-63 characters, lowercase letters, digits, and
hyphens, starting with a letter. Immutable.

- rule: router name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.router.description

`string`

Human-readable description of the router. Mutable.

### spec.router.bgp

`GcpHaVpnGatewayRouterBgp` · required

The router's BGP configuration: the ASN Google speaks as and the
default route advertisement. Required -- an HA VPN router exists to
run BGP, and every tunnel session needs the ASN.

- rule: {"required":true}
- rule: advertised_groups can only be set when advertise_mode is CUSTOM
- rule: advertised_ip_ranges can only be set when advertise_mode is CUSTOM

### spec.router.bgp.asn

`uint32` · required

Local BGP Autonomous System Number -- the number Google Cloud speaks as
toward every on-premises or peer-cloud router. Must be an RFC 6996
private ASN, 16-bit (64512-65534) or 32-bit (4200000000-4294967294),
and DIFFERENT from every peer's ASN. Fixed for the router's lifetime;
a change recreates the router and every tunnel session on it.

- rule: asn must be a private ASN (64512-65534 or 4200000000-4294967294)
- rule: {"required":true}

### spec.router.bgp.advertiseMode

`string`

Route advertisement mode. DEFAULT (the value when empty): advertise
every subnet range of the VPC to every peer. CUSTOM: advertise only
what advertised_groups and advertised_ip_ranges specify -- the shape
for exposing a curated set of ranges to the other side.

- rule: advertise_mode must be DEFAULT or CUSTOM

### spec.router.bgp.advertisedGroups

`[]string`

Prefix groups to advertise in CUSTOM mode, in addition to
advertised_ip_ranges. The API accepts exactly one group value:
ALL_SUBNETS (re-adds the default subnet advertisement on top of the
custom ranges).

- rule: {"repeated":{"items":{"string":{"in":["ALL_SUBNETS"]}}}}

### spec.router.bgp.advertisedIpRanges

`[]GcpHaVpnGatewayRouterBgpAdvertisedIpRange`

Individual IP ranges to advertise in CUSTOM mode, sent to all peers
in addition to any advertised_groups.

### spec.router.bgp.advertisedIpRanges[].range

`string` · required

The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").

- rule: {"required":true}

### spec.router.bgp.advertisedIpRanges[].description

`string`

Human-readable description of this advertised range.

### spec.router.bgp.keepaliveInterval

`int32`

Interval in seconds between BGP keepalive messages (20-60; 20 when
empty). Hold time -- how long a peer waits before declaring the session
dead -- is three times this value. Lower values fail over faster at
the cost of more control traffic; BFD on the session is the precise
tool for sub-second detection.

- rule: keepalive_interval must be between 20 and 60 seconds

### spec.router.bgp.identifierRange

`string`

Explicit range of valid BGP identifiers (what other vendors call the
router ID), as a link-local IPv4 CIDR from 169.254.0.0/16 of size at
least /30 (e.g. "169.254.8.0/30"). Must not overlap any session's
interface range on this router. Leave empty to let GCP choose.

### spec.router.encryptedInterconnectRouter

`bool`

Dedicate this router to encrypted Interconnect VLAN attachments (HA VPN
over Cloud Interconnect). Required when vpn_interfaces pins the gateway
to attachments; must be false otherwise. Immutable: an encrypted router
cannot be converted later.

### spec.router.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the router for org-policy and IAM
conditions. Keys `tagKeys/{id}`, values `tagValues/{id}`. Create-time
only: a change replaces the router.

### spec.deletionPolicy

`string`

What destroying this resource does to the gateway and router in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- both are deleted; fails while any GcpHaVpnConnection's
               tunnels still reference the gateway (destroy the
               connections first -- a chart's dependency order does
               this when they reference the gateway)
  "PREVENT" -- destroy FAILS; the guard for the gateway every site's
               device is configured against
  "ABANDON" -- the resources leave management but stay live in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `interconnect_interfaces_need_encrypted_router`: pinning vpn_interfaces to Interconnect attachments requires router.encrypted_interconnect_router = true (HA VPN over Cloud Interconnect); an internet-facing gateway leaves vpn_interfaces empty
- `vpn_interface_ids_unique`: each vpn_interfaces entry must name a different interface id (0 or 1)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpHaVpnGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.gateway_self_link` | `string` | The gateway's self link -- what a GcpHaVpnConnection's `gateway` references (its tunnels attach here), and what another Google Cloud gateway's connection names as its `peer.gcp_gateway`. |
| `status.outputs.gateway_name` | `string` | The gateway's name in GCP (the resolved value: metadata.name when the spec left gateway_name empty). |
| `status.outputs.region` | `string` | The region the gateway and router live in -- what a GcpHaVpnConnection's `region` references so the two can never disagree. |
| `status.outputs.interface_0_ip_address` | `string` | The public IP of gateway interface 0. Configure the on-premises device's first tunnel toward this address. |
| `status.outputs.interface_1_ip_address` | `string` | The public IP of gateway interface 1. Configure the on-premises device's second tunnel toward this address; a tunnel on each interface is what earns the 99.99% SLA. |
| `status.outputs.router_name` | `string` | The Cloud Router's name -- what a GcpHaVpnConnection's `router` references so its interfaces and BGP peers land on this router. |
| `status.outputs.router_self_link` | `string` | The Cloud Router's self link. |
| `status.outputs.router_asn` | `uint32` | The ASN the router speaks as -- the number the on-premises side configures as its BGP neighbor's ASN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpHaVpnConnection | `spec.gateway` | `status.outputs.gateway_self_link` |
| GcpHaVpnConnection | `spec.router` | `status.outputs.router_name` |
| GcpHaVpnConnection | `spec.region` | `status.outputs.region` |
| GcpHaVpnConnection | `spec.peer.gcpGateway` | `status.outputs.gateway_self_link` |

## See Also

- [Overview](../README.md)

# GcpHaVpnConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpHaVpnConnectionSpec connects one HA VPN gateway (GcpHaVpnGateway) to
ONE peer -- an on-premises site, another cloud, or another Google Cloud
VPC -- with one to four IPsec tunnels, each carrying a BGP session on the
gateway's Cloud Router. Declare one connection per site; add or remove
sites without touching the gateway.

The recommended 99.99% topology is TWO tunnels: one from each gateway
interface, to two addresses of the peer (redundancy_type
TWO_IPS_REDUNDANCY, or a Google peer, which always has two). Each tunnel
gets its own link-local /30 and BGP session; both advertise the same
routes so traffic fails over in seconds (or sub-second with BFD).

Composition: this kind creates the tunnels, one Cloud Router interface
and one BGP peer per tunnel on the gateway's router, and -- when the
peer is an external device -- the external VPN gateway resource that
holds its addresses. Nothing here has a life apart from the connection.

Cost: Google bills each tunnel per hour while it exists (a two-tunnel
connection is two tunnel-hours per hour) plus the traffic that leaves
Google through the tunnels at internet egress rates. Tunnels bill from
creation, whether or not the peer is up.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnConnection
metadata:
  name: planton-oss-e2e-gcpvpncn
  id: planton-oss-e2e-gcpvpncn
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcphavpnconnection
  annotations:
    planton.dev/e2e: "true"
    # Google-to-Google in one project: the hub gateway is the registry
    # prerequisite; the spoke network, spoke gateway, and the reverse
    # connection are extra fixture instances, in deploy order. BGP establishes
    # only when both directions exist.
    planton.dev/e2e-prerequisites: "catalog/gcp/gcphavpnconnection/e2e/prerequisites/gcpvpcnetwork-spoke.yaml,catalog/gcp/gcphavpnconnection/e2e/prerequisites/gcphavpngateway-spoke.yaml,catalog/gcp/gcphavpnconnection/e2e/prerequisites/gcphavpnconnection-spoke-to-hub.yaml"
  tags:
    - planton-e2e
spec:
  # The hub gateway, its router, and its region: three references to the same
  # prerequisite GcpHaVpnGateway.
  gateway:
    valueFrom:
      kind: GcpHaVpnGateway
      name: planton-oss-e2e-gcpvpngw-prereq
      fieldPath: status.outputs.gateway_self_link
  router:
    valueFrom:
      kind: GcpHaVpnGateway
      name: planton-oss-e2e-gcpvpngw-prereq
      fieldPath: status.outputs.router_name
  region:
    valueFrom:
      kind: GcpHaVpnGateway
      name: planton-oss-e2e-gcpvpngw-prereq
      fieldPath: status.outputs.region

  # The peer: the spoke fixture gateway. Google pairs interfaces itself.
  peer:
    gcpGateway:
      valueFrom:
        kind: GcpHaVpnGateway
        name: planton-oss-e2e-gcpvpncn-spoke-gw
        fieldPath: status.outputs.gateway_self_link

  # Two tunnels, one per gateway interface -- the 99.99% shape. The secrets
  # are fixture literals (the peer is our own fixture); a real manifest wires
  # them from a secrets manager. Tunnel names are project+region-scoped with
  # no soft-delete reservation.
  tunnels:
    - name: planton-oss-e2e-gcpvpncn-h2s-0
      vpnGatewayInterface: 0
      sharedSecret: planton-e2e-psk-0-not-a-real-secret
      bgpSession:
        interfaceIpRange: 169.254.100.1/30
        peerAsn: 64515
    - name: planton-oss-e2e-gcpvpncn-h2s-1
      vpnGatewayInterface: 1
      sharedSecret: planton-e2e-psk-1-not-a-real-secret
      bgpSession:
        interfaceIpRange: 169.254.101.1/30
        peerAsn: 64515

  # DELETE (default), PREVENT, ABANDON -- fanned to every resource.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.gateway` | `string \| valueFrom` | yes |  | GcpHaVpnGateway (`status.outputs.gateway_self_link`) |
| `spec.router` | `string \| valueFrom` | yes |  | GcpHaVpnGateway (`status.outputs.router_name`) |
| `spec.region` | `string \| valueFrom` | yes |  | GcpHaVpnGateway (`status.outputs.region`) |
| `spec.peer` | `GcpHaVpnConnectionPeer` | yes |  |  |
| `spec.peer.externalGateway` | `GcpHaVpnConnectionExternalGateway` |  |  |  |
| `spec.peer.externalGateway.name` | `string` |  |  |  |
| `spec.peer.externalGateway.redundancyType` | `string` | yes |  |  |
| `spec.peer.externalGateway.interfaces` | `[]GcpHaVpnConnectionExternalGatewayInterface` | yes |  |  |
| `spec.peer.externalGateway.interfaces[].id` | `int32` |  |  |  |
| `spec.peer.externalGateway.interfaces[].ipAddress` | `string` |  |  |  |
| `spec.peer.externalGateway.interfaces[].ipv6Address` | `string` |  |  |  |
| `spec.peer.externalGateway.description` | `string` |  |  |  |
| `spec.peer.externalGateway.labels` | `map<string, string>` |  |  |  |
| `spec.peer.gcpGateway` | `string \| valueFrom` |  |  | GcpHaVpnGateway (`status.outputs.gateway_self_link`) |
| `spec.tunnels` | `[]GcpHaVpnConnectionTunnel` | yes |  |  |
| `spec.tunnels[].name` | `string` | yes |  |  |
| `spec.tunnels[].vpnGatewayInterface` | `int32` |  |  |  |
| `spec.tunnels[].peerExternalGatewayInterface` | `int32` |  |  |  |
| `spec.tunnels[].sharedSecret` | `string` (sensitive) | yes |  |  |
| `spec.tunnels[].ikeVersion` | `int32` |  | `2` |  |
| `spec.tunnels[].localTrafficSelector` | `[]string` |  |  |  |
| `spec.tunnels[].remoteTrafficSelector` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite` | `GcpHaVpnConnectionCipherSuite` |  |  |  |
| `spec.tunnels[].cipherSuite.phase1` | `GcpHaVpnConnectionCipherSuitePhase1` |  |  |  |
| `spec.tunnels[].cipherSuite.phase1.encryption` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase1.integrity` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase1.prf` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase1.dh` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase2` | `GcpHaVpnConnectionCipherSuitePhase2` |  |  |  |
| `spec.tunnels[].cipherSuite.phase2.encryption` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase2.integrity` | `[]string` |  |  |  |
| `spec.tunnels[].cipherSuite.phase2.pfs` | `[]string` |  |  |  |
| `spec.tunnels[].labels` | `map<string, string>` |  |  |  |
| `spec.tunnels[].bgpSession` | `GcpHaVpnConnectionBgpSession` | yes |  |  |
| `spec.tunnels[].bgpSession.name` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.interfaceIpRange` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.ipVersion` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.peerAsn` | `uint32` | yes |  |  |
| `spec.tunnels[].bgpSession.peerIpAddress` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.advertisedRoutePriority` | `int32` |  |  |  |
| `spec.tunnels[].bgpSession.advertiseMode` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.advertisedGroups` | `[]string` |  |  |  |
| `spec.tunnels[].bgpSession.advertisedIpRanges` | `[]GcpHaVpnConnectionBgpAdvertisedIpRange` |  |  |  |
| `spec.tunnels[].bgpSession.advertisedIpRanges[].range` | `string` | yes |  |  |
| `spec.tunnels[].bgpSession.advertisedIpRanges[].description` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.enable` | `bool` |  | `true` |  |
| `spec.tunnels[].bgpSession.enableIpv4` | `bool` |  |  |  |
| `spec.tunnels[].bgpSession.enableIpv6` | `bool` |  |  |  |
| `spec.tunnels[].bgpSession.ipv6NexthopAddress` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.peerIpv6NexthopAddress` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.ipv4NexthopAddress` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.peerIpv4NexthopAddress` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.customLearnedIpRanges` | `[]GcpHaVpnConnectionBgpCustomLearnedIpRange` |  |  |  |
| `spec.tunnels[].bgpSession.customLearnedIpRanges[].range` | `string` | yes |  |  |
| `spec.tunnels[].bgpSession.customLearnedRoutePriority` | `int32` |  |  |  |
| `spec.tunnels[].bgpSession.bfd` | `GcpHaVpnConnectionBgpBfd` |  |  |  |
| `spec.tunnels[].bgpSession.bfd.sessionInitializationMode` | `string` | yes |  |  |
| `spec.tunnels[].bgpSession.bfd.minReceiveInterval` | `int32` |  |  |  |
| `spec.tunnels[].bgpSession.bfd.minTransmitInterval` | `int32` |  |  |  |
| `spec.tunnels[].bgpSession.bfd.multiplier` | `int32` |  |  |  |
| `spec.tunnels[].bgpSession.md5AuthenticationKey` | `GcpHaVpnConnectionBgpMd5AuthenticationKey` |  |  |  |
| `spec.tunnels[].bgpSession.md5AuthenticationKey.name` | `string` |  |  |  |
| `spec.tunnels[].bgpSession.md5AuthenticationKey.key` | `string` (sensitive) | yes |  |  |
| `spec.tunnels[].bgpSession.importPolicies` | `[]string` |  |  |  |
| `spec.tunnels[].bgpSession.exportPolicies` | `[]string` |  |  |  |
| `spec.tunnels[].description` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the tunnels and router interfaces are created in: a
reference to a GcpProject or the project ID as a literal. Must be the
gateway's project. Empty means the provider's default project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.gateway

`string | valueFrom` · required

The HA VPN gateway the tunnels leave from: a reference to a
GcpHaVpnGateway (its gateway_self_link output) or the gateway's self
link as a literal. Immutable.

- references: GcpHaVpnGateway (`status.outputs.gateway_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHaVpnGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_self_link}} -- a bare string does not parse

### spec.router

`string | valueFrom` · required

The Cloud Router the BGP sessions run on -- the gateway's router: a
reference to the same GcpHaVpnGateway (its router_name output) or the
router's name as a literal. Must be in the gateway's region and
network. Immutable.

- references: GcpHaVpnGateway (`status.outputs.router_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHaVpnGateway, name: <that resource's name>, fieldPath: status.outputs.router_name}} -- a bare string does not parse

### spec.region

`string | valueFrom` · required

The region of the gateway and router -- a reference to the same
GcpHaVpnGateway (its region output) or the region as a literal, so a
connection can never name a region its gateway is not in. Immutable.

- references: GcpHaVpnGateway (`status.outputs.region`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHaVpnGateway, name: <that resource's name>, fieldPath: status.outputs.region}} -- a bare string does not parse

### spec.peer

`GcpHaVpnConnectionPeer` · required

What the tunnels connect to: an external device or another Google
Cloud HA VPN gateway. Required.

- rule: {"required":true}
- rule: set exactly one of external_gateway (an on-premises or other-cloud device) or gcp_gateway (another Google Cloud HA VPN gateway)

### spec.peer.externalGateway

`GcpHaVpnConnectionExternalGateway`

The other side is an on-premises appliance or another cloud's VPN
endpoint, described by its public addresses. This connection creates
the external VPN gateway resource for it.

- rule: interfaces must match redundancy_type: 1 interface for SINGLE_IP_INTERNALLY_REDUNDANT, 2 for TWO_IPS_REDUNDANCY, 4 for FOUR_IPS_REDUNDANCY
- rule: interface ids must be unique and below the interface count (0, 0-1, or 0-3)

### spec.peer.externalGateway.name

`string`

Name of the external VPN gateway resource in GCP. Defaults to the
connection's metadata.name when empty. 1-63 characters, lowercase
letters, digits, and hyphens, starting with a letter. Immutable.

- rule: external gateway name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.peer.externalGateway.redundancyType

`string` · required

How many public addresses the device exposes, which fixes how many
interfaces it has and which HA topology the tunnels form:
  "SINGLE_IP_INTERNALLY_REDUNDANT" -- one address (the device is
      redundant behind it); two tunnels from the two gateway
      interfaces to that one address earn 99.9%
  "TWO_IPS_REDUNDANCY"  -- two addresses (two devices or one with two
      uplinks); one tunnel per gateway interface to each address --
      two tunnels -- earn Google's 99.99% SLA. The recommended shape.
  "FOUR_IPS_REDUNDANCY" -- four addresses; two tunnels per gateway
      interface, four in all, also 99.99%
Immutable.

- rule: redundancy_type must be SINGLE_IP_INTERNALLY_REDUNDANT, TWO_IPS_REDUNDANCY, or FOUR_IPS_REDUNDANCY
- rule: {"required":true}

### spec.peer.externalGateway.interfaces

`[]GcpHaVpnConnectionExternalGatewayInterface` · required

The device's public addresses, one per interface, ids 0..n-1. The
count must match redundancy_type. Immutable.

- rule: {"repeated":{"minItems":"1","maxItems":"4"}}
- rule: each external gateway interface sets exactly one of ip_address or ipv6_address

### spec.peer.externalGateway.interfaces[].id

`int32`

The interface's numeric id. Which ids are legal depends on the
gateway's redundancy_type: 0 for SINGLE_IP_INTERNALLY_REDUNDANT; 0 and
1 for TWO_IPS_REDUNDANCY; 0-3 for FOUR_IPS_REDUNDANCY. Each id appears
once.

- rule: {"int32":{"lte":3,"gte":0}}

### spec.peer.externalGateway.interfaces[].ipAddress

`string`

The public IPv4 address of the device's interface. Cannot be a Google
Compute Engine address (use peer.gcp_gateway for Google-to-Google).
Set this or ipv6_address, matching the gateway's gateway_ip_version.

- rule: ip_address must be an IPv4 address

### spec.peer.externalGateway.interfaces[].ipv6Address

`string`

The public IPv6 address of the device's interface, in any RFC 4291
form (e.g. 2001:db8::2d9:51:0:0). Cannot be a Compute Engine address.

- rule: ipv6_address must be an IPv6 address

### spec.peer.externalGateway.description

`string`

Human-readable description of the device. Immutable.

### spec.peer.externalGateway.labels

`map<string, string>`

User labels on the external gateway resource, merged with Planton's
platform labels (which win on key conflicts). Mutable.

### spec.peer.gcpGateway

`string | valueFrom`

The other side is another Google Cloud HA VPN gateway -- VPN between
two VPCs, in the same or different projects or organizations: a
reference to a GcpHaVpnGateway (its gateway_self_link output) or the
gateway's self link as a literal. Google pairs interfaces
automatically (tunnel on interface 0 here connects to interface 0
there), so tunnels leave peer_external_gateway_interface empty. The
other VPC declares its own GcpHaVpnConnection pointing back at this
side's gateway; the two together form the link.

- references: GcpHaVpnGateway (`status.outputs.gateway_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHaVpnGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_self_link}} -- a bare string does not parse

### spec.tunnels

`[]GcpHaVpnConnectionTunnel` · required

The tunnels, one to four, each with its BGP session. Two -- one per
gateway interface -- is the recommended HA shape.

- rule: {"repeated":{"minItems":"1","maxItems":"4"}}

### spec.tunnels[].name

`string` · required

Name of the tunnel in GCP. Required; unique in the region. 1-63
characters, lowercase letters, digits, and hyphens, starting with a
letter. Also the default name of the tunnel's BGP session. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.tunnels[].vpnGatewayInterface

`int32`

Which of the Google gateway's two interfaces this tunnel leaves from:
0 or 1. For the 99.99% SLA put at least one tunnel on each. Immutable.

- rule: {"int32":{"lte":1,"gte":0}}

### spec.tunnels[].peerExternalGatewayInterface

`int32` · optional (explicit presence)

Which interface of the EXTERNAL gateway this tunnel lands on (its id
in peer.external_gateway.interfaces). Set when the peer is an external
device; leave empty when the peer is a Google Cloud gateway (Google
pairs interfaces itself). Sent only when set. Immutable.

- rule: {"int32":{"lte":3,"gte":0}}

### spec.tunnels[].sharedSecret

`string` · required · sensitive

The IKE pre-shared key both ends authenticate with. Sensitive: never
logged or exported; in Terraform state only its hash is kept. No
content rule, because sensitive fields hold a managed-secret reference
on consuming platforms and a content-shape rule would reject every
reference. Google accepts 1-63 printable characters. Immutable: a
rotation recreates the tunnel.

- rule: {"string":{"minLen":"1"}}

### spec.tunnels[].ikeVersion

`int32` · optional (explicit presence)

IKE protocol version: 1 or 2. 2 when empty (Google's default and the
only version that supports IPv6, BFD-friendly rekeying, and modern
ciphers). Use 1 only for a legacy device. Always sent. Immutable.

- default: `2`
- rule: {"int32":{"lte":2,"gte":1}}

### spec.tunnels[].localTrafficSelector

`[]string`

Local traffic selectors -- the IPv4 CIDRs on the Google side the
tunnel carries when the tunnel is used in ROUTE-BASED (policy-less)
mode without BGP. With a BGP session (the normal HA VPN shape) Google
sets 0.0.0.0/0 and this stays empty; sent only when set. The ranges
must be disjoint. Immutable.

- rule: {"repeated":{"items":{"cel":[{"id":"valid_cidr","message":"each traffic selector must be an IPv4 CIDR like 10.0.0.0/16","expression":"this.isIpPrefix(4)"}]}}}

### spec.tunnels[].remoteTrafficSelector

`[]string`

Remote traffic selectors -- the IPv4 CIDRs on the peer side, same
rules as local_traffic_selector. Immutable.

- rule: {"repeated":{"items":{"cel":[{"id":"valid_cidr","message":"each traffic selector must be an IPv4 CIDR like 192.168.0.0/16","expression":"this.isIpPrefix(4)"}]}}}

### spec.tunnels[].cipherSuite

`GcpHaVpnConnectionCipherSuite`

Restrict the IKE ciphers this tunnel negotiates. Leave unset to accept
Google's defaults (which already exclude broken algorithms). Immutable.

### spec.tunnels[].cipherSuite.phase1

`GcpHaVpnConnectionCipherSuitePhase1`

Phase-1 (IKE SA) ciphers.

### spec.tunnels[].cipherSuite.phase1.encryption

`[]string`

Encryption algorithms.

### spec.tunnels[].cipherSuite.phase1.integrity

`[]string`

Integrity algorithms.

### spec.tunnels[].cipherSuite.phase1.prf

`[]string`

Pseudo-random functions.

### spec.tunnels[].cipherSuite.phase1.dh

`[]string`

Diffie-Hellman groups.

### spec.tunnels[].cipherSuite.phase2

`GcpHaVpnConnectionCipherSuitePhase2`

Phase-2 (child / IPsec SA) ciphers.

### spec.tunnels[].cipherSuite.phase2.encryption

`[]string`

Encryption algorithms.

### spec.tunnels[].cipherSuite.phase2.integrity

`[]string`

Integrity algorithms.

### spec.tunnels[].cipherSuite.phase2.pfs

`[]string`

Perfect-forward-secrecy groups.

### spec.tunnels[].labels

`map<string, string>`

User labels on the tunnel, merged with Planton's platform labels
(which win on key conflicts). The one mutable field on a tunnel.

### spec.tunnels[].bgpSession

`GcpHaVpnConnectionBgpSession` · required

The BGP session inside this tunnel: the router interface and peer.
Required -- an HA VPN tunnel without BGP carries no routes.

- rule: {"required":true}
- rule: advertised_groups can only be set when advertise_mode is CUSTOM
- rule: advertised_ip_ranges can only be set when advertise_mode is CUSTOM
- rule: peer_ip_address, when set, must be a link-local address in 169.254.0.0/16 like interface_ip_range (Google requires both ends of an IPv4 session in the same /30)

### spec.tunnels[].bgpSession.name

`string`

Name of the Cloud Router interface and BGP peer this session creates
(both take the same name). Defaults to the tunnel's name when empty.
RFC 1035, 1-63 characters; unique on the router. Immutable.

- rule: session name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.tunnels[].bgpSession.interfaceIpRange

`string`

The Cloud Router's address on the tunnel, as a link-local CIDR from
169.254.0.0/16 with a /30 mask (e.g. "169.254.10.1/30"): Google takes
the address as its own end and the peer takes the other usable address
of the /30. Every session on the router needs a different /30 and it
must not overlap the router's identifier_range. Required for IPv4
sessions; Google assigns one when empty only for IPv6-only sessions.
Immutable.

- rule: interface_ip_range must be a link-local IPv4 CIDR from 169.254.0.0/16 with a /30 mask, e.g. 169.254.10.1/30

### spec.tunnels[].bgpSession.ipVersion

`string` · optional (explicit presence)

The IP version of the router interface:
  ""     -- Google infers it from interface_ip_range (IPv4) or from the
            gateway's stack type
  "IPV4"
  "IPV6" -- for IPv6-only sessions on an IPV6_ONLY or dual-stack
            gateway; Google assigns the link-local addresses
Sent only when set. Immutable.

- rule: ip_version must be IPV4 or IPV6

### spec.tunnels[].bgpSession.peerAsn

`uint32` · required

The peer router's ASN -- the number the on-premises device (or the
other Google Cloud router) speaks as. Must differ from this router's
ASN. Any public or private ASN. Required. Mutable.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.tunnels[].bgpSession.peerIpAddress

`string` · optional (explicit presence)

The peer's address on the tunnel: the other usable address of
interface_ip_range (e.g. "169.254.10.2" for 169.254.10.1/30). Google
derives it from the interface range when empty, so it is normally left
unset; set it only when the device insists on a specific address.
Sent only when set. Immutable.

- rule: peer_ip_address must be an IPv4 address

### spec.tunnels[].bgpSession.advertisedRoutePriority

`int32` · optional (explicit presence)

The priority (MED) of routes advertised to this peer: where the peer
has more than one matching route of equal length, the LOWEST value
wins. Set different priorities on the tunnels of one connection to
make the peer prefer one tunnel (active/passive) instead of splitting
traffic (active/active, the default when equal). 0-65535; sent only
when set, so 0 is a real priority distinct from "let Google choose".
Mutable.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.tunnels[].bgpSession.advertiseMode

`string`

Per-session override of the router's advertisement mode. Empty
inherits the router's; DEFAULT advertises every subnet; CUSTOM
advertises only advertised_groups and advertised_ip_ranges below.
Mutable.

- rule: advertise_mode must be DEFAULT or CUSTOM

### spec.tunnels[].bgpSession.advertisedGroups

`[]string`

Prefix groups to advertise to this peer in CUSTOM mode; the API
accepts exactly ALL_SUBNETS. Overrides the router's list for this
session.

- rule: {"repeated":{"items":{"string":{"in":["ALL_SUBNETS"]}}}}

### spec.tunnels[].bgpSession.advertisedIpRanges

`[]GcpHaVpnConnectionBgpAdvertisedIpRange`

Individual ranges to advertise to this peer in CUSTOM mode, in
addition to any advertised_groups. Overrides the router's list for
this session.

### spec.tunnels[].bgpSession.advertisedIpRanges[].range

`string` · required

The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").

- rule: {"required":true}

### spec.tunnels[].bgpSession.advertisedIpRanges[].description

`string`

Human-readable description of this advertised range.

### spec.tunnels[].bgpSession.enable

`bool` · optional (explicit presence)

Whether the session is up. Default true. Setting false tears the
session down and withdraws its routes without deleting anything -- the
switch for draining a tunnel before maintenance. Always sent. Mutable.

- default: `true`

### spec.tunnels[].bgpSession.enableIpv4

`bool` · optional (explicit presence)

Carry IPv4 routes over this session. Google enables it by default when
the peer address is IPv4, so it is sent only when set -- set false on
a dual-stack gateway to make a session IPv6-only. Mutable.

### spec.tunnels[].bgpSession.enableIpv6

`bool`

Carry IPv6 routes over this session. Default false. Requires a gateway
with stack_type IPV4_IPV6 or IPV6_ONLY and a VPC with an internal IPv6
range. Mutable.

### spec.tunnels[].bgpSession.ipv6NexthopAddress

`string` · optional (explicit presence)

Google Cloud's IPv6 next-hop address for the session, from
2600:2d00:0:2::/64 or 2600:2d00:0:3::/64. Google assigns one when
empty; sent only when set. Mutable.

- rule: ipv6_nexthop_address must be an IPv6 address

### spec.tunnels[].bgpSession.peerIpv6NexthopAddress

`string` · optional (explicit presence)

The peer's IPv6 next-hop address, same range and rules as
ipv6_nexthop_address. Sent only when set. Mutable.

- rule: peer_ipv6_nexthop_address must be an IPv6 address

### spec.tunnels[].bgpSession.ipv4NexthopAddress

`string` · optional (explicit presence)

Google Cloud's IPv4 next-hop address for an IPv4-over-IPv6 session
(IPv4 routes exchanged over an IPv6 BGP session). Google assigns one
when empty; sent only when set. Mutable.

- rule: ipv4_nexthop_address must be an IPv4 address

### spec.tunnels[].bgpSession.peerIpv4NexthopAddress

`string` · optional (explicit presence)

The peer's IPv4 next-hop address for an IPv4-over-IPv6 session. Sent
only when set. Mutable.

- rule: peer_ipv4_nexthop_address must be an IPv4 address

### spec.tunnels[].bgpSession.customLearnedIpRanges

`[]GcpHaVpnConnectionBgpCustomLearnedIpRange`

Ranges to treat as learned from this peer even though it did not
advertise them -- for devices that cannot advertise everything behind
them. Mutable.

### spec.tunnels[].bgpSession.customLearnedIpRanges[].range

`string` · required

A CIDR prefix (an address without a mask is /32 or /128).

- rule: {"required":true}

### spec.tunnels[].bgpSession.customLearnedRoutePriority

`int32` · optional (explicit presence)

Priority applied to every custom learned range of this session.
0-65335; Google uses 100 when empty. Sent only when set, so 0 is a real
priority. Mutable.

- rule: {"int32":{"lte":65335,"gte":0}}

### spec.tunnels[].bgpSession.bfd

`GcpHaVpnConnectionBgpBfd`

BFD for fast failure detection on this session. Leave unset to rely
on BGP keepalives alone.

### spec.tunnels[].bgpSession.bfd.sessionInitializationMode

`string` · required

Who starts the BFD session:
  "ACTIVE"   -- Cloud Router initiates (the usual choice)
  "PASSIVE"  -- Cloud Router waits for the peer to initiate
  "DISABLED" -- BFD off for this session

- rule: session_initialization_mode must be ACTIVE, PASSIVE, or DISABLED
- rule: {"required":true}

### spec.tunnels[].bgpSession.bfd.minReceiveInterval

`int32`

Minimum interval in milliseconds between BFD packets RECEIVED from the
peer; the negotiated value is the greater of this and the peer's
transmit interval. 1000-30000; 1000 when empty.

- rule: min_receive_interval must be between 1000 and 30000 milliseconds

### spec.tunnels[].bgpSession.bfd.minTransmitInterval

`int32`

Minimum interval in milliseconds between BFD packets TRANSMITTED to
the peer; negotiated the same way. 1000-30000; 1000 when empty.

- rule: min_transmit_interval must be between 1000 and 30000 milliseconds

### spec.tunnels[].bgpSession.bfd.multiplier

`int32`

How many consecutive BFD packets may be missed before the peer is
declared down. 5-16; 5 when empty. Detection time is roughly the
negotiated interval times this multiplier.

- rule: multiplier must be between 5 and 16

### spec.tunnels[].bgpSession.md5AuthenticationKey

`GcpHaVpnConnectionBgpMd5AuthenticationKey`

MD5 authentication of the BGP session. Leave unset for an
unauthenticated session (the tunnel's IPsec already protects the
control traffic; MD5 defends against a misconfigured peer, not an
attacker).

### spec.tunnels[].bgpSession.md5AuthenticationKey.name

`string`

Name of the key. RFC 1035, 1-63 characters. Defaults to the tunnel's
name with a `-md5` suffix when empty.

- rule: md5 key name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.tunnels[].bgpSession.md5AuthenticationKey.key

`string` · required · sensitive

The key material -- the same string configured on the peer router.
Sensitive: never logged or exported. No content rule, because sensitive
fields hold a managed-secret reference on consuming platforms and a
content-shape rule would reject every reference.

- rule: {"string":{"minLen":"1"}}

### spec.tunnels[].bgpSession.importPolicies

`[]string`

Names of route policies (ROUTE_POLICY_TYPE_IMPORT) on the router,
applied to routes learned from this peer in order. Route policies are
created outside this kind (`gcloud compute routers update-route-policy`
or the provider's router route-policy resource) and referenced here by
name. Mutable.

### spec.tunnels[].bgpSession.exportPolicies

`[]string`

Names of route policies (ROUTE_POLICY_TYPE_EXPORT) on the router,
applied to routes advertised to this peer in order. Same provenance as
import_policies. Mutable.

### spec.tunnels[].description

`string`

Human-readable description of the tunnel. Immutable (Google keeps it
create-time on this resource).

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to every tunnel (and the external gateway,
when created) for org-policy and IAM conditions. Keys `tagKeys/{id}`,
values `tagValues/{id}`. Create-time only: a change replaces the
tunnels.

### spec.deletionPolicy

`string`

What destroying this resource does to the tunnels, router interfaces,
BGP peers, and external gateway in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- all are deleted; the site is disconnected
  "PREVENT" -- destroy FAILS; the guard for a production site link
  "ABANDON" -- the resources leave management but stay live in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `tunnel_names_unique`: each tunnel must have a distinct name
- `session_ranges_unique`: each tunnel's bgp_session.interface_ip_range must be a different /30
- `external_peer_tunnels_name_their_interface`: with peer.external_gateway every tunnel sets peer_external_gateway_interface; with peer.gcp_gateway none does (Google pairs the interfaces)
- `external_interface_ids_exist`: each tunnel's peer_external_gateway_interface must name an interface declared in peer.external_gateway.interfaces

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpHaVpnConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.tunnel_self_links` | `[]string` | The self links of the tunnels, in spec order. |
| `status.outputs.tunnel_names` | `[]string` | The tunnel names in GCP, in spec order. |
| `status.outputs.router_interface_names` | `[]string` | The names of the Cloud Router interfaces created for the sessions, in spec order (one per tunnel). |
| `status.outputs.bgp_peer_names` | `[]string` | The names of the BGP peers created for the sessions, in spec order (one per tunnel). `gcloud compute routers get-status` reports each peer's BGP state under this name. |
| `status.outputs.external_gateway_self_link` | `string` | The external VPN gateway's self link, when the peer is an external device; empty for a Google-to-Google connection. |
| `status.outputs.gateway_self_link` | `string` | The gateway the tunnels leave from, as a self link (the resolved reference). |
| `status.outputs.router_name` | `string` | The Cloud Router the sessions run on, by name (the resolved reference). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.gateway` | GcpHaVpnGateway | `status.outputs.gateway_self_link` |
| `spec.router` | GcpHaVpnGateway | `status.outputs.router_name` |
| `spec.region` | GcpHaVpnGateway | `status.outputs.region` |
| `spec.peer.gcpGateway` | GcpHaVpnGateway | `status.outputs.gateway_self_link` |

## See Also

- [Overview](../README.md)

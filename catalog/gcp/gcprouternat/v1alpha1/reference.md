# GcpRouterNat

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpRouterNatSpec defines a Cloud Router with a NAT gateway — the managed
egress path that lets instances without external IPs reach the internet
(public NAT) or other private networks (private NAT).

The router and the NAT are provisioned together as one node, and the
router's full inline surface is configurable here: BGP ASN and route
advertisement (router_bgp), description, encrypted-Interconnect
dedication, and resource-manager tags. Interface- and peer-level BGP
(the objects that terminate VPN tunnels and Interconnect attachments on
the router) are separate resources composed on top of this node.

Important behavioral notes:

  - router_name, nat_name, region, the network, endpoint_types, and type
    are immutable — changing them replaces the resource. Everything else
    (IP allocation, subnetwork scoping, port tuning, timeouts, logging,
    rules) updates in place, which is what makes NAT IP rotation and
    fleet-wide egress tuning zero-downtime operations.
  - With no nat_ips, GCP auto-allocates external IPs (AUTO_ONLY) and may
    change them as capacity scales. Reference GcpAddress reservations in
    nat_ips (MANUAL_ONLY) when third parties allowlist your egress IPs;
    the literal IPs live on those address nodes' outputs.
  - drain_nat_ips is an update-time lever: move an IP from nat_ips to
    drain_nat_ips to bleed connections off it before releasing it. The
    API rejects drain entries on a brand-new NAT and entries that were
    never in nat_ips.

## Example

```yaml
# Exercises the deep NAT surface offline: explicit subnetwork scoping with a
# secondary-range selection, the router's custom BGP advertisement, dynamic
# port allocation with power-of-two bounds, tuned timeouts, a NAT rule with
# a dedicated egress IP, full translation logging, and an explicit deletion
# policy.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRouterNat
metadata:
  name: hack-router-nat
spec:
  # project_id omitted — falls back to the provider's default project.
  routerName: hack-router
  natName: hack-nat
  region: us-central1
  vpcSelfLink:
    value: projects/hack-project/global/networks/hack-vpc
  routerDescription: egress router for the hack fleet
  # BGP posture for a router that also serves VPN/Interconnect sessions:
  # a private ASN and a curated custom advertisement.
  routerBgp:
    asn: 64514
    advertiseMode: CUSTOM
    advertisedGroups:
      - ALL_SUBNETS
    advertisedIpRanges:
      - range: 10.10.0.0/16
        description: on-prem reachable services
    keepaliveInterval: 30
  sourceSubnetworkIpRangesToNat: LIST_OF_SUBNETWORKS
  subnetworks:
    - subnetwork:
        value: projects/hack-project/regions/us-central1/subnetworks/hack-subnet
      sourceIpRangesToNat:
        - PRIMARY_IP_RANGE
        - LIST_OF_SECONDARY_IP_RANGES
      secondaryIpRangeNames:
        - pods
  natIps:
    - value: projects/hack-project/regions/us-central1/addresses/hack-nat-ip-a
    - value: projects/hack-project/regions/us-central1/addresses/hack-nat-ip-b
  minPortsPerVm: 64
  maxPortsPerVm: 4096
  enableDynamicPortAllocation: true
  udpIdleTimeoutSec: 30
  tcpEstablishedIdleTimeoutSec: 1200
  tcpTransitoryIdleTimeoutSec: 30
  tcpTimeWaitTimeoutSec: 60
  rules:
    - ruleNumber: 100
      match: destination.ip == '203.0.113.10'
      description: partner API egress pinned to a dedicated IP
      action:
        sourceNatActiveIps:
          - value: projects/hack-project/regions/us-central1/addresses/hack-nat-ip-b
  logFilter: ALL
  # DELETE keeps destroys real; PREVENT belongs on production egress paths.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.routerName` | `string` | yes |  |  |
| `spec.natName` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.vpcSelfLink` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.routerBgp` | `GcpRouterNatRouterBgp` |  |  |  |
| `spec.routerBgp.asn` | `uint32` |  |  |  |
| `spec.routerBgp.advertiseMode` | `string` |  |  |  |
| `spec.routerBgp.advertisedGroups` | `[]string` |  |  |  |
| `spec.routerBgp.advertisedIpRanges` | `[]GcpRouterNatRouterBgpAdvertisedIpRange` |  |  |  |
| `spec.routerBgp.advertisedIpRanges[].range` | `string` | yes |  |  |
| `spec.routerBgp.advertisedIpRanges[].description` | `string` |  |  |  |
| `spec.routerBgp.keepaliveInterval` | `int32` |  |  |  |
| `spec.routerBgp.identifierRange` | `string` |  |  |  |
| `spec.routerDescription` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.sourceSubnetworkIpRangesToNat` | `string` |  |  |  |
| `spec.subnetworks` | `[]GcpRouterNatSubnetwork` |  |  |  |
| `spec.subnetworks[].subnetwork` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.subnetworks[].sourceIpRangesToNat` | `[]string` |  |  |  |
| `spec.subnetworks[].secondaryIpRangeNames` | `[]string` |  |  |  |
| `spec.natIps` | `[]string \| valueFrom` |  |  | GcpAddress (`status.outputs.self_link`) |
| `spec.drainNatIps` | `[]string \| valueFrom` |  |  | GcpAddress (`status.outputs.self_link`) |
| `spec.autoNetworkTier` | `string` |  |  |  |
| `spec.minPortsPerVm` | `int32` |  |  |  |
| `spec.maxPortsPerVm` | `int32` |  |  |  |
| `spec.enableDynamicPortAllocation` | `bool` |  |  |  |
| `spec.enableEndpointIndependentMapping` | `bool` |  |  |  |
| `spec.endpointTypes` | `[]string` |  |  |  |
| `spec.udpIdleTimeoutSec` | `int32` |  |  |  |
| `spec.icmpIdleTimeoutSec` | `int32` |  |  |  |
| `spec.tcpEstablishedIdleTimeoutSec` | `int32` |  |  |  |
| `spec.tcpTransitoryIdleTimeoutSec` | `int32` |  |  |  |
| `spec.tcpTimeWaitTimeoutSec` | `int32` |  |  |  |
| `spec.rules` | `[]GcpRouterNatRule` |  |  |  |
| `spec.rules[].ruleNumber` | `uint32` |  |  |  |
| `spec.rules[].match` | `string` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].action` | `GcpRouterNatRuleAction` |  |  |  |
| `spec.rules[].action.sourceNatActiveIps` | `[]string \| valueFrom` |  |  | GcpAddress (`status.outputs.self_link`) |
| `spec.rules[].action.sourceNatDrainIps` | `[]string \| valueFrom` |  |  | GcpAddress (`status.outputs.self_link`) |
| `spec.rules[].action.sourceNatActiveRanges` | `[]string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.rules[].action.sourceNatDrainRanges` | `[]string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.logFilter` | `enum` |  | `ERRORS_ONLY` |  |
| `spec.encryptedInterconnectRouter` | `bool` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.sourceSubnetworkIpRangesToNat64` | `string` |  |  |  |
| `spec.nat64Subnetworks` | `[]string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project ID where the Cloud Router and NAT will be created.
If not specified, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.routerName

`string` · required

Name of the Cloud Router to create in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a lowercase letter or number.
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.natName

`string` · required

Name of the NAT configuration on the Cloud Router.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a lowercase letter or number.
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.region

`string` · required

GCP region for the Cloud Router and NAT (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true}

### spec.vpcSelfLink

`string | valueFrom` · required

The VPC network the router attaches to. Accepts a network self link.
Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.routerBgp

`GcpRouterNatRouterBgp`

The router's BGP configuration — ASN, route advertisement, keepalive,
identifier range. Only needed when the router also serves BGP
sessions (Cloud VPN / Interconnect on the same router); NAT-only
routers leave this unset and GCP assigns an ASN.

- rule: advertised_groups can only be set when advertise_mode is CUSTOM
- rule: advertised_ip_ranges can only be set when advertise_mode is CUSTOM

### spec.routerBgp.asn

`uint32`

Local BGP Autonomous System Number. Must be an RFC6996 private ASN,
16-bit (64512-65534) or 32-bit (4200000000-4294967294). Fixed for the
router's lifetime; every VPN tunnel and Interconnect attachment on
this router shares it.

- rule: asn must be a private ASN (64512-65534 or 4200000000-4294967294)

### spec.routerBgp.advertiseMode

`string`

Route advertisement mode. DEFAULT (the value when empty): advertise
all subnets visible to the router. CUSTOM: advertise only what
advertised_groups and advertised_ip_ranges specify — the shape for
exposing a curated set of ranges over VPN/Interconnect.

- rule: advertise_mode must be DEFAULT or CUSTOM

### spec.routerBgp.advertisedGroups

`[]string`

Prefix groups to advertise in CUSTOM mode, in addition to
advertised_ip_ranges. The API accepts exactly one group value:
ALL_SUBNETS (re-adds the default subnet advertisement on top of the
custom ranges).

- rule: {"repeated":{"items":{"string":{"in":["ALL_SUBNETS"]}}}}

### spec.routerBgp.advertisedIpRanges

`[]GcpRouterNatRouterBgpAdvertisedIpRange`

Individual IP ranges to advertise in CUSTOM mode, sent to all peers
of the router in addition to any advertised_groups.

### spec.routerBgp.advertisedIpRanges[].range

`string` · required

The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").

- rule: {"required":true}

### spec.routerBgp.advertisedIpRanges[].description

`string`

Human-readable description of this advertised range.

### spec.routerBgp.keepaliveInterval

`int32`

Interval in seconds between BGP keepalive messages (20-60, default
20). Hold time — how long a peer waits before declaring the session
dead — is three times this value.

- rule: keepalive_interval must be between 20 and 60 seconds

### spec.routerBgp.identifierRange

`string`

Explicit range of valid BGP identifiers (what other vendors call the
router ID), as a link-local IPv4 CIDR from 169.254.0.0/16 of size at
least /30 (e.g. "169.254.8.0/30"). Must not overlap any IPv4 BGP
session range on the router. Leave empty to let GCP choose.

### spec.routerDescription

`string`

Human-readable description of the Cloud Router's purpose.

### spec.type

`string`

NAT gateway type. PUBLIC (default): NAT to the internet using external
IPs. PRIVATE: NAT between VPC networks (Network Connectivity Center
spokes) using subnetwork ranges — no external IPs involved.
Immutable after creation.

- rule: type must be PUBLIC or PRIVATE

### spec.sourceSubnetworkIpRangesToNat

`string`

Which subnetworks (and which of their IP ranges) get NAT.
ALL_SUBNETWORKS_ALL_IP_RANGES (default when empty and subnetworks is
empty): every subnetwork in the region, primary + secondary ranges.
ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES: every subnetwork, primary only.
LIST_OF_SUBNETWORKS (implied when subnetworks is non-empty): only the
subnetworks listed below.

- rule: source_subnetwork_ip_ranges_to_nat must be ALL_SUBNETWORKS_ALL_IP_RANGES, ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES, or LIST_OF_SUBNETWORKS

### spec.subnetworks

`[]GcpRouterNatSubnetwork`

Per-subnetwork NAT scoping. Listing any subnetwork here selects
LIST_OF_SUBNETWORKS mode: only the listed subnetworks are NATed.

- rule: secondary_ip_range_names is required when source_ip_ranges_to_nat contains LIST_OF_SECONDARY_IP_RANGES
- rule: secondary_ip_range_names is only meaningful when source_ip_ranges_to_nat contains LIST_OF_SECONDARY_IP_RANGES

### spec.subnetworks[].subnetwork

`string | valueFrom` · required

The subnetwork to NAT. Accepts a subnetwork self link.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.subnetworks[].sourceIpRangesToNat

`[]string`

Which IP ranges of the subnetwork are NATed.
ALL_IP_RANGES: primary and all secondary ranges (the default when empty).
PRIMARY_IP_RANGE: only the primary range.
LIST_OF_SECONDARY_IP_RANGES: only the secondary ranges named in
secondary_ip_range_names — the shape for NATing GKE pod ranges without
exposing the node range, or vice versa.

- rule: {"repeated":{"items":{"string":{"in":["ALL_IP_RANGES","PRIMARY_IP_RANGE","LIST_OF_SECONDARY_IP_RANGES"]}}}}

### spec.subnetworks[].secondaryIpRangeNames

`[]string`

Names of the secondary ranges to NAT. Required when
source_ip_ranges_to_nat contains LIST_OF_SECONDARY_IP_RANGES.

### spec.natIps

`[]string | valueFrom`

Static external IPs the NAT translates through, each referencing a
GcpAddress reservation (EXTERNAL, in this region) by self link.
Non-empty selects MANUAL_ONLY allocation — the shape for stable,
allowlistable egress IPs. Empty selects AUTO_ONLY (GCP manages the
pool). Public NAT only.

- references: GcpAddress (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.drainNatIps

`[]string | valueFrom`

IPs being drained: existing connections continue, new connections stop.
Each entry must already be present in nat_ips (the API enforces this and
rejects drain entries on a brand-new NAT). Public NAT only.

- references: GcpAddress (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.autoNetworkTier

`string`

Network tier for auto-allocated NAT IPs: PREMIUM (default) or STANDARD.
Only meaningful for public NAT with auto-allocation.

- rule: auto_network_tier must be PREMIUM or STANDARD

### spec.minPortsPerVm

`int32`

Minimum ports reserved per VM (default: 64 static, 32 dynamic). Raise
when instances open many concurrent connections to the same
destination; must be a power of two when dynamic port allocation is
enabled.

- rule: {"int32":{"gte":0}}

### spec.maxPortsPerVm

`int32`

Maximum ports per VM. Only valid with dynamic port allocation, where it
caps how far a busy VM's allocation can grow; must be a power of two.

- rule: {"int32":{"gte":0}}

### spec.enableDynamicPortAllocation

`bool`

Dynamic port allocation: VMs start at min_ports_per_vm and grow toward
max_ports_per_vm on demand — better pool utilization for mixed
workloads. Mutually exclusive with endpoint-independent mapping.

### spec.enableEndpointIndependentMapping

`bool`

Endpoint-independent mapping: the same internal ip:port maps to the
same NAT ip:port regardless of destination (RFC 5128) — required by
some peer-to-peer and SIP workloads. Mutually exclusive with dynamic
port allocation.

### spec.endpointTypes

`[]string`

Which resource types this NAT serves: ENDPOINT_TYPE_VM (default),
ENDPOINT_TYPE_SWG (Secure Web Gateway), or ENDPOINT_TYPE_MANAGED_PROXY_LB
(regional load balancer proxies). A NAT serves exactly one endpoint
type. Immutable after creation.

- rule: {"repeated":{"maxItems":"1","items":{"string":{"in":["ENDPOINT_TYPE_VM","ENDPOINT_TYPE_SWG","ENDPOINT_TYPE_MANAGED_PROXY_LB"]}}}}

### spec.udpIdleTimeoutSec

`int32`

Idle timeout for UDP connections, in seconds (default 30).

- rule: {"int32":{"gte":0}}

### spec.icmpIdleTimeoutSec

`int32`

Idle timeout for ICMP connections, in seconds (default 30).

- rule: {"int32":{"gte":0}}

### spec.tcpEstablishedIdleTimeoutSec

`int32`

Idle timeout for established TCP connections, in seconds (default 1200).

- rule: {"int32":{"gte":0}}

### spec.tcpTransitoryIdleTimeoutSec

`int32`

Idle timeout for transitory (half-open) TCP connections, in seconds
(default 30).

- rule: {"int32":{"gte":0}}

### spec.tcpTimeWaitTimeoutSec

`int32`

How long a TCP connection lingers in TIME_WAIT before its NAT port is
reusable, in seconds (default 120). Lowering this frees ports faster
for high-churn workloads at the cost of stricter RFC conformance.

- rule: {"int32":{"gte":0}}

### spec.rules

`[]GcpRouterNatRule`

NAT rules: route matching egress connections through dedicated NAT IPs
or ranges (e.g. a stable source IP for one partner's API).

### spec.rules[].ruleNumber

`uint32`

Rule priority (0-65000). Lower numbers win when multiple rules match.

- rule: {"uint32":{"lte":65000}}

### spec.rules[].match

`string` · required

CEL expression selecting the traffic this rule applies to, evaluated
against destination attributes. Example:
"destination.ip == '203.0.113.10' || destination.ip == '203.0.113.11'"

- rule: {"required":true}

### spec.rules[].description

`string`

Human-readable description of the rule's intent.

### spec.rules[].action

`GcpRouterNatRuleAction`

The NAT IPs or ranges used for matching traffic.

### spec.rules[].action.sourceNatActiveIps

`[]string | valueFrom`

Static external IPs to NAT matching traffic through (public NAT only).
Each entry references a GcpAddress reservation by self link.

- references: GcpAddress (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.rules[].action.sourceNatDrainIps

`[]string | valueFrom`

IPs being drained out of this rule (public NAT only): existing
connections continue, no new connections are established. An IP must
already be in source_nat_active_ips before it can be drained.

- references: GcpAddress (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.rules[].action.sourceNatActiveRanges

`[]string | valueFrom`

Subnetworks whose primary ranges provide the NAT addresses for matching
traffic (private NAT only).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.rules[].action.sourceNatDrainRanges

`[]string | valueFrom`

Subnetwork ranges being drained (private NAT only).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.logFilter

`enum` · optional (explicit presence)

Log filter for NAT translation logging.
**Default:** ERRORS_ONLY (recommended for production to detect port exhaustion and connection failures).
Use DISABLED for non-production environments to reduce costs.
Use ALL for security auditing or detailed troubleshooting (generates significant log volume).

- default: `ERRORS_ONLY`

Allowed values (use exactly as shown):

- `DISABLED` -- Disable logging (not recommended for production)
- `ERRORS_ONLY` -- Log translation errors only (recommended for production)
- `ALL` -- Log all translations (use for security auditing or troubleshooting)
- `TRANSLATIONS_ONLY` -- Log successful translations only (connection-level audit trail without the error noise)

### spec.encryptedInterconnectRouter

`bool`

Dedicate the router to encrypted VLAN attachments (HA VPN over Cloud
Interconnect). An encrypted-Interconnect router carries only
encrypted attachments and cannot be converted later. Immutable after
creation.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the Cloud Router for org-policy and
IAM conditions. Keys in the form "tagKeys/{id}", values
"tagValues/{id}". Create-time only: changing them later replaces the
router.

### spec.deletionPolicy

`string`

Deletion policy for the router and NAT — what happens when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the router and NAT are deleted
  "PREVENT" -- destroy FAILS; protects the egress path of a
               production fleet from accidental teardown
  "ABANDON" -- the router and NAT are removed from management but
               keep serving traffic in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.sourceSubnetworkIpRangesToNat64

`string`

NAT64 scope — which subnetworks get IPv6-to-IPv4 translation.
ALL_IPV6_SUBNETWORKS: every subnetwork's IPv6 ranges are translated
(only ONE NAT per region in the network may claim this).
LIST_OF_IPV6_SUBNETWORKS: only the subnetworks in nat64_subnetworks.
Leave empty for no NAT64 (IPv4-only NAT).

- rule: source_subnetwork_ip_ranges_to_nat64 must be ALL_IPV6_SUBNETWORKS or LIST_OF_IPV6_SUBNETWORKS

### spec.nat64Subnetworks

`[]string | valueFrom`

Subnetworks whose IPv6 traffic the NAT64 gateway translates. Each
entry references a dual-stack GcpSubnetwork by self link. Listing any
subnetwork here selects LIST_OF_IPV6_SUBNETWORKS mode (mirroring how
subnetworks implies LIST_OF_SUBNETWORKS).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

## Validation Rules

- `dynamic_ports_conflicts_with_eim`: enable_dynamic_port_allocation and enable_endpoint_independent_mapping are mutually exclusive
- `max_ports_requires_dynamic_allocation`: max_ports_per_vm is only valid when enable_dynamic_port_allocation is true
- `max_ports_gte_min_ports`: max_ports_per_vm must be greater than or equal to min_ports_per_vm
- `dynamic_ports_must_be_powers_of_two`: min_ports_per_vm and max_ports_per_vm must be powers of two when dynamic port allocation is enabled
- `list_mode_requires_subnetworks`: subnetworks must be listed when source_subnetwork_ip_ranges_to_nat is LIST_OF_SUBNETWORKS
- `subnetworks_require_list_mode`: listing subnetworks requires source_subnetwork_ip_ranges_to_nat LIST_OF_SUBNETWORKS (or leave it empty to imply it)
- `private_nat_has_no_external_ips`: private NAT uses subnetwork ranges, not external IPs — nat_ips, drain_nat_ips, and auto_network_tier must be empty
- `drain_requires_manual_nat_ips`: drain_nat_ips requires nat_ips (an IP must be in the manual set before it can be drained)
- `rule_ips_only_for_public_nat`: private NAT rules use source_nat_active_ranges/source_nat_drain_ranges, not IPs
- `rule_ranges_only_for_private_nat`: public NAT rules use source_nat_active_ips/source_nat_drain_ips, not subnetwork ranges
- `nat64_list_mode_requires_subnetworks`: nat64_subnetworks must be listed when source_subnetwork_ip_ranges_to_nat64 is LIST_OF_IPV6_SUBNETWORKS
- `nat64_subnetworks_require_list_mode`: listing nat64_subnetworks requires source_subnetwork_ip_ranges_to_nat64 LIST_OF_IPV6_SUBNETWORKS (or leave it empty to imply it)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpRouterNat, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Name of the Cloud NAT gateway (as created in GCP). |
| `status.outputs.router_self_link` | `string` | Self-link URL of the Cloud Router carrying this NAT. |
| `status.outputs.nat_ips` | `[]string` | Self links of the static external IPs the NAT translates through. Populated only for manual allocation (nat_ips set in the spec); the literal IP of each entry is the referenced GcpAddress node's `address` output. Empty for auto-allocation, where GCP manages an unlisted pool that can change as capacity scales — use manual allocation when the egress IPs must be stable and allowlistable. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.vpcSelfLink` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetworks[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.natIps` | GcpAddress | `status.outputs.self_link` |
| `spec.drainNatIps` | GcpAddress | `status.outputs.self_link` |
| `spec.rules[].action.sourceNatActiveIps` | GcpAddress | `status.outputs.self_link` |
| `spec.rules[].action.sourceNatDrainIps` | GcpAddress | `status.outputs.self_link` |
| `spec.rules[].action.sourceNatActiveRanges` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.rules[].action.sourceNatDrainRanges` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.nat64Subnetworks` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## See Also

- [Overview](../README.md)

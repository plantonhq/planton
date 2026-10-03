# GcpRouterNat Guide

The judgment this guide protects: NAT failures almost never look like NAT
failures. They look like intermittent timeouts from application code — a
fleet quietly exhausting its NAT ports, or a partner suddenly rejecting an
egress IP that GCP rotated. Configure allocation and ports for the fleet
you will have, not the fleet you have today.

## Auto vs manual allocation is a business decision, not a technical one

Auto-allocation (`natIps` empty) is operationally free: GCP grows and
shrinks the IP pool with demand. Its cost is identity — the IPs can change,
so nothing external may ever depend on them. The moment a partner
allowlists your egress, a compliance record names it, or an API vendor
registers it for rate limits, switch to referenced `GcpAddress`
reservations. The rotation path is the reason the manual mode is safe to
commit to: add the new reservation to `natIps`, move the old one to
`drainNatIps`, and established connections bleed off with zero cut — then
remove it.

## Port exhaustion: the arithmetic worth doing once

Every VM gets `minPortsPerVm` ports (64 by default) per NAT IP for ALL its
outbound connections to the same destination ip:port. A service mesh
calling one upstream API at high concurrency hits 64 ports fast, and the
symptom is dropped SYNs, not an error mentioning NAT. Two levers, one
rule of thumb:

- **Dynamic port allocation** (`enableDynamicPortAllocation` with a
  `maxPortsPerVm` ceiling) for mixed fleets — idle VMs stay cheap, busy
  VMs grow. Powers of two only; incompatible with endpoint-independent
  mapping (the spec enforces both).
- **A raised static floor** when every VM behaves the same and you want
  deterministic capacity math: `(NAT IPs × 64,512) / minPortsPerVm` is
  your VM ceiling per gateway.

Leave `logFilter: ERRORS_ONLY` on in production — port exhaustion appears
there (`allocation_status: DROPPED`) long before anyone correlates the
application symptoms.

## One router, many hats — when to set routerBgp

A NAT-only router never needs `routerBgp`; GCP assigns an ASN and nothing
speaks BGP. Set it when the SAME regional router also terminates Cloud VPN
tunnels or Interconnect attachments: the ASN is fixed for the router's
lifetime (pick the private ASN your network team allocates, never default
into one that collides with on-prem), and `advertiseMode: CUSTOM` with
`advertisedIpRanges` is how you expose a curated set of ranges to peers
instead of every subnet. The BGP peers and interfaces themselves are
separate resources composed on top of this node — this kind owns the
router's own posture, not the sessions.

## Private NAT and NAT64 are different animals

- `type: PRIVATE` is NAT *between* private networks (Network Connectivity
  Center spokes) — no external IPs anywhere; addresses come from
  subnetwork ranges carried in rules (`sourceNatActiveRanges`). The spec
  keeps the public-IP machinery locked out of private NATs pre-deploy.
- NAT64 (`nat64Subnetworks` / `sourceSubnetworkIpRangesToNat64`) is public
  NAT for IPv6-only workloads reaching IPv4 destinations. Only one NAT per
  region in a network may claim `ALL_IPV6_SUBNETWORKS` — scope with the
  list form when a region runs several gateways.

## Conventions and gotchas

- Scoping to subnetworks is also a security boundary: a subnet outside the
  NAT's scope simply has no internet path (pair with Private Google Access
  so Google APIs stay reachable without consuming ports).
- GKE pod ranges: NAT the pods without the nodes (or vice versa) using
  `sourceIpRangesToNat: [LIST_OF_SECONDARY_IP_RANGES]` plus the range
  names — the whole reason per-subnetwork scoping carries range detail.
- `deletionPolicy: PREVENT` belongs on the production egress path: a
  destroyed NAT takes the whole private fleet's connectivity down at once,
  which is the single most disruptive "oops" this kind can express.
- Timeout tuning is real money for high-churn workloads:
  `tcpTimeWaitTimeoutSec` below the 120s default frees ports faster at
  the cost of stricter RFC conformance.

## On the diagram

The NAT node sits between the VPC and the internet, with `GcpAddress`
reservations rendering as first-class dependency nodes under manual
allocation — the allowlisted IPs are visible inventory, not hidden state.

## Pairs well with

- `GcpVpcNetwork` — the network the router attaches to (`vpcSelfLink`).
- `GcpAddress` — stable egress IPs (`natIps`, `drainNatIps`, rule IPs).
- `GcpSubnetwork` — scoped v4 coverage, NAT64 subnetworks, private-NAT
  ranges.
- `GcpGkeCluster` — private clusters compose Cloud NAT on their network
  for outbound pulls; no spec field links them, the network does.

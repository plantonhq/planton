# AWS VPC: The Composable Network Foundation

## What a VPC Is

An AWS Virtual Private Cloud is a logically isolated virtual network in a single
AWS region. It is defined by an IP address space and a handful of network-wide
settings -- tenancy, DNS behavior, and IPv6 enablement. Everything else that
people associate with "a network" -- subnets, route tables, internet gateways,
NAT gateways, endpoints -- are distinct resources that live *inside* a VPC and
have their own lifecycles.

`AwsVpc` models exactly that: the VPC and its address space, nothing more. It is
the root node of an AWS networking topology, and it is deliberately thin.

## Why Thin, and Why Composable

It is tempting to make a "VPC" component that also creates a sensible set of
subnets and gateways in one shot -- it feels convenient. But that convenience
comes at the cost of composability. A subnet has a different lifecycle than the
VPC: you add, resize, and retire subnets over a network's life while the VPC's
CIDR stays fixed. A NAT gateway is an expensive, optional, AZ-scoped resource
with its own Elastic IP. An internet gateway is a singleton attachment. Folding
all of these into the VPC turns independent, individually ownable resources into
hidden fields of one object, where they cannot be referenced, replaced, or
reasoned about on their own.

The composable model keeps each concern a first-class node:

- **`AwsVpc`** — the address space (this component).
- **`AwsSubnet`** — a subnet in one availability zone, with routing folded in
  (inline routes that create a subnet-owned route table, or a reference to an
  existing one).
- **`AwsInternetGateway`** — the VPC's attachment point for public ingress/egress.
- **`AwsNatGateway`** — outbound-only internet access for private subnets,
  composing an `AwsElasticIp` by reference.

Downstream resources reference the VPC by its `vpc_id` output and reference
specific subnets by their `subnet_id` outputs. The architecture graph then
mirrors the real topology: you can see, own, and change each piece
independently, and an agent composing a network from intent can assemble exactly
the pieces it needs.

## The Address Space Model

A VPC's address space is the one part that is genuinely intrinsic to the VPC, so
it lives here.

### Primary IPv4 CIDR

Every VPC has exactly one primary IPv4 CIDR, fixed at creation. There are two
ways to set it:

- **Explicit**: set `cidrBlock` (e.g. `10.0.0.0/16`). The mask must be between
  /16 and /28.
- **IPAM-allocated**: set `ipv4IpamPoolId` and `ipv4NetmaskLength` to let AWS IP
  Address Manager carve a block of the requested size from a managed pool. This
  is how larger organizations avoid CIDR collisions across many VPCs.

These are mutually exclusive modes (you cannot ask IPAM for a netmask *and* pin
an explicit CIDR via the netmask path), and exactly one primary source is
required. Choosing a non-overlapping range matters: VPCs that will ever be peered
or connected through a transit gateway must not share address space.

### Secondary IPv4 CIDRs

When a single CIDR runs out of room, or when a workload needs a distinct range
(such as the carrier-grade 100.64.0.0/10 space for pods in a large Kubernetes
cluster), additional CIDRs can be associated with the VPC. Each entry in
`secondaryIpv4CidrBlocks` becomes its own association resource, so secondary
ranges can be added or removed over time without recreating the VPC.

### IPv6

IPv6 is opt-in and dual-stack (a VPC always has IPv4; IPv6 is layered on):

- **Amazon-provided**: set `assignGeneratedIpv6CidrBlock` to get an Amazon /56,
  optionally advertised from a specific `ipv6CidrBlockNetworkBorderGroup` (for
  Local Zones / Wavelength).
- **IPAM-allocated**: set `ipv6IpamPoolId` with either `ipv6NetmaskLength` (a
  size) or `ipv6CidrBlock` (a specific block).

As with IPv4, the Amazon-provided and IPAM modes are mutually exclusive, and the
component validates the combinations up front so an invalid request fails fast
rather than mid-apply.

## Network-Wide Settings

- **Instance tenancy** (`instanceTenancy`): `default` places instances on shared
  hardware; `dedicated` forces single-tenant hardware for every instance in the
  VPC (rarely needed, and materially more expensive). AWS only supports moving
  `dedicated` back to `default` in place.
- **DNS support** (`enableDnsSupport`): the Amazon-provided DNS resolver inside
  the VPC. AWS enables this by default and most workloads depend on it, so it
  stays on unless you explicitly disable it.
- **DNS hostnames** (`enableDnsHostnames`): whether instances with public IPs
  also get public DNS names. Enable this for VPCs whose instances should be
  reachable by name, and for features like Route 53 private hosted zones.
- **Network Address Usage metrics** (`enableNetworkAddressUsageMetrics`): a
  CloudWatch metric tracking how the VPC's address space is consumed -- useful
  for capacity planning in large networks.

## Validation as Documentation

The address-space rules above are not just prose -- they are encoded as
cross-field validation on the spec, mirroring what AWS itself enforces: a primary
IPv4 source is required; explicit-CIDR and IPAM-netmask modes are mutually
exclusive; IPv6 amazon-provided and IPAM modes are mutually exclusive; an
explicit IPv6 CIDR or netmask requires an IPAM pool; netmask lengths must be in
the legal ranges. Invalid combinations are rejected before any cloud call, which
is what makes the component safe for both humans and agents to configure from the
spec alone.

## Dual-Engine Implementation

`AwsVpc` ships both a Terraform/OpenTofu module and a Pulumi (Go) module at
behavioral parity. Both create the VPC and its secondary-CIDR associations,
apply the same identity tags, and export the same outputs (`vpc_id`, `vpc_arn`,
`cidr_block`, `ipv6_cidr_block`, `owner_id`, the main/default route-table and
default security-group/network-ACL ids, and `region`). Whichever engine a team
standardizes on, the VPC behaves identically.

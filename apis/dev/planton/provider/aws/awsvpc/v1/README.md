# Overview

The AwsVpc API resource provisions an AWS Virtual Private Cloud (VPC): the
isolated virtual network that is the foundation for nearly every other AWS
resource. It is a thin, composable building block -- an IP address space plus a
few network-wide settings -- not a bundle of subnets and gateways.

## Why We Created This API Resource

A VPC is the root of an AWS network topology, but the resources that make a
network useful -- subnets, internet gateways, NAT gateways, route tables -- each
have their own lifecycle and are most powerful when modeled as independent,
referenceable nodes. This resource keeps the VPC itself clean so a topology can
be composed from first-class components rather than hidden inside one opaque
object. It lets you:

- **Define the address space deliberately**: one primary IPv4 CIDR, optional
  secondary IPv4 CIDRs, and optional IPv6 -- explicitly or from an IPAM pool.
- **Compose, don't bundle**: attach `AwsSubnet`, `AwsInternetGateway`,
  `AwsNatGateway`, and other components by reference, each owning its own
  lifecycle.
- **Stay consistent across environments**: the same declarative shape on both
  Terraform and Pulumi.

## Key Features

### Address Space

- **Primary IPv4 CIDR**: the VPC's main range (e.g. `10.0.0.0/16`), specified
  directly or allocated from an IPAM pool.
- **Secondary IPv4 CIDRs**: additional ranges associated with the VPC, added or
  removed without recreating it.
- **IPv6**: an Amazon-provided /56 or an IPAM-allocated block, for dual-stack
  networks.

### Network Settings

- **Instance tenancy**: shared (`default`) or single-tenant (`dedicated`)
  hardware.
- **DNS**: Amazon-provided DNS resolution and public DNS hostnames.
- **Network Address Usage metrics**: optional CloudWatch IP-consumption metrics
  for capacity planning.

## Benefits

- **Composability**: subnets and gateways are first-class components that
  reference the VPC, so the architecture graph reflects the real topology.
- **Feature depth**: secondary CIDRs, IPv6, and IPAM are available from day one,
  not deferred.
- **Consistency**: identical behavior across Terraform and Pulumi.
- **Correctness**: cross-field validation mirrors what AWS itself enforces, so
  invalid IP configurations are caught before deployment.

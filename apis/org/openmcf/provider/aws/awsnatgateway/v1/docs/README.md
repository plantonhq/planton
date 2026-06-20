# AwsNatGateway — Design Rationale and Research

This document explains why the `AwsNatGateway` spec is shaped the way it is and what is intentionally deferred.

## Why the NAT gateway is its own component

A NAT gateway is a real, independently ownable network node: it has its own lifecycle, its own id, its own placement, and it is referenced as a route target by the subnets it serves. Modeling it as a standalone component (rather than something created implicitly inside the VPC) lets a topology be composed from intent — an author or an agent creates exactly the gateways they want, in the subnets they choose, and wires private subnets to them explicitly. It also keeps the VPC thin: the VPC owns address space and DNS, while egress is a separate, composable concern.

## Connectivity type is required and explicit

A NAT gateway is one of two quite different things depending on `connectivity_type`:

- **public** — placed in a public subnet, fronted by an Elastic IP, and used to give private subnets outbound internet access.
- **private** — no Elastic IP, used for outbound access to other private networks (peered/transit/VPN) with no internet exposure.

AWS's own API defaults a new gateway to public, but the field is **required** here so the choice is always explicit in the manifest. Making it required (rather than defaulting to public) keeps the two IaC engines honest: neither the Terraform nor the Pulumi module has to invent a default, so they cannot drift from each other. It is modeled as a CEL-validated string (`in ['public', 'private']`) — matching the convention used elsewhere for small two-way choices (an `AwsSubnet`'s `private_dns_hostname_type_on_launch`), where a larger structural enumeration would instead be a proto enum.

## The Elastic IP is composed, never embedded

A public NAT gateway needs a stable outbound IPv4 address. Rather than allocating an IP inside this component, `allocation_id` references an `AwsElasticIp` (or accepts a literal `eipalloc-` id). This keeps the IP a first-class, independently ownable node with its own lifecycle — the platform-sanctioned way to give a NAT gateway a stable address — and avoids burying an allocation inside another resource where it cannot be reasoned about or reused.

## Cross-field validation mirrors the AWS resource

The AWS resource enforces a set of connectivity-type-dependent rules; the spec encodes the same constraints declaratively via message-level CEL, each with a human-readable message:

- a **public** gateway requires `allocation_id`;
- a **private** gateway forbids `allocation_id` and `secondary_allocation_ids`;
- `private_ip`, `secondary_private_ip_addresses`, and `secondary_private_ip_address_count` are valid only for a **private** gateway;
- the two ways to add secondary private IPs (an explicit list vs. a count) are mutually exclusive.

This means an invalid combination is rejected at validation time, before any cloud call — the kind of dense, self-describing contract a coding agent can configure correctly from the spec alone.

## Placement vs routing

A recurring source of confusion is that creating a NAT gateway grants egress. It does not. The gateway lives in one subnet; the private subnets that need egress must send their default route (`target_type = nat_gateway`) to it. This component therefore exports `nat_gateway_id`, which an `AwsSubnet` route consumes as its `target_id`. A public gateway additionally depends on its host subnet being public (routing to an internet gateway); that is a property of the subnet, not of this component.

## Outputs: no ARN

A NAT gateway has no ARN — neither the AWS resource nor the underlying API exposes one. The stack outputs therefore identify the gateway solely by `nat_gateway_id`, alongside its observed addresses (`public_ip`, `private_ip`), its `network_interface_id`, and the echoed `subnet_id`/`region`.

## What's intentionally deferred

Each item is additive and can land later without breaking the schema:

- **Regional / multi-AZ NAT availability mode** — AWS recently introduced a regional NAT gateway (`availability_mode`, `availability_zone_address`, auto-provisioned zones) that spans availability zones. It is intentionally not surfaced: the Pulumi AWS provider does not yet expose these inputs, so adding them would break the rule that the Terraform and Pulumi modules stay at full behavioral parity. When both engines support it, it can be added as new optional fields. The established high-availability pattern today — one zonal NAT gateway per AZ — is fully supported.
- **Live E2E for the inline-routing path** — a private subnet routing `0.0.0.0/0` to this gateway proves the full egress recipe end to end. Exercising it additionally requires the test harness to support per-scenario prerequisites; that framework enhancement is tracked separately. The gateway's own create/verify/destroy lifecycle is covered by the live E2E.

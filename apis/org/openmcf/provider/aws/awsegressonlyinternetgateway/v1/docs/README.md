# AwsEgressOnlyInternetGateway — Design Rationale and Research

This document explains why the `AwsEgressOnlyInternetGateway` spec is shaped the way it is and what is intentionally deferred.

## Why the egress-only gateway is its own component

An egress-only internet gateway is a real, independently ownable network node: it has its own lifecycle, its own id, and it is referenced as an IPv6 route target by subnets. Modeling it as a standalone component (rather than something created implicitly inside the VPC) lets a topology be composed from intent — an author or an agent creates exactly the gateway they want, attaches it to a specific dual-stack VPC, and wires private subnets' IPv6 egress to it explicitly. It keeps the VPC thin: the VPC owns address space and DNS, while IPv6 egress is a separate, composable concern.

## The spec is deliberately minimal

An egress-only internet gateway has essentially one configuration axis: which VPC it attaches to. The authoritative AWS surface (`aws_egress_only_internet_gateway`) exposes only `vpc_id` plus tags; even the `arn` and `owner_id` that an internet gateway computes do not exist here — the only computed attribute is the id. Designing to the 90/10 standard means resisting the urge to invent knobs that do not exist. The spec therefore carries `region` (to drive provider construction and as an echoed output) and `vpc_id` (the attachment), and nothing more. There is no CEL rule because there are no cross-field constraints to enforce.

## Attachment semantics: required and immutable

`vpc_id` is **required**: in a declarative graph an unattached gateway is dead weight, and the composition story is always "attach this gateway to that VPC." It is supplied as a `StringValueOrRef` with `default_kind = AwsVpc` so it can be a literal vpc-id (brownfield) or a reference resolved from an `AwsVpc`'s outputs (greenfield).

Unlike an internet gateway's attachment (which AWS lets you detach and re-attach), an egress-only gateway's attachment is **immutable**: AWS creates it bound to a VPC and offers no detach API, so changing `vpc_id` replaces the gateway (ForceNew). The field documentation states this explicitly so the behavior is unambiguous to a reader or an agent.

## IPv6-only, outbound-only — why there is no NAT

Every IPv6 address AWS assigns is globally routable, so there is no IPv6 network address translation. The egress-only internet gateway is the mechanism that provides the "outbound-but-not-inbound" guarantee for IPv6 that a NAT gateway provides for IPv4: it statefully allows return traffic for connections initiated from inside the VPC and drops unsolicited inbound. It also carries no per-hour or per-GB charge, unlike a NAT gateway.

## Attachment is not routing

Attaching the gateway does not, by itself, give a subnet IPv6 egress. Connectivity requires a subnet whose route table sends the IPv6 default route (`::/0`) to the gateway. This component therefore exports `egress_only_internet_gateway_id`, which an `AwsSubnet` route consumes as its `target_id` with `target_type = egress_only_internet_gateway`. The two components together express the private-IPv6-egress recipe; neither does it alone.

## Relationship to sibling gateways

- **Internet gateway** — bidirectional IPv4/IPv6. Use it for public subnets that must be reachable from the internet.
- **NAT gateway** — the IPv4 equivalent of this gateway: outbound-only access for private IPv4 subnets, living in a public subnet that routes to an internet gateway (and billed per hour and per GB).

## What's intentionally deferred

Each item is additive and can land later without breaking the schema:

- **Detached creation** — AWS attaches the gateway at creation; there is no dangling-gateway workflow to model, so `vpc_id` is required.
- **Live E2E for the inline-routing path** — a subnet routing `::/0` to this gateway proves the full private-IPv6-egress recipe end to end. Exercising it additionally requires the test harness to support per-scenario prerequisites (today prerequisites are resolved per-kind from the registry); that framework enhancement is tracked separately. The gateway's own create/attach/destroy lifecycle is covered by live E2E (it attaches cleanly to the `AwsVpc` prerequisite, verified with `DescribeEgressOnlyInternetGateways`).

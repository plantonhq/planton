# AwsInternetGateway — Design Rationale and Research

This document explains why the `AwsInternetGateway` spec is shaped the way it is and what is intentionally deferred.

## Why the internet gateway is its own component

An internet gateway is a real, independently ownable network node: it has its own lifecycle, its own id, and it is referenced as a route target by subnets. Modeling it as a standalone component (rather than something created implicitly inside the VPC) lets a topology be composed from intent — an author or an agent creates exactly the gateway they want, attaches it to a specific VPC, and wires public subnets to it explicitly. It also keeps the VPC thin: the VPC owns address space and DNS, while internet connectivity is a separate, composable concern.

## The spec is deliberately minimal

An internet gateway has essentially one configuration axis: which VPC it attaches to. The authoritative AWS surface (`aws_internet_gateway`) exposes only `vpc_id` plus tags; everything else (`arn`, `owner_id`) is computed. Designing to the 90/10 standard here means resisting the urge to invent knobs that do not exist. The spec therefore carries `region` (to drive provider construction and as an echoed output) and `vpc_id` (the attachment), and nothing more. There is no CEL rule because there are no cross-field constraints to enforce.

## Attachment semantics: required and updatable

`vpc_id` is **required**: in a declarative graph an unattached gateway is dead weight, and the composition story is always "attach this gateway to that VPC." It is supplied as a `StringValueOrRef` with `default_kind = AwsVpc` so it can be a literal vpc-id (brownfield) or a reference resolved from an `AwsVpc`'s outputs (greenfield).

Unlike a subnet's `vpc_id` (which is ForceNew), an internet gateway's attachment is **updatable**: AWS supports detaching a gateway from one VPC and attaching it to another without replacing the gateway. The field documentation states this explicitly so the behavior is unambiguous to a reader or an agent. A VPC can have at most one internet gateway attached at a time.

## Attachment is not exposure

A recurring source of confusion is that attaching an internet gateway makes a VPC "public." It does not. Connectivity requires a subnet whose route table sends a default route to the gateway. This component therefore exports `internet_gateway_id`, which an `AwsSubnet` route consumes as its `target_id` with `target_type = internet_gateway`. The two components together express the public-subnet recipe; neither does it alone.

## Relationship to sibling gateways

- **Egress-only internet gateway** — for IPv6-only outbound access with no inbound exposure. A separate, additive kind (a different AWS resource); not folded in here.
- **NAT gateway** — gives private-subnet instances outbound-only access. A NAT gateway lives in a public subnet that itself routes to an internet gateway, so the internet gateway is a prerequisite for the NAT egress topology, not a replacement for it.

## What's intentionally deferred

Each item is additive and can land later without breaking the schema:

- **Detached creation** — AWS allows creating a gateway without attaching it and attaching later. The declarative model has no use for a dangling gateway, so `vpc_id` is required; if a genuine attach-later workflow emerges, `vpc_id` could become optional without a breaking change to consumers.
- **Live E2E for the gateway's own create/attach/destroy lifecycle** — a live attach needs a VPC that does not already have an internet gateway. Until `AwsVpc` stops bundling its own internet gateway (a VPC may have only one at a time), the live test cannot stand up a suitable prerequisite. The verifier, `e2e/profile.yaml`, and scenario ship ready; the live run is enabled once a gateway-free VPC is available. The component is otherwise fully validated offline (spec validation, output conformance and cross-engine parity, `tofu validate`).
- **Live E2E for the inline-routing path** — a subnet routing `0.0.0.0/0` to this gateway proves the full public-subnet recipe end to end. Exercising it additionally requires the test harness to support per-scenario prerequisites (today prerequisites are resolved per-kind from the registry); that framework enhancement is tracked separately.

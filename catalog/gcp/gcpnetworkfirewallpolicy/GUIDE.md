# GcpNetworkFirewallPolicy Guide

The judgment this guide protects: a network firewall policy is the ONE
ordered rule set for a network, written where the network's owners can read
it whole. Prefer it to a pile of legacy `GcpFirewallRule`s, and keep global
and regional policies for what each is for.

## Global or regional

Google models the two scopes as separate resource families with identical
rules and associations; this kind is one shape and `region` picks the
family. Empty `region` is a GLOBAL policy: it governs all of an attached
network's traffic in every region, and is the everyday choice. A region is a
REGIONAL policy: it governs only that region's traffic, and it is the only
scope at which two things exist -- rules whose target is an internal managed
load balancer (`targetType: INTERNAL_MANAGED_LB` with
`targetForwardingRules`), and the RDMA and ultra-low-latency policy types
(`policyType`). The scope is immutable: a policy cannot move between
families or regions.

## Where it sits in the evaluation order

Hierarchical policies (organization, then folders) decide first; their
`goto_next` is what lets this policy see the packet at all. Then the
network's global network firewall policy, then its regional one, then legacy
VPC firewall rules -- unless the network's enforcement order has been
flipped to put legacy rules before network policies. `allow` and `deny` end
the decision; `goto_next` passes it along.

## Priority is identity

Google identifies a rule by its priority, and this kind keeps that truth: a
rule whose priority changes is destroyed and recreated at the new number
while every other rule is untouched. `ruleName` is a free label -- rename it
at will; renumber deliberately. Leave gaps (1000, 2000, ...) so a rule can
be slotted between two others later.

## One policy per network per scope

A network carries at most one global and, per region, one regional network
firewall policy association at a time. Associating a second fails until the
first is detached. To swap, detach first -- or let a chart destroy the old
policy before the new association is made. `policyType` must match the
network's profile: `VPC_POLICY` (the default when unset) for ordinary VPCs;
the RDMA and ULL types only attach to networks created with those profiles.

## Direction decides which match fields mean anything

On an `INGRESS` rule the `src*` fields describe where traffic comes from and
`dest*` may narrow the addresses inside; on `EGRESS` the roles flip. Secure
tags and networks exist only on the VPC side, so `srcSecureTags` and
`srcNetworks` carry meaning on ingress rules only. `srcNetworkContext` /
`destNetworkContext` express the same idea as a class (`INTERNET`,
`INTRA_VPC`, ...) without listing addresses.

## Targets narrow the rule

A rule applies to every VM on the attached networks unless it names targets:
`targetSecureTags` (VMs carrying one of the tag values) or
`targetServiceAccounts` (VMs running as one of the accounts). On a regional
policy, `targetType: INTERNAL_MANAGED_LB` with `targetForwardingRules`
applies the rule to internal Application Load Balancers instead of VMs. A
rule whose target tags are all INEFFECTIVE (deleted value, deleted network)
is silently ignored; log the rule and watch its logs.

## Everything the catalog can make is a reference

The project, VPC networks, tag values, service accounts, and forwarding
rules are catalog kinds, so every field that names one takes a `valueFrom`
reference (or a literal when the thing lives outside the chart). Address
groups, security profile groups, and Google's threat-intelligence lists are
literals: the first two are not yet catalog kinds and the lists are Google's.

## What is immutable

`policyName`, `region`, `policyType`, and `projectId` recreate the policy --
and with it every rule and association. A rule's `priority` recreates that
rule. An association's `name` and `network` recreate that association.
Everything else changes in place.

## Destroying

`deletionPolicy` fans to the policy, every rule, and every association.
`DELETE` (the default) detaches, removes, and deletes; traffic then falls
through to the next level. `PREVENT` is the guard for a network's baseline
deny. `ABANDON` leaves everything live and enforcing but unmanaged.

## Cost

Free. Firewall policies have no meter; `enableLogging` bills as Cloud
Logging ingestion on the project.

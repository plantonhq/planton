# GcpHierarchicalFirewallPolicy Guide

The judgment this guide protects: a hierarchical policy is the organization's
word on what may never happen, plus a `goto_next` that leaves everything else
to the people running the projects. Write it narrow and let it delegate.

## Where it sits in the evaluation order

Google evaluates a packet against the hierarchical policies from the
organization down through the folders (within each, lowest priority number
first), then the network's global network firewall policy, then the regional
one, then legacy VPC firewall rules. `allow` and `deny` end the decision;
`goto_next` passes it to the next level. An organization-level `deny` can
therefore never be undone by a project, and an organization-level `allow`
cannot be tightened below it -- which is why the everyday shape is a short
list of denies (and allows for Google's own ranges) followed by `goto_next`.
Google appends two implied `goto_next` rules at 2147483646 and 2147483647 to
every hierarchical policy; the spec keeps user priorities below them.

## Parent versus associations

`parent` is where the policy lives -- the node whose IAM governs who may edit
it and whose quota it counts against. `associations` is where it is
enforced. Usually they are the same node, but Google allows one policy to be
enforced on several folders, so a shared "regulated workloads" rule set can
be written once and attached to three branches of the tree. A folder or the
organization carries one hierarchical policy association at a time: to swap
policies, detach the old one first (or declare the new policy and let the
chart destroy the old one before the new association is made).

## Priority is identity

Google identifies a rule by its priority. This kind keeps that truth instead
of hiding it: a rule whose priority changes is destroyed and recreated at the
new number, while every other rule is untouched. Change a rule's content in
place freely; renumber deliberately. Leave gaps (1000, 2000, ...) so a rule
can be slotted between two others later without moving either.

## Direction decides which match fields mean anything

On an `INGRESS` rule the `src*` fields describe where traffic comes from
(the internet, a country, a threat list, another VPC, a tagged VM) and
`dest*` may narrow which addresses inside are reachable. On an `EGRESS` rule
the roles flip: `dest*` describes the outside and `src*` the VMs inside.
Secure tags and networks exist only on the VPC side, so `srcSecureTags` and
`srcNetworks` carry meaning on ingress rules only -- Google ignores them on
egress. `srcNetworkContext` / `destNetworkContext` express the same idea as a
class (`INTERNET`, `INTRA_VPC`, ...) without listing addresses.

## Targets narrow the rule

A rule applies to every VM beneath the association unless it names targets:
`targetResources` (specific VPC networks), `targetSecureTags` (VMs carrying
one of the tag values), or `targetServiceAccounts` (VMs running as one of
the accounts). Tags and service accounts cannot be combined on one rule --
Google selects by one or the other. A rule whose target tags are all
INEFFECTIVE (deleted value, deleted network) is silently ignored, which is
the one way a policy can stop enforcing without anyone changing it; log the
rule and watch its logs.

## Everything the catalog can make is a reference

Folders, VPC networks, tag values, and service accounts are all catalog
kinds, so every field that names one takes a `valueFrom` reference (or a
literal when the thing lives outside the chart). Address groups, security
profile groups, and Google's threat-intelligence lists are named as literals:
the first two are not yet catalog kinds and the lists are Google's own.

## What is immutable

`parent` and `shortName` recreate the policy -- and with it every rule and
association. A rule's `priority` recreates that rule. An association's
`name` and `target` recreate that association. Everything else changes in
place.

## Destroying

`deletionPolicy` fans to the policy, every rule, and every association.
`DELETE` (the default) detaches, removes, and deletes; traffic then falls
through to the next level as if the policy had never existed. `PREVENT` is
the guard for an organization's baseline. `ABANDON` leaves everything live
and enforcing but unmanaged.

## Cost

Free. Firewall policies have no meter; `enableLogging` bills as Cloud Logging
ingestion on the projects whose VMs the rule decides for.

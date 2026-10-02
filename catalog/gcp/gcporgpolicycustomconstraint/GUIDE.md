# GcpOrgPolicyCustomConstraint Guide

The judgment this guide protects: a custom constraint is a definition the
whole organization shares, so it is written once, named for what it
forbids, and enforced by policies -- never duplicated into one of them.

## Definition here, enforcement in GcpOrgPolicy

Creating a constraint blocks nothing. It becomes a guardrail when a
`GcpOrgPolicy` at some scope references it (`customConstraint` pointing at
this resource's `constraint` output, the `custom.<name>` handle). That
separation is deliberate: the constraint outlives any one policy, three
folders can enforce it with three small policies, and deleting one of
those policies leaves the rule intact for the other two.

## Write the condition over the resource Google shows you

The condition is CEL over the REST resource's fields as the service
exposes them to the constraint engine: `resource.management.autoUpgrade
== false` for a GKE node pool, `resource.settings.ipConfiguration.
ipv4Enabled == true` for a Cloud SQL instance. Google's per-service
"supported services" pages list the resource types, the visible fields,
and the methods each service evaluates. A constraint over an unsupported
field is rejected at apply time, never silently ignored.

## DENY describes the forbidden shape; ALLOW describes the only shape

Most guardrails are `DENY`: the condition says what must not happen. An
`ALLOW` constraint inverts that -- the condition describes the only
acceptable configuration and everything else is implicitly denied. Reach
for `ALLOW` when the allowed set is small and the forbidden set is
open-ended (only Shielded VMs; only these machine families).

## Write the description for the person it blocks

The description is the error message a blocked engineer reads. "Node
pools must enable auto-upgrade; set management.autoUpgrade to true" gets
them unblocked; "policy violation" does not.

## The name is permanent, the rule is not

The name (with Google's `custom.` prefix, which the module adds), the
organization, and the resource types are immutable -- changing any is a
delete and a create, and every policy enforcing the old name lapses in
between. The condition, action, methods, display name, and description
update in place, so tightening a live constraint is safe. Name the
constraint for the outcome (`requireShieldedVm`), not for today's
condition.

## Destroy order

A constraint that policies still enforce cannot be safely removed --
those policies start failing to apply. A chart whose policies reference
the constraint destroys them first; a hand-managed estate should do the
same. `PREVENT` is the right deletion policy for a constraint many
folders rely on.

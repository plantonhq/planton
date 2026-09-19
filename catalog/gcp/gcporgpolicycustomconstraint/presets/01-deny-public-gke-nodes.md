# Deny Public GKE Nodes

A `DENY` constraint over GKE node pools: any create or update that leaves
private nodes off is blocked. The organization's own rule, defined once;
enforce it on a folder or project with a `GcpOrgPolicy` whose
`customConstraint` references this resource.

## What it configures

- `resourceTypes: [container.googleapis.com/NodePool]` — the REST resource
  the condition reads.
- `condition` — CEL over the node pool's fields; true means "forbidden".
- `actionType: DENY` — block when the condition is true.
- `description` — the message the blocked engineer reads.

## Adjust before deploying

- **`organizationId`** — yours.
- **`constraintName`** — the bare name (no `custom.`); it becomes
  `custom.denyPublicGkeNodes`, the handle policies reference.

## When to choose something else

An allow-list (only these shapes are acceptable) takes the **Require
Shielded VM** preset.

# Team Scope with Clusters

## Use Case

Give a team a namespace and admin access on a specific cluster that joined the fleet at creation.

## When to Use

- A team that runs on named clusters rather than every cluster in the fleet
- Clusters created with `GcpGkeCluster.fleetProject`, bound through their `fleet_membership` output

## What This Creates

- The Fleet API on the fleet host project
- A team scope with one namespace, one role binding, and one cluster binding

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `membershipBindings[].membership` | a `GcpGkeCluster` reference | Another cluster, or a `GcpGkeFleetMembership` reference for an explicitly registered one. |
| `rbacRoleBindings[].role.predefinedRole` | `ADMIN` | `EDIT` or `VIEW`. |

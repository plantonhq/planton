# Team Scope

## Use Case

Give a team its slice of the fleet: two namespaces on every cluster bound to the scope, labeled with the team, and edit access for the team's Google group.

## When to Use

- Onboarding a team onto a shared fleet
- Declaring a team's namespaces and access once instead of per cluster

## What This Creates

- The Fleet API on the fleet host project
- A team scope with two fleet namespaces and one role binding

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scopeId` | `orders` | The team's name. |
| `namespaces` | two namespaces | The team's namespaces (Google reserves system names such as `kube-system`). |
| `rbacRoleBindings[].group` | `orders-devs@example.com` | The team's group, literal or a `GcpCloudIdentityGroup` reference. |
| `rbacRoleBindings[].role.predefinedRole` | `EDIT` | `ADMIN` or `VIEW`. |

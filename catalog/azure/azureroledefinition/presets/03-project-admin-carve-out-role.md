# Broad Grant with Carve-Outs and Assignable-Scope Control

This preset is the classic "almost-Contributor" role: a wildcard grant with
`notActions` trimming away the authorization plane. Project admins manage
every resource in their environments but can never change RBAC, policy, or
locks -- closing the privilege-escalation path where a Contributor-like role
grants itself more access.

It also demonstrates `assignableScopes` as a governance control: the
definition lives at subscription scope (so multiple resource groups can share
one definition), but Azure will refuse any assignment of it outside the two
listed project groups. The role's blast radius is pre-authorized at
definition time, independent of who holds assignment rights later.

Remember `notActions` is a carve-out from THIS role's grant, not a deny: if
the same principal holds another role granting RBAC writes, that other grant
still applies. Hard denies are Azure deny assignments, a separate mechanism.

## When to Use

- Project/environment admin roles that must not touch the authorization plane
- Any wildcard-based role -- always pair `*` with explicit carve-outs
- Roles whose assignment surface must be constrained ahead of time

## Key Configuration Choices

- **Assignable scopes by reference** -- in an Infra Chart, replace the
  literals with `valueFrom: { name: project-rg }` blocks; each element
  resolves through the `AzureResourceGroup` default kind
- **One management group max** -- Azure allows at most one management group
  in `assignableScopes`
- **Audit the carve-out list** -- `az provider operation show --namespace
  Microsoft.Authorization` enumerates the operations being excluded

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<org-prefix>` | Your organization's role-name prefix (names are tenant-unique) | Your naming convention |
| `<subscription-arm-id>` | `/subscriptions/{subscription-id}` | `az account show --query id` (prepend `/subscriptions/`) |
| `<project-resource-group-arm-id>` | ARM ID of the production project resource group | `az group show --name <rg> --query id` |
| `<project-staging-resource-group-arm-id>` | ARM ID of the staging project resource group | `az group show --name <rg> --query id` |

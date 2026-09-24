# Administrator with Roles and Permissions

## Pattern

A user whose standing is written down completely: the roles it holds and the one-off API permissions it has beside them, every one by reference to the role or resource server that defines it. The manifest is the whole statement of what the account may do.

## What It Does

- Creates the user in the referenced database connection with a minted password (no `password` declared).
- Assigns the referenced role authoritatively -- a role removed from the manifest is removed from the user on the next apply.
- Grants one direct permission (`read:audit-log`) on the referenced resource server, authoritatively in the same way.
- Records a team assignment in `appMetadata`, which the application reads at sign-in and the user cannot change.

## When to Use

- An operator account whose access is reviewed by reading one file.
- A role covers most of what the account needs and a single extra scope covers the rest; use a role instead when the extras grow.
- The role and the API are themselves declared as `Auth0Role` and `Auth0ResourceServer`, so nothing is retyped.

## Customization

- Replace the `Auth0Role` and `Auth0ResourceServer` references with your own declarations, or use literal `value: rol_...` and `value: https://...` for objects managed outside Planton.
- Remove `permissions` to let roles carry every grant; remove `roles` for a user whose standing is a handful of direct scopes.
- Change `appMetadata` to whatever your application reads; `userMetadata` is the sibling the user may edit themselves.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.connectionName.valueFrom.name` | The name of your `Auth0Connection` database connection |
| `spec.email` | The administrator's address |
| `spec.roles[].valueFrom.name` | The name of your `Auth0Role` |
| `spec.permissions[].resourceServerIdentifier.valueFrom.name` | The name of your `Auth0ResourceServer` |
| `spec.permissions[].name` | A scope defined on that resource server |

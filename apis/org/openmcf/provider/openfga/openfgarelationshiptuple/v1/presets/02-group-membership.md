# Group Membership Tuple

This preset adds a user to a group, enabling group-based access control. When the authorization model grants `group#member` access to resources, all members of the group inherit those permissions. This is the building block for RBAC systems where permissions are assigned to groups rather than individual users.

## When to Use

- Adding users to groups for inherited permissions
- RBAC systems where roles/groups are the primary permission mechanism
- Organizations where team membership drives resource access

## Key Configuration Choices

- **Structured user/object** -- proto-correct nested format enabling cross-resource references
- **member relation** (`relation: member`) -- standard group membership relation; must be defined in the authorization model
- **Group as object** -- the group itself is the object being related to; once a user is a member, any resource granting access to `group#member` becomes accessible

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<store-id>` | ID of the target OpenFgaStore | `OpenFgaStore` status outputs |
| `<user-id>` | User identifier (e.g., `anne`) | Your identity system |
| `<group-id>` | Group identifier (e.g., `engineering`, `admin-team`) | Your application |

## Related Presets

- **01-user-document-access** -- Use instead for direct user-to-resource permission grants without groups

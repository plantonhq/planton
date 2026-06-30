# Auth0 Role - Permissions

## Management API Scopes

Auth0 Role resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:roles` | List and retrieve roles and their permissions |
| Create | `create:roles` | Create new roles |
| Update | `update:roles` | Modify role name, description, and permission assignments |
| Delete | `delete:roles` | Remove roles from the tenant |

## Permission Assignment Scopes

Setting a role's permissions reads the resource servers that own the referenced scopes. The following scope is required in addition to the role scopes above:

| Operation | Required Scope | Description |
|-----------|---------------|-------------|
| Resolve scopes for assignment | `read:resource_servers` | Read resource servers to validate the scopes assigned to the role |

Managing the permission assignments themselves (add/remove permissions on a role) is covered by `update:roles`.

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete) including permission assignment, the minimum required scopes are:

```
read:roles create:roles update:roles delete:roles read:resource_servers
```

## Prerequisite Scopes Must Exist

The scopes referenced by a role's permissions must already be defined on their resource servers before they can be assigned. This component does not create scopes — use the `Auth0ResourceServer` component (or define scopes directly in Auth0) first. The M2M application does not need write access to resource servers to assign existing scopes to a role; `read:resource_servers` is sufficient.

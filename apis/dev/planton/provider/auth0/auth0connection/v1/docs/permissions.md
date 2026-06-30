# Auth0 Connection - Permissions

## Management API Scopes

Auth0 connection resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:connections` | List and retrieve connection configurations |
| Create | `create:connections` | Create new connections (Database, Social, Enterprise) |
| Update | `update:connections` | Modify connection settings, enabled clients, and options |
| Delete | `delete:connections` | Remove connections from the tenant |

## Additional Scopes

Depending on the connection type and operations performed, these supplementary scopes may be required:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read users | `read:users` | List users associated with a connection |
| Delete users | `delete:users` | Remove users from a database connection |
| Read stats | `read:stats` | Retrieve connection usage statistics |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:connections create:connections update:connections delete:connections
```

## Connection-Client Association

Connections are enabled for specific clients via the `enabled_clients` property on the connection object. Modifying this association requires `update:connections` scope. No additional client-specific scopes are needed for this operation.

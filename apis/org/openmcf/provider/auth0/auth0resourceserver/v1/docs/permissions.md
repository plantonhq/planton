# Auth0 Resource Server - Permissions

## Management API Scopes

Auth0 resource server resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:resource_servers` | List and retrieve API definitions and their scopes |
| Create | `create:resource_servers` | Create new APIs with identifier, scopes, and signing config |
| Update | `update:resource_servers` | Modify API settings, scopes, and token configuration |
| Delete | `delete:resource_servers` | Remove APIs from the tenant |

## Additional Scopes

Managing client access to resource servers requires client grant scopes:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read grants | `read:client_grants` | List which clients can access this API |
| Create grants | `create:client_grants` | Grant a client access to this API with specific scopes |
| Update grants | `update:client_grants` | Modify the scopes a client has for this API |
| Delete grants | `delete:client_grants` | Revoke a client's access to this API |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:resource_servers create:resource_servers update:resource_servers delete:resource_servers
```

If the automation also manages which clients can access the API, add:

```
read:client_grants create:client_grants update:client_grants delete:client_grants
```

# Auth0 Client - Permissions

## Management API Scopes

Auth0 client resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:clients` | List and retrieve application configurations |
| Create | `create:clients` | Create new applications (SPA, Web, M2M, Native) |
| Update | `update:clients` | Modify application settings, callbacks, and grants |
| Delete | `delete:clients` | Remove applications from the tenant |

## Additional Scopes

Depending on the client configuration, these supplementary scopes may also be required:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read grants | `read:client_grants` | List client grant associations |
| Create grants | `create:client_grants` | Associate clients with APIs |
| Update grants | `update:client_grants` | Modify granted scopes for a client-API pair |
| Delete grants | `delete:client_grants` | Remove client grant associations |
| Read keys | `read:client_keys` | Retrieve client signing credentials |
| Update keys | `update:client_keys` | Rotate client signing credentials |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:clients create:clients update:clients delete:clients
```

If the automation also manages which APIs a client can access, add:

```
read:client_grants create:client_grants update:client_grants delete:client_grants
```

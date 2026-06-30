# Auth0 Event Stream - Permissions

## Management API Scopes

Auth0 event stream (log stream) resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:log_streams` | List and retrieve event stream configurations |
| Create | `create:log_streams` | Create new event streams (webhook, EventBridge, etc.) |
| Update | `update:log_streams` | Modify stream destination, filters, and credentials |
| Delete | `delete:log_streams` | Remove event streams from the tenant |

## Related Scopes

Event streams deliver tenant log data. These additional scopes relate to the underlying log data:

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read logs | `read:logs` | Query tenant logs directly via the Management API |
| Read log users | `read:logs_users` | Access user-specific log entries |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, delete), the minimum required scopes are:

```
read:log_streams create:log_streams update:log_streams delete:log_streams
```

The `read:logs` and `read:logs_users` scopes are not required for managing event streams but may be useful for verifying that events are being generated correctly.

## Credential Sensitivity

Event stream configurations contain destination credentials (webhook tokens, API keys). The `read:log_streams` scope exposes these credentials in API responses. Grant this scope only to trusted automation principals.

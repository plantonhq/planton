# Auth0 Action - Permissions

## Management API Scopes

Auth0 Action resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:actions` | List and retrieve Action definitions and their code |
| Create | `create:actions` | Create new Actions with code, triggers, and secrets |
| Update | `update:actions` | Modify Action code, dependencies, and secrets |
| Delete | `delete:actions` | Remove Actions from the tenant |

## Deployment and Trigger Scopes

Managing Action deployments and trigger bindings requires the same core scopes. The following operations are covered by the scopes above:

| Operation | Required Scope | Description |
|-----------|---------------|-------------|
| Deploy Action | `update:actions` | Deploy a draft Action version to the pipeline |
| List versions | `read:actions` | Retrieve Action version history |
| Get trigger bindings | `read:actions` | List Actions bound to a specific trigger |
| Update trigger bindings | `update:actions` | Reorder or change Actions bound to a trigger |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, deploy, delete), the minimum required scopes are:

```
read:actions create:actions update:actions delete:actions
```

## Secrets in Actions

Action secrets (environment variables) are managed through the Action update endpoint. No additional scopes are required beyond `update:actions` to set or modify Action secrets.

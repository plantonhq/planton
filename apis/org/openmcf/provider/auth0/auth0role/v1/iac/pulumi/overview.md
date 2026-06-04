# Auth0Role Pulumi Module — Architecture Overview

## Resource Flow

```
Auth0RoleStackInput
  ├── target (Auth0Role manifest)
  │     ├── metadata.name → Pulumi resource name (and default role name)
  │     └── spec
  │           ├── name → Role display name (optional; defaults to metadata.name)
  │           ├── description → Role description (optional)
  │           └── permissions → RolePermissionsPermissionArray (conditional)
  └── provider_config → Auth0 Provider credentials
```

## Module Structure

| File | Purpose |
|---|---|
| `main.go` | Entry point: provider setup → role → permissions → outputs |
| `locals.go` | Extract spec fields into `Locals` struct |
| `role.go` | Create `auth0.Role` and conditionally `auth0.RolePermissions` |
| `outputs.go` | Export stack outputs (id, name, description) |

## Resource Creation Sequence

1. **Auth0 Provider** — Configured with domain/client_id/client_secret from `provider_config`, or falls back to environment variables.
2. **auth0.Role** — Created with the resolved name and optional description.
3. **auth0.RolePermissions** (conditional) — When `permissions` is non-empty, sets the role's authoritative permission list. Depends on the role resource. Each permission references a scope by name and the owning resource server's identifier (audience).

## Outputs

| Output | Source |
|---|---|
| `id` | `role.ID()` |
| `name` | `role.Name` |
| `description` | `role.Description` |

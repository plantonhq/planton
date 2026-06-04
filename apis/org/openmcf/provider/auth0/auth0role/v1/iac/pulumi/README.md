# Auth0Role — Pulumi Module

Pulumi Go module that creates and manages an Auth0 Role and its permissions.

## What It Creates

- `auth0.Role` — The role resource with name and description.
- `auth0.RolePermissions` (conditional) — Sets the role's complete permission list when `permissions` is non-empty. This is the authoritative set: a permission omitted on a later apply is removed from the role.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret)
- The permission scopes referenced by `permissions` must already exist on their resource servers (e.g., created via the `Auth0ResourceServer` component).

## Local Testing

```bash
# Install Pulumi Auth0 plugin
make install-pulumi-plugins

# Login to local state
pulumi login --local

# Preview with test manifest
make test

# Or use debug script
./debug.sh ../hack/manifest.yaml
```

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Architecture

See [overview.md](overview.md) for the resource flow and module structure.

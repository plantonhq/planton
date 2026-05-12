# Auth0Action — Pulumi Module

Pulumi Go module that creates and manages Auth0 Actions with optional trigger binding.

## What It Creates

- `auth0.Action` — The action resource with code, trigger, runtime, dependencies, and secrets.
- `auth0.TriggerAction` (conditional) — Binds the action to its trigger when `trigger_binding` is set.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret)

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

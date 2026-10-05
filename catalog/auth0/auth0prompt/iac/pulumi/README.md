# Auth0Prompt — Pulumi Module

Pulumi Go module that manages the login-flow settings of an existing Auth0 tenant.

## What It Creates

- `auth0.Prompt` — the tenant's one set of prompt settings: `universalLoginExperience`, `identifierFirst` and `webauthnPlatformFirstFactor`, each sent only when the spec sets it; a setting the spec leaves out keeps the tenant's value. The resource's delete is the provider's no-op, so the last-applied values stay in place.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:prompts` and `update:prompts`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps each unset field to nil (never sent), the twin of `iac/tf/locals.tf`.
- `module/prompt.go` applies the settings and `module/outputs.go` exports them as applied.

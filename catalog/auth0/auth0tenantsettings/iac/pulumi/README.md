# Auth0TenantSettings — Pulumi Module

Pulumi Go module that manages the presentation settings of an existing Auth0 tenant.

## What It Creates

- `auth0.Tenant` — the settings of the tenant the provider's credential belongs to. It creates no tenant: the Management API cannot. Only `friendlyName`, `pictureUrl`, `supportEmail` and `supportUrl` are set, each only when the spec sets it; every other tenant setting is left as it is. The resource's delete is the provider's no-op, so the last-applied values stay in place.
- `auth0.CustomDomainDefault` — the tenant's default domain, declared only when the spec sets `default_custom_domain`. Its delete only forgets the default (Auth0 has no way to unset one).

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:tenant_settings` and `update:tenant_settings`.

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps each unset field to nil (never sent), the twin of `iac/tf/locals.tf`.
- `module/tenant.go` applies the settings, `module/custom_domain_default.go` sets the default domain, and `module/outputs.go` exports them.

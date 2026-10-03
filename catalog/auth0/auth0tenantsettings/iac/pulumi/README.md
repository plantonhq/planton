# Auth0TenantSettings — Pulumi Module

Pulumi Go module that manages the settings of an existing Auth0 tenant.

## What It Creates

- `auth0.Tenant` — the settings of the tenant the provider's credential belongs to. It creates no tenant: the Management API cannot. `tenantArgs` sets each argument and block only when the spec sets it, so the tenant keeps every setting the spec leaves out -- except the six the provider resets on the first deploy (see the spec). The resource's delete is the provider's no-op, so the last-applied values stay in place.
- `auth0.CustomDomainDefault` — the tenant's default domain, declared only when the spec sets `default_custom_domain`. Its delete only forgets the default (Auth0 has no way to unset one).

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:tenant_settings` and `update:tenant_settings`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps each unset presentation string and reference to nil (never sent), the twin of `iac/tf/locals.tf`.
- `module/tenant.go` builds the tenant's arguments (`tenantArgs`, a pure function over the spec, tested in `tenant_test.go` to send nothing the spec leaves unset) and applies them; `module/custom_domain_default.go` sets the default domain, and `module/outputs.go` exports the settings.

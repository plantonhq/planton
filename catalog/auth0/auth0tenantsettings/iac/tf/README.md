# Auth0TenantSettings — Terraform Module

Terraform/OpenTofu module that manages the presentation settings of an existing Auth0 tenant.

## What It Creates

- `auth0_tenant` — the settings of the tenant the provider's credential belongs to. It creates no tenant: the Management API cannot. Only `friendly_name`, `picture_url`, `support_email` and `support_url` are set, each only when the spec sets it; every other tenant setting is left as it is. Destroy is the provider's no-op, so the last-applied values stay in place.
- `auth0_custom_domain_default` — the tenant's default domain, declared only when the spec sets `default_custom_domain` (`count`). Its delete only forgets the default (Auth0 has no way to unset one).

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:tenant_settings` and `update:tenant_settings`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `friendly_name`, `picture_url`, `support_email`, `support_url`, `default_custom_domain` -- each optional, at least one set |

## Outputs

| Name | Description |
|---|---|
| `friendly_name` | The tenant's name as people see it |
| `picture_url` | The URL of the tenant's logo |
| `support_email` | The support address the tenant's pages offer |
| `support_url` | The support page the tenant's pages link to |
| `default_custom_domain` | The tenant's default domain as set by this resource; empty when unmanaged |

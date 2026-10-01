# Auth0TenantSettings — Terraform Module

Terraform/OpenTofu module that manages the settings of an existing Auth0 tenant.

## What It Creates

- `auth0_tenant` — the settings of the tenant the provider's credential belongs to. It creates no tenant: the Management API cannot. Every argument is null and every block (`flags`, `session_cookie`, `sessions`, `oidc_logout`, `mtls`, `error_page`, `default_token_quota`, `country_codes`) absent unless the spec sets it, so the tenant keeps every setting the spec leaves out -- except the six the provider resets on the first deploy (see the spec). Destroy is the provider's no-op, so the last-applied values stay in place.
- `auth0_custom_domain_default` — the tenant's default domain, declared only when the spec sets `default_custom_domain` (`count`). Its delete only forgets the default (Auth0 has no way to unset one).

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:tenant_settings` and `update:tenant_settings`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | The tenant settings (`variables.tf`, generated from the spec) -- each optional, at least one set |

## Outputs

`friendly_name`, `picture_url`, `support_email`, `support_url`, `default_custom_domain` (empty when unmanaged), `default_audience`, `default_directory`, `client_id_metadata_document_supported`, `resource_parameter_profile`, `enable_dynamic_client_registration`, `dynamic_client_registration_security_mode`, `session_lifetime`, `idle_session_lifetime`, `session_cookie_mode` and `enabled_locales` -- the settings as the tenant carries them after the apply, managed or not.

# Auth0Branding — Terraform Module

Terraform/OpenTofu module that manages how an existing Auth0 tenant's Universal Login looks: the branding every page shares, and optionally the no-code theme of the login box.

## What It Creates

- `auth0_branding` — the branding of the tenant the provider's credential belongs to, declared only when the spec sets one of `logo_url`, `favicon_url`, `colors`, `font_url` or `universal_login_template` (`count`). Each setting is sent only when set (`font` and `universal_login` are blocks declared only then). Auth0 has no delete for the tenant's branding: destroy removes the page template (on a tenant with a custom domain) and leaves the last-applied logo, favicon, colors and font in place.
- `auth0_branding_theme` — the login box's theme, declared only when the spec sets `theme` (`count`). It is sent whole: `locals.tf` fills every field the spec leaves unset with Auth0's default (the provider's `internal/auth0/branding/resource_theme.go`), in one `theme_defaults` map. `identifiers` is sent only when declared. The provider adopts the tenant's existing theme on create; destroy deletes the theme.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:branding`, `update:branding`, `delete:branding` and `read:custom_domains`.
- A custom domain on the tenant when the spec sets `universal_login_template` (Auth0 refuses a page template on the canonical domain).

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `logo_url`, `favicon_url`, `colors`, `font_url`, `universal_login_template`, `theme` -- each optional, at least one set |

## Outputs

| Name | Description |
|---|---|
| `theme_id` | The id of the theme the branding applied; empty when the spec declares no theme |
| `logo_url` | The logo the tenant's pages show after the deployment; empty when the spec manages no branding setting |

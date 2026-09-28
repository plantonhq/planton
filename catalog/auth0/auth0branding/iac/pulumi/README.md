# Auth0Branding — Pulumi Module

Pulumi Go module that manages how an existing Auth0 tenant's Universal Login looks: the branding every page shares, and optionally the no-code theme of the login box.

## What It Creates

- `auth0.Branding` — the branding of the tenant the provider's credential belongs to, declared only when the spec sets one of `logo_url`, `favicon_url`, `colors`, `font_url` or `universal_login_template`. Each setting is sent only when set. Auth0 has no delete for the tenant's branding: the resource's delete removes the page template (on a tenant with a custom domain) and leaves the last-applied logo, favicon, colors and font in place.
- `auth0.BrandingTheme` — the login box's theme, declared only when the spec sets `theme`. It is sent whole: `module/locals.go` fills every field the spec leaves unset with Auth0's default (the provider's `internal/auth0/branding/resource_theme.go`), in one `themeDefaults` value. `identifiers` is sent only when declared. The provider adopts the tenant's existing theme on create; delete deletes the theme.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:branding`, `update:branding`, `delete:branding` and `read:custom_domains`.
- A custom domain on the tenant when the spec sets `universal_login_template` (Auth0 refuses a page template on the canonical domain).

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps each unset branding setting to nil (never sent) and resolves a declared theme against Auth0's defaults, the twin of `iac/tf/locals.tf`.
- `module/branding.go` applies the branding, `module/branding_theme.go` applies the theme, and `module/outputs.go` exports `theme_id` and `logo_url`.

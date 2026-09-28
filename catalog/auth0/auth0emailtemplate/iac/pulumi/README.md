# Auth0EmailTemplate — Pulumi Module

Pulumi Go module that customizes one of the emails an Auth0 tenant sends.

## What It Creates

- `auth0.EmailTemplate` — the template of one email (`verify_email`, `reset_email`, `welcome_email`, ...) in the tenant the provider's credential belongs to: its sender, subject and Liquid-templated HTML body, where its link leads afterwards, and how long that link lives. `syntax` defaults to `liquid` and `enabled` to `true`; the other optional settings are sent only when the spec sets them. An existing template is taken over. Auth0 cannot delete a template, so destroy disables it and the tenant sends Auth0's default email again.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:email_templates`, `create:email_templates` and `update:email_templates`.
- The tenant's email provider (Auth0EmailProvider): Auth0 refuses custom templates without one.

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` fills the spec's defaults (`liquid`, on) and maps each unset optional setting to nil (never sent), the twin of `iac/tf/locals.tf` (pinned by `module/locals_test.go`).
- `module/email_template.go` applies the template, and `module/outputs.go` exports it.

# Auth0EmailProvider — Pulumi Module

Pulumi Go module that manages the email provider of an Auth0 tenant.

## What It Creates

- `auth0.EmailProvider` — the one email provider of the tenant the provider's credential belongs to. The spec's service arm names it (`smtp`, `ses`, `sendgrid`, `sparkpost`, `mailgun`, `mandrill`, `azure_cs`, `ms365` or `custom`) and fills exactly the `credentials` and `settings` that service uses, every credential as a Pulumi secret; the `custom` arm sends the empty `credentials` block the provider requires. A tenant that already has a provider is taken over. Destroy deletes the provider, and the tenant falls back to Auth0's built-in test provider.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:email_provider`, `create:email_provider`, `update:email_provider` and `delete:email_provider`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps the service arm onto the provider's name, credentials and settings, the twin of `iac/tf/locals.tf` (pinned by `module/locals_test.go`).
- `module/email_provider.go` applies the provider, and `module/outputs.go` exports it.

# Auth0EmailProvider — Terraform Module

Terraform/OpenTofu module that manages the email provider of an Auth0 tenant.

## What It Creates

- `auth0_email_provider` — the one email provider of the tenant the provider's credential belongs to. The spec's service arm names it (`smtp`, `ses`, `sendgrid`, `sparkpost`, `mailgun`, `mandrill`, `azure_cs`, `ms365` or `custom`) and fills exactly the `credentials` and `settings` that service uses; the `custom` arm sends the empty `credentials` block the provider requires. A tenant that already has a provider is taken over. Destroy deletes the provider, and the tenant falls back to Auth0's built-in test provider.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:email_provider`, `create:email_provider`, `update:email_provider` and `delete:email_provider`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `default_from_address` (required), `enabled` (default `true`), and exactly one of `smtp`, `ses`, `sendgrid`, `sparkpost`, `mailgun`, `mandrill`, `azure_cs`, `ms365`, `custom` |

## Outputs

| Name | Description |
|---|---|
| `name` | The service the tenant sends through, as Auth0 names it |
| `default_from_address` | The sender of the tenant's emails |

`locals.tf` maps the service arm onto the provider's `name`, `credentials` and `settings` -- the same rule as the Pulumi module's `locals.go`.

# Auth0EmailTemplate — Terraform Module

Terraform/OpenTofu module that customizes one of the emails an Auth0 tenant sends.

## What It Creates

- `auth0_email_template` — the template of one email (`verify_email`, `reset_email`, `welcome_email`, ...) in the tenant the provider's credential belongs to: its sender, subject and Liquid-templated HTML body, where its link leads afterwards, and how long that link lives. `syntax` defaults to `liquid` and `enabled` to `true`; the other optional settings are sent only when the spec sets them. An existing template is taken over. Auth0 cannot delete a template, so destroy disables it and the tenant sends Auth0's default email again.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:email_templates`, `create:email_templates` and `update:email_templates`.
- The tenant's email provider (Auth0EmailProvider): Auth0 refuses custom templates without one.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `template`, `from`, `subject`, `body` (required); `syntax` (default `liquid`), `enabled` (default `true`), `result_url`, `url_lifetime_in_seconds`, `include_email_in_redirect` (optional) |

## Outputs

| Name | Description |
|---|---|
| `template` | The email this resource customizes |
| `enabled` | Whether the template is on |

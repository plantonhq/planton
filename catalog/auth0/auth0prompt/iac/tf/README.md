# Auth0Prompt — Terraform Module

Terraform/OpenTofu module that manages the login-flow settings of an existing Auth0 tenant.

## What It Creates

- `auth0_prompt` — the tenant's one set of prompt settings: `universal_login_experience`, `identifier_first` and `webauthn_platform_first_factor`, each sent only when the spec sets it; a setting the spec leaves out keeps the tenant's value. Destroy is the provider's no-op, so the last-applied values stay in place.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:prompts` and `update:prompts`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `universal_login_experience`, `identifier_first`, `webauthn_platform_first_factor` -- each optional, at least one set |

## Outputs

| Name | Description |
|---|---|
| `universal_login_experience` | The login experience the tenant runs: `new` or `classic` |
| `identifier_first` | Whether the login flow asks for the identifier first |
| `webauthn_platform_first_factor` | Whether the device's own authenticator is offered as the first factor |

# Auth0CustomDomainVerification — Terraform Module

Terraform/OpenTofu module that verifies an Auth0 custom domain and waits until it is ready.

## What It Creates

- `auth0_custom_domain_verification` — the verification of the custom domain `custom_domain_id` names, polled until Auth0 reports the domain `ready`. It has no update; its delete is the provider's no-op, so destroy leaves the domain verified.
- `data.auth0_custom_domain` — the verified domain, read back by id for the `domain` output.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `create:custom_domains` and `read:custom_domains`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Cloud resource metadata (`name`, `org`, `env`, ...) |
| `spec` | `custom_domain_id` -- the custom domain to verify |

## Outputs

| Name | Description |
|---|---|
| `custom_domain_id` | The verified custom domain's identifier (`cd_...`) |
| `domain` | The verified domain's name |
| `origin_domain_name` | The tenant host the domain serves from |
| `cname_api_key` | The self-managed proxy's key (sensitive; empty for Auth0-managed) |

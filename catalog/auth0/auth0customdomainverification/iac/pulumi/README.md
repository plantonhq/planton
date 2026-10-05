# Auth0CustomDomainVerification — Pulumi Module

Pulumi Go module that verifies an Auth0 custom domain and waits until it is ready.

## What It Creates

- `auth0.CustomDomainVerification` — the verification of the custom domain `custom_domain_id` names, polled until Auth0 reports the domain `ready`. It has no update; its delete is the provider's no-op, so destroy leaves the domain verified.
- A read of the verified domain (`auth0.LookupCustomDomain`, by id) for the `domain` output.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `create:custom_domains` and `read:custom_domains`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` resolves the custom domain id, the twin of `iac/tf/locals.tf`.
- `module/verification.go` verifies the domain; `module/outputs.go` exports it, with `cname_api_key` as a secret.

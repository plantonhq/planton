# Auth0CustomDomain — Pulumi Module

Pulumi Go module that creates a custom domain for an Auth0 tenant.

## What It Creates

- `auth0.CustomDomain` — the custom domain, in the tenant the provider's credential belongs to. The optional settings (client-IP header, TLS policy, relying party, metadata) are sent only when the spec sets them. Changing the domain or its type replaces it; destroy deletes it.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `create:custom_domains`, `read:custom_domains`, `update:custom_domains` and `delete:custom_domains`.

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` maps each unset setting to nil (never sent), the twin of `iac/tf/locals.tf`.
- `module/custom_domain.go` creates the domain.
- `module/outputs.go` exports it, and picks the DNS record to publish from Auth0's verification methods: the CNAME when offered, otherwise the first method (the same rule as `iac/tf/locals.tf`, pinned by `module/outputs_test.go`).

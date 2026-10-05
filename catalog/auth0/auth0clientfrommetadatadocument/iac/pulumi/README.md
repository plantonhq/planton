# Auth0ClientFromMetadataDocument — Pulumi Module

Pulumi Go module that registers an Auth0 application from its Client ID Metadata Document.

## What It Creates

- `auth0.ClientCimd` — the application, in the tenant the provider's credential belongs to (the tenant must allow registration from metadata documents). Auth0 fetches the document at `externalClientId` and registers the application from it; the module then sets only the settings the spec declares, so everything else stays as the document or the tenant set it. Changing `externalClientIdVersion` makes Auth0 fetch the document again; changing `externalClientId` replaces the application; destroy deletes it.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `create:clients`, `read:clients`, `update:clients` and `delete:clients`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` carries the resource name and the spec.
- `module/client.go` registers the application. `clientArgs` is a pure function of the spec that sets an argument only when the spec declares it (nil pointers, nil blocks and empty collections are left out), the twin of `iac/tf/locals.tf` and `iac/tf/main.tf`; `module/client_test.go` pins that an undeclared setting is never sent.
- `module/outputs.go` exports the application and the document's validation, read through `summarizeValidation` (the SDK's `Index(0)` panics on an empty list), the same rule as `iac/tf/outputs.tf`.

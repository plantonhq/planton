# Auth0ResourceServer — Pulumi Module

Pulumi Go module that creates an Auth0 Resource Server (API), its scopes, and the default grants every third-party application gets on it.

## What It Creates

- `auth0.ResourceServer` — the API, in the tenant the provider's credential belongs to. Every setting but `allow_offline_access`, `skip_consent_for_verifiable_first_party_clients` and `enforce_policies` is sent only when the spec declares it; each access-policy block is sent only with its policy. Changing `identifier` replaces the API; destroy deletes it.
- `auth0.ResourceServerScopes` — the API's authoritative scope list, when `spec.scopes` is non-empty.
- `auth0.ClientGrant` — one per `spec.third_party_client_default_grants` entry, with `default_for` `third_party_clients`, this API's identifier as audience, and no client id. Named `<metadata.name>-default-grant-<subject_type>`; created after the scopes.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `create:resource_servers`, `read:resource_servers`, `update:resource_servers` and `delete:resource_servers`, and the four `client_grants` scopes when default grants are declared (`../permissions.yaml`).

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` reads the spec; unset optional settings stay nil (never sent) -- the twin of `iac/tf/locals.tf`.
- `module/resourceserver.go` builds the API's arguments (`resourceServerArgs`, a pure function of the spec) and its scopes. `signing_secret` is sent as a Pulumi secret.
- `module/default_grants.go` declares the default grants (`defaultGrantArgs`, pure).
- `module/outputs.go` exports the API and `third_party_client_default_grant_ids`, keyed by subject type.
- `module/locals_test.go` pins that a spec declaring only its identifier sends none of the unmanaged settings and declares no grant.

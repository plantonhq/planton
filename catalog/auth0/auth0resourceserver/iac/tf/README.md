# Auth0ResourceServer — Terraform Module

Terraform/OpenTofu module that creates an Auth0 Resource Server (API), its scopes, and the default grants every third-party application gets on it.

## What It Creates

- `auth0_resource_server` — the API, in the tenant the provider's credential belongs to. Every setting but `allow_offline_access`, `skip_consent_for_verifiable_first_party_clients` and `enforce_policies` is sent only when the spec declares it (null attributes, blocks rendered only when declared); each access-policy block is rendered only with its policy. `signing_secret` is marked sensitive. Changing `identifier` replaces the API; destroy deletes it.
- `auth0_resource_server_scopes` — the API's authoritative scope list, when `spec.scopes` is non-empty.
- `auth0_client_grant.third_party_client_default_grants` — one per `spec.third_party_client_default_grants` entry, keyed by subject type, with `default_for = "third_party_clients"`, this API's identifier as audience, and no `client_id`; created after the scopes.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs the four `resource_servers` scopes, and the four `client_grants` scopes when default grants are declared (`../permissions.yaml`).

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `identifier` (required); token settings, `scopes`, the access policy (`subject_type_authorization`), proof of possession, token encryption, authorization details and policy, anonymous-session settings, Online Refresh Tokens, and `third_party_client_default_grants` (optional) |

## Outputs

| Name | Description |
|---|---|
| `id` | The internal Auth0 identifier |
| `identifier` | The API identifier (audience) |
| `name` | The friendly display name |
| `signing_alg` | The token signing algorithm |
| `signing_secret` | The signing secret (HS256 only, sensitive) |
| `token_lifetime` | Token validity duration in seconds |
| `token_lifetime_for_web` | Token validity for implicit/hybrid flows |
| `allow_offline_access` | Whether refresh tokens can be issued |
| `skip_consent_for_verifiable_first_party_clients` | Consent skip setting |
| `enforce_policies` | Whether RBAC is enabled |
| `token_dialect` | Access token format |
| `is_system` | Whether this is a system resource server |
| `client_id` | The client linked to the API, if any |
| `third_party_client_default_grant_ids` | The default grants' ids (`cgr_...`), keyed by subject type |

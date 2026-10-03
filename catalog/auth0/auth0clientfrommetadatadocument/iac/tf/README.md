# Auth0ClientFromMetadataDocument — Terraform Module

Terraform/OpenTofu module that registers an Auth0 application from its Client ID Metadata Document.

## What It Creates

- `auth0_client_cimd` — the application, in the tenant the provider's credential belongs to (the tenant must allow registration from metadata documents). Auth0 fetches the document at `external_client_id` and registers the application from it; the module then sets only the settings the spec declares, so everything else stays as the document or the tenant set it. Changing `external_client_id_version` makes Auth0 fetch the document again; changing `external_client_id` replaces the application; destroy deletes it.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `create:clients`, `read:clients`, `update:clients` and `delete:clients`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `external_client_id` (required); `external_client_id_version`, `app_type`, `grant_types`, `description`, `allowed_origins`, `web_origins`, `oidc_conformant`, `require_proof_of_possession`, `skip_non_verifiable_callback_uri_confirmation_prompt`, `redirection_policy`, `organization_discovery_methods`, `default_organization`, `client_metadata`, `jwt_configuration`, `refresh_token`, `token_quota` (optional) |

`variables.tf` is generated from the spec proto (`planton tofu generate-variables`); never edit it by hand.

## Outputs

| Name | Description |
|---|---|
| `client_id` | The application's id in the Management API (`tpc_...`) |
| `external_client_id` | The document's URL, the client_id presented in sign-in flows |
| `name` | The application's name, from the document |
| `app_type` | The application type Auth0 holds |
| `grant_types` | The grants Auth0 holds |
| `callbacks` | The redirect URIs, from the document |
| `logo_uri` | The logo, from the document |
| `jwks_uri` | Where the application publishes its public keys |
| `third_party_security_mode` | Always `strict` |
| `external_metadata_created_by` | `admin` or `client` |
| `validation_valid` | Whether the document passed validation at the last read |
| `validation_warnings` | What Auth0 ignored in the document |
| `validation_violations` | What prevents Auth0 from processing the document fully |

`locals.tf` maps empty lists and maps to null and renders each declared block once -- the same unset-means-unmanaged rule as the Pulumi module's `clientArgs`.

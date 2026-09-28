# Auth0ClientFromMetadataDocument

Registers an application in an [Auth0 tenant](https://auth0.com/docs/get-started/auth0-overview/create-applications/register-applications-with-cimd) from its Client ID Metadata Document -- a JSON file the application's owner hosts at an https URL. Auth0 fetches the document and registers the application from it: the path an MCP client takes to onboard itself, with no client secret to create or hand over.

## When to Use

- **MCP clients**: register an agent once per deployment from the document it publishes, and every instance signs in with the same registration.
- **Partner integrations**: an external application that manages its own metadata and keys, rotating them without sharing a secret.
- **Token policy in code**: the lifetimes, refresh-token rotation, origins and redirect policy the tenant sets over the document, reviewed like any other change.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ClientFromMetadataDocument
metadata:
  name: mcp-client
spec:
  externalClientId: https://mcp-client.example.com/.well-known/oauth-client-metadata
  externalClientIdVersion: 1
  description: MCP client for the team's tools
```

## Who Owns What

| Owner | Settings |
|---|---|
| The document only (reported as outputs) | name (`client_name`), redirect URIs (`redirect_uris`), logo, token-endpoint authentication (`none` or `private_key_jwt`), public keys (`jwks_uri`) |
| The document seeds, the spec may override | `appType`, `grantTypes`, `description` |
| The spec only | `allowedOrigins`, `webOrigins`, `oidcConformant`, `requireProofOfPossession`, `skipNonVerifiableCallbackUriConfirmationPrompt`, `redirectionPolicy`, `organizationDiscoveryMethods`, `defaultOrganization`, `clientMetadata`, `jwtConfiguration`, `refreshToken`, `tokenQuota` |

## Key Behaviors

- **Unset means unmanaged**: the module sends only what the spec declares; Auth0 keeps what the document or the tenant set.
- **The document is read once**: Auth0 fetches it at registration. Raise `externalClientIdVersion` after the document changes and the next apply fetches it again, then re-applies the spec's settings over it.
- **Keyed by URL**: registration is an upsert keyed by `externalClientId`, so an already-registered URL is taken over; changing the URL replaces the application with a new `client_id`. Destroy deletes it.
- **Strict third-party application**: it signs people in only through domain-level connections, reaches an API only through an explicit client grant, holds only the `authorization_code` and `refresh_token` grants, and its login fails while the tenant has active Rules.
- **Prerequisite**: the tenant's Client ID Metadata Document Registration setting must be on (`Auth0TenantSettings`).
- **Plans**: registration is on every plan; a confidential client (`private_key_jwt`) needs the Enterprise plan.
- **Permissions**: the credential needs `create:clients`, `read:clients`, `update:clients` and `delete:clients` (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `client_id` | The application's Management API id (`tpc_...`), what grants and connections name |
| `external_client_id` | The document's URL, the `client_id` presented in sign-in flows |
| `name`, `logo_uri`, `callbacks`, `jwks_uri` | What Auth0 took from the document |
| `app_type`, `grant_types` | What Auth0 holds, the spec's or the document's |
| `third_party_security_mode` | Always `strict` |
| `external_metadata_created_by` | `admin` or `client` |
| `validation_valid`, `validation_warnings`, `validation_violations` | Auth0's verdict on the document at the last read |

## Auth0 Documentation

- [Register applications with CIMD](https://auth0.com/docs/get-started/auth0-overview/create-applications/register-applications-with-cimd)
- [Security controls for third-party applications](https://auth0.com/docs/get-started/applications/third-party-applications/security-controls)
- [Terraform auth0_client_cimd](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/client_cimd)
- [Pulumi auth0.ClientCimd](https://www.pulumi.com/registry/packages/auth0/api-docs/clientcimd/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

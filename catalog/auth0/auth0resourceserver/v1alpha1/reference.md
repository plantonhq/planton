# Auth0ResourceServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0ResourceServerSpec defines the configuration for an Auth0 Resource Server (API).
In Auth0, Resource Servers represent APIs that your applications can request access to.
They define the audience parameter used in authorization requests and the scopes (permissions)
that can be granted to applications.

This spec covers the 80/20 use case for configuring Auth0 APIs:
- Custom backend APIs with JWT-based access control
- APIs requiring role-based access control (RBAC)
- APIs with defined scopes/permissions for fine-grained authorization

https://auth0.com/docs/get-started/apis

## Example

```yaml
# Auth0 Resource Server Test Manifest
# This file is used for testing the Auth0ResourceServer component
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: Your Auth0 tenant domain (e.g., your-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:resource_servers
#    - read:resource_servers
#    - update:resource_servers
#    - delete:resource_servers

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: test-api
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # Required: API identifier (audience)
  identifier: https://api.test.planton.dev/

  # Optional: Friendly display name
  name: Test API

  # Token configuration
  signing_alg: RS256
  token_lifetime: 86400         # 24 hours
  token_lifetime_for_web: 7200  # 2 hours

  # Access control
  allow_offline_access: true
  skip_consent_for_verifiable_first_party_clients: true
  enforce_policies: true
  token_dialect: access_token_authz

  # API Scopes/Permissions
  scopes:
    - name: read:items
      description: Read access to items
    - name: write:items
      description: Create and update items
    - name: delete:items
      description: Delete items
    - name: admin:all
      description: Full administrative access
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.identifier` | `string` | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.signingAlg` | `string` |  |  |  |
| `spec.allowOfflineAccess` | `bool` |  |  |  |
| `spec.tokenLifetime` | `int32` |  |  |  |
| `spec.tokenLifetimeForWeb` | `int32` |  |  |  |
| `spec.skipConsentForVerifiableFirstPartyClients` | `bool` |  |  |  |
| `spec.enforcePolicies` | `bool` |  |  |  |
| `spec.tokenDialect` | `string` |  |  |  |
| `spec.scopes` | `[]Auth0ResourceServerScope` |  |  |  |
| `spec.scopes[].name` | `string` | yes |  |  |
| `spec.scopes[].description` | `string` |  |  |  |

## Field Details

### spec.identifier

`string` · required

identifier is the unique identifier for the resource server.
This value is used as the "audience" parameter for authorization calls.
Typically a URI representing your API (e.g., "https://api.example.com/").
Cannot be changed once set.

Example: "https://api.mycompany.com/", "api.planton.live"

- rule: {"required":true}

### spec.name

`string`

name is a friendly display name for the resource server.
This is shown in the Auth0 dashboard and consent screens.
Cannot include `<` or `>` characters.

### spec.signingAlg

`string`

signing_alg is the algorithm used to sign access tokens for this API.
Options:
- "RS256": RSA using SHA-256 (asymmetric, recommended)
- "HS256": HMAC using SHA-256 (symmetric, requires client secret)
- "PS256": RSA-PSS using SHA-256
Default: RS256

- rule: {"string":{"in":["","RS256","HS256","PS256"]}}

### spec.allowOfflineAccess

`bool`

allow_offline_access indicates whether refresh tokens can be issued for this API.
When true, applications can request refresh tokens using the "offline_access" scope.
This allows applications to obtain new access tokens without user interaction.
Default: false

### spec.tokenLifetime

`int32`

token_lifetime is the duration (in seconds) that access tokens remain valid
when issued from the token endpoint.
Range: 0 to 2592000 (30 days)
Default: 86400 (24 hours)

- rule: {"int32":{"lte":2592000,"gte":0}}

### spec.tokenLifetimeForWeb

`int32`

token_lifetime_for_web is the duration (in seconds) that access tokens remain valid
when issued via implicit or hybrid flows.
This should typically be shorter than token_lifetime for security.
Cannot be greater than token_lifetime.
Range: 0 to 2592000 (30 days)
Default: 7200 (2 hours)

- rule: {"int32":{"lte":2592000,"gte":0}}

### spec.skipConsentForVerifiableFirstPartyClients

`bool`

skip_consent_for_verifiable_first_party_clients indicates whether to skip
the consent prompt for applications flagged as first-party.
When true, first-party applications don't show the consent screen to users.
Default: true

### spec.enforcePolicies

`bool`

enforce_policies enables RBAC authorization policies for this API.
When true, role and permission assignments are evaluated during login.
This allows you to use Auth0's built-in RBAC to control API access.
Requires token_dialect to be set to a value that includes permissions.
Default: false

https://auth0.com/docs/manage-users/access-control/rbac

### spec.tokenDialect

`string`

token_dialect determines the format of access tokens issued for this API.
Options:
- "access_token": Standard Auth0 JWT with claims
- "access_token_authz": Standard Auth0 JWT including RBAC permissions claims
- "rfc9068_profile": IETF JWT Access Token Profile compliant
- "rfc9068_profile_authz": IETF profile with RBAC permissions claims

Use "_authz" variants when enforce_policies is true to include permissions in tokens.
Default: access_token

https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles

- rule: {"string":{"in":["","access_token","access_token_authz","rfc9068_profile","rfc9068_profile_authz"]}}

### spec.scopes

`[]Auth0ResourceServerScope`

scopes defines the permissions that can be granted for this API.
Each scope represents a specific permission that applications can request.
Applications request scopes during authorization, and granted scopes
appear in the access token's "scope" claim.

Example scopes:
- read:users (permission to read user data)
- write:orders (permission to create/update orders)
- delete:products (permission to delete products)

https://auth0.com/docs/get-started/apis/api-settings#scopes

### spec.scopes[].name

`string` · required

name is the scope identifier used in OAuth flows.
Should follow the pattern: action:resource (e.g., "read:users", "write:orders")
This is what applications request and what appears in access tokens.

- rule: {"required":true}

### spec.scopes[].description

`string`

description is a human-readable explanation of what this scope grants.
Shown on consent screens and in the Auth0 dashboard.
Example: "Read access to user profiles"

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0ResourceServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the internal Auth0 identifier for this resource server. This is a unique string assigned by Auth0 when the resource is created. Used for API calls to manage the resource server. |
| `status.outputs.identifier` | `string` | identifier is the API identifier (audience) for this resource server. This is the value used in authorization requests as the "audience" parameter. Same as the identifier specified in the spec. |
| `status.outputs.name` | `string` | name is the friendly display name of the resource server. Derived from spec.name or metadata.name. |
| `status.outputs.signing_alg` | `string` | signing_alg is the algorithm used to sign tokens for this API. One of: RS256, HS256, PS256 |
| `status.outputs.signing_secret` | `string` | signing_secret is the secret used for signing tokens (HS256 only). This is only populated when signing_alg is HS256. IMPORTANT: Keep this secret secure and never expose in client-side code. |
| `status.outputs.token_lifetime` | `string` | token_lifetime is the configured token validity duration in seconds. |
| `status.outputs.token_lifetime_for_web` | `string` | token_lifetime_for_web is the token validity for implicit/hybrid flows. |
| `status.outputs.allow_offline_access` | `string` | allow_offline_access indicates if refresh tokens can be issued. |
| `status.outputs.skip_consent_for_verifiable_first_party_clients` | `string` | skip_consent_for_verifiable_first_party_clients indicates consent skip setting. |
| `status.outputs.enforce_policies` | `string` | enforce_policies indicates if RBAC is enabled for this API. |
| `status.outputs.token_dialect` | `string` | token_dialect is the access token format configured for this API. |
| `status.outputs.is_system` | `string` | is_system indicates if this is a system-managed resource server. System resource servers (like the Auth0 Management API) cannot be modified. |
| `status.outputs.client_id` | `string` | client_id is the associated client ID if one has been linked. Some resource servers may have an associated client for certain features. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0Client | `spec.apiGrants[].audience` | `status.outputs.identifier` |
| Auth0User | `spec.permissions[].resourceServerIdentifier` | `status.outputs.identifier` |

## See Also

- [Overview](../README.md)

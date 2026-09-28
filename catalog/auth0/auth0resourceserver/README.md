# Auth0 Resource Server

The Auth0ResourceServer component configures a Resource Server (API) in Auth0. Resource Servers represent the APIs that your applications can request access to, defining the audience parameter used in authorization requests and the scopes (permissions) that can be granted.

## Why Use Auth0 Resource Server?

When building APIs that need to be accessed by multiple applications, you need a way to:

1. **Define API Identity**: Give your API a unique identifier (audience) that applications use when requesting access
2. **Control Token Settings**: Configure how access tokens are issued, including lifetime and signing algorithm
3. **Define Permissions**: Create scopes that represent specific permissions applications can request
4. **Enable RBAC**: Use Auth0's built-in role-based access control for fine-grained authorization

## Key Features

- **API Identifier (Audience)**: Unique URI that identifies your API in OAuth flows
- **Token Configuration**: Control token lifetime, signing algorithm, and format
- **Scopes/Permissions**: Define granular permissions for API access
- **RBAC Support**: Enable role-based access control with permission claims in tokens
- **First-Party Consent Skip**: Automatically trust first-party applications
- **Access Policy**: Decide which applications can get a token at all, for people and for machines
- **Default Grants for Third-Party Applications**: Give every third-party application -- partners, AI agents, MCP clients registered through Dynamic Client Registration or a Client ID Metadata Document -- the same scopes without a grant per application
- **Token Protection**: Sender-constrained tokens (DPoP, mTLS), encrypted tokens (JWE), rich authorization request types

## Quick Start

### Basic API Configuration

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: my-api
spec:
  identifier: https://api.example.com/
  name: My Example API
  signingAlg: RS256
  tokenLifetime: 86400
  allowOfflineAccess: true
```

### API with Scopes

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: products-api
spec:
  identifier: https://api.example.com/products
  name: Products API
  signingAlg: RS256
  tokenLifetime: 3600
  scopes:
    - name: read:products
      description: Read product catalog
    - name: write:products
      description: Create and update products
    - name: delete:products
      description: Delete products
```

### RBAC-Enabled API

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: rbac-api
spec:
  identifier: https://api.example.com/v2
  name: RBAC-Enabled API
  signingAlg: RS256
  tokenLifetime: 3600
  enforcePolicies: true
  tokenDialect: access_token_authz
  skipConsentForVerifiableFirstPartyClients: true
  scopes:
    - name: read:users
      description: Read user profiles
    - name: write:users
      description: Create and update users
    - name: admin:users
      description: Full administrative access to users
```

### API for an MCP Server

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: mcp-server-api
spec:
  identifier: https://mcp.example.com/mcp
  name: Example MCP Server
  signingAlg: RS256
  tokenDialect: rfc9068_profile
  allowOfflineAccess: true
  skipConsentForVerifiableFirstPartyClients: true
  scopes:
    - name: tools:read
      description: List the server's tools and read their results
    - name: tools:call
      description: Call the server's tools on your behalf
  subjectTypeAuthorization:
    user:
      policy: require_client_grant
    client:
      policy: deny_all
  thirdPartyClientDefaultGrants:
    - subjectType: user
      scopes:
        - tools:read
        - tools:call
```

Every third-party application gets `tools:read` and `tools:call` on a person's behalf; no application gets a token for itself.

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `identifier` | string | Yes | Unique API identifier (audience). Typically a URI. Cannot be changed after creation. |
| `name` | string | No | Friendly display name for the API |
| `signing_alg` | string | No | Token signing algorithm: RS256 (default), HS256, PS256 |
| `allow_offline_access` | bool | No | Allow refresh tokens (default: false) |
| `token_lifetime` | int32 | No | Token validity in seconds (default: 86400, max: 2592000) |
| `token_lifetime_for_web` | int32 | No | Token lifetime for implicit/hybrid flows |
| `skip_consent_for_verifiable_first_party_clients` | bool | No | Skip consent for first-party apps (default: true) |
| `enforce_policies` | bool | No | Enable RBAC authorization (default: false) |
| `token_dialect` | string | No | Token format: access_token, access_token_authz, rfc9068_profile, rfc9068_profile_authz |
| `scopes` | list | No | API permissions/scopes (authoritative once set) |
| `allowOnlineAccess`, `allowOnlineAccessWithEphemeralSessions` | bool | No | Online Refresh Tokens (Beta) |
| `consentPolicy` | string | No | `transactional-authorization-with-mfa` or `null` |
| `tokenLifetimeForAnonymousAccessTokens` | int32 | No | Anonymous-session token lifetime, 86400 to 2592000 (Early Access) |
| `verificationLocation` | string | No | JWKS URL for token introspection |
| `signingSecret` | string | No | HS256 shared secret, sensitive, 16+ characters |
| `accessToken` | object | No | Anonymous-session claims mapping (Early Access) |
| `authorizationDetails` | list | No | Rich Authorization Request types |
| `authorizationPolicy` | object | No | Authorization policy id (Early Access) |
| `proofOfPossession` | object | No | DPoP or mTLS sender constraining |
| `subjectTypeAuthorization` | object | No | Access policy for users, clients and anonymous users |
| `tokenEncryption` | object | No | JWE token encryption with the API's public key |
| `thirdPartyClientDefaultGrants` | list | No | Default grants for third-party applications, one per subject type |

Unset means unmanaged for everything but `allowOfflineAccess`, `skipConsentForVerifiableFirstPartyClients` and `enforcePolicies`, which are always sent. The field reference (`v1alpha1/reference.md`) names the settings an adopted API must declare, and the plan each setting needs.

### Token Dialects

| Dialect | Description |
|---------|-------------|
| `access_token` | Standard Auth0 JWT with claims |
| `access_token_authz` | Auth0 JWT with RBAC permissions claims |
| `rfc9068_profile` | IETF JWT Access Token Profile compliant |
| `rfc9068_profile_authz` | IETF profile with RBAC permissions |

Use `_authz` variants when `enforce_policies` is true to include permissions in tokens.

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Auth0's internal resource server ID |
| `identifier` | The API audience/identifier |
| `name` | The resource server display name |
| `signing_alg` | Configured signing algorithm |
| `signing_secret` | HS256 signing secret (if applicable) |
| `token_lifetime` | Configured token lifetime |
| `is_system` | Whether this is a system resource server |
| `third_party_client_default_grant_ids` | The default grants' ids (`cgr_...`), keyed by subject type |

## Best Practices

1. **Use RS256 for Signing**: RS256 (asymmetric) is more secure than HS256 (symmetric) and doesn't require sharing secrets.

2. **Define Meaningful Scopes**: Follow the `action:resource` naming convention (e.g., `read:users`, `write:orders`).

3. **Enable RBAC for Complex Apps**: When you need role-based permissions, enable `enforce_policies` and use `access_token_authz` dialect.

4. **Set Appropriate Token Lifetimes**: Balance security with user experience. Shorter lifetimes are more secure.

5. **Pair the Access Policy with Grants**: Moving `subjectTypeAuthorization.user.policy` to `require_client_grant` locks out every application without a `user` grant; deploy the grants (Auth0 Client `apiGrants`, or `thirdPartyClientDefaultGrants`) with the policy.

6. **Use Refresh Tokens Carefully**: Only enable `allow_offline_access` when applications genuinely need to refresh tokens without user interaction.

## Related Resources

- [Auth0 APIs Documentation](https://auth0.com/docs/get-started/apis)
- [Access Token Profiles](https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles)
- [RBAC Documentation](https://auth0.com/docs/manage-users/access-control/rbac)
- [API Scopes](https://auth0.com/docs/get-started/apis/api-settings#scopes)
- [API Access Policies for Applications](https://auth0.com/docs/get-started/apis/api-access-policies-for-applications)
- [Client Grants and Default Permissions for Third-Party Applications](https://auth0.com/docs/get-started/applications/application-access-to-apis-client-grants)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

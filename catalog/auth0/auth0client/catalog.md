# Auth0 Application (Client)

Deploys an Auth0 Application (Client) -- the OAuth 2.0 client that users and services authenticate through -- with configurable grant types, token settings, and optional API access grants. Supports all four application types: `spa` and `native` are public clients that authenticate with PKCE and never receive a secret, while `regular_web` and `non_interactive` (machine-to-machine) are confidential clients that hold a client secret. API access is granted separately from authentication: an M2M client without an `apiGrants` entry can obtain tokens but cannot call any API.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auth0 Client (Application)** -- an application registered in Auth0 with the specified type, OAuth settings, callback URLs, JWT configuration, and refresh token behavior
- **Client Grants** -- created only when `apiGrants` is configured, one grant per entry authorizing this client to call the specified API with the listed scopes

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **An Auth0 Resource Server** (only for `apiGrants`) -- each grant's audience must identify an existing API. Provide the audience directly or reference an Auth0ResourceServer Infra Component via ValueFromRef.
- **An Auth0 Connection** (only for `enabledConnections`) -- restricting the client to specific identity providers requires the connections to exist. Provide connection names directly or reference Auth0Connection Infra Components via ValueFromRef.
- **M2M token quota** (only for `non_interactive` clients) -- client-credentials tokens count against the tenant's monthly M2M token quota (1,000/month on the free plan); the application object itself is free.

## Deploy

### Console

Open the deployment store, find **Auth0 Application (Client)**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the application type, and the OAuth, token, and API grant settings.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Client
metadata:
  name: acme-web-spa
  org: acme-corp
  env: prod
spec:
  applicationType: spa
  oidcConformant: true
  callbacks:
    - https://app.acme.com/callback
  allowedLogoutUrls:
    - https://app.acme.com
  webOrigins:
    - https://app.acme.com
  grantTypes:
    - authorization_code
    - refresh_token
```

```shell
planton apply -f auth0-client.yaml
```

This creates an OIDC-conformant Single Page Application in Auth0 -- a public client with no client secret, restricted to the registered callback and logout URLs. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the client to connections and resource servers deployed in the same InfraPipeline:

```yaml
spec:
  enabledConnections:
    - valueFrom:
        kind: Auth0Connection
        name: user-db
        fieldPath: status.outputs.name
  apiGrants:
    - audience:
        valueFrom:
          kind: Auth0ResourceServer
          name: backend-api
          fieldPath: status.outputs.identifier
      scopes:
        - read:data
```

The InfraPipeline resolves the dependency graph, deploys the connection and resource server first, then provisions the client with the resolved values.

## Key Configuration

These are the most important decisions when configuring an Auth0 Application. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Application type** -- Choose `applicationType` by where the credential can live, not by what the app looks like. `spa` and `native` run in environments that cannot keep a secret, so they get PKCE and no `client_secret`; `regular_web` and `non_interactive` run server-side and receive one. Registering a browser app as `regular_web` ships a secret you cannot protect; registering a backend as `spa` leaves it unable to use the client-credentials flow.

**API grants vs grant types** -- `grantTypes` only enables OAuth flows; `apiGrants` is what authorizes APIs. An M2M application with `client_credentials` in `grantTypes` but no `apiGrants` entry authenticates successfully and can call nothing. Each grant pairs an audience (resource server identifier) with the scopes to allow.

**JWT signing algorithm** -- Leave `jwtConfiguration.alg` at RS256 (the module's default when `jwtConfiguration` is set). JWKS-verifying consumers -- NextAuth and most OIDC libraries -- reject an HS256 id_token because they cannot fetch a symmetric key from the JWKS endpoint. Use HS256 only when every token consumer can securely hold the client secret.

**Refresh token rotation** -- Auth0's default is `non-rotating`, a legacy behavior. Set `refreshToken.rotationType: rotating` for production so each use invalidates the previous token, bounding the blast radius of a stolen refresh token. Pair it with `expirationType: expiring` and explicit `tokenLifetime`/`idleTokenLifetime` for time-bounded sessions.

**Enabled connections** -- Leaving `enabledConnections` empty makes every connection in the tenant available to this application; listing any restricts the application to exactly that list. This is the mirror image of the connection side, where an empty `enabledClients` list means no application can use the connection.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **Auth0Connection** (optional) | `enabledConnections` | `status.outputs.name` |
| **Auth0ResourceServer** (optional) | `apiGrants[].audience` | `status.outputs.identifier` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `client_id` | OAuth 2.0 client identifier (public) | Auth0Connection `enabledClients`, application OAuth configuration |
| `client_secret` | OAuth 2.0 client secret -- confidential clients (`regular_web`, `non_interactive`) only | Backend and M2M service credentials |
| `signing_keys` | Signing certificates for this client's RS256 tokens | Pinned JWT signature verification in backend services |
| `token_endpoint_auth_method` | How the client authenticates to the token endpoint | OAuth client library configuration |

## Common Patterns

**Single-page application** -- `applicationType: spa` with `callbacks`, `webOrigins`, and `allowedLogoutUrls` covering every environment the app runs in, and `grantTypes` of `authorization_code` plus `refresh_token`. There is no secret to manage; pair with rotating refresh tokens since the browser is the least trusted place a long-lived credential can live.

**Machine-to-machine service** -- `applicationType: non_interactive` with `grantTypes: [client_credentials]` and at least one `apiGrants` entry. The grant, not the application type, is what confers API access -- scope it to the minimum the service calls. Each token issued counts against the tenant's M2M token quota.

**Regular web application** -- `applicationType: regular_web` for server-rendered apps that can hold the `client_secret`. The secret enables confidential-client authentication at the token endpoint; the trade is that it becomes a credential you must store and rotate server-side.

## Works With

- [**Auth0 Connection (Identity Provider)**](/infra-catalog/auth0-connection) -- provides identity provider connections that this client can use for authentication via `enabledConnections`
- [**Auth0 Resource Server (API)**](/infra-catalog/auth0-resource-server) -- defines APIs that this client can be authorized to access via `apiGrants`

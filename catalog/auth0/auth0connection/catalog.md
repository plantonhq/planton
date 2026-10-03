# Auth0 Connection (Identity Provider)

Deploys an Auth0 Connection that bridges Auth0 with an identity source -- a hosted user database, a social provider (Google, Facebook, GitHub), or an enterprise identity provider (SAML, OIDC, Azure AD/Entra ID). Each connection carries exactly one strategy and its provider-specific options block, and the strategy is fixed for the connection's lifetime. A connection only becomes a login option for the applications listed in `enabledClients` -- left empty, no application can authenticate through it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auth0 Connection** -- a connection resource configured with the specified strategy, display name, and provider-specific options (database, social, SAML, OIDC, or Azure AD)
- **Connection Clients** -- created only when `enabledClients` is configured, a resource linking this connection to the specified Auth0 applications so they can offer it as a login option

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **OAuth credentials from the social provider** (only for social strategies) -- a client ID and secret from the provider's developer console, set in `socialOptions`.
- **Identity Provider metadata** (only for enterprise strategies) -- SAML sign-in endpoint and X.509 signing certificate for `samlOptions`, the issuer URL for `oidcOptions`, or the app registration credentials and tenant domain for `azureAdOptions`.
- **Auth0 Client application IDs** (only for `enabledClients`) -- provide client IDs directly or reference Auth0Client Infra Components via ValueFromRef.
- **The password-advanced-options entitlement** (only for `passwordHistorySize`, `passwordNoPersonalInfo`, or `passwordDictionary`) -- these database options call a paid Auth0 API and the deployment fails with a 403 on free and lower-tier tenants. Leave them unset unless the tenant carries the entitlement.

## Deploy

### Console

Open the deployment store, find **Auth0 Connection (Identity Provider)**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the strategy, and the options block that strategy requires.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Connection
metadata:
  name: user-db
  org: acme-corp
  env: prod
spec:
  strategy: auth0
  databaseOptions:
    passwordPolicy: good
    bruteForceProtection: true
```

```shell
planton apply -f auth0-connection.yaml
```

This creates an Auth0-hosted database connection with a `good` password policy and brute-force protection enabled. Until `enabledClients` lists at least one application, no application offers this connection as a login option. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the connection to applications deployed in the same InfraPipeline:

```yaml
spec:
  enabledClients:
    - valueFrom:
        kind: Auth0Client
        name: my-spa
        fieldPath: status.outputs.client_id
```

The InfraPipeline resolves the dependency graph, deploys the client first, then provisions the connection with the resolved client ID.

## Key Configuration

These are the most important decisions when configuring an Auth0 Connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Strategy** -- The `strategy` field is a one-way door: it determines the identity source type, which options block applies, and it cannot be changed after creation -- switching a connection from `auth0` to `samlp` means a new connection and a user migration. Use `auth0` for a hosted database, social strategy names like `google-oauth2` or `github` for social login, and `samlp`, `oidc`, or `waad` for corporate SSO.

**Enabled clients** -- The `enabledClients` field is the connection's blast door. If left empty, no application can authenticate through this connection -- users see nothing and login attempts fail, with no error at deploy time. List every application that should offer this connection, directly by client ID or by referencing Auth0Client Infra Components via ValueFromRef.

**Password policy** -- The `databaseOptions.passwordPolicy` field sets complexity requirements for database connections, from `none` to `excellent` (10+ characters with mixed case, numeric, and special). Use `good` or higher for production with `bruteForceProtection: true`. The advanced controls (`passwordHistorySize`, `passwordNoPersonalInfo`, `passwordDictionary`) require the paid entitlement noted in Before You Deploy.

**Enterprise provider configuration** -- Each enterprise strategy requires its own options block: `samlOptions` with the IdP's sign-in endpoint and signing certificate, `oidcOptions` with the issuer (Auth0 discovers the rest from `/.well-known/openid-configuration`), and `azureAdOptions` with the app registration credentials and tenant domain. A strategy deployed without its matching options block produces a connection that cannot authenticate anyone.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **Auth0Client** (optional) | `enabledClients` | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Unique Auth0 connection identifier (e.g., `con_...`) | Management API operations, tenant log queries |
| `name` | Unique connection name within the tenant | Auth0Client `enabledConnections` references |

## Common Patterns

**Hosted user database** -- `strategy: auth0` with a `good`-or-better password policy and brute-force protection. The default when you own the user store: Auth0 stores credentials as bcrypt hashes and you get signup, password reset, and MFA hooks without running a database.

**Social login** -- one connection per provider (`google-oauth2`, `github`, `facebook`) with that provider's OAuth credentials in `socialOptions`. The `scopes` list decides what profile data Auth0 receives; request only what the application reads.

**Enterprise SSO** -- `samlp`, `oidc`, or `waad`, typically one connection per corporate IdP. In B2B setups, provision a new connection per customer rather than reworking an existing one -- the strategy is immutable and each customer's IdP metadata is independent.

## Works With

- [**Auth0 Application (Client)**](/infra-catalog/auth0-client) -- applications that use this connection for authentication, linked via `enabledClients`

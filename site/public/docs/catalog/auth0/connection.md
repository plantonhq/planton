---
title: "Connection"
description: "Connection deployment documentation"
icon: "package"
order: 100
componentName: "auth0connection"
---

# Auth0 Connection

Deploys an Auth0 Connection that bridges Auth0 with an identity source, enabling users to authenticate via databases, social providers (Google, Facebook, GitHub), or enterprise identity providers (SAML, OIDC, Azure AD/Entra ID). Each connection is configured with a single strategy and its corresponding provider-specific options, then linked to one or more Auth0 applications.

## What Gets Created

When you deploy an Auth0Connection resource, OpenMCF provisions:

- **Auth0 Connection** — an `auth0_connection` resource configured with the specified strategy, display name, and provider-specific options (database, social, SAML, OIDC, or Azure AD)
- **Connection Clients** — when `enabledClients` is configured, an `auth0_connection_clients` resource that links the connection to the specified Auth0 applications

## Prerequisites

- **Auth0 credentials** configured via environment variables (`AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`) or OpenMCF provider config
- **An Auth0 tenant** with connection quota available
- **OAuth credentials from the social provider** if using a social strategy (Google, Facebook, GitHub, etc.)
- **Identity Provider metadata** if using an enterprise strategy (SAML sign-in endpoint and certificate, OIDC issuer, or Azure AD app registration)
- **Auth0 Client application IDs** if linking the connection to specific applications via `enabledClients`

## Quick Start

Create a file `auth0-connection.yaml`:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Connection
metadata:
  name: user-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0Connection.user-db
spec:
  strategy: auth0
  databaseOptions:
    passwordPolicy: good
    bruteForceProtection: true
```

Deploy:

```shell
openmcf apply -f auth0-connection.yaml
```

This creates an Auth0 hosted database connection with a "good" password policy and brute-force protection enabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `strategy` | `string` | Identity provider strategy. Determines the authentication method and which options block applies. | Must be one of: `auth0`, `google-oauth2`, `facebook`, `github`, `linkedin`, `twitter`, `microsoft-account`, `apple`, `samlp`, `oidc`, `waad`, `ad`, `adfs` |

### Optional Fields

#### General

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown on the Auth0 Universal Login page. |
| `isDomainConnection` | `bool` | `false` | Enables identifier-first authentication flows where Auth0 discovers the connection from the user's email domain. |
| `showAsButton` | `bool` | `true` | Controls whether the connection appears as a button on the Universal Login page. When `false`, the connection is only used if explicitly requested or discovered via domain. |
| `realms` | `string[]` | `[]` | Identifiers used to route authentication requests to this connection. Defaults to the connection name if not specified. |
| `metadata` | `map<string, string>` | `{}` | Custom key-value pairs stored with the connection. Maximum 10 pairs, keys and values up to 255 characters each. |

#### Enabled Clients (`enabledClients`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabledClients` | `StringValueOrRef[]` | `[]` | Auth0 application client IDs permitted to use this connection. Only listed applications show this connection as a login option. Can reference Auth0Client resources via `valueFrom`. |

#### Database Options (`databaseOptions`)

Applicable when `strategy` is `auth0`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `databaseOptions.passwordPolicy` | `string` | `good` | Password complexity level. Values: `none`, `low` (6+ chars), `fair` (8+ with mixed case), `good` (8+ with mixed case and numeric), `excellent` (10+ with mixed case, numeric, and special). |
| `databaseOptions.requiresUsername` | `bool` | `false` | Requires a username in addition to email during signup. |
| `databaseOptions.disableSignup` | `bool` | `false` | Prevents new user signups. Useful when onboarding is done programmatically. |
| `databaseOptions.bruteForceProtection` | `bool` | `true` | Blocks login attempts after multiple failures. |
| `databaseOptions.passwordHistorySize` | `int32` | `5` | Number of previous passwords to check against. Range: 0-24. Set to 0 to disable. |
| `databaseOptions.passwordNoPersonalInfo` | `bool` | `true` | Prevents passwords containing the user's name, username, or email. |
| `databaseOptions.passwordDictionary` | `bool` | `true` | Rejects common/weak passwords found in a dictionary. |
| `databaseOptions.mfaEnabled` | `bool` | `false` | Prompts users for a second authentication factor during login. |

#### Social Options (`socialOptions`)

Applicable when `strategy` is `google-oauth2`, `facebook`, `github`, `linkedin`, `twitter`, `microsoft-account`, or `apple`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `socialOptions.clientId` | `string` | — | **(required)** OAuth client ID from the social provider's developer console. |
| `socialOptions.clientSecret` | `string` | — | **(required)** OAuth client secret from the social provider's developer console. |
| `socialOptions.scopes` | `string[]` | per strategy | OAuth scopes to request. If not specified, strategy defaults apply. |
| `socialOptions.allowedAudiences` | `string[]` | `[]` | Restricts which audiences (applications) can use the connection. Only applicable for providers that support audience restrictions. |
| `socialOptions.upstreamParams` | `map<string, string>` | `{}` | Custom parameters forwarded to the upstream provider (e.g., `login_hint`, `prompt`). |

#### SAML Options (`samlOptions`)

Applicable when `strategy` is `samlp`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `samlOptions.signInEndpoint` | `string` | — | **(required)** Identity Provider Single Sign-On URL where Auth0 sends SAML authentication requests. |
| `samlOptions.signingCert` | `string` | — | **(required)** X.509 signing certificate from the Identity Provider in PEM format. |
| `samlOptions.signOutEndpoint` | `string` | `""` | Identity Provider Single Logout URL. |
| `samlOptions.entityId` | `string` | `""` | Unique identifier (Issuer) for the Identity Provider. Derived from metadata if not specified. |
| `samlOptions.protocolBinding` | `string` | `HTTP-Redirect` | How SAML requests are sent. Values: `urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect`, `urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST`. |
| `samlOptions.userIdAttribute` | `string` | NameID claim | SAML attribute used as the user identifier. |
| `samlOptions.signRequest` | `bool` | `false` | Whether Auth0 signs outgoing SAML requests. Required by some Identity Providers. |
| `samlOptions.signatureAlgorithm` | `string` | `rsa-sha256` | Signing algorithm. Values: `rsa-sha256`, `rsa-sha1` (legacy). |
| `samlOptions.digestAlgorithm` | `string` | `sha256` | Digest algorithm. Values: `sha256`, `sha1` (legacy). |
| `samlOptions.attributeMappings` | `map<string, string>` | `{}` | Maps SAML attributes to Auth0 profile fields (e.g., `email`, `name`, `given_name`). |

#### OIDC Options (`oidcOptions`)

Applicable when `strategy` is `oidc`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `oidcOptions.issuer` | `string` | — | **(required)** OIDC issuer URL. Auth0 fetches configuration from `/.well-known/openid-configuration`. |
| `oidcOptions.clientId` | `string` | — | **(required)** OAuth client ID from the OIDC provider. |
| `oidcOptions.clientSecret` | `string` | `""` | OAuth client secret. Required for authorization code flow. |
| `oidcOptions.scopes` | `string[]` | `["openid", "profile", "email"]` | OIDC scopes to request. `openid` is always implicit. |
| `oidcOptions.type` | `string` | `front_channel` | OIDC flow type. Values: `front_channel` (authorization code), `back_channel` (token endpoint directly). |
| `oidcOptions.authorizationEndpoint` | `string` | from discovery | Override for the authorization endpoint. |
| `oidcOptions.tokenEndpoint` | `string` | from discovery | Override for the token endpoint. |
| `oidcOptions.userinfoEndpoint` | `string` | from discovery | Override for the userinfo endpoint. |
| `oidcOptions.jwksUri` | `string` | from discovery | Override for the JWKS URI. |
| `oidcOptions.attributeMappings` | `map<string, string>` | `{}` | Maps OIDC claims to Auth0 profile fields. |

#### Azure AD Options (`azureAdOptions`)

Applicable when `strategy` is `waad`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `azureAdOptions.clientId` | `string` | — | **(required)** Application (client) ID from the Azure AD app registration. |
| `azureAdOptions.clientSecret` | `string` | — | **(required)** Client secret from the Azure AD app registration. |
| `azureAdOptions.domain` | `string` | — | **(required)** Azure AD tenant domain (e.g., `contoso.onmicrosoft.com` or `contoso.com`). |
| `azureAdOptions.tenantId` | `string` | `common` | Azure AD tenant (Directory) ID. Defaults to `common`, allowing any tenant. |
| `azureAdOptions.useCommonEndpoint` | `bool` | `false` | When `true`, allows users from any Azure AD tenant (multi-tenant). |
| `azureAdOptions.maxGroupsToRetrieve` | `int32` | `50` | Maximum number of groups retrieved from Azure AD. Set to 0 for no limit. |
| `azureAdOptions.shouldTrustEmailVerified` | `bool` | `true` | Whether to trust Azure AD's `email_verified` claim. |
| `azureAdOptions.apiEnableUsers` | `bool` | `false` | Enables listing Azure AD users via the Auth0 Management API. Requires Azure AD `Directory.Read.All` permission. |

## Examples

### Database Connection with Security Defaults

A hosted database connection with password policy, brute-force protection, and MFA:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Connection
metadata:
  name: app-users
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Connection.app-users
spec:
  strategy: auth0
  displayName: Sign up with Email
  enabledClients:
    - value: "abc123clientID"
  databaseOptions:
    passwordPolicy: excellent
    bruteForceProtection: true
    passwordHistorySize: 10
    passwordNoPersonalInfo: true
    passwordDictionary: true
    disableSignup: false
    mfaEnabled: true
```

### Google Social Login

A Google OAuth 2.0 connection linked to an application managed by an Auth0Client resource:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Connection
metadata:
  name: google-social
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Connection.google-social
spec:
  strategy: google-oauth2
  displayName: Sign in with Google
  enabledClients:
    - valueFrom:
        kind: Auth0Client
        name: my-spa
  socialOptions:
    clientId: "123456789.apps.googleusercontent.com"
    clientSecret: "GOCSPX-your-google-secret"
    scopes:
      - openid
      - profile
      - email
```

### Enterprise SAML Connection

A SAML connection for enterprise single sign-on with signed requests and attribute mappings:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Connection
metadata:
  name: corporate-sso
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Connection.corporate-sso
spec:
  strategy: samlp
  displayName: Company SSO
  isDomainConnection: true
  showAsButton: true
  enabledClients:
    - value: "abc123clientID"
  samlOptions:
    signInEndpoint: "https://idp.corp.example.com/sso/saml"
    signingCert: |
      -----BEGIN CERTIFICATE-----
      MIIDpTCCAo2gAwIBAgIJAL...your-certificate...
      -----END CERTIFICATE-----
    signOutEndpoint: "https://idp.corp.example.com/slo/saml"
    entityId: "https://idp.corp.example.com"
    signRequest: true
    signatureAlgorithm: rsa-sha256
    digestAlgorithm: sha256
    attributeMappings:
      email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      given_name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
      family_name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
```

### Azure AD Enterprise Connection

An Azure AD (Entra ID) connection restricted to a single tenant with group retrieval:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Connection
metadata:
  name: azure-ad-sso
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Connection.azure-ad-sso
spec:
  strategy: waad
  displayName: Microsoft Work Account
  isDomainConnection: true
  enabledClients:
    - value: "abc123clientID"
  azureAdOptions:
    clientId: "00000000-0000-0000-0000-000000000000"
    clientSecret: "your-azure-client-secret"
    domain: contoso.onmicrosoft.com
    tenantId: "11111111-1111-1111-1111-111111111111"
    useCommonEndpoint: false
    maxGroupsToRetrieve: 100
    shouldTrustEmailVerified: true
    apiEnableUsers: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Unique Auth0 connection identifier (e.g., `con_0000000000000001`). |
| `name` | `string` | Unique name of the connection within the Auth0 tenant. Derived from `metadata.name`. |
| `strategy` | `string` | Identity provider strategy type (e.g., `auth0`, `google-oauth2`, `samlp`, `oidc`, `waad`). |
| `isEnabled` | `string` | Whether the connection is currently enabled for authentication. |
| `provisioningTicketUrl` | `string` | Self-service setup URL. Available for certain enterprise connections (SAML, OIDC). |
| `callbackUrl` | `string` | Auth0 callback URL to register with the identity provider. Format: `https://{tenant}.auth0.com/login/callback`. |
| `metadataUrl` | `string` | SAML metadata URL. Only available for SAML connections. Format: `https://{tenant}.auth0.com/samlp/metadata/{connection_name}`. |
| `entityId` | `string` | SAML Service Provider Entity ID. Only available for SAML connections. Format: `urn:auth0:{tenant}:{connection_name}`. |
| `enabledClientIds` | `string[]` | List of Auth0 application client IDs linked to this connection. |
| `realms` | `string[]` | Realms/domains associated with this connection for identifier-first flows. |

## Related Components

- [Auth0Client](/docs/catalog/auth0/auth0client) — applications that use this connection for authentication, linked via `enabledClients`
- [Auth0ResourceServer](/docs/catalog/auth0/auth0resourceserver) — APIs that authenticated users may access after signing in through this connection
- [Auth0EventStream](/docs/catalog/auth0/auth0eventstream) — streams authentication events including login activity on this connection

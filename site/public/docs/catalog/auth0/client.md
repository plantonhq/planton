---
title: "Client"
description: "Client deployment documentation"
icon: "package"
order: 100
componentName: "auth0client"
---

# Auth0 Client

Deploys an Auth0 Application (Client) with configurable OAuth flows, token settings, and optional API access grants. Supports all four Auth0 application types — native, SPA, regular web, and machine-to-machine — with full control over callbacks, refresh token behavior, JWT signing, and organization-aware authentication.

## What Gets Created

When you deploy an Auth0Client resource, OpenMCF provisions:

- **Auth0 Client (Application)** — an `auth0_client` resource configured with the specified application type, OAuth settings, URL allowlists, and optional JWT/refresh token configuration
- **Client Grants** — created only when `apiGrants` is configured, one `auth0_client_grant` resource per entry authorizing this client to call the specified API with the listed scopes

## Prerequisites

- **Auth0 credentials** configured via environment variables (`AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`) or OpenMCF provider config
- **An Auth0 tenant** with sufficient application quota
- **An Auth0 Resource Server** if configuring `apiGrants` to authorize API access
- **An Auth0 Connection** if restricting the client to specific connections via `enabledConnections`

## Quick Start

Create a file `auth0-client.yaml`:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Client
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0Client.my-app
spec:
  applicationType: spa
```

Deploy:

```shell
openmcf apply -f auth0-client.yaml
```

This creates a Single Page Application in Auth0 with default OAuth settings and OIDC-conformant behavior.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `applicationType` | `string` | The type of Auth0 application. Determines which OAuth flows and security settings apply. | Must be one of: `native`, `spa`, `regular_web`, `non_interactive` |

### Optional Fields

#### General

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Free-text description of the application. Maximum 140 characters. |
| `logoUri` | `string` | `""` | URL of the application logo, displayed on consent and login pages. |
| `clientMetadata` | `map<string, string>` | `{}` | Custom metadata key-value pairs for application-specific configuration. Maximum 10 pairs. |
| `clientAliases` | `string[]` | `[]` | Alternative identifiers for this client, usable in authentication requests instead of client_id. |

#### URL Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `callbacks` | `string[]` | `[]` | Allowed callback URLs. Auth0 redirects here after authentication. Include both development and production URLs. |
| `allowedLogoutUrls` | `string[]` | `[]` | URLs that Auth0 can redirect to after logout. |
| `webOrigins` | `string[]` | `[]` | Allowed origins for web message response mode. Required for SPAs using popup or iframe-based authentication. |
| `allowedOrigins` | `string[]` | `[]` | CORS origins allowed for cross-origin requests from JavaScript applications. |

#### OAuth & Authentication

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `grantTypes` | `string[]` | per app type | OAuth grant types this application can use. Common values: `authorization_code`, `implicit`, `refresh_token`, `client_credentials`, `password`. If not specified, defaults are based on `applicationType`. |
| `oidcConformant` | `bool` | `false` | Enables stricter OIDC-conformant behavior. Recommended for new applications. |
| `isFirstParty` | `bool` | `false` | Marks this as a first-party application. First-party apps skip the user consent prompt. |
| `crossOriginAuthentication` | `bool` | `false` | Enables cross-origin authentication for embedded login forms in SPAs. |
| `crossOriginLoc` | `string` | `""` | URL for cross-origin verification fallback. Used with `crossOriginAuthentication` for certain browsers. |
| `sso` | `bool` | `false` | Enables Single Sign-On. Users already logged in won't need to re-authenticate. |
| `ssoDisabled` | `bool` | `false` | Explicitly disables SSO, requiring authentication for each session. |
| `isTokenEndpointIpHeaderTrusted` | `bool` | `false` | When `true`, Auth0 uses the `X-Forwarded-For` header for IP-based features. |

#### Login & Organization

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `customLoginPage` | `string` | `""` | Custom HTML for the login page. Only used when `customLoginPageOn` is `true`. |
| `customLoginPageOn` | `bool` | `false` | Enables the custom login page instead of Universal Login. |
| `initiateLoginUri` | `string` | `""` | URL to initiate login for OIDC third-party initiated login flows. |
| `organizationUsage` | `string` | `""` | How organizations are used with this app. Values: `deny`, `allow`, `require`. |
| `organizationRequireBehavior` | `string` | `""` | When `organizationUsage` is `require`, determines prompt behavior. Values: `no_prompt`, `pre_login_prompt`, `post_login_prompt`. |

#### JWT Configuration (`jwtConfiguration`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `jwtConfiguration.lifetimeInSeconds` | `int32` | `36000` | JWT expiration time in seconds. Range: 0–2592000 (30 days). |
| `jwtConfiguration.alg` | `string` | `RS256` | Signing algorithm. Values: `HS256` (symmetric), `RS256` (asymmetric, recommended), `PS256`. |
| `jwtConfiguration.secretEncoded` | `bool` | `false` | Whether the client secret is base64-encoded. Only relevant when `alg` is `HS256`. |
| `jwtConfiguration.scopes` | `map<string, string>` | `{}` | Custom scopes and their descriptions available for this application. |

#### Refresh Token Configuration (`refreshToken`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `refreshToken.rotationType` | `string` | `non-rotating` | Rotation behavior. Values: `non-rotating`, `rotating` (recommended — issues new token with each use). |
| `refreshToken.expirationType` | `string` | `non-expiring` | Expiration behavior. Values: `non-expiring`, `expiring`. |
| `refreshToken.tokenLifetime` | `int32` | `2592000` | Absolute lifetime in seconds. Token expires after this time regardless of activity. Only used when `expirationType` is `expiring`. |
| `refreshToken.idleTokenLifetime` | `int32` | `1296000` | Inactivity timeout in seconds. Token expires if not used within this time. Only used when `expirationType` is `expiring`. |
| `refreshToken.infiniteTokenLifetime` | `bool` | `false` | Allows tokens to never expire. Only valid when `expirationType` is `non-expiring`. |
| `refreshToken.infiniteIdleTokenLifetime` | `bool` | `false` | Allows tokens to never expire due to inactivity. Only valid when `expirationType` is `non-expiring`. |
| `refreshToken.leeway` | `int32` | `0` | Clock skew leeway in seconds for token validation. |

#### Native Social Login (`nativeSocialLogin`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `nativeSocialLogin.apple.enabled` | `bool` | `false` | Enables Sign in with Apple native integration. Only applicable for `native` application types. |
| `nativeSocialLogin.facebook.enabled` | `bool` | `false` | Enables Facebook native login integration. Only applicable for `native` application types. |

#### Mobile Configuration (`mobile`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mobile.android.appPackageName` | `string` | `""` | Android application package name (e.g., `com.example.myapp`). |
| `mobile.android.sha256CertFingerprints` | `string[]` | `[]` | SHA-256 fingerprints of Android signing certificates for App Links and secure deep linking. |
| `mobile.ios.teamId` | `string` | `""` | Apple Developer Team ID. Required for universal links. |
| `mobile.ios.appBundleIdentifier` | `string` | `""` | iOS application bundle identifier (e.g., `com.example.myapp`). |

#### OIDC Backchannel Logout (`oidcBackchannelLogout`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `oidcBackchannelLogout.backchannelLogoutUrls` | `string[]` | `[]` | URLs to receive logout tokens. Auth0 POSTs a logout token to these URLs on logout. |

#### Connections (`enabledConnections`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabledConnections` | `StringValueOrRef[]` | `[]` | Limits which Auth0 connections this application can use. If empty, all connections are available. Can reference Auth0Connection resources via `valueFrom`. |

#### API Grants (`apiGrants`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `apiGrants[].audience` | `StringValueOrRef` | — | API identifier (Resource Server identifier) this client is authorized to access. Required per grant entry. Can reference Auth0ResourceServer resources via `valueFrom`. |
| `apiGrants[].scopes` | `string[]` | `[]` | Permissions granted for this API. If empty, the client gets access with no specific scopes. |
| `apiGrants[].allowAnyOrganization` | `bool` | `false` | Whether any organization can be used with this grant. Only relevant when using Auth0 Organizations. |
| `apiGrants[].organizationUsage` | `string` | `""` | Whether organizations can be used with client credentials exchanges. Values: `deny`, `allow`, `require`. |

## Examples

### SPA with Callback URLs

A Single Page Application with development and production callback URLs:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Client
metadata:
  name: my-spa
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0Client.my-spa
spec:
  applicationType: spa
  description: React dashboard application
  callbacks:
    - https://app.example.com/callback
    - http://localhost:3000/callback
  allowedLogoutUrls:
    - https://app.example.com
    - http://localhost:3000
  webOrigins:
    - https://app.example.com
    - http://localhost:3000
  grantTypes:
    - authorization_code
    - refresh_token
  oidcConformant: true
  isFirstParty: true
```

### Machine-to-Machine with API Grants

A backend service client authorized to call a custom API and the Auth0 Management API:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Client
metadata:
  name: backend-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Client.backend-service
spec:
  applicationType: non_interactive
  description: Backend API service
  grantTypes:
    - client_credentials
  apiGrants:
    - audience: https://api.example.com/
      scopes:
        - read:resources
        - write:resources
    - audience: https://my-tenant.us.auth0.com/api/v2/
      scopes:
        - read:users
        - read:user_idp_tokens
```

### Full-Featured Web Application

A server-side web application with JWT configuration, rotating refresh tokens, and organization support:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Client
metadata:
  name: web-portal
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Client.web-portal
spec:
  applicationType: regular_web
  description: Customer portal web application
  callbacks:
    - https://portal.example.com/auth/callback
  allowedLogoutUrls:
    - https://portal.example.com
  grantTypes:
    - authorization_code
    - refresh_token
  oidcConformant: true
  isFirstParty: true
  sso: true
  organizationUsage: allow
  jwtConfiguration:
    lifetimeInSeconds: 3600
    alg: RS256
  refreshToken:
    rotationType: rotating
    expirationType: expiring
    tokenLifetime: 2592000
    idleTokenLifetime: 1296000
```

### Using Foreign Key References

Reference other OpenMCF-managed Auth0 resources instead of hardcoding identifiers:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Client
metadata:
  name: ref-client
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0Client.ref-client
spec:
  applicationType: non_interactive
  description: Service client with resource references
  grantTypes:
    - client_credentials
  enabledConnections:
    - valueFrom:
        kind: Auth0Connection
        name: my-db-connection
        field: status.outputs.name
  apiGrants:
    - audience:
        valueFrom:
          kind: Auth0ResourceServer
          name: my-api
          field: status.outputs.identifier
      scopes:
        - read:data
        - write:data
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Internal Auth0 identifier for the client |
| `client_id` | `string` | OAuth 2.0 client identifier. Safe to expose in client-side code. |
| `client_secret` | `string` | OAuth 2.0 client secret. Only available for `regular_web` and `non_interactive` application types. Keep secure — never expose in client-side code. |
| `name` | `string` | Name of the application, derived from `metadata.name` |
| `application_type` | `string` | Application type (`native`, `spa`, `regular_web`, `non_interactive`) |
| `signing_keys` | `Auth0SigningKey[]` | Signing keys for RS256 token signature verification. Each key contains `cert`, `pkcs7`, `subject`, and `thumbprint` fields. |
| `callback_url_template` | `string` | Whether callback URL templating is enabled |
| `allowed_clients` | `string[]` | Clients allowed to perform delegation for this client |
| `global` | `string` | Whether this is the tenant's global (default) client |
| `token_endpoint_auth_method` | `string` | Authentication method for the token endpoint (e.g., `none`, `client_secret_post`, `client_secret_basic`) |

## Related Components

- [Auth0ResourceServer](/docs/catalog/auth0/auth0resourceserver) — defines the APIs that this client can be authorized to access via `apiGrants`
- [Auth0Connection](/docs/catalog/auth0/auth0connection) — provides authentication connections that can be linked via `enabledConnections`
- [Auth0EventStream](/docs/catalog/auth0/auth0eventstream) — streams authentication events from the Auth0 tenant

# AWS Cognito User Pool Client

Deploy and manage a Cognito user pool app client using Planton -- the OAuth 2.0 / OIDC contract between ONE application and a user pool, covering grant types, redirect URLs, token lifetimes, authentication flows, refresh-token rotation, and Pinpoint analytics.

## Overview

An app client is deliberately its own resource rather than a field on the pool:

- **Many per pool** -- a pool serves a web frontend, a mobile app, and machine-to-machine services, each with its own OAuth contract.
- **Referenced by ID** -- the client ID is what downstream systems consume: an API Gateway JWT authorizer lists it as an audience, an ALB `authenticate_cognito` action takes it as `user_pool_client_id`, and application configs embed it for SDK sign-in.
- **Own lifecycle** -- clients are added, rotated, and retired without touching the pool or its users.

## When to Use

- Every application that authenticates against an `AwsCognitoUserPool` needs exactly one client.
- A public client (SPA, mobile) authenticating with PKCE -- `generateSecret: false`.
- A confidential client (server-side app) holding a secret -- `generateSecret: true`.
- A machine-to-machine client using the `client_credentials` grant with custom scopes from an `AwsCognitoResourceServer`.

## Prerequisites

- An `AwsCognitoUserPool` (referenced by `userPoolId`).
- (Optional) `AwsCognitoIdentityProvider` resources if the client offers federated sign-in.
- (Optional) An `AwsCognitoResourceServer` if the client requests custom scopes.

## ForceNew Fields (Cannot Change After Creation)

- `userPoolId` -- a client cannot move between pools.
- `generateSecret` -- confidential vs public is decided at creation.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPoolClient
metadata:
  name: web-app
  org: my-org
  env: prod
  id: web-app-prod
spec:
  region: us-east-1
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: my-auth
      fieldPath: status.outputs.user_pool_id
  explicitAuthFlows:
    - ALLOW_USER_SRP_AUTH
    - ALLOW_REFRESH_TOKEN_AUTH
  preventUserExistenceErrors: ENABLED
```

## Spec Reference

### Identity

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | `string` | Yes | AWS region. |
| `userPoolId` | `StringValueOrRef` | Yes | The pool this client authenticates against. ForceNew. |
| `generateSecret` | `bool` | No | Generate a client secret (confidential clients). ForceNew. |

### OAuth 2.0 / OIDC

| Field | Type | Description |
|-------|------|-------------|
| `allowedOauthFlowsUserPoolClient` | `bool` | Master switch for the OAuth fields below. |
| `allowedOauthFlows` | `string[]` | `code`, `implicit` (legacy), `client_credentials` (M2M; cannot mix with user-facing grants). |
| `allowedOauthScopes` | `string[]` | Built-ins (`openid`, `email`, `profile`, `phone`, `aws.cognito.signin.user.admin`) plus resource-server scopes (`{identifier}/{scope}`). |
| `callbackUrls` / `logoutUrls` | `string[]` | OAuth redirect targets (max 100 each). |
| `defaultRedirectUri` | `string` | Must be one of `callbackUrls`. |
| `supportedIdentityProviders` | `StringValueOrRef[]` | `COGNITO` for the pool's directory, plus IdP names -- literals or references to `AwsCognitoIdentityProvider` (which also orders the deployment graph). |

### Authentication flows

| Field | Type | Description |
|-------|------|-------------|
| `explicitAuthFlows` | `string[]` | `ALLOW_USER_SRP_AUTH`, `ALLOW_REFRESH_TOKEN_AUTH`, `ALLOW_USER_PASSWORD_AUTH`, `ALLOW_ADMIN_USER_PASSWORD_AUTH`, `ALLOW_CUSTOM_AUTH`, `ALLOW_USER_AUTH` (choice-based/passwordless), or the legacy pre-ALLOW spellings (unmixed). |
| `authSessionValidity` | `int32` | Challenge-handshake session validity, 3-15 minutes. |

### Token lifetimes

| Field | Type | Description |
|-------|------|-------------|
| `accessTokenValidity` / `idTokenValidity` | `int32` | 5 minutes - 24 hours in the configured unit (default hours). |
| `refreshTokenValidity` | `int32` | 60 minutes - 10 years in the configured unit (default days). |
| `tokenValidityUnits` | `object` | `seconds` / `minutes` / `hours` / `days` per token type. |
| `refreshTokenRotation.feature` | `string` | `ENABLED` / `DISABLED` -- each refresh issues a NEW refresh token. When ENABLED, `ALLOW_REFRESH_TOKEN_AUTH` must not be listed in `explicitAuthFlows` (AWS rejects the combination). |
| `refreshTokenRotation.retryGracePeriodSeconds` | `optional int32` | 0-60s the retired token keeps working after rotation. Tri-state: unset leaves the choice to AWS, an EXPLICIT 0 pins immediate retirement (grace disabled), 1-60 grants that many seconds of grace. |

### Security posture

| Field | Type | Description |
|-------|------|-------------|
| `enableTokenRevocation` | `bool` | Sign-out revokes tokens. AWS default: true. |
| `enablePropagateAdditionalUserContextData` | `bool` | Forward client IP/user-agent to threat protection in server-side flows. Requires a secret. |
| `preventUserExistenceErrors` | `string` | `ENABLED` (recommended -- blocks user enumeration) or `LEGACY`. |
| `riskConfiguration` | `object` | CLIENT-SCOPED threat-protection responses, overriding the pool-wide policy for this client only: per-risk-level account-takeover actions with notification templates (SES identity ref), compromised-credential blocking with event filters, and IP allow/block exceptions. Requires the POOL's threat protection to be AUDIT/ENFORCED -- AWS rejects it otherwise. |

### Attribute access and analytics

| Field | Type | Description |
|-------|------|-------------|
| `readAttributes` / `writeAttributes` | `string[]` | Attribute-level access for this client. Empty = all. |
| `analyticsConfiguration` | `object` | Pinpoint wiring: `applicationArn` XOR (`applicationId` + `externalId` + `roleArn` ref). |

## Outputs

| Output | Description |
|--------|-------------|
| `client_id` | The app client ID -- the JWT `aud` claim, the ALB `user_pool_client_id`, the SDK client identifier. |
| `client_secret` | The client secret; populated only when `generateSecret: true`. Sensitive -- treat as a credential. |
| `user_pool_id` | The pool this client belongs to, resolved from the spec reference -- app configs typically need the (pool id, client id) pair together. |

## Composing

```yaml
# API Gateway JWT authorizer audience:
audiences:
  - valueFrom:
      kind: AwsCognitoUserPoolClient
      name: web-app
      fieldPath: status.outputs.client_id
```

## Deliberately Omitted

- **`aws_cognito_managed_user_pool_client`**: adopts clients CREATED BY other AWS services (ELB, AppSync) by name pattern, with a no-op delete -- an adoption lifecycle that will be its own kind, never a mode of this one.
- **Per-kind tags**: app clients are not taggable in AWS.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

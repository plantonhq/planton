# AWS Cognito User Pool Client

Deploys a Cognito User Pool app client — the OAuth 2.0 / OIDC contract between ONE application and a user pool. Each application surface (web SPA, mobile app, M2M service) gets its own client with its own grant types, redirect URLs, token lifetimes, and client ID. Downstream systems reference that client ID (API Gateway JWT authorizer audiences, ALB authenticate-cognito actions, SDK configs), which is why the client is its own resource rather than a field on the pool.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cognito User Pool Client** — registered on the target user pool with the configured OAuth flows, scopes, redirect URLs, auth flows, token lifetimes, attribute access, and optional Pinpoint analytics. App clients are not taggable in AWS.
- **Client Secret** — minted only when `generateSecret` is true (confidential / M2M clients); public clients authenticate with PKCE instead.
- **Client-Scoped Risk Configuration** — created only when `riskConfiguration` is set; overrides the pool-wide threat-protection response policy for this client only (requires the pool's threat protection to be active).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **A Cognito User Pool** — the directory this application authenticates against. Reference an AwsCognitoUserPool Infra Component or provide the pool ID directly.
- **Identity providers** (only for federated sign-in) — list providers by name or reference AwsCognitoIdentityProvider resources in `supportedIdentityProviders`.
- **Resource server scopes** (only for custom scopes) — an AwsCognitoResourceServer minting the `identifier/scope-name` strings this client will request; mandatory for the `client_credentials` grant.

## Deploy

### Console

Open the deployment store, find **AWS Cognito User Pool Client**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the pool reference, OAuth posture, and token lifetimes. Start from the **Web SPA** preset in the [Presets](#presets) tab for public clients, **Machine-to-Machine (Client Credentials)** for machine grants, or **Server-Side Confidential** for server-side apps holding a secret.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoUserPoolClient
metadata:
  name: web-spa-client
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: app-auth
      fieldPath: status.outputs.user_pool_id
  allowedOauthFlowsUserPoolClient: true
  allowedOauthFlows:
    - code
  allowedOauthScopes:
    - openid
    - email
    - profile
  callbackUrls:
    - https://app.acme-corp.com/callback
  logoutUrls:
    - https://app.acme-corp.com/logout
  supportedIdentityProviders:
    - value: COGNITO
  explicitAuthFlows:
    - ALLOW_USER_SRP_AUTH
    - ALLOW_REFRESH_TOKEN_AUTH
  enableTokenRevocation: true
  preventUserExistenceErrors: ENABLED
```

```shell
planton apply -f cognito-user-pool-client.yaml
```

This creates a public SPA client on the referenced pool: Authorization Code grant with the standard OIDC scopes, SRP sign-in against the pool's own directory, refresh tokens, revocation on sign-out, and user-enumeration protection. An Infra Job tracks the provisioning in real time.

### InfraChart

When the client deploys alongside its pool and a federated provider in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: app-auth
      fieldPath: status.outputs.user_pool_id
  allowedOauthFlowsUserPoolClient: true
  allowedOauthFlows:
    - code
  allowedOauthScopes:
    - openid
    - email
  callbackUrls:
    - https://app.acme-corp.com/callback
  supportedIdentityProviders:
    - value: COGNITO
    - valueFrom:
        kind: AwsCognitoIdentityProvider
        name: corp-okta
        fieldPath: status.outputs.provider_name
```

The InfraPipeline resolves the dependency graph — pool, then provider, then client — so the provider exists before the client lists it. Custom resource-server scopes go into `allowedOauthScopes` using the resource server's `scope_identifiers` output values.

## Key Configuration

These are the most important decisions when configuring an app client. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Confidential or public is decided at birth** — `generateSecret` is create-only. On for server-side and M2M clients that can protect a secret; off for SPAs and mobile apps, which authenticate with PKCE. Changing your mind later means a new client and a new client ID everywhere downstream.

**One client per application surface, and M2M stands alone** — Authorization Code (`code`) is the modern default for anything user-facing; `implicit` is legacy (tokens leak into browser history). The `client_credentials` grant cannot share a client with user-facing grants — AWS rejects the mix — and it can request ONLY custom resource-server scopes, since the built-in OIDC scopes describe a user that doesn't exist in a machine flow.

**Refresh token rotation changes the rules** — with `refreshTokenRotation.feature: ENABLED`, each refresh issues a new token and retires the old one, shrinking a stolen token's blast radius. Drop `ALLOW_REFRESH_TOKEN_AUTH` from `explicitAuthFlows` when you enable it — rotation owns the refresh behavior and AWS rejects the combination. `retryGracePeriodSeconds` is tri-state: unset keeps AWS's default, an explicit `0` pins immediate retirement, 1–60 grants that many seconds for clients that lose the response carrying the new token.

**Token lifetimes are a value-unit pair** — `accessTokenValidity`/`idTokenValidity` (5 minutes to 24 hours) and `refreshTokenValidity` (60 minutes to 10 years) are interpreted in `tokenValidityUnits` (defaults: hours, hours, days). Shorter access tokens shrink the exposure window at the cost of more refresh traffic; the spec validates the bounds before AWS sees them.

**Keep `preventUserExistenceErrors: ENABLED`** — it returns the same error for wrong-password and nonexistent-user attempts, defeating user enumeration. LEGACY exists for backward compatibility only.

**Client-scoped threat protection has a cross-resource prerequisite** — `riskConfiguration` (account-takeover actions per risk level, compromised-credential response, IP allow/block exceptions) overrides the pool-wide policy for this client only, and requires the pool's threat protection (`userPoolAddOns.advancedSecurityMode` of AUDIT or ENFORCED) to be active — a requirement AWS enforces at apply, not the spec. For server-side flows, `enablePropagateAdditionalUserContextData` forwards the real end-user IP and user agent to threat protection (requires a client secret); without it, Cognito risk-scores your server's IP.

**Omitting `supportedIdentityProviders` enables everything** — AWS turns on ALL of the pool's providers for the client. List `COGNITO` and the intended federated providers explicitly so adding a provider to the pool later doesn't silently appear on this client's sign-in page.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCognitoUserPool** | `userPoolId` | `status.outputs.user_pool_id` |
| **AwsCognitoIdentityProvider** | `supportedIdentityProviders` | `status.outputs.provider_name` |
| **AwsIamRole** | `analyticsConfiguration.roleArn` | `status.outputs.role_arn` |
| **AwsSesEmailIdentity** | `riskConfiguration.accountTakeover.notifyConfiguration.sourceArn` | `status.outputs.identity_arn` |

Custom resource-server scopes travel as plain strings in `allowedOauthScopes` — copy them from the resource server's `scope_identifiers` output.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `client_id` | The app client ID — the public identifier and the `aud` claim | API Gateway JWT authorizer audiences, ALB authenticate-cognito actions, application SDK configs |
| `client_secret` | The client secret (sensitive; populated only when a secret was generated) | Confidential clients' token-endpoint authentication, wired as a secret |
| `user_pool_id` | The pool this client belongs to, resolved from the reference | Application configs needing the (pool id, client id) pair together |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public SPA with PKCE** — no secret, Authorization Code grant, SRP auth (the password never leaves the device), standard OIDC scopes, revocation and enumeration protection on. Add rotation later without replacing the client. Start from the **Web SPA** preset.

**Machine-to-machine per service** — `client_credentials` with a generated secret and custom resource-server scopes, short access-token lifetimes. Give each calling service its OWN client so each holds its own credential and can be revoked independently — sharing one M2M client across services makes revocation an outage. Start from the **Machine-to-Machine (Client Credentials)** preset.

**Server-side confidential** — Authorization Code with a secret, rotating refresh tokens with a small retry grace window, and end-user context propagated to threat protection so risk scoring sees the browser, not your backend. Start from the **Server-Side Confidential** preset.

## Works With

- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) — the user directory this client authenticates against, wired via the `userPoolId` reference
- [**AWS Cognito Identity Provider**](/infra-catalog/aws-cognito-identity-provider) — federated providers listed in `supportedIdentityProviders`
- [**AWS Cognito Resource Server**](/infra-catalog/aws-cognito-resource-server) — mints the custom OAuth scopes this client requests
- [**AWS SES Email Identity**](/infra-catalog/aws-ses-email-identity) — the sending identity for threat-protection notification emails

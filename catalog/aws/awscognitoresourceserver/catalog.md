# AWS Cognito Resource Server

Deploys a Cognito Resource Server — an OAuth 2.0 scope namespace attached to a user pool. Each scope becomes an `identifier/scope-name` string that app clients request and your APIs enforce. The resource server does not host users; it declares the permissions tokens may carry — and for machine-to-machine clients using the `client_credentials` grant, these custom scopes are the ONLY scopes they can request, making this component the prerequisite for any M2M authorization on the pool.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cognito Resource Server** — registered on the target user pool with a permanent identifier, a display name, and the configured OAuth scopes. Resource servers are not taggable in AWS, so this is the single resource the module manages.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **A Cognito User Pool** — the directory whose access tokens will carry these scopes. Reference an AwsCognitoUserPool Infra Component or provide the pool ID (`{region}_{poolId}`) directly.

## Deploy

### Console

Open the deployment store, find **AWS Cognito Resource Server**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the pool reference, identifier, and scopes. Start from the **API Scopes** preset in the [Presets](#presets) tab for a working read/write scope vocabulary.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoResourceServer
metadata:
  name: orders-api-scopes
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: app-auth
      fieldPath: status.outputs.user_pool_id
  identifier: https://api.acme-corp.com
  name: Orders API
  scopes:
    - scopeName: orders.read
      scopeDescription: Read order history
    - scopeName: orders.write
      scopeDescription: Create and update orders
```

```shell
planton apply -f cognito-resource-server.yaml
```

This registers a resource server on the referenced pool minting two requestable scopes — `https://api.acme-corp.com/orders.read` and `https://api.acme-corp.com/orders.write`. An Infra Job tracks the provisioning in real time.

### InfraChart

When the resource server deploys alongside its pool in one chart, wire the pool reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: app-auth
      fieldPath: status.outputs.user_pool_id
  identifier: https://api.acme-corp.com
  name: Orders API
  scopes:
    - scopeName: orders.read
      scopeDescription: Read order history
```

The InfraPipeline resolves the dependency graph, deploys the pool first, then registers the resource server on it — and app clients in the same chart can consume the emitted `scope_identifiers`.

## Key Configuration

These are the most important decisions when configuring a resource server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The identifier is forever** — `identifier` is the resource server's identity within the pool and the prefix of every scope it mints; changing it forces replacement, and every client requesting the old scope strings breaks. The convention worth following: use the API's audience URI (`https://api.acme-corp.com`), which makes token audiences self-describing.

**The pool binding is forever too** — `userPoolId` is create-only; a resource server cannot move between pools. Deploy one resource server per API per pool.

**Scopes update in place, with a token-lifetime tail** — adding scopes is safe any time. Removing a scope invalidates it for FUTURE tokens only: already-issued tokens carry it until they expire, so your API must keep tolerating (or actively rejecting) a removed scope for one token lifetime after the removal.

**Scope naming has a reserved character** — `/` separates identifier from scope name in the minted string, so it cannot appear inside a scope name (nor can spaces, `"`, or `\`). Dotted (`orders.read`) and coloned (`orders:write`) names both work; pick one convention per API and hold it — clients hard-code these strings.

**This is the M2M prerequisite** — a user pool client using the `client_credentials` grant can request only custom scopes, never the standard OpenID ones. No resource server means no requestable scopes means no working M2M client on the pool.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCognitoUserPool** | `userPoolId` | `status.outputs.user_pool_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `scope_identifiers` | The fully-qualified scope strings (`{identifier}/{scope_name}`) this server mints | The exact values an AwsCognitoUserPoolClient lists in `allowedOauthScopes` |
| `resource_server_identifier` | The identifier — the scope prefix access tokens carry | API-side token validation configuration (the audience/prefix to check) |
| `user_pool_id` | The pool this server belongs to, resolved from the reference | Consumers holding only this resource get both halves of the (pool, identifier) key AWS uses |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One resource server per API** — each API the pool protects gets its own identifier and scope vocabulary, so permission models evolve per API without cross-talk. The scope list starts coarse (`read`, `write`) and grows finer (`orders:read`, `admin`) as the API's authorization model matures — additions are free of migration cost, removals carry the token-lifetime tail. Start from the **API Scopes** preset.

**The M2M pair** — this resource server plus an AwsCognitoUserPoolClient with `allowedOauthFlows: [client_credentials]` listing the minted scopes. The client authenticates with its secret, receives an access token carrying exactly these scopes, and your API authorizes on them — no user in the loop.

## Works With

- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) — the user directory this resource server attaches to, wired via the `userPoolId` reference
- [**AWS Cognito User Pool Client**](/infra-catalog/aws-cognito-user-pool-client) — the app clients that request these scopes in their allowed OAuth scopes
- [**AWS Cognito Identity Provider**](/infra-catalog/aws-cognito-identity-provider) — federated sign-in providers on the same pool

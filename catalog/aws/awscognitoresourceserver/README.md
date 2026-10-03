# AWS Cognito Resource Server

Deploy and manage a Cognito resource server using Planton -- the OAuth 2.0 resource (an API) a user pool mints custom access-token scopes for.

## Overview

A resource server names an API (by an identifier, conventionally its audience URI) and defines the scope vocabulary access tokens can carry for it. Each scope becomes requestable by app clients as `{identifier}/{scope_name}`.

It is deliberately its own resource:

- **Many per pool** -- a pool protects many APIs, each with its own identifier and scopes.
- **Referenced by clients** -- the scope identifiers it mints are exactly what `AwsCognitoUserPoolClient.allowedOauthScopes` lists; machine-to-machine clients using the `client_credentials` grant can request ONLY these custom scopes.
- **Own lifecycle** -- scopes evolve with the API, not with the pool.

## When to Use

- Machine-to-machine authorization: a `client_credentials` client needs at least one custom scope to request.
- Fine-grained API permissions in access tokens (e.g. `https://api.example.com/orders:write`) that your API authorizes on.

## Prerequisites

- An `AwsCognitoUserPool` (referenced by `userPoolId`).

## ForceNew Fields (Cannot Change After Creation)

- `userPoolId` -- a resource server cannot move between pools.
- `identifier` -- the resource server's identity within the pool.

Scopes update in place; removing a scope invalidates it for future tokens (already-issued tokens carry it until they expire).

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoResourceServer
metadata:
  name: orders-api
  org: my-org
  env: prod
  id: orders-api-prod
spec:
  region: us-east-1
  userPoolId:
    valueFrom:
      kind: AwsCognitoUserPool
      name: my-auth
      fieldPath: status.outputs.user_pool_id
  identifier: https://api.example.com
  name: Orders API
  scopes:
    - scopeName: read
      scopeDescription: Read orders
    - scopeName: write
      scopeDescription: Create and update orders
```

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | `string` | Yes | AWS region. |
| `userPoolId` | `StringValueOrRef` | Yes | The pool this resource server belongs to. ForceNew. |
| `identifier` | `string` | Yes | Unique identifier within the pool -- the scope prefix access tokens carry. Conventionally the API's audience URI. 1-256 chars. ForceNew. |
| `name` | `string` | Yes | Display name in the Cognito console. 1-256 chars. |
| `scopes[].scopeName` | `string` | Yes | Scope name (no spaces, `/`, `"`, or `\`). Max 100 scopes. |
| `scopes[].scopeDescription` | `string` | Yes | Shown on consent screens. |

## Outputs

| Output | Description |
|--------|-------------|
| `resource_server_identifier` | The identifier -- the scope prefix in access tokens. |
| `scope_identifiers` | The fully-qualified scope strings (`{identifier}/{scope_name}`) app clients list in `allowedOauthScopes`. |
| `user_pool_id` | The pool this resource server belongs to, resolved from the spec reference. |

## Deliberately Omitted

- **Per-kind tags**: resource servers are not taggable in AWS.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

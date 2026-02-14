---
title: "Resource Server"
description: "Resource Server deployment documentation"
icon: "package"
order: 100
componentName: "auth0resourceserver"
---

# Auth0 Resource Server

Deploys an Auth0 Resource Server (API) with configurable token settings, scope definitions, and optional RBAC policy enforcement. Resource Servers define the APIs that applications request access to via the OAuth 2.0 `audience` parameter.

## What Gets Created

When you deploy an Auth0ResourceServer resource, OpenMCF provisions:

- **Auth0 Resource Server** — an `auth0_resource_server` resource representing the API, configured with the specified identifier, token lifetime, signing algorithm, and access control settings
- **Resource Server Scopes** — created only when `scopes` is configured, an `auth0_resource_server_scopes` resource defining the permissions available for this API

## Prerequisites

- **Auth0 credentials** configured via environment variables (`AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`) or OpenMCF provider config
- **An Auth0 tenant** with sufficient API quota
- **A unique API identifier** (audience URI) that has not already been registered in the tenant — identifiers cannot be changed after creation

## Quick Start

Create a file `auth0-api.yaml`:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0ResourceServer
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0ResourceServer.my-api
spec:
  identifier: https://api.example.com/
```

Deploy:

```shell
openmcf apply -f auth0-api.yaml
```

This creates a Resource Server in Auth0 with the identifier `https://api.example.com/`, using default token settings (RS256 signing, 86400-second token lifetime).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `identifier` | `string` | Unique identifier for the API, used as the `audience` parameter in authorization requests. Typically a URI (e.g., `https://api.example.com/`). Cannot be changed after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Friendly display name for the API. Shown in the Auth0 dashboard and consent screens. Cannot include `<` or `>` characters. |
| `signingAlg` | `string` | `RS256` | Algorithm used to sign access tokens. Values: `RS256` (asymmetric, recommended), `HS256` (symmetric, requires client secret), `PS256`. |
| `allowOfflineAccess` | `bool` | `false` | Whether refresh tokens can be issued for this API. When `true`, applications can request refresh tokens using the `offline_access` scope. |
| `tokenLifetime` | `int32` | `86400` | Duration in seconds that access tokens remain valid when issued from the token endpoint. Range: 0–2592000 (30 days). |
| `tokenLifetimeForWeb` | `int32` | `7200` | Duration in seconds that access tokens remain valid when issued via implicit or hybrid flows. Cannot exceed `tokenLifetime`. Range: 0–2592000 (30 days). |
| `skipConsentForVerifiableFirstPartyClients` | `bool` | `true` | Whether to skip the consent prompt for first-party applications. When `true`, first-party apps don't show the consent screen. |
| `enforcePolicies` | `bool` | `false` | Enables RBAC authorization policies. When `true`, role and permission assignments are evaluated during login. Requires `tokenDialect` set to an `_authz` variant to include permissions in tokens. |
| `tokenDialect` | `string` | `access_token` | Format of access tokens issued for this API. Values: `access_token`, `access_token_authz` (includes RBAC permissions), `rfc9068_profile` (IETF-compliant), `rfc9068_profile_authz` (IETF with RBAC). Use `_authz` variants when `enforcePolicies` is `true`. |
| `scopes` | `Auth0ResourceServerScope[]` | `[]` | Permissions that can be granted for this API. Each scope has a `name` (required) and optional `description`. |
| `scopes[].name` | `string` | — | Scope identifier used in OAuth flows (e.g., `read:users`, `write:orders`). Appears in access tokens. Required per scope entry. |
| `scopes[].description` | `string` | `""` | Human-readable explanation of what this scope grants. Shown on consent screens and in the Auth0 dashboard. |

## Examples

### API with Scopes

A backend API with defined read and write permissions:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0ResourceServer
metadata:
  name: orders-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.Auth0ResourceServer.orders-api
spec:
  identifier: https://api.example.com/orders
  name: Orders API
  signingAlg: RS256
  tokenLifetime: 86400
  allowOfflineAccess: true
  scopes:
    - name: read:orders
      description: Read access to orders
    - name: write:orders
      description: Create and update orders
    - name: delete:orders
      description: Delete orders
```

### RBAC-Enabled API

An API with role-based access control that includes permissions in access tokens:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0ResourceServer
metadata:
  name: admin-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0ResourceServer.admin-api
spec:
  identifier: https://api.example.com/admin
  name: Admin API
  signingAlg: RS256
  tokenLifetime: 3600
  tokenLifetimeForWeb: 1800
  skipConsentForVerifiableFirstPartyClients: true
  enforcePolicies: true
  tokenDialect: access_token_authz
  scopes:
    - name: read:users
      description: Read user profiles
    - name: write:users
      description: Create and update users
    - name: delete:users
      description: Delete users
    - name: read:roles
      description: Read role assignments
    - name: manage:roles
      description: Create, update, and delete roles
```

### Production API with Short-Lived Tokens

A production API with restricted token lifetimes and no offline access:

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0ResourceServer
metadata:
  name: payments-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.Auth0ResourceServer.payments-api
spec:
  identifier: https://api.example.com/payments
  name: Payments API
  signingAlg: RS256
  tokenLifetime: 900
  tokenLifetimeForWeb: 300
  allowOfflineAccess: false
  enforcePolicies: true
  tokenDialect: rfc9068_profile_authz
  scopes:
    - name: read:transactions
      description: View transaction history
    - name: create:payments
      description: Initiate new payments
    - name: manage:refunds
      description: Process refunds
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Internal Auth0 identifier for the resource server |
| `identifier` | `string` | API identifier (audience) for this resource server, same as the value specified in `spec.identifier` |
| `name` | `string` | Display name of the resource server |
| `signing_alg` | `string` | Algorithm used to sign tokens (`RS256`, `HS256`, or `PS256`) |
| `signing_secret` | `string` | Secret used for signing tokens. Only populated when `signingAlg` is `HS256`. Keep secure. |
| `token_lifetime` | `string` | Configured token validity duration in seconds |
| `token_lifetime_for_web` | `string` | Token validity for implicit/hybrid flows in seconds |
| `allow_offline_access` | `string` | Whether refresh tokens can be issued |
| `skip_consent_for_verifiable_first_party_clients` | `string` | Whether consent is skipped for first-party apps |
| `enforce_policies` | `string` | Whether RBAC is enabled for this API |
| `token_dialect` | `string` | Access token format configured for this API |
| `is_system` | `string` | Whether this is a system-managed resource server (e.g., Auth0 Management API) |
| `client_id` | `string` | Associated client ID, if one has been linked |

## Related Components

- [Auth0Client](/docs/catalog/auth0/auth0client) — applications that request access to this API via `apiGrants`
- [Auth0Connection](/docs/catalog/auth0/auth0connection) — authentication connections used by clients accessing this API
- [Auth0EventStream](/docs/catalog/auth0/auth0eventstream) — streams authentication and authorization events from the Auth0 tenant

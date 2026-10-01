# Auth0 Resource Server (API)

Deploys an Auth0 Resource Server — the API definition that applications request access tokens for, carrying the audience identifier, token signing and lifetime settings, scope definitions, optional RBAC policy enforcement, the access policy that decides which applications may get a token at all, and the default grants every third-party application gets. The identifier doubles as the OAuth 2.0 `audience` parameter and is permanent: it cannot be changed after creation. Scopes managed here are authoritative — each apply sets the API's complete scope list.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Auth0 Resource Server** — the API registered in the tenant with the specified identifier (audience), signing algorithm, token lifetimes, and access-control settings
- **Resource Server Scopes** — created only when `scopes` is non-empty; an authoritative scope set that makes the API's permission list exactly match the spec on every apply
- **Default Grants for Third-Party Applications** — one client grant per `thirdPartyClientDefaultGrants` entry (one per subject type), naming no application: every third-party application gets its scopes on this API

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** — an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **Management API scopes** — the M2M application behind the Provider Connection needs `create:resource_servers`, `read:resource_servers`, `update:resource_servers`, and `delete:resource_servers`. The scopes resource rides the same four. Default grants for third-party applications add `create:client_grants`, `read:client_grants`, `update:client_grants`, and `delete:client_grants`.
- **A plan that includes the advanced token settings you use** — token encryption, mTLS proof of possession and transactional authorization need the Enterprise plan with the Highly Regulated Identity add-on; anonymous-session settings are an Enterprise Early Access feature. Everything else works on every plan.
- **A unique API identifier** (audience URI) not already registered in the tenant. Identifiers cannot be changed after creation, so choose a stable URI like `https://api.example.com/` before deploying.

## Deploy

### Console

Open the deployment store, find **Auth0 Resource Server (API)**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the identifier and token settings, and an inline builder for the scope list.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0ResourceServer
metadata:
  name: backend-api
  org: acme-corp
  env: prod
spec:
  identifier: https://api.example.com/
  scopes:
    - name: read:data
      description: Read access to data
    - name: write:data
      description: Write access to data
```

```shell
planton apply -f auth0-resource-server.yaml
```

This creates an API with RS256 signing, Auth0's default 24-hour token lifetime, and two requestable scopes — no RBAC enforcement. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Auth0 Resource Server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Identifier (audience)** — The `identifier` is what applications send as the OAuth `audience` parameter, and it is permanent: it cannot be changed after creation, so renaming means creating a new Resource Server and migrating every client and role that references the old audience. Choose a stable URI like `https://api.example.com/` up front.

**Signing algorithm** — `signingAlg` defaults to RS256, where APIs validate tokens against Auth0's public JWKS endpoint and keys rotate without coordination. Choosing HS256 turns the API into a secret-holder: every consumer must store the shared `signing_secret`, and rotation requires coordinated updates between Auth0 and all of them. Use RS256 unless something specific forces HS256.

**Token lifetimes** — `tokenLifetime` (default 86400 seconds / 24 hours) bounds how long a stolen access token stays usable; for sensitive APIs, reduce it to minutes (e.g., 900). `tokenLifetimeForWeb` covers implicit and hybrid flows, should be shorter, and cannot exceed `tokenLifetime`.

**RBAC enforcement** — `enforcePolicies: true` makes Auth0 evaluate role and permission assignments at login. On its own it only evaluates: to make permissions land in the token, pair it with a `tokenDialect` ending in `_authz`. Without that pairing, backends see RBAC-filtered scopes but no permissions claim.

**Token dialect** — `tokenDialect` picks between two base formats — `access_token` (standard Auth0 JWT) and `rfc9068_profile` (IETF JWT Access Token Profile, for ecosystems that require the standard `at+jwt` shape) — and the `_authz` suffix on either variant embeds RBAC permissions in the token payload.

**Who can get a token** — `subjectTypeAuthorization` sets the API's access policy separately for tokens that act for a person (`user`: `allow_all`, `require_client_grant`, `deny_all`) and tokens an application gets for itself (`client`: `require_client_grant`, `deny_all`). `require_client_grant` is Auth0's least-privilege recommendation, but it locks out every application without a grant, so deploy the grants with it. Third-party applications need a grant under every policy.

**Default grants for third-party applications** — `thirdPartyClientDefaultGrants` gives every third-party application the same scopes on this API without a grant of its own: the way to open an API to partners, AI agents and MCP clients that register themselves through Dynamic Client Registration or a Client ID Metadata Document. A grant made for one application (Auth0 Application `apiGrants`) takes precedence.

**Adopting an existing API** — Unset means unmanaged, so an imported API keeps what it has for everything the spec leaves out, except `allowOfflineAccess`, `skipConsentForVerifiableFirstPartyClients` and `enforcePolicies` (always sent) and the few settings the field reference marks — declare their live values before the first apply.

**Scopes are authoritative** — When `scopes` is set, each apply makes it the API's complete permission list: removing an entry deletes that scope from the API, and any role permission or client grant built on it stops granting access. Follow the `action:resource` naming pattern (`read:users`, `write:orders`) and prefer granular scopes over a coarse `admin` — scope descriptions are what users see on consent screens.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies. Downstream resources reference the API by its identifier (audience) string rather than a typed reference.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `identifier` | The API identifier (audience) | Auth0 Client `apiGrants[].audience`, Auth0 Role `permissions[].resourceServerIdentifier` |
| `id` | Auth0's internal resource server ID | Management API automation against this API |
| `signing_secret` | Token-signing secret, populated only for HS256 | Backend token verification for HS256 APIs — treat as a credential |
| `third_party_client_default_grant_ids` | The default grants' ids (`cgr_...`), keyed by subject type | Importing and auditing the grants |

## Common Patterns

**API with granular scopes** — Define one scope per action-resource pair (`read:orders`, `write:orders`, `delete:orders`) so each client is granted exactly the subset it needs. The trade is administrative surface: more scopes to grant, but least-privilege tokens and consent screens that say something meaningful.

**RBAC-enabled API** — Set `enforcePolicies: true` with `tokenDialect: access_token_authz`, then group this API's scopes into Auth0 Roles assigned to users. Tokens arrive carrying a permissions claim, so the backend authorizes from the token alone without a lookup per request.

**Machine-to-machine API** — The identifier serves as the audience that M2M clients request via the client-credentials flow, with each client's allowed scopes granted through Auth0 Client `apiGrants`. Keep `allowOfflineAccess` off — M2M clients re-authenticate with their credentials and have no use for refresh tokens.

**MCP server API** — An API an MCP server exposes: people reach it through third-party MCP clients holding the default `user` grant, and no application gets a token for itself. Start from the **MCP Server API** preset.

## Works With

- [**Auth0 Application (Client)**](/cloud-catalog/auth0-client) — the applications authorized to call this API; each `apiGrants` entry references this audience with the scopes granted to that client
- [**Auth0 Role**](/cloud-catalog/auth0-role) — groups this API's scopes into assignable access tiers; each role permission references the audience and a scope defined here
- [**Auth0 Tenant Settings**](/cloud-catalog/auth0-tenant-settings) — the tenant-level half of opening an API to third-party applications, such as dynamic client registration

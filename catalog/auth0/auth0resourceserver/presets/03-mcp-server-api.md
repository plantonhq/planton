# MCP Server API

This preset defines the API an MCP server exposes to AI agents and their MCP clients. MCP clients are third-party applications, usually registered by themselves -- through Dynamic Client Registration or from a Client ID Metadata Document -- so nobody can grant them access one by one. A default grant for third-party applications gives every one of them the server's two scopes on a person's behalf, the person consents, and the token carries only what they approved.

## When to Use

- A remote MCP server whose tools act on the signed-in person's data
- Any API opened to applications you do not register yourself (partners, agents, dynamically registered clients)

## Key Configuration Choices

- **Per-app authorization for people** (`subjectTypeAuthorization.user.policy: require_client_grant`) -- only applications holding a grant may act for a person; third-party applications get theirs from the default grant, and a first-party application of yours needs its own grant (Auth0 Application `apiGrants` with subject type `user`). Auth0 recommends this least-privilege policy; `allow_all` would let every first-party application in the tenant in without one. Third-party applications always need a grant, whichever policy is set.
- **No machine access** (`subjectTypeAuthorization.client.policy: deny_all`) -- no application gets a token for itself through the client-credentials flow; the server only ever acts for people.
- **The default grant** (`thirdPartyClientDefaultGrants`, subject type `user`) -- the ceiling every third-party application gets without a grant of its own. A grant made for one application takes precedence, to widen or narrow a single client.
- **IETF JWT access tokens** (`tokenDialect: rfc9068_profile`) and **refresh tokens** (`allowOfflineAccess: true`) -- the token shape resource servers validate by standard, and sessions that outlive a one-hour access token.
- **Requested scopes** -- with per-app authorization, a client names the scopes it wants in each token request.

## Before You Deploy

Tenant-level settings the MCP flow also needs live on Auth0 Tenant Settings and Auth0 Connection, not here: Dynamic Client Registration (when clients register themselves), the Resource Parameter Compatibility Profile Auth0 recommends for MCP clients, and a connection promoted to the domain level -- third-party applications sign people in only through domain-level connections.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.identifier` | The MCP server's URL (sample: `https://mcp.example.com/mcp`) | Your server's public endpoint; it cannot change after creation |
| `spec.scopes` | The permissions the server's tools need | Your server's tool design |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-api-with-scopes** -- a first-party API without third-party access
- **02-rbac-api** -- permissions from the person's roles

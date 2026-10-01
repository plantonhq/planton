# MCP Client From Its Document

This preset registers an MCP client in your tenant from the Client ID Metadata Document the client hosts. Auth0 fetches the document, takes the client's name, redirect URIs, logo and keys from it, and registers it as a strict third-party application -- no client secret is ever created or handed over. The spec adds only what the tenant decides: a description and the redirect policy, with a version number that lets you pull a changed document on purpose.

## When to Use

- An MCP client (or any agent that publishes a Client ID Metadata Document) needs to sign people in to your tenant and call APIs you protect
- A partner integration manages its own metadata and keys, and you want its registration in code rather than clicked in the dashboard
- The first registration of a client, before you tune its token lifetimes

## Key Configuration Choices

- **The document's URL** (`externalClientId`) -- the client's identity in every sign-in flow; the document's own `client_id` must equal it exactly. Changing it registers a different application with a new `client_id`
- **A version to fetch again** (`externalClientIdVersion: 1`) -- Auth0 reads the document once at registration; raise the number after the client's owner changes the document, and the next apply fetches it again
- **A description of your own** (`description`) -- seeded from the document; declaring it keeps every plan clean
- **Redirect protection** (`redirectionPolicy: open_redirect_protection`) -- a failed sign-in shows Auth0's error page instead of redirecting to a callback someone else controls

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.externalClientId` | The URL the client serves its metadata document from (sample: `https://mcp-client.example.com/.well-known/oauth-client-metadata`) | The client's own documentation or its operator; it is also the `client_id` field inside the document |
| `spec.description` | Up to 140 characters on what the client is for | Your own words |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

Before applying, turn on the tenant's Client ID Metadata Document registration (Auth0 Tenant Settings), promote the connections the client may sign people in through to the domain level, and grant the client each API it needs -- a third-party application reaches no API without a client grant.

## Related Presets

- **02-rotating-refresh-tokens** -- the same registration with refresh tokens that rotate and expire on a schedule you set

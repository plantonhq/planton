# Rotating Refresh Tokens

This preset registers an MCP client from its document and keeps its sessions short-lived and theft-evident. The client holds the refresh_token grant, every exchange returns a new refresh token and retires the old one, and a token dies after 15 days unused or 30 days in all. It is the shape OAuth 2.1 and the MCP authorization specification expect of a public client that keeps a person signed in between tasks.

## When to Use

- An MCP client that runs long tasks on a person's behalf and must renew its access without asking them to sign in again
- A public client (a desktop or command-line agent) where a refresh token could be copied off the machine
- Any registration where you want the token lifetimes reviewed in code rather than left at Auth0's defaults

## Key Configuration Choices

- **The refresh grant** (`grantTypes: [authorization_code, refresh_token]`) -- without `refresh_token` the settings below never take effect; declared here, it also overrides a document that forgot it
- **Rotation** (`refreshToken.rotationType: rotating`) -- a replayed refresh token revokes the whole family, so a stolen token is caught the first time either party uses it
- **Expiry** (`expirationType: expiring`, `tokenLifetime: 2592000`, `idleTokenLifetime: 1296000`) -- Auth0 requires these applications' refresh tokens to expire; the idle lifetime can never exceed the absolute one
- **Retry overlap** (`leeway: 3`) -- a client whose response was lost may present the just-rotated token once more within three seconds without tripping reuse detection; `0` turns the overlap off

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.externalClientId` | The URL the client serves its metadata document from (sample: `https://mcp-client.example.com/.well-known/oauth-client-metadata`) | The client's own documentation or its operator |
| `spec.refreshToken.tokenLifetime` / `idleTokenLifetime` | How long a session may last in all, and unused (seconds) | Your security policy; up to 157788000 seconds absolute |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-mcp-client-from-its-document** -- the registration alone, with Auth0's own token settings

# MCP-Ready Tenant

This preset opens the tenant to MCP clients: AI tools and agents that reach your API on a person's behalf and have never been registered in the tenant. It turns on the three settings such a client relies on to find its way in on its own, and nothing else.

## When to Use

- You publish an MCP server backed by an API the tenant protects, and want MCP clients (Claude Code, VS Code, custom agents) to sign people in without you registering each client by hand

## Key Configuration Choices

- **Client ID Metadata Document registration** (`clientIdMetadataDocumentSupported: true`) -- a client hosts a JSON document on its own HTTPS domain, and that URL is its client ID. What it exposes: anyone controlling a domain can present a client; each registers as a strict third-party application (PKCE required, only domain-level connections), and its logins fail while the tenant has active Rules. Auth0 marks the setting Early Access.
- **The resource parameter** (`resourceParameterProfile: compatibility`) -- when a request carries no `audience`, Auth0 reads the RFC 8707 `resource` parameter MCP clients send to name your API. What it exposes: nothing new to strangers; `resource` stops being forwarded to an upstream identity provider, so an integration that relied on that forwarding breaks.
- **Dynamic Client Registration** (`flags.enableDynamicClientRegistration: true`) -- opens `/oidc/register`, where any client registers a third-party application without a token. What it exposes: registration is open to everyone. Each registration reaches only the APIs and scopes the tenant's default third-party permissions grant and signs in only through domain-level connections, so configure those before deploying this preset.
- **Unset is unmanaged** -- every other tenant setting keeps its current value

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-product-branded-login** -- the name and logo MCP users see when they sign in

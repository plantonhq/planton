# RBAC API

This preset defines an API whose tokens carry the permissions a person holds through their roles. Auth0 evaluates role and permission assignments during login (`enforcePolicies: true`) and writes the granted permissions into a `permissions` claim (`tokenDialect: access_token_authz`), so the backend authorizes each request from the token alone.

## When to Use

- An API whose access depends on who the person is, not only which application calls it
- Admin and read-only tiers, grouped into Auth0 Roles that reference this API's scopes

## Key Configuration Choices

- **RBAC on** (`enforcePolicies: true`) -- the scopes a token carries are limited to the person's role permissions
- **Permissions in the token** (`tokenDialect: access_token_authz`) -- without an `_authz` dialect, RBAC filters scopes but the token carries no permissions claim; use `rfc9068_profile_authz` for the IETF JWT profile instead
- **One-hour tokens** (`tokenLifetime: 3600`) -- a role change takes effect at the next token

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.identifier` | The API's audience (sample: `https://api.example.com/v2`) | A stable URI you choose; it cannot change after creation |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-api-with-scopes** -- the same API without RBAC
- **03-mcp-server-api** -- an API third-party applications reach through default grants

# API with Scopes

This preset defines a backend API with RS256-signed tokens and one scope per action on its resource. It is the starting point for almost every API: applications get tokens for its audience, and each application's grant (Auth0 Application `apiGrants`) or each person's roles (Auth0 Role) decide which of the scopes a token carries.

## When to Use

- A first-party backend your own web, mobile or machine-to-machine applications call
- Any API whose permissions you want to grant one action at a time

## Key Configuration Choices

- **RS256** (`signingAlg: RS256`) -- Auth0 holds the private key; the API verifies tokens against the tenant's public keys, and nothing secret is shared
- **Granular scopes** (`read:orders`, `write:orders`, `delete:orders`) -- least-privilege tokens and consent screens that say something meaningful
- **Access policy left to Auth0** -- `subjectTypeAuthorization` is unset, so a new API keeps Auth0's defaults: every first-party application may get a token on a person's behalf, and a machine-to-machine application needs a grant

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.identifier` | The API's audience (sample: `https://api.example.com/`) | A stable URI you choose; it cannot change after creation |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-rbac-api** -- the same API with role-based permissions in its tokens
- **03-mcp-server-api** -- an API third-party applications reach through default grants

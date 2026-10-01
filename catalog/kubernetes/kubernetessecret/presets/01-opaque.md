# Opaque Secret

This preset creates an opaque Kubernetes Secret with arbitrary key-value data. The most common secret type, used for storing credentials, API keys, connection strings, and other sensitive data.

## When to Use

- Storing application credentials (database passwords, API keys, tokens)
- Any sensitive configuration data that should not be in ConfigMaps
- Generic key-value secret data that does not fit a specialized type (TLS, Docker registry, etc.)

## Key Configuration Choices

- **Opaque type** -- the default and most versatile secret type; stores arbitrary key-value pairs
- **Three example keys** -- `username`, `password`, `api-key`; replace with your application's secret keys

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the secret | Your namespace management |
| `$secret/replace-with-your-username-secret` | The organization secret holding the application username | `planton secret set <slug> --string` |
| `$secret/replace-with-your-password-secret` | The organization secret holding the application password | `planton secret set <slug> --string` |
| `$secret/replace-with-your-api-key-secret` | The organization secret holding the API key or token | `planton secret set <slug> --string` |

Every value in `data` is a reference to a managed secret on Planton, never the value itself; a deploy without the platform takes the literal. For several keys that belong together, one key-value secret serves them all: `$secret/<slug>/<key>`.

## Related Presets

- **02-tls** -- TLS certificate and key pair
- **03-docker-registry** -- Docker registry authentication credentials

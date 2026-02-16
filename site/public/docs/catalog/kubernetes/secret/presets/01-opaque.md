---
title: "Opaque Secret"
description: "This preset creates an opaque Kubernetes Secret with arbitrary key-value data. The most common secret type, used for storing credentials, API keys, connection strings, and other sensitive data."
type: "preset"
rank: "01"
presetSlug: "01-opaque"
componentSlug: "secret"
componentTitle: "Secret"
provider: "kubernetes"
icon: "package"
order: 1
---

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
| `<your-username>` | Application username | Your credential management system |
| `<your-password>` | Application password | Your credential management system |
| `<your-api-key>` | API key or token | Your service provider dashboard |

## Related Presets

- **02-tls** -- TLS certificate and key pair
- **03-docker-registry** -- Docker registry authentication credentials

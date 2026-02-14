# Restricted Read-Only Application Credential

This preset creates an application credential with read-only access to compute, network, and block-storage APIs. It uses the `reader` role and further restricts access via `accessRules` to GET requests only. Ideal for monitoring agents, dashboards, or audit tools that should never modify infrastructure.

## When to Use

- Monitoring agents (Prometheus, Datadog) that need to list instances, networks, and volumes
- Audit and compliance tools that read infrastructure state
- CI/CD pipelines that only need to verify deployment status (not modify resources)

## Key Configuration Choices

- **Reader role** -- minimum privilege role for read operations
- **GET-only access rules** -- even if the reader role permits some writes, access rules enforce HTTP GET only
- **Restricted** (`unrestricted: false`, default) -- cannot create sub-credentials or trusts
- **No expiration** -- add `expiresAt` field if time-limited access is required
- **Auto-generated secret** -- OpenStack generates a random secret at creation time

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`. Adjust `accessRules` paths and services to match your monitoring requirements.

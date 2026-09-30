---
title: "Variables"
description: "Centralized runtime variables for your services and infrastructure — organization-scoped and environment-scoped, with dynamic infrastructure references"
icon: settings
order: 47
tags:
  - Variables
  - Configuration
---

# Variables

Variables in Planton store non-sensitive configuration — database hostnames, API endpoints, feature flags, region names, port numbers, or any value that services and infrastructure deployments need at runtime but that does not require encryption. For sensitive values, use [Secrets](/docs/secrets).

If you have used GitHub Repository Variables, GitLab CI/CD Variables (unprotected), or Kubernetes ConfigMaps, the model is familiar. Planton adds organization-wide and environment-specific scoping, dynamic references to infrastructure outputs, and a unified management experience across the CLI and web console.

In the web console, variables are managed from the **Variables** tab within **Service Hub** in the sidebar.

## Scoping: Organization vs Environment

Every variable has a scope that determines where it is available:

**Organization variables** are available across all environments. Use these for values that are the same everywhere — a company domain name, a shared API base URL, a feature flag that applies globally.

**Environment variables** are specific to a single environment. Use these for values that differ between environments — the database hostname for staging vs production, a logging level, or an environment-specific service URL.

When a service deployment references a variable, Planton resolves it within the appropriate scope. Environment-scoped variables take precedence when the deployment targets that environment.

## Creating Variables

### Using the Web Console

Navigate to **Variables** in the sidebar. Click **Create Variable** and provide:

- **Name** — A descriptive identifier (e.g., `database-host`, `api-base-url`).
- **Scope** — Organization or Environment.
- **Description** — What this variable is used for.
- **Value** — Either a literal string value or a dynamic reference to an infrastructure output.

<!-- SCREENSHOT: Create variable form
  Page: /orgs/{org}/variables (create dialog)
  Action: Show the variable creation form with scope selection
  Focus: The complete form with literal value and ValueFromRef options visible
  Alt: Variable creation form showing name, scope, description, and value fields with option to reference infrastructure outputs
-->

### Using the CLI

```bash
# List every variable in the organization, with its scope
planton variable list

# What one environment can read: its own variables and the organization's
planton variable list --env production

# Create or update a variable (the value is positional)
planton variable set database-host db.internal.example.com --env production

# Read one value in a script (exit 0 found, 3 not found)
planton variable get database-host --env production -o plain

# Delete a variable
planton variable delete database-host --env production
```

`planton variable list` reads every page, so the list is never cut off. Without `--env` it shows the whole organization; the saved context's environment never narrows it.

## Dynamic References

Variables do not have to be static strings. A variable can reference an output field from a deployed infrastructure resource. When the variable is resolved, Planton looks up the referenced resource and extracts the current value.

For example, you can create a variable that references the endpoint of a deployed PostgreSQL cluster:

```yaml
name: DATABASE_HOST
description: "Primary database hostname, resolved from the deployed Postgres instance"
value_from:
  kind: PostgresCluster
  env: production
  name: my-postgres
  field_path: "status.endpoint"
```

When a service references this variable, it receives the actual endpoint of the deployed database — not a hardcoded string. If the database is redeployed and the endpoint changes, the variable resolves to the new value automatically.

Dynamic references are useful for:

- **Database endpoints** that change when clusters are rebuilt or migrated
- **Load balancer addresses** assigned by the cloud provider
- **Service URLs** that vary between environments
- **Resource ARNs or IDs** needed by dependent services

## Referencing Variables from Services

Services reference variables by name. During the deployment pipeline, Planton resolves the references and injects the values into the service's runtime environment.

The reference syntax names the variable, or the group and the entry inside it:

```
$var/<slug>                  # an organization variable
$var/<group>/<entry>         # an entry in an organization variable group
$var/@<env>/<slug>           # an environment's own variable
```

### Viewing Resolved Configuration

The CLI provides commands to see the fully resolved configuration for a service, combining both variables and secrets:

```bash
# Display resolved environment variables as a table
planton service env-vars --env production

# Generate .env and .env_export files for local development
planton service dot-env --env production

# Generate with value overrides for local testing
planton service dot-env --env production --set DB_HOST=localhost --set DB_PORT=5433
```

The `env-vars` command displays two tables — one for variables and one for secrets — with a status indicator showing whether each reference resolved successfully or encountered an error. This is useful for debugging configuration issues before deploying.

The `dot-env` command writes `.env` and `.env_export` files to the current directory. The `--set` flag lets you override specific values for local development without modifying the actual variable.

## Where Variables Are Read

Variables are the source of truth for non-sensitive configuration across the platform:

**Connection fields** — Non-sensitive fields in [connections](/docs/connections) (such as AWS account IDs, regions, or an Auth0 tenant's domain) take a variable in place of a literal: `region: {variable: default-aws-region}`, or one entry of a variable group as `region: {variable: aws-defaults/region}`. Several connections can share one value, and a fact that a connection and your services both read is declared once. Connections read organization-scoped variables.

**Cloud Resource inputs** — Any non-sensitive input field on a Cloud Resource takes a `$var/...` reference in place of a literal. Instead of duplicating a VPC ID or subnet name across multiple resource definitions, you store it as a variable and reference it. The value is resolved just-in-time before the deployment executes, and the field's own rules (a format, a length) are checked on the resolved value.

## Related Documentation

- [Variable Groups](/docs/variables/variable-groups) — Grouped variable collections for managing related configuration as a unit
- [Secrets](/docs/secrets) — Sensitive value management with encryption
- [What is a Service?](/docs/ci-cd/what-is-a-service) — How services consume variables and secrets
- [Deployment Targets](/docs/ci-cd/deployment-targets) — How configuration is applied to deployments

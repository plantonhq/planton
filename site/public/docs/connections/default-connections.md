---
title: "Default Connections"
description: "Configure default credentials so deployments resolve the right connection automatically without explicit specification"
icon: settings
order: 80
tags:
  - Connect
  - Default Connections
  - Resolution
---

# Default Connections

When you create an Infra Component or deploy a service, you can explicitly specify which credential to use. But in practice, most organizations have a primary credential per provider — one main AWS account, one main GCP project — and specifying it on every deployment is tedious and error-prone.

Default connections solve this. You designate one credential as the default for a provider, and Planton uses it automatically when no credential is explicitly specified. This reduces configuration overhead while keeping the explicit option available for cases that need it.

## How Resolution Works

When a deployment needs credentials for a provider and none are explicitly specified, Planton resolves the default using a simple two-level hierarchy:

1. **Environment-level default** — If a default is set for this specific provider and environment, use it.
2. **Organization-level default** — If no environment-level default exists, fall back to the organization-level default.
3. **No default found** — If neither exists, the deployment fails with a clear error identifying the missing default.

The resolution is deterministic and transparent. There is no guessing, no heuristic selection, and no implicit fallback to "any available credential." If you haven't set up defaults, deployments that don't specify credentials will fail — and that's intentional.

### Why Two Levels?

Most organizations have one primary credential per provider that works for the majority of environments. The organization-level default covers that case. But some environments need different credentials — production might use a separate AWS account from staging, or a compliance environment might require a dedicated GCP project.

Environment-level defaults let you override the organization default where needed, without affecting other environments.

### Example

```
AWS defaults:
  Organization default: aws-main-account
  Production override: aws-prod-account (environment-level)
  
Resolution:
  Development → aws-main-account (org default)
  Staging     → aws-main-account (org default)
  Production  → aws-prod-account (env override wins)
```

## Prerequisites

A connection must be [authorized](/docs/connections/environment-mappings) for an environment before it can be set as the default for that environment. This ensures that defaults can only reference credentials that have been explicitly approved for use. If you try to set a default that isn't authorized, the operation fails.

## Setting Defaults via the Web Console

### From the Provider Matrix

The Provider Matrix view shows both authorization and default status for every connection-environment combination. To set a default:

1. Navigate to **Connections** and open the Provider Matrix (or open a specific connection's detail page).
2. Find the cell for the connection and environment where you want to set the default.
3. Click the cell and select **Set as Default** from the action menu.

<!-- SCREENSHOT: Provider Matrix with default action
  Page: /orgs/{org}/connections or /{org}/connection/{kind}/{name}
  Action: Show the Provider Matrix with a cell action popover showing "Set as Default" option
  Focus: The action popover on a specific cell
  Alt: Provider Matrix cell showing action menu with Set as Default, Remove Authorization, and other options
-->

### From a Connection Detail Page

1. Navigate to the connection's detail page.
2. In the Environments section, click the status cell for the environment where you want to set the default.
3. Select **Set as Default**.

## Setting Defaults via the CLI

### Set an organization-level default

```bash
# Set a default for all environments
planton connection default set --provider aws --connection my-aws-account
```

### Set an environment-level default

```bash
# Override the default for a specific environment
planton connection default set --provider aws --connection aws-prod --env production
```

A default only selects a connection; it doesn't authorize one. When the connection isn't [authorized](/docs/connections/environment-mappings) everywhere the new default reaches (that environment, or every environment for an organization default), `default set` saves the default and warns, printing the `planton connection auth create` (or `delete` then `create`) command that closes the gap. A job that falls back to an unauthorized default is refused before anything runs.

### View current defaults

```bash
# List all defaults
planton connection default list

# Filter by provider
planton connection default list --provider aws

# Get the default for a specific provider
planton connection default get --provider aws

# Get the default for a specific provider and environment
planton connection default get --provider aws --env production
```

### Test resolution

The `resolve` command simulates the resolution algorithm — it tells you which credential would be used for a given provider and environment without actually running a deployment. This is useful for verifying your default configuration before deploying.

```bash
# Which AWS credential would be used in production?
planton connection default resolve --provider aws --env production

# Which GCP credential would be used in development?
planton connection default resolve --provider gcp --env development
```

### Remove a default

```bash
# Remove the organization-level default for a provider
planton connection default unset --provider aws

# Remove an environment-level default
planton connection default unset --provider aws --env production
```

Removing an environment-level default causes that environment to fall back to the organization-level default. Removing the organization-level default means environments without their own defaults will have no default at all — deployments that don't specify credentials will fail.

## Common Patterns

### Simple Setup (One Account Per Provider)

For organizations with a single cloud account per provider:

```bash
# Set org-level defaults for your primary accounts
planton connection default set --provider aws --connection aws-main
planton connection default set --provider gcp --connection gcp-main
planton connection default set --provider cloudflare --connection cf-main
```

Every deployment in every environment uses these credentials automatically. No per-deployment configuration needed.

### Production Isolation

For organizations that separate production credentials:

```bash
# Org-level defaults for non-production environments
planton connection default set --provider aws --connection aws-dev

# Production overrides
planton connection default set --provider aws --connection aws-prod --env production
planton connection default set --provider gcp --connection gcp-prod --env production
```

Development and staging use the dev credentials. Production uses its own credentials. The `resolve` command lets you verify:

```bash
planton connection default resolve --provider aws --env staging
# → aws-dev

planton connection default resolve --provider aws --env production
# → aws-prod
```

### Multi-Region or Multi-Account

For organizations with region- or account-specific credentials:

```bash
# Default for most environments
planton connection default set --provider aws --connection aws-us-west

# Regional overrides
planton connection default set --provider aws --connection aws-eu-west --env eu-production
planton connection default set --provider aws --connection aws-ap-southeast --env ap-production
```

## Troubleshooting

### "No default connection for provider"

A deployment failed because no default credential was found.

**Fix**:
1. Check if a default exists: `planton connection default get --provider <provider>`
2. Check if it's authorized for the environment: `planton connection auth list --provider <provider>`
3. Set a default if none exists: `planton connection default set --provider <provider> --connection <slug>`

### "Nothing ran: environment … may not use the … connection …"

A default is set, but the underlying credential isn't authorized for the target environment. The job is refused before any resource runs, and the refusal names both and the fix:

```
Nothing ran: environment staging may not use the aws connection aws-prod. Authorize it
(planton connection auth create --provider aws --connection aws-prod --scope environment
--environments staging), or make it the organization's (--scope organization), then run the job again.
```

Following an infra pipeline, the CLI adds: `Refused before any resource ran: no infra job started, so nothing in the cloud changed.`

**Fix**:
1. Authorize the connection: `planton connection auth create --provider <provider> --connection <slug> --scope environment --environments <env>` (or `--scope organization`). If the connection already has an environment-scoped authorization, `planton connection auth delete --provider <provider> --connection <slug>` first, then create it again with every environment it should reach.
2. Or change the default to a connection that is already authorized.
3. Run the job again.

### Resolution returns unexpected credential

Use `resolve` to debug:

```bash
planton connection default resolve --provider aws --env production
```

If the result isn't what you expect, check whether an environment-level default is overriding the organization default, or whether the organization default was recently changed.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [Environment Mappings](/docs/connections/environment-mappings) — Authorize connections for environments (prerequisite for defaults)
- [Cloud Providers](/docs/connections/cloud-providers) — Connect cloud provider accounts
- [Infrastructure](/docs/infrastructure) — How infrastructure deployments use default connections

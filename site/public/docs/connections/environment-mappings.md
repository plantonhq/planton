---
title: "Environment Mappings"
description: "Control which credentials can be used in which environments through connection authorization"
icon: settings
order: 70
tags:
  - Connect
  - Authorization
  - Environments
  - Security
---

# Environment Mappings

Creating a connection gives Planton the credentials to authenticate with a provider. But having credentials is not the same as being allowed to use them everywhere. Environment mappings — technically, connection authorizations — control which credentials can be used in which environments.

This is a security boundary. Without it, a developer working in a development environment could accidentally deploy with production AWS credentials, or a staging database connection could be used in a production deployment. Authorization ensures that each environment uses only the credentials explicitly approved for it.

## How Authorization Works

Connections exist at the organization level. Environments are the deployment contexts within your organization (development, staging, production, and any custom environments you create). Authorization is the bridge between them.

When you authorize a connection for an environment, you're saying: "This credential is allowed to be used for deployments targeting this environment." When a deployment runs, Planton checks that the credential being used (whether selected explicitly or resolved as a default) is authorized for the target environment. If it isn't, the deployment fails.

### Two Authorization Scopes

You choose one of two scopes when authorizing a connection:

**Organization-wide** — The connection is authorized for all environments in the organization, including any environments created in the future. This is convenient for credentials that should be universally available, like a shared development AWS account or a GitHub connection.

**Environment-specific** — The connection is authorized only for the environments you explicitly list. New environments are not automatically included. This is the right choice for production credentials, compliance-sensitive connections, or any credential where you want explicit control over where it can be used.

You can change the scope at any time. Narrowing from organization-wide to environment-specific does not affect in-flight deployments, but future deployments in unauthorized environments will fail.

## The Provider Matrix

The web console provides a Provider Matrix view that shows the authorization status of all connections across all environments at a glance.

<!-- SCREENSHOT: Provider Matrix view
  Page: /orgs/{org}/connections (Provider Matrix tab or connection detail page)
  Action: Show the matrix with multiple connections as rows and environments as columns, with various status indicators
  Focus: The full matrix grid showing authorized, default, and not-authorized states
  Alt: Provider Matrix showing connection authorization status across development, staging, and production environments
-->

Each cell in the matrix represents one connection in one environment. The cell shows one of four states:

- **Not authorized** — The connection cannot be used in this environment. Click to authorize.
- **Authorized** — The connection can be used in this environment, but it's not the default.
- **Default** — The connection is authorized and set as the default for this provider in this environment.
- **Misconfigured** — The connection has an inconsistent state (e.g., set as default but not authorized). Follow the prompts to resolve.

## Managing Authorization via the Web Console

### Authorizing a Connection

1. Navigate to the connection's detail page (click the connection in the "Connected Providers" list, or click a cell in the Provider Matrix).
2. The **Environments** section shows the current authorization state.
3. Click **Configure Authorization** (if no authorization exists) or **Manage** (to modify an existing authorization).
4. Choose your scope:
   - **Organization-wide** — Authorize for all environments.
   - **Environment-specific** — Select individual environments from the list.
5. Save the authorization.

<!-- SCREENSHOT: Manage Authorization dialog
  Page: /{org}/connection/{kind}/{name}
  Action: Show the Manage Scope dialog with organization-wide and environment-specific options
  Focus: The scope selection and environment checkboxes
  Alt: Authorization scope dialog showing organization-wide option and environment-specific option with environment list
-->

### Removing Authorization

1. Navigate to the connection's detail page.
2. In the Environments section, click the cell for the environment you want to deauthorize (or use the Manage dialog to change scope).
3. Confirm the removal.

Removing authorization immediately prevents the connection from being used in that environment. Any deployments currently in progress continue with the credentials they started with, but new deployments will fail if no alternative authorized credential exists.

## Managing Authorization via the CLI

### Authorize a connection

```bash
# Authorize for all environments (organization-wide)
planton connection authorization create \
  --provider aws \
  --connection my-aws-prod \
  --scope organization

# Authorize for specific environments
planton connection authorization create \
  --provider gcp \
  --connection gcp-analytics \
  --scope environment \
  --environments staging,production
```

A connection has one authorization, so `create` is refused when one already exists. To add an environment to an environment-scoped authorization, or to change its scope, delete it and create it again with the scope and full environment list you want. `--provider` takes any provider Planton connects to, including `vault` and `openfga`.

A deployment into an environment the connection isn't authorized for is refused before anything runs, and the refusal carries the command: *"Nothing ran: environment staging may not use the aws connection aws-prod. Authorize it (planton connection auth create --provider aws --connection aws-prod --scope environment --environments staging), or make it the organization's (--scope organization), then run the job again."*

### List authorizations

```bash
# List all authorizations
planton connection authorization list

# Filter by provider
planton connection auth list --provider aws

# Output as JSON
planton connection auth list -o json
```

### View a specific authorization

```bash
# Get authorization details by provider and connection
planton connection auth get --provider aws --connection my-aws-prod
```

### Remove an authorization

```bash
# Remove authorization for a connection
planton connection auth delete --provider aws --connection my-aws-prod
```

## Common Patterns

### Separate Credentials Per Environment Tier

The most common pattern: different cloud accounts for different environment tiers.

```
aws-dev-account     → authorized for: development
aws-staging-account → authorized for: staging
aws-prod-account    → authorized for: production
```

Each environment uses credentials that match its security posture. Development uses a sandbox account with relaxed permissions. Production uses a hardened account with strict IAM policies.

### Shared Connections Across Tiers

Some connections are safe to share across all environments — typically git provider connections and container registries:

```
github-acme-org → authorized for: all environments (organization-wide)
ecr-shared      → authorized for: all environments (organization-wide)
```

These don't carry the same blast radius as cloud provider credentials, so organization-wide authorization is practical.

### Gradual Rollout

When introducing a new credential, start narrow and expand:

1. Authorize for development only.
2. Verify deployments work correctly.
3. Expand to staging.
4. Verify again.
5. Expand to production (or set as default).

This reduces risk when rotating credentials or switching providers.

## What Happens Without Authorization

If a deployment targets an environment where no authorized connection exists for the required provider, the deployment fails with a clear error message identifying the missing authorization. The fix is to either:

1. Authorize an existing connection for that environment, or
2. Create a new connection and authorize it.

Planton does not fall back to any credential — it fails explicitly rather than risking the use of an unintended credential.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [Default Connections](/docs/connections/default-connections) — How defaults work on top of authorization
- [Cloud Providers](/docs/connections/cloud-providers) — Connect cloud provider accounts

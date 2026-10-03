# Auth0 Role

Deploys an Auth0 Role and its complete set of API permissions (scopes) as a single Infra Component. Roles are the middle layer of Auth0 role-based access control (RBAC) -- they group the scopes defined on a Resource Server into a reusable access tier that you assign to users.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auth0 Role** -- a role configured with the specified display name and description
- **Role Permissions** -- created only when `permissions` is non-empty, an authoritative permission assignment that sets the role's complete scope list

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **An Auth0 tenant** with RBAC available.
- **Existing scopes** for any permission you grant. Each permission names a scope and the identifier (audience) of the Resource Server that owns it; the scope must already exist on that Resource Server (created via the Auth0 Resource Server component or directly in Auth0). A role can be deployed with no permissions and have them added later.

## Deploy

### Console

Open the deployment store, find **Auth0 Role**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the role's display name and description, and an inline builder for the permission set. Start from the **Role with Permissions** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Role
metadata:
  name: editor
  org: acme-corp
  env: prod
spec:
  name: Editor
  description: Read and write access to the orders API
  permissions:
    - name: read:orders
      resourceServerIdentifier: https://api.example.com/
    - name: write:orders
      resourceServerIdentifier: https://api.example.com/
```

```shell
planton apply -f auth0-role.yaml
```

This creates an Editor role and grants it two scopes on a single API. All spec fields are optional -- omit `permissions` to create an empty role, or omit `name` to default the role name to `metadata.name`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an Auth0 Role. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Display name** -- The `name` field sets the human-readable role name shown in the Auth0 dashboard and when assigning the role to users. When omitted, it defaults to `metadata.name`. Set it when you want a friendlier label (e.g., resource name `viewer`, display name `Viewer`).

**Permission set** -- The `permissions` array is the heart of the role. Each entry pairs a scope `name` (e.g., `read:orders`) with the `resourceServerIdentifier` (audience) of the API that defines it. A single role can span multiple APIs by listing permissions with different identifiers.

**Authoritative reconciliation** -- The permission set is authoritative: each deploy sets the role's permissions to exactly the list provided. Removing a permission from the spec revokes it from everyone who holds the role on the next apply -- there is no partial or additive mode.

**Audience accuracy** -- The `resourceServerIdentifier` must exactly match the target Resource Server's identifier. A mismatched audience silently grants nothing, so reuse the same value across permissions on the same API.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. Permissions reference a Resource Server by its identifier (audience) string rather than a typed reference.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | Auth0 role identifier (e.g., `rol_abc123`) | Assigning the role to users, Management API calls |
| `name` | Human-readable role name | Dashboards, audit logs |
| `description` | Role description | Access reviews, audit logs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Role with permissions** -- A named access tier backed by a focused set of scopes on one API. Start from the **Role with Permissions** preset.

**Administrator across APIs** -- A role that grants scopes spanning multiple Resource Servers. Start from the **Administrator Role (Multiple APIs)** preset.

**Empty role** -- A role created without permissions, to be populated later or assigned as a placeholder tier. Start from the **Role Without Permissions** preset.

## Works With

- [**Auth0 Resource Server (API)**](/infra-catalog/auth0-resource-server) -- defines the scopes a role grants; each permission references a Resource Server by its identifier (audience).

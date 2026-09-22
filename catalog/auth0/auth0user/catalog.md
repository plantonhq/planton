# Auth0 User

Deploys an Auth0 User -- an identity in one of your tenant's database or passwordless connections -- with its profile, its complete set of roles and direct API permissions, and, on a database connection, an initial password the module mints when you declare none. The user's identity-provider subject is an output, so any system that grants standing by subject reads it by reference instead of copying it from the dashboard.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Auth0 User** -- the user in the connection you reference, with the profile fields, verification flags, and metadata documents you declare
- **Initial Password** -- generated only when no password is declared on a database connection (24 characters, letters and digits), reported once in the outputs
- **User Roles** -- created only when `roles` is non-empty, an authoritative assignment that sets the user's complete role list
- **User Permissions** -- created only when `permissions` is non-empty, an authoritative assignment that sets the user's complete list of direct API permissions

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with Auth0 domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource. The Machine-to-Machine application behind it needs the `create:users`, `read:users`, `update:users`, and `delete:users` scopes, plus `read:roles` when roles are assigned.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **A database or passwordless connection** the user is created in -- an `Auth0Connection` with strategy `auth0`, `email`, or `sms`. Auth0 can only create users in connections it holds the credential for; users of social and enterprise connections arrive at first sign-in and cannot be declared.
- **Existing roles and scopes** for anything you assign. Roles are referenced by id (an `Auth0Role`); each direct permission names a scope on a Resource Server identifier (an `Auth0ResourceServer`). Both can be added later -- a user deploys fine with neither.

## Deploy

### Console

Open the deployment store, find **Auth0 User**, and click **Deploy**. The creation wizard walks you through the connection, who the person is, how they appear, what they may do, and how they sign in -- with "let Planton mint the password" as the recommended posture. Start from the **Staff Account with Minted Password** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0User
metadata:
  name: staff-root
  org: acme-corp
  env: prod
spec:
  connectionName:
    valueFrom:
      kind: Auth0Connection
      name: users
      fieldPath: status.outputs.name
  email: platform-root@acme.example
  name: Platform root
  emailVerified: true
  verifyEmail: false
  roles:
    - valueFrom:
        kind: Auth0Role
        name: administrator
        fieldPath: status.outputs.id
```

```shell
planton apply -f auth0-user.yaml
```

This creates a verified user in the `users` connection, mints its initial password, and assigns it the administrator role. A Stack Job tracks the provisioning in real time; read `status.outputs.password` once into the credential store that owns it.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the user to the connection, roles, and resource servers deployed in the same InfraPipeline:

```yaml
spec:
  connectionName:
    valueFrom:
      kind: Auth0Connection
      name: users
      fieldPath: status.outputs.name
  roles:
    - valueFrom:
        kind: Auth0Role
        name: administrator
        fieldPath: status.outputs.id
  permissions:
    - name: read:audit-log
      resourceServerIdentifier:
        valueFrom:
          kind: Auth0ResourceServer
          name: backend-api
          fieldPath: status.outputs.identifier
```

The InfraPipeline resolves the dependency graph, deploys the connection, role, and resource server first, then provisions the user with the resolved values. A control plane that grants standing by identity-provider subject reads `status.outputs.user_id` from this resource the same way.

## Key Configuration

These are the most important decisions when configuring an Auth0 User. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which connection** -- `connectionName` decides everything downstream: a database connection takes a password (declared or minted); a passwordless `email` or `sms` connection takes none and the user signs in with a one-time code. Reference the `Auth0Connection` so the connection is created first and its name is never retyped. The choice is immutable -- a user cannot move between connections.

**The password's three states** -- Leave `password` empty on a database connection and the module mints one, reporting it in `status.outputs.password`; this is the recommended posture, because a credential nobody typed is a credential nobody has to protect twice. Declare it as a managed-secret reference to bring your own (it is never echoed back). Set `passwordless: true` for a connection that has none.

**Vouching for the address** -- `emailVerified: true` marks the address as confirmed, and `verifyEmail: false` stops Auth0 from sending a confirmation; together they are the posture for an identity you own. For a person who should confirm their own address, leave both unset and Auth0 asks them.

**Standing as one file** -- `roles` and `permissions` are authoritative: each deploy sets the user's role list and direct-permission list to exactly what the manifest says, so an entry removed from the manifest is revoked on the next apply. Prefer roles; direct permissions are for the one-off scope a role would be too heavy for.

**Two kinds of metadata** -- `userMetadata` is the person's own (a locale, a preference); `appMetadata` is the application's (a plan, a team, an external id) and the user cannot change it. Neither carries roles or permissions -- those have first-class fields.

**Suspend before you delete** -- `blocked: true` refuses every sign-in while keeping the record and its subject; deleting the resource removes the subject for good, and anything that granted standing to it must be re-pointed.

## Outputs and Dependencies

### What This Component Consumes

| Reference | Kind | Field | Purpose |
|-----------|------|-------|---------|
| `connectionName` | Auth0Connection | `status.outputs.name` | The connection the user is created in |
| `roles[]` | Auth0Role | `status.outputs.id` | Roles assigned to the user |
| `permissions[].resourceServerIdentifier` | Auth0ResourceServer | `status.outputs.identifier` | The API that defines a directly granted scope |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_id` | The full identity-provider subject (e.g. `auth0\|66f1c2d3...`), the `sub` claim in every token | A control plane that grants standing by subject; audit configuration |
| `email` | The address as stored | Notifications, dashboards |
| `username` | The login name, when the connection requires one | Application configuration |
| `name`, `nickname`, `picture` | The profile as stored, including Auth0's defaults | Display |
| `connection_name` | The connection the user belongs to | Audit |
| `password` | The minted initial password, only when the module generated it | Read once into the credential store that owns it |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Staff account with minted password** -- An identity you own, verified, with no password anywhere in the files. Start from the **Staff Account with Minted Password** preset.

**Administrator with roles and permissions** -- A user whose complete standing is written down by reference. Start from the **Administrator with Roles and Permissions** preset.

**Passwordless support identity** -- A user in an email or SMS connection, one-time codes, no password to manage. Start from the **Passwordless Support Identity** preset.

## Works With

- [**Auth0 Connection (Identity Provider)**](/cloud-catalog/auth0-connection) -- the database or passwordless connection the user is created in; referenced by name.
- [**Auth0 Role**](/cloud-catalog/auth0-role) -- the roles assigned to the user; referenced by id.
- [**Auth0 Resource Server (API)**](/cloud-catalog/auth0-resource-server) -- defines the scopes a direct permission grants; referenced by identifier.

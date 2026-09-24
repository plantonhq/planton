# Auth0User -- Pulumi Module

Pulumi Go module that creates and manages an Auth0 User in a database or passwordless connection, with its authoritative roles and direct API permissions, minting the initial password when none is declared.

## What It Creates

- `random.RandomPassword` (conditional) -- The initial password, minted only when `password` is empty on a database connection (`passwordless` false): 24 characters, upper- and lower-case letters and digits, no symbols. Its generation-shape arguments are ignored after creation so an imported credential never regenerates on preview.
- `auth0.User` -- The user in the named connection, with the profile fields, verification flags, metadata documents, and the declared or minted password. No password is sent on a passwordless connection.
- `auth0.UserRoles` (conditional) -- Sets the user's complete role list when `roles` is non-empty. This is the authoritative set: a role omitted on a later apply is removed from the user.
- `auth0.UserPermissions` (conditional) -- Sets the user's complete list of direct API permissions when `permissions` is non-empty. Authoritative in the same way.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for a Machine-to-Machine application granted `create:users`, `read:users`, `update:users`, `delete:users`, and -- when roles are assigned -- `read:roles`.
- The connection named by `connection_name` must be a database (`auth0`) or passwordless (`email`, `sms`) connection that already exists (e.g., created via the `Auth0Connection` component). Users of social and enterprise connections cannot be created through the Management API.
- The roles referenced by `roles` and the scopes referenced by `permissions` must already exist (e.g., created via the `Auth0Role` and `Auth0ResourceServer` components).

## Local Testing

```bash
# Install Pulumi Auth0 plugin
make install-pulumi-plugins

# Login to local state
pulumi login --local

# Preview with test manifest
make test

# Or use debug script
./debug.sh ../../e2e/manifest.yaml
```

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Module Structure

| File | Purpose |
|---|---|
| `module/main.go` | Provider setup and the create sequence: user, roles, permissions, outputs |
| `module/locals.go` | The spec read once into plain values: references flattened, metadata rendered as JSON, the mint predicate |
| `module/user.go` | The user resource, the minted password, and the two authoritative companion resources |
| `module/outputs.go` | Stack outputs mapped onto `Auth0UserStackOutputs`; the password exported only when minted |

## Outputs

| Output | Description |
|---|---|
| `user_id` | The full identity-provider subject, connection prefix included (e.g. `auth0\|66f1c2d3...`) -- the `sub` claim in every token issued for the user |
| `email` | The email address as Auth0 stored it |
| `username` | The login name, when the connection requires one |
| `name` | The display name as stored, including Auth0's default when none was declared |
| `nickname` | The short name as stored, including Auth0's default when none was declared |
| `picture` | The avatar URL as stored, including Auth0's placeholder when none was declared |
| `connection_name` | The connection the user belongs to |
| `password` | The minted initial password (secret) -- exported only when the module generated it |

# Auth0User -- Terraform Module

Terraform/OpenTofu module that creates and manages an Auth0 User in a database or passwordless connection, with its authoritative roles and direct API permissions, minting the initial password when none is declared.

## What It Creates

- `random_password` (conditional) -- The initial password, minted only when `password` is empty on a database connection (`passwordless` false): 24 characters, upper- and lower-case letters and digits, no symbols. Its generation-shape arguments are ignored after creation so an imported credential never regenerates on plan.
- `auth0_user` -- The user in the named connection, with the profile fields, verification flags, metadata documents, and the declared or minted password. No password is sent on a passwordless connection.
- `auth0_user_roles` (conditional) -- Sets the user's complete role list when `roles` is non-empty. This is the authoritative set: a role omitted on a later apply is removed from the user.
- `auth0_user_permissions` (conditional) -- Sets the user's complete list of direct API permissions when `permissions` is non-empty. Authoritative in the same way.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables, for a Machine-to-Machine application granted `create:users`, `read:users`, `update:users`, `delete:users`, and -- when roles are assigned -- `read:roles`.
- The connection named by `connection_name` must be a database (`auth0`) or passwordless (`email`, `sms`) connection that already exists (e.g., created via the `Auth0Connection` component). Users of social and enterprise connections cannot be created through the Management API.
- The roles referenced by `roles` and the scopes referenced by `permissions` must already exist (e.g., created via the `Auth0Role` and `Auth0ResourceServer` components).

## Usage

```hcl
module "auth0_user" {
  source = "."

  metadata = {
    name = "staff-root"
  }

  spec = {
    connection_name = "users"
    email           = "platform-root@example.com"
    name            = "Platform root"
    email_verified  = true
    verify_email    = false
    roles           = ["rol_abc123"]
  }
}
```

With no `password`, the module mints one and reports it in the `password` output. Declare `password` to bring your own (it is never echoed back), or set `passwordless = true` for an email or SMS connection, where Auth0 refuses a password.

## Inputs

`variables.tf` is generated from the `Auth0UserSpec` proto (`planton tofu generate-variables Auth0User`); the field comments in `spec.proto` are the authority on every input.

| Variable | Type | Required | Description |
|---|---|---|---|
| `metadata` | object | Yes | Resource metadata (name, org, env) |
| `spec` | object | Yes | User specification: connection, profile, verification flags, password posture, metadata documents, roles, permissions |

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
| `password` | The minted initial password (sensitive) -- set only when the module generated it |

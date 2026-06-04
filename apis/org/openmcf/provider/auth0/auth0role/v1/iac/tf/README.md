# Auth0Role — Terraform Module

Terraform/OpenTofu module that creates and manages an Auth0 Role and its permissions.

## What It Creates

- `auth0_role` — The role resource with name and description.
- `auth0_role_permissions` (conditional) — Sets the role's complete permission list when `permissions` is non-empty. This is the authoritative set: a permission omitted on a later apply is removed from the role.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables.
- The permission scopes referenced by `permissions` must already exist on their resource servers (e.g., created via the `Auth0ResourceServer` component).

## Usage

```hcl
module "auth0_role" {
  source = "."

  metadata = {
    name = "viewer"
  }

  spec = {
    name        = "Viewer"
    description = "Read-only access to the orders API"
    permissions = [
      {
        name                       = "read:orders"
        resource_server_identifier = "https://api.example.com/"
      },
    ]
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|---|---|---|---|
| `metadata` | object | Yes | Resource metadata (name, org, env) |
| `spec` | object | Yes | Role specification (name, description, permissions) |

## Outputs

| Output | Description |
|---|---|
| `id` | Auth0 role identifier (e.g. rol_abc123) |
| `name` | Role name |
| `description` | Role description |

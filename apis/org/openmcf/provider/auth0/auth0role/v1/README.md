# Auth0Role

Manages an [Auth0 Role](https://auth0.com/docs/manage-users/access-control/rbac) — a named collection of API permissions (scopes) that implements Auth0's role-based access control (RBAC). A role groups scopes defined on one or more resource servers and can then be assigned to users.

## When to Use

- **Access tiers**: Define reusable roles (Administrator, Editor, Viewer) for an application.
- **RBAC**: Group API scopes into roles for assignment to users or groups.
- **Cross-API roles**: Aggregate scopes from multiple resource servers into one role.
- **Infrastructure as code**: Manage role-to-permission mappings as version-controlled config.

## Quick Start

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Role
metadata:
  name: viewer
  org: acme-corp
  env: production
spec:
  name: Viewer
  description: Read-only access to the orders API
  permissions:
    - name: read:orders
      resource_server_identifier: https://api.example.com/
```

## Key Behaviors

- **name**: Optional friendly role name. Defaults to `metadata.name` when omitted.
- **permissions**: The authoritative set of scopes granted to the role. The deployment manages the complete list — a permission removed from the spec is removed from the role on the next apply. When omitted, the role is created with no permissions.
- **permission scopes**: Each permission references a scope by `name` and the `resource_server_identifier` (audience) of the API that owns it. The scope must already exist on that resource server (see [Auth0ResourceServer](../../auth0resourceserver/v1/README.md)).

## Outputs

| Output | Description |
|---|---|
| `id` | Auth0 role identifier (e.g. `rol_abc123`) |
| `name` | Role name |
| `description` | Role description |

## Auth0 Documentation

- [Role-Based Access Control (RBAC)](https://auth0.com/docs/manage-users/access-control/rbac)
- [Create roles](https://auth0.com/docs/manage-users/access-control/configure-core-rbac/roles/create-roles)
- [Add permissions to roles](https://auth0.com/docs/manage-users/access-control/configure-core-rbac/roles/add-permissions-to-roles)
- [Terraform auth0_role](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/role)
- [Terraform auth0_role_permissions](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/role_permissions)

# Auth0Role — Research Documentation

## What Are Auth0 Roles?

Auth0 Roles are a core building block of Auth0's role-based access control (RBAC). A role is a named collection of **permissions**, where each permission is a scope defined on a **resource server** (an API registered in the tenant). Roles are assigned to users; when a user authenticates, their roles' permissions can be embedded in the issued access token (when the resource server enables RBAC and an `_authz` token dialect).

The RBAC model in Auth0 has three layers:

1. **Resource Server (API)** — defines the available scopes (e.g., `read:orders`). Modeled by the `Auth0ResourceServer` component.
2. **Role** — groups a subset of those scopes into a named tier (e.g., `Editor`). Modeled by this component.
3. **User assignment** — assigns roles to users. This is a runtime/identity operation, intentionally out of scope for infrastructure-as-code (see "80/20 Scoping" below).

## The Role and Permission Resources

The Auth0 provider models roles and their permissions as separate resources:

| Concern | Terraform | Pulumi |
|---|---|---|
| The role itself (name, description) | `auth0_role` | `auth0.Role` |
| The role's full permission set (authoritative) | `auth0_role_permissions` | `auth0.RolePermissions` |
| A single appended permission (non-authoritative) | `auth0_role_permission` | `auth0.RolePermission` |

This component folds the role and its permission set into a single deployable unit using the **authoritative** resource (`auth0_role_permissions` / `auth0.RolePermissions`). A role with no permissions is inert, so combining the two in one manifest is the natural 80/20 shape: deploy a role and the scopes it grants in one apply.

### Why the authoritative set, not the append resource

The Auth0 provider warns against mixing `auth0_role_permissions` (manages the complete list) with `auth0_role_permission` (appends one). Choosing the authoritative resource gives declarative, drift-correcting behavior consistent with the rest of Planton: the manifest is the source of truth, and a permission removed from the spec is removed from the role on the next apply. The append-style single-permission resource is intentionally not modeled.

## Permissions Reference Existing Scopes

A role permission is a reference to a scope that must **already exist** on the named resource server. It is identified by:

- `name` — the scope name (e.g., `read:orders`), and
- `resource_server_identifier` — the identifier (audience) of the resource server that owns the scope.

If the scope or resource server does not exist when the role is applied, the Auth0 Management API rejects the permission assignment. The typical workflow is to deploy an `Auth0ResourceServer` (which defines the scopes) before, or alongside, the roles that grant them.

## 80/20 Scoping Decision

**In scope:**

- Role name and description.
- The authoritative set of permissions (scope + resource server identifier).

**Out of scope (intentionally):**

- **User-to-role assignment.** Assigning roles to users (`auth0_user_roles` / `auth0_organization_member_roles`) mixes runtime identity (which users exist) into infrastructure config. User lifecycle is managed through Auth0's user APIs, SSO/JIT provisioning, or organization membership flows — not through static IaC. Keeping roles and assignments separate is the standard, clean separation.
- **The append-style single-permission resource** (`auth0_role_permission`), for the drift-correctness reasons above.
- **Description length / name format validation** beyond presence. Auth0 enforces its own limits server-side; the component does not duplicate undocumented constraints to avoid rejecting input the API would accept.

## Provider Parity

| Feature | Terraform | Pulumi | Planton Spec |
|---|---|---|---|
| Create role | `auth0_role` | `auth0.Role` | `Auth0RoleSpec` (name, description) |
| Set permissions (authoritative) | `auth0_role_permissions` | `auth0.RolePermissions` | `permissions` list |
| Append single permission | `auth0_role_permission` | `auth0.RolePermission` | Not supported (80/20) |
| Assign role to users | `auth0_user_roles` | `auth0.UserRoles` | Not supported (runtime identity) |

## Best Practices

- **Least privilege**: grant a role only the scopes its tier genuinely needs.
- **Deploy APIs first**: ensure the resource server and its scopes exist before assigning them to a role.
- **Name roles for humans**: use clear `spec.name` values (Administrator, Editor, Viewer); the role name is what appears in the Auth0 dashboard and in assignment UIs.
- **Keep roles small and composable**: prefer several focused roles over one broad role when different users need different subsets of access.

## Common Pitfalls

- **Referencing a non-existent scope**: the apply fails if `name`/`resource_server_identifier` does not match an existing scope on the resource server.
- **Mixing authoritative and append management**: managing the same role's permissions both here and via an external append-style process causes churn. This component owns the full set.
- **Expecting user assignments**: this component does not assign the role to anyone; assignment is a separate identity operation.

## References

- [Auth0 RBAC](https://auth0.com/docs/manage-users/access-control/rbac)
- [Create roles](https://auth0.com/docs/manage-users/access-control/configure-core-rbac/roles/create-roles)
- [Add permissions to roles](https://auth0.com/docs/manage-users/access-control/configure-core-rbac/roles/add-permissions-to-roles)
- [Terraform auth0_role](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/role)
- [Terraform auth0_role_permissions](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/role_permissions)
- [Pulumi auth0.Role](https://www.pulumi.com/registry/packages/auth0/api-docs/role/)
- [Pulumi auth0.RolePermissions](https://www.pulumi.com/registry/packages/auth0/api-docs/rolepermissions/)

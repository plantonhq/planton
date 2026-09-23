# Auth0Role

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0RoleSpec defines the configuration for an Auth0 Role.
In Auth0, a Role is a named collection of permissions (scopes) that can be
assigned to users. Roles are the building block of Auth0's role-based access
control (RBAC): permissions are defined on an Auth0 Resource Server (API),
grouped into roles here, and then assigned to users.

This spec covers the 80/20 use case for managing Auth0 Roles:
- Defining a role with a human-readable name and description
- Granting the role a set of API permissions (scopes), each scoped to the
  resource server (API) that owns it

Permissions are folded into this component: the IaC modules create the role
AND set its permissions in a single deployment, so a role is useful out of the
box. The permission set is authoritative -- the deployment manages the complete
list of permissions for the role.

https://auth0.com/docs/manage-users/access-control/rbac
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/role
https://www.pulumi.com/registry/packages/auth0/api-docs/role/

## Example

```yaml
# Auth0 Role Test Manifest
# This file is used for testing the Auth0Role component
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: Your Auth0 tenant domain (e.g., your-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:roles
#    - read:roles
#    - update:roles
#    - delete:roles
#    - read:resource_servers
#
# 3. The referenced resource server identifier and its scopes must already exist
#    in the tenant (e.g., created via the Auth0ResourceServer component).

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Role
metadata:
  name: test-viewer-role
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # Optional: friendly display name (defaults to metadata.name)
  name: Test Viewer

  # Optional: description
  description: Read-only access for testing the Auth0Role component

  # Authoritative set of permissions granted to the role
  permissions:
    - name: read:items
      resource_server_identifier: https://api.test.planton.dev/
    - name: read:orders
      resource_server_identifier: https://api.test.planton.dev/
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.permissions` | `[]Auth0RolePermission` |  |  |  |
| `spec.permissions[].name` | `string` | yes |  |  |
| `spec.permissions[].resourceServerIdentifier` | `string` | yes |  |  |

## Field Details

### spec.name

`string`

name is the human-readable name of the role (e.g., "Administrator", "Editor").
Shown in the Auth0 dashboard and used when assigning the role to users.
When omitted, the role name defaults to metadata.name.

### spec.description

`string`

description is a human-readable explanation of what the role grants.
Shown in the Auth0 dashboard alongside the role name.
Example: "Full administrative access to the orders API".

### spec.permissions

`[]Auth0RolePermission`

permissions is the set of API permissions (scopes) granted to this role.
Each permission references a scope defined on an Auth0 Resource Server (API),
identified by the scope name and the resource server's identifier (audience).

The set is authoritative: the deployment manages the complete list of
permissions for the role. Omitting a previously-assigned permission removes it
from the role. When empty, the role is created with no permissions (it can
still be assigned to users and have permissions added later out-of-band).

Example:
  permissions:
    - name: read:orders
      resource_server_identifier: https://api.example.com/
    - name: write:orders
      resource_server_identifier: https://api.example.com/

https://auth0.com/docs/manage-users/access-control/configure-core-rbac/roles/create-roles

- rule: Each permission needs a name -- the scope as defined on the resource server, e.g. 'read:orders'.
- rule: Each permission must say which API it belongs to -- the resource server identifier (audience), e.g. 'https://api.example.com/'.

### spec.permissions[].name

`string` · required

name is the permission (scope) name as defined on the resource server.
This is the same value that appears in an access token's "scope" claim.
Example: "read:orders", "write:users", "delete:products".

- rule: {"required":true}

### spec.permissions[].resourceServerIdentifier

`string` · required

resource_server_identifier is the identifier (audience) of the Auth0 Resource
Server that owns this permission. This is the API's unique identifier, typically
a URI -- the same value used as the "audience" parameter in authorization requests.
Example: "https://api.example.com/".

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Role, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the Auth0 role (e.g., "rol_abc123"). Assigned by Auth0 and used to reference the role when assigning it to users and in Management API calls. |
| `status.outputs.name` | `string` | name is the human-readable name of the role. Reflects spec.name, or metadata.name when spec.name is omitted. |
| `status.outputs.description` | `string` | description is the human-readable description of the role. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0User | `spec.roles` | `status.outputs.id` |

## See Also

- [Overview](../README.md)

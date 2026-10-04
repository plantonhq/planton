# GcpGkeFleetScope

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpGkeFleetScopeSpec declares a team scope in a GKE fleet
(`google_gke_hub_scope`) together with everything that defines the
team's slice of the fleet: its fleet namespaces
(`google_gke_hub_namespace`), who gets which access
(`google_gke_hub_scope_rbac_role_binding`), and which clusters the team
may use (`google_gke_hub_membership_binding`).

A scope is the unit of fleet team management. Binding a cluster to it
creates the scope's namespaces on that cluster and applies its role
bindings there; adding a namespace creates it on every bound cluster.
The namespaces, role bindings, and cluster bindings have no meaning
outside their scope and are owned by whoever owns the team's slice, so
they live here rather than as blocks of their own.

Clusters are bound by their fleet membership's full name, whichever way
they joined: a GcpGkeCluster with fleet_project exports it as
fleet_membership, an explicit GcpGkeFleetMembership as name.

Important behavioral notes:

  - The scope lives in a fleet: Google requires the fleet first, so
    project_id references the GcpGkeFleet by default.
  - scope_id and every child's ID are create-time decisions; labels,
    namespace labels, role-binding principals and roles, and a cluster
    binding's target update in place.
  - Destroy deletes the bindings, the namespaces (removing them from
    every bound cluster), and the scope.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetScope
metadata:
  name: team-orders
spec:
  projectId:
    value: my-gcp-project
  scopeId: orders
  namespaceLabels:
    team: orders
  namespaces:
    - scopeNamespaceId: orders-api
    - scopeNamespaceId: orders-workers
  rbacRoleBindings:
    - scopeRbacRoleBindingId: orders-devs
      group:
        value: orders-devs@example.com
      role:
        predefinedRole: EDIT
    - scopeRbacRoleBindingId: orders-oncall
      user:
        value: oncall@example.com
      role:
        predefinedRole: VIEW
  membershipBindings:
    - membershipBindingId: orders-uc1
      membership:
        value: projects/my-gcp-project/locations/us-central1/memberships/orders-uc1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpGkeFleet (`status.outputs.project_id`) |
| `spec.scopeId` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.namespaceLabels` | `map<string, string>` |  |  |  |
| `spec.namespaces` | `[]GcpGkeFleetScopeNamespace` |  |  |  |
| `spec.namespaces[].scopeNamespaceId` | `string` | yes |  |  |
| `spec.namespaces[].labels` | `map<string, string>` |  |  |  |
| `spec.namespaces[].namespaceLabels` | `map<string, string>` |  |  |  |
| `spec.rbacRoleBindings` | `[]GcpGkeFleetScopeRbacRoleBinding` |  |  |  |
| `spec.rbacRoleBindings[].scopeRbacRoleBindingId` | `string` | yes |  |  |
| `spec.rbacRoleBindings[].user` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.rbacRoleBindings[].group` | `string \| valueFrom` |  |  | GcpCloudIdentityGroup (`status.outputs.group_email`) |
| `spec.rbacRoleBindings[].role` | `GcpGkeFleetScopeRole` | yes |  |  |
| `spec.rbacRoleBindings[].role.predefinedRole` | `string` |  |  |  |
| `spec.rbacRoleBindings[].role.customRole` | `string` |  |  |  |
| `spec.rbacRoleBindings[].labels` | `map<string, string>` |  |  |  |
| `spec.membershipBindings` | `[]GcpGkeFleetScopeMembershipBinding` |  |  |  |
| `spec.membershipBindings[].membershipBindingId` | `string` | yes |  |  |
| `spec.membershipBindings[].membership` | `string \| valueFrom` | yes |  | GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`) |
| `spec.membershipBindings[].labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The fleet host project the scope lives in: a literal project ID or a
GcpGkeFleet reference (the fleet the scope belongs to, which Google
requires first and a chart then orders before the scope). Empty means
the provider's default project. Immutable.

- references: GcpGkeFleet (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleet, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scopeId

`string`

The scope's ID, unique in the fleet; usually the team's name. 1-63
lowercase letters, digits, and hyphens, starting and ending with a
letter or digit. Defaults to metadata.name. Immutable.

- rule: scope_id must be 1-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit

### spec.labels

`map<string, string>`

Labels on the scope resource itself. The platform attribution labels
are added on top and win on key conflicts.

### spec.namespaceLabels

`map<string, string>`

Kubernetes labels Google applies to every namespace of this scope on
every bound cluster. On a key collision with a namespace's own
namespace_labels, this scope-level value wins.

### spec.namespaces

`[]GcpGkeFleetScopeNamespace`

The scope's fleet namespaces. Each is created as a Kubernetes namespace
on every cluster bound to the scope (onboarding a namespace that
already exists there).

### spec.namespaces[].scopeNamespaceId

`string` · required

The namespace's name, which is also the Kubernetes namespace created
on every bound cluster: a DNS label (1-63 lowercase letters, digits,
and hyphens). Google reserves the system namespaces (default,
kube-system, gke-connect, istio-system, config-management-system, and
the others the spec refuses). Immutable.

- rule: scope_namespace_id must be a DNS label: 1-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit
- rule: scope_namespace_id is a namespace Google reserves for the system
- rule: {"required":true}

### spec.namespaces[].labels

`map<string, string>`

Labels on the fleet namespace resource. The platform attribution
labels are added on top and win on key conflicts.

### spec.namespaces[].namespaceLabels

`map<string, string>`

Kubernetes labels Google applies to this namespace on every bound
cluster; the scope's namespace_labels win on a key collision.

### spec.rbacRoleBindings

`[]GcpGkeFleetScopeRbacRoleBinding`

Who gets which Kubernetes access within the scope's namespaces on the
bound clusters.

- rule: a role binding names exactly one principal: user or group

### spec.rbacRoleBindings[].scopeRbacRoleBindingId

`string` · required

The binding's ID, unique in the scope. 1-63 lowercase letters, digits,
and hyphens. Immutable.

- rule: scope_rbac_role_binding_id must be 1-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit
- rule: {"required":true}

### spec.rbacRoleBindings[].user

`string | valueFrom`

A user, as the clusters see it: a person's email ("alice@example.com")
or a service account's email -- a GcpServiceAccount reference, for the
CI or workload identity that deploys into the team's namespaces.
Exactly one of user or group.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.rbacRoleBindings[].group

`string | valueFrom`

A Google group's email, the usual way to give a whole team access:
a literal ("team-a@example.com") or a GcpCloudIdentityGroup reference.
Exactly one of user or group.

- references: GcpCloudIdentityGroup (`status.outputs.group_email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudIdentityGroup, name: <that resource's name>, fieldPath: status.outputs.group_email}} -- a bare string does not parse

### spec.rbacRoleBindings[].role

`GcpGkeFleetScopeRole` · required

The access granted. Required.

- rule: {"required":true}
- rule: a role is exactly one of predefined_role or custom_role

### spec.rbacRoleBindings[].role.predefinedRole

`string`

One of Google's predefined roles, applied in each of the scope's
namespaces:
  "ADMIN" -- full control of the namespace, including its RBAC
  "EDIT"  -- read and write most objects; no RBAC changes
  "VIEW"  -- read-only

- rule: predefined_role must be ADMIN, EDIT, or VIEW

### spec.rbacRoleBindings[].role.customRole

`string`

The name of a Kubernetes ClusterRole on the bound clusters. Google
honors it only when the fleet's rbacrolebindingactuation feature
(GcpGkeFleetFeature with feature "rbacrolebindingactuation") lists it
in allowed_custom_roles.

### spec.rbacRoleBindings[].labels

`map<string, string>`

Labels on the role binding resource. The platform attribution labels
are added on top and win on key conflicts.

### spec.membershipBindings

`[]GcpGkeFleetScopeMembershipBinding`

The clusters the team may use, each by its fleet membership.

### spec.membershipBindings[].membershipBindingId

`string` · required

The binding's ID, unique in the scope. 1-63 lowercase letters, digits,
and hyphens. Immutable.

- rule: membership_binding_id must be 1-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit
- rule: {"required":true}

### spec.membershipBindings[].membership

`string | valueFrom` · required

The cluster's fleet membership, by full name
("projects/{p}/locations/{l}/memberships/{id}"): a GcpGkeFleetMembership
reference, a GcpGkeCluster reference (its fleet_membership output, for
a cluster that joined through fleet_project), or the literal name. The
membership must be in this scope's fleet. Both modules derive the
binding's project, location, and membership ID from it. Immutable.

- references: GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`)
- rule: membership must be a full membership name: projects/{project}/locations/{location}/memberships/{id}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleetMembership, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.membershipBindings[].labels

`map<string, string>`

Labels on the binding resource. The platform attribution labels are
added on top and win on key conflicts.

### spec.deletionPolicy

`string`

What destroy does, for the scope and everything folded into it:
  "" / "DELETE" -- bindings, namespaces, and the scope are deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.unique_namespace_ids`: scope_namespace_id must be unique within the scope
- `spec.unique_rbac_role_binding_ids`: scope_rbac_role_binding_id must be unique within the scope
- `spec.unique_membership_binding_ids`: membership_binding_id must be unique within the scope

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeFleetScope, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/global/scopes/{scope_id}. |
| `status.outputs.scope_id` | `string` | The scope's ID. |
| `status.outputs.uid` | `string` | Google's unique identifier for the scope. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpGkeFleet | `status.outputs.project_id` |
| `spec.rbacRoleBindings[].user` | GcpServiceAccount | `status.outputs.email` |
| `spec.rbacRoleBindings[].group` | GcpCloudIdentityGroup | `status.outputs.group_email` |
| `spec.membershipBindings[].membership` | GcpGkeFleetMembership | `status.outputs.name` |
| `spec.membershipBindings[].membership` | GcpGkeCluster | `status.outputs.fleet_membership` |

## See Also

- [Overview](../README.md)

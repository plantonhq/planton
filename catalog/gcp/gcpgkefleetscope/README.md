# GCP GKE Fleet Scope

Declares a team scope in a GKE fleet together with the team's slice of it: fleet namespaces created on every bound cluster, who gets which Kubernetes access in them, and which clusters the team may use. Clusters are bound by their fleet membership, whichever way they joined -- a `GcpGkeCluster` with `fleetProject` (its `fleet_membership` output) or an explicit `GcpGkeFleetMembership`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `gkehub.googleapis.com` on the fleet host project (never disabled on destroy)
- **Scope** -- one `gke_hub_scope`
- **Fleet namespaces** -- one `gke_hub_namespace` per `namespaces` entry
- **Role bindings** -- one `gke_hub_scope_rbac_role_binding` per `rbacRoleBindings` entry
- **Cluster bindings** -- one `gke_hub_membership_binding` per `membershipBindings` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/gkehub.admin` (or the scope permissions in `iac/permissions.yaml`) on the fleet host project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpGkeFleet`** -- Google requires the fleet before a scope; reference its `project_id` from `projectId`.

### Optional Dependencies

- **`GcpGkeCluster`** / **`GcpGkeFleetMembership`** -- clusters to bind (`membershipBindings[].membership`).
- **`GcpCloudIdentityGroup`** / **`GcpServiceAccount`** -- role-binding principals.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetScope
metadata:
  name: team-orders
spec:
  projectId:
    valueFrom:
      kind: GcpGkeFleet
      name: platform-fleet
      fieldPath: status.outputs.project_id
  scopeId: orders
  namespaces:
    - scopeNamespaceId: orders-api
  rbacRoleBindings:
    - scopeRbacRoleBindingId: orders-devs
      group:
        value: orders-devs@example.com
      role:
        predefinedRole: EDIT
  membershipBindings:
    - membershipBindingId: prod-us-east1
      membership:
        valueFrom:
          kind: GcpGkeCluster
          name: prod-us-east1
          fieldPath: status.outputs.fleet_membership
```

```shell
planton apply -f gke-fleet-scope.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The fleet host project (`GcpGkeFleet` ref). Immutable. |
| `scopeId` | `string` | `metadata.name` | The scope's ID, usually the team's name. Immutable. |
| `labels` | `map` | -- | Labels on the scope resource. |
| `namespaceLabels` | `map` | -- | Kubernetes labels on every namespace of the scope (they win on key collisions). |
| `namespaces[]` | list | -- | `scopeNamespaceId` (a DNS label; system names refused), `labels`, `namespaceLabels`. |
| `rbacRoleBindings[]` | list | -- | `scopeRbacRoleBindingId`; exactly one of `user` or `group`; `role` with exactly one of `predefinedRole` (`ADMIN`, `EDIT`, `VIEW`) or `customRole`; `labels`. |
| `membershipBindings[]` | list | -- | `membershipBindingId`; `membership` (a full membership name or reference); `labels`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`, fanned to every child. |

### Validation Rules

- Child IDs are unique within the scope; namespace IDs are DNS labels and not one of Google's reserved system namespaces.
- A role binding names exactly one principal and exactly one role form.
- A literal membership is `projects/{project}/locations/{location}/memberships/{id}`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/global/scopes/{scope_id}` |
| `scope_id` | `string` | The scope's ID |
| `uid` | `string` | Google's unique identifier for the scope |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Bindings live with the scope.** A cluster that joined through `fleetProject` has a membership Google created; binding it here is the only way to give it to a team, and the same shape serves explicit memberships.
- **Same fleet only.** A bound membership must be in this scope's fleet project.
- **Custom roles need an allowlist.** Google honors `customRole` only when the fleet's `rbacrolebindingactuation` feature lists it.
- **Destroy** deletes the bindings, the namespaces (removing them from every bound cluster), and the scope.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpGkeFleet** -- the fleet the scope lives in
- **GcpGkeCluster** -- clusters that join through `fleetProject` (bind through `fleet_membership`)
- **GcpGkeFleetMembership** -- explicitly registered clusters
- **GcpGkeFleetFeature** -- `rbacrolebindingactuation` allowlists custom roles

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

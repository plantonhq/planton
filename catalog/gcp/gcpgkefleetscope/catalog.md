# GCP GKE Fleet Scope

Gives a team its slice of a shared GKE fleet. Declare the team's namespaces once and they appear on every cluster bound to the scope; grant the team's group edit or admin access once and it applies in those namespaces on every cluster; add a cluster to the scope and the team's namespaces and access follow it there.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- the Fleet API on the fleet host project
- **Scope** -- the team scope
- **Fleet namespaces, role bindings, and cluster bindings** -- one resource per declared entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with fleet admin permissions on the fleet host project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP GKE Fleet Scope**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team Scope** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetScope
metadata:
  name: team-orders
  org: acme-corp
  env: prod
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
```

```shell
planton apply -f gke-fleet-scope.yaml
```

This gives the orders team a namespace and edit access across the clusters bound to its scope. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference the fleet's `status.outputs.project_id` from `projectId`, and each cluster's `status.outputs.fleet_membership` (a `GcpGkeCluster` with `fleetProject`) or `status.outputs.name` (a `GcpGkeFleetMembership`) from `membershipBindings[].membership`.

## Key Configuration

These are the most important decisions when configuring a team scope. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespaces** -- the Kubernetes namespaces the team owns on every bound cluster.

**Access** -- one role binding per group or user, with Google's `ADMIN`, `EDIT`, or `VIEW` roles, or an allowlisted custom ClusterRole.

**Clusters** -- which fleet clusters the team may use.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGkeFleet** | `projectId` | `status.outputs.project_id` |
| **GcpGkeCluster** | `membershipBindings[].membership` | `status.outputs.fleet_membership` |
| **GcpGkeFleetMembership** | `membershipBindings[].membership` | `status.outputs.name` |
| **GcpCloudIdentityGroup** | `rbacRoleBindings[].group` | `status.outputs.group_email` |
| **GcpServiceAccount** | `rbacRoleBindings[].user` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The scope's resource name | Audits, tooling |
| `scope_id` | The scope's ID | Tooling that addresses the scope by ID |
| `uid` | The scope's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team scope** -- namespaces and group access. Start from the **Team Scope** preset.

**Team on named clusters** -- the same, bound to specific clusters. Start from the **Team Scope with Clusters** preset.

## Works With

- [**GCP GKE Fleet**](/infra-catalog/gcp-gke-fleet) -- the fleet the scope lives in
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- clusters to bind
- [**GCP GKE Fleet Membership**](/infra-catalog/gcp-gke-fleet-membership) -- explicitly registered clusters
- [**GCP Cloud Identity Group**](/infra-catalog/gcp-cloud-identity-group) -- the team's group

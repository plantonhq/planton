# GCP GKE Fleet Membership

Registers a cluster with a GKE fleet explicitly: a cluster created without `fleetProject`, a cluster created outside Planton, or a cluster in another project. Optionally turns on fleet Workload Identity through the cluster's OIDC issuer. The membership's `name` is what team scopes and per-cluster feature settings reference.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `gkehub.googleapis.com` on the fleet host project (never disabled on destroy)
- **Membership** -- one `gke_hub_membership`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/gkehub.admin` (or the membership permissions in `iac/permissions.yaml`), plus `container.clusters.get` on the cluster on the fleet host project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpGkeFleet`** -- declared first, so the registration never creates the fleet implicitly; reference its `project_id` from `projectId`.

### Optional Dependencies

- **`GcpGkeCluster`** -- the cluster to register (`gkeCluster`, its `cluster_id`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetMembership
metadata:
  name: legacy-cluster
spec:
  projectId:
    valueFrom:
      kind: GcpGkeFleet
      name: platform-fleet
      fieldPath: status.outputs.project_id
  gkeCluster:
    value: projects/my-gcp-project/locations/us-central1/clusters/legacy-cluster
```

```shell
planton apply -f gke-fleet-membership.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The fleet host project (`GcpGkeFleet` ref). Immutable. |
| `membershipId` | `string` | `metadata.name` | The membership's ID. Immutable. |
| `location` | `string` | `global` | Where the membership lives. Immutable. |
| `gkeCluster` | `string` / ref | -- | The cluster's ID (`GcpGkeCluster` ref, `cluster_id`); sent as Google's `//container.googleapis.com/...` resource link. Immutable. |
| `issuer` | `string` | -- | The cluster's OIDC issuer; turns on fleet Workload Identity. Immutable. |
| `labels` | `map` | -- | Labels on the membership. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `membershipId` is 1-63 lowercase letters, digits, or hyphens.
- `issuer` is an `https://` URL shorter than 2000 characters.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/memberships/{id}` |
| `membership_id` | `string` | The membership's ID |
| `location` | `string` | The membership's location |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Never for a cluster with `fleetProject`.** That cluster is already registered; reference its `fleet_membership` output instead.
- **The issuer's form.** For a GKE cluster it is `https://container.googleapis.com/v1/` followed by the cluster's `cluster_id`; Google requires the `locations/` form, so the cluster's `self_link` (which uses `zones/` for zonal clusters) does not fit.
- **Replacing the membership** drops its scope bindings and per-cluster feature settings with it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpGkeFleet** -- the fleet the cluster joins
- **GcpGkeCluster** -- the cluster, or the alternative registration path through `fleetProject`
- **GcpGkeFleetScope** -- binds the membership to a team
- **GcpGkeFleetFeature** -- per-cluster settings through `membershipConfigs`

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GCP GKE Fleet Feature

Turns on one GKE fleet feature and configures it: Config Sync (`configmanagement`), Policy Controller (`policycontroller`), Cloud Service Mesh (`servicemesh`), multi-cluster ingress and service discovery, fleet logging (`fleetobservability`), upgrade sequencing (`clusterupgrade`), custom roles for team scopes (`rbacrolebindingactuation`), and fleet tenancy's workload identity (`workloadidentity`). Fleet-wide defaults reach every cluster in the fleet, including clusters that join later; `membershipConfigs` overrides them for named clusters.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `gkehub.googleapis.com` and the feature's own API (never disabled on destroy)
- **Feature** -- one `gke_hub_feature`
- **Per-cluster settings** -- one `gke_hub_feature_membership` per `membershipConfigs` entry (an entry of the feature's map in Google, not an object of its own)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/gkehub.admin` (or the feature permissions in `iac/permissions.yaml`) on the fleet host project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpGkeFleet`** -- the fleet the feature configures; reference its `project_id` from `projectId`.

### Optional Dependencies

- **`GcpGkeCluster`** / **`GcpGkeFleetMembership`** -- clusters for `membershipConfigs` and the multi-cluster ingress config cluster.
- **`GcpServiceAccount`** -- Config Sync's repository and metrics identities.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetFeature
metadata:
  name: config-sync
spec:
  projectId:
    valueFrom:
      kind: GcpGkeFleet
      name: platform-fleet
      fieldPath: status.outputs.project_id
  feature: configmanagement
  fleetDefaultMemberConfig:
    configmanagement:
      management: MANAGEMENT_AUTOMATIC
      configSync:
        enabled: true
        sourceFormat: unstructured
        git:
          secretType: none
          syncRepo: https://github.com/example/platform-config
          syncBranch: main
```

```shell
planton apply -f gke-fleet-feature.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `feature` | `string` | The feature's Google name, e.g. `configmanagement`, `policycontroller`, `servicemesh`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The fleet host project (`GcpGkeFleet` ref). Immutable. |
| `location` | `string` | `global` | Immutable. |
| `labels` | `map` | -- | Labels on the feature. |
| `multiclusteringress` | object | -- | `configMembership` -- the config cluster's membership. |
| `fleetobservability` | object | -- | `loggingConfig.defaultConfig.mode`, `loggingConfig.fleetScopeLogsConfig.mode` (`COPY` or `MOVE`). |
| `clusterupgrade` | object | -- | `upstreamFleets` (one fleet today), `postConditions.soaking`, `gkeUpgradeOverrides[]`. |
| `rbacrolebindingactuation` | object | -- | `allowedCustomRoles` for team scopes. |
| `workloadidentity` | object | -- | `scopeTenancyPool` (`GcpWorkloadIdentityPool` ref). |
| `fleetDefaultMemberConfig` | object | -- | `configmanagement`, `mesh`, or `policycontroller` defaults for every cluster. |
| `membershipConfigs[]` | list | -- | `membership` plus one of `configmanagement` (adds `stopSyncing`, `deploymentOverrides`), `mesh`, or `policycontroller`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`, fanned to the per-cluster settings. |

### Validation Rules

- Each settings block is accepted only on its feature (`mesh` on `servicemesh`); Google's API holds one spec per feature.
- `membershipConfigs` only on `configmanagement`, `servicemesh`, and `policycontroller`; exactly one block per entry; a membership at most once.
- Config Sync syncs from `git` or `oci`, with a required `secretType` and `syncRepo`; Policy Controller requires `installSpec`; Mesh requires `management`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/features/{feature}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Creating adopts.** A feature that is already on is taken over: its settings become this block's.
- **Destroy turns the feature off**, except `rbacrolebindingactuation`, which Google never deletes -- destroy empties its allowlist.
- **Turned-down levers are not offered:** Hierarchy Controller (Config Sync 1.20 removed it) and Policy Controller through Config Management (Config Sync 1.21) -- use the `policycontroller` feature.
- **Cost.** Config Management, Policy Controller, team management, and upgrade sequencing carry no charge of their own; Managed Service Mesh and Multi Cluster Ingress have their own SKUs.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpGkeFleet** -- the fleet the feature configures
- **GcpGkeFleetScope** -- `customRole` bindings need the `rbacrolebindingactuation` allowlist
- **GcpGkeCluster** / **GcpGkeFleetMembership** -- clusters the per-cluster settings target

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

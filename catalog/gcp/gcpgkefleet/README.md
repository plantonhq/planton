# GCP GKE Fleet

Declares a project's GKE fleet: the one fleet a fleet host project holds, its display name, and the defaults every cluster that joins it inherits (Binary Authorization evaluation and GKE security posture). Team scopes, fleet features, and memberships live inside it and reference it from their `projectId`, which orders them after the fleet in a chart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `gkehub.googleapis.com` (the Fleet API) on the host project (never disabled on destroy)
- **Fleet** -- the project's one fleet (`default` in `global`)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/gkehub.admin` (or the fleet permissions in `iac/permissions.yaml`) on the host project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpProject`** -- the fleet host project, by reference (`projectId`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleet
metadata:
  name: platform-fleet
spec:
  projectId:
    value: my-gcp-project
  displayName: Platform fleet
  defaultClusterConfig:
    securityPostureConfig:
      mode: BASIC
      vulnerabilityMode: VULNERABILITY_BASIC
```

```shell
planton apply -f gke-fleet.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The fleet host project (`GcpProject` ref). Immutable. |
| `displayName` | `string` | derived from the project | 4-30 characters of letters, digits, hyphens, quotes, spaces, `!`. |
| `defaultClusterConfig.binaryAuthorizationConfig.evaluationMode` | `string` | Google's setting | `DISABLED` or `POLICY_BINDINGS`. |
| `defaultClusterConfig.binaryAuthorizationConfig.policyBindings` | `string[]` | -- | GKE platform policies, `projects/{number}/platforms/gke/policies/{id}`. |
| `defaultClusterConfig.securityPostureConfig.mode` | `string` | Google's setting | `DISABLED`, `BASIC`, or `ENTERPRISE`. |
| `defaultClusterConfig.securityPostureConfig.vulnerabilityMode` | `string` | Google's setting | `VULNERABILITY_DISABLED`, `VULNERABILITY_BASIC`, or `VULNERABILITY_ENTERPRISE`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `displayName` is 4-30 allowed characters.
- `policyBindings` are unique platform policy names and need `evaluationMode: POLICY_BINDINGS`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `project_id` | `string` | The host project's ID -- what every fleet child's `projectId` references |
| `name` | `string` | `projects/{project}/locations/global/fleets/default` |
| `uid` | `string` | Google's unique identifier for the fleet |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Declare the fleet first.** Google also creates the fleet implicitly when the first cluster registers in a project; a fleet declared after that collides with it. Every fleet child names this kind as its prerequisite.
- **One per project.** A project holds exactly one fleet.
- **Destroy deletes the fleet**, and Google refuses while memberships or scopes remain.
- **Labels and the compliance posture default are not offered yet:** the pinned Pulumi SDK (pulumi-gcp v9.37.0) lacks both, and an argument one engine cannot send is never a one-engine field. They arrive with pulumi-gcp v10.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpGkeFleetScope** -- team scopes inside the fleet
- **GcpGkeFleetFeature** -- fleet features (Config Sync, Policy Controller, Service Mesh, multi-cluster ingress, upgrade sequencing)
- **GcpGkeFleetMembership** and **GcpGkeCluster** (`fleetProject`) -- the two ways a cluster joins

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

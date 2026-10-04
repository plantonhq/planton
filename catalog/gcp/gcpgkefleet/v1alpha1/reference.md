# GcpGkeFleet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpGkeFleetSpec declares a project's GKE fleet (`google_gke_hub_fleet`):
the one fleet a fleet host project holds, its display name, and the
defaults every cluster that joins it inherits.

A fleet is how a platform team runs many clusters as one: team scopes
(GcpGkeFleetScope) give teams namespaces and access across the clusters
bound to them, fleet features (GcpGkeFleetFeature) turn on Config Sync,
Policy Controller, Cloud Service Mesh, multi-cluster ingress, and
upgrade sequencing fleet-wide, and memberships (GcpGkeFleetMembership,
or GcpGkeCluster.fleet_project) bring clusters in. Those blocks live
inside this one and reference it from their project_id.

Important behavioral notes:

  - A project holds exactly one fleet, always named "default" in
    "global". Google also creates it implicitly when the first cluster
    registers in a project that has none; declare this block BEFORE any
    cluster joins (every fleet child names it as its prerequisite), or
    its create collides with the implicit fleet.
  - Everything except the project updates in place.
  - Destroy deletes the fleet. Google refuses while memberships or
    scopes still exist, so a chart tears the children down first.
  - Fleet labels and the compliance posture default are not offered:
    the pinned Pulumi SDK (pulumi-gcp v9.37.0) lacks both, and an
    argument one engine cannot send is never a one-engine field. They
    arrive with pulumi-gcp v10.

## Example

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
    binaryAuthorizationConfig:
      evaluationMode: POLICY_BINDINGS
      policyBindings:
        - projects/123456789012/platforms/gke/policies/baseline
    securityPostureConfig:
      mode: BASIC
      vulnerabilityMode: VULNERABILITY_BASIC
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.defaultClusterConfig` | `GcpGkeFleetDefaultClusterConfig` |  |  |  |
| `spec.defaultClusterConfig.binaryAuthorizationConfig` | `GcpGkeFleetBinaryAuthorizationConfig` |  |  |  |
| `spec.defaultClusterConfig.binaryAuthorizationConfig.evaluationMode` | `string` |  |  |  |
| `spec.defaultClusterConfig.binaryAuthorizationConfig.policyBindings` | `[]string` |  |  |  |
| `spec.defaultClusterConfig.securityPostureConfig` | `GcpGkeFleetSecurityPostureConfig` |  |  |  |
| `spec.defaultClusterConfig.securityPostureConfig.mode` | `string` |  |  |  |
| `spec.defaultClusterConfig.securityPostureConfig.vulnerabilityMode` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The fleet host project: a literal project ID or a GcpProject
reference. Empty means the provider's default project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string`

The fleet's name as the console and gcloud show it. 4-30 characters of
letters, digits, hyphens, single or double quotes, spaces, and
exclamation points. Empty lets Google derive one from the project's
name.

- rule: display_name must be 4-30 characters of letters, digits, hyphens, quotes, spaces, or exclamation points

### spec.defaultClusterConfig

`GcpGkeFleetDefaultClusterConfig`

Defaults applied to every cluster in the fleet, existing and future.

### spec.defaultClusterConfig.binaryAuthorizationConfig

`GcpGkeFleetBinaryAuthorizationConfig`

Binary Authorization for every cluster in the fleet: which GKE
platform policies their workloads are evaluated against.

- rule: policy_bindings take effect only with evaluation_mode POLICY_BINDINGS

### spec.defaultClusterConfig.binaryAuthorizationConfig.evaluationMode

`string`

How workloads are evaluated:
  "DISABLED"        -- no Binary Authorization evaluation
  "POLICY_BINDINGS" -- workloads are audited against policy_bindings
Empty leaves Google's current setting.

- rule: evaluation_mode must be DISABLED or POLICY_BINDINGS

### spec.defaultClusterConfig.binaryAuthorizationConfig.policyBindings

`[]string`

The GKE platform policies to audit against, each the policy's
relative name: "projects/{project_number}/platforms/gke/policies/{policy_id}".
Platform policies are created through the Binary Authorization API or
console (no catalog block creates one), so these are names, not
references. Distinct from the project-level GcpBinaryAuthorizationPolicy.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^projects/[0-9]+/platforms/gke/policies/[^/]+$"}}}}

### spec.defaultClusterConfig.securityPostureConfig

`GcpGkeFleetSecurityPostureConfig`

GKE security posture for every cluster in the fleet: workload
configuration auditing and vulnerability scanning.

### spec.defaultClusterConfig.securityPostureConfig.mode

`string`

Workload configuration auditing:
  "DISABLED"   -- off
  "BASIC"      -- Google's standard configuration checks
  "ENTERPRISE" -- advanced posture capabilities; check Google's
                  security posture pricing before choosing it
Empty leaves Google's current setting.

- rule: mode must be DISABLED, BASIC, or ENTERPRISE

### spec.defaultClusterConfig.securityPostureConfig.vulnerabilityMode

`string`

Vulnerability scanning of running workloads:
  "VULNERABILITY_DISABLED", "VULNERABILITY_BASIC" (OS vulnerabilities),
  "VULNERABILITY_ENTERPRISE" (adds language-package vulnerabilities).
Empty leaves Google's current setting.

- rule: vulnerability_mode must be VULNERABILITY_DISABLED, VULNERABILITY_BASIC, or VULNERABILITY_ENTERPRISE

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the fleet is deleted (Google refuses while
                   memberships or scopes remain)
  "PREVENT"     -- destroy fails; a guard for a fleet teams depend on
  "ABANDON"     -- the fleet leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeFleet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_id` | `string` | The fleet host project's ID -- the value every fleet child's project_id references, so a chart places scopes, features, and memberships inside this fleet and orders them after it. |
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/global/fleets/default. |
| `status.outputs.uid` | `string` | Google's unique identifier for the fleet. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGkeCluster | `spec.fleetProject` | `status.outputs.project_id` |
| GcpGkeFleetFeature | `spec.projectId` | `status.outputs.project_id` |
| GcpGkeFleetFeature | `spec.clusterupgrade.upstreamFleets` | `status.outputs.project_id` |
| GcpGkeFleetMembership | `spec.projectId` | `status.outputs.project_id` |
| GcpGkeFleetScope | `spec.projectId` | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)

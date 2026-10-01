# GcpGkeFleetMembership

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGkeFleetMembershipSpec registers a cluster with a GKE fleet
(`google_gke_hub_membership`): the explicit form of joining a fleet.

There are two ways a GKE cluster joins a fleet, and a cluster uses
exactly one:

  - At creation, through GcpGkeCluster.fleet_project. Google creates the
    membership itself and the cluster exports its name as
    status.outputs.fleet_membership. This is the common path; no
    membership block is declared.
  - Explicitly, through this block: for a cluster created without
    fleet_project, a cluster created outside Planton, or a cluster in
    another project. Never declare this block for a cluster that sets
    fleet_project -- the cluster is already registered.

Either way the membership's full name is what team scopes
(GcpGkeFleetScope.membership_bindings) and per-cluster feature settings
(GcpGkeFleetFeature.membership_configs) reference.

Important behavioral notes:

  - membership_id, location, the cluster, and the issuer are
    create-time decisions; changing one replaces the membership (and
    drops its scope bindings and per-cluster feature settings with it).
    Labels update in place.
  - Destroy unregisters the cluster from the fleet; the cluster itself
    is untouched.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetMembership
metadata:
  name: orders-uc1
spec:
  projectId:
    value: my-gcp-project
  gkeCluster:
    value: projects/my-gcp-project/locations/us-central1/clusters/orders-uc1
  issuer: https://container.googleapis.com/v1/projects/my-gcp-project/locations/us-central1/clusters/orders-uc1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpGkeFleet (`status.outputs.project_id`) |
| `spec.membershipId` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.gkeCluster` | `string \| valueFrom` |  |  | GcpGkeCluster (`status.outputs.cluster_id`) |
| `spec.issuer` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The fleet host project the membership lives in: a literal project ID
or a GcpGkeFleet reference (the fleet this cluster joins, which orders
the membership after the fleet in a chart). Empty means the provider's
default project. Immutable.

- references: GcpGkeFleet (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleet, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.membershipId

`string`

The membership's ID, unique in the fleet; by convention the cluster's
name. 1-63 lowercase letters, digits, and hyphens, starting and ending
with a letter or digit. Defaults to metadata.name. Immutable.

- rule: membership_id must be 1-63 lowercase letters, digits, or hyphens, starting and ending with a letter or digit

### spec.location

`string`

Where the membership lives: "global" (the default, and what explicit
registration normally uses) or a region. Immutable.

- rule: location must be global or a region

### spec.gkeCluster

`string | valueFrom`

The GKE cluster to register: a GcpGkeCluster reference (its
cluster_id, "projects/{p}/locations/{l}/clusters/{name}") or that
literal path. Both modules send it as Google's resource link,
"//container.googleapis.com/projects/{p}/locations/{l}/clusters/{name}".
Omit it only for a cluster outside Google Cloud that registers through
its own Connect agent. Immutable.

- references: GcpGkeCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.issuer

`string`

The cluster's OIDC issuer, which turns on fleet Workload Identity for
this membership: Google then trusts tokens from this issuer within the
fleet's workload identity pool ("{project}.hub.id.goog"). For a GKE
cluster it is "https://container.googleapis.com/v1/" followed by the
cluster's cluster_id output -- Google requires the locations/ form, so
the cluster's self_link (zones/ for zonal clusters) does not fit.
Empty leaves fleet Workload Identity off. Immutable: Google refuses an
issuer change in place.

- rule: issuer must be an https:// URL shorter than 2000 characters

### spec.labels

`map<string, string>`

Labels on the membership. The platform attribution labels are added on
top and win on key conflicts.

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the cluster is unregistered from the fleet
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the membership leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeFleetMembership, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/memberships/{membership_id} -- what GcpGkeFleetScope.membership_bindings and GcpGkeFleetFeature.membership_configs reference. |
| `status.outputs.membership_id` | `string` | The membership's ID. |
| `status.outputs.location` | `string` | The membership's location. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpGkeFleet | `status.outputs.project_id` |
| `spec.gkeCluster` | GcpGkeCluster | `status.outputs.cluster_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDeployTarget | `spec.anthosCluster.membership` | `status.outputs.name` |
| GcpDeployTarget | `spec.associatedEntities[].anthosClusters[].membership` | `status.outputs.name` |
| GcpGkeFleetFeature | `spec.multiclusteringress.configMembership` | `status.outputs.name` |
| GcpGkeFleetFeature | `spec.membershipConfigs[].membership` | `status.outputs.name` |
| GcpGkeFleetScope | `spec.membershipBindings[].membership` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

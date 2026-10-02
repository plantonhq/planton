# GcpBigQueryReservation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryReservationSpec defines a BigQuery slot reservation
(`google_bigquery_reservation`) -- dedicated query capacity in an
administration project, with the assignments that route projects,
folders, or an organization onto it folded in (they live under exactly
one reservation).

Capacity model: slot_capacity is the baseline billed around the clock;
autoscale_max_slots adds slots only while queries need them, billed
while in use. Idle baseline slots are shared with other reservations in
the same admin project unless ignore_idle_slots is set. Slots bought as
a GcpBigQueryCapacityCommitment in the same admin project and location
lower the price of the baseline; a reservation group
(GcpBigQueryReservationGroup) pools idle slots among its members first.

Immutable: project_id, location, reservation_name, edition (a change
replaces the reservation). Capacity, autoscaling, concurrency, idle-slot
sharing, the group, the secondary location, and labels update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservation
metadata:
  name: analytics
spec:
  projectId:
    value: bq-admin-project
  location: US
  edition: ENTERPRISE
  # 100 slots always on, up to 300 more while queries need them.
  slotCapacity: 100
  autoscaleMaxSlots: 300
  assignments:
    - assignee:
        projectId:
          value: analytics-project
      jobType: QUERY
    - assignee:
        folderId:
          value: "123456789012"
      jobType: PIPELINE
  labels:
    team: data-platform
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  |  |  |
| `spec.reservationName` | `string` |  |  |  |
| `spec.slotCapacity` | `int64` |  |  |  |
| `spec.edition` | `string` |  |  |  |
| `spec.autoscaleMaxSlots` | `int64` |  |  |  |
| `spec.ignoreIdleSlots` | `bool` |  |  |  |
| `spec.concurrency` | `int64` |  |  |  |
| `spec.reservationGroup` | `string \| valueFrom` |  |  | GcpBigQueryReservationGroup (`status.outputs.name`) |
| `spec.secondaryLocation` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.assignments` | `[]GcpBigQueryReservationAssignment` |  |  |  |
| `spec.assignments[].assignee` | `GcpBigQueryReservationAssignee` | yes |  |  |
| `spec.assignments[].assignee.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.assignments[].assignee.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.assignments[].assignee.organizationId` | `string` |  |  |  |
| `spec.assignments[].jobType` | `string` | yes |  |  |
| `spec.assignments[].principal` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The administration project the reservation lives in and bills to: a
literal project ID or a GcpProject reference. If omitted, the
provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

Where the slots are: a multi-region (US, EU) or a region
(asia-northeast1). Assigned projects' jobs use the reservation only
for data in this location. Empty means US (Google's default).
Immutable.

### spec.reservationName

`string`

The reservation's name -- letters, digits, and dashes. Defaults to
metadata.name. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z0-9-]+$"}}

### spec.slotCapacity

`int64`

Baseline slots, always allocated and billed -- 0 is valid with
autoscaling, for pay-while-used capacity. Queries may exceed it by
borrowing idle slots unless ignore_idle_slots is set. Mutable in
place.

- rule: {"int64":{"gte":"0"}}

### spec.edition

`string`

The BigQuery edition, which sets the features and the per-slot price:
  STANDARD        -- autoscaling only, no commitments, fewer features
  ENTERPRISE      -- commitments, BigQuery ML, BI Engine, and more
  ENTERPRISE_PLUS -- adds managed disaster recovery and compliance
                     controls
Empty lets Google choose. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STANDARD","ENTERPRISE","ENTERPRISE_PLUS"]}}

### spec.autoscaleMaxSlots

`int64`

The most slots autoscaling may add on top of slot_capacity (a
multiple of 50). 0 or empty turns autoscaling off.

- rule: {"int64":{"gte":"0"}}

### spec.ignoreIdleSlots

`bool`

true caps queries at this reservation's own slots; false (Google's
default) lets them borrow idle slots from other reservations in the
same admin project.

### spec.concurrency

`int64`

The soft limit on queries running at once. 0 (Google's default) sizes
it automatically from the reservation's slots.

- rule: {"int64":{"gte":"0"}}

### spec.reservationGroup

`string | valueFrom`

The group the reservation belongs to, which pools idle slots among its
members before other reservations: a GcpBigQueryReservationGroup
reference (its name output) or a literal group name or full path.

- references: GcpBigQueryReservationGroup (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryReservationGroup, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.secondaryLocation

`string`

Managed disaster recovery: the location of a secondary replica
(Enterprise Plus). Setting it on create makes a failover reservation;
setting or clearing it later converts one way or the other. Failover
itself is an operational action outside the spec.

### spec.labels

`map<string, string>`

Labels on the reservation. The platform attribution labels are added
on top and win on a key conflict.

### spec.assignments

`[]GcpBigQueryReservationAssignment`

The projects, folders, or organization whose jobs use the
reservation, each for one job type. Each is unique by assignee,
job_type, and principal. An assignment is PENDING until the
reservation has capacity to give.

### spec.assignments[].assignee

`GcpBigQueryReservationAssignee` · required

Whose jobs use the reservation.

- rule: {"required":true}
- rule: set exactly one of project_id, folder_id, or organization_id

### spec.assignments[].assignee.projectId

`string | valueFrom`

A project: a literal project ID or a GcpProject reference. The modules
send projects/{id}.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.assignments[].assignee.folderId

`string | valueFrom`

A folder: its numeric ID, literal or a GcpFolder reference. The
modules send folders/{id}.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.assignments[].assignee.organizationId

`string`

An organization: the numeric organization ID, without the
organizations/ prefix. The modules send organizations/{id}.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.assignments[].jobType

`string` · required

Which jobs:
  QUERY      -- interactive and batch queries, and scripts
  PIPELINE   -- load, export, and copy jobs
  CONTINUOUS -- continuous queries

- rule: {"required":true,"string":{"in":["QUERY","PIPELINE","CONTINUOUS"]}}

### spec.assignments[].principal

`string`

Narrow the assignment to one principal's jobs: jobs that principal
runs use this reservation, everyone else's fall back to the
project/folder/organization assignment (then on-demand). Formats:
  principal://goog/subject/USER_EMAIL
  principal://iam.googleapis.com/projects/-/serviceAccounts/SA_EMAIL
  principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL/subject/SUBJECT
Empty covers every principal.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^principal://.+$"}}

### spec.deletionPolicy

`string`

What happens to the reservation and its assignments when this
resource is destroyed:
  "" / "DELETE" -- deleted (assigned jobs fall back to on-demand)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- they leave management and keep running (and
                   billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.assignments_unique`: each assignment must be unique by assignee, job_type, and principal

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryReservation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/reservations/{reservation_name}. |
| `status.outputs.reservation_name` | `string` | The reservation's name (the last segment of name). |
| `status.outputs.location` | `string` | The reservation's location. |
| `status.outputs.assignment_names` | `[]string` | The assignments' resource names, in the order they are declared. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.reservationGroup` | GcpBigQueryReservationGroup | `status.outputs.name` |
| `spec.assignments[].assignee.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.assignments[].assignee.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)

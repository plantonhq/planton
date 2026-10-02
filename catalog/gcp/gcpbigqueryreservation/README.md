# GCP BigQuery Reservation

A BigQuery slot reservation -- dedicated query capacity in an administration project, with baseline slots, autoscaling, idle-slot sharing, an optional reservation group, managed disaster recovery, and the assignments that route projects, folders, or an organization onto it folded in.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Reservation** -- a `bigquery_reservation` carrying the platform attribution labels
- **Assignments** -- one `bigquery_reservation_assignment` per `assignments` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the admin project and on every assignee on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservationGroup`** -- a group the reservation joins (`reservationGroup`).
- **`GcpBigQueryCapacityCommitment`** -- committed slots in the same admin project and location lower the price of the baseline.
- **`GcpProject` / `GcpFolder`** -- assignees, by reference.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservation
metadata:
  name: analytics
spec:
  location: US
  edition: ENTERPRISE
  slotCapacity: 100
  autoscaleMaxSlots: 300
  assignments:
    - assignee:
        projectId:
          value: analytics-project
      jobType: QUERY
```

```shell
planton apply -f bigquery-reservation.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `slotCapacity` | `int64` | Baseline slots, billed always; 0 is valid with autoscaling. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The admin project. |
| `location` | `string` | `US` | Multi-region or region. Immutable. |
| `reservationName` | `string` | `metadata.name` | Letters, digits, dashes. Immutable. |
| `edition` | `string` | Google's choice | `STANDARD`, `ENTERPRISE`, `ENTERPRISE_PLUS`. Immutable. |
| `autoscaleMaxSlots` | `int64` | off | Slots autoscaling may add above the baseline. |
| `ignoreIdleSlots` | `bool` | `false` | `true` caps queries at this reservation's own slots. |
| `concurrency` | `int64` | automatic | Soft limit on concurrent queries. |
| `reservationGroup` | `StringValueOrRef` | none | A `GcpBigQueryReservationGroup` reference. |
| `secondaryLocation` | `string` | none | Managed disaster recovery replica (Enterprise Plus). |
| `labels` | `map<string,string>` | `{}` | Labels; attribution labels win. |
| `assignments` | `list` | none | `assignee` (exactly one of `projectId`, `folderId`, `organizationId`), `jobType` (`QUERY`, `PIPELINE`, `CONTINUOUS`), optional `principal`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` -- applies to the assignments too. |

### Validation Rules

- Each assignment has exactly one assignee and is unique by assignee, job type, and principal.
- Job types, editions, and the numeric organization id follow Google's forms; principals start with `principal://`.
- Slot counts are non-negative.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/reservations/{name}` |
| `reservation_name` | `string` | The reservation's name |
| `location` | `string` | The location |
| `assignment_names` | `[]string` | The assignments' full names, in declared order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Assignments are immutable.** Any change to an assignment replaces it; its jobs fall back to on-demand for the moment in between.
- **Assigning needs rights on both sides.** Creating an assignment needs `bigquery.reservationAssignments.create` on the admin project and on the assignee.
- **On-demand pinning is not here.** Assigning a project to BigQuery's built-in `none` reservation (to force on-demand pricing) is a standalone assignment this block does not express.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpBigQueryCapacityCommitment** -- committed slots the reservation draws on
- **GcpBigQueryReservationGroup** -- idle-slot sharing among reservations
- **GcpBigQueryDataset** -- the data the assigned jobs query

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

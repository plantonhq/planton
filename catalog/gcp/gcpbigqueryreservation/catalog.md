# GCP BigQuery Reservation

Gives BigQuery workloads predictable, dedicated capacity: baseline slots that are always there, autoscaling that adds slots only while queries need them, and assignments that route whole projects, folders, or an organization onto the capacity -- in one block owned by the team that manages the BigQuery bill.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Reservation** -- a `bigquery_reservation` carrying the platform attribution labels
- **Assignments** -- one `bigquery_reservation_assignment` per `assignments` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the admin project and on every assignee on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservationGroup`** -- a group the reservation joins (`reservationGroup`).
- **`GcpBigQueryCapacityCommitment`** -- committed slots in the same admin project and location lower the price of the baseline.
- **`GcpProject` / `GcpFolder`** -- assignees, by reference.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Reservation**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Autoscale Only** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservation
metadata:
  name: analytics
  org: acme-corp
  env: prod
spec:
  location: US
  edition: ENTERPRISE
  slotCapacity: 100
  autoscaleMaxSlots: 300
  assignments:
    - assignee:
        projectId:
          value: acme-analytics
      jobType: QUERY
```

```shell
planton apply -f bigquery-reservation.yaml
```

This reserves 100 always-on Enterprise slots in `US`, lets autoscaling add up to 300 more, and routes `acme-analytics`' queries onto them. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference assignee projects and folders as `GcpProject` / `GcpFolder`, and a `GcpBigQueryReservationGroup` from `reservationGroup`. Keep `GcpBigQueryCapacityCommitment` blocks in the same admin project.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Baseline and autoscale** -- `slotCapacity` bills around the clock; `autoscaleMaxSlots` adds slots only while queries use them. A zero baseline with autoscaling is pay-while-used capacity.

**Edition** -- `STANDARD` is autoscale-only; `ENTERPRISE` adds commitments and more features; `ENTERPRISE_PLUS` adds managed disaster recovery. It cannot change.

**Assignments** -- route a project, folder, or organization -- or one principal inside them -- onto the reservation per job type; anything unassigned runs on-demand.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId`, `assignments[].assignee.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `assignments[].assignee.folderId` | `status.outputs.folder_id` |
| **GcpBigQueryReservationGroup** | `reservationGroup` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The reservation's full name | Audit, console links |
| `reservation_name` | The name | BigQuery job configuration (`reservation`) |
| `assignment_names` | The assignments | Audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Pay while used** -- A zero-baseline Standard reservation that autoscales, assigned to one project. Start from the **Autoscale Only** preset.

**Production capacity** -- Always-on Enterprise slots with autoscale headroom, assigned to a folder and a project. Start from the **Baseline with Assignments** preset.

**Disaster recovery** -- An Enterprise Plus reservation replicated to a secondary region. Start from the **Managed Disaster Recovery** preset.

## Works With

- [**GCP BigQuery Capacity Commitment**](/cloud-catalog/gcp-bigquery-capacity-commitment) -- discounted baseline slots
- [**GCP BigQuery Reservation Group**](/cloud-catalog/gcp-bigquery-reservation-group) -- idle-slot sharing
- [**GCP Project**](/cloud-catalog/gcp-project) -- assignees

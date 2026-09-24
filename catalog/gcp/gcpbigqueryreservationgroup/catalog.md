# GCP BigQuery Reservation Group

Groups BigQuery reservations so they lend idle slots to each other first -- keeping spare capacity inside a team or tier before it flows to the rest of the admin project.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Reservation group** -- a `bigquery_reservation_group`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservation`** -- the reservations that join the group.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Reservation Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team Tier** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservationGroup
metadata:
  name: tier-1
  org: acme-corp
  env: prod
spec:
  location: US
```

```shell
planton apply -f bigquery-reservation-group.yaml
```

This creates the `tier-1` group in `US` for reservations to join. A Stack Job tracks the provisioning in real time.

### InfraChart

Declare the group next to the `GcpBigQueryReservation` blocks that reference its `name` output from `reservationGroup`.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Membership** -- reservations join from their own `reservationGroup`; the group lists nothing itself.

**Sharing** -- idle slots go to the group's other reservations before anyone else in the admin project.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The group's full name | A reservation's `reservationGroup` |
| `reservation_group_name` | The name | Console |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team tier** -- A group for one team's or one tier's reservations. Start from the **Team Tier** preset.

**Protected production tier** -- An EU group that production reservations depend on, protected by `PREVENT`. Start from the **Protected EU Group** preset.

## Works With

- [**GCP BigQuery Reservation**](/cloud-catalog/gcp-bigquery-reservation) -- the members

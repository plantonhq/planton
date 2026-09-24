# GCP BigQuery Reservation Group

A BigQuery reservation group -- a named set of reservations in one administration project and location that share idle slots with each other before any other reservation. Reservations join by referencing the group from their `reservationGroup` field; the group itself holds no capacity.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Reservation group** -- a `bigquery_reservation_group`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservation`** -- the reservations that join the group.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservationGroup
metadata:
  name: tier-1
spec:
  location: US
```

```shell
planton apply -f bigquery-reservation-group.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| -- | -- | Nothing is required; the name defaults to `metadata.name`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The admin project. |
| `location` | `string` | `US` | Immutable. |
| `reservationGroupName` | `string` | `metadata.name` | Letters, digits, dashes. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `reservationGroupName` is letters, digits, and dashes.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/reservationGroups/{name}` -- what a reservation's `reservationGroup` takes |
| `reservation_group_name` | `string` | The name |
| `location` | `string` | The location |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Same project and location.** A group and its reservations share an admin project and location.
- **Remove members before destroy.** Reservations referencing the group depend on it, so in one environment they are destroyed first.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpBigQueryReservation** -- the reservations that join the group

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

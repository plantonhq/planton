# GcpBigQueryReservationGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryReservationGroupSpec defines a BigQuery reservation group
(`google_bigquery_reservation_group`) -- a named set of reservations in
one administration project and location that share idle slots with each
other before any other reservation. Reservations join by referencing
this group from their reservation_group field; the group itself has no
capacity and costs nothing.

Immutable: project_id, location, reservation_group_name (a change
replaces the group).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryReservationGroup
metadata:
  name: tier-1
spec:
  projectId:
    value: bq-admin-project
  location: US
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  |  |  |
| `spec.reservationGroupName` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The administration project the group lives in (the same one as its
reservations): a literal project ID or a GcpProject reference. If
omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

The location of the group and its reservations: a multi-region (US,
EU) or a region. Empty means US (Google's default). Immutable.

### spec.reservationGroupName

`string`

The group's name -- letters, digits, and dashes. Defaults to
metadata.name. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z0-9-]+$"}}

### spec.deletionPolicy

`string`

What happens to the group when this resource is destroyed:
  "" / "DELETE" -- deleted (remove its reservations from it first)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the group leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryReservationGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/reservationGroups/{name}. What a reservation's reservation_group field references. |
| `status.outputs.reservation_group_name` | `string` | The group's name (the last segment of name). |
| `status.outputs.location` | `string` | The group's location. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBigQueryReservation | `spec.reservationGroup` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

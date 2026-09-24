# GCP BigQuery Capacity Commitment

A BigQuery capacity commitment -- slots bought for a fixed term at a discount in an administration project and location, pooled across every reservation there. **It is a purchase**: creating one starts a billed term, and Google refuses to delete it before the term ends.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Capacity commitment** -- a `bigquery_capacity_commitment`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservation`** -- reservations in the same admin project and location draw on the committed slots.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryCapacityCommitment
metadata:
  name: enterprise-annual-100
spec:
  location: US
  slotCount: 100
  plan: ANNUAL
  edition: ENTERPRISE
  deletionPolicy: ABANDON
```

```shell
planton apply -f bigquery-capacity-commitment.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `slotCount` | `int64` | Slots committed. Immutable. |
| `plan` | `string` | The commitment plan -- `ANNUAL` or `THREE_YEAR` with editions. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The admin project. |
| `location` | `string` | `US` | Immutable. |
| `capacityCommitmentId` | `string` | `metadata.name` | Lowercase letters, digits, dashes. Immutable. |
| `renewalPlan` | `string` | Google's default | The plan the commitment renews into. |
| `edition` | `string` | Google's choice | `STANDARD`, `ENTERPRISE`, `ENTERPRISE_PLUS`. Immutable. |
| `enforceSingleAdminProjectPerOrg` | `bool` | `false` | Fail if another project in the organization holds a commitment. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (refused before the term ends), `PREVENT`, or `ABANDON`. |

### Validation Rules

- `slotCount` is at least 1 and `plan` is required.
- `capacityCommitmentId` is lowercase letters, digits, and dashes, 1-64 characters, not starting or ending with a dash.
- `edition` takes only Google's three editions.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/capacityCommitments/{id}` |
| `state` | `string` | `PENDING`, `ACTIVE`, or `FAILED` |
| `commitment_start_time` | `string` | When the current term started |
| `commitment_end_time` | `string` | When it ends -- the earliest a delete succeeds |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **This spends money for a year or more.** Review `slotCount`, `plan`, and `edition` before applying -- none can be taken back.
- **Destroy fails before the term ends.** Use `deletionPolicy: ABANDON` so the block can leave management while the commitment runs out its term.
- **Commitments are pooled.** Every reservation in the same admin project and location draws on them; they belong to no single reservation.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpBigQueryReservation** -- the reservations that draw on the committed slots

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GCP BigQuery Capacity Commitment

A BigQuery capacity commitment -- slots bought for a fixed term at a discount in an administration project and location, pooled across every reservation there. **It is a purchase**: creating one starts a billed term, and Google refuses to delete it before the term ends.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

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
  renewalPlan: NONE
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
| `plan` | `string` | The commitment plan -- `ANNUAL` or `THREE_YEAR` with editions (Google's `CommitmentPlan` values). Moves only to a longer plan. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The admin project. |
| `location` | `string` | `US` | `US` or `EU` in upper case, or a region in lower case. Immutable. |
| `capacityCommitmentId` | `string` | `metadata.name` | Lowercase letters, digits, dashes. Immutable. |
| `renewalPlan` | `string` | Google's default | `NONE` ends the commitment at its term (needs `edition`); `ANNUAL` or `THREE_YEAR` renew it. |
| `edition` | `string` | Google's choice | `STANDARD`, `ENTERPRISE`, `ENTERPRISE_PLUS`. Immutable. |
| `enforceSingleAdminProjectPerOrg` | `bool` | `false` | Google's single-admin-project guard -- never sent by the pinned provider, so it guards nothing; changing it replaces the commitment. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (refused before the term ends), `PREVENT`, or `ABANDON`. |

### Validation Rules

- `slotCount` is at least 1 and `plan` is required.
- `plan` and `renewalPlan` take Google's `CommitmentPlan` values; `NONE` is a renewal plan only, and needs `edition`.
- `location` is `US`, `EU`, or a lowercase region.
- `capacityCommitmentId` is lowercase letters, digits, and dashes, 1-64 characters, not starting or ending with a dash.
- `edition` takes only Google's three editions.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/capacityCommitments/{id}` |
| `state` | `string` | `PENDING`, `ACTIVE`, or `FAILED` |
| `commitment_start_time` | `string` | When the original commitment started; renewals don't move it |
| `commitment_end_time` | `string` | When the current term ends -- a delete before it fails; a renewal moves it |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **This spends money for a year or more.** Review `slotCount`, `plan`, and `edition` before applying -- none can be taken back.
- **A commitment cannot be deleted before its term ends.** Use `deletionPolicy: ABANDON` so the block can leave management; the commitment then keeps billing to the end of its term with `renewalPlan: NONE`, and through every renewal otherwise.
- **A replacing change is never a resize.** Under `ABANDON` it leaves the old commitment billing and attempts a new purchase; under `DELETE` the apply fails while the term runs.
- **`NONE` ends it; the manifest then re-buys it.** Once Google removes a commitment at the end of its term, remove the block -- a deploy of the unchanged manifest buys a new one.
- **Commitments are pooled.** Every reservation in the same admin project and location draws on them; they belong to no single reservation.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpBigQueryReservation** -- the reservations that draw on the committed slots

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

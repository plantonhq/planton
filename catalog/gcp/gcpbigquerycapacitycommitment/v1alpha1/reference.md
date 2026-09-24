# GcpBigQueryCapacityCommitment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryCapacityCommitmentSpec defines a BigQuery capacity commitment
(`google_bigquery_capacity_commitment`) -- a purchase of slots for a
fixed term at a discount, in an administration project and location.
Commitments are pooled: every reservation (GcpBigQueryReservation) in
the same admin project and location draws on them, so a commitment
belongs to no single reservation.

THIS IS A PURCHASE. Creating one starts a billed term. Google refuses to
delete a commitment before its term ends, so destroying this resource
fails until then -- set deletion_policy to ABANDON when the block should
be able to leave management early, and plan renewals with renewal_plan.

Immutable: project_id, location, capacity_commitment_id, slot_count,
edition, enforce_single_admin_project_per_org (a change would buy a new
commitment). plan (to a longer one) and renewal_plan update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryCapacityCommitment
metadata:
  name: enterprise-annual-500
spec:
  projectId:
    value: bq-admin-project
  location: US
  slotCount: 500
  plan: ANNUAL
  renewalPlan: ANNUAL
  edition: ENTERPRISE
  enforceSingleAdminProjectPerOrg: true
  # Google refuses to delete a commitment before its term ends.
  deletionPolicy: ABANDON
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  |  |  |
| `spec.capacityCommitmentId` | `string` |  |  |  |
| `spec.slotCount` | `int64` |  |  |  |
| `spec.plan` | `string` | yes |  |  |
| `spec.renewalPlan` | `string` |  |  |  |
| `spec.edition` | `string` |  |  |  |
| `spec.enforceSingleAdminProjectPerOrg` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The administration project the commitment is bought in: a literal
project ID or a GcpProject reference. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

Where the slots are: a multi-region (US, EU) or a region. Only
reservations in the same location use them. Empty means US (Google's
default). Immutable.

### spec.capacityCommitmentId

`string`

The commitment's ID -- lowercase letters, digits, and dashes, not
starting or ending with a dash, at most 64 characters. Defaults to
metadata.name. Google does not keep the ID if the commitment is later
split or merged. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z0-9](?:[-a-z0-9]{0,62}[a-z0-9])?$"}}

### spec.slotCount

`int64`

Slots committed. Immutable.

- rule: {"int64":{"gte":"1"}}

### spec.plan

`string` · required

The commitment plan (Google's CommitmentPlan values). With BigQuery
editions the plans on offer are ANNUAL (one year) and THREE_YEAR; the
FLEX, MONTHLY, TRIAL, and *_FLAT_RATE plans belong to the legacy
flat-rate model. A plan can move to a longer term in place, never a
shorter one.

- rule: {"required":true}

### spec.renewalPlan

`string`

The plan the commitment converts to when its term ends (for plans
that renew), e.g. ANNUAL to renew for another year. Changing it
extends the committed period. Empty leaves Google's default.

### spec.edition

`string`

The edition the slots are for: STANDARD, ENTERPRISE, or
ENTERPRISE_PLUS -- it must match the reservations that draw on the
commitment. Empty lets Google choose. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STANDARD","ENTERPRISE","ENTERPRISE_PLUS"]}}

### spec.enforceSingleAdminProjectPerOrg

`bool`

Fail the purchase if another project in the organization already
holds a capacity commitment -- a guard for organizations that keep
every commitment in one admin project. Immutable.

### spec.deletionPolicy

`string`

What happens to the commitment when this resource is destroyed:
  "" / "DELETE" -- deleted, which Google refuses before the term ends
                   (the destroy then fails)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the commitment leaves management and runs out its
                   term (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryCapacityCommitment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/capacityCommitments/{id}. |
| `status.outputs.state` | `string` | The commitment's state (PENDING, ACTIVE, or FAILED). |
| `status.outputs.commitment_start_time` | `string` | When the current term started (ACTIVE commitments only). |
| `status.outputs.commitment_end_time` | `string` | When the current term ends (ACTIVE commitments only) -- the earliest a delete succeeds. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)

# GcpBigQueryCapacityCommitment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpBigQueryCapacityCommitmentSpec defines a BigQuery capacity commitment
(`google_bigquery_capacity_commitment`) -- a purchase of slots for a
fixed term at a discount, in an administration project and location.
Commitments are pooled: every reservation (GcpBigQueryReservation) in
the same admin project and location draws on them, so a commitment
belongs to no single reservation.

THIS IS A PURCHASE. Creating one starts a billed term. A commitment
cannot be deleted before its term ends: Google refuses the delete, so
destroying this resource with DELETE fails until then. Set
deletion_policy to ABANDON when the block should be able to leave
management early, and decide with renewal_plan whether the commitment
ends at its term (NONE) or renews.

Immutable: project_id, location, capacity_commitment_id, slot_count,
edition, enforce_single_admin_project_per_org (a change would buy a new
commitment). plan (to a longer one) and renewal_plan update in place.

A replacing change is never a resize. Under ABANDON it leaves this
commitment billing to the end of its term, unmanaged, and attempts a new
purchase; under DELETE the apply fails while the term runs. More slots
is a second commitment; fewer is Google's split into two commitments,
then a delete of one after its end time.

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
reservations in the same admin project and location use them; a
commitment in another project or location does nothing for them. Empty
means US (Google's default). Write US and EU in upper case and a region
in lower case: the provider never reads the location back, so another
spelling is another location to it, and replaces (re-buys) the
commitment. Immutable.

- rule: location must be US or EU in upper case, or a region in lower case such as us-central1.

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
shorter one: Google refuses a shorter plan. After a renewal Google
reports the renewal plan as the commitment's plan, so keep plan equal
to renewal_plan, or update plan after each renewal -- otherwise the
next deploy re-commits to a longer plan or fails on a shorter one.

- rule: {"required":true,"string":{"in":["FLEX","FLEX_FLAT_RATE","TRIAL","MONTHLY","MONTHLY_FLAT_RATE","ANNUAL","ANNUAL_FLAT_RATE","THREE_YEAR"]}}

### spec.renewalPlan

`string`

What the commitment becomes when its term ends:
  NONE       -- it ends: Google removes it at the end of its term (it
                needs edition set). Once Google removes it, remove this
                block, or destroy it with ABANDON before the term ends:
                a deploy of the unchanged manifest after the term buys
                a new commitment.
  ANNUAL     -- it renews for another year, and keeps renewing
  THREE_YEAR -- it renews for three years
(and the legacy flat-rate plans). Changing it extends the committed
period. Empty leaves Google's default, and ANNUAL commitments renew by
default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FLEX","FLEX_FLAT_RATE","TRIAL","MONTHLY","MONTHLY_FLAT_RATE","ANNUAL","ANNUAL_FLAT_RATE","THREE_YEAR","NONE"]}}

### spec.edition

`string`

The edition the slots are for: STANDARD, ENTERPRISE, or
ENTERPRISE_PLUS -- it must match the reservations that draw on the
commitment. Empty lets Google choose. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STANDARD","ENTERPRISE","ENTERPRISE_PLUS"]}}

### spec.enforceSingleAdminProjectPerOrg

`bool`

Google's guard that fails the purchase if another project in the
organization already holds a capacity commitment. The pinned provider
records this value but never sends it to Google (Google takes it as a
query parameter of the create call, which the provider leaves out), so
it guards nothing; changing it still replaces -- re-buys -- the
commitment. Leave it unset. Immutable.

### spec.deletionPolicy

`string`

What happens to the commitment when this resource is destroyed:
  "" / "DELETE" -- deleted, which Google refuses before the term ends
                   (the destroy then fails)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the commitment leaves management and keeps billing
                   in GCP: to the end of its term when renewal_plan is
                   NONE, and through every renewal otherwise
The policy also governs a replacing change (see the message comment).

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.renewal_none_needs_edition`: renewal_plan NONE needs an edition: Google accepts NONE only for an edition commitment.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryCapacityCommitment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/capacityCommitments/{id}. |
| `status.outputs.state` | `string` | The commitment's state (PENDING, ACTIVE, or FAILED). |
| `status.outputs.commitment_start_time` | `string` | When the original commitment started (ACTIVE commitments only); renewals don't move it. |
| `status.outputs.commitment_end_time` | `string` | When the current term ends (ACTIVE commitments only). A delete before it fails. A renewal moves it forward, so a commitment that renews never reaches it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)

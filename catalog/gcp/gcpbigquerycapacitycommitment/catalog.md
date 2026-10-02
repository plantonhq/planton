# GCP BigQuery Capacity Commitment

Buys BigQuery capacity at a discount for a one- or three-year term, pooled across every reservation in an administration project -- the finance team's lever on the BigQuery bill, declared as its own block because it outlives any single reservation.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `bigqueryreservation.googleapis.com` on the admin project (never disabled on destroy)
- **Capacity commitment** -- a `bigquery_capacity_commitment`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with BigQuery resource admin permissions (`roles/bigquery.resourceAdmin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpBigQueryReservation`** -- reservations in the same admin project and location draw on the committed slots.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Capacity Commitment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Annual Enterprise Commitment** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryCapacityCommitment
metadata:
  name: enterprise-annual-500
  org: acme-corp
  env: prod
spec:
  location: US
  slotCount: 500
  plan: ANNUAL
  renewalPlan: ANNUAL
  edition: ENTERPRISE
  deletionPolicy: ABANDON
```

```shell
planton apply -f bigquery-capacity-commitment.yaml
```

This commits to 500 Enterprise slots in `US` for a year, renewing annually. A Stack Job tracks the provisioning in real time.

### InfraChart

Declare commitments in the same admin project and location as the `GcpBigQueryReservation` blocks that use them; keep them in a stack that is rarely destroyed.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Slots and term** -- `slotCount` and `plan` are what you buy; `ANNUAL` and `THREE_YEAR` are the editions-era plans.

**Renewal** -- `renewalPlan` decides what happens at the end of the term; changing it extends the commitment.

**Destroy** -- `ABANDON` is the honest setting -- Google will not delete a commitment before its term ends.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The commitment's full name | Audit |
| `commitment_end_time` | When the term ends | Renewal planning |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Annual commitment** -- 100 Enterprise slots for a year, renewing annually. Start from the **Annual Enterprise Commitment** preset.

**Long-term commitment** -- Enterprise Plus slots for three years, guarded to one admin project per organization. Start from the **Three-Year Commitment** preset.

## Works With

- [**GCP BigQuery Reservation**](/cloud-catalog/gcp-bigquery-reservation) -- the capacity the commitment discounts

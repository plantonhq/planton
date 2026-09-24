# GcpBigQueryCapacityCommitment Guide

The judgment this guide protects: a commitment is a contract, not infrastructure. Applying one spends money for a year or more and cannot be undone, so treat the manifest like a purchase order -- reviewed by whoever owns the BigQuery bill, kept in a stack that is rarely destroyed, and set to leave management gracefully.

## What you buy

`slotCount` slots of an `edition` for a `plan`'s term, in an admin project and location. With BigQuery editions, Google offers `ANNUAL` and `THREE_YEAR` plans; the `FLEX`, `MONTHLY`, `TRIAL`, and `*_FLAT_RATE` plans belong to the legacy flat-rate model. A plan can move to a longer term in place, never a shorter one. `renewalPlan` decides what the commitment becomes when the term ends; changing it extends the committed period.

## How it is used

Commitments are pooled: every `GcpBigQueryReservation` in the same admin project and location draws on them, and the committed slots lower the price of those reservations' baseline. That is why a commitment is its own block -- it belongs to no single reservation and outlives them. `enforceSingleAdminProjectPerOrg` refuses the purchase if another project in the organization already holds a commitment, for organizations that keep all capacity in one place.

## Destroy

Google refuses to delete a commitment before its term ends, so with the default `DELETE` a destroy fails until then. `ABANDON` lets the block leave management while the commitment runs out its term in Google Cloud -- the setting every preset uses. The `commitment_end_time` output is the earliest a delete can succeed.

# GcpBigQueryCapacityCommitment Guide

The judgment this guide protects: a commitment is a contract, not infrastructure. Applying one spends money for a year or more and cannot be undone, so treat the manifest like a purchase order -- reviewed by whoever owns the BigQuery bill, kept in a stack that is rarely destroyed, and set to leave management gracefully.

## What you buy

`slotCount` slots of an `edition` for a `plan`'s term, in an admin project and location. With BigQuery editions, Google offers `ANNUAL` and `THREE_YEAR` plans; the `FLEX`, `MONTHLY`, `TRIAL`, and `*_FLAT_RATE` plans belong to the legacy flat-rate model. A plan can move to a longer term in place, never a shorter one. `renewalPlan` decides what the commitment becomes when the term ends: `NONE` ends it (it needs `edition` set), `ANNUAL` or `THREE_YEAR` renew it, and `ANNUAL` commitments renew by default. Changing it extends the committed period. After a renewal Google reports the renewal plan as the commitment's plan, so keep `plan` equal to `renewalPlan` -- otherwise the next deploy re-commits to a longer plan or fails on a shorter one.

A replacing change (slots, edition, location, id, project) is never a resize. More slots is a second commitment; fewer is Google's split into two commitments, then a delete of one after its end time.

## How it is used

Commitments are pooled: every `GcpBigQueryReservation` in the same admin project and location draws on them, and the committed slots lower the price of those reservations' baseline. That is why a commitment is its own block -- it belongs to no single reservation and outlives them. Google's `enforceSingleAdminProjectPerOrg` guard is never sent by the pinned provider, so it guards nothing here; leave it unset, since changing it replaces the commitment.

## Destroy

A commitment cannot be deleted before its term ends: Google refuses, so with the default `DELETE` a destroy fails until then. `ABANDON` lets the block leave management while the commitment keeps billing in Google Cloud -- to the end of its term with `renewalPlan: NONE`, and through every renewal otherwise. Every preset pairs `ABANDON` with `NONE`. The same policy governs a replacing change: under `ABANDON` it leaves the old commitment billing and attempts a new purchase; under `DELETE` the apply fails while the term runs.

The clean exit: set `renewalPlan: NONE` (with `edition` set) and apply; then either destroy with `ABANDON`, or let Google remove the commitment at the end of its term and remove the block. A deploy of the unchanged manifest after Google removes it buys a new commitment. `commitment_end_time` is when the current term ends; a renewal moves it forward.

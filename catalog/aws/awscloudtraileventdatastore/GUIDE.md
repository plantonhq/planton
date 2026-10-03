# AwsCloudTrailEventDataStore — Kind Guide

Authored operational judgment for the CloudTrail Lake component: the
design decisions behind the spec's shape, and what to know before
operating event data stores in production.

## Service availability: CloudTrail Lake is closed to new customers

AWS no longer accepts new CloudTrail Lake customers: on an account that has never created an event data store, `CreateEventDataStore` fails with `InvalidParameterException: CloudTrail Lake is no longer accepting new customers. Existing customers can continue to use their event data stores.` (live-verified 2026-08-25, identical on both engines). This component therefore deploys ONLY on accounts grandfathered into the service — one that already holds (or previously created) an event data store. There is no known exception process. If your account is not grandfathered, use the AwsCloudTrail kind's trail delivery to S3/CloudWatch instead; nothing in this kind's spec can route around the account-level wall.

## Design decisions

- **Its own kind, not a trail arm.** An event data store deploys with
  zero trails and owns its own billing, retention, and
  termination-protection lifecycle — no trail edge exists in the
  provider's schema.
- **Provider-default toggles are tri-state.** `multi_region_enabled`
  and `termination_protection_enabled` both default TRUE at AWS;
  unset means "take the AWS default", and the modules render only an
  explicit choice so they never fight it.
- **`suspend` is write-only at AWS.** The API never reports it back,
  so it is asserted on every apply and invisible to imports — a
  freshly imported store shows it unset, and the first apply
  re-asserts it as a server-side no-op.
- **Selector rules are taught, not invented.** AWS requires every
  advanced selector to carry an `eventCategory` condition — a
  server-side rule the provider does not pre-check; the spec comments
  carry it and the live lane holds the contract.
- **The KMS key is fixed at creation.** Changing `kms_key_id`
  replaces the store (and its ingested history) — pick the key before
  first ingestion, and prefer a multi-region key for a multi-region
  store. Losing the key makes the store unreadable.

## Operating event data stores in production

- **The teardown is two steps by AWS design.** With termination
  protection on (the default), a destroy FAILS: apply
  `termination_protection_enabled: false` first, then destroy. The
  delete is soft — `PENDING_DELETION` for 7 days, the name reserved,
  the store still describable (and restorable via the console) until
  the purge.
- **Ingestion is the bill.** Lake charges per GB ingested (extendable
  pricing) or less on ingest but per GB retained (fixed pricing, 7-year
  cap). An unscoped store ingests EVERY management event; scope
  selectors to what investigations actually need.
- **Pricing mode is a one-way door in practice.** AWS allows changing
  fixed → extendable only; pick deliberately at creation.
- **Pause instead of delete.** `suspend: true` stops the meter on new
  ingestion while keeping history queryable — the right lever for
  stores kept for occasional investigations.
- **Organization stores run from the management account** (or the
  delegated CloudTrail administrator) with all-features Organizations
  enabled.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

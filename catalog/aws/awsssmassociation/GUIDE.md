# AwsSsmAssociation — Kind Guide

Authored operational judgment for the association component: the
design decisions behind the spec's shape, and what to know before
operating associations in production.

## Design decisions

- **This kind exists because associations bind ANY document.** The
  provider's document reference is a free string with no validator and
  no structural edge to a user document — AWS-managed documents are
  the documented norm. As a document satellite, "schedule
  AWS-RunPatchBaseline nightly" would have been unrepresentable; as
  its own kind, the document is a value-or-reference field
  (`documentName`) that carries both cases honestly.
- **No registry prerequisite** for the same reason: nothing is
  schema-required. Chart wiring to a customer AwsSsmDocument rides the
  reference's `valueFrom` arm.
- **Identity is the AWS-generated UUID**, not the name —
  `association_name` is console display metadata, and the import map
  derives `association_id` from the outputs.

## Operating associations in production

- **Changing the document replaces the association**; every other
  change creates a new association VERSION in place — AWS versions
  associations whole, so the provider sends the full argument set on
  update.
- **AWS materializes the document's declared parameter defaults** into
  the stored parameters map. A freshly imported association shows the
  merged result — the first plan reconciles it (declared
  write-normalized in the import map); post-apply idempotency holds.
- **`wait_for_success_timeout_seconds` is a create-time gate only** —
  it fails the DEPLOY unless the first run succeeds within the window.
  Never set it when no matching targets exist yet: the wait times out
  by construction. It is never read back (config-only in the import
  map).
- **`apply_only_at_cron_interval` prevents the immediate first run**
  on create/update — without it, State Manager applies the document
  right away and then on schedule.
- **`max_errors: "0"` is the strictest honest setting**: one failure
  stops further scheduling for that interval. With tag targets on
  large fleets, prefer percentages for both rate controls.
- **`sync_compliance: MANUAL` hands compliance to an external
  process** (PutComplianceItems) — the association stops writing its
  own compliance records; don't set it casually.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

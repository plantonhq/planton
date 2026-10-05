# AwsOrganizationAccount — Kind Guide

Authored operational judgment for the member-account kind: the
design decisions behind the spec's shape, and what to know before
operating member accounts in production.

## Design decisions

- **The account-settings satellites fold onto the created account.**
  Alternate contacts, the primary contact, and region enablement all
  carry an optional target-account edge in the provider — here that
  edge is ALWAYS the folded pivot's ID, which is what makes them
  satellites. (The same AWS APIs can manage the CALLER's own account
  when the target is omitted — a settings-singleton use case this kind
  deliberately does not model; it is revisitable on demand as its own
  future kind.)
- **Alternate contacts are three typed fields, not a list.** AWS keeps
  at most one contact per category (billing/operations/security) — the
  exactly-one-per-type contract is structural in the spec, so a
  duplicate-category mistake cannot be written.
- **`roleName` is write-once and the engines ignore later changes.**
  AWS exposes NO API to read the bootstrap role back; without the
  ignore, importing an existing account would plan a destructive
  replacement to "set" a value AWS can never echo. A genuine role
  change is an account rebuild — do it deliberately, not through this
  field.
- **`closeOnDeletion` is the modeled delete contract**, not an
  engine flag exclusion: which destroy AWS performs (remove vs close)
  is real account-lifecycle configuration an operator must choose.

## Operating member accounts in production

- **Treat member accounts as long-lived.** Removal requires the
  account to carry standalone billing information and leaves it ALIVE
  outside the org; closure suspends it for ~90 days
  (PENDING_CLOSURE), holds its email and account ID through that
  window, and is quota-limited (~10% of accounts per rolling 30 days).
  Plan account topology like schema design, not like instances.
- **Root emails are forever-unique across AWS** — closed accounts hold
  their email through the closure window. Use plus-addressed or
  per-account aliases (`aws+workloads-prod@example.com`) so a
  recreated account never fights its predecessor.
- **The contact arms are eventually consistent** — the provider polls
  until writes are visible (the module inherits that patience); a
  slow-but-green apply here is normal.
- **Clearing an optional primary-contact leaf does NOT clear it at
  AWS** — the Put API has no unset semantics; the last value stays on
  file. Overwrite deliberately, never by omission.
- **Region enablement is slow and sticky**: ~60 minutes each way, and
  removing an entry keeps the region's last state (delete is a no-op
  by AWS design). Disable explicitly with `enabled: false` when a
  region must actually close.
- **GovCloud companions are import-only afterwards** — the
  `createGovcloud` flag creates the pair once; the GovCloud account is
  managed by importing it in the GovCloud partition, never through
  this resource.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

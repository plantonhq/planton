# AwsSsmDocument — Kind Guide

Authored operational judgment for the document component: the design
decisions behind the spec's shape, and what to know before operating
documents in production.

## Design decisions

- **Associations are NOT this document's satellite.** The provider's
  association resource references documents by a free string with no
  validator — AWS-managed documents (`AWS-RunShellScript`,
  `AmazonCloudWatch-ManageAgent`) are the documented norm, with no
  user document anywhere. Folding associations here would have made
  the most common State Manager use unrepresentable; they are their
  own AwsSsmAssociation kind.
- **Sharing is a typed account-ID list**, not the provider's flat
  `{type, account_ids}` map — `Share` is the only legal type (the
  provider validates it), so the type key carries no configuration.
  `All` shares publicly and is pattern-legal.
- **`metadata.name` is the document name** — document names allow
  hyphens (3–128 of letters, digits, `_.-`), so no explicit name
  field is needed.

## Operating documents in production

- **Every content update creates a new document version** and the
  module promotes it to the default. Schema-1.x (legacy command)
  documents only accept updates when the content itself changed — an
  AWS rule; touching only tags is fine, touching only permissions is
  fine, but a no-op content rewrite is rejected.
- **`versionName` labels are immutable forever per document** — AWS
  rejects reusing a label on different content. Treat them like git
  tags: bump per release.
- **DeleteDocument is asynchronous.** A deleted document reads as
  `Deleting` briefly, and re-creating the same name during that window
  fails. Rapid destroy/recreate cycles of the same name need a wait or
  a fresh name.
- **Attachments are write-only at AWS** — no API returns attachment
  metadata, so a freshly imported document shows none and the first
  apply re-asserts them (declared config-only in the import map).
- **Un-sharing on delete is automatic**: the provider removes all
  permissions before deleting, in account-ID batches of 20.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

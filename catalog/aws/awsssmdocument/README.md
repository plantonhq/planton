<p align="center">
  <img src="logo.svg" alt="AWS SSM Document" width="80"/>
</p>

# AWS SSM Document

Manage a [customer-owned AWS Systems Manager document](https://docs.aws.amazon.com/systems-manager/latest/userguide/documents.html)
— a reusable definition of the actions Systems Manager performs: run a
command on managed nodes, drive an automation runbook, configure a
session, and the other document types AWS accepts on the same API.

## What Gets Managed

- **The document** (`metadata.name` is the document name): its
  **content** (JSON or YAML per `documentFormat`), **documentType**
  (Command, Automation, Session, Policy, Package, and the specialty
  types), a **targetType** restriction, an immutable per-version
  **versionName** label, **attachments** for artifact-carrying types,
  and **sharing** (account IDs or `All` for public).
- Updating the content creates a **new document version** and promotes
  it to the default.

State Manager associations (binding a document to targets on a
schedule) are deliberately NOT part of this kind — an association
binds ANY document, AWS-managed included; see
[AwsSsmAssociation](../awsssmassociation).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

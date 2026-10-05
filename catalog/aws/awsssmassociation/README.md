<p align="center">
  <img src="logo.svg" alt="AWS SSM Association" width="80"/>
</p>

# AWS SSM Association

Manage an [AWS Systems Manager State Manager association](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-state.html)
— the binding of an SSM document to targets on a schedule: "run this
document on these machines, this often".

## What Gets Managed

- **The association** (AWS identifies it by a generated UUID): the
  **document** it runs (an AWS-managed name like `AWS-RunPatchBaseline`
  as a literal, or a customer [AwsSsmDocument](../awsssmdocument) by
  reference), pinned to a **documentVersion**; the document's
  **parameters**; up to five **targets** (instance IDs, tags, resource
  groups); the **schedule** (cron/rate, optionally only-at-interval);
  **compliance** severity and sync mode; **rate controls**
  (maxConcurrency/maxErrors); Change Calendar gating; and S3
  **output delivery**.

The document reference is deliberately a value-or-reference: an
association binds ANY document — AWS-managed documents are first-class
— which is why this is its own kind rather than a document
satellite.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

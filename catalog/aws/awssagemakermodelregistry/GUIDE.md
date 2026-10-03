# AwsSagemakerModelRegistry — Kind Guide

Authored operational judgment for the model registry component: the
design decisions behind the spec's shape, and what to know before
running a model registry in production.

## Design decisions

- **The group's AWS name derives from `metadata.name`.** SageMaker's
  charset accepts the manifest name as-is — no separate name field to
  drift from the resource's identity.
- **The resource policy is folded into the spec.** AWS models the
  policy as a separate resource (an idempotent
  `PutModelPackageGroupPolicy` upsert), but it has no life of its own —
  one group, at most one policy. Folding it as structured JSON keeps
  the manifest one document; presence in the spec is presence in AWS,
  and removing it deletes the policy.
- **No version fields, by design.** Model package versions register
  into the group from training pipelines and SDK calls — an imperative,
  high-frequency act that has no place in declarative infrastructure.
  The spec describes the shelf, not the books.
- **The description is a create-time contract.** Upstream marks it
  ForceNew — an edit replaces the group. The spec caps it at 1024
  characters and the docs say it plainly: write it once, well.

## Running a model registry in production

- **Never edit the description of a live group.** A description edit
  replaces the group (provider-enforced ForceNew) — treat the
  description as part of the group's identity and write it before the
  first model version registers.
- **Iterate on sharing through the policy.** The resource policy is the
  one in-place-updateable arm — add and remove cross-account principals
  freely without touching the group.
- **Grant the minimum verbs.** Read-side consumers need
  `sagemaker:DescribeModelPackage` / `sagemaker:ListModelPackages`;
  only accounts that register models into your group need
  `sagemaker:CreateModelPackage`.
- **Groups are free — organize generously.** One group per model family
  keeps approval workflows and lineage legible; creating another costs
  nothing.
- **Tags are the safe mutable surface.** Ownership, cost-center, and
  lifecycle labels can evolve without replacing anything.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

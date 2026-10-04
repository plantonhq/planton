# AwsSagemakerPipeline — Kind Guide

Authored operational judgment for the pipeline kind: the design
decisions behind the spec's shape, and what to know before running ML
pipelines in production.

## Design decisions

- **The pipeline's AWS name derives from `metadata.name`.** SageMaker's
  charset accepts the manifest name as-is — no separate name field to
  drift from the resource's identity. The display name is separate
  because the provider REQUIRES one: omitted, the modules reuse the
  pipeline name.
- **The definition comes from exactly one place.** A CEL rule enforces
  AWS's own CreatePipeline contract — `definition` XOR
  `definition_s3_location` — at manifest time instead of at apply.
- **The inline definition is structured data, not a string.** The spec
  carries the pipeline-definition JSON as a `Struct`, so manifests stay
  diffable YAML rather than an opaque escaped-JSON blob.
- **The S3 arm's blind spot is taught on the spec field.** AWS's
  describe API returns only the RESOLVED definition, never the S3
  location — the location is config-only on import and drift on the S3
  object is invisible to refresh. The field's own comment says so, so
  nobody learns it from an incident.

## Running ML pipelines in production

- **Generate definitions; don't hand-write them.** The SageMaker Python
  SDK's `pipeline.definition()` produces the definition JSON — commit
  its output into the manifest (or upload it to S3) rather than
  authoring the schema by hand.
- **A green apply is the validity claim.** AWS validates the step graph
  server-side at create — a malformed DAG fails the apply, not the
  first execution.
- **Prefer the inline arm when definitions fit.** Inline, the
  definition lives in the manifest and drifts visibly. On the S3 arm,
  a changed object is invisible to refresh — if you must use it, pin
  `version_id` and treat every definition change as a manifest change.
- **Pipelines are free — executions bill.** Keep as many pipeline
  shells as your workflows need; cost only accrues when executions run
  their steps.
- **Everything except the name updates in place.** Iterate on the
  definition, display name, role, and parallelism cap without
  replacement; `parallelism_max_steps` is a default across executions.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

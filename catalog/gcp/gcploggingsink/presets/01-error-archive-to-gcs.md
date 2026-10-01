# Error Archive to GCS

The cheapest compliance archive: every ERROR-and-above entry in the
project lands as hourly JSON batches in a Cloud Storage bucket.

## What it configures

- A project-scoped sink (no `scope` block — the ambient project).
- A `gcsBucket` destination — the module renders the
  `storage.googleapis.com/...` URI from the bucket name.
- `severity>=ERROR` — errors are evidence; INFO is noise at archive
  prices.

## The deploy's second half

Grant the sink's `writer_identity` output
`roles/storage.objectCreator` on the bucket — through a
`GcpGcsBucketIamMember` in the same chart whose `member` references the
sink's `status.outputs.writer_identity` (never the bucket's own
`iamMembers`: that would make the bucket depend on the sink that depends
on it). Without the grant the sink reports
success and exports NOTHING.

## Adjust before deploying

- **gcsBucket** — reference a GcpGcsBucket resource via valueFrom in
  charts so the grant flow stays declarative.
- Consider bucket lifecycle rules (the bucket kind's surface) to age
  archives to Coldline.

## When to choose something else

Need to QUERY the logs? The **Audit Logs to BigQuery** preset. Feeding
an external pipeline? The **Log Stream to Pub/Sub** preset.

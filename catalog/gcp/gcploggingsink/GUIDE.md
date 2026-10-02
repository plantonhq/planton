# GcpLoggingSink Guide

The judgment this guide protects: a sink without its destination grant is
a silent no-op. GCP creates the sink, reports success, and exports
nothing — no error surfaces anywhere. The writer-identity grant is the
deploy's second half, not an optional extra.

## The grant flow is the whole game

Every sink gets a `writer_identity` service account, and that identity
must hold write access ON THE DESTINATION: `roles/storage.objectCreator`
(bucket), `roles/bigquery.dataEditor` (dataset),
`roles/pubsub.publisher` (topic). Grant it with a standalone grant block
that depends on both the sink and the destination: a
`GcpGcsBucketIamMember` for a bucket, a `GcpPubSubTopicIamMember` for a
topic, each with `member` referencing the sink's
`status.outputs.writer_identity` (already `serviceAccount:{email}`).
The grant deploys after both, and a recreated sink (new identity)
re-grants automatically. Never grant it through the destination kind's
own IAM fields: the sink references its destination, so a destination that references the sink's identity back would be a dependency cycle. A BigQuery dataset's access list is authoritative
on `GcpBigQueryDataset.access`, so grant `roles/bigquery.dataEditor`
outside the catalog, or as a literal `access` entry once the identity is
known.

## Renaming is recreation is a new identity

`sinkName` is ForceNew, and the replacement sink mints a NEW writer
identity — the old grant now authorizes a dead account and the new
identity has nothing. Chart-wired grants heal this; hand-managed grants
silently break the export.

## The empty filter is a decision, not a default posture

An empty `filter` exports EVERYTHING in scope. That is deliberate for
audit archives and ruinously expensive for BigQuery destinations on
chatty projects. Start narrow (`severity>=ERROR`, a resource.type
scope) and widen deliberately; use `exclusions` with `sample()` to keep
high-volume noise at a fraction.

## intercept_children rewires OTHER people's logging

An org/folder sink with `interceptChildren: true` takes matching logs
away from the children's own sinks — child projects stop seeing those
entries in their exports. It exists for compliance capture; using it
without telling the child teams is how platform teams break someone
else's pipeline from two scopes above.

## BigQuery destinations: partitioned, always

`usePartitionedTables: true` is strictly better than the legacy
date-sharded table names (partition pruning, expiration policies, sane
SQL) and is required to remain on the unique writer. The spec leaves the
provider default (false) untouched for fidelity; the presets set it.

## Teardown discipline

`PREVENT` is right for compliance-mandated exports — deleting the sink
stops the export instantly while the destination keeps only history.
`ABANDON` keeps exporting unmanaged. Already-exported data always
survives sink deletion.

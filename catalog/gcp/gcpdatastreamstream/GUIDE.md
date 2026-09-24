# GcpDatastreamStream Guide

The judgment this guide protects: a stream is a standing bill for every GiB it moves, so it replicates exactly the objects you need, in the shape your analysts query, and starts only when you say so.

## Before it runs

Each source needs its own preparation, outside the catalog: MySQL binary logging in ROW format; PostgreSQL logical decoding, a publication, and a replication slot (one slot per stream -- an idle slot makes the server retain WAL, so drop it with the stream); Oracle archive and supplemental logging; SQL Server CDC or transaction-log access; a Spanner change stream. Create the stream `NOT_STARTED` (the default), check it in the console, then set `desiredState: RUNNING`.

## Backfill

`backfillAll` copies what exists before streaming changes; list big or irrelevant tables under its `*ExcludedObjects`. `backfillNone: true` replicates only changes from the start -- the choice when history already lives in the destination. The backfill bills at its own per-GiB rate.

## Landing in BigQuery

- **`writeMode: MERGE`** (Google's default) -- tables mirror the source's current state, merged by primary key; `dataFreshness` trades staleness for BigQuery compute.
- **`writeMode: APPEND_ONLY`** -- every change is a row with its change type: history, audits, slowly changing dimensions. It is immutable, so choose deliberately.
- **Datasets** -- `singleTargetDataset` puts every table in one dataset (a `GcpBigQueryDataset` reference, trimmed to Google's `projects/{p}/datasets/{d}` form); `sourceHierarchyDatasets` creates one dataset per source schema, prefixed and optionally CMEK-encrypted.
- **`ruleSets`** -- partition or cluster a source object's table when Datastream creates it. Rules apply at creation, so declare them before the stream starts.
- **`blmtConfig`** -- write Apache Iceberg tables to your bucket through a BigQuery cloud-resource connection, whose service account needs storage access.

## Landing in Cloud Storage

Avro keeps types exactly; JSON (optionally gzip, optionally with an Avro schema file) is easier for ad hoc tools. Rotation by time (15-60 seconds) or size decides file count and freshness.

## Grants

Datastream acts as `service-{project-number}@gcp-sa-datastream.iam.gserviceaccount.com`. Grant it BigQuery Data Editor where it writes, object access on Cloud Storage destinations, and `roles/cloudkms.cryptoKeyEncrypterDecrypter` on any key -- each on the resource being accessed.

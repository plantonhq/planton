# MySQL to Cloud Storage (Avro)

## Use Case

Land every change from a MySQL database as Avro files in a Cloud Storage data lake, for Dataflow, Spark, or BigLake to process downstream.

## When to Use

- Feeding a lake or a custom processing pipeline instead of BigQuery directly
- Keeping a raw, replayable change log

## What This Creates

- A stream from the `shop` database, read by GTID, into Avro files under `/shop` in the `raw-lake` profile's bucket, rotated every 60 seconds or 100 MB, created `NOT_STARTED`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `sourceConfig.mysqlSourceConfig.cdcMethod` | `GTID` | `GTID` survives a failover to a replica; use `BINARY_LOG_POSITION` when GTIDs are off. |
| `destinationConfig.gcsDestinationConfig.avroFileFormat` | `true` | Use `jsonFileFormat` (optionally gzip) for JSON files instead. |
| `destinationConfig.gcsDestinationConfig.fileRotationInterval` | `60s` | Google allows 15 to 60 seconds; shorter means fresher, smaller files. |

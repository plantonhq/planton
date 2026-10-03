# AwsDynamodb

An Amazon DynamoDB table: key schema and indexes, capacity in any of AWS's three shapes (on-demand, provisioned, pre-warmed), streams, Global Tables v2 multi-region replication, encryption, recovery, and the table-scoped governance surface (resource policy, Kinesis change-data destination, CloudWatch contributor insights).

The table name comes from `metadata.name` (create-time immutable in AWS). A table is a true leaf -- nothing has to exist before it -- and everything it composes with attaches by reference: the KMS key that encrypts it (`AwsKmsKey`), the Kinesis stream that receives its change data (`AwsKinesisStream`), and the S3 bucket an import seeds it from (`AwsS3Bucket`).

## Spec highlights

- **Capacity, three-shaped**: `PAY_PER_REQUEST` (the recommended default -- zero capacity planning, optional `onDemandThroughput` spend ceiling), `PROVISIONED` (reserved units on the table and each GSI, for reserved-capacity pricing), and `warmThroughput` (pre-warmed instant capacity for launch events; increase-only).
- **Auto scaling on provisioned tables**: `autoscaling` folds Application Auto Scaling -- target-tracking on read/write utilization (bounds + a 20-90% target) plus named, timezone-aware `scheduledAdjustments`. Live capacity is scaler-owned on every provisioned table (without the block the modules register pinned min = max targets from `provisionedThroughput`), which is what makes adding or removing autoscaling an in-place update.
- **Indexes**: global secondary indexes with per-index capacity, warm throughput, and multi-attribute keys (1-4 HASH + 0-4 RANGE elements); local secondary indexes stated honestly as create-time-only.
- **Global Tables v2**: folded `replicas` (per-region KMS key, PITR, deletion protection, tag propagation) with `consistencyMode` up to Multi-Region Strong Consistency -- exactly two STRONG replicas, or one STRONG replica plus a `globalTableWitness` region. Validation pins the topology AWS accepts before anything deploys.
- **Create sources**: restore from point-in-time (same-account by name, cross-account/region by ARN), restore from a backup ARN, or seed from S3 (`importTable` -- CSV, DynamoDB JSON, or Ion; billed as a one-time import).
- **Folded table-scoped satellites**: a resource-based IAM `resourcePolicy` (the policy document as native YAML, plus the `confirmRemoveSelfResourceAccess` guard for deliberate lockdown policies), the `kinesisStreamingDestination` (one per table, by AWS's own rule), and `contributorInsights` on the table and opted-in GSIs -- each materializes as its own provider resource in both engines, so edits are in-place.
- **Recovery and safety**: point-in-time recovery with a tunable 1-35 day window, deletion protection, table class, and TTL.

## Outputs

`table_name`, `table_arn`, `table_id`, `stream_arn`, `stream_label` -- the name and ARN are the join keys IAM policies and application configuration consume; the stream ARN is what Lambda event-source mappings attach to.

## How it works

Both the Terraform/OpenTofu and Pulumi modules implement the same contract: one `aws_dynamodb_table` keyed by `metadata.name`, plus one provider resource per configured satellite (resource policy, Kinesis destination, contributor insights per table/index, and -- on provisioned tables -- the Application Auto Scaling targets, policies, and scheduled actions). Spec values are the exact AWS API strings, so manifests read like AWS documentation and values pass through the engines untranslated.

On provisioned tables both engines hand live read/write capacity to Application Auto Scaling and stop reconciling it on the table resource: with `autoscaling` configured the scaler tracks your target inside your bounds, and without it the modules register pinned min = max targets from `provisionedThroughput` -- AWS moves out-of-range capacity into the bounds on every target update, so a capacity edit in the manifest still lands.

## References

- [DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html)
- [Read/write capacity modes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadWriteCapacityMode.html)
- [Global Tables](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GlobalTables.html)
- [Warm throughput](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/warm-throughput.html)
- [Importing from S3](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataImport.HowItWorks.html)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

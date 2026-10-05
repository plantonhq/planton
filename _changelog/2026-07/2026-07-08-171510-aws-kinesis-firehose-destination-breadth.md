# AWS Kinesis Firehose: Full Destination Breadth, MSK Source, Secrets Manager Credentials, and a Typed Processor Pipeline

**Date**: July 8, 2026
**Type**: Feature, Breaking Change
**Components**: API Definitions, Provider Framework, Terraform Modules, Pulumi Modules, E2E Harness

## Summary

`AwsKinesisFirehose` grows from four destinations to the complete provider surface: **Splunk**, **Snowflake** (Snowpipe Streaming), **Apache Iceberg** (Glue-cataloged tables), and **OpenSearch Serverless** join extended S3, OpenSearch, HTTP endpoint, and Redshift. The **Amazon MSK source** turns Kafka topics into delivery pipelines with zero consumer code, **Secrets Manager credentials** keep destination secrets out of manifests and IaC state on all four credentialed destinations, and the Lambda-only `processing` block is restructured into an **ordered pipeline of six typed processor arms** — completing dynamic partitioning, which the old spec enabled with no way to define partition keys. All four live dual-engine lanes are green with a zero-orphan sweep.

## What Was Built

### Four new destination arms (oneof, exactly-one, ForceNew)

- **`splunk`** — HEC endpoint/type/token, indexer-acknowledgment timeout (180–600s), and the tightest buffering caps of any destination (0–60s / 1–5 MiB, CEL-enforced). Splunk has no destination-level role — HEC authorization is the token, and the S3 backup configuration carries its own role.
- **`snowflake`** — account URL + database/schema/table, key-pair authentication (user + RSA private key, both `(sensitive)`), a dedicated Snowflake ingestion role, PrivateLink VPCE, and the three data-loading modes (JSON column mapping XOR VARIANT content/metadata columns, requiredness CEL-coupled). Defaults to Snowpipe Streaming's near-real-time buffering (0s / 1 MiB).
- **`iceberg`** — the Glue catalog ARN (create-time immutable), per-table routing with `unique_keys` upserts (change-data-capture straight into the lakehouse), `append_only` mode, per-table error prefixes.
- **`opensearch_serverless`** — collection endpoint + fixed index, VPC delivery, the 100 MiB buffer cap.
- The `opensearch` arm also gains the previously missing `document_id_options` (FIREHOSE_DEFAULT vs NO_DOCUMENT_ID), and the OpenSearch/HTTP buffer caps are corrected to the provider's real 100 MiB (destination-level CEL; the shared hints message keeps its universal 0–900 / 1–128 envelope).

### Amazon MSK source

`msk_source` — cluster ARN ref (`AwsMskCluster.status.outputs.cluster_arn`), topic, PRIVATE/PUBLIC connectivity, an IAM role carrying the kafka-cluster data-plane permissions, and `read_from_timestamp` point-in-time rewind. Mutually exclusive with `kinesis_stream_source` and with delivery-stream SSE (both provider constraints as CEL).

### Secrets Manager credentials (the recommended production mode)

On exactly the four destinations the provider supports (Redshift, Splunk, HTTP endpoint, Snowflake): a `secrets_manager` block whose presence IS the enable switch (ForceNew, documented). Exactly-one-authentication-mode CELs pair it against the plaintext arms (`username`+`password`, `hec_token`, `access_key`, `user`+`private_key`), so a manifest can never carry both. The credential never lands in the manifest or IaC state, and rotation in Secrets Manager needs no delivery-stream update.

### Typed processor pipeline (breaking restructure)

`processing.processors` is an ordered list where each entry sets exactly one typed arm (the exactly-one-of CEL class):

| Arm | Purpose |
|-----|---------|
| `lambda` | Batch transformation — keeps its typed FK ref to `AwsLambda` |
| `metadata_extraction` | JQ partition-key extraction — completes dynamic partitioning |
| `decompression` | GZIP payloads (CloudWatch Logs subscriptions) |
| `cloudwatch_log_processing` | Unwrap CW-Logs subscription envelopes |
| `append_delimiter` | JSON-lines output (extended_s3 only) |
| `record_deaggregation` | Split KPL/delimited aggregates (extended_s3 only) |

The provider's raw `{type, parameters[]}` name/value model was deliberately NOT exposed: it would have destroyed the Lambda FK edge and pushed AWS's internal parameter vocabulary onto users. Both engines normalize the typed arms once (a shared `locals` translation in Terraform, a shared `normalizeProcessors` in Pulumi) and render the provider shape per destination.

**Live-caught constraint:** AWS rejects `RecordDeAggregation` (and per its docs, `AppendDelimiterToRecord`) on every destination except S3 — the first live Splunk lane failed with `InvalidArgumentException: RecordDeAggregationProcessor is not allowed for any destination type other than S3`. Now an authoring-time CEL on all seven non-S3 destinations, with the S3-only pipeline arms proven live in the extended_s3 lane instead.

### Breaking changes (no users; no compatibility shims per policy)

- `processing` reshaped from the Lambda-only block to the typed pipeline (field renumbering across the spec).
- Redshift `password` converged from a target-less `StringValueOrRef` to the family's plain-string `(sensitive)` convention.
- The dishonest "Secrets Manager not supported in v1" comments retired along with the deferral ledger entries.

### Outputs, presets, docs

- Outputs gain `destination_id` + `version_id` (the UpdateDestination coordinates); conformance case extended.
- Presets: two new marquee patterns (`05-snowflake-streaming` with Secrets Manager credentials, `06-iceberg-lakehouse` with unique-key upserts); `04-s3-parquet-analytics` now demonstrates real dynamic partitioning (metadata-extraction processor + `!{partitionKeyFromQuery:...}` prefix — the old preset enabled partitioning with no key source).
- **Two pre-existing preset defects fixed**: every preset manifest was missing the required `region` (invalid since authoring — nothing exercised them), and the presets used a legacy subdirectory layout invisible to the public-site catalog mirror (its scanner only reads flat `NN-name.yaml`+`.md` pairs) — they had never published. Flattened to the current convention; the kind's presets now appear on the site for the first time.
- README/catalog page/architecture doc rewritten to the eight-destination, three-source surface with the omissions ledger reduced to two honest entries (legacy Elasticsearch arm — a superseded API for the same domain fleet the `opensearch` arm serves; prefix expressions live in the `prefix` string).

## Engines

- **Terraform**: all new arms in `main.tf` (dynamic blocks keyed off the oneof-derived destination string), the processor normalization in `locals.tf`, the generator-owned `variables.tf` regenerated (drift-guard enrolled), floor lifted `>= 6.0.0` → `>= 6.8.0` (iceberg `append_only` is the newest attribute rendered).
- **Pulumi**: four new per-destination files following the module's per-arm layout; MSK source; secrets-manager blocks; document-ID options; pulumi-aws v7.35.0 carries every block (verified in the SDK source before implementation). The commented-out `binary: ./debug.sh` scaffold removed from `Pulumi.yaml`.
- Zero PARITY-EXCEPTIONs; `validate-outputs` green on both modules.

## Validation

- **Offline gate all green**: `make protos` ×3, 100+ spec/CEL test cases (a case per new rule and arm, including the S3-only processor guard and every exactly-one-auth pair), targeted package + Pulumi builds, `make build-go`, `secret-coverage --check`, `validate-refs --check`, drift guard, outputs conformance, `tofu validate` + offline `tofu plan` proof for all four hack manifests (extended_s3, snowflake, iceberg, MSK→OpenSearch-Serverless), `pulumi preview` proof for the three offline-only arms, all 12 manifests/scenarios CLI-validated, site catalog regenerated, scaffolding-leakage grep clean.
- **Live dual-engine E2E 4/4 green**: extended_s3 chain with the full S3-only processor pipeline + dynamic partitioning (Pulumi 5m50s / Terraform 6m20s) and the Splunk destination with the CloudWatch-Logs pipeline against a placeholder HEC endpoint — AWS activates the stream without probing the endpoint (Pulumi 5m29s / Terraform 5m55s). Zero-orphan sweep clean (Firehose, IAM, S3, tagged-resource query).
- **Deliberately not live, each with the reason recorded in the profile**: the MSK source (needs a provisioned MSK cluster whose own lane is deferred), OpenSearch/OpenSearch-Serverless/HTTP/Redshift/Snowflake (live external infrastructure or an external SaaS account), and Iceberg (the Glue catalog ARN embeds the account ID, which a committed manifest cannot carry — the SES identity-policy class). All proven by spec tests + the offline plan/preview gate.

## Live-Run Operational Learning (folded into the forge rule)

A fresh private `PULUMI_HOME` has no plugin cache, so every Pulumi command resolves plugins via anonymous GitHub API calls; back-to-back lanes exhaust the anonymous rate limit, and the failure lands wherever the next resolution happens — including `pulumi destroy`/`stack rm` during dependency teardown. Fixtures then orphan and every later lane fails DEPENDENCIES-UP with `EntityAlreadyExists`/`BucketAlreadyOwnedByYou` conflicts. Exporting `GITHUB_TOKEN` for live lanes eliminates the class; the guidance now lives in the kind-forge workflow rule, and the update rule gained a preset-layout gate (flat pairs + required-field validation) from the preset findings.

## Deferred Surface (recorded)

The legacy `elasticsearch` destination arm is the only skipped provider surface (superseded API for the same domain fleet). Everything else the provider models for this resource is now in the spec.

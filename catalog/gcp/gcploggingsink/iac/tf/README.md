# GcpLoggingSink - Terraform Module

This Terraform module provisions a Cloud Logging sink (`google_logging_project_sink`, `google_logging_folder_sink`, `google_logging_organization_sink`, or `google_logging_billing_account_sink` — exactly one, selected by `spec.scope`). It is the Terraform-side implementation of the Planton `GcpLoggingSink` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one logging sink — the routing rule that exports log entries matching a filter to a destination. The destination URI is rendered by the module from whichever arm the spec sets (GCS bucket, BigQuery dataset, Pub/Sub topic, or a raw URI escape hatch), so manifests reference resources naturally instead of hand-assembling `service.googleapis.com/...` strings.

The one post-create step every sink needs: grant the `writer_identity` output write access on the destination, or the sink silently exports nothing. Scope differences are modeled, not smoothed over — writer-identity controls exist only on project sinks; children routing only on folder/org sinks. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcploggingsink/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpLoggingSink spec | — |

The `spec` object includes: `scope` (at most one of project_id/folder_id/organization_id/billing_account; empty means the provider's default project), `destination` (exactly one of gcs_bucket/bigquery_dataset/pubsub_topic/raw_uri, plus `use_partitioned_tables` for BigQuery), `filter`, `exclusions` (carve-outs with their own filters), `include_children`/`intercept_children` (folder/org only), `unique_writer_identity` (project only; default true, sent explicitly), `custom_writer_identity` (project only), `sink_name` (empty defaults to `metadata.name`), `description`, `disabled`, and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `sink_name` | The sink name as it exists in GCP |
| `writer_identity` | `serviceAccount:{email}` — grant this write access on the destination |

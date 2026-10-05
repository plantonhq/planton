# GcpCloudRun - Terraform Module

This Terraform module provisions a Cloud Run service (`google_cloud_run_v2_service`). It is the Terraform-side implementation of the Planton `GcpCloudRun` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Cloud Run Admin API (`run.googleapis.com`) so a fresh project works first try, then creates the service with the full provider surface: multiple containers (the serving container plus sidecars with `depends_on` startup ordering), literal and Secret Manager environment variables, startup probes (HTTP/TCP/gRPC), liveness and readiness probes (HTTP/gRPC — the API accepts no TCP arm on either), all five volume types (Cloud SQL sockets, Secret Manager files, in-memory or disk scratch, GCS FUSE with mount options, NFS) with `sub_path` mounts, per-revision and service-level scaling (including MANUAL mode and the service-wide max), revision traffic splitting with tags, direct VPC egress or a Serverless VPC Access connector, GPU node selection, CMEK image encryption, Binary Authorization, session affinity, custom audiences, deploy-from-source builds (`build_config` + container `base_image_uri`), Identity-Aware Proxy (`iap_enabled`), multi-region services (`multi_region_settings` with `region = "global"`), service and revision annotations/labels, `default_uri_disabled`, and `health_check_disabled`.

Public access is the additive-IAM path: when `allow_unauthenticated` is true a `google_cloud_run_v2_service_iam_member` grants `roles/run.invoker` to `allUsers`; `invoker_iam_disabled` is the org-policy alternative that switches the IAM check off instead (the spec rejects setting both). `deletion_protection` defaults to true — a destroy fails until the manifest opts out — and `deletion_policy` sets the Terraform-side destroy stance (PREVENT / ABANDON).

An empty `project_id` falls back to the provider's default project; empty optional strings become null so the provider omits them from the API payload. Enum values arrive from the spec as the API's own names and pass through untranslated. The module runs on the plain `google` provider — every modeled field is GA at the pinned major.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpcloudrun/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpCloudRun spec | — |

The `spec` object includes: `region` (required; `"global"` for multi-region services), `containers` (required; image, command/args, env with Secret Manager references, single serving port with `h2c` support, cpu/memory limits + `cpu_idle`/`startup_cpu_boost`, volume mounts with `sub_path`, startup/liveness/readiness probes, `depends_on`, `base_image_uri`), `volumes` (cloud_sql_instance / secret / empty_dir / gcs with `mount_options` / nfs — exactly one source each), `service_account` (empty = Compute Engine default), `scaling` + `service_scaling` (incl. the service-wide `max_instance_count`), `max_instance_request_concurrency`, `timeout_seconds`, `execution_environment`, `session_affinity`, `encryption_key`, `revision`, `revision_labels`/`revision_annotations`, `vpc_access` (connector XOR network_interfaces + egress), `node_selector`/`gpu_zonal_redundancy_disabled`, `ingress`, `allow_unauthenticated` XOR `invoker_iam_disabled`, `custom_audiences`, `traffic`, `launch_stage`, `binary_authorization`, `build_config`, `iap_enabled`, `default_uri_disabled`, `health_check_disabled`, `multi_region_settings`, `annotations`, `deletion_protection` (default true), `deletion_policy`, and `project_id` (empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `url` | Canonical serving URL of the service |
| `service_name` | Name of the service as created in GCP (the serverless-NEG handle) |
| `revision` | Latest ready revision name |
| `location` | Region the service is deployed in |
| `uid` | Server-assigned unique identifier |
| `urls` | Every URL serving this service |

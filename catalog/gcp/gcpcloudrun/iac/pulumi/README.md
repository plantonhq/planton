# GcpCloudRun - Pulumi Module

This Pulumi (Go) module provisions a Cloud Run service (`cloudrunv2.Service`). It is the Pulumi-side implementation of the Planton `GcpCloudRun` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Cloud Run Admin API (`run.googleapis.com`) so a fresh project works first try, then creates the service with the full provider surface: multiple containers (the serving container plus sidecars with `dependsOn` startup ordering), literal and Secret Manager environment variables, startup probes (HTTP/TCP/gRPC), liveness and readiness probes (HTTP/gRPC — the API accepts no TCP arm on either), all five volume types (Cloud SQL sockets, Secret Manager files, in-memory or disk scratch, GCS FUSE with mount options, NFS) with `subPath` mounts, per-revision and service-level scaling (including MANUAL mode and the service-wide max), revision traffic splitting with tags, direct VPC egress or a Serverless VPC Access connector, GPU node selection, CMEK image encryption, Binary Authorization, session affinity, custom audiences, deploy-from-source builds (`buildConfig` + container `baseImageUri`), Identity-Aware Proxy (`iapEnabled`), multi-region services (`multiRegionSettings` with `region: global`), service and revision annotations/labels, `defaultUriDisabled`, and `healthCheckDisabled`.

Public access is the additive-IAM path: when `allow_unauthenticated` is true a `cloudrunv2.ServiceIamMember` grants `roles/run.invoker` to `allUsers`; `invoker_iam_disabled` is the org-policy alternative that switches the IAM check off instead (the spec rejects setting both). `deletion_protection` defaults to true — a destroy fails until the manifest opts out — and `deletion_policy` sets the engine-side destroy stance (PREVENT / ABANDON).

An empty `project_id` falls back to the provider's default project — the ambient-project contract every GCP kind honors. Enum values pass through as the API's own names, untranslated.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Build

```bash
cd catalog/gcp/gcpcloudrun/iac/pulumi
make build
```

## Outputs

| Name | Description |
|------|-------------|
| `url` | Canonical serving URL of the service |
| `service_name` | Name of the service as created in GCP (the serverless-NEG handle) |
| `revision` | Latest ready revision name |
| `location` | Region the service is deployed in |
| `uid` | Server-assigned unique identifier |
| `urls` | Every URL serving this service |

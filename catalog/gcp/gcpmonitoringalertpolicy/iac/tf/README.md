# GcpMonitoringAlertPolicy - Terraform Module

This Terraform module provisions a Cloud Monitoring alert policy (`google_monitoring_alert_policy`). It is the Terraform-side implementation of the Planton `GcpMonitoringAlertPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one alert policy — the rule that watches metrics or logs (threshold, absence, log match, MQL, PromQL, or SQL conditions, combined by `combiner`) and notifies the referenced channels when incidents open. Each condition carries exactly one condition-type arm, enforced by the spec's validations (the provider leaves the API's oneof unchecked client-side).

`enabled` is sent explicitly on every apply so disabling a policy actually reaches the API. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpmonitoringalertpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpMonitoringAlertPolicy spec | — |

The `spec` object includes: `combiner` (required — AND/OR/AND_WITH_MATCHING_RESOURCE), `conditions` (1–6, each with exactly one condition arm), `notification_channels` (channel resource names), `severity` (CRITICAL/ERROR/WARNING), `enabled` (default true, sent explicitly), `alert_strategy` (auto-close, rate limit — required for log-based policies — and re-notification), `documentation` (runbook content + links), `display_name` (empty defaults to `metadata.name`), `project_id` (empty falls back to the provider default project), `labels` (user metadata), and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `policy_name` | `projects/{project}/alertPolicies/{id}` |

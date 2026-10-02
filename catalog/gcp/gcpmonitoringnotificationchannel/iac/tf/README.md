# GcpMonitoringNotificationChannel - Terraform Module

This Terraform module provisions a Cloud Monitoring notification channel (`google_monitoring_notification_channel`). It is the Terraform-side implementation of the Planton `GcpMonitoringNotificationChannel` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one notification channel — the delivery endpoint (email, Slack, PagerDuty, SMS, webhook, or Pub/Sub) that alert policies notify when incidents open or close. Alert policies reference the channel by its resource name (the `channel_name` output).

Two label surfaces exist and are never conflated: the provider's `labels` argument is the type-specific channel configuration (fed from `spec.channel_labels`), while `user_labels` is freeform metadata (fed from `spec.labels` merged with the platform attribution labels). Credentials ride the `sensitive_labels` block, stored and redacted API-side. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpmonitoringnotificationchannel/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpMonitoringNotificationChannel spec | — |

The `spec` object includes: `type` (required — email/sms/slack/pagerduty/webhook_tokenauth/webhook_basicauth/pubsub; GCP validates against its live type catalog), `channel_labels` (type-specific non-secret config; credential keys are refused by validation), `sensitive_labels` (auth_token/password/service_key — secrets), `enabled` (default true, sent explicitly), `force_delete`, `display_name` (empty defaults to `metadata.name`), `project_id` (empty falls back to the provider default project), `labels` (user metadata), and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `channel_name` | `projects/{project}/notificationChannels/{id}` — the value alert policies reference |
| `verification_status` | Verification state (SMS/email channels deliver only after verification) |

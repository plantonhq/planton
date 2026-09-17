# GcpMonitoringUptimeCheck - Terraform Module

This Terraform module provisions a Cloud Monitoring uptime check (`google_monitoring_uptime_check_config`). It is the Terraform-side implementation of the Planton `GcpMonitoringUptimeCheck` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one uptime check — the probe Google runs against a public URL, monitored resource, resource group, or synthetic-monitor Cloud Function from multiple regions on a fixed cadence. A check only measures; pair it with a `GcpMonitoringAlertPolicy` filtering on `uptime_check_passed` and this check's `uptime_check_id` to actually page.

Exactly one target block and — except for synthetic monitors — exactly one check block render, enforced by the spec's validations before the module runs. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpmonitoringuptimecheck/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpMonitoringUptimeCheck spec | — |

The `spec` object includes: `timeout` (required, 1s–60s), `period` (60s/300s/600s/900s; empty defaults to 300s), one target (`monitored_resource` | `resource_group` | `synthetic_monitor`), one check (`http_check` | `tcp_check`; omit both only for a synthetic monitor), `content_matchers` (response assertions), `checker_type`, `selected_regions`, `log_check_failures`, `display_name` (empty defaults to `metadata.name`), `project_id` (empty falls back to the provider default project), `labels` (user metadata), and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `uptime_check_name` | `projects/{project}/uptimeCheckConfigs/{id}` |
| `uptime_check_id` | The bare check ID — the value alert policies filter on to page on failures |

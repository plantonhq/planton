# GCP Log Metric - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Logging log-based metric using Planton's `GcpLogMetric` API. The module is written in Go and creates `logging.Metric` — the bridge from log entries matching a filter to a chartable Cloud Monitoring metric (a counter, or a distribution extracted from entry fields).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Cloud Logging API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── log_metric.go          # Metric creation + descriptor/bucket expansion
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `metric_name` | `name` | Defaults to metadata.name |
| `filter` | `filter` | Required |
| `bucket_name` | `bucket_name` | Bucket-scoped metrics; references a GcpLogBucket's `bucket_name` output |
| `disabled` | `disabled` | Sent EXPLICITLY on every apply — a true→false transition must reach the API |
| `metric_descriptor.*` | `metric_descriptor.*` | Kind/value-type/unit/display-name/label schema |
| `value_extractor`, `label_extractors` | same names | The DISTRIBUTION extraction surface |
| `bucket_options.*` | `bucket_options.*` | Explicit/exponential/linear histogram layouts |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | Omitted when empty (provider default DELETE) |

The module also enables `logging.googleapis.com` on the target project
(`disable_on_destroy` false — tearing down one metric never disables
logging project-wide).

## Outputs

| Output | Description |
|---|---|
| `metric_name` | The metric name — address it from Monitoring as `logging.googleapis.com/user/{metric_name}` |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```

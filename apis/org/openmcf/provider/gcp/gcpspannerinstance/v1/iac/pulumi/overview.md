# GcpSpannerInstance Pulumi Module Architecture

## Overview

This module creates a Google Cloud Spanner instance using the `spanner.NewInstance` resource from the `pulumi-gcp` provider.

## Resource Flow

```
StackInput (GcpSpannerInstanceStackInput)
  ├── target.spec.project_id    → spanner.Instance.Project
  ├── target.spec.instance_name → spanner.Instance.Name
  ├── target.spec.config        → spanner.Instance.Config
  ├── target.spec.display_name  → spanner.Instance.DisplayName
  ├── target.spec.num_nodes     → spanner.Instance.NumNodes
  ├── target.spec.processing_units → spanner.Instance.ProcessingUnits
  ├── target.spec.autoscaling_config → spanner.Instance.AutoscalingConfig
  ├── target.spec.instance_type → spanner.Instance.InstanceType
  ├── target.spec.edition       → spanner.Instance.Edition
  ├── target.spec.default_backup_schedule_type → spanner.Instance.DefaultBackupScheduleType
  ├── target.spec.force_destroy → spanner.Instance.ForceDestroy
  └── locals.GcpLabels          → spanner.Instance.Labels
```

## Capacity Model

The module handles three mutually exclusive capacity modes:

1. **Fixed nodes** (`num_nodes > 0`): Sets `NumNodes` directly
2. **Fixed processing units** (`processing_units > 0`): Sets `ProcessingUnits` directly
3. **Autoscaling** (`autoscaling_config != nil`): Sets `AutoscalingConfig` with limits and targets

The proto-level CEL validations ensure mutual exclusion before the module runs.

## Labels

Framework labels are computed in `locals.go` and applied to the instance. Labels include:
- `openmcf-resource: true`
- `openmcf-resource-name: {instance_name}`
- `openmcf-resource-kind: gcpspannerinstance`
- `openmcf-organization: {org}` (if set)
- `openmcf-environment: {env}` (if set)
- `openmcf-resource-id: {id}` (if set)

## Outputs

- `instance_id`: Fully qualified path `projects/{project}/instances/{name}`
- `instance_name`: Short name (used by GcpSpannerDatabase)
- `state`: CREATING or READY

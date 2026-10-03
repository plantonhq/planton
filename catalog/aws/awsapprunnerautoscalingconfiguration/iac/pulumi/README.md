# AwsAppRunnerAutoScalingConfiguration — Pulumi Module

## Overview

This Pulumi module registers an App Runner auto scaling configuration version -- the reusable scaling policy (warm floor, scale-out ceiling, per-instance concurrency) that App Runner services adopt by ARN.

## Module Structure

```
module/
  main.go                        — Entry point: provider setup, orchestrate resource creation
  locals.go                      — Identity tag set
  auto_scaling_configuration.go  — apprunner.AutoScalingConfigurationVersion resource + output exports
  outputs.go                     — Output key constants
```

## IaC Inputs

The module reads `AwsAppRunnerAutoScalingConfigurationIacInput` which contains:
- `target` — The fully-specified `AwsAppRunnerAutoScalingConfiguration` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `configuration_arn` | The revision-carrying ARN services reference |
| `configuration_revision` | The revision this deployment registered |
| `latest` | Whether this revision is the newest under the name |

## Key Implementation Notes

- **Everything is ForceNew by design**: AWS versions this resource; any value change registers the next revision under the same name, and the new ARN rolls referencing services through the resource graph.
- **Explicit naming**: the configuration's cloud name is `metadata.name`, matching the Terraform module's physical identity.

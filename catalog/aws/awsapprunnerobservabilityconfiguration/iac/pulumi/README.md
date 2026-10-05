# AwsAppRunnerObservabilityConfiguration — Pulumi Module

## Overview

This Pulumi module registers an App Runner observability configuration version -- the reusable tracing policy (AWS X-Ray via a managed OpenTelemetry collector) that App Runner services adopt by ARN.

## Module Structure

```
module/
  main.go                          — Entry point: provider setup, orchestrate resource creation
  locals.go                        — Identity tag set
  observability_configuration.go   — apprunner.ObservabilityConfiguration resource + output exports
  outputs.go                       — Output key constants
```

## IaC Inputs

The module reads `AwsAppRunnerObservabilityConfigurationIacInput` which contains:
- `target` — The fully-specified `AwsAppRunnerObservabilityConfiguration` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `configuration_arn` | The revision-carrying ARN services reference |
| `configuration_revision` | The revision this deployment registered |
| `latest` | Whether this revision is the newest under the name |

## Key Implementation Notes

- **Trace settings are ForceNew by design**: AWS versions this resource; changes register the next revision under the same name, and the new ARN rolls referencing services.
- **Explicit naming**: the configuration's cloud name is `metadata.name`, matching the Terraform module's physical identity.
- **The trace block is optional**: emitted only when the spec configures tracing; a configuration without it is valid but inert.

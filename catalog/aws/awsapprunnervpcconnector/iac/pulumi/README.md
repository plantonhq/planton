# AwsAppRunnerVpcConnector — Pulumi Module

## Overview

This Pulumi module provisions an App Runner VPC connector -- the managed ENI set that App Runner services route their outbound traffic through to reach private VPC resources.

## Module Structure

```
module/
  main.go           — Entry point: provider setup, orchestrate resource creation
  locals.go         — Identity tag set
  vpc_connector.go  — apprunner.VpcConnector resource + output exports
  outputs.go        — Output key constants
```

## IaC Inputs

The module reads `AwsAppRunnerVpcConnectorIacInput` which contains:
- `target` — The fully-specified `AwsAppRunnerVpcConnector` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `vpc_connector_arn` | The ARN services set as their egress `vpc_connector_arn` |
| `vpc_connector_revision` | The revision of this connector under its name |
| `status` | Lifecycle status (`ACTIVE` when ready) |

## Key Implementation Notes

- **Immutable attachment**: AWS has no update API for connectors -- changing subnets or security groups replaces the connector (a new revision under the same name).
- **Explicit naming**: the connector's cloud name is `metadata.name`, matching the Terraform module's physical identity.
- **Security groups are required**: AWS mandates at least one group; targets' groups must admit ingress from it.

# AwsHttpApiVpcLink — Pulumi Module

## Overview

This Pulumi module provisions an API Gateway v2 VPC link -- the managed ENI set that HTTP API private integrations route through to reach internal ALB/NLB/Cloud Map backends.

## Module Structure

```
module/
  main.go      — Entry point: provider setup, orchestrate resource creation
  locals.go    — Identity tag set
  vpc_link.go  — apigatewayv2.VpcLink resource + output exports
  outputs.go   — Output key constants
```

## IaC Inputs

The module reads `AwsHttpApiVpcLinkIacInput` which contains:
- `target` — The fully-specified `AwsHttpApiVpcLink` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `vpc_link_id` | The VPC link ID (what private integrations set as `connection_id`) |
| `vpc_link_arn` | ARN of the VPC link |

## Key Implementation Notes

- **Immutable attachment**: AWS has no update API for subnets/security groups -- changing either replaces the link. Only the name mutates in place.
- **Explicit naming**: the link's cloud name is `metadata.name`, matching the Terraform module's physical identity.
- **Empty security groups**: a link without security groups applies no filtering on its side; the target's security groups govern reachability.

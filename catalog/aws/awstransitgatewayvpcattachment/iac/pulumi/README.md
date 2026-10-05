# AwsTransitGatewayVpcAttachment — Pulumi Module

## Overview

This Pulumi module provisions a Transit Gateway VPC attachment -- the connection that plugs one VPC into a Transit Gateway hub through ENIs in the chosen subnets.

## Module Structure

```
module/
  main.go            — Entry point: provider setup, orchestrate resource creation
  locals.go          — Identity tag set + enable/disable dial helpers
  vpc_attachment.go  — ec2transitgateway.VpcAttachment resource
  outputs.go         — Output key constants
```

## IaC Inputs

The module reads `AwsTransitGatewayVpcAttachmentIacInput` which contains:
- `target` — The fully-specified `AwsTransitGatewayVpcAttachment` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `attachment_id` | The attachment ID referenced by route table associations, propagations, and routes |
| `attachment_arn` | The attachment ARN for IAM policies |
| `vpc_owner_id` | The AWS account that owns the attached VPC |

## Behavior Notes

- The gateway and VPC references are create-time immutable; the subnet set
  updates in place.
- Tri-state options (proto `optional bool`) are omitted when unset so the
  provider/AWS default -- or the gateway-inherited value, for the
  default-table membership pair -- applies.

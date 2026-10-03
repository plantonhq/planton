# AwsTransitGateway — Pulumi Module

## Overview

This Pulumi module provisions an AWS Transit Gateway -- the regional hub that VPC attachments, route tables, VPN connections, and Direct Connect gateways compose onto.

## Module Structure

```
module/
  main.go             — Entry point: provider setup, orchestrate resource creation
  locals.go           — Identity tag set + enable/disable dial helpers
  transit_gateway.go  — ec2transitgateway.TransitGateway resource
  outputs.go          — Output key constants
```

## IaC Inputs

The module reads `AwsTransitGatewayIacInput` which contains:
- `target` — The fully-specified `AwsTransitGateway` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `transit_gateway_id` | The gateway ID referenced by attachments, route tables, and subnet routes |
| `transit_gateway_arn` | The gateway ARN for IAM policies and RAM sharing |
| `owner_id` | The owning AWS account ID |
| `association_default_route_table_id` | Default association route table (empty when disabled) |
| `propagation_default_route_table_id` | Default propagation route table (empty when disabled) |

## Behavior Notes

- Tri-state dials (proto `optional bool`) are omitted when unset so the
  provider/AWS default applies -- the same null fall-through as the
  Terraform module.
- The default-table association/propagation dials carry AWS's asymmetric
  replacement rule: disable -> enable replaces the gateway; enable ->
  disable updates in place.

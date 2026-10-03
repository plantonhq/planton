# AwsTransitGatewayRouteTable — Pulumi Module

## Overview

This Pulumi module provisions a Transit Gateway route table -- one isolated routing domain inside a Transit Gateway hub -- together with its folded associations, propagations, static routes, and prefix list references.

## Module Structure

```
module/
  main.go        — Entry point: provider setup, orchestrate resource creation
  locals.go      — Identity tag set
  route_table.go — RouteTable + per-member association/propagation/route/prefix-list resources
  outputs.go     — Output key constants
```

## IaC Inputs

The module reads `AwsTransitGatewayRouteTableIacInput` which contains:
- `target` — The fully-specified `AwsTransitGatewayRouteTable` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `route_table_id` | The route table ID (tgw-rtb-...) |
| `route_table_arn` | The route table ARN for IAM policies |

## Behavior Notes

- Every folded member materializes as its own provider resource keyed by a
  value unique within the table (attachment ID, destination CIDR, prefix
  list ID), so membership changes are surgical.
- An attachment can be ASSOCIATED with at most one route table across the
  whole gateway; AWS enforces that at apply time (documents cannot see each
  other). Propagations carry no such limit.

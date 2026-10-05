# AwsHttpApiGateway — Pulumi Module

## Overview

This Pulumi module provisions an API Gateway HTTP API (v2) as one coherent unit: the API, deduplicated integrations (Lambda proxy, HTTP proxy, private VPC-link proxy, or AWS service actions), optional authorizers, the routes that bind them, and the deployment stage.

## Module Structure

```
module/
  main.go         — Entry point: provider setup, api → integrations → authorizers → routes → stage
  locals.go       — Identity tags + whole-object integration dedup
  api.go          — aws_apigatewayv2_api (CORS, version, IP address type)
  integration.go  — Deduplicated integrations (all three backend shapes)
  authorizer.go   — JWT + REQUEST (Lambda) authorizers
  route.go        — Routes binding integrations and authorizers
  stage.go        — Stage with access logs, throttling, per-route settings
  outputs.go      — Output key constants
```

## IaC Inputs

The module reads `AwsHttpApiGatewayIacInput` which contains:
- `target` — The fully-specified `AwsHttpApiGateway` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `api_id` | The API identifier (the join key domain mappings reference) |
| `api_endpoint` | The default endpoint URL |
| `api_arn` | ARN of the API |
| `execution_arn` | Execution ARN prefix for Lambda resource policies |
| `stage_invoke_url` | Invoke URL of the deployed stage |
| `stage_name` | Name of the deployed stage |

## Key Implementation Notes

- **Integration dedup**: routes whose integration blocks are identical (proto equality across every field) share one integration resource; the Terraform module dedups with the same whole-object rule.
- **auto_deploy presence rule**: an omitted `auto_deploy` means true; only an explicit `false` turns it off. Identical rule in the Terraform module.
- **Stage after routes**: per-route settings reference routes by key, so the stage depends on the created routes.
- **Service integrations**: `integration_subtype` integrations send no `integration_uri` (the action's parameters ride `request_parameters`) and are pinned to payload format 1.0 by AWS.
- **Explicit naming**: the API's cloud name is `metadata.name`, matching the Terraform module's physical identity.

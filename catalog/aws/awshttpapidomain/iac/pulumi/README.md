# AwsHttpApiDomain — Pulumi Module

## Overview

This Pulumi module provisions an API Gateway v2 custom domain name and its API mappings: TLS termination with an ACM certificate, optional mutual TLS, and one mapping resource per configured API.

## Module Structure

```
module/
  main.go         — Entry point: provider setup, orchestrate resource creation
  locals.go       — Identity tag set
  domain_name.go  — apigatewayv2.DomainName + per-entry apigatewayv2.ApiMapping + exports
  outputs.go      — Output key constants
```

## IaC Inputs

The module reads `AwsHttpApiDomainIacInput` which contains:
- `target` — The fully-specified `AwsHttpApiDomain` resource
- `provider_config` — AWS credentials/region resolution

## Outputs

| Key | Description |
|-----|-------------|
| `domain_name` | The custom domain name (the domain's join key) |
| `domain_name_arn` | ARN of the domain name resource |
| `target_domain_name` | The regional domain to alias/CNAME to from DNS |
| `hosted_zone_id` | Route 53 alias target zone of the regional endpoint |

## Key Implementation Notes

- **Endpoint type / security policy hardcoded**: v2 domains accept only REGIONAL + TLS_1_2, so neither is a spec field.
- **Root mapping addressing**: the empty mapping key (the domain root) gets the `root` resource-name alias; the real (absent) key is what reaches AWS.
- **Nested config outputs**: `target_domain_name` and `hosted_zone_id` are read from the domain-name configuration output type -- the DNS composition surface.

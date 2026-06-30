# Azure Cache for Redis Deployment Component

**Date**: February 14, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Provider Framework

## Summary

Added AzureRedisCache (enum 431, id_prefix: azred) as a complete Planton deployment component with both Pulumi and Terraform IaC modules. This is the 16th Azure resource forged in the azure-resource-expansion sub-project (R15 in the queue), completing the database/cache category alongside PostgreSQL, MySQL, MSSQL, and CosmosDB.

## Problem Statement / Motivation

Azure Cache for Redis is a critical piece of enterprise Azure architectures, serving as the caching and session management layer for web applications, container apps, and microservices. Without it, the database-stack, container-apps-environment, and web-app-environment infra charts lack an optional cache tier.

### Pain Points

- No Planton component for Azure's managed Redis service
- Infra charts couldn't include an optional caching layer
- Users deploying Azure workloads had to manage Redis separately

## Solution / What's New

A complete AzureRedisCache deployment component covering all three SKU tiers (Basic, Standard, Premium) with 11 corrections from the original T02 spec design based on deep Terraform provider research.

### Key Design Decisions

- **SKU family auto-derived**: Users specify `sku_name` (Basic/Standard/Premium) and `capacity` (size). The Azure-required `family` field ("C" or "P") is auto-derived in IaC modules -- zero user-facing complexity for a three-field Azure API quirk
- **maxmemory_policy exposed**: The single most important Redis configuration field, absent from the original plan. Controls eviction behavior when memory is full
- **Structured patch schedules**: T02 spec had `repeated string patch_schedule` (wrong). Corrected to `repeated AzureRedisPatchSchedule` with `day_of_week`, `start_hour_utc`, and `maintenance_window`
- **Firewall rule name constraint**: Azure requires `^\w+$` (no hyphens). CEL validation enforces this at the spec level
- **string+CEL for enums**: Follows established pattern (NSG, LB, PostgreSQL, CosmosDB) using provider-authentic string values with CEL `in` validation instead of proto enums

## Implementation Details

### Proto API (4 files)

- `spec.proto`: 15 fields on the main message, 2 supporting messages (AzureRedisPatchSchedule, AzureRedisFirewallRule), 8 CEL validations
- `stack_outputs.proto`: 5 outputs (redis_id, hostname, ssl_port, primary_access_key, primary_connection_string)
- `api.proto`: Standard KRM wiring (api_version, kind, metadata, spec, status)
- `stack_input.proto`: Target + AzureProviderConfig

### Validation Tests (39 tests)

- 19 positive tests covering all valid enum values, optional field combinations, StringValueOrRef references
- 20 negative tests covering missing required fields, invalid enum values, constraint violations
- Notably tests: firewall rule names with hyphens (must fail), capacity bounds, patch schedule day validation

### Pulumi Module

- Uses `redis.NewCache` and `redis.NewFirewallRule` from `pulumi-azure/sdk/v6/go/azure/redis`
- Family derivation in `locals.go`: `if sku_name == "Premium" then "P" else "C"`
- Patch schedules as `CachePatchScheduleArray` embedded in cache args
- Firewall rules created with explicit `DependsOn` on cache

### Terraform Module

- `azurerm_redis_cache.main` with dynamic `patch_schedule` block
- `azurerm_redis_firewall_rule.rules` via `for_each`
- Family computed in `locals.tf`: `var.spec.sku_name == "Premium" ? "P" : "C"`

### Corrections from T02 Spec (11 total)

1. Added `resource_group` (StringValueOrRef) -- missing from T02
2. Added `region` (string) -- missing from T02
3. Changed `sku` proto enum to `sku_name` string+CEL
4. Auto-derived `family` (not user-facing)
5. Corrected `patch_schedule` from `repeated string` to structured message
6. Added `minimum_tls_version`
7. Added `public_network_access_enabled`
8. Added `maxmemory_policy` (8 valid eviction policies)
9. Added `zones` for availability zone support
10. Added firewall rule name format validation (`^\w+$`)
11. Validated `redis_version` values ("4", "6")

## Benefits

- **Database category complete**: PostgreSQL + MySQL + MSSQL + CosmosDB + Redis covers all major Azure data services
- **Infra chart ready**: Redis can now be an optional component in database-stack, container-apps-environment, and web-app-environment charts
- **Production-quality docs**: README with SKU comparison tables, eviction policy guidance, network access patterns; 7 YAML examples; comprehensive research docs

## Impact

- **Users**: Can deploy Azure Cache for Redis through Planton with the same declarative YAML pattern as all other resources
- **Infra charts**: Database-stack and app-environment charts can now include optional caching layers
- **Downstream**: AzurePrivateEndpoint can reference `redis_id` for private connectivity

## Related Work

- Part of: 20260212.05.sp.azure-resource-expansion (R15 of 24)
- Follows: R14 AzureCosmosdbAccount (database category complete)
- Next: R16 AzureServicePlan (app hosting category begins)
- Database trifecta + cache: PostgreSQL (R11) + MySQL (R12) + MSSQL (R13) + CosmosDB (R14) + Redis (R15)

## Code Metrics

- **Files created**: ~33 (proto, Go, HCL, Markdown, YAML, shell)
- **Tests**: 39 spec validation tests (39 pass, 0 fail)
- **Enum**: 431 (AzureRedisCache)
- **ID prefix**: azred

---

**Status**: Production Ready
**Timeline**: Single session

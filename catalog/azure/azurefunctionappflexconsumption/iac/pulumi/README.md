# AzureFunctionAppFlexConsumption Pulumi Module

This directory contains the Pulumi IaC implementation for the `AzureFunctionAppFlexConsumption` component.

## Structure

```
pulumi/
├── main.go          # Entrypoint (loads IaC input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── README.md        # This file
└── module/
    ├── main.go      # Resource creation (azurerm_function_app_flex_consumption)
    ├── locals.go    # Local variable initialization + enum wire maps
    └── outputs.go   # Output key constants
```

## What It Creates

One `appservice.AppFlexConsumption` (the classic-SDK wrapper of `azurerm_function_app_flex_consumption`): the Flex Consumption Function App with its runtime declaration, deployment-storage binding, scale configuration (instance memory, fan-out ceiling, HTTP concurrency, always-ready pools), site config, app settings, connection strings, sticky settings, managed identity, and Easy Auth v2.

## Behavior Notes

- **The FC1 gate**: the provider verifies the referenced plan's SKU is FC1 at create time -- the plan reference must point at a Flex Consumption plan.
- **storage_container_type is a constant**: `blobContainer` is the argument's single legal value, so the spec omits it and this module sends the constant.
- **The storage-auth triangle** is spec-enforced (key with connection-string auth, identity id with user-assigned auth), mirroring the provider's create-time checks.
- **Presence-guarded proto defaults**: unset optionals deploy the spec's documented defaults (instance memory 2048, max instances 100, https_only true, TLS floors 1.2), never the Go zero value.
- **Write-only fields** (`storage_access_key`, `zip_deploy_file`, `app_service_logs`) do not round-trip -- the provider re-sends them from configuration; re-supplying them on import is expected, not drift.
- **Outputs** flatten onto the same proto fields as the Terraform module (outbound IP sets exported as real lists; identity and site-credential outputs populated via ApplyT guards).

## Building

```bash
go build -o /dev/null ./catalog/azure/azurefunctionappflexconsumption/iac/pulumi
```

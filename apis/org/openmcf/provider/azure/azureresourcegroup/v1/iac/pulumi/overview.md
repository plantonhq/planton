# AzureResourceGroup Pulumi Module -- Architecture Overview

## Purpose

This module is the Pulumi implementation for the `AzureResourceGroup` OpenMCF component.
It translates the protobuf-defined spec into Azure infrastructure using the Pulumi
Azure Classic SDK.

## Architecture

```
AzureResourceGroupStackInput
  ├── target (AzureResourceGroup)
  │     ├── metadata (name, org, env)
  │     └── spec (name, region)
  └── provider_config (credentials)
         │
         ▼
  ┌─────────────────┐
  │  module/main.go  │  Creates azure provider + resource group
  │  module/locals.go│  Initializes tags from metadata
  │  module/outputs.go│ Defines output constant names
  └─────────────────┘
         │
         ▼
  Stack Outputs: resource_group_id, resource_group_name, region
```

## Module Structure

- `main.go` -- Pulumi entrypoint that loads stack input and calls module
- `module/main.go` -- Core implementation: creates provider and resource group
- `module/locals.go` -- Initializes locals struct with tags from metadata
- `module/outputs.go` -- Output constant definitions for consistent naming

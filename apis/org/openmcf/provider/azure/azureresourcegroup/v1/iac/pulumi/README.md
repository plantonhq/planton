# AzureResourceGroup Pulumi Module

## Overview

This Pulumi module provisions an Azure Resource Group using the Azure Classic provider
(`pulumi-azure`). It creates a single `azurerm_resource_group` with OpenMCF metadata tags.

## Resources Created

- `core.ResourceGroup` -- the Azure Resource Group

## Inputs

The module receives an `AzureResourceGroupStackInput` containing:

- `target.spec.name` -- resource group name
- `target.spec.region` -- Azure region
- `target.metadata` -- OpenMCF metadata (name, org, env) used for tagging
- `provider_config` -- Azure credentials (client_id, client_secret, subscription_id, tenant_id)

## Outputs

| Output | Description |
|--------|-------------|
| `resource_group_id` | Azure Resource Manager ID |
| `resource_group_name` | Name of the created resource group |
| `region` | Azure region |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest openmcf
```

## Debugging

```bash
./debug.sh       # Start Delve debugger on port 2345
```

# Azure Function App Environment

Serverless function hosting on Azure with monitoring and optional secrets management.

## What This Chart Deploys

| Resource | Kind | Condition |
|----------|------|-----------|
| Resource Group | `AzureResourceGroup` | Always |
| Service Plan | `AzureServicePlan` | Always |
| Storage Account | `AzureStorageAccount` | Always |
| Log Analytics Workspace | `AzureLogAnalyticsWorkspace` | Always |
| Application Insights | `AzureApplicationInsights` | `create_app_insights` |
| Function App | `AzureFunctionApp` | Always |
| Key Vault | `AzureKeyVault` | `create_key_vault` |

## Architecture

The chart creates a Linux Function App with its required dependencies:

- **Service Plan** defines the compute tier (Consumption Y1 for pay-per-execution,
  Elastic Premium EP1-EP3 for pre-warmed instances, or Dedicated B1-P3v3)
- **Storage Account** provides the runtime storage required by Azure Functions
  for triggers, bindings, and execution state
- **Log Analytics Workspace** collects function execution logs
- **Application Insights** (optional, default on) provides request tracing,
  dependency tracking, and performance metrics
- **Key Vault** (optional) stores secrets that Function Apps can reference
  via Key Vault References

## Runtime Stacks

| `runtime_stack` | Description | Example `runtime_version` |
|-----------------|-------------|---------------------------|
| `node` | Node.js | `20`, `18` |
| `python` | Python | `3.11`, `3.10` |
| `java` | Java | `17`, `11` |
| `dotnet` | .NET (in-process) | `8.0`, `6.0` |
| `dotnet-isolated` | .NET (isolated worker) | `8.0`, `6.0` |
| `custom` | Custom handler | (none) |

## Parameters

### Foundation

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Azure region | `eastus` |
| `resource_group_name` | Resource group suffix | `func-rg` |

### Service Plan

| Parameter | Description | Default |
|-----------|-------------|---------|
| `plan_name` | Plan name | `func-plan` |
| `sku_name` | SKU (Y1, EP1-EP3, B1-P3v3) | `Y1` |
| `os_type` | OS (Linux/Windows) | `Linux` |

### Storage

| Parameter | Description | Default |
|-----------|-------------|---------|
| `storage_account_name` | Globally unique name | `myfuncsa12345` |

### Function App

| Parameter | Description | Default |
|-----------|-------------|---------|
| `function_app_name` | App name | `my-func-app` |
| `runtime_stack` | Runtime language | `node` |
| `runtime_version` | Language version | `20` |
| `functions_extension_version` | Functions runtime | `~4` |

### Monitoring & Secrets

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_app_insights` | Enable APM | `true` |
| `create_key_vault` | Enable Key Vault | `false` |
| `key_vault_name` | Key Vault name | `my-func-kv` |

## Example

Deploy a Python Function App on Elastic Premium with monitoring:

```yaml
params:
  region: westus2
  resource_group_name: myapi-func-rg
  plan_name: myapi-plan
  sku_name: EP1
  storage_account_name: myapifuncsa
  function_app_name: myapi-functions
  runtime_stack: python
  runtime_version: "3.11"
  create_key_vault: true
  key_vault_name: myapi-kv
```

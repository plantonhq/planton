# AzureServicePlan Terraform Module

This directory contains the Terraform IaC implementation for the `AzureServicePlan` kind.

## Structure

```
tf/
├── main.tf          # Service Plan resource definition
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values
├── locals.tf        # Local computations (tags)
├── provider.tf      # Azure provider configuration
└── README.md        # This file
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Service Plan | `azurerm_service_plan` | Always |

## Usage

```hcl
module "service_plan" {
  source = "./path/to/module"

  metadata = {
    name = "my-plan"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region            = "eastus"
    resource_group    = "my-rg"
    service_plan_name = "my-plan"
    os_type           = "LINUX"
    sku_name          = "PREMIUM_P1V3"
    worker_count      = 3
  }
}
```

The `os_type` and `sku_name` values arrive as the spec enum's name strings;
`locals.tf` maps them to Azure's wire spellings (`Linux`, `P1v3`).

## Outputs

| Output | Description |
|--------|-------------|
| `service_plan_id` | ARM resource ID of the Service Plan |
| `service_plan_name` | Name of the Service Plan |
| `os_type` | Configured OS type in Azure's spelling |
| `sku_name` | Configured SKU name in Azure's spelling |
| `kind` | Azure's computed plan classification |
| `reserved` | Whether the plan runs Linux workers |

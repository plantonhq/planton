# AzureServicePlan Terraform Module

This directory contains the Terraform IaC implementation for the `AzureServicePlan` component.

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
    region         = "eastus"
    resource_group = "my-rg"
    name           = "my-plan"
    os_type        = "Linux"
    sku_name       = "P1v3"
    worker_count   = 3
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `plan_id` | ARM resource ID of the Service Plan |
| `plan_name` | Name of the Service Plan |
| `os_type` | Configured OS type |
| `sku_name` | Configured SKU name |

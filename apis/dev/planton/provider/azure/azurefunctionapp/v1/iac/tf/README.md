# AzureFunctionApp Terraform Module

This directory contains the Terraform IaC implementation for the `AzureFunctionApp` component.

## Structure

```
tf/
├── main.tf          # Linux Function App resource definition
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values (7 outputs matching stack_outputs.proto)
├── locals.tf        # Local computations (tags)
├── provider.tf      # Azure provider configuration
└── README.md        # This file
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Linux Function App | `azurerm_linux_function_app` | Always |

## Usage

```hcl
module "function_app" {
  source = "./path/to/module"

  metadata = {
    name = "my-function-app"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region               = "eastus"
    resource_group       = "my-rg"
    name                 = "my-function-app"
    service_plan_id      = "/subscriptions/.../providers/Microsoft.Web/serverfarms/my-plan"
    storage_account_name = "mystorageaccount"

    site_config = {
      application_stack = {
        python_version = "3.12"
      }
      always_on = true
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `function_app_id` | ARM resource ID of the Function App |
| `default_hostname` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | List of outbound IP addresses |
| `identity_principal_id` | System-assigned identity principal ID (empty if no identity) |
| `identity_tenant_id` | System-assigned identity tenant ID (empty if no identity) |
| `custom_domain_verification_id` | Domain verification ID for custom domain binding |
| `kind` | Resource kind string (e.g., `functionapp,linux`) |

# AzureLinuxWebApp Terraform Module

This directory contains the Terraform IaC implementation for the `AzureLinuxWebApp` component.

## Structure

```
tf/
├── main.tf          # Linux Web App resource definition
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values (7 outputs matching stack_outputs.proto)
├── locals.tf        # Local computations (tags)
├── provider.tf      # Azure provider configuration
└── README.md        # This file
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Linux Web App | `azurerm_linux_web_app` | Always |

## Prerequisites

- **Azure credentials**: Configure via `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID` environment variables
- **App Service Plan**: An existing App Service Plan (the plan ARM ID is a required input)
- **Terraform 1.5+**: Required for module compatibility

## Usage

```hcl
module "web_app" {
  source = "./path/to/module"

  metadata = {
    name = "my-web-app"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "my-rg"
    name           = "my-web-app"
    service_plan_id = "/subscriptions/.../providers/Microsoft.Web/serverfarms/my-plan"

    site_config = {
      application_stack = {
        python_version = "3.12"
      }
      always_on        = true
      health_check_path = "/health"
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `web_app_id` | ARM resource ID of the Web App |
| `default_hostname` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | List of outbound IP addresses |
| `identity_principal_id` | System-assigned identity principal ID (empty if no identity) |
| `identity_tenant_id` | System-assigned identity tenant ID (empty if no identity) |
| `custom_domain_verification_id` | Domain verification ID for custom domain binding |
| `kind` | Resource kind string (e.g., `app,linux`) |

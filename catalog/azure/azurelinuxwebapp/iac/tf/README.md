# AzureLinuxWebApp Terraform Module

This directory contains the Terraform IaC implementation for the `AzureLinuxWebApp` kind.

## Structure

```
tf/
├── main.tf          # Linux Web App resource definition (all blocks)
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values matching outputs.proto
├── locals.tf        # Local computations (tags + enum-to-wire-value maps)
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
    region          = "eastus"
    resource_group  = "my-rg"
    web_app_name    = "my-web-app"
    service_plan_id = "/subscriptions/.../providers/Microsoft.Web/serverfarms/my-plan"

    site_config = {
      application_stack = {
        python_version = "3.12"
      }
      always_on         = true
      health_check_path = "/health"
    }
  }
}
```

Enum-valued fields arrive as the spec enum's name strings (e.g.
`minimum_tls_version = "TLS_1_2"`, `ftps_state = "DISABLED"`,
`identity.type = "SYSTEM_ASSIGNED"`); `locals.tf` maps them to Azure's
wire spellings.

## Outputs

| Output | Description |
|--------|-------------|
| `web_app_id` | ARM resource ID of the Web App |
| `default_hostname` | Default hostname (`{web_app_name}.azurewebsites.net`) |
| `outbound_ip_addresses` | Currently active outbound IP addresses |
| `possible_outbound_ip_addresses` | Every outbound IP the platform could ever use |
| `identity_principal_id` | System-assigned identity principal ID (empty if no identity) |
| `identity_tenant_id` | System-assigned identity tenant ID (empty if no identity) |
| `custom_domain_verification_id` | Domain verification ID for custom domain binding |
| `kind` | Resource kind string (e.g., `app,linux`) |
| `hosting_environment_id` | App Service Environment ARM ID (empty outside ASE) |
| `site_credential_name` | Publishing credential username |
| `site_credential_password` | Publishing credential password (sensitive) |

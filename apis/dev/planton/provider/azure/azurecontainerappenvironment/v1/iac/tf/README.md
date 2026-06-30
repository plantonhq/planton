# AzureContainerAppEnvironment Terraform Module

This directory contains the Terraform IaC implementation for the `AzureContainerAppEnvironment` component.

## Structure

```
tf/
├── main.tf          # Container App Environment resource definition
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values
├── locals.tf        # Local computations (tags, logs_destination)
├── provider.tf      # Azure provider configuration
└── README.md        # This file
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Container App Environment | `azurerm_container_app_environment` | Always |

## Usage

```hcl
module "container_app_env" {
  source = "./path/to/module"

  metadata = {
    name = "my-env"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                     = "eastus"
    resource_group             = "my-rg"
    name                       = "my-apps-env"
    infrastructure_subnet_id   = "/subscriptions/.../subnets/apps"
    log_analytics_workspace_id = "/subscriptions/.../workspaces/law"
    zone_redundancy_enabled    = true
    workload_profiles = [
      {
        name                  = "general"
        workload_profile_type = "D4"
        minimum_count         = 2
        maximum_count         = 8
      }
    ]
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `environment_id` | ARM resource ID of the Container App Environment |
| `default_domain` | Default domain for apps in this environment |
| `static_ip_address` | Static IP address of the environment |
| `platform_reserved_cidr` | Reserved IP range for platform infrastructure |
| `platform_reserved_dns_ip_address` | Internal DNS server IP |

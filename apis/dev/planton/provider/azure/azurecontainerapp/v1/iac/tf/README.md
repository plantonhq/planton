# AzureContainerApp Terraform Module

This directory contains the Terraform IaC implementation for the `AzureContainerApp` component.

## Structure

```
tf/
├── main.tf          # Container App resource definition with all dynamic blocks
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Output values
├── locals.tf        # Local computations (tags)
├── provider.tf      # Azure provider configuration
└── README.md        # This file
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Container App | `azurerm_container_app` | Always |

## Usage

```hcl
module "container_app" {
  source = "./path/to/module"

  metadata = {
    name = "my-app"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    resource_group               = "my-rg"
    name                         = "my-web-app"
    container_app_environment_id = "/subscriptions/.../managedEnvironments/my-env"
    revision_mode                = "Single"

    containers = [
      {
        name   = "web"
        image  = "mcr.microsoft.com/k8se/quickstart:latest"
        cpu    = 0.5
        memory = "1Gi"

        env = [
          { name = "APP_ENV", value = "production" },
          { name = "DB_PASSWORD", secret_name = "db-password" }
        ]

        liveness_probe = {
          transport = "HTTP"
          port      = 8080
          path      = "/healthz"
        }

        readiness_probe = {
          transport = "HTTP"
          port      = 8080
          path      = "/ready"
        }
      }
    ]

    min_replicas = 1
    max_replicas = 10

    http_scale_rules = [
      {
        name                = "http-rule"
        concurrent_requests = "100"
      }
    ]

    secrets = [
      { name = "db-password", value = "supersecret" }
    ]

    ingress = {
      external_enabled = true
      target_port      = 8080

      traffic_weight = [
        {
          latest_revision = true
          percentage      = 100
        }
      ]

      cors_policy = {
        allowed_origins = ["https://example.com"]
        allowed_methods = ["GET", "POST"]
      }
    }

    dapr = {
      app_id   = "my-web-app"
      app_port = 8080
    }

    identity = {
      type = "SystemAssigned"
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `container_app_id` | ARM resource ID of the Container App |
| `latest_revision_name` | Name of the latest Container Revision |
| `latest_revision_fqdn` | FQDN of the latest Container Revision |
| `outbound_ip_addresses` | Outbound IP addresses of the Container App |
| `ingress_fqdn` | Ingress FQDN (empty string if ingress is not configured) |

---
title: "Docker Container"
description: "This preset deploys a containerized Function App running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry password),..."
type: "preset"
rank: "02"
presetSlug: "02-docker-container"
componentSlug: "azurefunctionapp-research-design-documentation"
componentTitle: "AzureFunctionApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 2
---

# Docker Container

This preset deploys a containerized Function App running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry password), always-on mode for zero cold starts, and secure defaults. This is the standard pattern for custom runtimes, multi-language stacks, or pre-built container images.

## When to Use

- Custom runtimes not natively supported by Azure Functions (Rust, Go, etc. via custom handler)
- Pre-built container images from your CI/CD pipeline
- Applications with complex system dependencies packaged in a container
- Teams standardizing on container-based deployments across all compute platforms

## Key Configuration Choices

- **Docker application stack** (`docker`) -- Runs a custom container image instead of a managed runtime
- **ACR managed identity** (`container_registry_use_managed_identity: true`) -- Credential-free image pulls from Azure Container Registry; no registry username/password needed
- **System-assigned identity** (`type: SystemAssigned`) -- Enables managed identity for ACR and other Azure service access; the identity's `principal_id` must have `AcrPull` role on the registry
- **Always on** (`always_on: true`) -- Prevents the app from being unloaded during idle periods; critical for container-based function apps on Dedicated/Premium plans
- **Health check** (`health_check_path: /api/health`) -- Azure monitors this endpoint and removes unhealthy instances from the load balancer
- **FTPS disabled** (`ftps_state: Disabled`) -- No FTP deployment; container images are pulled from the registry
- **Elastic Premium or Dedicated plan required** -- Docker-based function apps cannot run on Consumption (Y1) plans

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-elastic-premium-plan-id>` | ARM ID of an Elastic Premium (EP*) or Dedicated (P*v3) App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-storage-account-name>` | Name of the storage account for Functions runtime | Azure portal or `AzureStorageAccount` status outputs |
| `<your-storage-access-key>` | Access key for the storage account | Azure portal -> Storage Account -> Access keys |
| `<your-registry>.azurecr.io` | Azure Container Registry login server | Azure portal -> Container Registry -> Login server |
| `<your-org>/<your-function-app>` | Container image name (without tag) | Your container registry |
| `image_tag: latest` | Container image tag (use a specific version for production) | Your CI/CD pipeline |

## Related Presets

- **01-python-http-api** -- Use instead for a Python-native function app without container packaging
- **03-enterprise-elastic-premium** -- Use instead for production workloads needing VNet integration, IP restrictions, and full security hardening

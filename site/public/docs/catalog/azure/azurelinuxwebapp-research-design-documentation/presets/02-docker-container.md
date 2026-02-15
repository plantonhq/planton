---
title: "Docker Container"
description: "This preset deploys a containerized web application running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry..."
type: "preset"
rank: "02"
presetSlug: "02-docker-container"
componentSlug: "azurelinuxwebapp-research-design-documentation"
componentTitle: "AzureLinuxWebApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 2
---

# Docker Container

This preset deploys a containerized web application running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry password), always-on mode for zero cold starts, and secure defaults. This is the standard pattern for custom runtimes, polyglot services, or pre-built container images.

## When to Use

- Custom runtimes not natively supported by Azure App Service (Rust, Go 1.22+, Elixir, etc.)
- Pre-built container images from your CI/CD pipeline
- Applications with complex system dependencies packaged in a container
- Teams standardizing on container-based deployments across all compute platforms
- Multi-language services (e.g., a Go backend serving a compiled frontend)

## Key Configuration Choices

- **Docker application stack** (`docker`) -- Runs a custom container image instead of a managed runtime
- **ACR managed identity** (`container_registry_use_managed_identity: true`) -- Credential-free image pulls from Azure Container Registry; no registry username/password needed
- **System-assigned identity** (`type: SystemAssigned`) -- Enables managed identity for ACR and other Azure service access; the identity's `principal_id` must have `AcrPull` role on the registry
- **Always on** (`always_on: true`) -- Prevents the app from being unloaded during idle periods; critical for container-based web apps on Dedicated/Premium plans
- **WEBSITES_PORT** (`WEBSITES_PORT: "8080"`) -- Tells Azure which port the container listens on; change to match your application's HTTP port
- **Health check** (`health_check_path: /health`) -- Azure monitors this endpoint and removes unhealthy instances from the load balancer
- **FTPS disabled** (`ftps_state: Disabled`) -- No FTP deployment; container images are pulled from the registry
- **Premium or Dedicated plan recommended** -- Ensures always-on support and sufficient resources for container workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-premium-plan-id>` | ARM ID of a Premium (P*v3) or Standard (S*) App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-registry>.azurecr.io` | Azure Container Registry login server | Azure portal -> Container Registry -> Login server |
| `<your-org>/<your-web-app>` | Container image name (without tag) | Your container registry |
| `image_tag: latest` | Container image tag (use a specific version for production) | Your CI/CD pipeline |

## Related Presets

- **01-node-web-api** -- Use instead for a Node.js-native web API without container packaging
- **03-enterprise-private-web-app** -- Use instead for production workloads needing VNet integration, IP restrictions, and full security hardening

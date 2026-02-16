---
title: "Python HTTP API"
description: "This preset deploys a Python 3.12 Function App configured for HTTP-triggered APIs with Application Insights monitoring, a health check endpoint, CORS, and secure defaults. It is the most common..."
type: "preset"
rank: "01"
presetSlug: "01-python-http-api"
componentSlug: "function-app"
componentTitle: "Function App"
provider: "azure"
icon: "package"
order: 1
---

# Python HTTP API

This preset deploys a Python 3.12 Function App configured for HTTP-triggered APIs with Application Insights monitoring, a health check endpoint, CORS, and secure defaults. It is the most common starting point for Python serverless APIs on Azure Functions.

## When to Use

- Python REST APIs or webhook handlers triggered by HTTP requests
- Lightweight backend services that benefit from serverless scaling
- Applications that need Application Insights telemetry for monitoring and diagnostics
- APIs serving a web frontend that requires CORS configuration

## Key Configuration Choices

- **Python 3.12** (`python_version: "3.12"`) -- Current LTS version; best balance of performance and ecosystem support
- **Application Insights** (`application_insights_connection_string`) -- Automatic telemetry for requests, dependencies, exceptions, and traces
- **Health check** (`health_check_path: /api/health`) -- Azure monitors this endpoint and removes unhealthy instances from the load balancer
- **HTTPS only** (`https_only: true`) -- All HTTP requests are redirected to HTTPS
- **FTPS disabled** (`ftps_state: Disabled`) -- No FTP file deployment; use CI/CD pipelines instead
- **TLS 1.2 minimum** (`minimum_tls_version: "1.2"`) -- Industry-standard minimum TLS version
- **CORS** (`allowed_origins`) -- Single allowed origin; replace with your frontend domain
- **Worker extensions** (`PYTHON_ENABLE_WORKER_EXTENSIONS: "1"`) -- Enables Python worker extensions for Application Insights SDK integration

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-service-plan-id>` | ARM ID of the App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-storage-account-name>` | Name of the storage account for Functions runtime | Azure portal or `AzureStorageAccount` status outputs |
| `<your-storage-access-key>` | Access key for the storage account | Azure portal -> Storage Account -> Access keys |
| `<your-app-insights-connection-string>` | Application Insights connection string | Azure portal or `AzureApplicationInsights` status outputs |
| `allowed_origins: https://myapp.example.com` | Your frontend domain for CORS | Your application domain |

## Related Presets

- **02-docker-container** -- Use instead for custom container images or runtimes not natively supported
- **03-enterprise-elastic-premium** -- Use instead for production workloads needing VNet integration, managed identity, and pre-warmed instances

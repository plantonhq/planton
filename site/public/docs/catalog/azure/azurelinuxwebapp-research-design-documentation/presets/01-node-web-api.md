---
title: "Node.js Web API"
description: "This preset deploys a Node.js 22 LTS web API with health check monitoring, HTTP/2 for improved performance, CORS for cross-origin requests, and Application Insights telemetry. It is the standard..."
type: "preset"
rank: "01"
presetSlug: "01-node-web-api"
componentSlug: "azurelinuxwebapp-research-design-documentation"
componentTitle: "AzureLinuxWebApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 1
---

# Node.js Web API

This preset deploys a Node.js 22 LTS web API with health check monitoring, HTTP/2 for improved performance, CORS for cross-origin requests, and Application Insights telemetry. It is the standard starting point for Node.js REST APIs, GraphQL endpoints, and web services on Azure App Service.

## When to Use

- Node.js REST APIs or GraphQL endpoints serving web and mobile clients
- Express, Fastify, or Hapi web services
- Next.js API routes running as a standalone server
- APIs that serve a web frontend requiring CORS configuration

## Key Configuration Choices

- **Node.js 22 LTS** (`node_version: "22-lts"`) -- Current Long-Term Support version; best balance of performance, stability, and ecosystem support
- **Application Insights** (`application_insights_connection_string`) -- Automatic telemetry for HTTP requests, dependencies, exceptions, and traces
- **Health check** (`health_check_path: /health`) -- Azure monitors this endpoint and removes unhealthy instances from the load balancer
- **HTTP/2** (`http2_enabled: true`) -- Multiplexing and header compression for improved API latency and throughput
- **HTTPS only** (`https_only: true`) -- All HTTP requests are redirected to HTTPS
- **FTPS disabled** (`ftps_state: Disabled`) -- No FTP file deployment; use CI/CD pipelines instead
- **TLS 1.2 minimum** (`minimum_tls_version: "1.2"`) -- Industry-standard minimum TLS version
- **CORS** (`allowed_origins`) -- Single allowed origin; replace with your frontend domain
- **Production mode** (`NODE_ENV: production`) -- Enables production optimizations in Node.js frameworks

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-service-plan-id>` | ARM ID of the App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-app-insights-connection-string>` | Application Insights connection string | Azure portal or `AzureApplicationInsights` status outputs |
| `allowed_origins: https://myapp.example.com` | Your frontend domain for CORS | Your application domain |

## Related Presets

- **02-docker-container** -- Use instead for custom container images or runtimes not natively supported
- **03-enterprise-private-web-app** -- Use instead for production workloads needing VNet integration, managed identity, and IP restrictions

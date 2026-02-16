---
title: "Enterprise Private Web App"
description: "This preset deploys a production-grade Python web application on a Premium App Service Plan with VNet integration, IP restrictions (default deny), Application Insights monitoring, Key Vault secret..."
type: "preset"
rank: "03"
presetSlug: "03-enterprise-private-web-app"
componentSlug: "linux-web-app"
componentTitle: "Linux Web App"
provider: "azure"
icon: "package"
order: 3
---

# Enterprise Private Web App

This preset deploys a production-grade Python web application on a Premium App Service Plan with VNet integration, IP restrictions (default deny), Application Insights monitoring, Key Vault secret references, diagnostic logging, and full security hardening. This is the standard pattern for enterprise web applications that require network isolation, compliance controls, and credential-free authentication.

## When to Use

- Production web applications that must comply with enterprise security policies (VNet isolation, IP restrictions, TLS 1.2+)
- Workloads that need VNet integration to access private databases, Redis, or other VNet-connected services
- Services using Key Vault references for secrets management (no plain-text credentials in manifests or app_settings)
- Applications requiring comprehensive logging for auditing and troubleshooting (application logs, HTTP logs, failed request tracing)
- Multi-instance deployments with health check monitoring for high availability

## Key Configuration Choices

- **Premium plan** -- Dedicated compute with VNet integration, always-on, and enhanced performance
- **Multiple workers** (`worker_count: 3`) -- Three instances for high availability and load distribution
- **Always on** (`always_on: true`) -- Prevents cold starts; critical for production workloads
- **VNet integration** (`virtual_network_subnet_id` + `vnet_route_all_enabled: true`) -- All outbound traffic routes through the VNet for network inspection and private resource access
- **IP restrictions with default deny** -- Only corporate office and VPN CIDRs can access the Web App; all other traffic is denied
- **SCM mirrors main restrictions** (`scm_use_main_ip_restriction: true`) -- Kudu/SCM endpoint uses the same IP restrictions as the main site
- **Key Vault references** (`@Microsoft.KeyVault(...)` in `app_settings`) -- Secrets are fetched from Key Vault at runtime using the managed identity
- **System-assigned identity** (`type: SystemAssigned`) -- Credential-free access to Key Vault and other Azure services
- **Comprehensive logging** -- Application logs at Information level, HTTP logs with 7-day retention, and failed request tracing for diagnostics
- **Health check with eviction** (`health_check_eviction_time_in_min: 5`) -- Unhealthy instances are evicted after 5 minutes of continuous failure
- **HTTP/2** (`http2_enabled: true`) -- Multiplexing and header compression for improved performance
- **CORS with credentials** -- Single allowed origin with `support_credentials: true` for cookie-based auth

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-premium-plan-id>` | ARM ID of a Premium (P1v3/P2v3/P3v3) App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-app-insights-connection-string>` | Application Insights connection string | Azure portal or `AzureApplicationInsights` status outputs |
| `<your-vnet-subnet-id>` | ARM ID of the subnet delegated to Microsoft.Web/serverFarms | Azure portal or `AzureSubnet` status outputs (`subnet_id`) |
| `<your-keyvault>.vault.azure.net` | Key Vault hostname for secret references | Azure portal -> Key Vault -> Overview |
| `ip_address: 203.0.113.0/24` | Your corporate office CIDR | Network administrator |
| `ip_address: 198.51.100.0/24` | Your VPN gateway CIDR | Network administrator |
| `allowed_origins: https://portal.example.com` | Your frontend domain for CORS | Your application domain |

## Related Presets

- **01-node-web-api** -- Use instead for simpler Node.js APIs without VNet or managed identity requirements
- **02-docker-container** -- Use instead for custom container images with ACR-based deployments

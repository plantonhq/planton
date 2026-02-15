---
title: "Enterprise Elastic Premium"
description: "This preset deploys a production-grade Function App on an Elastic Premium plan with VNet integration, managed identity for storage (no access keys), Key Vault secret references, pre-warmed instances..."
type: "preset"
rank: "03"
presetSlug: "03-enterprise-elastic-premium"
componentSlug: "azurefunctionapp-research-design-documentation"
componentTitle: "AzureFunctionApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 3
---

# Enterprise Elastic Premium

This preset deploys a production-grade Function App on an Elastic Premium plan with VNet integration, managed identity for storage (no access keys), Key Vault secret references, pre-warmed instances for zero cold starts, IP restrictions, runtime scale monitoring, and full security hardening. This is the standard pattern for enterprise serverless workloads that require network isolation, credential-free authentication, and production resilience.

## When to Use

- Production Function Apps that must comply with enterprise security policies (VNet isolation, IP restrictions, TLS 1.2+)
- Latency-sensitive APIs or event processors that cannot tolerate cold starts
- Workloads that need VNet integration to access private databases, Redis, or other VNet-connected services
- Services using Key Vault references for secrets management (no plain-text credentials in manifests or app_settings)
- High-throughput event processors that benefit from runtime scale monitoring and elastic scale-out

## Key Configuration Choices

- **Elastic Premium plan** -- Pre-warmed instances eliminate cold starts; elastic scale-out handles traffic spikes
- **Pre-warmed instances** (`elastic_instance_minimum: 2`, `pre_warmed_instance_count: 3`) -- 2 always-running instances plus 3 warm standby; total of 5 instances ready to handle requests instantly
- **Scale limit** (`app_scale_limit: 30`) -- Caps elastic scale-out at 30 instances to control costs
- **Managed identity storage** (`storage_uses_managed_identity: true`) -- No storage access key needed; the Function App's system-assigned identity must have Storage Blob Data Owner and Storage Queue Data Contributor roles
- **VNet integration** (`virtual_network_subnet_id` + `vnet_route_all_enabled: true`) -- All outbound traffic routes through the VNet for network inspection and private resource access
- **IP restrictions** -- Only corporate office and VPN CIDRs can access the Function App; default action is Deny
- **SCM mirrors main restrictions** (`scm_use_main_ip_restriction: true`) -- Kudu/SCM endpoint uses the same IP restrictions as the main site
- **Key Vault references** (`@Microsoft.KeyVault(...)` in `app_settings`) -- Secrets are fetched from Key Vault at runtime using the managed identity
- **Builtin logging disabled** (`builtin_logging_enabled: false`) -- Application Insights provides telemetry; disabling AzureWebJobsDashboard avoids duplicate logging and storage costs
- **Runtime scale monitoring** (`runtime_scale_monitoring_enabled: true`) -- Functions runtime directly monitors event sources for more accurate scaling decisions
- **HTTP/2** (`http2_enabled: true`) -- Multiplexing and header compression for improved API performance
- **CORS with credentials** -- Single allowed origin with `support_credentials: true` for cookie-based auth

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the Azure resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-elastic-premium-plan-id>` | ARM ID of an Elastic Premium (EP1/EP2/EP3) App Service Plan | Azure portal or `AzureServicePlan` status outputs (`plan_id`) |
| `<your-storage-account-name>` | Name of the storage account for Functions runtime | Azure portal or `AzureStorageAccount` status outputs |
| `<your-app-insights-connection-string>` | Application Insights connection string | Azure portal or `AzureApplicationInsights` status outputs |
| `<your-vnet-subnet-id>` | ARM ID of the subnet delegated to Microsoft.Web/serverFarms | Azure portal or `AzureSubnet` status outputs (`subnet_id`) |
| `<your-keyvault>.vault.azure.net` | Key Vault hostname for secret references | Azure portal -> Key Vault -> Overview |
| `ip_address: 203.0.113.0/24` | Your corporate office CIDR | Network administrator |
| `ip_address: 198.51.100.0/24` | Your VPN gateway CIDR | Network administrator |
| `allowed_origins: https://portal.example.com` | Your frontend domain for CORS | Your application domain |

## Related Presets

- **01-python-http-api** -- Use instead for simpler Python APIs without VNet or managed identity requirements
- **02-docker-container** -- Use instead for custom container images with ACR-based deployments

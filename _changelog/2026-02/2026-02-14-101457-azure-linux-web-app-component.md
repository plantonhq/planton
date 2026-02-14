# AzureLinuxWebApp: Full-Feature Web Hosting Component

**Date**: February 14, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi IaC, Terraform IaC, Presets

## Summary

Added AzureLinuxWebApp (R20) as a full-feature OpenMCF deployment component for Azure Linux Web Apps (`azurerm_linux_web_app`). This is the web hosting sibling of AzureFunctionApp (R19), providing a complete declarative interface for deploying web applications on Azure App Service (Linux) with 13 message types, ~100 fields, 7 outputs, 106 passing tests, and dual IaC implementations (Pulumi + Terraform).

## Problem Statement / Motivation

Azure App Service is one of the most widely used web hosting platforms, but the original T02 plan spec was significantly underspecified -- only 10 fields and 3 message types versus the ~50+ fields and 10+ nested blocks in the actual Terraform provider. A production-grade platform needs comprehensive coverage of the real provider surface, not a skeleton.

### Pain Points

- T02 spec was missing `resource_group` (StringValueOrRef), `region`, and 14 other critical fields
- Docker configuration was oversimplified (2 fields vs 4+ needed including registry auth)
- Java application stack was missing `java_server_version` (RequiredWith in provider)
- SiteConfig had only 3 fields instead of the 22 needed for production deployments
- No diagnostic logging support (logs block)
- No web-app-specific features (client affinity, health check eviction)

## Solution / What's New

Full-feature AzureLinuxWebApp with 17 corrections from the T02 plan spec, thoroughly researched against the Terraform provider source (`linux_web_app_resource.go`).

### Key Features

- **13 message types** covering the complete resource surface
- **Richer application stack** than FunctionApp: .NET, Node.js, Python, PHP, Ruby, Go, Java (server/version/server_version), Docker
- **Top-level logs block** with application logs (file system level), HTTP logs (retention), failed request tracing, detailed error messages
- **Web-app-specific fields**: `client_affinity_enabled` (ARR session stickiness), `enabled` toggle, `health_check_eviction_time_in_min`
- **Full security surface**: client certificates (mTLS), IP restrictions with headers, TLS version control, FTPS state
- **Opinionated secure defaults**: `https_only=true`, `use_32_bit_worker=false`, `ftps_state="Disabled"`, `minimum_tls_version="1.2"`
- **Application Insights integration**: via app_settings injection (`APPLICATIONINSIGHTS_CONNECTION_STRING`), not a site_config field

### 17 Corrections from T02 Spec

1. Added `resource_group` (StringValueOrRef) -- per DD05
2. Added `region` (string, required) -- per established pattern
3. Restructured Docker as block (4+ fields, not 2)
4. Added `java_server_version` to application stack
5. Expanded SiteConfig from 3 to 22 fields
6. Removed `auto_heal_enabled` (useless without trigger config)
7. Changed `connection_strings[].value` to StringValueOrRef
8. Changed identity to string+CEL (not proto enum)
9. Added `public_network_access_enabled`
10. Added `key_vault_reference_identity_id` (StringValueOrRef)
11. Added `storage_mounts` (repeated)
12. Added `logs` block (simplified, file system only)
13. Added `client_affinity_enabled`
14. Added `client_certificate_enabled` + `client_certificate_mode`
15. Added `enabled` toggle
16. Enriched outputs from 4 to 7
17. Added name validation CEL

## Implementation Details

### File Structure (38 files)

```
apis/org/openmcf/provider/azure/azurelinuxwebapp/v1/
├── spec.proto              # 13 message types, ~100 fields
├── stack_outputs.proto     # 7 outputs
├── api.proto               # KRM wiring
├── stack_input.proto       # IaC module input
├── spec_test.go            # 106 validation tests
├── *.pb.go                 # Generated proto stubs
├── BUILD.bazel             # Gazelle-managed
├── README.md               # User-facing overview
├── examples.md             # 6 YAML examples
├── docs/README.md          # Comprehensive research
├── iac/
│   ├── hack/manifest.yaml  # Test manifest
│   ├── pulumi/
│   │   ├── module/main.go  # Resource creation (LinuxWebApp)
│   │   ├── module/locals.go
│   │   ├── module/outputs.go
│   │   ├── main.go         # Entrypoint
│   │   └── ...
│   └── tf/
│       ├── main.tf         # azurerm_linux_web_app
│       ├── variables.tf
│       ├── locals.tf
│       ├── outputs.tf
│       └── provider.tf
└── presets/
    ├── 01-node-web-api.*
    ├── 02-docker-container.*
    └── 03-enterprise-private-web-app.*
```

### Key Design Decisions

- **Application Insights injection**: Unlike FunctionApp which has `siteConfig.ApplicationInsightsConnectionString`, LinuxWebApp doesn't expose this field. Instead, we merge `APPLICATIONINSIGHTS_CONNECTION_STRING` into `app_settings` in both Pulumi and Terraform modules.
- **Logs as top-level block**: FunctionApp has `app_service_logs` inside `site_config`. WebApp has a richer top-level `logs` block with application logs, HTTP logs, failed request tracing, and detailed error messages. Blob storage variants deferred to v2.
- **Opinionated defaults**: `use_32_bit_worker` defaults to `false` (provider default is `true`) because 64-bit is the right choice for modern apps.

## Benefits

- **Complete web hosting coverage** for Azure: ServicePlan + FunctionApp + LinuxWebApp now covers the full App Service stack
- **106 validation tests** ensuring proto constraints match Azure API reality
- **Dual IaC** with feature parity between Pulumi and Terraform
- **Infra-chart ready** with 8 StringValueOrRef fields for dependency wiring
- **3 presets** for common deployment patterns (Node.js API, Docker container, Enterprise private)

## Impact

- **Users**: Can now deploy Azure Linux Web Apps declaratively with the same consistency as all other OpenMCF resources
- **Infra charts**: The `web-app-environment` chart can now be built (T03 phase)
- **Platform completeness**: App hosting category is now complete with 5 resources (ServicePlan, ContainerAppEnvironment, ContainerApp, FunctionApp, LinuxWebApp)

## Related Work

- R19 AzureFunctionApp -- sibling component, shares ~80% of structure
- R16 AzureServicePlan -- required dependency (provides compute tier)
- DD04 -- Linux-only decision (Windows variants excluded)
- DD05 -- AzureResourceGroup as first-class resource
- T03 -- Infra chart phase (web-app-environment chart, pending)

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)

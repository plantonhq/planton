# AzureFunctionApp Pulumi Module: Architecture Overview

## Resource Graph

The AzureFunctionApp module creates a single resource with complex nested configuration:

```
AzureFunctionApp
└── appservice.LinuxFunctionApp (azurerm_linux_function_app)
    ├── site_config
    │   ├── application_stack (runtime: dotnet/node/python/java/powershell/docker/custom)
    │   ├── cors
    │   ├── ip_restrictions[]
    │   ├── scm_ip_restrictions[]
    │   └── app_service_logs
    ├── connection_strings[]
    ├── storage_accounts[] (mounts)
    └── identity (SystemAssigned / UserAssigned / both)
```

## Data Flow

```
AzureFunctionAppStackInput
├── target.metadata          → Azure tags (resource, resource_name, resource_kind, org, env)
├── target.spec.region       → LinuxFunctionAppArgs.Location
├── target.spec.resource_group → locals.ResourceGroupName (via .GetValue())
├── target.spec.name         → LinuxFunctionAppArgs.Name
├── target.spec.service_plan_id → LinuxFunctionAppArgs.ServicePlanId (via .GetValue())
├── target.spec.storage_account_name → LinuxFunctionAppArgs.StorageAccountName (via .GetValue())
├── target.spec.storage_account_access_key → LinuxFunctionAppArgs.StorageAccountAccessKey (optional)
├── target.spec.storage_uses_managed_identity → LinuxFunctionAppArgs.StorageUsesManagedIdentity (optional)
├── target.spec.functions_extension_version → LinuxFunctionAppArgs.FunctionsExtensionVersion (optional, default "~4")
├── target.spec.site_config  → LinuxFunctionAppArgs.SiteConfig (required, complex nested)
├── target.spec.app_settings → LinuxFunctionAppArgs.AppSettings (optional map)
├── target.spec.connection_strings → LinuxFunctionAppArgs.ConnectionStrings (optional repeated)
├── target.spec.application_insights_connection_string → SiteConfig.ApplicationInsightsConnectionString
├── target.spec.https_only   → LinuxFunctionAppArgs.HttpsOnly (optional, default true)
├── target.spec.public_network_access_enabled → LinuxFunctionAppArgs.PublicNetworkAccessEnabled (optional)
├── target.spec.builtin_logging_enabled → LinuxFunctionAppArgs.BuiltinLoggingEnabled (optional)
├── target.spec.virtual_network_subnet_id → LinuxFunctionAppArgs.VirtualNetworkSubnetId (optional)
├── target.spec.identity     → LinuxFunctionAppArgs.Identity (optional)
├── target.spec.key_vault_reference_identity_id → LinuxFunctionAppArgs.KeyVaultReferenceIdentityId (optional)
├── target.spec.client_certificate_enabled → LinuxFunctionAppArgs.ClientCertificateEnabled (optional)
├── target.spec.client_certificate_mode → LinuxFunctionAppArgs.ClientCertificateMode (optional)
├── target.spec.client_certificate_exclusion_paths → LinuxFunctionAppArgs.ClientCertificateExclusionPaths (optional)
├── target.spec.content_share_force_disabled → LinuxFunctionAppArgs.ContentShareForceDisabled (optional)
└── target.spec.storage_mounts → LinuxFunctionAppArgs.StorageAccounts (optional repeated)
```

## Output Wiring

```
LinuxFunctionApp.ID()                    → function_app_id
LinuxFunctionApp.DefaultHostname         → default_hostname
LinuxFunctionApp.OutboundIpAddresses     → outbound_ip_addresses
LinuxFunctionApp.Identity.PrincipalId    → identity_principal_id (conditional)
LinuxFunctionApp.Identity.TenantId       → identity_tenant_id (conditional)
LinuxFunctionApp.CustomDomainVerificationId → custom_domain_verification_id
LinuxFunctionApp.Kind                    → kind
```

## Design Notes

- **Single resource, complex config**: Unlike simpler resources (e.g., AzureServicePlan), the
  Function App has deeply nested configuration (site_config → application_stack, cors,
  ip_restrictions, etc.). The module uses helper builder functions for each nested structure.

- **Application Insights wiring**: The connection string is set on `site_config` (not as a
  top-level arg) because that's where the Pulumi provider expects it. The proto puts
  `application_insights_connection_string` on the parent spec for ergonomics, and the module
  routes it to `SiteConfig.ApplicationInsightsConnectionString`.

- **Storage auth pattern**: Storage can be authenticated via access key or managed identity.
  The module checks which field is set and configures accordingly. These are mutually exclusive
  in the Azure API.

- **Identity output handling**: Identity outputs (principal_id, tenant_id) are only populated
  when the Function App has a system-assigned identity. The module uses `ApplyT` to safely
  extract these values, returning empty strings when identity is not configured.

- **Docker configuration**: The application stack's Docker config maps to the `Dockers` array
  field on the Pulumi provider (plural, though only one entry is used). Registry credentials
  are optional when using managed identity for ACR.

- **IP restrictions**: Main site and SCM site have separate IP restriction types in Pulumi
  (`LinuxFunctionAppSiteConfigIpRestrictionArgs` vs `LinuxFunctionAppSiteConfigScmIpRestrictionArgs`),
  requiring separate builder functions despite identical proto structures.

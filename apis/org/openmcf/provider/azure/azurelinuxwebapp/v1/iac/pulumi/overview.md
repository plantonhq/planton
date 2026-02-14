# AzureLinuxWebApp Pulumi Module: Architecture Overview

## Resource Graph

The AzureLinuxWebApp module creates a single resource with complex nested configuration:

```
AzureLinuxWebApp
└── appservice.LinuxWebApp (azurerm_linux_web_app)
    ├── site_config
    │   ├── application_stack (runtime: dotnet/node/python/php/ruby/go/java/docker)
    │   ├── cors
    │   ├── ip_restrictions[]
    │   └── scm_ip_restrictions[]
    ├── connection_strings[]
    ├── storage_accounts[] (mounts)
    ├── logs
    │   ├── application_logs
    │   └── http_logs
    └── identity (SystemAssigned / UserAssigned / both)
```

## Data Flow

```
AzureLinuxWebAppStackInput
├── target.metadata          → Azure tags (resource, resource_name, resource_kind, org, env)
├── target.spec.region       → LinuxWebAppArgs.Location
├── target.spec.resource_group → locals.ResourceGroupName (via .GetValue())
├── target.spec.name         → LinuxWebAppArgs.Name
├── target.spec.service_plan_id → LinuxWebAppArgs.ServicePlanId (via .GetValue())
├── target.spec.site_config  → LinuxWebAppArgs.SiteConfig (required, complex nested)
│   ├── .application_stack   → SiteConfig.ApplicationStack (runtime selection)
│   ├── .always_on           → SiteConfig.AlwaysOn (optional)
│   ├── .app_command_line    → SiteConfig.AppCommandLine (optional)
│   ├── .health_check_path   → SiteConfig.HealthCheckPath (optional)
│   ├── .minimum_tls_version → SiteConfig.MinimumTlsVersion (default "1.2")
│   ├── .ftps_state          → SiteConfig.FtpsState (default "Disabled")
│   ├── .worker_count        → SiteConfig.WorkerCount (optional)
│   ├── .http2_enabled       → SiteConfig.Http2Enabled (optional)
│   ├── .websockets_enabled  → SiteConfig.WebsocketsEnabled (optional)
│   ├── .use_32_bit_worker   → SiteConfig.Use32BitWorker (default false)
│   ├── .vnet_route_all_enabled → SiteConfig.VnetRouteAllEnabled (optional)
│   ├── .load_balancing_mode → SiteConfig.LoadBalancingMode (default "LeastRequests")
│   ├── .cors                → SiteConfig.Cors (optional)
│   ├── .ip_restrictions     → SiteConfig.IpRestrictions (optional)
│   └── .container_registry_use_managed_identity → SiteConfig.ContainerRegistryUseManagedIdentity
├── target.spec.app_settings → LinuxWebAppArgs.AppSettings (optional map)
├── target.spec.connection_strings → LinuxWebAppArgs.ConnectionStrings (optional repeated)
├── target.spec.application_insights_connection_string → AppSettings["APPLICATIONINSIGHTS_CONNECTION_STRING"]
├── target.spec.https_only   → LinuxWebAppArgs.HttpsOnly (optional, default true)
├── target.spec.public_network_access_enabled → LinuxWebAppArgs.PublicNetworkAccessEnabled (optional)
├── target.spec.enabled      → LinuxWebAppArgs.Enabled (optional, default true)
├── target.spec.virtual_network_subnet_id → LinuxWebAppArgs.VirtualNetworkSubnetId (optional)
├── target.spec.identity     → LinuxWebAppArgs.Identity (optional)
├── target.spec.key_vault_reference_identity_id → LinuxWebAppArgs.KeyVaultReferenceIdentityId (optional)
├── target.spec.client_affinity_enabled → LinuxWebAppArgs.ClientAffinityEnabled (optional)
├── target.spec.client_certificate_enabled → LinuxWebAppArgs.ClientCertificateEnabled (optional)
├── target.spec.client_certificate_mode → LinuxWebAppArgs.ClientCertificateMode (optional)
├── target.spec.client_certificate_exclusion_paths → LinuxWebAppArgs.ClientCertificateExclusionPaths (optional)
├── target.spec.storage_mounts → LinuxWebAppArgs.StorageAccounts (optional repeated)
└── target.spec.logs         → LinuxWebAppArgs.Logs (optional)
    ├── .application_logs    → Logs.ApplicationLogs
    ├── .http_logs           → Logs.HttpLogs
    ├── .failed_request_tracing → Logs.FailedRequestTracing
    └── .detailed_error_messages → Logs.DetailedErrorMessages
```

## Output Wiring

```
LinuxWebApp.ID()                         → web_app_id
LinuxWebApp.DefaultHostname              → default_hostname
LinuxWebApp.OutboundIpAddresses          → outbound_ip_addresses
LinuxWebApp.Identity.PrincipalId         → identity_principal_id (conditional)
LinuxWebApp.Identity.TenantId            → identity_tenant_id (conditional)
LinuxWebApp.CustomDomainVerificationId   → custom_domain_verification_id
LinuxWebApp.Kind                         → kind
```

## Design Notes

- **Single resource, complex config**: The Web App has deeply nested configuration
  (site_config → application_stack, cors, ip_restrictions, etc.) plus top-level logs.
  The module uses helper builder functions for each nested structure.

- **Application Insights wiring**: Unlike Function Apps where the connection string is
  set on `SiteConfig.ApplicationInsightsConnectionString`, Web Apps inject it as an
  app_setting (`APPLICATIONINSIGHTS_CONNECTION_STRING`). The module merges user-provided
  `app_settings` with the Application Insights connection string.

- **Logs block handling**: Web App logs are a top-level block (not nested in site_config
  like Function Apps). The module conditionally builds the `Logs` argument only when
  the proto `logs` field is populated.

- **Docker configuration**: The application stack's Docker config maps to the `Docker`
  nested type on the Pulumi provider. Registry credentials are optional when using
  managed identity for ACR.

- **Identity output handling**: Identity outputs (principal_id, tenant_id) are only
  populated when the Web App has a system-assigned identity. The module uses `ApplyT`
  to safely extract these values, returning empty strings when identity is not configured.

- **IP restrictions**: Main site and SCM site have separate IP restriction types in Pulumi
  (`LinuxWebAppSiteConfigIpRestrictionArgs` vs `LinuxWebAppSiteConfigScmIpRestrictionArgs`),
  requiring separate builder functions despite identical proto structures.

- **Java triple**: When `java_version` is set, `java_server` and `java_server_version`
  must also be provided. The module validates this constraint and maps all three fields
  to the application stack configuration.

# AzureFunctionApp: Research & Design Documentation

## Executive Summary

Azure Linux Function App (`Microsoft.Web/sites` kind `functionapp,linux`) is Azure's serverless compute platform for event-driven workloads. It hosts functions triggered by HTTP requests, queue messages, timer schedules, blob storage events, and dozens of other Azure service events. Unlike AWS Lambda (which abstracts compute entirely), Azure Functions run on an explicit App Service Plan, giving operators control over cost model, scale behavior, and networking.

This document captures the research, design rationale, and 80/20 scoping decisions behind the `AzureFunctionApp` OpenMCF component.

## Deployment Landscape

### Azure Functions vs AWS Lambda vs GCP Cloud Functions

| Dimension | Azure Functions | AWS Lambda | GCP Cloud Functions |
|-----------|----------------|------------|---------------------|
| **Compute model** | App Service Plan (explicit) | Fully managed (no plan) | Fully managed (no plan) |
| **Cold start control** | Elastic Premium pre-warming | Provisioned concurrency | Min instances |
| **Max execution** | 5 min (Consumption), unlimited (Premium/Dedicated) | 15 min | 60 min (2nd gen) |
| **Scale ceiling** | 200 (Consumption), 100 (EP), plan-defined (Dedicated) | 1000 concurrent (soft) | 3000 concurrent (soft) |
| **Container support** | Yes (Docker on Premium/Dedicated) | Yes (container images) | Yes (2nd gen) |
| **VNet integration** | Premium/Dedicated only | Yes (VPC) | Yes (VPC connector) |
| **Language runtimes** | .NET, Node, Python, Java, PowerShell, Custom | .NET, Node, Python, Java, Ruby, Go, Custom | .NET, Node, Python, Java, Go, Ruby, PHP |
| **Pricing** | Consumption (pay-per-exec), Premium (per-instance), Dedicated (per-instance) | Pay-per-request + per-GB-s | Pay-per-invocation + per-GB-s |

**Key Azure differentiator**: The explicit App Service Plan provides a spectrum from fully serverless (Consumption Y1) to dedicated VM instances (P3v3). This makes Azure Functions suitable for workloads that don't fit the pure serverless model (long-running processes, VNet-integrated backends, Docker containers).

### Azure Functions Architecture

```
                    ┌──────────────────────────────────┐
                    │         App Service Plan          │
                    │   (Consumption / EP / Dedicated)  │
                    └────────────┬─────────────────────┘
                                 │
                    ┌────────────▼─────────────────────┐
                    │       Linux Function App          │
                    │  ┌─────────────────────────────┐  │
                    │  │   Application Stack          │  │
                    │  │   (Python / Node / .NET /    │  │
                    │  │    Java / Docker / Custom)   │  │
                    │  └─────────────────────────────┘  │
                    │  ┌─────────────────────────────┐  │
                    │  │   Functions Runtime (~4)     │  │
                    │  │   - Trigger bindings         │  │
                    │  │   - Input/output bindings    │  │
                    │  │   - Extension bundles        │  │
                    │  └─────────────────────────────┘  │
                    └────────────┬─────────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
     ┌────────▼────────┐  ┌─────▼──────┐  ┌───────▼────────┐
     │ Storage Account │  │ App Insights│  │  VNet / Subnet │
     │  (required)     │  │ (optional)  │  │  (optional)    │
     └─────────────────┘  └────────────┘  └────────────────┘
```

## Runtime Comparison

### Consumption (Y1)

The Consumption plan is Azure's fully serverless tier:

| Property | Value |
|----------|-------|
| **Cost model** | Per-execution + per-GB-second |
| **Idle cost** | $0 (scales to 0 workers) |
| **Scale range** | 0 to 200 instances (automatic) |
| **Cold start** | Yes (seconds to tens of seconds) |
| **Max execution** | 5 minutes (configurable to 10 min) |
| **VNet integration** | Not supported |
| **Private endpoints** | Not supported |
| **Always On** | Not supported |
| **Best for** | Infrequent, bursty workloads; dev/test; cost-sensitive |

Consumption plans are managed by Azure -- you don't control instance count or pre-warming. The `app_scale_limit` field on `site_config` caps the maximum to control costs.

### Elastic Premium (EP1-EP3)

Elastic Premium combines serverless elasticity with enterprise features:

| Property | Value |
|----------|-------|
| **Cost model** | Per-instance per-hour (pre-warmed) + elastic overage |
| **Idle cost** | Min 1 instance always running |
| **Scale range** | 1 to 100 instances (elastic) |
| **Cold start** | Eliminated by pre-warmed instances |
| **Max execution** | Unlimited |
| **VNet integration** | Supported |
| **Private endpoints** | Supported |
| **Always On** | Automatic |
| **Best for** | Production APIs, latency-sensitive, VNet-integrated |

Key fields for EP plans:
- `elastic_instance_minimum`: Minimum always-running instances
- `pre_warmed_instance_count`: Additional warm standby instances
- `app_scale_limit`: Maximum elastic scale-out

### Dedicated (B1-P3v3)

Dedicated plans run Function Apps on fixed App Service VMs:

| Property | Value |
|----------|-------|
| **Cost model** | Per-instance per-hour (same as Web Apps) |
| **Idle cost** | Full instance cost (always running) |
| **Scale range** | Plan's `worker_count` (manual or auto-scale) |
| **Cold start** | None (with `always_on: true`) |
| **Max execution** | Unlimited |
| **VNet integration** | Supported (Standard and above) |
| **Private endpoints** | Supported |
| **Always On** | Must be explicitly set to `true` |
| **Best for** | Shared compute with Web Apps, long-running jobs, predictable cost |

Critical: On Dedicated plans, `always_on` must be `true` in `site_config`, otherwise Azure may unload the Function App after idle periods.

## Storage Requirement Deep-Dive

Every Azure Function App requires a Storage Account. This is non-optional and architectural:

### What Storage Is Used For

1. **Trigger state**: Blob triggers use lease blobs; queue triggers use poison queue metadata
2. **Execution logs**: The `azure-webjobs-hosts` container stores function execution history
3. **Durable Functions**: Orchestration state, history tables, and work-item queues
4. **Content share**: Function App code is stored in an Azure File Share (unless `content_share_force_disabled`)
5. **Key management**: Internal cryptographic keys for trigger webhooks

### Authentication Options

| Method | Field | Security | Notes |
|--------|-------|----------|-------|
| Access key | `storage_account_access_key` | Shared secret | Simple but requires key rotation |
| Managed identity | `storage_uses_managed_identity: true` | Credential-free | Requires RBAC: Storage Blob Data Owner + Storage Queue Data Contributor |

The managed identity approach is recommended for production:
- No secrets in manifests or environment variables
- Automatic credential rotation by Azure AD
- Auditable via Azure Activity Log

### Storage Account Requirements

- **Performance**: Standard (HDD) is sufficient for most workloads; Premium (SSD) for high-throughput triggers
- **Replication**: LRS is sufficient (Functions don't need geo-redundancy for runtime state)
- **Networking**: Must be accessible from the Function App (same VNet or public endpoint)
- **Account kind**: StorageV2 (General Purpose v2) recommended

## Application Stack Options

### Managed Runtimes

| Runtime | Field | Supported Versions | Notes |
|---------|-------|--------------------|-------|
| Python | `python_version` | 3.8, 3.9, 3.10, 3.11, 3.12, 3.13, 3.14 | Most popular for data/ML workloads |
| Node.js | `node_version` | 12, 14, 16, 18, 20, 22, 24 | Popular for HTTP APIs and webhooks |
| .NET | `dotnet_version` | 3.1, 6.0, 7.0, 8.0, 9.0, 10.0 | Isolated worker model recommended |
| Java | `java_version` | 8, 11, 17, 21 | Enterprise workloads |
| PowerShell | `powershell_core_version` | 7, 7.2, 7.4 | Azure automation scripts |

### Docker Containers

The `docker` application stack runs custom container images:

```yaml
site_config:
  application_stack:
    docker:
      registry_url: https://myregistry.azurecr.io
      image_name: myorg/my-function-app
      image_tag: v1.0.0
```

Container requirements:
- Must include the Azure Functions runtime base image (or custom handler)
- Supported on Elastic Premium and Dedicated plans only (not Consumption)
- ACR authentication via managed identity (`container_registry_use_managed_identity: true`) or registry credentials

### Custom Handler

The `use_custom_runtime` flag enables any language (Rust, Go, etc.) by implementing a lightweight HTTP server that communicates with the Functions host process. The custom handler receives trigger payloads as HTTP requests and returns responses.

## Networking Modes

### Public Access (Default)

The Function App is accessible via its `{name}.azurewebsites.net` hostname from the public internet. Outbound traffic uses Azure's shared SNAT IPs.

### VNet Integration

Setting `virtual_network_subnet_id` enables outbound VNet integration:

- **Outbound traffic**: Routes through the specified subnet
- **Private resources**: Access databases, Redis, and other VNet-connected services without public endpoints
- **Route all traffic**: `vnet_route_all_enabled: true` routes ALL outbound traffic (including public internet) through the VNet
- **Subnet delegation**: The subnet must be delegated to `Microsoft.Web/serverFarms`
- **Not supported**: Consumption plans (Y1)

### Private Endpoint (Not in v1)

Private endpoints give the Function App a private IP address within the VNet, making it inaccessible from the public internet. This is modeled as a separate `AzurePrivateEndpoint` component in OpenMCF, not as a field on AzureFunctionApp.

### IP Restrictions

The `ip_restrictions` list on `site_config` provides IP-based access control:

- **IP CIDR rules**: Allow/deny specific IP ranges
- **Service tag rules**: Allow Azure service tags (e.g., `AzureFrontDoor.Backend`)
- **VNet rules**: Allow traffic from specific subnets
- **Default action**: `ip_restriction_default_action` sets the fallback (Allow or Deny)
- **SCM site**: Separate restrictions for the Kudu/SCM endpoint

## Identity and Security

### Managed Identity

The `identity` block supports three modes:

| Mode | Field | Description |
|------|-------|-------------|
| System-assigned | `type: "SystemAssigned"` | Identity tied to Function App lifecycle; auto-deleted when app is deleted |
| User-assigned | `type: "UserAssigned"` + `identity_ids` | Pre-created identity with independent lifecycle; shared across resources |
| Both | `type: "SystemAssigned,UserAssigned"` | Both identity types simultaneously |

Outputs `identity_principal_id` and `identity_tenant_id` are populated for SystemAssigned identities.

### Key Vault References

App settings can reference Key Vault secrets using the `@Microsoft.KeyVault(SecretUri=...)` syntax. The `key_vault_reference_identity_id` field specifies which identity authenticates with Key Vault. If not set, the system-assigned identity is used.

### Client Certificates (mTLS)

The `client_certificate_enabled` and `client_certificate_mode` fields enable mutual TLS:

- **Required**: All requests must present a valid certificate
- **Optional**: Certificate is requested but not required
- **OptionalInteractiveUser**: Certificate is optional for browser users

### HTTPS Only

`https_only: true` (default) redirects all HTTP requests to HTTPS. This is a secure-by-default override of the Azure API default (`false`).

### FTPS State

`ftps_state: "Disabled"` (default) completely disables FTP/FTPS file deployment. This is a security best practice -- deployments should use CI/CD pipelines, not FTP.

## Design Decisions

### DD01: Linux-Only

Windows Function Apps (`azurerm_windows_function_app`) are excluded. Rationale:
- Linux covers >90% of serverless workloads (Python, Node.js, Docker containers)
- Windows adds a separate resource type with different field semantics
- The `AzureServicePlan` supports both OS types, so the compute tier is ready
- Can be added as `AzureWindowsFunctionApp` if demand exists

### DD02: StringValueOrRef for Service Plan, Storage, and Subnet

All upstream references use `StringValueOrRef` with `default_kind` annotations:
- `service_plan_id` -> `AzureServicePlan.status.outputs.plan_id`
- `storage_account_name` -> `AzureStorageAccount.status.outputs.storage_account_name`
- `virtual_network_subnet_id` -> `AzureSubnet.status.outputs.subnet_id`
- `application_insights_connection_string` -> `AzureApplicationInsights.status.outputs.connection_string`

This enables infra chart composition via `valueFrom` while supporting literal strings for standalone use.

### DD03: Storage Access Key is Not Exported

The `AzureStorageAccount` component intentionally does not export access keys in its `status.outputs` (exporting secrets through status is an anti-pattern). This means:
- `storage_account_access_key` must be provided as a literal or external secret reference
- `storage_uses_managed_identity: true` is the recommended alternative (no secret management)

### DD04: Secure Defaults

Several defaults differ from Azure API defaults to enforce security:
- `https_only: true` (Azure default: false)
- `ftps_state: "Disabled"` (Azure default: "AllAllowed")
- `minimum_tls_version: "1.2"` (Azure default: "1.2" -- matches)

### DD05: site_config as Required Object

Unlike `app_settings` (optional map), `site_config` is required because it contains `application_stack`, which is essential -- a Function App without a runtime is not functional. Making `site_config` required ensures users always specify a runtime.

### DD06: Omitted auth_settings

Azure App Service Authentication ("Easy Auth") provides built-in identity provider integration (Azure AD, Google, Facebook, etc.). We omitted it because:
- Complex configuration surface: 20+ sub-fields across `auth_settings` and `auth_settings_v2`
- Most production apps implement authentication in application code
- Can be added in v2 when demand materializes

### DD07: Omitted Deployment Slots

Deployment slots enable blue-green deployments. Omitted because:
- Slots are modeled as child resources (`Microsoft.Web/sites/slots`)
- Require additional concepts: slot settings, swap operations, traffic splitting
- Significant complexity for a feature that many teams handle via CI/CD pipelines
- Deferred to v2

### DD08: Omitted backup

Automated backup configuration. Omitted because:
- Most function apps are stateless (code is deployed from CI/CD)
- Backup is primarily useful for apps with local state or configuration drift
- Niche feature with low demand for serverless workloads

## Terraform Provider Analysis

### Source Files

- `internal/services/appservice/linux_function_app_resource.go` -- Resource implementation
- `internal/services/appservice/helpers/function_app_schema.go` -- Schema helpers
- `internal/services/appservice/helpers/function_app_slot_schema.go` -- Slot schema
- `internal/services/appservice/validate/function_app.go` -- Validation helpers

### Key Behaviors

1. **Name uniqueness**: Globally unique across Azure (forms `{name}.azurewebsites.net`)
2. **Storage validation**: Either `storage_account_access_key` or `storage_uses_managed_identity` must be set
3. **ForceNew fields**: `name`, `location`, `resource_group_name`, `service_plan_id` (when switching between Dynamic and non-Dynamic tiers)
4. **System-managed app_settings**: Azure auto-manages `AzureWebJobsStorage`, `FUNCTIONS_WORKER_RUNTIME`, `FUNCTIONS_EXTENSION_VERSION`, `WEBSITE_CONTENTAZUREFILECONNECTIONSTRING`, `WEBSITE_CONTENTSHARE`
5. **Docker requires EP or Dedicated**: Container-based function apps cannot run on Consumption plans

### API Version

- Azure API: `Microsoft.Web` version `2023-12-01`
- Resource ID: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/sites/{name}`

## Pulumi Provider Analysis

### Package

- `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice`
- Resource: `appservice.NewLinuxFunctionApp`
- All spec fields map directly to `LinuxFunctionAppArgs` properties

### Field Mapping

| Spec Field | Pulumi Property |
|------------|----------------|
| `name` | `Name` |
| `region` | `Location` |
| `resource_group` | `ResourceGroupName` |
| `service_plan_id` | `ServicePlanId` |
| `storage_account_name` | `StorageAccountName` |
| `storage_account_access_key` | `StorageAccountAccessKey` |
| `storage_uses_managed_identity` | `StorageUsesManagedIdentity` |
| `functions_extension_version` | `FunctionsExtensionVersion` |
| `site_config` | `SiteConfig` |
| `app_settings` | `AppSettings` |
| `connection_strings` | `ConnectionStrings` |
| `identity` | `Identity` |
| `virtual_network_subnet_id` | `VirtualNetworkSubnetId` |

## 80/20 Scoping Rationale

The included fields cover the following production scenarios:

| Scenario | Key Fields Used |
|----------|----------------|
| Python HTTP API | `application_stack.python_version`, `health_check_path`, `app_settings` |
| Node.js event processor | `application_stack.node_version`, `connection_strings`, `runtime_scale_monitoring_enabled` |
| Docker container function | `application_stack.docker`, `container_registry_use_managed_identity` |
| VNet-integrated backend | `virtual_network_subnet_id`, `vnet_route_all_enabled`, `ip_restrictions` |
| Elastic Premium with pre-warming | `elastic_instance_minimum`, `pre_warmed_instance_count`, `app_scale_limit` |
| Credential-free deployment | `storage_uses_managed_identity`, `identity`, `key_vault_reference_identity_id` |

Excluded features that can be added in v2:

| Feature | Reason for Deferral |
|---------|-------------------|
| `auth_settings` / `auth_settings_v2` | Complex surface (20+ fields), most apps handle auth in code |
| Deployment slots | Child resource with swap operations, handled by CI/CD |
| `backup` | Niche for stateless function apps |
| `sticky_settings` | Requires deployment slots |
| `zip_deploy_file` | Teams use CI/CD pipelines for deployment |
| Windows Function Apps | Separate resource type; Linux covers >90% of workloads |

## Downstream Dependencies

### Resources that AzureFunctionApp Consumes

| Upstream Resource | Field | Reference Path |
|-------------------|-------|---------------|
| AzureServicePlan | `service_plan_id` | `status.outputs.plan_id` |
| AzureStorageAccount | `storage_account_name` | `status.outputs.storage_account_name` |
| AzureApplicationInsights | `application_insights_connection_string` | `status.outputs.connection_string` |
| AzureSubnet | `virtual_network_subnet_id` | `status.outputs.subnet_id` |
| AzureResourceGroup | `resource_group` | `status.outputs.resource_group_name` |
| AzureUserAssignedIdentity | `identity.identity_ids`, `key_vault_reference_identity_id` | `status.outputs.identity_id` |

### Infra Charts

| Chart | Role |
|-------|------|
| `function-app-environment` | Leaf resource (ServicePlan -> StorageAccount -> AppInsights -> Subnet -> FunctionApp) |

## References

- [Azure Functions documentation](https://learn.microsoft.com/en-us/azure/azure-functions/)
- [Terraform azurerm_linux_function_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/linux_function_app)
- [Azure Functions pricing](https://azure.microsoft.com/en-us/pricing/details/functions/)
- [Azure Functions hosting options](https://learn.microsoft.com/en-us/azure/azure-functions/functions-scale)
- [Azure Functions networking options](https://learn.microsoft.com/en-us/azure/azure-functions/functions-networking-options)
- [Azure Functions identity-based connections](https://learn.microsoft.com/en-us/azure/azure-functions/functions-reference#configure-an-identity-based-connection)

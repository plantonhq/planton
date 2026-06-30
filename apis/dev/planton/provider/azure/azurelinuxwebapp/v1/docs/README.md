# AzureLinuxWebApp: Research & Design Documentation

## Executive Summary

Azure Linux Web App (`Microsoft.Web/sites` kind `app,linux`) is Azure's managed web hosting platform for running long-lived HTTP workloads on App Service. It hosts web applications, REST APIs, GraphQL endpoints, server-rendered frontends, and containerized microservices. Unlike serverless Function Apps that are event-driven and scale to zero, Web Apps are designed for always-on workloads that serve continuous HTTP traffic.

This document captures the research, design rationale, and 80/20 scoping decisions behind the `AzureLinuxWebApp` Planton component.

## Azure App Service Overview

### What It Is

Azure App Service is a fully managed PaaS (Platform as a Service) for hosting web applications, APIs, and mobile backends. Under the hood, it runs on Azure VMs managed by Microsoft, with automatic OS patching, load balancing, and health monitoring. The developer provides application code or a container image; Azure handles the infrastructure.

### How It Works

```
                    ┌──────────────────────────────────┐
                    │         App Service Plan          │
                    │  (Free / Basic / Standard /       │
                    │   Premium / Isolated)             │
                    └────────────┬─────────────────────┘
                                 │
                    ┌────────────▼─────────────────────┐
                    │        Linux Web App              │
                    │  ┌─────────────────────────────┐  │
                    │  │   Application Stack          │  │
                    │  │   (.NET / Node / Python /    │  │
                    │  │    PHP / Ruby / Go / Java /  │  │
                    │  │    Docker container)         │  │
                    │  └─────────────────────────────┘  │
                    │  ┌─────────────────────────────┐  │
                    │  │   Web Server                 │  │
                    │  │   - Gunicorn (Python)        │  │
                    │  │   - PM2 (Node.js)            │  │
                    │  │   - Kestrel (.NET)           │  │
                    │  │   - Tomcat/JBoss (Java)      │  │
                    │  │   - Custom (Docker)          │  │
                    │  └─────────────────────────────┘  │
                    └────────────┬─────────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
     ┌────────▼────────┐  ┌─────▼──────┐  ┌───────▼────────┐
     │  App Insights   │  │ Key Vault  │  │  VNet / Subnet │
     │  (optional)     │  │ (optional) │  │  (optional)    │
     └─────────────────┘  └────────────┘  └────────────────┘
```

The Web App sits atop an App Service Plan, which provides the compute resources. Multiple Web Apps can share a single plan. The plan's SKU determines performance, features, and cost:

| Tier | SKU Examples | Use Case |
|------|-------------|----------|
| Free/Shared | F1, D1 | Development, testing, hobby projects |
| Basic | B1, B2, B3 | Simple production apps, low traffic |
| Standard | S1, S2, S3 | Production apps, auto-scale, deployment slots |
| Premium | P1v3, P2v3, P3v3 | High-performance, VNet integration, zone redundancy |
| Isolated | I1v2, I2v2, I3v2 | Dedicated environment (App Service Environment) |

## Linux vs Windows App Service

### Why Linux-Only (DD04)

Azure App Service supports both Linux and Windows operating systems, each with a separate Terraform/Pulumi resource type:

| Dimension | Linux (`azurerm_linux_web_app`) | Windows (`azurerm_windows_web_app`) |
|-----------|------|---------|
| **Runtimes** | .NET, Node, Python, PHP, Ruby, Go, Java, Docker | .NET, Node, PHP, Java (no Python, Ruby, Go) |
| **Container support** | Full Docker support | Windows containers (limited) |
| **Cost** | Generally lower (no Windows licensing) | Higher (includes Windows Server licensing) |
| **Market share** | Growing majority for new deployments | Legacy workloads, .NET Framework |
| **Custom runtimes** | Docker containers for any language | Limited to supported runtimes |

The decision to scope to Linux-only (`AzureLinuxWebApp`) is based on:

1. **Coverage**: Linux supports all runtimes that Windows does, plus Python, Ruby, Go, and full Docker container support
2. **Industry trend**: New web application deployments overwhelmingly target Linux; Windows is primarily legacy
3. **API cleanliness**: Separate resource types (`azurerm_linux_web_app` vs `azurerm_windows_web_app`) have different field semantics -- combining them would create a confusing API surface
4. **Extensibility**: `AzureWindowsWebApp` can be added later if demand exists; the `AzureServicePlan` already supports both OS types

## Deployment Landscape

### Comparison with Competing Cloud Platforms

| Dimension | Azure App Service (Web App) | AWS Elastic Beanstalk | GCP Cloud Run | GCP App Engine | DigitalOcean App Platform |
|-----------|---------------------------|----------------------|--------------|----------------|--------------------------|
| **Model** | Managed PaaS (explicit plan) | Managed PaaS (auto-provisioned EC2) | Serverless containers | Managed PaaS (auto-scale) | Managed PaaS |
| **Container support** | Yes (Docker) | Yes (Docker) | Native container platform | Yes (Flex environment) | Yes (Docker) |
| **Scale to zero** | No (always-on model) | No | Yes | Yes (Standard env only) | No |
| **VNet integration** | Standard+ tier | VPC integration | VPC connector | VPC connector | VPC (limited) |
| **Custom domains** | Yes (with SSL) | Yes | Yes (with managed certs) | Yes | Yes |
| **Deployment slots** | Yes (Standard+ tier) | Environment cloning | Revisions with traffic splitting | Traffic splitting | Staging deployments |
| **OS control** | None (managed) | Limited (AMI selection) | None (serverless) | None (managed) | None (managed) |
| **Pricing** | Per-plan-instance/hour | Per-EC2-instance/hour | Per-request + per-vCPU-s | Per-instance-hour | Per-container/month |
| **Min commitment** | Free tier available | Free tier available | Pay-per-use | Free tier available | $5/month |

### When to Choose Azure App Service Over Alternatives

**Choose App Service when**:
- You want a fully managed platform with zero infrastructure management
- Your workload is an always-on web application (not batch or event-driven)
- You need rich Java support (Tomcat, JBoss EAP) with enterprise-grade features
- You're already in the Azure ecosystem with Active Directory, Key Vault, and VNet
- You need deployment slots for blue-green deployments (deferred to v2)

**Choose Azure Container Apps instead when**:
- You need Kubernetes-style features (Dapr, KEDA, sidecar containers)
- You want scale-to-zero for cost optimization on infrequent workloads
- You need multiple containers per deployment unit

**Choose Azure Functions instead when**:
- Your workload is event-driven (queue messages, blob events, timer triggers)
- You want pay-per-execution pricing (Consumption plan)
- You need serverless scale-to-zero behavior

## Application Runtime Options

### Managed Runtimes

Web Apps support a broader set of managed runtimes compared to Function Apps. Each runtime includes a pre-configured web server:

| Runtime | Field | Supported Versions | Default Web Server | Notes |
|---------|-------|--------------------|-------------------|-------|
| .NET | `dotnet_version` | 3.1, 6.0, 7.0, 8.0, 9.0, 10.0 | Kestrel | Isolated worker model; ASP.NET Core |
| Node.js | `node_version` | 12-lts, 14-lts, 16-lts, 18-lts, 20-lts, 22-lts, 24-lts | PM2 | Express, Fastify, Next.js SSR |
| Python | `python_version` | 3.7, 3.8, 3.9, 3.10, 3.11, 3.12, 3.13 | Gunicorn | Flask, FastAPI, Django |
| PHP | `php_version` | 7.4, 8.0, 8.1, 8.2, 8.3, 8.4 | Apache + mod_php / nginx | Laravel, WordPress, Symfony |
| Ruby | `ruby_version` | 2.6, 2.7 | Puma | **Deprecated** -- limited support |
| Go | `go_version` | 1.18, 1.19 | Custom handler | **Deprecated** -- use Docker instead |
| Java SE | `java_version` + `java_server: JAVA` | 8, 11, 17, 21 | Embedded (Spring Boot JAR) | Executable JAR with embedded server |
| Java Tomcat | `java_version` + `java_server: TOMCAT` | 8, 11, 17, 21 + Tomcat 8.5-10.1 | Apache Tomcat | WAR deployments, servlet-based apps |
| Java JBoss | `java_version` + `java_server: JBOSSEAP` | 8, 11, 17, 21 + JBoss 7.x-8.x | Red Hat JBoss EAP | Jakarta EE enterprise applications |

### Java Application Servers

Java on Azure App Service is unique because it supports configurable application servers via the `java_server` and `java_server_version` fields:

| Server | Versions | Use Case |
|--------|----------|----------|
| **JAVA** (SE) | 8, 11, 17, 21 | Spring Boot executable JARs, Quarkus, Micronaut -- apps with embedded servers |
| **TOMCAT** | 8.5, 9.0, 10.0, 10.1 | Traditional servlet-based apps (WAR files), Spring MVC, Jakarta Servlet |
| **JBOSSEAP** | 7, 7.4, 8.0 | Full Jakarta EE applications, enterprise workloads requiring JBoss-specific features |

### Docker Container Deployment

The `docker` application stack enables running any containerized application:

```yaml
site_config:
  application_stack:
    docker:
      registry_url: https://myregistry.azurecr.io
      image_name: myorg/my-web-app
      image_tag: v1.0.0
```

#### Container Requirements

- Must run a web server listening on the port specified by `WEBSITES_PORT` (default: 8080)
- Any HTTP-capable runtime or framework can be used (Rust, Go, Elixir, etc.)
- Image must be accessible from the specified registry at deployment time

#### ACR Integration

Azure Container Registry (ACR) integration supports two authentication modes:

| Method | Configuration | Security |
|--------|--------------|----------|
| Managed identity | `container_registry_use_managed_identity: true` + `identity.type: SystemAssigned` | Credential-free; identity needs `AcrPull` role |
| Registry credentials | `docker.registry_username` + `docker.registry_password` | Shared secret; requires credential rotation |

The managed identity approach is strongly recommended for production:
- No secrets in manifests or environment variables
- Automatic credential rotation by Azure AD
- Auditable via Azure Activity Log

## Networking

### Public Access (Default)

By default, the Web App is accessible via its `{name}.azurewebsites.net` hostname from the public internet. Outbound traffic uses Azure's shared SNAT IP addresses.

### VNet Integration

Setting `virtual_network_subnet_id` enables outbound VNet integration:

- **Outbound traffic**: Routes through the specified subnet
- **Private resources**: Access databases, Redis, and other VNet-connected services without public endpoints
- **Route all traffic**: `vnet_route_all_enabled: true` routes ALL outbound traffic (including public internet) through the VNet
- **Subnet delegation**: The subnet must be delegated to `Microsoft.Web/serverFarms`
- **Tier requirements**: Not supported on Free (F1) and Shared (D1) tiers

### Private Endpoints (Not in v1)

Private endpoints give the Web App a private IP address within the VNet, making it inaccessible from the public internet. This is modeled as a separate `AzurePrivateEndpoint` component in Planton, not as a field on AzureLinuxWebApp.

### IP Restrictions

The `ip_restrictions` list on `site_config` provides IP-based access control:

- **IP CIDR rules**: Allow/deny specific IP ranges (e.g., `203.0.113.0/24`)
- **Service tag rules**: Allow Azure service tags (e.g., `AzureFrontDoor.Backend`)
- **VNet rules**: Allow traffic from specific subnets
- **Default action**: `ip_restriction_default_action` sets the fallback (Allow or Deny)
- **SCM site**: Separate restrictions for the Kudu/SCM endpoint, or mirror main site with `scm_use_main_ip_restriction`
- **Header filters**: Filter by `X-Forwarded-For`, `X-Forwarded-Host`, `X-Azure-FDID`, `X-FD-HealthProbe` (for Front Door integration)

## Identity and Security

### Managed Identity

The `identity` block supports three modes:

| Mode | Field | Description |
|------|-------|-------------|
| System-assigned | `type: "SystemAssigned"` | Identity tied to Web App lifecycle; auto-deleted when app is deleted |
| User-assigned | `type: "UserAssigned"` + `identity_ids` | Pre-created identity with independent lifecycle; shared across resources |
| Both | `type: "SystemAssigned,UserAssigned"` | Both identity types simultaneously |

Outputs `identity_principal_id` and `identity_tenant_id` are populated for SystemAssigned identities.

### Key Vault References

App settings can reference Key Vault secrets using the `@Microsoft.KeyVault(SecretUri=...)` syntax:

```yaml
app_settings:
  DB_CONNECTION: "@Microsoft.KeyVault(SecretUri=https://my-kv.vault.azure.net/secrets/db-connection)"
```

The `key_vault_reference_identity_id` field specifies which identity authenticates with Key Vault. If not set, the system-assigned identity is used. The identity must have the `Key Vault Secrets User` role (or equivalent policy).

### Client Certificates (mTLS)

The `client_certificate_enabled` and `client_certificate_mode` fields enable mutual TLS:

| Mode | Behavior |
|------|----------|
| **Required** | All requests must present a valid certificate |
| **Optional** | Certificate is requested but not required |
| **OptionalInteractiveUser** | Certificate is optional for browser users |

The `client_certificate_exclusion_paths` field allows excluding specific paths (e.g., health check endpoints) from certificate validation.

### HTTPS Only

`https_only: true` (default) redirects all HTTP requests to HTTPS. This is a secure-by-default override of the Azure API default (`false`).

### FTPS State

`ftps_state: "Disabled"` (default) completely disables FTP/FTPS file deployment. This is a security best practice -- deployments should use CI/CD pipelines, not FTP.

### TLS Version

`minimum_tls_version: "1.2"` (default) enforces TLS 1.2 as the minimum version for incoming HTTPS connections. TLS 1.0 and 1.1 are deprecated by major browsers and security standards. The SCM (Kudu) site has a separate `scm_minimum_tls_version` field.

## Monitoring

### Application Insights Integration

The `application_insights_connection_string` field connects the Web App to Application Insights for APM telemetry:

- **Request tracing**: Automatic tracking of incoming HTTP requests with response times, status codes, and dependencies
- **Dependency tracking**: Outbound calls to databases, HTTP services, and Azure services
- **Exception logging**: Unhandled exceptions with stack traces and context
- **Custom telemetry**: Application-specific metrics and traces via the SDK
- **Live metrics**: Real-time view of incoming requests, failures, and performance

The connection string is injected into the application via the `APPLICATIONINSIGHTS_CONNECTION_STRING` app setting (or via the site_config in the provider layer).

### Diagnostic Logs

The `logs` block provides four logging capabilities:

| Feature | Field | Description |
|---------|-------|-------------|
| Application logs | `application_logs.file_system_level` | Captures application output (Off, Error, Warning, Information, Verbose) |
| HTTP logs | `http_logs.retention_in_mb` + `retention_in_days` | HTTP request/response details with configurable retention |
| Failed request tracing | `failed_request_tracing` | Detailed traces for HTTP 4xx/5xx responses |
| Detailed error messages | `detailed_error_messages` | Rich error pages (disable in production for security) |

Unlike Function Apps where logging is nested inside `site_config` as `app_service_logs`, Web App logs are a top-level block with richer configuration options.

## Scaling

### Worker Count

The `worker_count` field on `site_config` controls the number of instances allocated to the Web App. The maximum depends on the plan tier:

| Tier | Max Workers | Auto-Scale |
|------|-------------|-----------|
| Free/Shared | 1 | No |
| Basic | 3 | No (manual only) |
| Standard | 10 | Yes |
| Premium | 30 | Yes |
| Isolated | 100 | Yes |

### Always On

The `always_on` field prevents the Web App from being unloaded after idle periods:

| Tier | Always On Default | Notes |
|------|------------------|-------|
| Free (F1) | Not supported | App will be unloaded after idle |
| Basic+ | Must be explicitly set | Without `always_on: true`, the app may have cold starts after idle periods |
| Premium | Recommended `true` | Critical for production workloads |

Without `always_on: true` on Basic+ plans, Azure may unload the application after approximately 20 minutes of inactivity, resulting in cold start latency for the next request.

### Load Balancing

The `load_balancing_mode` field controls how requests are distributed across instances:

| Mode | Description |
|------|-------------|
| LeastRequests (default) | Routes to the instance with fewest active requests |
| WeightedRoundRobin | Round-robin with weight-based distribution |
| LeastResponseTime | Routes to the instance with lowest response time |
| WeightedTotalTraffic | Routes based on total traffic weight |
| RequestHash | Routes based on request hash (sticky by URL) |
| PerSiteRoundRobin | Round-robin per site (useful for multi-app plans) |

## Design Decisions

### DD01: Separate from AzureFunctionApp

Web Apps and Function Apps share the underlying `Microsoft.Web/sites` resource type but have fundamentally different behaviors:

| Aspect | Web App | Function App |
|--------|---------|-------------|
| Execution model | Always-on HTTP server | Event-driven, trigger-based |
| Scale behavior | Worker count (manual or auto-scale) | Per-function scaling |
| Storage requirement | Not required | Required (for trigger state) |
| Runtime extensions | Standard web servers | Functions runtime + extension bundles |
| Billing model | Per-instance/hour | Per-execution (Consumption) or per-instance |

Modeling them as separate components (`AzureLinuxWebApp` and `AzureFunctionApp`) provides clarity and prevents confusing cross-contamination of fields.

### DD02: Broader Runtime Support

Web Apps support more runtimes than Function Apps:

| Runtime | Web App | Function App |
|---------|---------|-------------|
| .NET | Yes | Yes |
| Node.js | Yes | Yes |
| Python | Yes | Yes |
| Java | Yes (SE/Tomcat/JBoss) | Yes (SE only) |
| PHP | Yes | No |
| Ruby | Yes (deprecated) | No |
| Go | Yes (deprecated) | No |
| Docker | Yes | Yes |
| Custom handler | N/A | Yes |

### DD03: StringValueOrRef for All Upstream References

All upstream references use `StringValueOrRef` with `default_kind` annotations:

| Field | Default Kind | Default Field Path |
|-------|-------------|-------------------|
| `resource_group` | `AzureResourceGroup` | `status.outputs.resource_group_name` |
| `service_plan_id` | `AzureServicePlan` | `status.outputs.plan_id` |
| `virtual_network_subnet_id` | `AzureSubnet` | `status.outputs.subnet_id` |
| `application_insights_connection_string` | `AzureApplicationInsights` | `status.outputs.connection_string` |
| `key_vault_reference_identity_id` | `AzureUserAssignedIdentity` | `status.outputs.identity_id` |

This enables infra chart composition via `valueFrom` while supporting literal strings for standalone use.

### DD04: Linux-Only

Windows Web Apps (`azurerm_windows_web_app`) are excluded. Rationale:
- Linux covers >90% of new web application deployments
- Windows adds a separate resource type with different field semantics (e.g., virtual applications, .NET Framework support)
- The `AzureServicePlan` supports both OS types, so the compute tier is ready
- Can be added as `AzureWindowsWebApp` if demand exists

### DD05: Logs as Top-Level Block

Unlike Function Apps where logging is nested inside `site_config.app_service_logs`, Web App logs are a top-level block. This matches the Terraform provider's structure and provides richer configuration:

| Function App Logs | Web App Logs |
|-------------------|-------------|
| `site_config.app_service_logs.disk_quota_mb` | `logs.http_logs.retention_in_mb` |
| `site_config.app_service_logs.retention_period_days` | `logs.http_logs.retention_in_days` |
| N/A | `logs.application_logs.file_system_level` |
| N/A | `logs.failed_request_tracing` |
| N/A | `logs.detailed_error_messages` |

### DD06: Secure Defaults

Several defaults differ from Azure API defaults to enforce security:

| Field | Planton Default | Azure API Default | Rationale |
|-------|----------------|-------------------|-----------|
| `https_only` | `true` | `false` | Enforce HTTPS-only by default |
| `ftps_state` | `"Disabled"` | `"AllAllowed"` | Disable insecure FTP access |
| `minimum_tls_version` | `"1.2"` | `"1.2"` | Matches (already secure) |
| `use_32_bit_worker` | `false` | `true` | 64-bit workers are recommended for production |
| `client_affinity_enabled` | `false` | `true` (Azure) | Disable ARR cookies for stateless apps |

### DD07: Client Affinity Default Override

Azure defaults `client_affinity_enabled` to `true` (ARR session affinity cookies). Planton defaults to `false` because:
- Most modern web apps are stateless (session in Redis/DB, not in-memory)
- ARR cookies add overhead and prevent even load distribution
- Can be explicitly enabled for the rare stateful web app

### DD08: Omitted auth_settings

Azure App Service Authentication ("Easy Auth") provides built-in identity provider integration (Azure AD, Google, Facebook, etc.). We omitted it because:
- Complex configuration surface: 20+ sub-fields across `auth_settings` and `auth_settings_v2`
- Most production apps implement authentication in application code (custom middleware, JWT validation)
- Easy Auth is primarily useful for rapid prototyping, not production
- Can be added in v2 when demand materializes

### DD09: Omitted Deployment Slots

Deployment slots enable blue-green deployments and traffic splitting. Omitted because:
- Slots are modeled as child resources (`Microsoft.Web/sites/slots`)
- Require additional concepts: slot settings, swap operations, traffic splitting percentages
- Significant complexity for a feature that many teams handle via CI/CD pipelines (Kubernetes-style rolling updates, or separate staging environments)
- Deferred to v2

### DD10: Omitted backup

Automated backup configuration. Omitted because:
- Most web apps are stateless (code is deployed from CI/CD)
- Backup is primarily useful for apps with local state or configuration drift
- Niche feature with low demand for modern web architectures

## Terraform Provider Analysis

### Source Files

- `internal/services/appservice/linux_web_app_resource.go` -- Resource implementation
- `internal/services/appservice/helpers/web_app_schema.go` -- Schema helpers
- `internal/services/appservice/helpers/web_app_slot_schema.go` -- Slot schema
- `internal/services/appservice/validate/web_app.go` -- Validation helpers

### Key Behaviors

1. **Name uniqueness**: Globally unique across Azure (forms `{name}.azurewebsites.net`)
2. **ForceNew fields**: `name`, `location`, `resource_group_name`
3. **Application stack mutex**: Exactly one runtime must be specified (validation in provider)
4. **Java triple**: `java_version`, `java_server`, `java_server_version` must be set together
5. **Docker container**: Container images require the `docker` block within `application_stack`
6. **Logs vs site_config**: Logging is a top-level `logs` block, not nested in `site_config`

### API Version

- Azure API: `Microsoft.Web` version `2023-12-01`
- Resource ID: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/sites/{name}`

## Pulumi Provider Analysis

### Package

- `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice`
- Resource: `appservice.NewLinuxWebApp`
- All spec fields map directly to `LinuxWebAppArgs` properties

### Field Mapping

| Spec Field | Pulumi Property |
|------------|----------------|
| `name` | `Name` |
| `region` | `Location` |
| `resource_group` | `ResourceGroupName` |
| `service_plan_id` | `ServicePlanId` |
| `site_config` | `SiteConfig` |
| `app_settings` | `AppSettings` |
| `connection_strings` | `ConnectionStrings` |
| `identity` | `Identity` |
| `virtual_network_subnet_id` | `VirtualNetworkSubnetId` |
| `https_only` | `HttpsOnly` |
| `public_network_access_enabled` | `PublicNetworkAccessEnabled` |
| `enabled` | `Enabled` |
| `client_affinity_enabled` | `ClientAffinityEnabled` |
| `client_certificate_enabled` | `ClientCertificateEnabled` |
| `client_certificate_mode` | `ClientCertificateMode` |
| `client_certificate_exclusion_paths` | `ClientCertificateExclusionPaths` |
| `key_vault_reference_identity_id` | `KeyVaultReferenceIdentityId` |
| `logs` | `Logs` |

## 80/20 Scoping Rationale

### What's Included

The included fields cover the following production scenarios:

| Scenario | Key Fields Used |
|----------|----------------|
| Python web API (Flask/FastAPI/Django) | `application_stack.python_version`, `health_check_path`, `app_settings` |
| Node.js web app (Express/Next.js SSR) | `application_stack.node_version`, `cors`, `http2_enabled` |
| Java enterprise app | `application_stack.java_*`, `connection_strings`, `always_on` |
| Docker containerized service | `application_stack.docker`, `container_registry_use_managed_identity` |
| VNet-integrated backend | `virtual_network_subnet_id`, `vnet_route_all_enabled`, `ip_restrictions` |
| Enterprise with security controls | `identity`, `key_vault_reference_identity_id`, `client_certificate_*`, `logs` |
| Multi-instance with load balancing | `worker_count`, `always_on`, `load_balancing_mode` |

### What's Excluded (Deferred to v2)

| Feature | Reason for Deferral |
|---------|-------------------|
| `auth_settings` / `auth_settings_v2` | Complex surface (20+ fields), most apps handle auth in code |
| Deployment slots | Child resource with swap operations, handled by CI/CD |
| `backup` | Niche for stateless web apps |
| `sticky_settings` | Requires deployment slots |
| `zip_deploy_file` | Teams use CI/CD pipelines for deployment |
| Windows Web Apps | Separate resource type; Linux covers >90% of workloads |
| Custom domain bindings | Separate resource (`AzureAppServiceCustomHostnameBinding`) |
| Auto-scale rules | Managed via `AzureMonitorAutoScaleSetting` resource |

## Downstream Dependencies

### Resources that AzureLinuxWebApp Consumes

| Upstream Resource | Field | Reference Path |
|-------------------|-------|---------------|
| AzureServicePlan | `service_plan_id` | `status.outputs.plan_id` |
| AzureApplicationInsights | `application_insights_connection_string` | `status.outputs.connection_string` |
| AzureSubnet | `virtual_network_subnet_id` | `status.outputs.subnet_id` |
| AzureResourceGroup | `resource_group` | `status.outputs.resource_group_name` |
| AzureUserAssignedIdentity | `identity.identity_ids`, `key_vault_reference_identity_id` | `status.outputs.identity_id` |

### Infra Charts

| Chart | Role |
|-------|------|
| `web-app-environment` | Leaf resource (ServicePlan -> AppInsights -> Subnet -> LinuxWebApp) |

## Best Practices for Production

### Security Checklist

- [ ] `https_only: true` -- enforce HTTPS
- [ ] `minimum_tls_version: "1.2"` -- minimum TLS 1.2
- [ ] `ftps_state: Disabled` -- no FTP access
- [ ] `identity.type: SystemAssigned` -- managed identity for Azure service access
- [ ] Key Vault references for secrets (`@Microsoft.KeyVault(SecretUri=...)`)
- [ ] `ip_restriction_default_action: Deny` -- explicit allow-list for IP restrictions
- [ ] `scm_use_main_ip_restriction: true` -- protect the Kudu/SCM endpoint
- [ ] `use_32_bit_worker: false` -- 64-bit workers for production

### Performance Checklist

- [ ] `always_on: true` -- prevent cold starts (Basic+ tier required)
- [ ] `http2_enabled: true` -- multiplexing and compression
- [ ] `health_check_path` configured -- Azure removes unhealthy instances
- [ ] `worker_count` sized for expected load
- [ ] `load_balancing_mode: LeastRequests` (default, good for most workloads)
- [ ] Application Insights enabled for monitoring

### Networking Checklist

- [ ] `virtual_network_subnet_id` -- VNet integration for private resource access
- [ ] `vnet_route_all_enabled: true` -- route all outbound traffic through VNet
- [ ] Subnet delegated to `Microsoft.Web/serverFarms`
- [ ] NSG rules on subnet allow outbound traffic to required Azure services
- [ ] IP restrictions configured for inbound access control

## References

- [Azure App Service documentation](https://learn.microsoft.com/en-us/azure/app-service/)
- [Terraform azurerm_linux_web_app](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/linux_web_app)
- [Azure App Service pricing](https://azure.microsoft.com/en-us/pricing/details/app-service/linux/)
- [Azure App Service networking](https://learn.microsoft.com/en-us/azure/app-service/networking-features)
- [Azure App Service managed identity](https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity)
- [Azure App Service TLS/SSL](https://learn.microsoft.com/en-us/azure/app-service/configure-ssl-bindings)
- [Azure App Service health check](https://learn.microsoft.com/en-us/azure/app-service/monitor-instances-health-check)
- [Azure App Service configuration](https://learn.microsoft.com/en-us/azure/app-service/configure-common)

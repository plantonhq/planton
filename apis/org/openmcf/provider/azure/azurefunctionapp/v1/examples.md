# AzureFunctionApp Examples

## Minimal Python Function App

The simplest configuration for a Python HTTP API. Uses a literal resource group and storage account with key-based authentication.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: py-functions
spec:
  region: eastus
  resource_group: my-rg
  name: py-functions
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/serverfarms/my-plan
  storage_account_name: pyfuncstorage
  storage_account_access_key: <storage-access-key>
  site_config:
    application_stack:
      python_version: "3.12"
```

## Node.js Function App with Application Insights

Node.js 20 runtime with Application Insights telemetry, HTTPS enforcement, and a health check endpoint.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: node-api
spec:
  region: westus2
  resource_group: api-rg
  name: node-api-functions
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/api-rg/providers/Microsoft.Web/serverfarms/api-plan
  storage_account_name: nodeapifuncstorage
  storage_account_access_key: <storage-access-key>
  application_insights_connection_string: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://eastus-0.in.applicationinsights.azure.com/
  https_only: true
  app_settings:
    NODE_ENV: production
    API_BASE_URL: https://api.example.com
  site_config:
    application_stack:
      node_version: "20"
    health_check_path: /api/health
    minimum_tls_version: "1.2"
```

## Docker Containerized Function App

Custom container image running on an Elastic Premium plan. Uses managed identity for ACR authentication instead of registry credentials.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: custom-fn
spec:
  region: eastus
  resource_group: containers-rg
  name: custom-fn-app
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/containers-rg/providers/Microsoft.Web/serverfarms/ep-plan
  storage_account_name: customfnstorage
  storage_account_access_key: <storage-access-key>
  identity:
    type: SystemAssigned
  site_config:
    application_stack:
      docker:
        registry_url: https://myregistry.azurecr.io
        image_name: myorg/my-function-app
        image_tag: v1.2.3
    container_registry_use_managed_identity: true
    health_check_path: /api/health
    always_on: true
```

## Enterprise Function App with VNet + Identity + IP Restrictions

Production-grade configuration with VNet integration, system-assigned managed identity, IP-based access control, CORS, connection strings, and full security hardening.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: enterprise-fn
spec:
  region: eastus
  resource_group: production-rg
  name: enterprise-event-processor
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/production-rg/providers/Microsoft.Web/serverfarms/prod-plan
  storage_account_name: enterprisefnstorage
  storage_uses_managed_identity: true
  application_insights_connection_string: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://eastus-0.in.applicationinsights.azure.com/
  https_only: true
  builtin_logging_enabled: false
  virtual_network_subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/production-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/functions-subnet
  identity:
    type: SystemAssigned
  app_settings:
    ENVIRONMENT: production
    SERVICE_BUS_CONNECTION: "@Microsoft.KeyVault(SecretUri=https://my-kv.vault.azure.net/secrets/sb-connection)"
  connection_strings:
    - name: Database
      type: PostgreSQL
      value: "Host=mydb.postgres.database.azure.com;Database=appdb;Port=5432;User Id=appuser;Password=<password>;Ssl Mode=Require;"
  site_config:
    application_stack:
      python_version: "3.12"
    always_on: true
    health_check_path: /api/health
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    http2_enabled: true
    vnet_route_all_enabled: true
    runtime_scale_monitoring_enabled: true
    cors:
      allowed_origins:
        - https://portal.example.com
        - https://admin.example.com
      support_credentials: true
    ip_restrictions:
      - name: allow-office
        priority: 100
        action: Allow
        ip_address: 203.0.113.0/24
        description: Corporate office CIDR
      - name: allow-vpn
        priority: 200
        action: Allow
        ip_address: 198.51.100.0/24
        description: VPN gateway CIDR
    ip_restriction_default_action: Deny
    scm_use_main_ip_restriction: true
```

## Elastic Premium with Pre-Warmed Instances + Managed Identity Storage

High-performance serverless configuration using Elastic Premium plan with pre-warmed instances, managed identity for storage (no access keys), and scale limits for cost control.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: premium-fn
spec:
  region: westeurope
  resource_group: premium-rg
  name: high-perf-functions
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/premium-rg/providers/Microsoft.Web/serverfarms/ep2-plan
  storage_account_name: premiumfnstorage
  storage_uses_managed_identity: true
  application_insights_connection_string: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://westeurope-0.in.applicationinsights.azure.com/
  identity:
    type: SystemAssigned
  app_settings:
    FUNCTIONS_WORKER_PROCESS_COUNT: "4"
    AzureWebJobsDisableHomepage: "true"
  site_config:
    application_stack:
      python_version: "3.12"
    app_scale_limit: 30
    elastic_instance_minimum: 2
    pre_warmed_instance_count: 3
    health_check_path: /api/health
    http2_enabled: true
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    runtime_scale_monitoring_enabled: true
```

## Infra Chart: valueFrom Pattern

In an infra chart, the Function App references upstream resources via `valueFrom`. This example shows the full dependency chain: ResourceGroup -> ServicePlan, StorageAccount, ApplicationInsights, Subnet -> FunctionApp.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: event-processor
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: event-processor
  service_plan_id:
    valueFrom:
      kind: AzureServicePlan
      name: functions-plan
      fieldPath: status.outputs.plan_id
  storage_account_name:
    valueFrom:
      kind: AzureStorageAccount
      name: func-storage
      fieldPath: status.outputs.storage_account_name
  storage_uses_managed_identity: true
  application_insights_connection_string:
    valueFrom:
      kind: AzureApplicationInsights
      name: func-insights
      fieldPath: status.outputs.connection_string
  virtual_network_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: functions-subnet
      fieldPath: status.outputs.subnet_id
  identity:
    type: SystemAssigned
  https_only: true
  builtin_logging_enabled: false
  site_config:
    application_stack:
      python_version: "3.12"
    always_on: true
    health_check_path: /api/health
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    vnet_route_all_enabled: true
    ip_restrictions:
      - name: allow-vnet-only
        priority: 100
        action: Allow
        virtual_network_subnet_id:
          valueFrom:
            kind: AzureSubnet
            name: app-subnet
            fieldPath: status.outputs.subnet_id
        description: Allow traffic from app subnet
    ip_restriction_default_action: Deny
```

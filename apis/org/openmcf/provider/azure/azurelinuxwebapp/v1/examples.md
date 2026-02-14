# AzureLinuxWebApp Examples

## Minimal Python Web App

The simplest configuration for a Python web application. Uses a literal resource group, HTTPS enforcement, and a health check endpoint.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: py-webapp
spec:
  region: eastus
  resource_group: my-rg
  name: py-webapp
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/serverfarms/my-plan
  site_config:
    application_stack:
      python_version: "3.12"
    health_check_path: /health
    minimum_tls_version: "1.2"
  https_only: true
```

## Node.js API with CORS

Node.js 22 LTS web API with CORS configuration, HTTP/2 for improved performance, and Application Insights monitoring. Suitable for REST APIs serving a web frontend.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: node-api
spec:
  region: westus2
  resource_group: api-rg
  name: node-api-webapp
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/api-rg/providers/Microsoft.Web/serverfarms/api-plan
  application_insights_connection_string: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://westus2-0.in.applicationinsights.azure.com/
  https_only: true
  app_settings:
    NODE_ENV: production
    API_BASE_URL: https://api.example.com
  site_config:
    application_stack:
      node_version: "22-lts"
    health_check_path: /api/health
    http2_enabled: true
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    cors:
      allowed_origins:
        - https://myapp.example.com
        - https://admin.example.com
      support_credentials: true
```

## Java with Tomcat

Java 17 web application running on Apache Tomcat 10.0 with always-on enabled. This is the standard pattern for Java servlet-based applications (Spring MVC, Jakarta EE) deployed as WAR files.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: java-webapp
spec:
  region: eastus
  resource_group: java-rg
  name: java-webapp
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/java-rg/providers/Microsoft.Web/serverfarms/standard-plan
  https_only: true
  app_settings:
    JAVA_OPTS: "-Xms512m -Xmx1024m"
    SPRING_PROFILES_ACTIVE: production
  connection_strings:
    - name: Database
      type: PostgreSQL
      value: "Host=mydb.postgres.database.azure.com;Database=appdb;Port=5432;User Id=appuser;Password=<password>;Ssl Mode=Require;"
  site_config:
    application_stack:
      java_version: "17"
      java_server: TOMCAT
      java_server_version: "10.0"
    always_on: true
    health_check_path: /actuator/health
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    http2_enabled: true
```

## Docker Container

Custom Docker container web app with system-assigned managed identity for ACR authentication. No registry password needed -- the identity's `principal_id` must have `AcrPull` role on the container registry. The container must serve HTTP on port 8080 (configurable via `WEBSITES_PORT`).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: container-webapp
spec:
  region: eastus
  resource_group: containers-rg
  name: container-webapp
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/containers-rg/providers/Microsoft.Web/serverfarms/premium-plan
  https_only: true
  identity:
    type: SystemAssigned
  app_settings:
    WEBSITES_PORT: "8080"
    ENVIRONMENT: production
  site_config:
    application_stack:
      docker:
        registry_url: https://myregistry.azurecr.io
        image_name: myorg/my-web-app
        image_tag: v2.1.0
    container_registry_use_managed_identity: true
    health_check_path: /health
    always_on: true
    minimum_tls_version: "1.2"
    ftps_state: Disabled
```

## Enterprise Private Web App

Production-grade configuration with VNet integration, IP restrictions (default deny), Application Insights monitoring, Key Vault secret references, diagnostic logging, and client certificate support. This is the standard pattern for enterprise web applications that require network isolation and compliance controls.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: enterprise-webapp
spec:
  region: eastus
  resource_group: production-rg
  name: enterprise-webapp
  service_plan_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/production-rg/providers/Microsoft.Web/serverfarms/premium-plan
  application_insights_connection_string: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://eastus-0.in.applicationinsights.azure.com/
  https_only: true
  public_network_access_enabled: true
  virtual_network_subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/production-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/webapp-subnet
  identity:
    type: SystemAssigned
  client_certificate_enabled: true
  client_certificate_mode: Optional
  client_certificate_exclusion_paths: "/health;/readiness"
  app_settings:
    ENVIRONMENT: production
    DB_CONNECTION: "@Microsoft.KeyVault(SecretUri=https://my-kv.vault.azure.net/secrets/db-connection)"
    API_KEY: "@Microsoft.KeyVault(SecretUri=https://my-kv.vault.azure.net/secrets/api-key)"
  logs:
    application_logs:
      file_system_level: Information
    http_logs:
      retention_in_mb: 50
      retention_in_days: 7
    failed_request_tracing: true
    detailed_error_messages: false
  site_config:
    application_stack:
      python_version: "3.12"
    always_on: true
    health_check_path: /health
    health_check_eviction_time_in_min: 5
    worker_count: 3
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    http2_enabled: true
    vnet_route_all_enabled: true
    load_balancing_mode: LeastRequests
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
      - name: allow-front-door
        priority: 300
        action: Allow
        service_tag: AzureFrontDoor.Backend
        description: Azure Front Door backend traffic
        headers:
          x_azure_fdid:
            - 00000000-0000-0000-0000-000000000000
    ip_restriction_default_action: Deny
    scm_use_main_ip_restriction: true
```

## Infra Chart: valueFrom Pattern

In an infra chart, the Web App references upstream resources via `valueFrom`. This example shows the full dependency chain: ResourceGroup -> ServicePlan, ApplicationInsights, Subnet -> LinuxWebApp. All upstream references use `StringValueOrRef` for composability.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: my-api
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: my-api
  service_plan_id:
    valueFrom:
      kind: AzureServicePlan
      name: web-plan
      fieldPath: status.outputs.plan_id
  application_insights_connection_string:
    valueFrom:
      kind: AzureApplicationInsights
      name: web-insights
      fieldPath: status.outputs.connection_string
  virtual_network_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: webapp-subnet
      fieldPath: status.outputs.subnet_id
  identity:
    type: SystemAssigned
  https_only: true
  logs:
    application_logs:
      file_system_level: Information
    http_logs:
      retention_in_mb: 50
      retention_in_days: 7
  site_config:
    application_stack:
      python_version: "3.12"
    always_on: true
    health_check_path: /health
    minimum_tls_version: "1.2"
    ftps_state: Disabled
    http2_enabled: true
    vnet_route_all_enabled: true
    ip_restrictions:
      - name: allow-app-subnet
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

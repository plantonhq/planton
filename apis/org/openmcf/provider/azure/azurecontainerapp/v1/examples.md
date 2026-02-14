# AzureContainerApp Examples

## 1. Minimal -- Single Container, No Ingress

The simplest configuration. A single container with no ingress, no scaling rules, no secrets.
The app is only accessible from within the Container App Environment via its name.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: hello-world
spec:
  resource_group: dev-rg
  name: hello-world
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/dev-rg/providers/Microsoft.App/managedEnvironments/dev-env
  containers:
    - name: hello
      image: mcr.microsoft.com/k8se/quickstart:latest
      cpu: 0.25
      memory: "0.5Gi"
```

## 2. Web Service with Ingress

HTTP service with external ingress, traffic routed to the latest revision, and health probes.
The app gets a public FQDN: `{name}.{environment-default-domain}`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: my-web-app
spec:
  resource_group: staging-rg
  name: my-web-app
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/staging-rg/providers/Microsoft.App/managedEnvironments/staging-env
  containers:
    - name: web
      image: myregistry.azurecr.io/my-web-app:v1.0.0
      cpu: 0.5
      memory: "1Gi"
      liveness_probe:
        transport: HTTP
        port: 8080
        path: /healthz
        interval_seconds: 10
        failure_count_threshold: 3
      readiness_probe:
        transport: HTTP
        port: 8080
        path: /ready
        interval_seconds: 5
        success_count_threshold: 2
  min_replicas: 1
  max_replicas: 5
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 3. Private Registry with Secrets

ACR registry with secret-based password authentication. DB connection string
passed as a secret-backed environment variable.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: my-api
spec:
  resource_group: staging-rg
  name: my-api
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/staging-rg/providers/Microsoft.App/managedEnvironments/staging-env
  containers:
    - name: api
      image: mycompany.azurecr.io/my-api:v2.1.0
      cpu: 0.5
      memory: "1Gi"
      env:
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          secret_name: db-connection
  secrets:
    - name: acr-password
      value: "my-acr-password-here"
    - name: db-connection
      value: "Server=mydb.database.windows.net;Database=mydb;User Id=admin;Password=s3cret;"
  registries:
    - server: mycompany.azurecr.io
      username: mycompany
      password_secret_name: acr-password
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 4. Production with Scaling

HTTP-based auto-scaling with min 2 / max 20 replicas. Liveness and readiness
probes protect against unhealthy instances and slow starts.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: prod-api
  org: mycompany
  env: production
spec:
  resource_group: production-rg
  name: prod-api
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.App/managedEnvironments/prod-env
  containers:
    - name: api
      image: mycompany.azurecr.io/prod-api:v3.0.0
      cpu: 1.0
      memory: "2Gi"
      env:
        - name: ASPNETCORE_ENVIRONMENT
          value: Production
      liveness_probe:
        transport: HTTP
        port: 8080
        path: /healthz
        initial_delay_in_seconds: 5
        interval_seconds: 10
        timeout_seconds: 3
        failure_count_threshold: 3
      readiness_probe:
        transport: HTTP
        port: 8080
        path: /ready
        interval_seconds: 5
        timeout_seconds: 2
        success_count_threshold: 2
        failure_count_threshold: 3
  min_replicas: 2
  max_replicas: 20
  cooldown_period_in_seconds: 180
  polling_interval_in_seconds: 15
  termination_grace_period_seconds: 30
  http_scale_rules:
    - name: http-requests
      concurrent_requests: "100"
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 5. Background Worker -- Scale-to-Zero

No ingress. Custom KEDA scale rule using `azure-servicebus` scaler.
`min_replicas: 0` enables scale-to-zero (no cost when the queue is empty).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: order-processor
spec:
  resource_group: production-rg
  name: order-processor
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.App/managedEnvironments/prod-env
  containers:
    - name: worker
      image: mycompany.azurecr.io/order-processor:v1.5.0
      cpu: 0.5
      memory: "1Gi"
      env:
        - name: QUEUE_CONNECTION
          secret_name: sb-connection
  min_replicas: 0
  max_replicas: 10
  custom_scale_rules:
    - name: queue-depth
      custom_rule_type: azure-servicebus
      metadata:
        queueName: orders
        messageCount: "10"
        namespace: my-namespace
      authentication:
        - secret_name: sb-connection
          trigger_parameter: connection
  secrets:
    - name: sb-connection
      value: "Endpoint=sb://my-namespace.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=..."
```

## 6. Multi-Container with Init Container

Main container plus an init container for database migration. A shared EmptyDir
volume passes migration status between init and main containers.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: my-app
spec:
  resource_group: staging-rg
  name: my-app
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/staging-rg/providers/Microsoft.App/managedEnvironments/staging-env
  init_containers:
    - name: db-migrate
      image: mycompany.azurecr.io/db-migrator:v1.0.0
      cpu: 0.25
      memory: "0.5Gi"
      env:
        - name: DATABASE_URL
          secret_name: db-connection
      command:
        - /bin/sh
        - -c
        - "migrate -database $DATABASE_URL -path /migrations up && echo 'done' > /shared/migration-status"
      volume_mounts:
        - name: shared-data
          path: /shared
  containers:
    - name: app
      image: mycompany.azurecr.io/my-app:v2.0.0
      cpu: 0.5
      memory: "1Gi"
      env:
        - name: DATABASE_URL
          secret_name: db-connection
      volume_mounts:
        - name: shared-data
          path: /shared
  volumes:
    - name: shared-data
      storage_type: EmptyDir
  secrets:
    - name: db-connection
      value: "Server=mydb.database.windows.net;Database=mydb;User Id=admin;Password=s3cret;"
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 7. Dapr-Enabled Microservice

Dapr sidecar with gRPC protocol for service-to-service invocation.
Other Dapr-enabled apps invoke this service using `app_id: "payment-service"`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: payment-service
spec:
  resource_group: production-rg
  name: payment-service
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.App/managedEnvironments/prod-env
  containers:
    - name: payment
      image: mycompany.azurecr.io/payment-service:v1.3.0
      cpu: 0.5
      memory: "1Gi"
      env:
        - name: DAPR_GRPC_PORT
          value: "50001"
  min_replicas: 2
  max_replicas: 8
  http_scale_rules:
    - name: http-requests
      concurrent_requests: "50"
  dapr:
    app_id: payment-service
    app_port: 50001
    app_protocol: grpc
  ingress:
    external_enabled: false
    target_port: 50001
    transport: http2
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 8. Full Enterprise

UserAssigned identity, Key Vault secrets, ACR with managed identity, ingress with
IP restrictions and CORS, multiple containers, volumes, probes, and scale rules.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: enterprise-api
  org: megacorp
  env: production
spec:
  resource_group: production-rg
  name: enterprise-api
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.App/managedEnvironments/prod-env
  revision_mode: Single
  containers:
    - name: api
      image: megacorp.azurecr.io/enterprise-api:v5.0.0
      cpu: 1.0
      memory: "2Gi"
      env:
        - name: ASPNETCORE_ENVIRONMENT
          value: Production
        - name: DB_CONNECTION
          secret_name: db-conn
        - name: REDIS_URL
          secret_name: redis-url
      liveness_probe:
        transport: HTTP
        port: 8080
        path: /healthz
        initial_delay_in_seconds: 10
        interval_seconds: 15
        timeout_seconds: 5
        failure_count_threshold: 3
      readiness_probe:
        transport: HTTP
        port: 8080
        path: /ready
        interval_seconds: 5
        success_count_threshold: 2
      startup_probe:
        transport: HTTP
        port: 8080
        path: /startup
        initial_delay_in_seconds: 0
        interval_seconds: 3
        timeout_seconds: 5
        failure_count_threshold: 30
      volume_mounts:
        - name: config-vol
          path: /app/config
    - name: sidecar-proxy
      image: megacorp.azurecr.io/auth-proxy:v1.2.0
      cpu: 0.25
      memory: "0.5Gi"
  volumes:
    - name: config-vol
      storage_type: EmptyDir
  min_replicas: 3
  max_replicas: 50
  cooldown_period_in_seconds: 120
  polling_interval_in_seconds: 10
  termination_grace_period_seconds: 60
  http_scale_rules:
    - name: http-load
      concurrent_requests: "75"
  secrets:
    - name: db-conn
      key_vault_secret_id: https://megacorp-kv.vault.azure.net/secrets/db-connection
      identity: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity
    - name: redis-url
      key_vault_secret_id: https://megacorp-kv.vault.azure.net/secrets/redis-url
      identity: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity
  registries:
    - server: megacorp.azurecr.io
      identity: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity
  identity:
    type: UserAssigned
    identity_ids:
      - /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity
  ingress:
    external_enabled: true
    target_port: 8080
    transport: http
    ip_security_restrictions:
      - name: allow-office
        action: Allow
        ip_address_range: 203.0.113.0/24
        description: Corporate office CIDR
      - name: allow-vpn
        action: Allow
        ip_address_range: 198.51.100.0/24
        description: Corporate VPN CIDR
      - name: deny-all
        action: Deny
        ip_address_range: 0.0.0.0/0
        description: Deny all other traffic
    cors_policy:
      allowed_origins:
        - https://app.megacorp.com
        - https://admin.megacorp.com
      allowed_methods:
        - GET
        - POST
        - PUT
        - DELETE
        - OPTIONS
      allowed_headers:
        - Content-Type
        - Authorization
        - X-Request-Id
      max_age_in_seconds: 3600
      allow_credentials_enabled: true
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 9. Infra Chart: valueFrom Pattern

All `StringValueOrRef` fields using `valueFrom` to reference upstream resources.
This is the typical pattern inside an infra chart where resources are wired via DAG dependencies.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: platform-api
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: platform-api
  container_app_environment_id:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: platform-env
      fieldPath: status.outputs.environment_id
  containers:
    - name: api
      image: mycompany.azurecr.io/platform-api:v1.0.0
      cpu: 0.5
      memory: "1Gi"
  identity:
    type: UserAssigned
    identity_ids:
      - valueFrom:
          kind: AzureUserAssignedIdentity
          name: platform-identity
          fieldPath: status.outputs.identity_id
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - latest_revision: true
        percentage: 100
```

## 10. Blue-Green Deployment

Multiple revision mode with traffic split 80/20 between the stable and canary revisions.
Each revision gets a labeled FQDN for direct testing: `stable.{app-fqdn}` and `canary.{app-fqdn}`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: my-service
spec:
  resource_group: production-rg
  name: my-service
  container_app_environment_id: /subscriptions/sub-id/resourceGroups/production-rg/providers/Microsoft.App/managedEnvironments/prod-env
  revision_mode: Multiple
  containers:
    - name: app
      image: mycompany.azurecr.io/my-service:v3.0.0
      cpu: 0.5
      memory: "1Gi"
      liveness_probe:
        transport: HTTP
        port: 8080
        path: /healthz
      readiness_probe:
        transport: HTTP
        port: 8080
        path: /ready
  revision_suffix: v3-canary
  min_replicas: 2
  max_replicas: 10
  http_scale_rules:
    - name: http-requests
      concurrent_requests: "100"
  ingress:
    external_enabled: true
    target_port: 8080
    traffic_weight:
      - revision_suffix: v2-stable
        percentage: 80
        label: stable
      - latest_revision: true
        percentage: 20
        label: canary
```

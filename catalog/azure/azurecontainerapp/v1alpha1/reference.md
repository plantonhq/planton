# AzureContainerApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureContainerAppSpec** defines the configuration for creating an Azure
Container App (Microsoft.App/containerApps), a serverless container workload
running inside an Azure Container Apps Managed Environment.

A Container App is a continuously running containerized service -- it
defines one or more containers, their resource allocations, scaling rules,
ingress configuration, and runtime secrets. For run-to-completion
workloads, use AzureContainerAppJob instead.

**Relationship to AzureContainerAppEnvironment**:

Every Container App runs inside an Environment. The Environment provides the
shared networking boundary, logging, Dapr infrastructure, and compute capacity.
The Container App defines the workload: which containers to run, how to scale
them, how to expose them, and what secrets they need.

**Revision model**:

Container Apps uses a revision-based deployment model. Each change to the
template section (containers, scale, volumes) creates a new revision.

SINGLE mode (default):
  - Only one revision is active at a time
  - New revisions automatically replace the old one
  - Simplest model for most workloads

MULTIPLE mode:
  - Multiple revisions can be active simultaneously
  - Traffic can be split across revisions via traffic_weight
  - Enables blue-green and canary deployment patterns

**Containers**:

At least one container is required. Each container specifies an image, CPU/memory
allocation, environment variables (literal or secret-backed), optional command/args
overrides, health probes, and volume mounts.

Init containers run to completion before main containers start. They share the
same volume mounts and secrets but cannot have health probes.

**Scaling**:

Container Apps scale based on KEDA-compatible rules:
- HTTP concurrent requests (most common for web services)
- TCP concurrent connections
- Azure Queue length
- Custom KEDA scalers (Kafka, Prometheus, Redis, cron, etc.)

Scale ranges from min_replicas to max_replicas. Setting min_replicas=0 enables
scale-to-zero (no cost when idle).

**Secrets and registries**:

Secrets can be plain-text values or Key Vault references (requiring managed
identity). They are referenced by name in container environment variables,
registry passwords, scale-rule authentication, and SECRET-type volumes.

Private container registries authenticate via username/password (referencing
a secret) or managed identity.

**Ingress**:

Optional HTTP/TCP ingress with traffic splitting, IP security restrictions,
CORS, and client certificate mode (mTLS). When ingress is not configured,
the app is only accessible from within the environment via the app name.

**No region field**: Container App location is computed from its environment.

**ForceNew fields** (changing these destroys and recreates):
`container_app_name`, `resource_group`, `container_app_environment_id`.

**Referenced by**: None (leaf workload resource)

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerApp
metadata:
  name: test-app
spec:
  resource_group:
    value: test-rg
  container_app_name: test-container-app
  container_app_environment_id:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.App/managedEnvironments/test-env
  revision_mode: MULTIPLE
  workload_profile_name: dedicated-d4
  max_inactive_revisions: 25
  revision_suffix: v2
  min_replicas: 1
  max_replicas: 30
  cooldown_period_in_seconds: 120
  polling_interval_in_seconds: 15
  termination_grace_period_seconds: 30
  containers:
    - name: web
      image: mcr.microsoft.com/k8se/quickstart:latest
      cpu: 0.5
      memory: "1Gi"
      command: ["/bin/server"]
      args: ["--port", "8080"]
      env:
        - name: MODE
          value: production
        - name: API_KEY
          secret_name: api-key
      liveness_probe:
        transport: HTTP_GET
        port: 8080
        path: /healthz
        initial_delay_in_seconds: 5
        failure_count_threshold: 5
        headers:
          - name: X-Probe
            value: liveness
      readiness_probe:
        transport: TCP_SOCKET
        port: 8080
        success_count_threshold: 2
      startup_probe:
        transport: HTTPS_GET
        port: 8443
        path: /started
        failure_count_threshold: 60
      volume_mounts:
        - name: shared-data
          path: /data
          sub_path: web
  init_containers:
    - name: migrate
      image: mcr.microsoft.com/k8se/quickstart:latest
      command: ["/bin/migrate"]
      env:
        - name: DB_URL
          secret_name: db-url
  volumes:
    - name: shared-data
      storage_type: AZURE_FILE
      storage_name:
        value: test-env-storage
      mount_options: uid=1000,gid=1000
    - name: scratch
      storage_type: EMPTY_DIR
    - name: certs
      storage_type: SECRET
  http_scale_rules:
    - name: http-load
      concurrent_requests: "100"
  azure_queue_scale_rules:
    - name: work-queue
      queue_name: work
      queue_length: 5
      authentication:
        - secret_name: queue-conn
          trigger_parameter: connection
  custom_scale_rules:
    - name: cron-scale
      custom_rule_type: cron
      metadata:
        timezone: UTC
        start: 0 8 * * 1-5
        end: 0 18 * * 1-5
        desiredReplicas: "5"
      identity_id:
        value: System
  secrets:
    - name: api-key
      value: plain-text-secret
    - name: db-url
      key_vault_secret_id: https://test-vault.vault.azure.net/secrets/db-url
      identity: System
    - name: queue-conn
      value: DefaultEndpointsProtocol=https;AccountName=test
  registries:
    - server:
        value: testregistry.azurecr.io
      identity: System
  ingress:
    external_enabled: true
    target_port: 8080
    transport: AUTO
    allow_insecure_connections: false
    client_certificate_mode: ACCEPT
    traffic_weight:
      - revision_suffix: v1
        percentage: 80
      - revision_suffix: v2
        percentage: 20
        label: canary
    ip_security_restrictions:
      - name: office
        action: ALLOW
        ip_address_range: 203.0.113.0/24
        description: Office CIDR
    cors:
      allowed_origins:
        - https://example.com
      allowed_headers:
        - Content-Type
        - Authorization
      allowed_methods:
        - GET
        - POST
      exposed_headers:
        - X-Request-Id
      max_age_in_seconds: 3600
      allow_credentials_enabled: true
  dapr:
    app_id: test-app
    app_port: 8080
    app_protocol: DAPR_HTTP
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    user_assigned_identity_ids:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.containerAppName` | `string` | yes |  |  |
| `spec.containerAppEnvironmentId` | `string \| valueFrom` | yes |  | AzureContainerAppEnvironment (`status.outputs.environment_id`) |
| `spec.revisionMode` | `enum` |  |  |  |
| `spec.workloadProfileName` | `string` |  |  |  |
| `spec.maxInactiveRevisions` | `int32` |  |  |  |
| `spec.containers` | `[]AzureContainerAppContainer` | yes |  |  |
| `spec.containers[].name` | `string` | yes |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].cpu` | `double` |  |  |  |
| `spec.containers[].memory` | `string` | yes |  |  |
| `spec.containers[].env` | `[]AzureContainerAppEnvVar` |  |  |  |
| `spec.containers[].env[].name` | `string` | yes |  |  |
| `spec.containers[].env[].value` | `string` |  |  |  |
| `spec.containers[].env[].secretName` | `string` |  |  |  |
| `spec.containers[].command` | `[]string` |  |  |  |
| `spec.containers[].args` | `[]string` |  |  |  |
| `spec.containers[].livenessProbe` | `AzureContainerAppProbe` |  |  |  |
| `spec.containers[].livenessProbe.transport` | `enum` |  |  |  |
| `spec.containers[].livenessProbe.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.path` | `string` |  |  |  |
| `spec.containers[].livenessProbe.host` | `string` |  |  |  |
| `spec.containers[].livenessProbe.headers` | `[]AzureContainerAppProbeHeader` |  |  |  |
| `spec.containers[].livenessProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].livenessProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].livenessProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].readinessProbe` | `AzureContainerAppProbe` |  |  |  |
| `spec.containers[].readinessProbe.transport` | `enum` |  |  |  |
| `spec.containers[].readinessProbe.port` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.path` | `string` |  |  |  |
| `spec.containers[].readinessProbe.host` | `string` |  |  |  |
| `spec.containers[].readinessProbe.headers` | `[]AzureContainerAppProbeHeader` |  |  |  |
| `spec.containers[].readinessProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].readinessProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].readinessProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].readinessProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].readinessProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].readinessProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].startupProbe` | `AzureContainerAppProbe` |  |  |  |
| `spec.containers[].startupProbe.transport` | `enum` |  |  |  |
| `spec.containers[].startupProbe.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.path` | `string` |  |  |  |
| `spec.containers[].startupProbe.host` | `string` |  |  |  |
| `spec.containers[].startupProbe.headers` | `[]AzureContainerAppProbeHeader` |  |  |  |
| `spec.containers[].startupProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].startupProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].startupProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].startupProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].startupProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].startupProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].volumeMounts` | `[]AzureContainerAppVolumeMount` |  |  |  |
| `spec.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].path` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.initContainers` | `[]AzureContainerAppInitContainer` |  |  |  |
| `spec.initContainers[].name` | `string` | yes |  |  |
| `spec.initContainers[].image` | `string` | yes |  |  |
| `spec.initContainers[].cpu` | `double` |  |  |  |
| `spec.initContainers[].memory` | `string` |  |  |  |
| `spec.initContainers[].env` | `[]AzureContainerAppEnvVar` |  |  |  |
| `spec.initContainers[].env[].name` | `string` | yes |  |  |
| `spec.initContainers[].env[].value` | `string` |  |  |  |
| `spec.initContainers[].env[].secretName` | `string` |  |  |  |
| `spec.initContainers[].command` | `[]string` |  |  |  |
| `spec.initContainers[].args` | `[]string` |  |  |  |
| `spec.initContainers[].volumeMounts` | `[]AzureContainerAppVolumeMount` |  |  |  |
| `spec.initContainers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.initContainers[].volumeMounts[].path` | `string` | yes |  |  |
| `spec.initContainers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.volumes` | `[]AzureContainerAppVolume` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].storageType` | `enum` |  |  |  |
| `spec.volumes[].storageName` | `string \| valueFrom` |  |  | AzureContainerAppEnvironmentStorage (`status.outputs.storage_name`) |
| `spec.volumes[].mountOptions` | `string` |  |  |  |
| `spec.minReplicas` | `int32` |  | `0` |  |
| `spec.maxReplicas` | `int32` |  | `10` |  |
| `spec.cooldownPeriodInSeconds` | `int32` |  | `300` |  |
| `spec.pollingIntervalInSeconds` | `int32` |  | `30` |  |
| `spec.revisionSuffix` | `string` |  |  |  |
| `spec.terminationGracePeriodSeconds` | `int32` |  | `0` |  |
| `spec.httpScaleRules` | `[]AzureContainerAppHttpScaleRule` |  |  |  |
| `spec.httpScaleRules[].name` | `string` | yes |  |  |
| `spec.httpScaleRules[].concurrentRequests` | `string` | yes |  |  |
| `spec.httpScaleRules[].authentication` | `[]AzureContainerAppScaleRuleAuth` |  |  |  |
| `spec.httpScaleRules[].authentication[].secretName` | `string` | yes |  |  |
| `spec.httpScaleRules[].authentication[].triggerParameter` | `string` |  |  |  |
| `spec.tcpScaleRules` | `[]AzureContainerAppTcpScaleRule` |  |  |  |
| `spec.tcpScaleRules[].name` | `string` | yes |  |  |
| `spec.tcpScaleRules[].concurrentRequests` | `string` | yes |  |  |
| `spec.tcpScaleRules[].authentication` | `[]AzureContainerAppScaleRuleAuth` |  |  |  |
| `spec.tcpScaleRules[].authentication[].secretName` | `string` | yes |  |  |
| `spec.tcpScaleRules[].authentication[].triggerParameter` | `string` |  |  |  |
| `spec.azureQueueScaleRules` | `[]AzureContainerAppAzureQueueScaleRule` |  |  |  |
| `spec.azureQueueScaleRules[].name` | `string` | yes |  |  |
| `spec.azureQueueScaleRules[].queueName` | `string` | yes |  |  |
| `spec.azureQueueScaleRules[].queueLength` | `int32` |  |  |  |
| `spec.azureQueueScaleRules[].authentication` | `[]AzureContainerAppScaleRuleAuth` | yes |  |  |
| `spec.azureQueueScaleRules[].authentication[].secretName` | `string` | yes |  |  |
| `spec.azureQueueScaleRules[].authentication[].triggerParameter` | `string` |  |  |  |
| `spec.customScaleRules` | `[]AzureContainerAppCustomScaleRule` |  |  |  |
| `spec.customScaleRules[].name` | `string` | yes |  |  |
| `spec.customScaleRules[].customRuleType` | `string` | yes |  |  |
| `spec.customScaleRules[].metadata` | `map<string, string>` | yes |  |  |
| `spec.customScaleRules[].authentication` | `[]AzureContainerAppScaleRuleAuth` |  |  |  |
| `spec.customScaleRules[].authentication[].secretName` | `string` | yes |  |  |
| `spec.customScaleRules[].authentication[].triggerParameter` | `string` |  |  |  |
| `spec.customScaleRules[].identityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.secrets` | `[]AzureContainerAppSecret` |  |  |  |
| `spec.secrets[].name` | `string` | yes |  |  |
| `spec.secrets[].value` | `string` (sensitive) |  |  |  |
| `spec.secrets[].keyVaultSecretId` | `string` |  |  |  |
| `spec.secrets[].identity` | `string` |  |  |  |
| `spec.registries` | `[]AzureContainerAppRegistry` |  |  |  |
| `spec.registries[].server` | `string \| valueFrom` | yes |  | AzureContainerRegistry (`status.outputs.login_server`) |
| `spec.registries[].username` | `string` |  |  |  |
| `spec.registries[].passwordSecretName` | `string` |  |  |  |
| `spec.registries[].identity` | `string` |  |  |  |
| `spec.ingress` | `AzureContainerAppIngress` |  |  |  |
| `spec.ingress.externalEnabled` | `bool` |  | `false` |  |
| `spec.ingress.targetPort` | `int32` |  |  |  |
| `spec.ingress.exposedPort` | `int32` |  |  |  |
| `spec.ingress.transport` | `enum` |  |  |  |
| `spec.ingress.allowInsecureConnections` | `bool` |  | `false` |  |
| `spec.ingress.clientCertificateMode` | `enum` |  |  |  |
| `spec.ingress.trafficWeight` | `[]AzureContainerAppTrafficWeight` | yes |  |  |
| `spec.ingress.trafficWeight[].latestRevision` | `bool` |  | `false` |  |
| `spec.ingress.trafficWeight[].revisionSuffix` | `string` |  |  |  |
| `spec.ingress.trafficWeight[].percentage` | `int32` |  |  |  |
| `spec.ingress.trafficWeight[].label` | `string` |  |  |  |
| `spec.ingress.ipSecurityRestrictions` | `[]AzureContainerAppIpSecurityRestriction` |  |  |  |
| `spec.ingress.ipSecurityRestrictions[].name` | `string` | yes |  |  |
| `spec.ingress.ipSecurityRestrictions[].action` | `enum` |  |  |  |
| `spec.ingress.ipSecurityRestrictions[].ipAddressRange` | `string` | yes |  |  |
| `spec.ingress.ipSecurityRestrictions[].description` | `string` |  |  |  |
| `spec.ingress.cors` | `AzureContainerAppCors` |  |  |  |
| `spec.ingress.cors.allowedOrigins` | `[]string` | yes |  |  |
| `spec.ingress.cors.allowedHeaders` | `[]string` |  |  |  |
| `spec.ingress.cors.allowedMethods` | `[]string` |  |  |  |
| `spec.ingress.cors.exposedHeaders` | `[]string` |  |  |  |
| `spec.ingress.cors.maxAgeInSeconds` | `int32` |  |  |  |
| `spec.ingress.cors.allowCredentialsEnabled` | `bool` |  | `false` |  |
| `spec.dapr` | `AzureContainerAppDapr` |  |  |  |
| `spec.dapr.appId` | `string` | yes |  |  |
| `spec.dapr.appPort` | `int32` |  |  |  |
| `spec.dapr.appProtocol` | `enum` |  |  |  |
| `spec.identity` | `AzureContainerAppIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Container App will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: Changing this destroys and recreates the app.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.containerAppName

`string` · required

The name of the Container App.
Lowercase alphanumeric characters, hyphens, and dots; must start and
end with an alphanumeric character; no consecutive hyphens; at most
32 characters.

This name is used in the app's FQDN: {name}.{environment-default-domain}

**ForceNew**: Changing this destroys and recreates the app.

- rule: Container App name must be lowercase alphanumeric characters, hyphens, or dots, start and end with an alphanumeric character, and contain no consecutive hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.containerAppEnvironmentId

`string | valueFrom` · required

The Container App Environment where this app will run.
The environment provides networking, logging, and compute capacity.

**ForceNew**: Changing this destroys and recreates the app.

- references: AzureContainerAppEnvironment (`status.outputs.environment_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerAppEnvironment, name: <that resource's name>, fieldPath: status.outputs.environment_id}} -- a bare string does not parse

### spec.revisionMode

`enum`

The revision operating mode. Controls how many revisions can be active.

SINGLE: only one revision active at a time (simplest; new revisions
replace the old automatically). Unspecified deploys SINGLE.
MULTIPLE: multiple revisions active simultaneously with traffic
splitting -- the blue-green / canary model.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_revision_mode_unspecified` -- Not specified -- deploys SINGLE, the right choice for most workloads.
- `SINGLE` -- Only one revision active at a time; new revisions replace the old.
- `MULTIPLE` -- Multiple revisions active simultaneously with traffic splitting -- enables blue-green and canary patterns.

### spec.workloadProfileName

`string`

The workload profile name to run this app on.
References a profile defined in the Container App Environment.

Omit to use the default Consumption (serverless) profile.
Set to a named profile (e.g., "gpu-pool", "high-memory") for dedicated compute.

### spec.maxInactiveRevisions

`int32` · optional (explicit presence)

Maximum number of inactive revisions to retain.
Older inactive revisions beyond this limit are automatically purged.

Range: 0-100. Omit to use Azure's default (typically 100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.containers

`[]AzureContainerAppContainer` · required

Main containers for the app. At least one container is required.
Each container runs continuously and can have health probes.

- rule: {"repeated":{"minItems":"1"}}
- rule: liveness probes do not support success_count_threshold, and failure_count_threshold must be between 1 and 30
- rule: readiness probe failure_count_threshold must be between 1 and 48
- rule: startup probes do not support success_count_threshold

### spec.containers[].name

`string` · required

Container name. Must be unique within the app.
Lowercase alphanumeric, hyphens, or dots; starts and ends with an
alphanumeric character; max 46 characters.

- rule: container name must be lowercase alphanumeric, hyphens, or dots, starting and ending with an alphanumeric character
- rule: {"required":true,"string":{"minLen":"1","maxLen":"46"}}

### spec.containers[].image

`string` · required

Container image in repository:tag format.
Examples: "mcr.microsoft.com/k8se/quickstart:latest", "myregistry.azurecr.io/myapp:v1.2.3"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].cpu

`double`

CPU allocation in vCPU cores.
Examples: 0.25, 0.5, 1.0, 2.0

Valid values depend on the workload profile. For Consumption plan:
0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0. CPU and memory scale
together (0.25 vCPU pairs with 0.5Gi, and so on).

- rule: {"double":{"gte":0.1}}

### spec.containers[].memory

`string` · required

Memory allocation in Gi format.
Examples: "0.5Gi", "1Gi", "2Gi", "4Gi"

Must match the CPU allocation. For Consumption plan with 0.25 vCPU,
memory must be "0.5Gi". See Azure docs for valid CPU/memory pairs.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env

`[]AzureContainerAppEnvVar`

Environment variables for the container.
Each variable is a literal value or a reference to a secret by name.

- rule: an environment variable takes either a literal value or a secret_name, not both -- move the literal into the app's secrets list to reference it

### spec.containers[].env[].name

`string` · required

Environment variable name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env[].value

`string`

Literal value. Mutually exclusive with secret_name.

### spec.containers[].env[].secretName

`string`

Reference to a secret name defined in the app's `secrets` list.
When set, the env var's value comes from the secret.

### spec.containers[].command

`[]string`

Container command (entrypoint override).
Overrides the image's default ENTRYPOINT.

### spec.containers[].args

`[]string`

Container arguments.
Overrides the image's default CMD.

### spec.containers[].livenessProbe

`AzureContainerAppProbe`

Liveness probe. Determines if the container is alive.
If the probe fails, the container is restarted.

Liveness probes never carry success_count_threshold, and their
failure_count_threshold tops out at 30.

### spec.containers[].livenessProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
- `TCP_SOCKET` -- Port connectivity check (Azure transport "TCP").
- `HTTP_GET` -- HTTP GET request, healthy on 200-399 (Azure transport "HTTP").
- `HTTPS_GET` -- HTTPS GET request (Azure transport "HTTPS").

### spec.containers[].livenessProbe.port

`int32`

Port to probe. Range: 1-65535.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].livenessProbe.path

`string`

URI path for HTTP/HTTPS probes. Ignored for TCP probes.
Example: "/healthz", "/ready"

### spec.containers[].livenessProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].livenessProbe.headers

`[]AzureContainerAppProbeHeader`

HTTP headers to include in probe requests. Only for HTTP/HTTPS probes.

### spec.containers[].livenessProbe.headers[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].livenessProbe.headers[].value

`string` · required

Header value.

- rule: {"required":true}

### spec.containers[].livenessProbe.initialDelayInSeconds

`int32` · optional (explicit presence)

Seconds to wait before starting probes after container start.

Unset deploys the probe type's own default: 1 for liveness, 0 for
readiness and startup.
Range: 0-60

- rule: {"int32":{"lte":60,"gte":0}}

### spec.containers[].livenessProbe.intervalSeconds

`int32` · optional (explicit presence)

Seconds between probe executions.

Default: 10
Range: 1-240

- default: `10`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].livenessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds to wait for a probe response before timing out.

Default: 1
Range: 1-240

- default: `1`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].livenessProbe.failureCountThreshold

`int32` · optional (explicit presence)

Number of consecutive failures before the probe is considered failed.

Default: 3
Ceiling by probe type: 30 (liveness), 48 (readiness), 240 (startup) --
enforced on the container.

- default: `3`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].livenessProbe.successCountThreshold

`int32` · optional (explicit presence)

Number of consecutive successes before the probe is considered successful.
Only readiness probes support this (enforced on the container).

Default: 3
Range: 1-10

- default: `3`
- rule: {"int32":{"lte":10,"gte":1}}

### spec.containers[].readinessProbe

`AzureContainerAppProbe`

Readiness probe. Determines if the container is ready to serve traffic.
If the probe fails, the container is removed from load balancing.

The only probe type that supports success_count_threshold; its
failure_count_threshold tops out at 48.

### spec.containers[].readinessProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
- `TCP_SOCKET` -- Port connectivity check (Azure transport "TCP").
- `HTTP_GET` -- HTTP GET request, healthy on 200-399 (Azure transport "HTTP").
- `HTTPS_GET` -- HTTPS GET request (Azure transport "HTTPS").

### spec.containers[].readinessProbe.port

`int32`

Port to probe. Range: 1-65535.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].readinessProbe.path

`string`

URI path for HTTP/HTTPS probes. Ignored for TCP probes.
Example: "/healthz", "/ready"

### spec.containers[].readinessProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].readinessProbe.headers

`[]AzureContainerAppProbeHeader`

HTTP headers to include in probe requests. Only for HTTP/HTTPS probes.

### spec.containers[].readinessProbe.headers[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].readinessProbe.headers[].value

`string` · required

Header value.

- rule: {"required":true}

### spec.containers[].readinessProbe.initialDelayInSeconds

`int32` · optional (explicit presence)

Seconds to wait before starting probes after container start.

Unset deploys the probe type's own default: 1 for liveness, 0 for
readiness and startup.
Range: 0-60

- rule: {"int32":{"lte":60,"gte":0}}

### spec.containers[].readinessProbe.intervalSeconds

`int32` · optional (explicit presence)

Seconds between probe executions.

Default: 10
Range: 1-240

- default: `10`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].readinessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds to wait for a probe response before timing out.

Default: 1
Range: 1-240

- default: `1`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].readinessProbe.failureCountThreshold

`int32` · optional (explicit presence)

Number of consecutive failures before the probe is considered failed.

Default: 3
Ceiling by probe type: 30 (liveness), 48 (readiness), 240 (startup) --
enforced on the container.

- default: `3`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].readinessProbe.successCountThreshold

`int32` · optional (explicit presence)

Number of consecutive successes before the probe is considered successful.
Only readiness probes support this (enforced on the container).

Default: 3
Range: 1-10

- default: `3`
- rule: {"int32":{"lte":10,"gte":1}}

### spec.containers[].startupProbe

`AzureContainerAppProbe`

Startup probe. Determines when the container has finished starting.
During startup, liveness and readiness probes are disabled.
Useful for slow-starting applications.

Startup probes never carry success_count_threshold; their
failure_count_threshold tops out at 240 (slow starters get headroom).

### spec.containers[].startupProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
- `TCP_SOCKET` -- Port connectivity check (Azure transport "TCP").
- `HTTP_GET` -- HTTP GET request, healthy on 200-399 (Azure transport "HTTP").
- `HTTPS_GET` -- HTTPS GET request (Azure transport "HTTPS").

### spec.containers[].startupProbe.port

`int32`

Port to probe. Range: 1-65535.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.path

`string`

URI path for HTTP/HTTPS probes. Ignored for TCP probes.
Example: "/healthz", "/ready"

### spec.containers[].startupProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].startupProbe.headers

`[]AzureContainerAppProbeHeader`

HTTP headers to include in probe requests. Only for HTTP/HTTPS probes.

### spec.containers[].startupProbe.headers[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].startupProbe.headers[].value

`string` · required

Header value.

- rule: {"required":true}

### spec.containers[].startupProbe.initialDelayInSeconds

`int32` · optional (explicit presence)

Seconds to wait before starting probes after container start.

Unset deploys the probe type's own default: 1 for liveness, 0 for
readiness and startup.
Range: 0-60

- rule: {"int32":{"lte":60,"gte":0}}

### spec.containers[].startupProbe.intervalSeconds

`int32` · optional (explicit presence)

Seconds between probe executions.

Default: 10
Range: 1-240

- default: `10`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds to wait for a probe response before timing out.

Default: 1
Range: 1-240

- default: `1`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.failureCountThreshold

`int32` · optional (explicit presence)

Number of consecutive failures before the probe is considered failed.

Default: 3
Ceiling by probe type: 30 (liveness), 48 (readiness), 240 (startup) --
enforced on the container.

- default: `3`
- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.successCountThreshold

`int32` · optional (explicit presence)

Number of consecutive successes before the probe is considered successful.
Only readiness probes support this (enforced on the container).

Default: 3
Range: 1-10

- default: `3`
- rule: {"int32":{"lte":10,"gte":1}}

### spec.containers[].volumeMounts

`[]AzureContainerAppVolumeMount`

Volume mounts for the container. References volumes defined in the
spec's `volumes` field by name.

### spec.containers[].volumeMounts[].name

`string` · required

Name of the volume to mount. Must match a volume name in the spec's `volumes` field.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumeMounts[].path

`string` · required

Absolute path inside the container where the volume is mounted.
Example: "/data", "/mnt/config"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumeMounts[].subPath

`string`

Sub-path within the volume to mount. When omitted, the entire volume is mounted.

### spec.initContainers

`[]AzureContainerAppInitContainer`

Init containers run to completion before main containers start.
Use for database migrations, configuration generation, or downloading
assets. Init containers cannot have health probes.

### spec.initContainers[].name

`string` · required

Container name. Must be unique within the app (across both containers
and init containers).

- rule: init container name must be lowercase alphanumeric, hyphens, or dots, starting and ending with an alphanumeric character
- rule: {"required":true,"string":{"minLen":"1","maxLen":"46"}}

### spec.initContainers[].image

`string` · required

Container image in repository:tag format.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].cpu

`double` · optional (explicit presence)

CPU allocation in vCPU cores. Optional for init containers.
When omitted, inherits from the app's overall resource allocation.

### spec.initContainers[].memory

`string` · optional (explicit presence)

Memory allocation in Gi format. Optional for init containers.
When omitted, inherits from the app's overall resource allocation.

### spec.initContainers[].env

`[]AzureContainerAppEnvVar`

Environment variables for the init container.

- rule: an environment variable takes either a literal value or a secret_name, not both -- move the literal into the app's secrets list to reference it

### spec.initContainers[].env[].name

`string` · required

Environment variable name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].env[].value

`string`

Literal value. Mutually exclusive with secret_name.

### spec.initContainers[].env[].secretName

`string`

Reference to a secret name defined in the app's `secrets` list.
When set, the env var's value comes from the secret.

### spec.initContainers[].command

`[]string`

Container command (entrypoint override).

### spec.initContainers[].args

`[]string`

Container arguments.

### spec.initContainers[].volumeMounts

`[]AzureContainerAppVolumeMount`

Volume mounts for the init container.

### spec.initContainers[].volumeMounts[].name

`string` · required

Name of the volume to mount. Must match a volume name in the spec's `volumes` field.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumeMounts[].path

`string` · required

Absolute path inside the container where the volume is mounted.
Example: "/data", "/mnt/config"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumeMounts[].subPath

`string`

Sub-path within the volume to mount. When omitted, the entire volume is mounted.

### spec.volumes

`[]AzureContainerAppVolume`

Volumes available to containers in this app.
Containers reference volumes by name in their volume_mounts field.

- rule: AZURE_FILE and NFS_AZURE_FILE volumes require storage_name (the environment storage resource backing the share); EMPTY_DIR and SECRET volumes must omit it

### spec.volumes[].name

`string` · required

Volume name. Referenced by containers in their volume_mounts field.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.volumes[].storageType

`enum`

Storage type backing the volume. Unspecified deploys EMPTY_DIR.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_volume_storage_type_unspecified` -- Not specified -- deploys EMPTY_DIR (ephemeral scratch space).
- `EMPTY_DIR` -- Ephemeral storage local to the replica; lost on termination.
- `AZURE_FILE` -- Persistent SMB Azure Files share via an environment storage resource.
- `NFS_AZURE_FILE` -- Persistent NFS Azure Files share via an environment storage resource; requires a VNet-injected environment.
- `SECRET` -- Mounts the app's secrets as files inside the container.

### spec.volumes[].storageName

`string | valueFrom`

Name of the Container App Environment storage resource backing the
volume. Required for AZURE_FILE and NFS_AZURE_FILE volumes; must be
omitted otherwise. Can be a literal name or a reference to an
AzureContainerAppEnvironmentStorage output.

- references: AzureContainerAppEnvironmentStorage (`status.outputs.storage_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerAppEnvironmentStorage, name: <that resource's name>, fieldPath: status.outputs.storage_name}} -- a bare string does not parse

### spec.volumes[].mountOptions

`string`

Comma-separated mount options for the volume (SMB/NFS mounts).
Example: "uid=1000,gid=1000"

### spec.minReplicas

`int32` · optional (explicit presence)

Minimum number of replicas. Set to 0 for scale-to-zero (no cost when idle).

Default: 0
Range: 0-300

- default: `0`
- rule: {"int32":{"lte":300,"gte":0}}

### spec.maxReplicas

`int32` · optional (explicit presence)

Maximum number of replicas the app can scale to.

Default: 10
Range: 1-300

- default: `10`
- rule: {"int32":{"lte":300,"gte":1}}

### spec.cooldownPeriodInSeconds

`int32` · optional (explicit presence)

Scale cooldown period in seconds. After a scale event, the scaler waits
this duration before evaluating scale rules again.

Default: 300 (5 minutes)

- default: `300`
- rule: {"int32":{"gte":1}}

### spec.pollingIntervalInSeconds

`int32` · optional (explicit presence)

KEDA polling interval in seconds. How often scale rules are evaluated.

Default: 30

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.revisionSuffix

`string`

Manual revision suffix. When set, the revision name becomes
"{app-name}--{revision_suffix}". When omitted, Azure auto-generates a suffix.

Useful for canary deployments and explicit revision tracking.

### spec.terminationGracePeriodSeconds

`int32` · optional (explicit presence)

Termination grace period in seconds. How long to wait for containers
to gracefully shut down before forcefully terminating them.

Default: 0 (immediate termination)
Range: 0-600

- default: `0`
- rule: {"int32":{"lte":600,"gte":0}}

### spec.httpScaleRules

`[]AzureContainerAppHttpScaleRule`

HTTP scale rules. Scale based on concurrent HTTP requests.
Most common for web services and APIs.

### spec.httpScaleRules[].name

`string` · required

Rule name. Must be unique within the app's scale rules.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.httpScaleRules[].concurrentRequests

`string` · required

Number of concurrent HTTP requests that triggers scaling.
Must be a positive integer as a string (KEDA convention).

Example: "100" means scale up when more than 100 concurrent requests per replica.

- rule: concurrent_requests must be a positive integer written as a string, e.g. "100"
- rule: {"required":true}

### spec.httpScaleRules[].authentication

`[]AzureContainerAppScaleRuleAuth`

Authentication configuration for the scale rule.
References secrets for KEDA scaler authentication.

### spec.httpScaleRules[].authentication[].secretName

`string` · required

Name of the secret in the app's `secrets` list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.httpScaleRules[].authentication[].triggerParameter

`string`

Scaler-specific trigger parameter name that this secret maps to.
Required for Azure Queue and custom scale rules; optional for HTTP/TCP
rules (their scaler has a single implicit parameter).

Examples:
- Azure Queue: "connection"
- Kafka: "sasl" or "tls"
- Custom: varies by scaler

### spec.tcpScaleRules

`[]AzureContainerAppTcpScaleRule`

TCP scale rules. Scale based on concurrent TCP connections.

### spec.tcpScaleRules[].name

`string` · required

Rule name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tcpScaleRules[].concurrentRequests

`string` · required

Number of concurrent TCP connections that triggers scaling.
Must be a positive integer as a string.

- rule: concurrent_requests must be a positive integer written as a string, e.g. "100"
- rule: {"required":true}

### spec.tcpScaleRules[].authentication

`[]AzureContainerAppScaleRuleAuth`

Authentication configuration for the scale rule.

### spec.tcpScaleRules[].authentication[].secretName

`string` · required

Name of the secret in the app's `secrets` list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tcpScaleRules[].authentication[].triggerParameter

`string`

Scaler-specific trigger parameter name that this secret maps to.
Required for Azure Queue and custom scale rules; optional for HTTP/TCP
rules (their scaler has a single implicit parameter).

Examples:
- Azure Queue: "connection"
- Kafka: "sasl" or "tls"
- Custom: varies by scaler

### spec.azureQueueScaleRules

`[]AzureContainerAppAzureQueueScaleRule`

Azure Queue scale rules. Scale based on Azure Storage Queue depth.

- rule: Azure Queue scale rule authentication entries must set trigger_parameter (typically "connection") so KEDA knows which scaler parameter the secret feeds

### spec.azureQueueScaleRules[].name

`string` · required

Rule name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.azureQueueScaleRules[].queueName

`string` · required

Azure Storage Queue name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.azureQueueScaleRules[].queueLength

`int32`

Queue length threshold. When the queue has more messages than this value,
the app scales up.

- rule: {"int32":{"gte":1}}

### spec.azureQueueScaleRules[].authentication

`[]AzureContainerAppScaleRuleAuth` · required

Authentication configuration. Required for Azure Queue scale rules --
KEDA needs credentials to read the queue depth. Each entry must name
the trigger_parameter the secret maps to (typically "connection").

- rule: {"repeated":{"minItems":"1"}}

### spec.azureQueueScaleRules[].authentication[].secretName

`string` · required

Name of the secret in the app's `secrets` list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.azureQueueScaleRules[].authentication[].triggerParameter

`string`

Scaler-specific trigger parameter name that this secret maps to.
Required for Azure Queue and custom scale rules; optional for HTTP/TCP
rules (their scaler has a single implicit parameter).

Examples:
- Azure Queue: "connection"
- Kafka: "sasl" or "tls"
- Custom: varies by scaler

### spec.customScaleRules

`[]AzureContainerAppCustomScaleRule`

Custom KEDA scale rules. Scale based on any KEDA-supported scaler
(Kafka, Prometheus, Redis, cron, cpu, memory, etc.).

- rule: custom scale rule authentication entries must set trigger_parameter so KEDA knows which scaler parameter the secret feeds

### spec.customScaleRules[].name

`string` · required

Rule name. Lowercase alphanumeric characters, hyphens, and periods.

- rule: scale rule name must be lowercase alphanumeric characters, hyphens, or periods
- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customScaleRules[].customRuleType

`string` · required

KEDA scaler type identifier.
Examples: "kafka", "prometheus", "redis", "cron", "cpu", "memory",
"postgresql", "mysql", "rabbitmq", "azure-servicebus"

- rule: custom_rule_type must be a KEDA scaler Azure Container Apps supports, e.g. kafka, prometheus, redis, cron, cpu, memory, azure-servicebus, rabbitmq (see keda.sh/docs/scalers)
- rule: {"required":true}

### spec.customScaleRules[].metadata

`map<string, string>` · required

Scaler-specific metadata. Keys and values depend on the scaler type.

Example for cron: {"timezone": "UTC", "start": "0 8 * * 1-5", "end": "0 18 * * 1-5", "desiredReplicas": "5"}
Example for cpu: {"type": "Utilization", "value": "70"}
Example for kafka: {"bootstrapServers": "kafka:9092", "consumerGroup": "my-group", "topic": "my-topic", "lagThreshold": "100"}

- rule: {"map":{"minPairs":"1"}}

### spec.customScaleRules[].authentication

`[]AzureContainerAppScaleRuleAuth`

Authentication configuration for the scale rule. Each entry must name
the trigger_parameter the secret maps to.

### spec.customScaleRules[].authentication[].secretName

`string` · required

Name of the secret in the app's `secrets` list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customScaleRules[].authentication[].triggerParameter

`string`

Scaler-specific trigger parameter name that this secret maps to.
Required for Azure Queue and custom scale rules; optional for HTTP/TCP
rules (their scaler has a single implicit parameter).

Examples:
- Azure Queue: "connection"
- Kafka: "sasl" or "tls"
- Custom: varies by scaler

### spec.customScaleRules[].identityId

`string | valueFrom`

Managed identity KEDA uses to execute the scale rule (workload
identity for the scaler instead of connection-string secrets).
The literal "System" (the app's system-assigned identity) or a User
Assigned Identity ARM resource ID -- reference the identity that is
also in the app's identity block so the scaler and the workload share
one principal. With an identity set, Azure scalers need no
`authentication` entries at all: grant the identity the data-plane
read role on the scaled resource (e.g. Azure Service Bus Data
Receiver for queue depth) and leave the connection secret out of the
app entirely.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.secrets

`[]AzureContainerAppSecret`

Secrets available to the app. Referenced by name in container env vars,
registry password_secret_name, scale-rule authentication, and
SECRET-type volumes.

Secrets can be plain-text values or Key Vault references. Key Vault
references require a managed identity.

- rule: a secret takes either a plain-text value or a key_vault_secret_id, not both
- rule: key_vault_secret_id requires identity ("System" or a user-assigned identity ARM ID) so the app can read the vault, and identity is only meaningful with key_vault_secret_id

### spec.secrets[].name

`string` · required

Secret name. Lowercase alphanumeric or hyphens, max 253 characters.
This name is used to reference the secret elsewhere in the spec.

- rule: secret name must be lowercase alphanumeric or hyphens, start and end with alphanumeric
- rule: {"required":true,"string":{"minLen":"1","maxLen":"253"}}

### spec.secrets[].value

`string` · sensitive

Plain-text secret value. Mutually exclusive with key_vault_secret_id.

### spec.secrets[].keyVaultSecretId

`string`

Key Vault secret URI. Format:
https://{vault-name}.vault.azure.net/secrets/{secret-name}
(versionless -- tracks the latest version) or with an explicit version:
https://{vault-name}.vault.azure.net/secrets/{secret-name}/{version}

Requires `identity` for Key Vault access.
Mutually exclusive with `value`.

### spec.secrets[].identity

`string`

Identity for Key Vault access. Required when key_vault_secret_id is
set (and only meaningful then).

Value is either:
- "System": use the app's system-assigned managed identity
- A User Assigned Identity ARM resource ID

### spec.registries

`[]AzureContainerAppRegistry`

Private container registry credentials. Required to pull images from
private registries (ACR, Docker Hub, GitHub Container Registry, etc.).

Each registry authenticates via username/password (referencing a secret)
or managed identity.

- rule: a registry authenticates with either a managed identity or a username + password_secret_name pair -- exactly one mode, and username and password_secret_name always travel together

### spec.registries[].server

`string | valueFrom` · required

Registry server hostname. Reference an AzureContainerRegistry's
login_server output (the default wiring -- the same shape the
Container Instance's registry credential carries) or pass a literal
such as "ghcr.io" or "docker.io".

- references: AzureContainerRegistry (`status.outputs.login_server`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerRegistry, name: <that resource's name>, fieldPath: status.outputs.login_server}} -- a bare string does not parse

### spec.registries[].username

`string`

Registry username. Required with password_secret_name.

### spec.registries[].passwordSecretName

`string`

Secret name containing the registry password.
Must reference a secret defined in the app's `secrets` list.
Required with username.

### spec.registries[].identity

`string`

Managed identity for registry authentication.
Value is either "System" (system-assigned) or a User Assigned Identity
ARM resource ID. Alternative to username/password.

### spec.ingress

`AzureContainerAppIngress`

Ingress configuration. When set, the app is accessible via HTTP/TCP.
When omitted, the app is only accessible from within the environment
via service discovery (app name).

- rule: exposed_port is only valid with transport TCP -- HTTP transports always listen on 443/80

### spec.ingress.externalEnabled

`bool` · optional (explicit presence)

Whether the app is accessible from outside the environment.

true: App is accessible from the internet (or VNet if environment is internal).
false: App is only accessible from within the environment.

Default: false

- default: `false`

### spec.ingress.targetPort

`int32`

Target port on the container to route traffic to. Range: 1-65535.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.ingress.exposedPort

`int32` · optional (explicit presence)

Exposed port for TCP transport. Range: 1-65535.
Only applicable when transport is TCP.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.ingress.transport

`enum`

Transport protocol. Unspecified deploys AUTO (detects HTTP/1.1 vs
HTTP/2). Use HTTP2 for gRPC services and TCP (with exposed_port) for
raw TCP protocols.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_ingress_transport_unspecified` -- Not specified -- deploys AUTO.
- `AUTO` -- Auto-detect HTTP/1.1 vs HTTP/2.
- `HTTP` -- HTTP/1.1 only.
- `HTTP2` -- HTTP/2 (gRPC and long-lived streaming connections).
- `TCP` -- Raw TCP -- pair with exposed_port for the externally visible port.

### spec.ingress.allowInsecureConnections

`bool` · optional (explicit presence)

Allow insecure (HTTP) connections. When false, only HTTPS is accepted
and HTTP requests are redirected to HTTPS.

Default: false (HTTPS only)

- default: `false`

### spec.ingress.clientCertificateMode

`enum`

Client certificate mode for mTLS. Unspecified leaves Azure's default
behavior (no client certificate requirement).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_ingress_client_certificate_mode_unspecified` -- Not specified -- Azure's default behavior (no client certificate requirement).
- `ACCEPT` -- Accept client certificates when presented, but do not require them.
- `REQUIRE` -- Require a client certificate on every request.
- `IGNORE` -- Ignore client certificates entirely.

### spec.ingress.trafficWeight

`[]AzureContainerAppTrafficWeight` · required

Traffic weight distribution across revisions.
At least one traffic weight is required when ingress is configured.

For SINGLE revision mode, one weight with latest_revision=true and
percentage=100. For MULTIPLE revision mode, weights split traffic
across named revisions.

Total percentage across all weights must equal 100.

- rule: {"repeated":{"minItems":"1"}}
- rule: each traffic weight targets exactly one thing: set latest_revision: true OR name a revision_suffix, never both and never neither

### spec.ingress.trafficWeight[].latestRevision

`bool` · optional (explicit presence)

Route traffic to the latest revision.
Mutually exclusive with revision_suffix.

- default: `false`

### spec.ingress.trafficWeight[].revisionSuffix

`string`

Target a specific revision by suffix.
The full revision name is "{app-name}--{revision_suffix}".
Mutually exclusive with latest_revision.

### spec.ingress.trafficWeight[].percentage

`int32`

Percentage of traffic to route to this revision.
All weights must sum to 100. Range: 0-100.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.ingress.trafficWeight[].label

`string`

Optional label prefix for this traffic weight.
When set, the revision gets an additional FQDN: {label}.{app-fqdn}

### spec.ingress.ipSecurityRestrictions

`[]AzureContainerAppIpSecurityRestriction`

IP security restrictions. Allow or deny traffic from specific IP ranges.
All rules must share the same action (Azure evaluates them as one
allowlist or one denylist).

### spec.ingress.ipSecurityRestrictions[].name

`string` · required

Rule name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ingress.ipSecurityRestrictions[].action

`enum`

Action to take for matching traffic. All rules on an ingress must
share the same action.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_ip_restriction_action_unspecified` -- Not specified -- invalid; pick ALLOW or DENY.
- `ALLOW` -- Permit traffic from the IP range.
- `DENY` -- Block traffic from the IP range.

### spec.ingress.ipSecurityRestrictions[].ipAddressRange

`string` · required

IP address or CIDR range.
Examples: "203.0.113.0/24", "10.0.0.1"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ingress.ipSecurityRestrictions[].description

`string`

Optional description of the rule.

### spec.ingress.cors

`AzureContainerAppCors`

CORS (Cross-Origin Resource Sharing) rules for browser-based clients.

### spec.ingress.cors.allowedOrigins

`[]string` · required

Allowed origins. At least one origin is required.
Use "*" to allow all origins (not recommended for production).
Examples: "https://example.com", "https://*.contoso.com"

- rule: {"repeated":{"minItems":"1"}}

### spec.ingress.cors.allowedHeaders

`[]string`

Allowed HTTP headers.
Examples: "Content-Type", "Authorization", "X-Custom-Header"

### spec.ingress.cors.allowedMethods

`[]string`

Allowed HTTP methods.
Examples: "GET", "POST", "PUT", "DELETE", "OPTIONS"

### spec.ingress.cors.exposedHeaders

`[]string`

Headers to expose to the browser.

### spec.ingress.cors.maxAgeInSeconds

`int32` · optional (explicit presence)

Maximum time in seconds that preflight results can be cached.

- rule: {"int32":{"gte":0}}

### spec.ingress.cors.allowCredentialsEnabled

`bool` · optional (explicit presence)

Whether to include credentials (cookies, authorization headers)
in CORS requests.

Default: false

- default: `false`

### spec.dapr

`AzureContainerAppDapr`

Dapr sidecar configuration. When set, a Dapr sidecar is injected
alongside the app container, enabling Dapr building blocks
(service invocation, pub/sub, state management, etc.).

### spec.dapr.appId

`string` · required

Dapr application identifier. Used for service discovery and invocation.
Other Dapr-enabled apps invoke this app using this ID, and environment
Dapr components scope themselves to it.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dapr.appPort

`int32` · optional (explicit presence)

Application port that the Dapr sidecar communicates with.
This is the port your application listens on.

### spec.dapr.appProtocol

`enum`

Protocol used for Dapr-to-app communication. Unspecified deploys
DAPR_HTTP.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_dapr_protocol_unspecified` -- Not specified -- deploys DAPR_HTTP.
- `DAPR_HTTP` -- HTTP/1.1 (the default; Azure protocol "http").
- `DAPR_GRPC` -- gRPC, for gRPC-based applications (Azure protocol "grpc").

### spec.identity

`AzureContainerAppIdentity`

Managed identity configuration for the app. Enables the app to
authenticate with Azure services (Key Vault, ACR, Storage, etc.)
without managing credentials.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the app's lifecycle), USER_ASSIGNED (bring
identities from user_assigned_identity_ids, shareable across
resources), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the app's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the app exists.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned principal and user-assigned identities.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type includes USER_ASSIGNED. Each entry references
an AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form Azure resource tags applied to the app, merged over the
platform's metadata-derived tags (user tags win on key collision) --
the hooks for cost allocation, chargeback reports, and Azure Policy
governance rules that filter or group by them. Updatable in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureContainerApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.container_app_id` | `string` | The Azure Resource Manager ID of the Container App. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.App/containerApps/{name} |
| `status.outputs.container_app_name` | `string` | The name of the Container App. |
| `status.outputs.latest_revision_name` | `string` | The name of the latest revision. Format: {app-name}--{suffix} Useful for CD pipelines to verify which revision is active after deployment. |
| `status.outputs.latest_revision_fqdn` | `string` | The FQDN of the latest revision. This FQDN points directly to the latest revision, bypassing traffic splitting. |
| `status.outputs.outbound_ip_addresses` | `[]string` | Outbound IP addresses used by the Container App for egress traffic. Use these to configure firewall allowlists on external services (databases, APIs, etc.) that the app connects to. |
| `status.outputs.ingress_fqdn` | `string` | The ingress FQDN of the Container App. This is the primary user-facing endpoint. Only populated when ingress is configured. In SINGLE revision mode, this is the same as latest_revision_fqdn. In MULTIPLE revision mode, this is the app's main FQDN that distributes traffic according to the configured traffic weights. |
| `status.outputs.custom_domain_verification_id` | `string` | The value Azure expects in the asuid.{domain} TXT record when binding a custom domain to this app -- publish it before creating the binding. |
| `status.outputs.identity_principal_id` | `string` | The principal (object) ID of the app's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal roles (AcrPull, Key Vault Secrets User, etc.) to let the app authenticate keylessly. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.containerAppEnvironmentId` | AzureContainerAppEnvironment | `status.outputs.environment_id` |
| `spec.volumes[].storageName` | AzureContainerAppEnvironmentStorage | `status.outputs.storage_name` |
| `spec.customScaleRules[].identityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.registries[].server` | AzureContainerRegistry | `status.outputs.login_server` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureContainerAppCustomDomain | `spec.containerAppId` | `status.outputs.container_app_id` |

## See Also

- [Overview](../README.md)

# AzureContainerAppJob

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureContainerAppJobSpec** defines the configuration for creating an
Azure Container App Job (Microsoft.App/jobs), a run-to-completion
containerized workload inside an Azure Container Apps Managed Environment.

Where a Container App runs continuously and serves traffic, a Job starts,
does its work, and exits: database migrations, batch processing, report
generation, queue draining, CI runners. Each execution runs the job's
template to completion (or until `replica_timeout_in_seconds`).

**Trigger models** (exactly one per job):

Manual (`manual_trigger`):
  - Executions start on demand (CLI, SDK, or another system calling the
    job's start API)
  - The on-demand batch model

Schedule (`schedule_trigger`):
  - Executions start on a cron expression (Cron format, UTC)
  - The nightly-report / periodic-cleanup model

Event (`event_trigger`):
  - Executions start when a KEDA scaler fires (queue depth, Kafka lag,
    any KEDA source) -- the scale block controls execution fan-out
  - The queue-worker model: each message batch gets its own execution

**Parallelism**: each trigger carries `parallelism` (how many replicas an
execution runs simultaneously) and `replica_completion_count` (how many
must succeed for the execution to count as successful) -- the
indexed-parallel-work model.

**Containers**: the job's template mirrors a Container App's containers
(image, resources, env, probes, volume mounts) minus continuous-serving
concerns -- there is no ingress, no revision model, and no scale-to-zero
(executions ARE the scaling unit).

**ForceNew fields** (changing these destroys and recreates the job):
`job_name`, `region`, `resource_group`, `container_app_environment_id`,
and the choice of trigger block.

**Referenced by**: None (leaf workload resource)

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppJob
metadata:
  name: test-job
spec:
  region: eastus
  resource_group:
    value: test-rg
  job_name: test-batch-job
  container_app_environment_id:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.App/managedEnvironments/test-env
  replica_timeout_in_seconds: 1800
  replica_retry_limit: 2
  workload_profile_name: dedicated-d4
  containers:
    - name: worker
      image: mcr.microsoft.com/k8se/quickstart-jobs:latest
      cpu: 0.5
      memory: "1Gi"
      command: ["/bin/process"]
      args: ["--batch-size", "100"]
      env:
        - name: MODE
          value: batch
        - name: QUEUE_CONN
          secret_name: queue-conn
      startup_probe:
        transport: TCP_SOCKET
        port: 8080
        failure_count_threshold: 60
      volume_mounts:
        - name: work-data
          path: /work
  init_containers:
    - name: fetch-config
      image: mcr.microsoft.com/k8se/quickstart-jobs:latest
      command: ["/bin/fetch-config"]
  volumes:
    - name: work-data
      storage_type: AZURE_FILE
      storage_name:
        value: test-env-storage
  event_trigger:
    parallelism: 2
    replica_completion_count: 2
    scale:
      max_executions: 20
      min_executions: 0
      polling_interval_in_seconds: 60
      rules:
        - name: queue-depth
          custom_rule_type: azure-queue
          metadata:
            queueName: work
            queueLength: "5"
          authentication:
            - secret_name: queue-conn
              trigger_parameter: connection
  secrets:
    - name: queue-conn
      value: DefaultEndpointsProtocol=https;AccountName=test
    - name: db-password
      key_vault_secret_id: https://test-vault.vault.azure.net/secrets/db-password
      identity: System
  registries:
    - server:
        value: testregistry.azurecr.io
      identity: System
  identity:
    type: SYSTEM_ASSIGNED
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.jobName` | `string` | yes |  |  |
| `spec.containerAppEnvironmentId` | `string \| valueFrom` | yes |  | AzureContainerAppEnvironment (`status.outputs.environment_id`) |
| `spec.replicaTimeoutInSeconds` | `int32` | yes |  |  |
| `spec.replicaRetryLimit` | `int32` |  |  |  |
| `spec.workloadProfileName` | `string` |  |  |  |
| `spec.containers` | `[]AzureContainerAppJobContainer` | yes |  |  |
| `spec.containers[].name` | `string` | yes |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].cpu` | `double` |  |  |  |
| `spec.containers[].memory` | `string` | yes |  |  |
| `spec.containers[].env` | `[]AzureContainerAppJobEnvVar` |  |  |  |
| `spec.containers[].env[].name` | `string` | yes |  |  |
| `spec.containers[].env[].value` | `string` |  |  |  |
| `spec.containers[].env[].secretName` | `string` |  |  |  |
| `spec.containers[].command` | `[]string` |  |  |  |
| `spec.containers[].args` | `[]string` |  |  |  |
| `spec.containers[].livenessProbe` | `AzureContainerAppJobProbe` |  |  |  |
| `spec.containers[].livenessProbe.transport` | `enum` |  |  |  |
| `spec.containers[].livenessProbe.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.path` | `string` |  |  |  |
| `spec.containers[].livenessProbe.host` | `string` |  |  |  |
| `spec.containers[].livenessProbe.headers` | `[]AzureContainerAppJobProbeHeader` |  |  |  |
| `spec.containers[].livenessProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].livenessProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].livenessProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].readinessProbe` | `AzureContainerAppJobProbe` |  |  |  |
| `spec.containers[].readinessProbe.transport` | `enum` |  |  |  |
| `spec.containers[].readinessProbe.port` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.path` | `string` |  |  |  |
| `spec.containers[].readinessProbe.host` | `string` |  |  |  |
| `spec.containers[].readinessProbe.headers` | `[]AzureContainerAppJobProbeHeader` |  |  |  |
| `spec.containers[].readinessProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].readinessProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].readinessProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].readinessProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].readinessProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].readinessProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].startupProbe` | `AzureContainerAppJobProbe` |  |  |  |
| `spec.containers[].startupProbe.transport` | `enum` |  |  |  |
| `spec.containers[].startupProbe.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.path` | `string` |  |  |  |
| `spec.containers[].startupProbe.host` | `string` |  |  |  |
| `spec.containers[].startupProbe.headers` | `[]AzureContainerAppJobProbeHeader` |  |  |  |
| `spec.containers[].startupProbe.headers[].name` | `string` | yes |  |  |
| `spec.containers[].startupProbe.headers[].value` | `string` | yes |  |  |
| `spec.containers[].startupProbe.initialDelayInSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.intervalSeconds` | `int32` |  | `10` |  |
| `spec.containers[].startupProbe.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.containers[].startupProbe.failureCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].startupProbe.successCountThreshold` | `int32` |  | `3` |  |
| `spec.containers[].volumeMounts` | `[]AzureContainerAppJobVolumeMount` |  |  |  |
| `spec.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].path` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.initContainers` | `[]AzureContainerAppJobInitContainer` |  |  |  |
| `spec.initContainers[].name` | `string` | yes |  |  |
| `spec.initContainers[].image` | `string` | yes |  |  |
| `spec.initContainers[].cpu` | `double` |  |  |  |
| `spec.initContainers[].memory` | `string` |  |  |  |
| `spec.initContainers[].env` | `[]AzureContainerAppJobEnvVar` |  |  |  |
| `spec.initContainers[].env[].name` | `string` | yes |  |  |
| `spec.initContainers[].env[].value` | `string` |  |  |  |
| `spec.initContainers[].env[].secretName` | `string` |  |  |  |
| `spec.initContainers[].command` | `[]string` |  |  |  |
| `spec.initContainers[].args` | `[]string` |  |  |  |
| `spec.initContainers[].volumeMounts` | `[]AzureContainerAppJobVolumeMount` |  |  |  |
| `spec.initContainers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.initContainers[].volumeMounts[].path` | `string` | yes |  |  |
| `spec.initContainers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.volumes` | `[]AzureContainerAppJobVolume` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].storageType` | `enum` |  |  |  |
| `spec.volumes[].storageName` | `string \| valueFrom` |  |  | AzureContainerAppEnvironmentStorage (`status.outputs.storage_name`) |
| `spec.volumes[].mountOptions` | `string` |  |  |  |
| `spec.manualTrigger` | `AzureContainerAppJobManualTrigger` |  |  |  |
| `spec.manualTrigger.parallelism` | `int32` |  | `1` |  |
| `spec.manualTrigger.replicaCompletionCount` | `int32` |  | `1` |  |
| `spec.scheduleTrigger` | `AzureContainerAppJobScheduleTrigger` |  |  |  |
| `spec.scheduleTrigger.cronExpression` | `string` | yes |  |  |
| `spec.scheduleTrigger.parallelism` | `int32` |  | `1` |  |
| `spec.scheduleTrigger.replicaCompletionCount` | `int32` |  | `1` |  |
| `spec.eventTrigger` | `AzureContainerAppJobEventTrigger` |  |  |  |
| `spec.eventTrigger.parallelism` | `int32` |  | `1` |  |
| `spec.eventTrigger.replicaCompletionCount` | `int32` |  | `1` |  |
| `spec.eventTrigger.scale` | `AzureContainerAppJobEventScale` |  |  |  |
| `spec.eventTrigger.scale.maxExecutions` | `int32` |  | `100` |  |
| `spec.eventTrigger.scale.minExecutions` | `int32` |  | `0` |  |
| `spec.eventTrigger.scale.pollingIntervalInSeconds` | `int32` |  | `30` |  |
| `spec.eventTrigger.scale.rules` | `[]AzureContainerAppJobEventScaleRule` |  |  |  |
| `spec.eventTrigger.scale.rules[].name` | `string` | yes |  |  |
| `spec.eventTrigger.scale.rules[].customRuleType` | `string` | yes |  |  |
| `spec.eventTrigger.scale.rules[].metadata` | `map<string, string>` | yes |  |  |
| `spec.eventTrigger.scale.rules[].authentication` | `[]AzureContainerAppJobScaleRuleAuth` |  |  |  |
| `spec.eventTrigger.scale.rules[].authentication[].secretName` | `string` | yes |  |  |
| `spec.eventTrigger.scale.rules[].authentication[].triggerParameter` | `string` | yes |  |  |
| `spec.eventTrigger.scale.rules[].identityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.secrets` | `[]AzureContainerAppJobSecret` |  |  |  |
| `spec.secrets[].name` | `string` | yes |  |  |
| `spec.secrets[].value` | `string` (sensitive) |  |  |  |
| `spec.secrets[].keyVaultSecretId` | `string` |  |  |  |
| `spec.secrets[].identity` | `string` |  |  |  |
| `spec.registries` | `[]AzureContainerAppJobRegistry` |  |  |  |
| `spec.registries[].server` | `string \| valueFrom` | yes |  | AzureContainerRegistry (`status.outputs.login_server`) |
| `spec.registries[].username` | `string` |  |  |  |
| `spec.registries[].passwordSecretName` | `string` |  |  |  |
| `spec.registries[].identity` | `string` |  |  |  |
| `spec.identity` | `AzureContainerAppJobIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Container App Job will be created.
Must match the environment's region.
Examples: "eastus", "westus2", "westeurope".

**ForceNew**: Changing this destroys and recreates the job.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Container App Job will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: Changing this destroys and recreates the job.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.jobName

`string` · required

The name of the Container App Job.
Lowercase alphanumeric characters and hyphens; must start and end with
an alphanumeric character; no consecutive hyphens; at most 32
characters.

**ForceNew**: Changing this destroys and recreates the job.

- rule: Job name must be lowercase alphanumeric characters or hyphens, start and end with an alphanumeric character, and contain no consecutive hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.containerAppEnvironmentId

`string | valueFrom` · required

The Container App Environment where this job will run.
The environment provides networking, logging, and compute capacity.

**ForceNew**: Changing this destroys and recreates the job.

- references: AzureContainerAppEnvironment (`status.outputs.environment_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerAppEnvironment, name: <that resource's name>, fieldPath: status.outputs.environment_id}} -- a bare string does not parse

### spec.replicaTimeoutInSeconds

`int32` · required

Maximum seconds a replica may run before Azure terminates it -- the
job's hard deadline. Size it to the slowest legitimate execution;
a replica killed by this timeout counts as failed (and retries per
replica_retry_limit).

- rule: {"required":true,"int32":{"gte":1}}

### spec.replicaRetryLimit

`int32` · optional (explicit presence)

How many times a failed replica is retried before the execution is
marked failed. 0 means no retries. Omit to use Azure's default (0).

- rule: {"int32":{"gte":0}}

### spec.workloadProfileName

`string`

The workload profile name to run this job on.
References a profile defined in the Container App Environment.

Omit to use the default Consumption (serverless) profile.

### spec.containers

`[]AzureContainerAppJobContainer` · required

Main containers for the job. At least one container is required.
All containers in a replica run to completion.

- rule: {"repeated":{"minItems":"1"}}
- rule: liveness probes do not support success_count_threshold, and failure_count_threshold must be between 1 and 30
- rule: readiness probe failure_count_threshold must be between 1 and 48
- rule: startup probes do not support success_count_threshold

### spec.containers[].name

`string` · required

Container name. Must be unique within the job.
Lowercase alphanumeric, hyphens, or dots; starts and ends with an
alphanumeric character; max 46 characters.

- rule: container name must be lowercase alphanumeric, hyphens, or dots, starting and ending with an alphanumeric character
- rule: {"required":true,"string":{"minLen":"1","maxLen":"46"}}

### spec.containers[].image

`string` · required

Container image in repository:tag format.
Examples: "mcr.microsoft.com/k8se/quickstart-jobs:latest", "myregistry.azurecr.io/batch:v1.2.3"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].cpu

`double`

CPU allocation in vCPU cores.
Examples: 0.25, 0.5, 1.0, 2.0. CPU and memory scale together
(0.25 vCPU pairs with 0.5Gi, and so on).

- rule: {"double":{"gte":0.1}}

### spec.containers[].memory

`string` · required

Memory allocation in Gi format.
Examples: "0.5Gi", "1Gi", "2Gi", "4Gi". Must match the CPU allocation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env

`[]AzureContainerAppJobEnvVar`

Environment variables for the container.
Each variable is a literal value or a reference to a secret by name.

- rule: an environment variable takes either a literal value or a secret_name, not both -- move the literal into the job's secrets list to reference it

### spec.containers[].env[].name

`string` · required

Environment variable name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env[].value

`string`

Literal value. Mutually exclusive with secret_name.

### spec.containers[].env[].secretName

`string`

Reference to a secret name defined in the job's `secrets` list.
When set, the env var's value comes from the secret.

### spec.containers[].command

`[]string`

Container command (entrypoint override).

### spec.containers[].args

`[]string`

Container arguments.

### spec.containers[].livenessProbe

`AzureContainerAppJobProbe`

Liveness probe. If the probe fails, the replica is restarted.
Liveness probes never carry success_count_threshold, and their
failure_count_threshold tops out at 30.

### spec.containers[].livenessProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_job_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
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

### spec.containers[].livenessProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].livenessProbe.headers

`[]AzureContainerAppJobProbeHeader`

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

`AzureContainerAppJobProbe`

Readiness probe. The only probe type that supports
success_count_threshold; its failure_count_threshold tops out at 48.

### spec.containers[].readinessProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_job_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
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

### spec.containers[].readinessProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].readinessProbe.headers

`[]AzureContainerAppJobProbeHeader`

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

`AzureContainerAppJobProbe`

Startup probe. Disables liveness/readiness checks until the container
has finished starting; failure_count_threshold tops out at 240.

### spec.containers[].startupProbe.transport

`enum`

Probe transport type: TCP_SOCKET (port connectivity check), HTTP_GET
(GET request), or HTTPS_GET (GET over TLS).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_job_probe_transport_unspecified` -- Not specified -- invalid; pick TCP_SOCKET, HTTP_GET, or HTTPS_GET.
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

### spec.containers[].startupProbe.host

`string`

Hostname for the probe request. When omitted, defaults to the container IP.

### spec.containers[].startupProbe.headers

`[]AzureContainerAppJobProbeHeader`

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

`[]AzureContainerAppJobVolumeMount`

Volume mounts for the container. References volumes defined in the
spec's `volumes` field by name.

### spec.containers[].volumeMounts[].name

`string` · required

Name of the volume to mount. Must match a volume name in the spec's `volumes` field.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumeMounts[].path

`string` · required

Absolute path inside the container where the volume is mounted.
Example: "/data", "/mnt/work"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumeMounts[].subPath

`string`

Sub-path within the volume to mount. When omitted, the entire volume is mounted.

### spec.initContainers

`[]AzureContainerAppJobInitContainer`

Init containers run to completion before main containers start.

### spec.initContainers[].name

`string` · required

Container name. Must be unique within the job (across both containers
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

### spec.initContainers[].memory

`string` · optional (explicit presence)

Memory allocation in Gi format. Optional for init containers.

### spec.initContainers[].env

`[]AzureContainerAppJobEnvVar`

Environment variables for the init container.

- rule: an environment variable takes either a literal value or a secret_name, not both -- move the literal into the job's secrets list to reference it

### spec.initContainers[].env[].name

`string` · required

Environment variable name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].env[].value

`string`

Literal value. Mutually exclusive with secret_name.

### spec.initContainers[].env[].secretName

`string`

Reference to a secret name defined in the job's `secrets` list.
When set, the env var's value comes from the secret.

### spec.initContainers[].command

`[]string`

Container command (entrypoint override).

### spec.initContainers[].args

`[]string`

Container arguments.

### spec.initContainers[].volumeMounts

`[]AzureContainerAppJobVolumeMount`

Volume mounts for the init container.

### spec.initContainers[].volumeMounts[].name

`string` · required

Name of the volume to mount. Must match a volume name in the spec's `volumes` field.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumeMounts[].path

`string` · required

Absolute path inside the container where the volume is mounted.
Example: "/data", "/mnt/work"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumeMounts[].subPath

`string`

Sub-path within the volume to mount. When omitted, the entire volume is mounted.

### spec.volumes

`[]AzureContainerAppJobVolume`

Volumes available to containers in this job.
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

- `azure_container_app_job_volume_storage_type_unspecified` -- Not specified -- deploys EMPTY_DIR (ephemeral scratch space).
- `EMPTY_DIR` -- Ephemeral storage local to the replica; lost on termination.
- `AZURE_FILE` -- Persistent SMB Azure Files share via an environment storage resource.
- `NFS_AZURE_FILE` -- Persistent NFS Azure Files share via an environment storage resource; requires a VNet-injected environment.
- `SECRET` -- Mounts the job's secrets as files inside the container.

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

### spec.manualTrigger

`AzureContainerAppJobManualTrigger`

Start executions on demand (CLI, SDK, or the job start API).

**ForceNew**: Switching trigger types destroys and recreates the job.

### spec.manualTrigger.parallelism

`int32` · optional (explicit presence)

How many replicas an execution runs simultaneously.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.manualTrigger.replicaCompletionCount

`int32` · optional (explicit presence)

How many replicas must complete successfully for the execution to
count as successful.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.scheduleTrigger

`AzureContainerAppJobScheduleTrigger`

Start executions on a cron schedule (UTC).

**ForceNew**: Switching trigger types destroys and recreates the job.

### spec.scheduleTrigger.cronExpression

`string` · required

Cron expression controlling when executions start (UTC).
Standard five-field Cron format.
Examples: "0 2 * * *" (02:00 daily), "*/15 * * * *" (every 15 minutes).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.scheduleTrigger.parallelism

`int32` · optional (explicit presence)

How many replicas an execution runs simultaneously.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.scheduleTrigger.replicaCompletionCount

`int32` · optional (explicit presence)

How many replicas must complete successfully for the execution to
count as successful.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.eventTrigger

`AzureContainerAppJobEventTrigger`

Start executions when a KEDA scaler fires (queue depth, Kafka lag,
any KEDA source).

**ForceNew**: Switching trigger types destroys and recreates the job.

### spec.eventTrigger.parallelism

`int32` · optional (explicit presence)

How many replicas an execution runs simultaneously.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.eventTrigger.replicaCompletionCount

`int32` · optional (explicit presence)

How many replicas must complete successfully for the execution to
count as successful.

Default: 1

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.eventTrigger.scale

`AzureContainerAppJobEventScale`

The scale contract: how event pressure fans out into executions.

### spec.eventTrigger.scale.maxExecutions

`int32` · optional (explicit presence)

Maximum number of executions running at once.

Default: 100

- default: `100`
- rule: {"int32":{"gte":1}}

### spec.eventTrigger.scale.minExecutions

`int32` · optional (explicit presence)

Minimum number of executions kept running.

Default: 0

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.eventTrigger.scale.pollingIntervalInSeconds

`int32` · optional (explicit presence)

How often (seconds) KEDA polls the event source.

Default: 30

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.eventTrigger.scale.rules

`[]AzureContainerAppJobEventScaleRule`

The KEDA scale rules that trigger executions.

### spec.eventTrigger.scale.rules[].name

`string` · required

Rule name. Lowercase alphanumeric characters, hyphens, and periods.

- rule: scale rule name must be lowercase alphanumeric characters, hyphens, or periods
- rule: {"required":true,"string":{"minLen":"1"}}

### spec.eventTrigger.scale.rules[].customRuleType

`string` · required

KEDA scaler type identifier.
Examples: "azure-queue", "azure-servicebus", "kafka", "rabbitmq"

- rule: custom_rule_type must be a KEDA scaler Azure Container Apps supports, e.g. azure-queue, azure-servicebus, kafka, rabbitmq (see keda.sh/docs/scalers)
- rule: {"required":true}

### spec.eventTrigger.scale.rules[].metadata

`map<string, string>` · required

Scaler-specific metadata. Keys and values depend on the scaler type.
Example for azure-queue: {"queueName": "work", "queueLength": "5"}

- rule: {"map":{"minPairs":"1"}}

### spec.eventTrigger.scale.rules[].authentication

`[]AzureContainerAppJobScaleRuleAuth`

Authentication configuration for the scale rule -- maps a secret to
the scaler parameter it feeds.

### spec.eventTrigger.scale.rules[].authentication[].secretName

`string` · required

Name of the secret in the job's `secrets` list.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.eventTrigger.scale.rules[].authentication[].triggerParameter

`string` · required

Scaler-specific trigger parameter name that this secret maps to
(e.g. "connection" for azure-queue).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.eventTrigger.scale.rules[].identityId

`string | valueFrom`

Managed identity KEDA uses to execute the scale rule (workload
identity for the scaler instead of connection-string secrets).
The literal "System" (the job's system-assigned identity) or a User
Assigned Identity ARM resource ID -- reference the identity that is
also in the job's identity block so the scaler and the workload share
one principal. With an identity set, Azure scalers need no
`authentication` entries at all: grant the identity the data-plane
read role on the scaled resource (e.g. Azure Service Bus Data
Receiver for queue depth) and leave the connection secret out of the
job entirely.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.secrets

`[]AzureContainerAppJobSecret`

Secrets available to the job. Referenced by name in container env vars,
registry password_secret_name, and event-trigger scale-rule
authentication.

- rule: a secret takes either a plain-text value or a key_vault_secret_id, not both
- rule: key_vault_secret_id requires identity ("System" or a user-assigned identity ARM ID) so the job can read the vault, and identity is only meaningful with key_vault_secret_id

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

Key Vault secret URI (versionless to track the latest version, or
with an explicit version pinned).

Requires `identity` for Key Vault access.
Mutually exclusive with `value`.

### spec.secrets[].identity

`string`

Identity for Key Vault access. Required when key_vault_secret_id is
set (and only meaningful then).

Value is either:
- "System": use the job's system-assigned managed identity
- A User Assigned Identity ARM resource ID

### spec.registries

`[]AzureContainerAppJobRegistry`

Private container registry credentials. Required to pull images from
private registries (ACR, Docker Hub, GitHub Container Registry, etc.).

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
Must reference a secret defined in the job's `secrets` list.
Required with username.

### spec.registries[].identity

`string`

Managed identity for registry authentication.
Value is either "System" (system-assigned) or a User Assigned Identity
ARM resource ID. Alternative to username/password.

### spec.identity

`AzureContainerAppJobIdentity`

Managed identity configuration for the job. Enables the job to
authenticate with Azure services (Key Vault, ACR, Storage, etc.)
without managing credentials.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the job's lifecycle), USER_ASSIGNED (bring
identities from user_assigned_identity_ids, shareable across
resources), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_job_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the job's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the job exists.
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

Free-form Azure resource tags applied to the job, merged over the
platform's metadata-derived tags (user tags win on key collision).
Updatable in place.

## Validation Rules

- `job_exactly_one_trigger`: a job has exactly one trigger: manual_trigger (on-demand), schedule_trigger (cron), or event_trigger (KEDA scaler)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureContainerAppJob, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.job_id` | `string` | The Azure Resource Manager ID of the Container App Job. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.App/jobs/{name} Use this to start manual executions (az containerapp job start) and for programmatic access. |
| `status.outputs.job_name` | `string` | The name of the Container App Job. |
| `status.outputs.event_stream_endpoint` | `string` | The endpoint streaming the job's execution events (started, running, succeeded, failed) -- the hook for execution monitoring. |
| `status.outputs.outbound_ip_addresses` | `[]string` | Outbound IP addresses used by the job's replicas for egress traffic. Use these to configure firewall allowlists on external services the job connects to. |
| `status.outputs.identity_principal_id` | `string` | The principal (object) ID of the job's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal roles (AcrPull, Storage Queue Data Reader, etc.) to let the job authenticate keylessly. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.containerAppEnvironmentId` | AzureContainerAppEnvironment | `status.outputs.environment_id` |
| `spec.volumes[].storageName` | AzureContainerAppEnvironmentStorage | `status.outputs.storage_name` |
| `spec.eventTrigger.scale.rules[].identityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.registries[].server` | AzureContainerRegistry | `status.outputs.login_server` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)

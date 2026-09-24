# Azure Container App Job

Provisions an Azure Container App Job -- a run-to-completion containerized workload inside a Container App Environment. Where a Container App runs continuously and serves traffic, a Job starts, does its work, and exits: database migrations, batch processing, report generation, queue draining. Each execution runs the job's template to completion; there is no ingress, no revision model, and no scale-to-zero, because executions ARE the scaling unit. Every job carries exactly one trigger -- manual, cron schedule, or KEDA event scaler -- and that choice defines its operating model.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Container App Job** -- in the referenced resource group and environment, with exactly one trigger (manual on-demand, cron schedule in UTC, or a KEDA event scaler), the container template (main + init containers, env vars, probes, volume mounts), and the execution bounds (replica timeout, retry limit)
- **Managed identity wiring** (when configured) -- system-assigned and/or user-assigned identities for keyless Key Vault reads, registry pulls, and event-source polling
- **Governance tags** -- your tags merged over the Planton-derived resource tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureResourceGroup** the job will live in. Reference its `resource_group_name` output via ValueFromRef.
- **An AzureContainerAppEnvironment** providing networking, logging, and compute. The job's `region` must MATCH the environment's region -- Azure rejects a mismatch at deploy.
- **For private images** -- registry credentials, or (recommended) an AzureUserAssignedIdentity holding AcrPull, pre-granted so the very first execution can pull.
- **For persistent results** -- an AzureContainerAppEnvironmentStorage registration the job's volumes can reference; EmptyDir scratch space dies with each replica.

## Deploy

### Console

Open the deployment store, find **Azure Container App Job**, and click **Deploy**. The creation wizard leads with the trigger decision (manual / schedule / event -- the job's defining choice), then walks the execution bounds (with Azure's 1800-second portal default seeded as the timeout recommendation), identity, the declare-before-reference secret namespace, registries, the KEDA scale rules (on the event trigger only), the container template, probes, volumes, and tags. Start from the **Scheduled Batch Job** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppJob
metadata:
  name: nightly-report
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  jobName: nightly-report
  containerAppEnvironmentId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.App/managedEnvironments/prod-apps-env"
  replicaTimeoutInSeconds: 1800
  scheduleTrigger:
    cronExpression: "0 2 * * *"
  containers:
    - name: report
      image: myregistry.azurecr.io/report:v3.1.0
      cpu: 0.5
      memory: 1Gi
```

```shell
planton apply -f job.yaml
```

This creates a job that runs the report container to completion at 2 AM UTC every night, with a 30-minute deadline per execution. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the job onto resources deployed in the same InfraPipeline -- the resource group, the environment, and a pre-granted identity for keyless scaling:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: batch-rg
      fieldPath: status.outputs.resource_group_name
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: apps-env
      fieldPath: status.outputs.environment_id
  eventTrigger:
    scale:
      rules:
        - name: queue-depth
          customRuleType: azure-servicebus
          metadata:
            queueName: orders
            messageCount: "5"
          identityId:
            valueFrom:
              kind: AzureUserAssignedIdentity
              name: batch-identity
              fieldPath: status.outputs.identity_id
```

The InfraPipeline resolves the dependency graph, deploys the environment and identity first, then provisions the job.

## Key Configuration

These are the most important decisions when configuring a job. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Trigger** -- exactly one per job, and it defines the operating model. `manualTrigger` is the on-demand batch (start via CLI, SDK, or the start API); `scheduleTrigger` runs a five-field cron in UTC (overlapping windows start NEW executions alongside running ones); `eventTrigger` is the queue-worker model -- a KEDA scaler polls the source and fans executions out up to `maxExecutions`. Every trigger carries the parallelism / completion-count pair (Azure defaults 1/1) for indexed parallel work. The CHOICE of trigger block is fixed at creation -- switching trigger types replaces the job and discards its execution history, as does changing `region`, `resourceGroup`, `jobName`, or `containerAppEnvironmentId`; the dials inside the trigger, the execution bounds, and the whole container template edit in place and apply to the NEXT execution.

**Replica timeout** -- required with NO Azure default: the hard deadline after which a replica is terminated and counted as FAILED. Size it to the slowest LEGITIMATE execution -- an undersized timeout converts healthy slow runs into failures and retries them into the same wall. Pair long timeouts with a liveness probe so hung processes restart in seconds instead of burning the window.

**Keyless posture** -- a user-assigned identity attached to the job can pull the image (AcrPull), read Key Vault secrets, and poll the event source (the scale rule's `identityId`) -- zero stored credentials, nothing to rotate, and pre-granted identities work on the very first execution.

**Secrets** -- declared once, referenced by name from env vars, registry passwords, and scale-rule auth. Each value is either a managed org secret or a Key Vault reference the job reads with its identity.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds[]`, scale-rule `identityId` | `status.outputs.identity_id` |
| **AzureContainerAppEnvironmentStorage** | `volumes[].storageName` | `status.outputs.storage_name` |
| **AzureContainerRegistry** | `registries[].server` | `status.outputs.login_server` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `job_id` | Azure Resource Manager ID of the job | What start calls and role-assignment scopes address |
| `job_name` | The job's name within the resource group | CLI/SDK start commands and runbooks |
| `event_stream_endpoint` | The job's event stream endpoint | Live execution monitoring |
| `outbound_ip_addresses` | Outbound IPs of the hosting environment | Firewall allowlists on databases and APIs the job calls |
| `identity_principal_id` | Principal ID of the system-assigned identity | The RBAC grant target for keyless access (empty unless system-assigned) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scheduled batch** -- a nightly report or periodic cleanup on a UTC cron with one container and fail-fast retries. Start from the **Scheduled Batch Job** preset.

**Queue worker** -- an event trigger scaling on Service Bus queue depth with a keyless scaler identity: each message batch gets its own execution, and `minExecutions: 0` scales to nothing between bursts. Start from the **Event-Triggered Queue Worker** preset.

**On-demand migration** -- a manual trigger run by the deploy pipeline before releasing the app: one replica, zero retries (a half-applied migration must never blindly re-run). Start from the **On-Demand Job (Manual Trigger)** preset.

## Works With

- [**Azure Resource Group**](/cloud-catalog/azure-resource-group) -- where the job lives
- [**Azure Container App Environment**](/cloud-catalog/azure-container-app-environment) -- provides the compute, network, and logging the job runs on
- [**Azure Container App**](/cloud-catalog/azure-container-app) -- the continuously-serving sibling; jobs drain the queues apps fill
- [**Azure Container App Environment Storage**](/cloud-catalog/azure-container-app-environment-storage) -- persistent Azure Files volumes for execution results
- [**Azure User Assigned Identity**](/cloud-catalog/azure-user-assigned-identity) -- keyless registry pulls, vault reads, and event-source polling
- [**Azure Service Bus Queue**](/cloud-catalog/azure-service-bus-queue) -- the classic event-trigger source for the queue-worker pattern

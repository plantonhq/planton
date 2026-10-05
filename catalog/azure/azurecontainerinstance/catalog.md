# Azure Container Instance

Deploys an Azure Container Instance container group -- serverless containers billed per second, with no cluster or VM to manage. One group holds one or more containers sharing a lifecycle, network namespace, and volumes (the sidecar pattern), plus optional one-shot init containers that run to completion before the main containers start. Almost the entire group is create-only -- after create, Azure applies only identity and tag changes in place; any other change replaces the group -- so design the group's shape before shipping workloads.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Container group** -- the containers with their images, CPU/memory, ports, environment, volumes, and probes; init containers; registry credentials; managed identity; optional Log Analytics diagnostics, custom DNS, customer-managed-key encryption, and the network posture (a public IP with optional DNS label, a private IP in a delegated subnet, or no IP at all)
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource ID); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A registry credential per private registry** (only for private images) -- an `imageRegistryCredentials` entry per registry server; prefer the user-assigned-identity form (grant it AcrPull on the registry) over username/password.
- **A delegated subnet** (only for the Private posture) -- the subnet referenced by `subnetId` must carry the `Microsoft.ContainerInstance/containerGroups` delegation (the AzureSubnet kind's `delegations` field).
- **A storage account and file share** (only for `azureFile` volumes) -- the mount authenticates with the account's access key; managed-identity mounts are not in the provider's surface.
- **A Log Analytics workspace** (only for diagnostics) -- referenced by its customer ID (the GUID, not the ARM ID) and shared key.

## Deploy

### Console

Open the deployment store, find **Azure Container Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public Web Container**, **Private VNet Worker**, or **Run-Once Job** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerInstance
metadata:
  name: hello-web
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  name: hello-web
  region: eastus
  osType: Linux
  dnsNameLabel: hello-web-acme
  dnsNameLabelReusePolicy: TenantReuse
  containers:
    - name: web
      image: mcr.microsoft.com/azuredocs/aci-helloworld:latest
      cpu: 0.5
      memory: 1
      ports:
        - port: 80
```

```shell
planton apply -f container-instance.yaml
```

This runs one always-restarting Linux container on a public IP at `hello-web-acme.eastus.azurecontainer.io`, billed per second while it runs. An Infra Job tracks the provisioning in real time.

### InfraChart

The private-worker shape composes with a delegated subnet, an identity, and a registry deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  name: queue-worker
  region: eastus
  osType: Linux
  ipAddressType: Private
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: aci-subnet
      fieldPath: status.outputs.subnet_id
  imageRegistryCredentials:
    - server:
        valueFrom:
          kind: AzureContainerRegistry
          name: platform-acr
          fieldPath: status.outputs.login_server
      userAssignedIdentityId:
        valueFrom:
          kind: AzureUserAssignedIdentity
          name: worker-identity
          fieldPath: status.outputs.identity_id
  identity:
    type: USER_ASSIGNED
    identityIds:
      - valueFrom:
          kind: AzureUserAssignedIdentity
          name: worker-identity
          fieldPath: status.outputs.identity_id
  containers:
    - name: worker
      image: platformacr.azurecr.io/queue-worker:1.4.2
      cpu: 1
      memory: 2
```

The InfraPipeline resolves the dependency graph, deploys the subnet, identity, and registry first, then provisions the group with the resolved references.

## Key Configuration

These are the most important decisions when configuring a container group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Design the shape first -- almost nothing updates in place** -- Azure applies exactly two changes to a live group: identity and tags. Everything else -- images, resources, ports, volumes, probes, networking -- replaces the group. Three fields are worse than that: `cpuLimit`, `memoryLimit`, and `keyVaultUserAssignedIdentityId` are accepted on update and silently never applied, so a manifest change there looks successful and does nothing. Treat all three as create-only.

**This is a group, not a pod with a scheduler** -- There is no rescheduling, rolling update, or horizontal scaling. A liveness probe restarts a container in place on the same host; if the host dies, an "Always" group restarts elsewhere, but a "Never" job does not rerun. Reach for this kind when the unit of work IS the group -- jobs, burst workers, simple services, sidecar pairs; reach for Container Apps or AKS when you need orchestration.

**Pick one of three network postures and mean it** -- Public (the default) gives an internet-facing IP; add `dnsNameLabel` for a stable name, and set `dnsNameLabelReusePolicy` stricter than the default "Unsecure" if anything long-lived points at the label -- a released label under "Unsecure" is claimable by anyone (dangling-DNS takeover). Private joins a delegated subnet; Azure serializes group operations per subnet, so parallel deploys into one subnet queue up. None means no group IP -- and it is the only posture Spot priority accepts. Manifest validation rejects the contradictory combinations (a DNS label on a private or IP-less group) before Azure silently discards them.

**Match the restart policy to the workload** -- "Always" (the default) for services, "Never" for run-once jobs (the group shows Terminated when done -- read the exit state before deleting), "OnFailure" for batch work that should retry non-zero exits. The policy is create-only, like the rest of the shape.

**The secrets never come back** -- Secure environment variables, volume storage keys, inline secret files, registry passwords, and the Log Analytics workspace key are all write-only: Azure never returns them on reads. Reference managed secrets as `$secret/<slug>` in the manifest rather than pasting values. Two consequences: an import of an existing group cannot recover them (re-supply them in the manifest), and a manifest carrying secure environment variables plans a one-time destroy-and-recreate on the first apply after import -- the platform cannot verify the running group's secrets match yours. Rotating a secure value also replaces the group, by design.

**Volumes: four forms, one choice each** -- `azureFile` is the only persistent form, and it needs the storage account key, so treat that manifest as secret-bearing. `emptyDir` is group-lifetime scratch; the SAME name across containers is one shared volume -- that is the sharing mechanism (init seeds, worker reads). `gitRepo` clones at start -- pin a `revision` or every replacement group gets whatever the branch head is that day. `secret` mounts inline files whose values must be base64 of the file content; plain text fails at mount time, not validation.

**Spot is for rerunnable work only** -- `priority: Spot` trades evictability for a steep discount and requires `ipAddressType: None`. It pairs naturally with `restartPolicy: Never` jobs that can simply run again; never put a service on it.

**Log Analytics: leave logType unset or use ContainerInsights** -- Wiring `diagnosticsLogAnalytics` with just the workspace ID and key ships logs under Azure's server-side default schema and always works. `logType: ContainerInstanceLogs` cannot currently be deployed -- the provider attaches a metadata object whenever a log type is set, and ARM rejects metadata for that type (`LogAnalyticsMetadataNotAllowed`, live-proven); validation rejects it up front. `ContainerInsights` accepts metadata, but the keys are a closed vocabulary (`pod-uuid`, `cluster-resource-id`, `node-name`) -- for your own labels, use tags.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (Private posture) | `subnetId` | `status.outputs.subnet_id` |
| **AzureContainerRegistry** (private images) | `imageRegistryCredentials[].server` | `status.outputs.login_server` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds[]`, `imageRegistryCredentials[].userAssignedIdentityId`, `keyVaultUserAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureStorageShare** (azureFile volumes) | `containers[].volumes[].azureFile.shareName` | `status.outputs.share_name` |
| **AzureStorageAccount** (azureFile volumes) | `containers[].volumes[].azureFile.storageAccountName` / `storageAccountKey` | `status.outputs.storage_account_name` / `status.outputs.primary_access_key` |
| **AzureLogAnalyticsWorkspace** (diagnostics) | `diagnosticsLogAnalytics.workspaceId` / `workspaceKey` | `status.outputs.workspace_customer_id` / `status.outputs.primary_shared_key` |
| **AzureKeyVaultKey** (CMK encryption) | `keyVaultKeyId` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ip_address` | The group's IP -- public or private per the posture; empty for "None" | DNS records, upstream proxies, health monitors |
| `fqdn` | `{dnsNameLabel}.{region}.azurecontainer.io`; empty without a label | CNAME targets for stable public addressing |
| `identity_principal_id` | The system-assigned identity's principal ID (empty unless SYSTEM_ASSIGNED) | AzureRoleAssignment grants so containers reach Key Vault, storage, or ACR without embedded secrets |

`container_group_id`, `container_group_name`, and `identity_tenant_id` identify the group and its identity's tenant; nothing downstream consumes them by reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public web container** -- one always-on Linux container with a public IP, a stable DNS label under a strict reuse policy, and a cheap liveness probe. The simplest service shape; also the honest way to try an image in a real Azure network before committing to Container Apps or AKS. Start from the **Public Web Container** preset.

**Private VNet worker** -- a queue processor in a delegated subnet, pulling from ACR as a user-assigned identity, no public surface and no secrets in the manifest. The same identity serves the workload's own Azure access, and identity is the one setting you can rotate without replacing the group. Start from the **Private VNet Worker** preset.

**Run-once job** -- `restartPolicy: Never`, an init container seeding a shared `emptyDir`, optionally Spot priority with `ipAddressType: None` for the discount. The group terminates when the work ends and bills only for the seconds it ran; read the exit state before deleting. Start from the **Run-Once Job** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the container group lives in
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- the delegated subnet private groups join
- [**Azure Container Registry**](/infra-catalog/azure-container-registry) -- the private registry images pull from
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- keyless registry pulls and the containers' own Azure access
- [**Azure Storage Share**](/infra-catalog/azure-storage-share) / [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the persistent `azureFile` volume form
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- destination for container logs and events
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- customer-managed-key encryption of the group's ephemeral state

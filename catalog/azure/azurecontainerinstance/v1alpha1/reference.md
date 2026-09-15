# AzureContainerInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureContainerInstanceSpec** defines an Azure Container Instance
container group -- serverless containers: hand Azure an image plus
CPU and memory, and it runs, billed per second while it runs. One
group holds one or more containers that share a lifecycle, network,
and volumes (the sidecar pattern), plus optional one-shot init
containers that run to completion before the main containers start.

Almost the entire group is CREATE-ONLY: after create, Azure applies
only identity and tag changes in place -- any other change replaces
the group. Design the group's shape before shipping workloads.

Networking is one of three postures: "Public" (an internet-facing IP,
optionally with a DNS name label), "Private" (an IP inside a subnet
delegated to Microsoft.ContainerInstance/containerGroups), or "None"
(no group IP at all -- and the only posture Spot priority accepts).

## Example

```yaml
# Deep-shape example for docs and offline validation: a maximal PUBLIC
# container group exercising every arm -- two containers (one with all
# four volume forms across the pair, probes in both forms, limits,
# secure environment variables), an init container seeding a shared
# scratch volume, both registry-credential forms, combined identity,
# Log Analytics diagnostics with typed metadata, custom DNS, zones,
# narrowed exposed ports, and customer-managed-key encryption.
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerInstance
metadata:
  name: test-container-instance
  id: test-container-instance
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: app-aci
  region: eastus
  osType: Linux
  restartPolicy: Always
  sku: Standard
  ipAddressType: Public
  dnsNameLabel: app-aci-acme
  dnsNameLabelReusePolicy: TenantReuse
  zones:
    - "1"
  # Narrow the group's surface to the web port -- the sidecar's metrics
  # port stays group-internal.
  exposedPorts:
    - port: 80
  initContainers:
    - name: seed
      image: mcr.microsoft.com/cbl-mariner/busybox:2.0
      commands: ["sh", "-c", "echo ready > /work/ready"]
      environmentVariables:
        SEED_MODE: static
      volumes:
        - name: work
          mountPath: /work
          emptyDir: true
  containers:
    - name: web
      image: mcr.microsoft.com/azuredocs/aci-helloworld:latest
      cpu: 0.5
      memory: 1
      cpuLimit: 1
      memoryLimit: 1.5
      ports:
        - port: 80
      environmentVariables:
        APP_MODE: production
      secureEnvironmentVariables:
        API_TOKEN: test-token-value
      volumes:
        - name: work
          mountPath: /work
          emptyDir: true
        - name: content
          mountPath: /content
          readOnly: true
          azureFile:
            shareName:
              value: app-content
            storageAccountName:
              value: appcontentsa
            storageAccountKey:
              value: dGVzdC1zdG9yYWdlLWtleQ==
      livenessProbe:
        httpGet:
          path: /
          port: 80
          scheme: http
        periodSeconds: 30
        failureThreshold: 3
      readinessProbe:
        exec: ["ls", "/work/ready"]
        initialDelaySeconds: 5
        successThreshold: 1
        timeoutSeconds: 2
    - name: sidecar
      image: registry.example.com/acme/metrics-sidecar:2.1.0
      cpu: 0.25
      memory: 0.5
      ports:
        - port: 9090
      commands: ["/bin/sidecar", "--listen", ":9090"]
      security:
        privilegeEnabled: false
      volumes:
        - name: rules
          mountPath: /etc/rules
          gitRepo:
            url: https://github.com/acme/metric-rules
            directory: rules
            revision: v1.4.0
        - name: creds
          mountPath: /etc/creds
          secret:
            client.pem: dGVzdC1jbGllbnQtcGVt
  imageRegistryCredentials:
    - server:
        value: acme.azurecr.io
      userAssignedIdentityId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-mi
    - server:
        value: registry.example.com
      username: puller
      password: test-registry-password
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-mi
  diagnosticsLogAnalytics:
    workspaceId:
      value: 00000000-0000-0000-0000-000000000000
    workspaceKey:
      value: dGVzdC13b3Jrc3BhY2Uta2V5
    logType: ContainerInsights
    # Metadata keys are a CLOSED ARM vocabulary (pod-uuid /
    # cluster-resource-id / node-name -- the Kubernetes-shaped tags
    # Container Insights understands); values are free strings.
    metadata:
      cluster-resource-id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster
      node-name: aci-connector
  dnsConfig:
    nameservers:
      - 10.0.0.10
    searchDomains:
      - internal.acme.com
    options:
      - ndots:2
  keyVaultKeyId:
    value: https://app-vault.vault.azure.net/keys/aci-cmk/0123456789abcdef0123456789abcdef
  keyVaultUserAssignedIdentityId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-mi
  tags:
    workload: web
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.osType` | `string` | yes |  |  |
| `spec.restartPolicy` | `string` |  |  |  |
| `spec.sku` | `string` |  |  |  |
| `spec.priority` | `string` |  |  |  |
| `spec.ipAddressType` | `string` |  |  |  |
| `spec.dnsNameLabel` | `string` |  |  |  |
| `spec.dnsNameLabelReusePolicy` | `string` |  |  |  |
| `spec.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.zones` | `[]string` |  |  |  |
| `spec.exposedPorts` | `[]AzureContainerInstancePort` |  |  |  |
| `spec.exposedPorts[].port` | `int32` |  |  |  |
| `spec.exposedPorts[].protocol` | `string` |  |  |  |
| `spec.containers` | `[]AzureContainerInstanceContainer` | yes |  |  |
| `spec.containers[].name` | `string` | yes |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].cpu` | `double` | yes |  |  |
| `spec.containers[].memory` | `double` | yes |  |  |
| `spec.containers[].cpuLimit` | `double` |  |  |  |
| `spec.containers[].memoryLimit` | `double` |  |  |  |
| `spec.containers[].ports` | `[]AzureContainerInstancePort` |  |  |  |
| `spec.containers[].ports[].port` | `int32` |  |  |  |
| `spec.containers[].ports[].protocol` | `string` |  |  |  |
| `spec.containers[].environmentVariables` | `map<string, string>` |  |  |  |
| `spec.containers[].secureEnvironmentVariables` | `map<string, string>` (sensitive) |  |  |  |
| `spec.containers[].commands` | `[]string` |  |  |  |
| `spec.containers[].volumes` | `[]AzureContainerInstanceVolume` |  |  |  |
| `spec.containers[].volumes[].name` | `string` | yes |  |  |
| `spec.containers[].volumes[].mountPath` | `string` | yes |  |  |
| `spec.containers[].volumes[].readOnly` | `bool` |  |  |  |
| `spec.containers[].volumes[].azureFile` | `AzureContainerInstanceVolumeAzureFile` |  |  |  |
| `spec.containers[].volumes[].azureFile.shareName` | `string \| valueFrom` | yes |  | AzureStorageShare (`status.outputs.share_name`) |
| `spec.containers[].volumes[].azureFile.storageAccountName` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.containers[].volumes[].azureFile.storageAccountKey` | `string \| valueFrom` (sensitive) | yes |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.containers[].volumes[].emptyDir` | `bool` |  |  |  |
| `spec.containers[].volumes[].gitRepo` | `AzureContainerInstanceVolumeGitRepo` |  |  |  |
| `spec.containers[].volumes[].gitRepo.url` | `string` | yes |  |  |
| `spec.containers[].volumes[].gitRepo.directory` | `string` |  |  |  |
| `spec.containers[].volumes[].gitRepo.revision` | `string` |  |  |  |
| `spec.containers[].volumes[].secret` | `map<string, string>` (sensitive) |  |  |  |
| `spec.containers[].security` | `AzureContainerInstanceSecurity` |  |  |  |
| `spec.containers[].security.privilegeEnabled` | `bool` |  |  |  |
| `spec.containers[].livenessProbe` | `AzureContainerInstanceProbe` |  |  |  |
| `spec.containers[].livenessProbe.exec` | `[]string` |  |  |  |
| `spec.containers[].livenessProbe.httpGet` | `AzureContainerInstanceProbeHttpGet` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders` | `map<string, string>` |  |  |  |
| `spec.containers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe` | `AzureContainerInstanceProbe` |  |  |  |
| `spec.containers[].readinessProbe.exec` | `[]string` |  |  |  |
| `spec.containers[].readinessProbe.httpGet` | `AzureContainerInstanceProbeHttpGet` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.httpHeaders` | `map<string, string>` |  |  |  |
| `spec.containers[].readinessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.successThreshold` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.initContainers` | `[]AzureContainerInstanceInitContainer` |  |  |  |
| `spec.initContainers[].name` | `string` | yes |  |  |
| `spec.initContainers[].image` | `string` | yes |  |  |
| `spec.initContainers[].environmentVariables` | `map<string, string>` |  |  |  |
| `spec.initContainers[].secureEnvironmentVariables` | `map<string, string>` (sensitive) |  |  |  |
| `spec.initContainers[].commands` | `[]string` |  |  |  |
| `spec.initContainers[].volumes` | `[]AzureContainerInstanceVolume` |  |  |  |
| `spec.initContainers[].volumes[].name` | `string` | yes |  |  |
| `spec.initContainers[].volumes[].mountPath` | `string` | yes |  |  |
| `spec.initContainers[].volumes[].readOnly` | `bool` |  |  |  |
| `spec.initContainers[].volumes[].azureFile` | `AzureContainerInstanceVolumeAzureFile` |  |  |  |
| `spec.initContainers[].volumes[].azureFile.shareName` | `string \| valueFrom` | yes |  | AzureStorageShare (`status.outputs.share_name`) |
| `spec.initContainers[].volumes[].azureFile.storageAccountName` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.initContainers[].volumes[].azureFile.storageAccountKey` | `string \| valueFrom` (sensitive) | yes |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.initContainers[].volumes[].emptyDir` | `bool` |  |  |  |
| `spec.initContainers[].volumes[].gitRepo` | `AzureContainerInstanceVolumeGitRepo` |  |  |  |
| `spec.initContainers[].volumes[].gitRepo.url` | `string` | yes |  |  |
| `spec.initContainers[].volumes[].gitRepo.directory` | `string` |  |  |  |
| `spec.initContainers[].volumes[].gitRepo.revision` | `string` |  |  |  |
| `spec.initContainers[].volumes[].secret` | `map<string, string>` (sensitive) |  |  |  |
| `spec.initContainers[].security` | `AzureContainerInstanceSecurity` |  |  |  |
| `spec.initContainers[].security.privilegeEnabled` | `bool` |  |  |  |
| `spec.imageRegistryCredentials` | `[]AzureContainerInstanceImageRegistryCredential` |  |  |  |
| `spec.imageRegistryCredentials[].server` | `string \| valueFrom` | yes |  | AzureContainerRegistry (`status.outputs.login_server`) |
| `spec.imageRegistryCredentials[].username` | `string` |  |  |  |
| `spec.imageRegistryCredentials[].password` | `string` (sensitive) |  |  |  |
| `spec.imageRegistryCredentials[].userAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.identity` | `AzureContainerInstanceIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.diagnosticsLogAnalytics` | `AzureContainerInstanceLogAnalytics` |  |  |  |
| `spec.diagnosticsLogAnalytics.workspaceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_customer_id`) |
| `spec.diagnosticsLogAnalytics.workspaceKey` | `string \| valueFrom` (sensitive) | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.primary_shared_key`) |
| `spec.diagnosticsLogAnalytics.logType` | `string` |  |  |  |
| `spec.diagnosticsLogAnalytics.metadata` | `map<string, string>` |  |  |  |
| `spec.dnsConfig` | `AzureContainerInstanceDnsConfig` |  |  |  |
| `spec.dnsConfig.nameservers` | `[]string` | yes |  |  |
| `spec.dnsConfig.searchDomains` | `[]string` |  |  |  |
| `spec.dnsConfig.options` | `[]string` |  |  |  |
| `spec.keyVaultKeyId` | `string \| valueFrom` |  |  | AzureKeyVaultKey (`status.outputs.key_id`) |
| `spec.keyVaultUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the container group lives in. Can be a
literal string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The container group's name. The provider validates only that it is
non-empty; Azure itself enforces the final form at create
(lowercase letters, numbers, and dashes travel safely).

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.region

`string` · required

The Azure region the group is created in, e.g. "eastus".

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.osType

`string` · required

The operating system of every container in the group: "Linux" or
"Windows". Windows groups pull much larger images and support
fewer features (no init containers reach general availability
there).

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"in":["Linux","Windows"]}}

### spec.restartPolicy

`string`

What happens when a container exits: "Always" restart (the
provider default when unset -- long-running services), "Never"
(run-once jobs; the group shows Terminated when done), or
"OnFailure" (retry only non-zero exits).

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","Always","Never","OnFailure"]}}

### spec.sku

`string`

The group's SKU: unset means the provider default, "Standard".
"Confidential" runs the group on AMD SEV-SNP hardware (confidential
computing); "Dedicated" places it on dedicated hosts;
"NotSpecified" is Azure's legacy null token.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","Standard","Dedicated","Confidential","NotSpecified"]}}

### spec.priority

`string`

Scheduling priority: unset or "Regular" is normal capacity; "Spot"
runs on spare capacity at steep discount and can be evicted at any
time. Spot groups cannot have a group IP -- Azure requires
ip_address_type "None" (the provider enforces this before apply,
mirrored here as a validation rule).

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","Regular","Spot"]}}

### spec.ipAddressType

`string`

The group's IP posture: unset means the provider default,
"Public" (an internet-facing IP). "Private" gives the group an IP
inside subnet_id's subnet; "None" gives the group no IP at all
(required for Spot priority).

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","Public","Private","None"]}}

### spec.dnsNameLabel

`string`

A DNS name label for the public IP: the group becomes
{label}.{region}.azurecontainer.io. Only meaningful with the
"Public" posture -- rejected here (as are
dns_name_label_reuse_policy and exposed_ports) when
ip_address_type is "None", because Azure would silently discard
them (the provider sends no IP object at all in that posture).
Conflicts with subnet_id.

**ForceNew**: changing this destroys and recreates the group.

### spec.dnsNameLabelReusePolicy

`string`

Who may reuse the DNS name label after this group releases it:
unset means the provider default, "Unsecure" (anyone). The
stricter scopes ("Noreuse", "ResourceGroupReuse",
"SubscriptionReuse", "TenantReuse") protect against dangling-DNS
takeover of a label your systems still point at. Only meaningful
alongside dns_name_label -- rejected when ip_address_type is
"None" for the same silent-discard reason as the label itself.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","Noreuse","ResourceGroupReuse","SubscriptionReuse","TenantReuse","Unsecure"]}}

### spec.subnetId

`string | valueFrom`

The subnet the group joins for the "Private" posture. The subnet
must be delegated to Microsoft.ContainerInstance/containerGroups.
Can be a literal ARM ID or a reference to an AzureSubnet output.
Conflicts with dns_name_label (a private group has no public DNS).
Azure serializes container-group operations per subnet, so groups
sharing a subnet deploy one at a time.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.zones

`[]string`

Availability zones to pin the group into, e.g. ["1"]. Unset lets
Azure place the group.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.exposedPorts

`[]AzureContainerInstancePort`

Ports exposed on the GROUP's IP. Omit to expose every port every
container declares (the provider derives the group ports). When
set, each entry must match a port+protocol some container
declares -- the provider rejects the apply otherwise, mirrored
here as a validation rule.

**ForceNew**: changing this destroys and recreates the group.

### spec.exposedPorts[].port

`int32`

The port number (1-65535).

**ForceNew**: changing this destroys and recreates the group.

- rule: port must be between 1 and 65535

### spec.exposedPorts[].protocol

`string`

The protocol: unset means the provider default, "TCP". "UDP" is
the other choice.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","TCP","UDP"]}}

### spec.containers

`[]AzureContainerInstanceContainer` · required

The containers that run in the group -- at least one. All
containers share the group's lifecycle, network namespace, and
volumes; they address each other on localhost.

**ForceNew**: changing this list destroys and recreates the group.

- rule: {"repeated":{"minItems":"1"}}

### spec.containers[].name

`string` · required

The container's name -- unique within the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].image

`string` · required

The image to run, e.g.
"mcr.microsoft.com/azuredocs/aci-helloworld:latest". Private
registries need a matching image_registry_credentials entry on the
spec.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].cpu

`double` · required

The vCPU cores REQUESTED for this container, e.g. 0.5 or 1. The
group's total is what Azure bills.

- rule: {"required":true}

### spec.containers[].memory

`double` · required

The memory REQUESTED for this container, in GB, e.g. 0.5 or 1.5.

- rule: {"required":true}

### spec.containers[].cpuLimit

`double` · optional (explicit presence)

An upper vCPU LIMIT the container may burst to. BEHAVIOR: the
provider applies this at CREATE only -- changing it alone later is
silently never applied (the provider's update path covers only
identity and tags), so treat it as create-only in practice.

- rule: {"double":{"gte":0}}

### spec.containers[].memoryLimit

`double` · optional (explicit presence)

An upper memory LIMIT in GB the container may burst to. The same
create-only-in-practice behavior as cpu_limit.

- rule: {"double":{"gte":0}}

### spec.containers[].ports

`[]AzureContainerInstancePort`

Ports this container listens on. The group exposes the union of
all containers' ports unless the spec's exposed_ports narrows it.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].ports[].port

`int32`

The port number (1-65535).

**ForceNew**: changing this destroys and recreates the group.

- rule: port must be between 1 and 65535

### spec.containers[].ports[].protocol

`string`

The protocol: unset means the provider default, "TCP". "UDP" is
the other choice.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","TCP","UDP"]}}

### spec.containers[].environmentVariables

`map<string, string>`

Plain environment variables, name to value.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].secureEnvironmentVariables

`map<string, string>` · sensitive

SECRET environment variables, name to value -- hidden from the
portal, the CLI, and API reads (Azure never returns the values;
both engines re-send them from configuration on updates).

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].commands

`[]string`

Override the image's entrypoint/command. Reads echo the image's
own entrypoint back when this is unset -- expected, not drift.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].volumes

`[]AzureContainerInstanceVolume`

Volumes mounted into this container. Each entry is exactly one of
the four volume forms (Azure File share, empty scratch dir, git
checkout, or inline secret files). An empty_dir volume with the
SAME name in several containers is one shared scratch volume --
that is the sharing mechanism; other forms must use unique names
across the group.

**ForceNew**: changing this destroys and recreates the group.

- rule: set exactly one volume form: azure_file, empty_dir, git_repo, or secret

### spec.containers[].volumes[].name

`string` · required

The volume's name. An empty_dir with the SAME name in several
containers is ONE shared scratch volume (how containers share
files); the other forms need names unique across the group.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumes[].mountPath

`string` · required

Where the volume mounts inside the container, e.g. "/data".

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumes[].readOnly

`bool`

Mount read-only. Unset means the provider default, writable.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].volumes[].azureFile

`AzureContainerInstanceVolumeAzureFile`

An Azure File share mounted into the container -- the persistent
form. Set exactly one volume form.

### spec.containers[].volumes[].azureFile.shareName

`string | valueFrom` · required

The file share's name. Can be a literal or a reference to an
AzureStorageShare output.

- references: AzureStorageShare (`status.outputs.share_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageShare, name: <that resource's name>, fieldPath: status.outputs.share_name}} -- a bare string does not parse

### spec.containers[].volumes[].azureFile.storageAccountName

`string | valueFrom` · required

The storage account holding the share. Can be a literal or a
reference to an AzureStorageAccount output. The container group
MOUNTS the share and lives in its own resource group (or its subnet),
so the reference is access, not placement, on a diagram -- the same
reasoning the Container App environment's storage registration
carries for the identical pair of fields.

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.containers[].volumes[].azureFile.storageAccountKey

`string | valueFrom` · required · sensitive

The storage account's access key. SECRET -- Azure never returns it
on reads; both engines re-send it from configuration on updates.
Reference an AzureStorageAccount's primary_access_key output or
pass a literal. Access, not placement, like storage_account_name.

- references: AzureStorageAccount (`status.outputs.primary_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_access_key}} -- a bare string does not parse

### spec.containers[].volumes[].emptyDir

`bool`

An empty scratch directory living as long as the group. Set
exactly one volume form.

### spec.containers[].volumes[].gitRepo

`AzureContainerInstanceVolumeGitRepo`

A git repository cloned into the volume at group start. Set
exactly one volume form.

### spec.containers[].volumes[].gitRepo.url

`string` · required

The repository URL, e.g. "https://github.com/acme/config".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumes[].gitRepo.directory

`string`

The directory inside the volume to clone into. Unset clones into
the volume root.

### spec.containers[].volumes[].gitRepo.revision

`string`

The commit hash or branch to check out. Unset checks out the
default branch's head.

### spec.containers[].volumes[].secret

`map<string, string>` · sensitive

Inline secret files: file name to content. Values must be
BASE64-ENCODED (Azure decodes them into the mounted files) and are
never returned on reads.

### spec.containers[].security

`AzureContainerInstanceSecurity`

The container's security context (Linux only).

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].security.privilegeEnabled

`bool`

Run the container PRIVILEGED (full host device access). Leave
false for anything internet-facing.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].livenessProbe

`AzureContainerInstanceProbe`

Restart the container when this probe fails.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].livenessProbe.exec

`[]string`

Run this command inside the container; exit 0 is healthy. Set
exec and/or http_get -- Azure runs whichever is present.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.containers[].livenessProbe.httpGet

`AzureContainerInstanceProbeHttpGet`

Probe an HTTP(S) endpoint the container serves. (ENGINE SHAPE: the
provider's schema accepts a LIST here but its own code keeps only
the last entry, so this models the single probe Azure actually
receives.)

### spec.containers[].livenessProbe.httpGet.path

`string`

The request path, e.g. "/healthz".

### spec.containers[].livenessProbe.httpGet.port

`int32`

The port to probe (1-65535).

- rule: port must be between 1 and 65535

### spec.containers[].livenessProbe.httpGet.scheme

`string`

The scheme: unset means "http" (Azure's default); "https" is the
other choice (the provider's own lowercase vocabulary). The
modules SEND "http" explicitly when unset -- ARM materializes the
scheme on reads and the provider treats it as replace-forcing, so
omitting it on the wire would make every re-apply propose a
destroy+create (live-proven).

- rule: {"string":{"in":["","http","https"]}}

### spec.containers[].livenessProbe.httpGet.httpHeaders

`map<string, string>`

Extra HTTP headers sent with each probe request, name to value.

### spec.containers[].livenessProbe.initialDelaySeconds

`int32`

Seconds to wait after container start before the first probe.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].livenessProbe.periodSeconds

`int32`

Seconds between probes. Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].livenessProbe.failureThreshold

`int32`

Consecutive failures before the probe is considered failed. Unset
rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].livenessProbe.successThreshold

`int32`

Consecutive successes before the probe is considered passing.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].livenessProbe.timeoutSeconds

`int32`

Seconds a single probe may run before counting as a failure.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].readinessProbe

`AzureContainerInstanceProbe`

Hold traffic until this probe succeeds.

**ForceNew**: changing this destroys and recreates the group.

### spec.containers[].readinessProbe.exec

`[]string`

Run this command inside the container; exit 0 is healthy. Set
exec and/or http_get -- Azure runs whichever is present.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.containers[].readinessProbe.httpGet

`AzureContainerInstanceProbeHttpGet`

Probe an HTTP(S) endpoint the container serves. (ENGINE SHAPE: the
provider's schema accepts a LIST here but its own code keeps only
the last entry, so this models the single probe Azure actually
receives.)

### spec.containers[].readinessProbe.httpGet.path

`string`

The request path, e.g. "/healthz".

### spec.containers[].readinessProbe.httpGet.port

`int32`

The port to probe (1-65535).

- rule: port must be between 1 and 65535

### spec.containers[].readinessProbe.httpGet.scheme

`string`

The scheme: unset means "http" (Azure's default); "https" is the
other choice (the provider's own lowercase vocabulary). The
modules SEND "http" explicitly when unset -- ARM materializes the
scheme on reads and the provider treats it as replace-forcing, so
omitting it on the wire would make every re-apply propose a
destroy+create (live-proven).

- rule: {"string":{"in":["","http","https"]}}

### spec.containers[].readinessProbe.httpGet.httpHeaders

`map<string, string>`

Extra HTTP headers sent with each probe request, name to value.

### spec.containers[].readinessProbe.initialDelaySeconds

`int32`

Seconds to wait after container start before the first probe.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].readinessProbe.periodSeconds

`int32`

Seconds between probes. Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].readinessProbe.failureThreshold

`int32`

Consecutive failures before the probe is considered failed. Unset
rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].readinessProbe.successThreshold

`int32`

Consecutive successes before the probe is considered passing.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.containers[].readinessProbe.timeoutSeconds

`int32`

Seconds a single probe may run before counting as a failure.
Unset rides Azure's default.

- rule: {"int32":{"gte":0}}

### spec.initContainers

`[]AzureContainerInstanceInitContainer`

One-shot containers that run TO COMPLETION, in order, before the
main containers start -- schema migrations, config fetchers,
volume seeders. Init containers carry no CPU/memory of their own,
no ports, and no probes.

**ForceNew**: changing this list destroys and recreates the group.

### spec.initContainers[].name

`string` · required

The init container's name -- unique within the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].image

`string` · required

The image to run.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].environmentVariables

`map<string, string>`

Plain environment variables, name to value.

### spec.initContainers[].secureEnvironmentVariables

`map<string, string>` · sensitive

SECRET environment variables -- the same never-read-back handling
as the main containers' secure variables.

### spec.initContainers[].commands

`[]string`

Override the image's entrypoint/command.

### spec.initContainers[].volumes

`[]AzureContainerInstanceVolume`

Volumes mounted into this init container -- typically the same
named empty_dir a main container mounts, so the init step can seed
files the workload reads.

- rule: set exactly one volume form: azure_file, empty_dir, git_repo, or secret

### spec.initContainers[].volumes[].name

`string` · required

The volume's name. An empty_dir with the SAME name in several
containers is ONE shared scratch volume (how containers share
files); the other forms need names unique across the group.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumes[].mountPath

`string` · required

Where the volume mounts inside the container, e.g. "/data".

**ForceNew**: changing this destroys and recreates the group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumes[].readOnly

`bool`

Mount read-only. Unset means the provider default, writable.

**ForceNew**: changing this destroys and recreates the group.

### spec.initContainers[].volumes[].azureFile

`AzureContainerInstanceVolumeAzureFile`

An Azure File share mounted into the container -- the persistent
form. Set exactly one volume form.

### spec.initContainers[].volumes[].azureFile.shareName

`string | valueFrom` · required

The file share's name. Can be a literal or a reference to an
AzureStorageShare output.

- references: AzureStorageShare (`status.outputs.share_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageShare, name: <that resource's name>, fieldPath: status.outputs.share_name}} -- a bare string does not parse

### spec.initContainers[].volumes[].azureFile.storageAccountName

`string | valueFrom` · required

The storage account holding the share. Can be a literal or a
reference to an AzureStorageAccount output. The container group
MOUNTS the share and lives in its own resource group (or its subnet),
so the reference is access, not placement, on a diagram -- the same
reasoning the Container App environment's storage registration
carries for the identical pair of fields.

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.initContainers[].volumes[].azureFile.storageAccountKey

`string | valueFrom` · required · sensitive

The storage account's access key. SECRET -- Azure never returns it
on reads; both engines re-send it from configuration on updates.
Reference an AzureStorageAccount's primary_access_key output or
pass a literal. Access, not placement, like storage_account_name.

- references: AzureStorageAccount (`status.outputs.primary_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_access_key}} -- a bare string does not parse

### spec.initContainers[].volumes[].emptyDir

`bool`

An empty scratch directory living as long as the group. Set
exactly one volume form.

### spec.initContainers[].volumes[].gitRepo

`AzureContainerInstanceVolumeGitRepo`

A git repository cloned into the volume at group start. Set
exactly one volume form.

### spec.initContainers[].volumes[].gitRepo.url

`string` · required

The repository URL, e.g. "https://github.com/acme/config".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.initContainers[].volumes[].gitRepo.directory

`string`

The directory inside the volume to clone into. Unset clones into
the volume root.

### spec.initContainers[].volumes[].gitRepo.revision

`string`

The commit hash or branch to check out. Unset checks out the
default branch's head.

### spec.initContainers[].volumes[].secret

`map<string, string>` · sensitive

Inline secret files: file name to content. Values must be
BASE64-ENCODED (Azure decodes them into the mounted files) and are
never returned on reads.

### spec.initContainers[].security

`AzureContainerInstanceSecurity`

The init container's security context (Linux only).

### spec.initContainers[].security.privilegeEnabled

`bool`

Run the container PRIVILEGED (full host device access). Leave
false for anything internet-facing.

**ForceNew**: changing this destroys and recreates the group.

### spec.imageRegistryCredentials

`[]AzureContainerInstanceImageRegistryCredential`

Credentials for pulling the containers' images from private
registries -- one entry per registry server. Prefer the
user-assigned-identity form over username/password: no secret to
rotate, and Azure grants the identity AcrPull on the registry.

**ForceNew**: changing this destroys and recreates the group.

### spec.imageRegistryCredentials[].server

`string | valueFrom` · required

The registry server, e.g. "myregistry.azurecr.io". Can be a
literal or a reference to an AzureContainerRegistry output.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureContainerRegistry (`status.outputs.login_server`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerRegistry, name: <that resource's name>, fieldPath: status.outputs.login_server}} -- a bare string does not parse

### spec.imageRegistryCredentials[].username

`string`

The registry username, for the username/password form.

**ForceNew**: changing this destroys and recreates the group.

### spec.imageRegistryCredentials[].password

`string` · sensitive

The registry password. SECRET -- Azure never returns it on reads.

**ForceNew**: changing this destroys and recreates the group.

### spec.imageRegistryCredentials[].userAssignedIdentityId

`string | valueFrom`

The user-assigned identity to pull as (grant it AcrPull on the
registry). Can be a literal ARM ID or a reference to an
AzureUserAssignedIdentity output.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.identity

`AzureContainerInstanceIdentity`

The group's managed identity -- how its containers authenticate to
other Azure services (Key Vault, storage, ACR) without embedded
secrets. The ONLY group setting besides tags that Azure updates in
place.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the group; USER_ASSIGNED brings identities you manage
(grantable on other resources BEFORE the group exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_container_instance_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the group. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity). Wire value: "UserAssigned".
- `SYSTEM_AND_USER_ASSIGNED` -- Both at once. Wire value: "SystemAssigned, UserAssigned".

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the group, by ARM ID. Reference
AzureUserAssignedIdentity resources so grants can be composed
before the group is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.diagnosticsLogAnalytics

`AzureContainerInstanceLogAnalytics`

Ship the containers' stdout/stderr and events to a Log Analytics
workspace. Azure never returns the workspace key on reads, so
both engines re-send it from configuration when the group is
updated. (This models the provider's diagnostics block, whose only
member is this Log Analytics form.)

**ForceNew**: changing this destroys and recreates the group.

- rule: metadata requires log_type -- the provider only sends metadata alongside a log type and silently discards it otherwise
- rule: log_type ContainerInstanceLogs cannot be deployed: the provider always sends a metadata object alongside a log type and ARM rejects metadata for this log type (LogAnalyticsMetadataNotAllowed) -- leave log_type unset for plain log shipping or use ContainerInsights
- rule: metadata keys are a closed vocabulary ARM validates (InvalidLogAnalyticsMetadataKeys): only pod-uuid, cluster-resource-id, and node-name are accepted

### spec.diagnosticsLogAnalytics.workspaceId

`string | valueFrom` · required

The workspace's CUSTOMER ID -- the GUID agents authenticate
against (the portal's "Workspace ID"), NOT the ARM resource ID.
Reference an AzureLogAnalyticsWorkspace's workspace_customer_id
output or pass the GUID literally.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_customer_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_customer_id}} -- a bare string does not parse

### spec.diagnosticsLogAnalytics.workspaceKey

`string | valueFrom` · required · sensitive

The workspace's shared key. SECRET -- Azure never returns it on
reads; both engines re-send it from configuration on updates.
Reference the workspace's primary_shared_key output or pass a
literal.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureLogAnalyticsWorkspace (`status.outputs.primary_shared_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.primary_shared_key}} -- a bare string does not parse

### spec.diagnosticsLogAnalytics.logType

`string`

The log schema: "ContainerInsights" (the richer, dashboard-ready
schema) or "ContainerInstanceLogs" (plain per-container logs).
Unset rides Azure's default.

"ContainerInstanceLogs" cannot currently be DEPLOYED: whenever a
log type is set the provider unconditionally attaches a metadata
object (even empty), and ARM rejects metadata for this log type
with LogAnalyticsMetadataNotAllowed (live-proven). A CEL rejects
the value with the same teaching so the failure surfaces at
validation, not mid-deploy; leave log_type unset for plain log
shipping (Azure's server-side default) or use "ContainerInsights".
Lift the CEL when the pinned provider stops sending empty
metadata.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"string":{"in":["","ContainerInsights","ContainerInstanceLogs"]}}

### spec.diagnosticsLogAnalytics.metadata

`map<string, string>`

Extra metadata attached to every log record. Requires log_type --
the provider only sends metadata alongside a log type and silently
discards it otherwise, so the pairing is enforced here.

The keys are a CLOSED vocabulary ARM validates server-side
(live-proven, InvalidLogAnalyticsMetadataKeys): only "pod-uuid",
"cluster-resource-id", and "node-name" are accepted -- the
Kubernetes-shaped tags Container Insights understands. Values are
free strings. A CEL enforces the key set so the failure surfaces
at validation, not mid-deploy.

**ForceNew**: changing this destroys and recreates the group.

### spec.dnsConfig

`AzureContainerInstanceDnsConfig`

Custom DNS for the group's containers -- the resolvers to use
instead of Azure's, plus optional search domains and resolver
options.

**ForceNew**: changing this destroys and recreates the group.

### spec.dnsConfig.nameservers

`[]string` · required

The DNS servers to use, in order -- at least one.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dnsConfig.searchDomains

`[]string`

Search domains appended to unqualified lookups.

**ForceNew**: changing this destroys and recreates the group.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.dnsConfig.options

`[]string`

Resolver options, e.g. "ndots:2".

**ForceNew**: changing this destroys and recreates the group.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.keyVaultKeyId

`string | valueFrom`

Customer-managed-key encryption for the group's ephemeral state: a
VERSIONED Key Vault key identifier
(https://{vault}.vault.azure.net/keys/{name}/{version}). Azure
Container Instances pins this exact key VERSION -- rotation does
not follow automatically (unlike consumers that accept versionless
identifiers). Reference an AzureKeyVaultKey's key_id output or
pass a literal versioned identifier.

**ForceNew**: changing this destroys and recreates the group.

- references: AzureKeyVaultKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.keyVaultUserAssignedIdentityId

`string | valueFrom`

The user-assigned identity Azure authenticates as when unwrapping
key_vault_key_id. It must be attached to the group (identity block)
and hold get/unwrap/wrap on the vault before create. BEHAVIOR: the
provider applies this at CREATE only -- changing it alone later is
silently never applied (the provider's update path skips it), so
treat it as create-only in practice.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the group, merged over the Planton-derived
metadata tags (user values win on key conflicts). Updatable in
place.

## Validation Rules

- `container_instance_spot_requires_ip_none`: priority "Spot" requires ip_address_type "None" -- Azure gives Spot groups no group IP (the provider enforces this before apply)
- `container_instance_subnet_conflicts_dns_label`: subnet_id and dns_name_label cannot be set together -- a subnet-joined group is private and has no public DNS name
- `container_instance_dns_label_requires_group_ip`: dns_name_label, dns_name_label_reuse_policy, and exposed_ports require an ip_address_type other than "None" -- without a group IP Azure silently discards them
- `container_instance_exposed_ports_declared_on_containers`: every exposed_ports entry must match a port and protocol declared by some container's ports -- Azure only exposes ports a container listens on (the provider enforces this before apply)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureContainerInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.container_group_id` | `string` | The container group's Azure Resource Manager ID. |
| `status.outputs.container_group_name` | `string` | The container group's name. |
| `status.outputs.ip_address` | `string` | The group's IP address -- public or private depending on ip_address_type. Empty for the "None" posture. |
| `status.outputs.fqdn` | `string` | The group's FQDN ({dns_name_label}.{region}.azurecontainer.io). Empty unless dns_name_label is set on a public group. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the group's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal access on other resources (Key Vault, storage) so the group's containers can reach them without embedded secrets. |
| `status.outputs.identity_tenant_id` | `string` | The tenant ID of the group's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.containers[].volumes[].azureFile.shareName` | AzureStorageShare | `status.outputs.share_name` |
| `spec.containers[].volumes[].azureFile.storageAccountName` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.containers[].volumes[].azureFile.storageAccountKey` | AzureStorageAccount | `status.outputs.primary_access_key` |
| `spec.initContainers[].volumes[].azureFile.shareName` | AzureStorageShare | `status.outputs.share_name` |
| `spec.initContainers[].volumes[].azureFile.storageAccountName` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.initContainers[].volumes[].azureFile.storageAccountKey` | AzureStorageAccount | `status.outputs.primary_access_key` |
| `spec.imageRegistryCredentials[].server` | AzureContainerRegistry | `status.outputs.login_server` |
| `spec.imageRegistryCredentials[].userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.diagnosticsLogAnalytics.workspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_customer_id` |
| `spec.diagnosticsLogAnalytics.workspaceKey` | AzureLogAnalyticsWorkspace | `status.outputs.primary_shared_key` |
| `spec.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.key_id` |
| `spec.keyVaultUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)

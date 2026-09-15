# AzureMachineLearningComputeCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningComputeClusterSpec** defines an auto-scaling
compute cluster on an Azure Machine Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/computes/{name})
-- the pool of VMs that training jobs and pipelines run on. The
cluster scales between its configured node bounds and, with
min_node_count 0, costs nothing while idle.

**Only identity, scale_settings and tags update in place** -- every
other field replaces the cluster (the provider's own contract).
Replacement is routine for clusters (they hold no data), but
running jobs on the cluster fail when it is replaced.

**The cluster is an ARM child of its workspace** -- it has no
resource group of its own. Uniquely in the ML family, the cluster
CAN live in a different region from its workspace (`region` here is
the compute location); note that ARM reports the cluster envelope
at the WORKSPACE's region, so cross-region clusters read back at
the workspace region while their nodes run where `region` says.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the low-priority
# wire value, the full scale-settings block, a system-assigned identity,
# the SSH admin account with both credential arms, VNet placement with
# both default-true booleans explicitly false, and the closed-by-default
# SSH port explicitly opened.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningComputeCluster
metadata:
  name: test-ml-compute-cluster
  org: test-org
  env: dev
spec:
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace
  name: test-spot-cluster
  region: eastus
  vmSize: STANDARD_DS3_V2
  vmPriority: LOW_PRIORITY
  scaleSettings:
    minNodeCount: 0
    maxNodeCount: 8
    scaleDownNodesAfterIdleDuration: PT15M
  identity:
    type: SYSTEM_ASSIGNED
  ssh:
    adminUsername: azureuser
    adminPassword:
      value: offline-plan-placeholder-password
    keyValue: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQoffline-plan-placeholder example
  sshPublicAccessEnabled: true
  localAuthEnabled: false
  nodePublicIpEnabled: false
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/ml-compute
  description: Offline-plan cluster exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.vmSize` | `string` | yes |  |  |
| `spec.vmPriority` | `enum` | yes |  |  |
| `spec.scaleSettings` | `AzureMachineLearningComputeClusterScaleSettings` | yes |  |  |
| `spec.scaleSettings.maxNodeCount` | `int32` |  |  |  |
| `spec.scaleSettings.minNodeCount` | `int32` |  |  |  |
| `spec.scaleSettings.scaleDownNodesAfterIdleDuration` | `string` | yes |  |  |
| `spec.identity` | `AzureMachineLearningComputeClusterIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.ssh` | `AzureMachineLearningComputeClusterSsh` |  |  |  |
| `spec.ssh.adminUsername` | `string` | yes |  |  |
| `spec.ssh.adminPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.ssh.keyValue` | `string` |  |  |  |
| `spec.sshPublicAccessEnabled` | `bool` |  |  |  |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.nodePublicIpEnabled` | `bool` |  | `true` |  |
| `spec.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.workspaceId

`string | valueFrom` · required

The Machine Learning workspace the cluster belongs to, by ARM ID.
Fixed at creation.

- references: AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningWorkspace, name: <that resource's name>, fieldPath: status.outputs.machine_learning_workspace_id}} -- a bare string does not parse

### spec.name

`string` · required

The cluster's name, unique on the workspace: 3-32 characters,
letters, digits and hyphens, starting with a letter and ending
with a letter or digit (the service's own rule, mirrored from the
provider) -- what jobs and pipelines reference as their compute
target. Changing the name replaces the cluster.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z][a-zA-Z0-9-]{1,30}[a-zA-Z0-9]$"}}

### spec.region

`string` · required

The Azure region the cluster's NODES run in, e.g. "eastus".
Clusters (unlike compute instances) may run in a different region
from their workspace -- useful when GPU quota lives elsewhere.
ARM still reports the cluster envelope at the workspace's region.
Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vmSize

`string` · required

The VM size of each node, e.g. "Standard_DS2_v2" or
"Standard_NC6s_v3" (GPU). Regional VM-family quota gates what
actually provisions. Fixed at creation.

- rule: {"required":true}

### spec.vmPriority

`enum` · required

Whether nodes are regular dedicated VMs or evictable low-priority
(spot-class) VMs at a steep discount. LOW_PRIORITY suits
fault-tolerant training (checkpointed jobs); evicted nodes take
running work with them. Fixed at creation.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_compute_cluster_vm_priority_unspecified` -- Not specified: rejected -- the provider requires an explicit priority.
- `DEDICATED` -- Regular dedicated VMs (wire value "Dedicated") -- nodes stay up until the cluster scales them down.
- `LOW_PRIORITY` -- Evictable low-priority VMs (wire value "LowPriority") -- deep discount, but Azure may evict nodes at any time; suit checkpointed, fault-tolerant training only.

### spec.scaleSettings

`AzureMachineLearningComputeClusterScaleSettings` · required

How the cluster scales. The one substantive setting that updates
in place -- tune bounds freely on a live cluster.

- rule: {"required":true}
- rule: max_node_count must be greater than or equal to min_node_count

### spec.scaleSettings.maxNodeCount

`int32`

The most nodes the cluster grows to. Regional vCPU quota for the
VM family gates what actually provisions.

- rule: {"int32":{"gte":0}}

### spec.scaleSettings.minNodeCount

`int32`

The fewest nodes the cluster keeps while idle. 0 is the
economical posture -- the cluster costs nothing between jobs, at
the price of node spin-up latency when work arrives.

- rule: {"int32":{"gte":0}}

### spec.scaleSettings.scaleDownNodesAfterIdleDuration

`string` · required

How long a node sits idle before scaling down, as an ISO-8601
duration (e.g. "PT30M", "PT2M"). Longer keeps warm nodes between
closely-spaced jobs; shorter saves cost.

- rule: scale_down_nodes_after_idle_duration must be an ISO-8601 duration, e.g. PT30M
- rule: {"required":true}

### spec.identity

`AzureMachineLearningComputeClusterIdentity`

The cluster's managed identity -- how jobs on the cluster reach
storage, Key Vault and the container registry without embedded
credentials. Updatable in place.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the cluster; USER_ASSIGNED brings identities you manage
(grantable data / registry access BEFORE the cluster exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_compute_cluster_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the cluster.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the cluster, by ARM ID. Reference
AzureUserAssignedIdentity resources so storage / registry grants
can be composed before the cluster is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.ssh

`AzureMachineLearningComputeClusterSsh`

The administrator account created on each node for SSH access.
Only meaningful with ssh_public_access_enabled true (or from
within the cluster's virtual network when false). The whole block
is fixed at creation.

- rule: the ssh block requires at least one of admin_password or key_value

### spec.ssh.adminUsername

`string` · required

The admin account's username on every node. Fixed at creation.

- rule: {"required":true}

### spec.ssh.adminPassword

`string | valueFrom` · sensitive

The admin account's password. Reference a secret rather than
embedding the literal. SSH keys (key_value) are the production
path; at least one credential is required (the provider's own
contract).

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ssh.keyValue

`string`

The admin account's SSH PUBLIC key (e.g. "ssh-rsa AAAA..."),
installed on every node. Public material -- not a secret. The
provider names this argument `key_value`.

### spec.sshPublicAccessEnabled

`bool`

Whether the nodes' SSH port answers the public internet. Azure's
default is false (closed) -- the right posture; enable it only
for debugging with the ssh block configured. Fixed at creation.

### spec.localAuthEnabled

`bool` · optional (explicit presence)

Whether Azure Machine Learning token authentication to the
cluster works. Unspecified applies true (the provider's default).
Set false to force Entra ID auth only -- pair with an identity.
Fixed at creation.

- default: `true`

### spec.nodePublicIpEnabled

`bool` · optional (explicit presence)

Whether each node gets a public IP. Unspecified applies true (the
provider's default). Set false for VNet-only clusters -- typically
together with subnet_id or a workspace managed network. Fixed at
creation.

- default: `true`

### spec.subnetId

`string | valueFrom`

The subnet the cluster's nodes are placed in, by ARM ID. Leave
unset to let Azure network the nodes (a workspace managed network
assigns one, read back after apply). Fixed at creation.
On a diagram the cluster is an ARM child of its workspace
(workspaces/{ws}/computes/{name}) and lives there; the subnet is
where its nodes attach, so the reference is access, not placement
-- otherwise a cluster on its own subnet would be drawn outside
the workspace it belongs to (the AKS node pool's rule).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.description

`string`

What the cluster is for. Fixed at creation (the provider's own
contract -- changing the description replaces the cluster).

### spec.tags

`map<string, string>`

Free-form tags applied to the cluster, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Updatable in
place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningComputeCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.machine_learning_compute_cluster_id` | `string` | The Azure Resource Manager ID of the compute cluster. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/computes/{name} |
| `status.outputs.machine_learning_compute_cluster_name` | `string` | The cluster's name -- what jobs and pipelines reference as their compute target within the workspace. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the cluster's system-assigned identity, when one is enabled -- what storage / Key Vault / ACR grants bind to so jobs on the cluster can reach data and images. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.workspaceId` | AzureMachineLearningWorkspace | `status.outputs.machine_learning_workspace_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMachineLearningBatchDeployment | `spec.computeId` | `status.outputs.machine_learning_compute_cluster_id` |

## See Also

- [Overview](../README.md)

# AzureMachineLearningComputeInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningComputeInstanceSpec** defines a compute
instance on an Azure Machine Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/computes/{name})
-- a single always-on VM serving as one data scientist's cloud
workstation for notebooks, interactive debugging, and small jobs.

**Nothing updates in place.** The provider has NO update path for
this resource -- every change, tags included, replaces the instance
(its OS disk and local files go with it; keep work in datastores
and git).

**The instance lives in its workspace's region** (the service's own
rule -- unlike clusters, instances cannot run elsewhere), and its
NAME is reserved region-wide per subscription: two instances in the
same region cannot share a name even across different workspaces,
and a deleted instance's name can stay reserved briefly.

**An instance is a running VM billing around the clock** unless
stopped -- it does not auto-scale to zero the way clusters do.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: personal
# authorization with a user assignment, a system-assigned identity,
# SSH enabled with a public key, VNet placement with both default-true
# booleans explicitly false, description, and tags.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningComputeInstance
metadata:
  name: test-ml-compute-instance
  org: test-org
  env: dev
spec:
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace
  name: test-alice-dev
  virtualMachineSize: STANDARD_DS3_V2
  authorizationType: personal
  assignToUser:
    tenantId: 00000000-0000-0000-0000-000000000000
    objectId: 11111111-1111-1111-1111-111111111111
  identity:
    type: SYSTEM_ASSIGNED
  localAuthEnabled: false
  ssh:
    publicKey: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQoffline-plan-placeholder example
  nodePublicIpEnabled: false
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/ml-workstations
  description: Offline-plan instance exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.virtualMachineSize` | `string` | yes |  |  |
| `spec.authorizationType` | `string` |  |  |  |
| `spec.assignToUser` | `AzureMachineLearningComputeInstanceAssignToUser` |  |  |  |
| `spec.assignToUser.tenantId` | `string` |  |  |  |
| `spec.assignToUser.objectId` | `string` |  |  |  |
| `spec.identity` | `AzureMachineLearningComputeInstanceIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.ssh` | `AzureMachineLearningComputeInstanceSsh` |  |  |  |
| `spec.ssh.publicKey` | `string` | yes |  |  |
| `spec.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.nodePublicIpEnabled` | `bool` |  | `true` |  |
| `spec.description` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.workspaceId

`string | valueFrom` · required

The Machine Learning workspace the instance belongs to, by ARM
ID. The instance always runs in this workspace's region. Fixed at
creation.

- references: AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningWorkspace, name: <that resource's name>, fieldPath: status.outputs.machine_learning_workspace_id}} -- a bare string does not parse

### spec.name

`string` · required

The instance's name: a letter followed by 3-24 letters, digits or
hyphens (the provider's own regex -- its error message understates
the length by one). Reserved REGION-WIDE per subscription, not
per workspace. Changing the name replaces the instance.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z][a-zA-Z0-9-]{3,24}$"}}

### spec.virtualMachineSize

`string` · required

The VM size, e.g. "STANDARD_DS3_V2" (the provider compares
case-insensitively). This VM runs around the clock until the
instance is stopped or deleted. Fixed at creation.

- rule: {"required":true}

### spec.authorizationType

`string`

The instance's authorization mode. "personal" (the only value the
provider accepts) locks the instance to one user -- its creator,
or the user in assign_to_user. Unspecified leaves the service
default (also personal-style single-user access). Fixed at
creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["personal"]}}

### spec.assignToUser

`AzureMachineLearningComputeInstanceAssignToUser`

Assign the instance to a user OTHER than the deploying principal
-- the admin-provisions-for-the-team pattern (pair with
authorization_type "personal"). Fixed at creation.

### spec.assignToUser.tenantId

`string`

The Entra ID tenant the user lives in.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.assignToUser.objectId

`string`

The user's Entra ID object ID.

### spec.identity

`AzureMachineLearningComputeInstanceIdentity`

The instance's managed identity -- how ITS workloads reach
storage and Key Vault without embedded credentials. Unlike the
cluster's, the whole block is fixed at creation.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the instance; USER_ASSIGNED brings identities you manage
(grantable data access BEFORE the instance exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_compute_instance_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the instance.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the instance, by ARM ID. Reference
AzureUserAssignedIdentity resources so storage / Key Vault grants
can be composed before the instance is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.localAuthEnabled

`bool` · optional (explicit presence)

Whether Azure Machine Learning token authentication to the
instance works. Unspecified applies true (the provider's
default). Set false to force Entra ID auth only. Fixed at
creation.

- default: `true`

### spec.ssh

`AzureMachineLearningComputeInstanceSsh`

SSH access to the instance. Absent means the SSH port is
DISABLED (the provider's own contract); present enables it with
the given public key -- the service assigns the username and
port, surfaced as outputs. Fixed at creation.

### spec.ssh.publicKey

`string` · required

The SSH PUBLIC key granted access (e.g. "ssh-rsa AAAA...").
Public material -- not a secret. The service assigns the admin
username and port, surfaced in the outputs.

- rule: {"required":true}

### spec.subnetId

`string | valueFrom`

The subnet the instance is placed in, by ARM ID. Only legal when
the workspace does NOT use a managed network (Azure then networks
the instance itself). Fixed at creation.
On a diagram the instance is an ARM child of its workspace
(workspaces/{ws}/computes/{name}) and lives there; the subnet is
where its network interface attaches, so the reference is access,
not placement (the compute cluster's and the AKS node pool's rule).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.nodePublicIpEnabled

`bool` · optional (explicit presence)

Whether the instance gets a public IP. Unspecified applies true
(the provider's default). When set false, the provider requires a
subnet_id UNLESS the workspace runs a managed network (its
isolation mode networks the instance) -- a live-workspace-state
contract enforced at apply time, not expressible at manifest
time. Fixed at creation.

- default: `true`

### spec.description

`string`

What the instance is for. Fixed at creation (the provider's own
contract -- changing the description replaces the instance).

### spec.tags

`map<string, string>`

Free-form tags applied to the instance, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Fixed at
creation (the provider's own contract -- tags included, unusually).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningComputeInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.machine_learning_compute_instance_id` | `string` | The Azure Resource Manager ID of the compute instance. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/computes/{name} |
| `status.outputs.machine_learning_compute_instance_name` | `string` | The instance's name -- what its owner selects as their compute in notebooks and the ML studio. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the instance's system-assigned identity, when one is enabled -- what storage / Key Vault grants bind to. |
| `status.outputs.ssh_username` | `string` | The admin username for SSH access, assigned by the service -- populated only when the ssh block is configured. |
| `status.outputs.ssh_port` | `int32` | The port the instance answers SSH on, assigned by the service -- populated only when the ssh block is configured. When the block is absent both engines must still emit a numeric zero (not a nil / empty string): a raw nil Pulumi export fails the harness int32 parse. Use Elem() on the optional SSH fields; Terraform already falls back with try(..., 0). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.workspaceId` | AzureMachineLearningWorkspace | `status.outputs.machine_learning_workspace_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## See Also

- [Overview](../README.md)

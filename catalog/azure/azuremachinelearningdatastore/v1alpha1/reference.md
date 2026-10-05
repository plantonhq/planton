# AzureMachineLearningDatastore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningDatastoreSpec** defines a datastore on an
Azure Machine Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/dataStores/{name})
-- the saved connection that tells the workspace where data lives
and how to reach it. ONE kind covers the three storage flavors as
variants: a blob container (blob_storage), a Data Lake Gen2
filesystem (data_lake_gen2), or an Azure Files share (file_share).
Exactly one variant block is set; the block IS the datastore type.

**Nearly everything is fixed at creation.** Only the credentials
inside the variant blocks and service_data_identity update in
place; `name`, the storage target, `description` and `tags` are all
ForceNew (the provider's own contract -- description and tags
included, unusually).

**Credentials are write-only.** ARM never returns account keys, SAS
tokens or client secrets -- the provider echoes them from
configuration. Reference secrets rather than embedding literals in
manifests.

**The datastore is an ARM child of its workspace** -- it has no
region or resource group of its own (ARM derives both through the
workspace).

## Example

```yaml
# Offline-plan test manifest. Exercises the blob variant's full
# surface: the container reference, the default-datastore claim, the
# service-data identity enum, and a SAS credential (write-only on ARM;
# the provider echoes it from configuration). The other two variants
# are mutually exclusive with this one by design (exactly one variant
# block) -- their plans are exercised by their own scenario shapes in
# the proof lane.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningDatastore
metadata:
  name: test-ml-datastore
  org: test-org
  env: dev
spec:
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace
  name: training_data
  description: Offline-plan datastore exercising the blob variant
  serviceDataIdentity: WORKSPACE_SYSTEM_ASSIGNED_IDENTITY
  blobStorage:
    storageContainerId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testmlstorage/blobServices/default/containers/training-data
    isDefault: false
    sharedAccessSignature:
      value: "sv=2024-01-01&ss=b&srt=co&sp=rl&sig=offline-plan-placeholder"
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.serviceDataIdentity` | `enum` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.blobStorage` | `AzureMachineLearningDatastoreBlobStorage` |  |  |  |
| `spec.blobStorage.storageContainerId` | `string \| valueFrom` | yes |  | AzureStorageContainer (`status.outputs.container_id`) |
| `spec.blobStorage.isDefault` | `bool` |  |  |  |
| `spec.blobStorage.accountKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.blobStorage.sharedAccessSignature` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.dataLakeGen2` | `AzureMachineLearningDatastoreDataLakeGen2` |  |  |  |
| `spec.dataLakeGen2.storageContainerId` | `string \| valueFrom` | yes |  | AzureStorageDataLakeGen2Filesystem (`status.outputs.filesystem_id`) |
| `spec.dataLakeGen2.tenantId` | `string` |  |  |  |
| `spec.dataLakeGen2.clientId` | `string` |  |  |  |
| `spec.dataLakeGen2.clientSecret` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.dataLakeGen2.authorityUrl` | `string` |  |  |  |
| `spec.fileShare` | `AzureMachineLearningDatastoreFileShare` |  |  |  |
| `spec.fileShare.storageFileshareId` | `string \| valueFrom` | yes |  | AzureStorageShare (`status.outputs.share_id`) |
| `spec.fileShare.accountKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.fileShare.sharedAccessSignature` | `string \| valueFrom` (sensitive) |  |  |  |

## Field Details

### spec.workspaceId

`string | valueFrom` · required

The Machine Learning workspace the datastore is registered on, by
ARM ID. Fixed at creation.

- references: AzureMachineLearningWorkspace (`status.outputs.machine_learning_workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningWorkspace, name: <that resource's name>, fieldPath: status.outputs.machine_learning_workspace_id}} -- a bare string does not parse

### spec.name

`string` · required

The datastore's name, unique on the workspace: 1-255 characters,
starting with an alphanumeric character, then alphanumerics,
hyphens or underscores (the provider's own rule) -- what jobs and
data assets reference as their datastore. Changing the name
replaces the datastore.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_-]{0,254}$"}}

### spec.description

`string`

What the datastore is for. Fixed at creation (the provider's own
contract -- changing the description replaces the datastore).

### spec.serviceDataIdentity

`enum`

How the WORKSPACE SERVICE reaches the data for service-side
operations (data profiling, previews) -- distinct from the
credentials jobs use. Unspecified applies the provider default,
NONE. The workspace-identity modes require the corresponding
identity to hold data access on the storage target.

Allowed values (use exactly as shown):

- `azure_machine_learning_datastore_service_data_identity_unspecified` -- Not specified: the provider applies "None".
- `NONE` -- No service-side identity access (wire value "None") -- the blob variant then requires explicit credentials.
- `WORKSPACE_SYSTEM_ASSIGNED_IDENTITY` -- The workspace's system-assigned identity (wire value "WorkspaceSystemAssignedIdentity").
- `WORKSPACE_USER_ASSIGNED_IDENTITY` -- The workspace's user-assigned identity (wire value "WorkspaceUserAssignedIdentity").

### spec.tags

`map<string, string>`

Free-form tags applied to the datastore object, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Fixed at
creation (the provider's own contract).

### spec.blobStorage

`AzureMachineLearningDatastoreBlobStorage`

The blob-container variant: the datastore points at a blob
container. The only variant where is_default can be set.

### spec.blobStorage.storageContainerId

`string | valueFrom` · required

The blob container the datastore points at, by ARM ID
(.../blobServices/default/containers/{name}). Fixed at creation.

- references: AzureStorageContainer (`status.outputs.container_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageContainer, name: <that resource's name>, fieldPath: status.outputs.container_id}} -- a bare string does not parse

### spec.blobStorage.isDefault

`bool`

Make this the workspace's default datastore (where job outputs
land unless directed elsewhere). Only settable on the blob
variant -- the other variants read it back from the service.

### spec.blobStorage.accountKey

`string | valueFrom` · sensitive

The storage account's access key. Reference a secret rather than
embedding the literal. When both account_key and
shared_access_signature are set, the provider sends the SAS token
(its own precedence). ARM never returns the value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.blobStorage.sharedAccessSignature

`string | valueFrom` · sensitive

A SAS token scoped to the container. Reference a secret rather
than embedding the literal. ARM never returns the value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.dataLakeGen2

`AzureMachineLearningDatastoreDataLakeGen2`

The Data Lake Gen2 variant: the datastore points at a filesystem
on a hierarchical-namespace storage account. Authenticates with a
service principal or workspace identity (no account key / SAS).

- rule: tenant_id, client_id and client_secret must be set together (service-principal auth) or all left unset

### spec.dataLakeGen2.storageContainerId

`string | valueFrom` · required

The Data Lake Gen2 filesystem the datastore points at, by ARM ID
(.../blobServices/default/containers/{name} -- a filesystem IS a
container on a hierarchical-namespace account). Fixed at
creation.

- references: AzureStorageDataLakeGen2Filesystem (`status.outputs.filesystem_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageDataLakeGen2Filesystem, name: <that resource's name>, fieldPath: status.outputs.filesystem_id}} -- a bare string does not parse

### spec.dataLakeGen2.tenantId

`string`

Service-principal auth: the Entra ID tenant ID. All three of
tenant_id, client_id and client_secret come together or not at
all; leave all unset to use workspace-identity or no credentials.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.dataLakeGen2.clientId

`string`

Service-principal auth: the application (client) ID.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.dataLakeGen2.clientSecret

`string | valueFrom` · sensitive

Service-principal auth: the client secret. Reference a secret
rather than embedding the literal. ARM never returns the value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.dataLakeGen2.authorityUrl

`string`

The authority URL used for service-principal authentication --
only for sovereign/custom clouds; leave unset for public Azure.

### spec.fileShare

`AzureMachineLearningDatastoreFileShare`

The Azure Files variant: the datastore points at a file share.
Requires exactly one of account key or SAS token (the provider's
own contract -- identity modes do not relax it here).

- rule: exactly one of account_key or shared_access_signature must be set on the file_share variant

### spec.fileShare.storageFileshareId

`string | valueFrom` · required

The file share the datastore points at, by ARM ID
(.../fileServices/default/shares/{name}). Fixed at creation.

- references: AzureStorageShare (`status.outputs.share_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageShare, name: <that resource's name>, fieldPath: status.outputs.share_id}} -- a bare string does not parse

### spec.fileShare.accountKey

`string | valueFrom` · sensitive

The storage account's access key. Reference a secret rather than
embedding the literal. ARM never returns the value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.fileShare.sharedAccessSignature

`string | valueFrom` · sensitive

A SAS token scoped to the share. Reference a secret rather than
embedding the literal. ARM never returns the value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Validation Rules

- `exactly_one_variant`: exactly one of blob_storage, data_lake_gen2 or file_share must be set -- the block is the datastore type
- `blob_auth_requires_credentials_when_identity_none`: the blob_storage variant requires account_key or shared_access_signature when service_data_identity is NONE (or unset)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningDatastore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.datastore_id` | `string` | The Azure Resource Manager ID of the datastore. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/dataStores/{name} |
| `status.outputs.datastore_name` | `string` | The datastore's name -- what jobs and data assets reference within the workspace. |
| `status.outputs.is_default` | `bool` | Whether this datastore is the workspace's default. Settable only on the blob variant; the other variants read the service's answer. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.workspaceId` | AzureMachineLearningWorkspace | `status.outputs.machine_learning_workspace_id` |
| `spec.blobStorage.storageContainerId` | AzureStorageContainer | `status.outputs.container_id` |
| `spec.dataLakeGen2.storageContainerId` | AzureStorageDataLakeGen2Filesystem | `status.outputs.filesystem_id` |
| `spec.fileShare.storageFileshareId` | AzureStorageShare | `status.outputs.share_id` |

## See Also

- [Overview](../README.md)

# AzureStorageContainer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureStorageContainerSpec** defines the configuration for creating a
blob container inside an Azure Storage Account: the namespace unit of
blob storage. Applications organize objects into containers the way
filesystems use top-level directories -- one per data domain (uploads,
logs, backups, artifacts) -- and Azure scopes public access, RBAC data
roles, encryption scopes, and lifecycle prefixes at the container level.

Containers are many-per-account with independent lifecycles, which is why
they are a first-class kind referencing the account rather than a list
folded into the account's spec. The parent is fixed at creation: a
container cannot move between accounts.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageContainer
metadata:
  name: test-storage-container
spec:
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/plantonhackstorage
  containerName: hack-uploads
  # Exercises the access-type enum mapping (blob = anonymous blob reads,
  # no listing -- the CDN-origin pattern).
  containerAccessType: BLOB
  # Exercises the encryption-scope pair.
  defaultEncryptionScope:
    value: hackscope
  encryptionScopeOverrideEnabled: false
  metadata:
    purpose: hack-validation
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.containerName` | `string` | yes |  |  |
| `spec.containerAccessType` | `enum` |  |  |  |
| `spec.defaultEncryptionScope` | `string \| valueFrom` |  |  | AzureStorageEncryptionScope (`status.outputs.encryption_scope_name`) |
| `spec.encryptionScopeOverrideEnabled` | `bool` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.storageAccountId

`string | valueFrom` · required

The storage account the container lives in, by ARM ID. References an
AzureStorageAccount's storage_account_id output so the account and its
containers compose in one manifest set. Fixed at creation.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.containerName

`string` · required

The container's name: 3-63 lowercase letters, digits, and hyphens;
starts with a letter or digit; no consecutive hyphens. Unique within
the account (it becomes the URL path segment:
https://{account}.blob.core.windows.net/{name}). Changing the name
replaces the container.

- rule: container_name must be 3-63 lowercase letters, digits, and hyphens, starting and ending with a letter or digit, with no consecutive hyphens
- rule: {"required":true,"string":{"minLen":"3","maxLen":"63"}}

### spec.containerAccessType

`enum`

Who can read the container WITHOUT credentials. Unspecified means
PRIVATE (no anonymous access -- the right posture for everything that
isn't a public website/CDN origin). Anonymous access also requires
the account's allow_nested_items_to_be_public to be true; when the
account forbids it, this field is forced to private regardless.

Allowed values (use exactly as shown):

- `azure_storage_container_access_type_unspecified` -- Not specified: PRIVATE.
- `PRIVATE` -- No anonymous access -- every read requires authorization. The recommended posture.
- `BLOB` -- Anonymous reads of BLOBS by direct URL, but no container listing. The public-website/CDN-origin pattern.
- `CONTAINER` -- Anonymous reads AND container listing -- anyone can enumerate every blob. Rarely appropriate; prefer BLOB.

### spec.defaultEncryptionScope

`string | valueFrom`

The encryption scope applied to blobs that don't name their own --
sub-account key isolation (e.g. per-tenant keys inside one account).
References an AzureStorageEncryptionScope's name output; the scope
must live on the SAME account as the container. Fixed at creation.

- references: AzureStorageEncryptionScope (`status.outputs.encryption_scope_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageEncryptionScope, name: <that resource's name>, fieldPath: status.outputs.encryption_scope_name}} -- a bare string does not parse

### spec.encryptionScopeOverrideEnabled

`bool` · optional (explicit presence)

Whether individual blob writes may OVERRIDE the default encryption
scope with their own. Azure's default is true; set false to make the
default scope mandatory for every blob in the container. Only
meaningful with default_encryption_scope. Fixed at creation.

### spec.metadata

`map<string, string>`

Free-form metadata key/value pairs stored on the container (visible
to anyone who can read container properties -- not for secrets).
Keys must be valid C# identifiers per Azure's rule and lowercase
(Azure lowercases keys on read). Hyphens fail at apply
("MetaData must start with letters or an underscores and be all
lowercase" -- live-caught on Pulumi when a fixture used e2e-suite).

## Validation Rules

- `storage_container_scope_override_needs_scope`: encryption_scope_override_enabled is only meaningful when default_encryption_scope is set
- `storage_container_metadata_keys_are_lowercase_csharp_identifiers`: container metadata keys must be lowercase C# identifiers (letter-or-underscore, then letters/digits/underscores -- no hyphens)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureStorageContainer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.container_id` | `string` | The Azure Resource Manager ID of the container. Role assignments (Storage Blob Data Reader/Contributor) scope to it for container-level data access. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Storage/storageAccounts/{account}/blobServices/default/containers/{name} |
| `status.outputs.container_name` | `string` | The container's name -- what SDKs, function bindings, and app settings reference within the account. |
| `status.outputs.storage_account_name` | `string` | The name of the storage account the container lives in, parsed from the resolved account ID -- saves consumers a second reference when they need the account/container pair. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.defaultEncryptionScope` | AzureStorageEncryptionScope | `status.outputs.encryption_scope_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureEventHub | `spec.captureDescription.destination.blobContainerName` | `status.outputs.container_name` |
| AzureMachineLearningDatastore | `spec.blobStorage.storageContainerId` | `status.outputs.container_id` |
| AzureStorageLocalUser | `spec.permissionScopes[].resourceName` | `status.outputs.container_name` |
| AzureStorageObjectReplication | `spec.rules[].sourceContainerName` | `status.outputs.container_name` |
| AzureStorageObjectReplication | `spec.rules[].destinationContainerName` | `status.outputs.container_name` |
| KubernetesLangfuse | `spec.blobStorage.target.azureBlob.container` | `status.outputs.container_name` |
| KubernetesLangfuse | `spec.blobStorage.mediaUpload.target.azureBlob.container` | `status.outputs.container_name` |
| KubernetesLangfuse | `spec.blobStorage.batchExport.target.azureBlob.container` | `status.outputs.container_name` |

## See Also

- [Overview](../README.md)

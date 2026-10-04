# AwsFsxDataRepositoryAssociation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxDataRepositoryAssociationSpec defines the desired configuration for an
Amazon FSx data repository association — the link between a directory on an
FSx for Lustre file system and an S3 bucket or prefix.

Data repository associations are the modern S3 integration for Lustre (and
the ONLY one available on PERSISTENT_2 file systems): each association maps
one file-system path to one S3 data repository, with independent
bidirectional sync policies — auto-import keeps the Lustre namespace in
sync as objects change in S3, auto-export writes file-system changes back
to the bucket. A file system supports up to 8 associations (25 per
account), each with its own lifecycle: create, update, and delete
associations without ever touching the file system.

Typical composition: an AwsFsxLustreFileSystem provides `file_system_id`;
one association per dataset links "s3://training-data/2026/" to
"/datasets/2026" and a second links "s3://model-artifacts/" to "/output"
with auto-export, giving compute jobs POSIX access in and results out.

Key design notes:
- `file_system_id`, `file_system_path`, and `data_repository_path` are
  ForceNew — an association's identity is the (file system, path, bucket)
  triple; changing any of them replaces the association.
- The sync policies and `imported_file_chunk_size` update in place.
- Only Lustre file systems WITHOUT the legacy in-spec import_path arm can
  carry associations (AWS forbids mixing the two S3-link generations).
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxDataRepositoryAssociation
metadata:
  org: example-org
  env: dev
  name: my-training-data-link
  id: awsfxdra-my-training-data-link-dev
spec:
  region: us-west-2
  file_system_id:
    value: fs-0123456789abcdef0
  file_system_path: /datasets/2026
  data_repository_path: s3://example-training-data/2026/
  auto_import_events:
    - NEW
    - CHANGED
    - DELETED
  auto_export_events:
    - NEW
    - CHANGED
  imported_file_chunk_size: 2048
  batch_import_meta_data_on_create: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.fileSystemId` | `string \| valueFrom` | yes |  | AwsFsxLustreFileSystem (`status.outputs.file_system_id`) |
| `spec.fileSystemPath` | `string` | yes |  |  |
| `spec.dataRepositoryPath` | `string` | yes |  |  |
| `spec.autoImportEvents` | `[]string` |  |  |  |
| `spec.autoExportEvents` | `[]string` |  |  |  |
| `spec.importedFileChunkSize` | `int32` |  |  |  |
| `spec.batchImportMetaDataOnCreate` | `bool` |  |  |  |
| `spec.deleteDataInFilesystem` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the association will be created — the file system's
region. Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.fileSystemId

`string | valueFrom` · required

The FSx for Lustre file system the association attaches to. Required.
ForceNew. The file system must not use the legacy in-spec S3 link
(import_path) — AWS forbids mixing the two generations.

- references: AwsFsxLustreFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsFsxLustreFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.fileSystemPath

`string` · required

The path on the file system that maps to the data repository, beginning
with "/" (e.g., "/datasets/2026" or "/" for the whole namespace).
1-4096 characters. ForceNew.

Paths must not overlap between associations on the same file system —
each directory subtree belongs to at most one repository.

- rule: file_system_path must begin with '/' (e.g., '/datasets') and be at most 4096 characters
- rule: {"required":true}

### spec.dataRepositoryPath

`string` · required

The S3 data repository URI, with an optional prefix (e.g.,
"s3://training-data" or "s3://training-data/2026/"). 3-900 characters.
ForceNew.

- rule: data_repository_path must be an S3 URI beginning with s3:// (3-900 characters)
- rule: {"required":true}

### spec.autoImportEvents

`[]string`

S3 events that automatically IMPORT metadata into the Lustre namespace —
how the file system tracks the bucket after creation.

- "NEW": objects added to the bucket appear as files.
- "CHANGED": changed objects refresh their file metadata.
- "DELETED": deleted objects remove their files.

Empty means no automatic import (the namespace reflects the bucket only
as of creation, or via manual import tasks). Updates in place.

- rule: {"repeated":{"maxItems":"3","unique":true}}

### spec.autoExportEvents

`[]string`

File-system events that automatically EXPORT back to S3 — how the bucket
tracks the file system.

- "NEW": new files are written to the bucket.
- "CHANGED": changed file contents/metadata update their objects.
- "DELETED": deleted files remove their objects.

Empty means no automatic export (write results back with manual export
tasks). Updates in place.

- rule: {"repeated":{"maxItems":"3","unique":true}}

### spec.importedFileChunkSize

`int32` · optional (explicit presence)

Stripe configuration for imported files: the maximum amount of data per
file (in MiB) stored on a single physical disk. Range: 1-512000; AWS
defaults to 1024. Updates in place.

- rule: {"int32":{"lte":512000,"gte":1}}

### spec.batchImportMetaDataOnCreate

`bool`

Run a batch import of the existing S3 metadata into the file-system path
as soon as the association is created — without it, only objects that
change AFTER creation (per auto_import_events) appear in the namespace.
Create-time behavior.

### spec.deleteDataInFilesystem

`bool`

Delete the data in the file-system path when the association is deleted.
By default the files remain on the file system (only the S3 link goes
away). Delete-time behavior; use deliberately.

## Validation Rules

- `auto_import_events_valid`: auto_import_events entries must be 'NEW', 'CHANGED', or 'DELETED'
- `auto_export_events_valid`: auto_export_events entries must be 'NEW', 'CHANGED', or 'DELETED'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxDataRepositoryAssociation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.association_id` | `string` | The AWS-assigned association ID (dra-...). The identifier FSx data repository task APIs and the console use to address this link. |
| `status.outputs.association_arn` | `string` | The Amazon Resource Name of the association, for IAM resource-level permissions. |
| `status.outputs.file_system_id` | `string` | The ID of the Lustre file system the association is attached to (fs-...), echoed for downstream composition without re-resolving the file-system reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.fileSystemId` | AwsFsxLustreFileSystem | `status.outputs.file_system_id` |

## See Also

- [Overview](../README.md)

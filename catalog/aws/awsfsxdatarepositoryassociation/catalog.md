# AWS FSx Data Repository Association

Links a directory on an FSx for Lustre file system to an S3 bucket or prefix with independent bidirectional sync policies -- auto-import keeps the Lustre namespace in sync as objects change in S3, and auto-export writes file-system changes back to the bucket. Associations are the modern S3 integration for Lustre and the only one available on PERSISTENT_2 file systems. Each association has its own lifecycle: create, update, and delete links without ever touching the file system.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data Repository Association** -- the link between one file-system path (e.g., `/datasets/2026`) and one S3 data repository (e.g., `s3://training-data/2026/`). A file system supports up to 8 associations (25 per account)
- **Automatic Import Policy** -- configured only when `autoImportEvents` is non-empty; NEW/CHANGED/DELETED object events in S3 create, refresh, or remove the corresponding Lustre file metadata
- **Automatic Export Policy** -- configured only when `autoExportEvents` is non-empty; NEW/CHANGED/DELETED file events on Lustre write back to the bucket
- **Optional Batch Import** -- when `batchImportMetaDataOnCreate` is enabled, a full metadata import of the existing repository runs at creation so pre-existing objects appear immediately

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An FSx for Lustre file system** that does NOT use the legacy in-spec S3 link (`import_path`) -- AWS forbids mixing the two S3-integration generations. Reference an AwsFsxLustreFileSystem Infra Component or provide an fs-... id directly. PERSISTENT_2 file systems accept only associations.
- **An S3 bucket** (or prefix) to link. The bucket may live in another region, but same-region buckets avoid inter-region transfer charges on every sync event.
- **Non-overlapping paths** -- each directory subtree on the file system can belong to at most one repository.

## Deploy

### Console

Open the deployment store, find **AWS FSx Data Repository Association**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Training Data Import Link** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxDataRepositoryAssociation
metadata:
  name: training-data-link
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  fileSystemId:
    valueFrom:
      kind: AwsFsxLustreFileSystem
      name: training-fs
      fieldPath: status.outputs.file_system_id
  fileSystemPath: /datasets/2026
  dataRepositoryPath: s3://training-data/2026/
  autoImportEvents:
    - NEW
    - CHANGED
    - DELETED
  batchImportMetaDataOnCreate: true
```

```shell
planton apply -f dra.yaml
```

This links `/datasets/2026` on the file system to the bucket prefix, imports the existing metadata at creation, and keeps the namespace tracking the bucket as objects are added, changed, or deleted. An Infra Job tracks the provisioning in real time.

### InfraChart

When the association deploys alongside its file system in one chart, wire the file-system reference via ValueFromRef -- here an export link writing an output directory back to S3:

```yaml
spec:
  fileSystemId:
    valueFrom:
      kind: AwsFsxLustreFileSystem
      name: training-fs
      fieldPath: status.outputs.file_system_id
  fileSystemPath: /output
  dataRepositoryPath: s3://model-artifacts/
  autoExportEvents:
    - NEW
    - CHANGED
```

The InfraPipeline resolves the dependency graph, deploys the file system first, then creates the association on it.

## Key Configuration

These are the most important decisions when configuring a data repository association. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The identity triple** -- `fileSystemId`, `fileSystemPath`, and `dataRepositoryPath` ARE the association: all three are ForceNew, and changing any of them replaces the link (a new dra-... id; the data on both sides is untouched). The file-system path must begin with `/`; the repository is an `s3://` URI with an optional prefix.

**Sync directions are independent** -- `autoImportEvents` (S3 → Lustre) and `autoExportEvents` (Lustre → S3) each take any subset of NEW/CHANGED/DELETED and update in place. Empty on both sides is a real state: the link reflects the bucket only as of creation (or a batch import), kept current with manual import/export tasks. An input dataset typically imports everything and exports nothing; an output directory does the opposite. Deletion export is a deliberate choice -- with it on, removing a file removes its S3 object.

**Batch import at creation** -- without `batchImportMetaDataOnCreate`, only objects that change AFTER creation appear in the namespace; existing datasets stay invisible until a manual import task. Turn it on when linking a bucket that already holds data.

**Delete-time behavior** -- by default, deleting the association only removes the S3 link and the files remain. `deleteDataInFilesystem` inverts that for scratch/cache directories whose contents are pure copies of S3 data.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsFsxLustreFileSystem** | `fileSystemId` | `status.outputs.file_system_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `association_id` | AWS association identifier (dra-...) | Manual data-repository tasks, AWS console/CLI operations |
| `association_arn` | Amazon Resource Name of the association | IAM policies for resource-level permissions |
| `file_system_id` | The linked Lustre file system's id, echoed | Downstream composition without re-resolving the Lustre reference |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Training data import** -- link an input path to a dataset prefix with all three import events and a creation-time batch import. Compute jobs get POSIX access to the bucket's contents, kept current automatically. Start from the **Training Data Import Link** preset.

**Results export** -- link an output path to an artifacts bucket with NEW and CHANGED export events. Job outputs land in S3 without any copy step. Start from the **Results Export Link** preset.

**Both on one file system** -- the canonical ML topology pairs an import link (`/datasets` ← training data) with an export link (`/output` → artifacts) on the same PERSISTENT_2 file system; jobs mount Lustre once and see both.

## Works With

- [**AWS FSx Lustre File System**](/infra-catalog/aws-fsx-lustre-file-system) -- the file system the association attaches to; provides `file_system_id`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the data repository backing the linked path

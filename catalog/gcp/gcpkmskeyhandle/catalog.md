# GCP KMS Key Handle

Gets a customer-managed encryption key from Cloud KMS Autokey for a resource you are about to create. Name the resource type and location, and Autokey creates an HSM key following Google's recommended settings, gives the service permission to use it, and returns the key -- no key ring design, no IAM grants, no key administrator in the loop.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `cloudkms.googleapis.com` on the handle's project
- **Key handle** -- one `kms.KeyHandle`, and through it the Autokey key

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to create key handles in the project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP KMS Key Handle**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Bucket Key** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyHandle
metadata:
  name: orders-bucket-key
  org: acme-corp
  env: prod
spec:
  projectId:
    value: orders-prod
  location: us-central1
  resourceTypeSelector: storage.googleapis.com/Bucket
```

```shell
planton apply -f kms-key-handle.yaml
```

This returns a key the orders bucket can use for CMEK. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference the handle's `status.outputs.kms_key` from the protected resource's customer-managed key field -- `GcpGcsBucket.kmsKeyName`, `GcpComputeDisk.kmsKey`, `GcpBigQueryDataset.kmsKeyName`, `GcpPubSubTopic.kmsKeyName`, and the other key fields that list `GcpKmsKeyHandle`.

## Key Configuration

These are the most important decisions when configuring a key handle. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Resource type** -- `resourceTypeSelector` names the kind of resource the key protects, as Google spells it (`storage.googleapis.com/Bucket`).

**Location** -- must match the resource's location exactly.

**Project** -- the protected resource's project; Autokey must be on for it or its folder.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `kms_key` | The Autokey key's resource name | The protected resource's customer-managed key field |
| `name` | The handle's resource name | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bucket key** -- a key for a Cloud Storage bucket. Start from the **Bucket Key** preset.

**Dataset key** -- a default key for a BigQuery dataset in a multi-region. Start from the **BigQuery Dataset Key** preset.

## Works With

- [**GCP KMS Autokey Config**](/infra-catalog/gcp-kms-autokey-config) -- turns Autokey on
- [**GCP Cloud Storage Bucket**](/infra-catalog/gcp-gcs-bucket) -- a common consumer of the key
- [**GCP BigQuery Dataset**](/infra-catalog/gcp-bigquery-dataset) -- a dataset's default key
- [**GCP Compute Disk**](/infra-catalog/gcp-compute-disk) -- a disk's key

# GcpBigQueryDataset -- Pulumi Architecture Overview

## Execution Flow

```
StackInput (GcpBigQueryDatasetStackInput)
  |
  +-- target: GcpBigQueryDataset (api.proto envelope)
  |     +-- metadata: CloudResourceMetadata
  |     +-- spec: GcpBigQueryDatasetSpec
  |           +-- project_id (StringValueOrRef -> GcpProject)
  |           +-- dataset_id
  |           +-- location
  |           +-- friendly_name, description
  |           +-- default_table_expiration_ms, default_partition_expiration_ms
  |           +-- max_time_travel_hours
  |           +-- is_case_insensitive, default_collation
  |           +-- storage_billing_model
  |           +-- delete_contents_on_destroy
  |           +-- kms_key_name (StringValueOrRef -> GcpKmsKey)
  |           +-- access[] (repeated GcpBigQueryDatasetAccessEntry)
  |
  +-- provider_config: GcpProviderConfig

  v module.Resources()

  1. initializeLocals() -> Locals { GcpLabels, spec ref }
  2. pulumigoogleprovider.Get() -> gcp.Provider
  3. dataset() -> bigquery.NewDataset
       +-- Maps all spec fields to DatasetArgs
       +-- Applies framework GcpLabels
       +-- Conditionally sets optional fields (non-zero/non-empty check)
       +-- Maps CMEK to DefaultEncryptionConfigurationArgs
       +-- Maps access entries to DatasetAccessTypeArray
       +-- Exports dataset_id, self_link, project, creation_time
```

## Resource Mapping

| Spec Field | Pulumi Property | Notes |
|------------|-----------------|-------|
| `project_id` | `Project` | From StringValueOrRef.GetValue() |
| `dataset_id` | `DatasetId` | Required, immutable |
| `location` | `Location` | Immutable after creation |
| `friendly_name` | `FriendlyName` | Optional |
| `description` | `Description` | Optional |
| `default_table_expiration_ms` | `DefaultTableExpirationMs` | int64 -> int conversion |
| `default_partition_expiration_ms` | `DefaultPartitionExpirationMs` | int64 -> int conversion |
| `max_time_travel_hours` | `MaxTimeTravelHours` | int32 -> string conversion |
| `is_case_insensitive` | `IsCaseInsensitive` | Immutable after creation |
| `default_collation` | `DefaultCollation` | Optional |
| `storage_billing_model` | `StorageBillingModel` | LOGICAL or PHYSICAL |
| `delete_contents_on_destroy` | `DeleteContentsOnDestroy` | Safety flag |
| `kms_key_name` | `DefaultEncryptionConfiguration.KmsKeyName` | CMEK |
| `access[]` | `Accesses` | Array of DatasetAccessTypeArgs |
| (framework) | `Labels` | Computed from metadata |

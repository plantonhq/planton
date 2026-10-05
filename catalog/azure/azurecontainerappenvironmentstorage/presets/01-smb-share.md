# SMB Azure Files Share (Read-Write)

This preset registers a standard SMB Azure Files share on a Container App Environment as read-write working storage. Apps and jobs then mount it by declaring an `AZURE_FILE` volume whose `storage_name` references this registration -- one registration can back volumes in any number of workloads.

## When to Use

- Persistent working storage that survives replica restarts and is shared across replicas
- Content directories, upload staging areas, shared caches
- The common case -- SMB works in both external and VNet-injected environments

## Key Configuration Choices

- **Read-write access** (`accessMode: READ_WRITE`) -- Workloads can write; use READ_ONLY for shared configuration or reference data
- **SMB path** (`accountName` + `accessKey`) -- The share is addressed by storage account name and authenticated with an account key; the key is the one field that rotates in place
- **References over literals** -- `shareName` and `accountName` resolve from `AzureStorageShare` / `AzureStorageAccount` outputs when composed in an Infra Chart

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<container-app-environment-id>` | ARM ID of the Container App Environment | `AzureContainerAppEnvironment` status outputs |
| `<azure-files-share-name>` | Name of the Azure Files share | `AzureStorageShare` status outputs (share_name) |
| `<storage-account-name>` | Name of the storage account holding the share | `AzureStorageAccount` status outputs (storage_account_name) |
| `<storage-account-access-key>` | Account access key for the SMB mount | `AzureStorageAccount` status outputs (primary_access_key) |

## Related Presets

- **02-nfs-share** -- Use instead for NFS shares in VNet-injected environments (premium FileStorage accounts)

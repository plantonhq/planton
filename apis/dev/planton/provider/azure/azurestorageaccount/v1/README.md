# Overview

The **Azure Storage Account API Resource** provides a consistent and standardized interface for deploying and managing Azure Storage Accounts within our infrastructure. This resource simplifies the process of creating storage accounts with blob containers, configuring replication strategies, and managing network access controls.

## Purpose

We developed this API resource to streamline the deployment of Azure Storage Accounts across various applications and services. By offering a unified interface, it reduces the complexity involved in storage management, enabling users to:

- **Create Storage Accounts**: Effortlessly provision Azure Storage Accounts with best-practice defaults.
- **Configure Replication**: Choose from multiple replication strategies (LRS, ZRS, GRS, GZRS) based on durability requirements.
- **Manage Blob Containers**: Create and configure blob containers with appropriate access levels.
- **Control Network Access**: Restrict storage access to specific IP ranges and Virtual Networks.
- **Enable Data Protection**: Configure soft delete and versioning for blob data recovery.

## Key Features

- **Consistent Interface**: Aligns with our existing APIs for deploying cloud infrastructure and services.
- **Simplified Configuration**: Abstracts the complexities of Azure Storage, enabling quicker setups without deep Azure expertise.
- **Multiple Storage Types**: Support for StorageV2, BlobStorage, BlockBlobStorage, and FileStorage.
- **Flexible Replication**: Choose from locally redundant (LRS) to geo-zone-redundant (GZRS) storage.
- **Access Tier Optimization**: Configure Hot or Cool access tiers for cost optimization.
- **Security First**: HTTPS-only traffic, TLS 1.2 minimum, and network ACL controls.
- **Data Protection**: Built-in support for blob versioning and soft delete.

## Spec Fields (Key Configuration)

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Azure region for deployment (e.g., "eastus", "westus2") |
| `resource_group` | string | Resource group name (must exist) |
| `account_kind` | enum | Storage account kind: STORAGE_V2, BLOB_STORAGE, BLOCK_BLOB_STORAGE, FILE_STORAGE |
| `account_tier` | enum | Performance tier: STANDARD or PREMIUM |
| `replication_type` | enum | Replication: LRS, ZRS, GRS, GZRS, RA_GRS, RA_GZRS |
| `access_tier` | enum | Blob access tier: HOT or COOL |
| `enable_https_traffic_only` | bool | Require HTTPS for all requests (default: true) |
| `min_tls_version` | enum | Minimum TLS version: TLS1_0, TLS1_1, TLS1_2 |
| `network_rules` | object | Network ACL configuration (IP rules, VNet rules) |
| `blob_properties` | object | Blob service properties (versioning, soft delete) |
| `containers` | list | Blob containers to create |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `storage_account_id` | Azure Resource Manager ID of the storage account |
| `storage_account_name` | Name of the storage account |
| `primary_blob_endpoint` | Primary blob service endpoint URL |
| `primary_queue_endpoint` | Primary queue service endpoint URL |
| `primary_table_endpoint` | Primary table service endpoint URL |
| `primary_file_endpoint` | Primary file service endpoint URL |
| `primary_dfs_endpoint` | Primary Data Lake Storage endpoint URL |
| `primary_web_endpoint` | Primary static website endpoint URL |
| `container_url_map` | Map of container names to their URLs |
| `region` | Azure region where deployed |
| `resource_group` | Resource group name |

## How It Works

This resource deploys Azure Storage Accounts using:

- **Pulumi Module**: `iac/pulumi/module/` - Go-based deployment using Pulumi Azure SDK
- **Terraform Module**: `iac/tf/` - HCL-based deployment using Azure Provider

Both modules provide feature parity and create identical resources.

## Use Cases

- **Application Data Storage**: Store application data, logs, and backups
- **Static Website Hosting**: Host static websites using blob storage
- **Data Lake**: Build data lakes using Data Lake Storage Gen2
- **File Shares**: Provide SMB file shares for applications
- **Archive Storage**: Long-term retention of infrequently accessed data
- **Container Data**: Store container images and artifacts

## References

- [Azure Storage Account Overview](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-overview)
- [Storage Replication Options](https://learn.microsoft.com/en-us/azure/storage/common/storage-redundancy)
- [Blob Access Tiers](https://learn.microsoft.com/en-us/azure/storage/blobs/access-tiers-overview)
- [Network Security for Storage](https://learn.microsoft.com/en-us/azure/storage/common/storage-network-security)
- [Blob Soft Delete](https://learn.microsoft.com/en-us/azure/storage/blobs/soft-delete-blob-overview)

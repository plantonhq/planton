# AliCloudLogProject

Manages an Alibaba Cloud Simple Log Service (SLS) project with optional bundled log stores and full-text indexes.

## Overview

An SLS project is the top-level container for log data in Alibaba Cloud. This component creates the project and optionally provisions log stores within it. Each log store can have a full-text search index created automatically, making ingested logs immediately searchable.

### What Gets Created

- **SLS Project** -- the regional container for log data
- **Log Stores** (optional) -- individual storage units within the project for collecting and querying logs
- **Full-Text Indexes** (optional, per store) -- enables log search with case-insensitive matching and standard tokenization

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`) |
| `projectName` | string | Globally unique SLS project name (3-63 chars, lowercase letters, digits, hyphens) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable project description |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags applied to the project |
| `logStores` | list | `[]` | Log stores to create within the project |

### Log Store Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | (required) | Log store name (3-63 chars) |
| `retentionDays` | int | `30` | Data retention in days (1-3650; 3650 = permanent) |
| `shardCount` | int | `2` | Number of write shards |
| `autoSplit` | bool | `true` | Auto-split shards when throughput exceeds capacity |
| `maxSplitShardCount` | int | `64` | Maximum shards after auto-splitting |
| `enableIndex` | bool | `true` | Create a full-text search index |
| `appendMeta` | bool | `true` | Append receive time and client IP to log entries |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `project_name` | The SLS project name |
| `project_id` | The SLS project resource ID |
| `log_store_names` | Map of log store names created within the project |

## Related Components

- **AliCloudAckManagedCluster** -- references this project for cluster audit and event logging
- **AliCloudFcFunction** -- references this project for function execution logging
- **AliCloudSaeApplication** -- references this project for application logging

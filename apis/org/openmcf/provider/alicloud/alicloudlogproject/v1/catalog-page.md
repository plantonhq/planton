# AliCloud Log Project

Deploys an Alibaba Cloud Simple Log Service (SLS) project with bundled log stores and full-text indexes. The component provisions the project, creates each specified log store, and enables full-text search indexing per store by default — ensuring logs are immediately queryable after ingestion.

## What Gets Created

When you deploy an AliCloudLogProject resource, OpenMCF provisions:

- **SLS Project** — the regional container for log data, created with the specified name, description, resource group, and tags
- **Log Stores** — one `alicloud_log_store` per entry in `logStores`, each with configurable retention, shard count, auto-split, and metadata enrichment
- **Full-Text Store Indexes** — one `alicloud_log_store_index` per log store where `enableIndex` is true (the default), configured with case-insensitive matching and standard tokenization

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`) or OpenMCF provider config
- **A globally unique project name** — SLS project names are unique across all Alibaba Cloud accounts within a region

## Quick Start

Create a file `log-project.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: my-log-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudLogProject.my-log-project
spec:
  region: cn-hangzhou
  projectName: my-app-logs
  logStores:
    - name: app-logs
```

Deploy:

```shell
openmcf apply -f log-project.yaml
```

This creates an SLS project named `my-app-logs` in `cn-hangzhou` with one log store (`app-logs`) and a full-text search index on that store.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region where the SLS project will be created (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `projectName` | `string` | Globally unique SLS project name. Lowercase letters, digits, and hyphens only. Must start and end with a letter or digit. | Required; 3-63 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the project. |
| `resourceGroupId` | `string` | `""` | Alibaba Cloud resource group ID for organizational grouping. If omitted, the default resource group is used. |
| `tags` | `map<string, string>` | `{}` | Key-value tags applied to the SLS project. Merged with standard OpenMCF tags. |
| `logStores` | `AliCloudLogStore[]` | `[]` | Log stores to create within this project. See fields below. |
| `logStores[].name` | `string` | — | Log store name. Must be unique within the project. (Required per store; 3-63 characters) |
| `logStores[].retentionDays` | `int` | `30` | Data retention period in days. Range: 1-3650. Set to 3650 for permanent retention. |
| `logStores[].shardCount` | `int` | `2` | Number of write shards. Each shard supports ~5 MB/s write throughput. Range: 1-256. |
| `logStores[].autoSplit` | `bool` | `true` | Automatically split shards when write throughput exceeds capacity. |
| `logStores[].maxSplitShardCount` | `int` | `64` | Maximum shards after auto-splitting. Only effective when `autoSplit` is true. Range: 1-256. |
| `logStores[].enableIndex` | `bool` | `true` | Create a full-text search index for this store. When true, logs are immediately searchable after ingestion. |
| `logStores[].appendMeta` | `bool` | `true` | Append log receive time and client IP as metadata fields on each log entry. |

## Examples

### Minimal Project

An empty SLS project with no log stores. Stores can be added by updating the manifest.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: empty-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudLogProject.empty-project
spec:
  region: cn-hangzhou
  projectName: my-empty-project
```

### Development with Single Store

A project for a development environment with one log store using short retention and minimal shards.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: dev-logging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudLogProject.dev-logging
spec:
  region: cn-hangzhou
  projectName: dev-app-logs
  description: Development environment logging
  logStores:
    - name: app-logs
      retentionDays: 7
      shardCount: 1
```

### Production Multi-Store

Separate stores for application logs, audit trails, and access logs with distinct retention and shard configurations. Tags enable cost attribution and organizational filtering.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: prod-logging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudLogProject.prod-logging
spec:
  region: cn-shanghai
  projectName: prod-platform-logs
  description: Production platform logging
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
  logStores:
    - name: app-logs
      retentionDays: 90
      shardCount: 4
      autoSplit: true
      maxSplitShardCount: 64
      enableIndex: true
      appendMeta: true
    - name: audit-logs
      retentionDays: 365
      shardCount: 2
      enableIndex: true
    - name: access-logs
      retentionDays: 30
      shardCount: 2
      enableIndex: true
```

### Archive Store Without Indexing

A project for compliance archival where query capability is not needed. Disabling indexing eliminates index storage costs.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: archive-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudLogProject.archive-project
spec:
  region: cn-hangzhou
  projectName: compliance-archive
  logStores:
    - name: regulatory-archive
      retentionDays: 3650
      shardCount: 1
      autoSplit: false
      enableIndex: false
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `project_name` | `string` | The SLS project name (also serves as the project identifier in SLS APIs) |
| `project_id` | `string` | The SLS project resource ID |
| `log_store_names` | `map<string, string>` | Map of log store names created within the project. Key and value are both the store name. Downstream components can reference specific stores via `StringValueOrRef`. |

## Related Components

- [AliCloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) — references this project for cluster audit and event logging
- [AliCloudFcFunction](/docs/catalog/alicloud/alicloudfcfunction) — references this project for function execution logging
- [AliCloudSaeApplication](/docs/catalog/alicloud/alicloudsaeapplication) — references this project for application logging

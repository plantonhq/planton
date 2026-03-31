# AliCloudLogProject Terraform Examples

Apply any example below using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --auto-approve
```

---

## Minimal SLS Project

Creates a project with no log stores. Stores can be added later by updating the
manifest.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: my-log-project
spec:
  region: cn-hangzhou
  projectName: my-app-logging
```

This creates a single `alicloud_log_project` resource. No stores or indexes are
provisioned.

---

## Development with Single Store

A project with one log store using short retention and a single shard. Full-text
indexing is enabled by default, making logs immediately searchable.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: dev-logging
  env: development
spec:
  region: cn-hangzhou
  projectName: dev-app-logs
  description: Development environment logging
  logStores:
    - name: app-logs
      retentionDays: 7
      shardCount: 1
      enableIndex: true
```

Resources created:
- `alicloud_log_project.main`
- `alicloud_log_store.stores["app-logs"]`
- `alicloud_log_store_index.indexes["app-logs"]`

---

## Production with Multiple Stores and Tags

Separate stores for application logs, audit trails, and access logs. Each store
has distinct retention and shard configuration. Tags enable cost attribution.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: prod-logging
  org: my-org
  env: production
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

Resources created:
- `alicloud_log_project.main`
- `alicloud_log_store.stores["app-logs"]`, `["audit-logs"]`, `["access-logs"]`
- `alicloud_log_store_index.indexes["app-logs"]`, `["audit-logs"]`, `["access-logs"]`

---

## Archive Store Without Indexing

A store used purely for compliance archival where query capability is not needed.
Disabling indexing eliminates index storage costs.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudLogProject
metadata:
  name: archive-logging
  org: my-org
  env: production
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

Resources created:
- `alicloud_log_project.main`
- `alicloud_log_store.stores["regulatory-archive"]`
- No `alicloud_log_store_index` (indexing disabled)

---

## After Deploying

Confirm the project exists using the Alibaba Cloud CLI:

```shell
aliyun sls GetProject --project <project-name>
```

List log stores within the project:

```shell
aliyun sls ListLogStores --project <project-name>
```

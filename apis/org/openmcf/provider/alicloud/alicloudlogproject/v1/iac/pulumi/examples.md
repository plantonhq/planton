# AlicloudLogProject Pulumi Examples

Apply any example below using the OpenMCF CLI:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal Project

Creates an SLS project with no log stores. Stores can be added later by updating
the manifest.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudLogProject
metadata:
  name: my-log-project
spec:
  region: cn-hangzhou
  projectName: my-app-logging
```

---

## Development Environment

A project with a single log store using short retention and minimal shards.
Full-text indexing is enabled by default.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudLogProject
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

---

## Production with Multiple Log Stores

Separate stores for application logs and audit trails, each with appropriate
retention and shard configuration. Tags enable cost attribution and filtering.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudLogProject
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

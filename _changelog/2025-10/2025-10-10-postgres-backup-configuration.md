# PostgreSQL Backup and Disaster Recovery Support

**Date**: October 10, 2025  
**Type**: Feature  
**Components**: PostgresOperatorKubernetes, PostgresKubernetes

## Summary

Added comprehensive backup and disaster recovery capabilities for PostgreSQL databases deployed on Kubernetes using the Zalando Postgres Operator. Supports automated backups to S3-compatible storage (Cloudflare R2, AWS S3, MinIO, etc.) with Point-in-Time Recovery (PITR), operator-level defaults, and per-database configuration overrides.

## Motivation

PostgreSQL databases deployed on Kubernetes lacked automated backup capabilities, creating significant data loss risks. Users needed:
- Automated continuous backup of databases
- Point-in-Time Recovery for disaster scenarios
- Flexible backup configuration at both operator and database levels
- Support for S3-compatible storage providers
- Database cloning capabilities for testing and development

## What's New

### 1. Operator-Level Backup Configuration

PostgresOperatorKubernetes now supports default backup settings that apply to all managed databases:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresOperatorKubernetes
metadata:
  name: postgres-operator
spec:
  backup_config:
    r2_config:
      cloudflare_account_id: "your-account-id"
      bucket_name: "postgres-backups"
      access_key_id: "your-access-key"
      secret_access_key: "your-secret-key"
    s3_prefix_template: "backups/$(SCOPE)/$(PGVERSION)"
    backup_schedule: "0 2 * * *"
    enable_wal_g_backup: true
    enable_wal_g_restore: true
    enable_clone_wal_g_restore: true
```

**Features**:
- Automatic Secret creation for storage credentials
- ConfigMap generation with WAL-G environment variables
- Operator injects backup configuration into all database pods
- Variables `$(SCOPE)` and `$(PGVERSION)` auto-replaced with database name and version

### 2. Per-Database Backup Overrides

PostgresKubernetes resources can override operator-level backup settings:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: my-database
spec:
  container:
    replicas: 2
    resources: {...}
    disk_size: "10Gi"
  backup_config:
    s3_prefix: "my-bucket/custom/path/$(PGVERSION)"
    backup_schedule: "0 3 * * *"
    enable_backup: true
    enable_restore: true
    enable_clone: true
```

**Features**:
- Fully optional - databases inherit operator settings if not specified
- Override any combination of settings
- Manager-agnostic design (not tied to Zalando implementation)
- Disable backups for specific databases if needed

### 3. Backup Capabilities

**Continuous WAL Archiving**:
- Write-Ahead Log files streamed to S3-compatible storage in real-time
- Minimal data loss (typically seconds) in disaster scenarios

**Scheduled Base Backups**:
- Full database backups on customizable schedules
- Cron expression support (e.g., `"0 2 * * *"` for 2 AM daily)

**Point-in-Time Recovery (PITR)**:
- Restore database to any point in time
- Critical for recovering from logical errors or data corruption

**Database Cloning**:
- Create copies of databases from backups
- Essential for testing, development, and staging environments

## Implementation Details

### New API Messages

**PostgresOperatorKubernetes** (operator-level, Zalando-specific):

```protobuf
message PostgresOperatorKubernetesBackupConfig {
  PostgresOperatorKubernetesBackupR2Config r2_config = 1;
  string s3_prefix_template = 2;
  string backup_schedule = 3;
  bool enable_wal_g_backup = 4;
  bool enable_wal_g_restore = 5;
  bool enable_clone_wal_g_restore = 6;
}

message PostgresOperatorKubernetesBackupR2Config {
  string cloudflare_account_id = 1;
  string bucket_name = 2;
  string access_key_id = 3;
  string secret_access_key = 4;
}
```

**PostgresKubernetes** (per-database, manager-agnostic):

```protobuf
message PostgresKubernetesBackupConfig {
  string s3_prefix = 1;
  string backup_schedule = 2;
  optional bool enable_backup = 3;
  optional bool enable_restore = 4;
  optional bool enable_clone = 5;
}
```

### Pulumi Module Implementations

**PostgresOperatorKubernetes Module**:
- `backup_config.go`: Creates Secret and ConfigMap from manifest
- `postgres_operator.go`: Wires ConfigMap into Helm release
- Helm values updated with `pod_environment_configmap` reference

**PostgresKubernetes Module**:
- `backup_config.go`: Translates generic config to Zalando env vars
- `main.go`: Adds `Env` field to PostgreSQL CRD with overrides

### Storage Backend Support

Works with any S3-compatible storage:
- **Cloudflare R2** (fully tested)
- **AWS S3**
- **Google Cloud Storage** (S3-compatible API)
- **MinIO**
- **Ceph**
- Any S3-compatible object storage

### Backup Tool Integration

Uses **WAL-G** (Write-Ahead Log - Go):
- Industry-standard PostgreSQL backup tool
- Continuous WAL archiving
- Incremental backups
- Compression and encryption support
- Built into Zalando Spilo container images

## Architecture

### Inheritance Model

```
┌─────────────────────────────────────┐
│  PostgresOperatorKubernetes         │
│  (Operator-Level Defaults)          │
│                                     │
│  - Secret: credentials              │
│  - ConfigMap: backup env vars       │
│  - Schedule: 0 2 * * *             │
│  - Path: backups/$(SCOPE)/$(PG)    │
└──────────────┬──────────────────────┘
               │ Inherits
               ▼
┌─────────────────────────────────────┐
│  PostgresKubernetes                 │
│  (Per-Database)                     │
│                                     │
│  Option 1: No backup_config         │
│  → Inherits all operator settings   │
│                                     │
│  Option 2: With backup_config       │
│  → Overrides specified settings     │
│  → Inherits rest from operator      │
└─────────────────────────────────────┘
```

### Resource Flow

```
Manifest with backup_config
    ↓
Pulumi Module
    ↓
┌─────────────────┐
│ Kubernetes      │
│ - Secret        │ ← Operator creates
│ - ConfigMap     │ ← Operator creates
│ - PostgreSQL CR │ ← Database with env overrides
└─────────────────┘
    ↓
Zalando Operator
    ↓
┌─────────────────┐
│ Spilo Pod       │
│ - Postgres      │
│ - Patroni       │
│ - WAL-G         │ ← Configured via env vars
└─────────────────┘
    ↓
S3-Compatible Storage
```

## Usage Examples

### Basic Operator Setup

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresOperatorKubernetes
metadata:
  name: postgres-operator
spec:
  target_cluster:
    kubernetesProviderConfigId: k8s-cluster-01
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
  backup_config:
    r2_config:
      cloudflare_account_id: "abc123"
      bucket_name: "postgres-backups"
      access_key_id: "R2_ACCESS_KEY"
      secret_access_key: "R2_SECRET_KEY"
    backup_schedule: "0 2 * * *"
```

### Database with Inherited Backup

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: production-db
spec:
  container:
    replicas: 2
    disk_size: "100Gi"
# No backup_config - inherits operator settings
```

### Database with Custom Backup Path

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: critical-db
spec:
  container:
    replicas: 3
    disk_size: "500Gi"
  backup_config:
    # Custom path for compliance/audit requirements
    s3_prefix: "compliance-backups/critical/$(PGVERSION)"
    # More frequent backups
    backup_schedule: "0 */6 * * *"
```

### Database with Backups Disabled

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: temporary-db
spec:
  container:
    replicas: 1
    disk_size: "10Gi"
  backup_config:
    # Explicitly disable backups for temporary database
    enable_backup: false
    enable_restore: false
    enable_clone: false
```

## Verification

After deploying databases with backup configuration:

```bash
# Check backup environment variables in pod
POD=$(kubectl get pods -n <namespace> -l application=spilo -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD -n <namespace> -- env | grep -E "WALG|BACKUP"

# Expected output:
# USE_WALG_BACKUP=true
# USE_WALG_RESTORE=true
# CLONE_USE_WALG_RESTORE=true
# WALG_S3_PREFIX=s3://bucket/backups/db-name/15
# BACKUP_SCHEDULE=0 2 * * *
# AWS_ENDPOINT=https://...
# AWS_REGION=auto

# List available backups
kubectl exec $POD -n <namespace> -- \
  envdir /run/etc/wal-e.d/env wal-g backup-list

# Check backup files in storage bucket
# Browse to: bucket/backups/db-name/15/
# - basebackups_005/ - Full backups
# - wal_005/ - WAL files
```

## Security Considerations

- **Credentials**: Stored in Kubernetes Secrets (not in ConfigMaps)
- **Secret Management**: Secrets created automatically from manifest credentials
- **Access Control**: Use Kubernetes RBAC to restrict Secret access
- **Encryption**: WAL-G supports encryption at rest and in transit
- **Bucket Permissions**: Use least-privilege IAM policies for storage access

## Performance Impact

- **WAL Archiving**: Minimal overhead (~1-2% CPU)
- **Base Backups**: Brief I/O spike during backup, non-blocking for queries
- **Storage**: Compressed backups (typically 30-50% of database size)
- **Network**: Continuous but low bandwidth for WAL streaming

## Migration Guide

### For Existing PostgresOperatorKubernetes Deployments

1. Add `backup_config` to operator manifest
2. Deploy updated operator (creates Secret and ConfigMap)
3. Existing databases restart automatically to pick up backup configuration
4. Verify backups appearing in storage bucket

### For Existing PostgresKubernetes Databases

No changes required! Databases automatically inherit operator-level backup settings.

To add per-database overrides:
1. Add `backup_config` section to database manifest
2. Specify only the settings you want to override
3. Redeploy database

## Benefits

1. **Data Protection**: Continuous backup with minimal data loss potential
2. **Disaster Recovery**: Restore to any point in time
3. **Compliance**: Automated backup schedules meet regulatory requirements
4. **Flexibility**: Configure at operator or database level
5. **Cost Effective**: Use affordable S3-compatible storage (R2, S3, MinIO)
6. **Testing**: Clone production databases for staging/development
7. **Self-Service**: Developers can configure backups per database
8. **Operator Defaults**: Set sensible defaults once for all databases

## Future Enhancements

- Support for additional backup tools (pg_basebackup, pgBackRest)
- Backup retention policies
- Backup encryption key management
- Cross-region replication
- Backup monitoring and alerting
- Web UI for backup management
- Backup restore automation
- Backup testing schedules

## Related Documentation

- WAL-G Documentation: https://github.com/wal-g/wal-g
- Zalando Postgres Operator: https://github.com/zalando/postgres-operator
- Spilo (Postgres + Patroni + WAL-G): https://github.com/zalando/spilo

## Breaking Changes

None. This is an additive feature that is fully backward compatible.


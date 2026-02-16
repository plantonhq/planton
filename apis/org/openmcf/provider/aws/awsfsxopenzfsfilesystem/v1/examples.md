# AwsFsxOpenzfsFileSystem Examples

## 1. Minimal — Development File System

The smallest possible OpenZFS file system for development and testing.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: my-org
  env: dev
  name: dev-nfs
  id: awsfxz-dev-nfs-dev
spec:
  storage_capacity_gib: 64
  throughput_capacity: 160
  subnet_ids:
    - value: subnet-0123456789abcdef0
```

## 2. Production — SINGLE_AZ_2 with Compression and Backups

A production-ready single-AZ file system with ZSTD compression, customer-managed encryption, daily backups, and NFS exports configured for the entire VPC.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: my-org
  env: production
  name: app-data
  id: awsfxz-app-data-production
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 1024
  throughput_capacity: 640
  subnet_ids:
    - value: subnet-0123456789abcdef0
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/my-kms-key
  root_volume_configuration:
    data_compression_type: ZSTD
    record_size_kib: 128
    nfs_exports:
      client_configurations:
        - clients: "*"
          options:
            - rw
            - crossmnt
            - no_root_squash
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  copy_tags_to_backups: true
  copy_tags_to_volumes: true
  weekly_maintenance_start_time: "7:02:00"
```

## 3. Multi-AZ — High Availability with Provisioned IOPS

A mission-critical Multi-AZ deployment with automatic failover, provisioned IOPS for guaranteed performance, NFS exports restricted to a VPC CIDR, and per-user storage quotas.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: my-org
  env: production
  name: ha-shared-fs
  id: awsfxz-ha-shared-fs-production
spec:
  deployment_type: MULTI_AZ_1
  storage_capacity_gib: 2048
  throughput_capacity: 2560
  subnet_ids:
    - value: subnet-0123456789abcdef0
    - value: subnet-0987654321fedcba0
  preferred_subnet_id:
    value: subnet-0123456789abcdef0
  endpoint_ip_address_range: "198.19.255.0/24"
  route_table_ids:
    - value: rtb-0123456789abcdef0
    - value: rtb-0987654321fedcba0
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/my-kms-key
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 160000
  root_volume_configuration:
    data_compression_type: ZSTD
    record_size_kib: 128
    copy_tags_to_snapshots: true
    nfs_exports:
      client_configurations:
        - clients: "10.0.0.0/16"
          options:
            - rw
            - crossmnt
    user_and_group_quotas:
      - id: 0
        storage_capacity_quota_gib: 500
        type: USER
      - id: 1000
        storage_capacity_quota_gib: 200
        type: USER
      - id: 100
        storage_capacity_quota_gib: 1000
        type: GROUP
  automatic_backup_retention_days: 30
  daily_automatic_backup_start_time: "02:00"
  copy_tags_to_backups: true
  copy_tags_to_volumes: true
  weekly_maintenance_start_time: "7:04:00"
```

## 4. Cross-Resource Reference with valueFrom

Referencing subnets, security groups, and KMS key from other OpenMCF-managed resources using `valueFrom`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: my-org
  env: production
  name: shared-nfs
  id: awsfxz-shared-nfs-production
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 512
  throughput_capacity: 320
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        id: awsvpc-main-production
        fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        id: awssg-nfs-production
        fieldPath: status.outputs.security_group_id
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      id: awskms-data-production
      fieldPath: status.outputs.key_arn
  root_volume_configuration:
    data_compression_type: LZ4
    nfs_exports:
      client_configurations:
        - clients: "*"
          options:
            - rw
            - no_root_squash
```

## 5. Analytics Workload — Large Records, No Compression

An OpenZFS file system optimized for analytics and data processing with large sequential I/O. Uses 1024 KiB record size for optimal streaming throughput, no compression (data is already compressed or incompressible), and a read-only root volume.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: my-org
  env: production
  name: analytics-data
  id: awsfxz-analytics-data-production
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 4096
  throughput_capacity: 2560
  subnet_ids:
    - value: subnet-0123456789abcdef0
  security_group_ids:
    - value: sg-0123456789abcdef0
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 200000
  root_volume_configuration:
    data_compression_type: NONE
    record_size_kib: 1024
    read_only: true
    nfs_exports:
      client_configurations:
        - clients: "10.0.0.0/8"
          options:
            - ro
            - crossmnt
            - root_squash
  automatic_backup_retention_days: 3
  daily_automatic_backup_start_time: "06:00"
```

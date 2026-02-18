# AwsFsxLustreFileSystem Examples

Apply manifests with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

or

```shell
openmcf tofu apply --manifest <yaml-path> --auto-approve
```

Provider credentials (AWS access key, secret, region) are supplied via stack input, not in the spec.

---

## 1. Minimal Scratch (Dev/Test Fast Processing)

The simplest configuration: SCRATCH_2 with 1200 GiB SSD. No backups, no replication. Ideal for ephemeral data processing, CI pipelines, or quick experiments.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: dev-scratch-fsx
  org: my-org
spec:
  region: us-west-2
  storage_capacity_gib: 1200
  subnet_id:
    value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
```

**Outputs:** `file_system_id`, `dns_name`, `mount_name`. Mount with:

```bash
sudo mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx
```

**Note:** SCRATCH_2 provides 200 MB/s/TiB baseline throughput with burst to 1300 MB/s/TiB. Data is not replicated — hardware failure means data loss.

---

## 2. SCRATCH_2 with S3 Import (Data Processing Pipeline)

Import data from S3, process it on Lustre, and export results back to S3. Uses `valueFrom` for subnet and security group references.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: s3-pipeline-fsx
  org: my-org
  labels:
    workload: etl-pipeline
spec:
  region: us-west-2
  storage_capacity_gib: 2400
  data_compression_type: LZ4
  import_path: s3://my-data-lake/raw-datasets/
  export_path: s3://my-data-lake/processed-output/
  subnet_id:
    valueFrom:
      kind: AwsVpc
      name: data-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lustre-clients-sg
        fieldPath: status.outputs.security_group_id
```

**How it works:** File metadata is imported from S3 at creation. File data is lazy-loaded on first access (HSM). Changes on the file system are automatically exported back to S3.

**Note:** `import_path` and `export_path` are ForceNew and only supported on SCRATCH deployments. For PERSISTENT deployments, use a separate data repository association resource.

---

## 3. PERSISTENT_2 for ML Training (High Throughput, SSD)

Production-grade persistent storage for distributed ML training. 1000 MB/s/TiB throughput, LZ4 compression, customer-managed KMS key, automatic backups, and provisioned metadata IOPS.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: ml-training-fsx
  org: my-org
  labels:
    environment: production
    workload: ml-training
spec:
  region: us-west-2
  deployment_type: PERSISTENT_2
  storage_capacity_gib: 4800
  storage_type: SSD
  per_unit_storage_throughput: 1000
  data_compression_type: LZ4
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-prod-key
      fieldPath: status.outputs.key_arn
  subnet_id:
    valueFrom:
      kind: AwsVpc
      name: ml-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ml-lustre-sg
        fieldPath: status.outputs.security_group_id
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "04:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:03:00"
  metadata_configuration:
    mode: USER_PROVISIONED
    iops: 12000
  log_configuration:
    destination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: fsx-audit-logs
        fieldPath: status.outputs.log_group_arn
    level: WARN_ERROR
```

**Aggregate throughput:** 4800 GiB ≈ 4.69 TiB × 1000 MB/s/TiB ≈ 4690 MB/s. Sufficient for 8–16 GPU instances reading training data simultaneously.

**Metadata IOPS:** 12000 provisioned IOPS handles workloads that create/list many small files (e.g., checkpointing).

---

## 4. PERSISTENT_1 HDD for Data Lake (High Capacity, Lower Cost)

Large-capacity persistent storage using HDD for workloads where throughput is more important than latency: log analysis, genomics pipelines, or data lake staging.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: datalake-fsx
  org: my-org
  labels:
    environment: production
    workload: data-lake
spec:
  region: us-west-2
  deployment_type: PERSISTENT_1
  storage_capacity_gib: 12000
  storage_type: HDD
  per_unit_storage_throughput: 40
  data_compression_type: LZ4
  subnet_id:
    valueFrom:
      kind: AwsVpc
      name: analytics-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: analytics-lustre-sg
        fieldPath: status.outputs.security_group_id
  automatic_backup_retention_days: 14
  daily_automatic_backup_start_time: "02:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "1:05:00"
```

**Aggregate throughput:** 12000 GiB ≈ 11.72 TiB × 40 MB/s/TiB ≈ 469 MB/s. HDD-backed storage is significantly cheaper per GiB than SSD.

**S3 integration note:** For PERSISTENT deployments, S3 integration is done via a separate data repository association (not `import_path`/`export_path`). This allows adding, modifying, and removing S3 links without replacing the file system.

---

## 5. PERSISTENT_2 with Minimal Configuration

A production-ready PERSISTENT_2 file system with sensible defaults: 250 MB/s/TiB throughput, automatic metadata IOPS, no custom KMS key. Good starting point when you need durability without complex configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: workload-fsx
  org: my-org
  labels:
    environment: staging
spec:
  region: us-west-2
  deployment_type: PERSISTENT_2
  storage_capacity_gib: 2400
  per_unit_storage_throughput: 250
  subnet_id:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lustre-sg
        fieldPath: status.outputs.security_group_id
  automatic_backup_retention_days: 3
```

**Note:** Encryption at rest uses the AWS-managed FSx key by default. All Lustre file systems are encrypted — no opt-in required.

---

## CLI Flows

Validate manifest:

```bash
openmcf validate --manifest ./fsx-lustre.yaml
```

Get outputs after deployment:

```bash
openmcf pulumi stack output file_system_id --stack my-org/project/prod
openmcf pulumi stack output dns_name --stack my-org/project/prod
openmcf pulumi stack output mount_name --stack my-org/project/prod
```

Mount from EC2:

```bash
sudo amazon-linux-extras install -y lustre
sudo mkdir -p /mnt/fsx
sudo mount -t lustre fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com@tcp:/fsx /mnt/fsx
```

For more architecture details and integration patterns, see [docs/README.md](docs/README.md).

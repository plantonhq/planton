# AwsFsxWindowsFileSystem Examples

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

## 1. Minimal with AWS Managed AD

The simplest configuration: SINGLE_AZ_2 with 32 GiB SSD and 32 MB/s throughput. Joined to an AWS Managed Microsoft AD. Ideal for development, testing, or proof-of-concept environments.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: dev-windows-fs
  org: my-org
spec:
  storage_capacity_gib: 32
  throughput_capacity: 32
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  active_directory_id:
    value: d-0123456789
```

**Outputs:** `file_system_id`, `dns_name`, `remote_administration_endpoint`. Mount with:

```cmd
net use Z: \\<dns_name>\share
```

**Note:** Default 7-day backup retention applies. Set `automatic_backup_retention_days: 0` to disable backups for dev environments.

---

## 2. Production with Self-Managed AD and Audit Logging

Production-grade single-AZ file system with self-managed Active Directory, audit logging for compliance, customer-managed KMS encryption, and 7-day backup retention.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: prod-windows-fs
  org: my-org
  labels:
    environment: production
    department: engineering
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 500
  storage_type: SSD
  throughput_capacity: 256
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/your-kms-key-id
  self_managed_active_directory:
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.10"
      - "10.0.0.11"
    username: svc-fsx-join
    password: "<your-service-account-password>"
    file_system_administrators_group: FSx Admins
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: FAILURE_ONLY
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "01:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Audit logging:** File access events (open, read, write, delete) are logged at all levels. File share access events (connect, disconnect) capture failures only. Logs go to the default `/aws/fsx/windows` CloudWatch log group.

**Security note:** For production, prefer `domain_join_service_account_secret_arn` over inline `username`/`password` to avoid credentials in manifests.

---

## 3. Multi-AZ High Availability

Mission-critical deployment with automatic failover across two availability zones. DNS aliases enable transparent access via custom hostnames. Provisioned IOPS for consistent performance under heavy load.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: ha-windows-fs
  org: my-org
  labels:
    environment: production
    tier: critical
spec:
  deployment_type: MULTI_AZ_1
  storage_capacity_gib: 1000
  storage_type: SSD
  throughput_capacity: 512
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0b2c3d4e5f600002
  preferred_subnet_id:
    value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/your-kms-key-id
  active_directory_id:
    value: d-0123456789
  aliases:
    - finance.corp.example.com
    - shared.corp.example.com
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: SUCCESS_AND_FAILURE
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 100000
  automatic_backup_retention_days: 30
  daily_automatic_backup_start_time: "03:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Failover behavior:** In a MULTI_AZ_1 deployment, the standby file server in the second subnet automatically takes over if the preferred file server fails. Failover typically completes in under 30 seconds. DNS aliases and the `remote_administration_endpoint` follow the active server.

**DNS aliases:** Create CNAME records pointing `finance.corp.example.com` and `shared.corp.example.com` to the file system's `dns_name` output.

---

## 4. HDD Storage for Capacity-Optimized Workloads

Cost-effective HDD-based file system for workloads where capacity and sequential throughput matter more than latency: file archival, bulk data migration, or large shared drives.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: archive-windows-fs
  org: my-org
  labels:
    environment: production
    workload: file-archive
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 2000
  storage_type: HDD
  throughput_capacity: 64
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  active_directory_id:
    value: d-0123456789
  automatic_backup_retention_days: 14
  daily_automatic_backup_start_time: "02:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "1:04:00"
```

**HDD constraints:** HDD storage requires a minimum of 2000 GiB and is only supported on SINGLE_AZ_2 and MULTI_AZ_1 deployment types. SINGLE_AZ_1 does not support HDD.

**Cost savings:** HDD storage is significantly cheaper per GiB than SSD, making it suitable for large datasets where sub-millisecond latency is not required.

---

## 5. Cross-Resource Wiring with valueFrom

Production deployment demonstrating `valueFrom` references to other OpenMCF resources. The file system references an AwsVpc for subnets, an AwsSecurityGroup for network access, an AwsKmsKey for encryption, and an AwsCloudwatchLogGroup for audit log destination.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: wired-windows-fs
  org: my-org
  labels:
    environment: production
spec:
  deployment_type: MULTI_AZ_1
  storage_capacity_gib: 500
  storage_type: SSD
  throughput_capacity: 256
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: corp-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: corp-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  preferred_subnet_id:
    valueFrom:
      kind: AwsVpc
      name: corp-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: smb-access-sg
        fieldPath: status.outputs.security_group_id
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-encryption-key
      fieldPath: status.outputs.key_arn
  self_managed_active_directory:
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.10"
      - "10.0.0.11"
    domain_join_service_account_secret_arn:
      value: arn:aws:secretsmanager:us-east-1:123456789012:secret:fsx-domain-join-creds
    file_system_administrators_group: FSx Admins
    organizational_unit_distinguished_name: "OU=FSx,DC=corp,DC=example,DC=com"
  aliases:
    - files.corp.example.com
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: SUCCESS_AND_FAILURE
    audit_log_destination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: fsx-audit-logs
        fieldPath: status.outputs.log_group_arn
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 50000
  automatic_backup_retention_days: 30
  daily_automatic_backup_start_time: "03:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Cross-resource wiring:** `valueFrom` resolves at deployment time. OpenMCF reads the referenced resource's outputs and injects the values into this resource's spec. This eliminates hardcoded IDs and enables fully declarative infrastructure graphs.

**Secrets Manager integration:** `domain_join_service_account_secret_arn` points to a Secrets Manager secret containing `{"username": "...", "password": "..."}`. This is the recommended approach for production — no credentials in the manifest.

---

## CLI Flows

Validate manifest:

```bash
openmcf validate --manifest ./fsx-windows.yaml
```

Get outputs after deployment:

```bash
openmcf pulumi stack output file_system_id --stack my-org/project/prod
openmcf pulumi stack output dns_name --stack my-org/project/prod
openmcf pulumi stack output remote_administration_endpoint --stack my-org/project/prod
```

Mount from Windows:

```cmd
net use Z: \\fs-0123456789abcdef0.corp.example.com\share
```

Mount from Linux (with cifs-utils):

```bash
sudo mount -t cifs //fs-0123456789abcdef0.corp.example.com/share /mnt/fsx -o username=user,password=pass,domain=CORP
```

PowerShell remote administration:

```powershell
Enter-PSSession -ComputerName <remote_administration_endpoint> -ConfigurationName FsxRemoteAdmin
```

For more architecture details and integration patterns, see [docs/README.md](docs/README.md).

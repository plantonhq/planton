# AwsFsxOntapFileSystem Examples

Apply manifests with OpenMCF:

```shell
openmcf pulumi preview --manifest <yaml-path> --stack <stack-name>
openmcf pulumi update --manifest <yaml-path> --stack <stack-name> --yes
```

or

```shell
openmcf tofu init --manifest <yaml-path>
openmcf tofu plan --manifest <yaml-path>
openmcf tofu apply --manifest <yaml-path> --auto-approve
```

Provider credentials (AWS access key, secret, region) are supplied via stack input, not in the spec.

---

## 1. Minimal Single-AZ Development

The simplest configuration: SINGLE_AZ_2 with 1 TiB SSD and 128 MB/s throughput. One HA pair. No backups. Ideal for development, testing, or proof-of-concept.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: ontap-fs-dev
  id: awsfxo-ontap-fs-dev
  org: my-org
  env: dev
spec:
  region: us-east-1
  storage_capacity_gib: 1024
  throughput_capacity_per_ha_pair: 128
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
```

**Outputs:** `file_system_id`, `dns_name`, `management_dns_name`. After provisioning, create an SVM and volumes to expose NFS/SMB/iSCSI shares.

**Note:** Default `automatic_backup_retention_days: 0` disables backups. ONTAP snapshots provide point-in-time recovery independently.

---

## 2. Production Single-AZ with Backups and Encryption

Production-grade single-AZ file system with customer-managed KMS encryption, 7-day automatic backups, and security groups.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: prod-ontap-fs
  id: awsfxo-prod-ontap-fs
  org: my-org
  env: production
  labels:
    environment: production
spec:
  region: us-east-1
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 2048
  storage_type: SSD
  throughput_capacity_per_ha_pair: 512
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/your-kms-key-id
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Backups:** Daily at 05:00 UTC. 7-day retention. ONTAP snapshots can be configured separately on SVMs/volumes.

---

## 3. Scale-Out with Multiple HA Pairs

SINGLE_AZ_2 with 4 HA pairs. Each HA pair provides 512 MB/s; total throughput = 4 × 512 = 2048 MB/s. Suitable for high-throughput workloads (media processing, large databases, data lakes).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: scaleout-ontap-fs
  id: awsfxo-scaleout-ontap-fs
  org: my-org
  env: production
  labels:
    workload: high-throughput
spec:
  region: us-east-1
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 8192
  storage_type: SSD
  throughput_capacity_per_ha_pair: 512
  ha_pairs: 4
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/your-kms-key-id
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  weekly_maintenance_start_time: "7:02:00"
```

**Throughput:** 4 HA pairs × 512 MB/s = 2048 MB/s aggregate. SINGLE_AZ_2 allows increasing `ha_pairs` without replacement.

---

## 4. Multi-AZ High Availability

MULTI_AZ_2 deployment with automatic failover across two availability zones. Requires two subnets in different AZs and a preferred subnet. Endpoint IP address range enables seamless failover.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: ha-ontap-fs
  id: awsfxo-ha-ontap-fs
  org: my-org
  env: production
  labels:
    tier: critical
spec:
  region: us-east-1
  deployment_type: MULTI_AZ_2
  storage_capacity_gib: 2048
  storage_type: SSD
  throughput_capacity_per_ha_pair: 512
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0b2c3d4e5f600002
  preferred_subnet_id:
    value: subnet-0a1b2c3d4e5f00001
  endpoint_ip_address_range: 10.0.100.0/24
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-east-1:123456789012:key/your-kms-key-id
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Failover:** In a failover event, the standby file server in the other subnet takes over automatically. DNS and floating IPs follow the active server.

**Note:** `endpoint_ip_address_range` must be a CIDR within the VPC that does not overlap with existing subnets. AWS assigns floating IPs from this range.

---

## 5. Cross-Resource References with valueFrom

Production deployment using `valueFrom` to reference other OpenMCF resources. Eliminates hardcoded IDs and enables declarative dependency graphs.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: wired-ontap-fs
  id: awsfxo-wired-ontap-fs
  org: my-org
  env: production
spec:
  region: us-east-1
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 2048
  storage_type: SSD
  throughput_capacity_per_ha_pair: 512
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: corp-vpc
        fieldPath: status.outputs.private_subnets.[0].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ontap-access-sg
        fieldPath: status.outputs.security_group_id
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-encryption-key
      fieldPath: status.outputs.key_arn
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

**Cross-resource wiring:** `valueFrom` resolves at deployment time. OpenMCF reads the referenced resource's outputs and injects the values. Ensures correct ordering (VPC and security group must exist before the file system).

**Multi-AZ with valueFrom:**

```yaml
spec:
  region: us-east-1
  deployment_type: MULTI_AZ_2
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
  endpoint_ip_address_range: 10.0.100.0/24
```

---

## CLI Flows

Validate manifest:

```bash
openmcf validate --manifest ./ontap-fs.yaml
```

Get outputs after deployment:

```bash
openmcf pulumi stack output file_system_id --stack my-org/project/prod
openmcf pulumi stack output management_dns_name --stack my-org/project/prod
```

ONTAP CLI access (requires `fsx_admin_password` in spec):

```bash
ssh fsxadmin@<management_dns_name>
```

For architecture details and integration patterns, see [docs/README.md](docs/README.md).

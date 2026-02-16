# AwsElasticFileSystem Examples

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

## 1. Minimal EFS (Just Subnet IDs)

The simplest configuration: one subnet per AZ, no encryption override, default bursting throughput.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: minimal-efs
  org: my-org
spec:
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
```

**Outputs:** `file_system_id`, `dns_name`, `mount_target_ids`, `mount_target_dns_names`. Use `file_system_id` for EKS PersistentVolumes or ECS task definitions.

---

## 2. Encrypted with Elastic Throughput

Encryption at rest with a customer-managed KMS key, plus elastic throughput for unpredictable access patterns. Uses `valueFrom` for `kms_key_id` (AwsKmsKey), `subnet_ids` (AwsVpc), and `security_group_ids` (AwsSecurityGroup).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: encrypted-elastic-efs
  org: my-org
spec:
  encrypted: true
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: efs-encryption-key
      fieldPath: status.outputs.key_arn
  throughput_mode: elastic
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        fieldPath: status.outputs.security_group_id
```

**Note:** `elastic` throughput requires `generalPurpose` performance mode (default). Throughput scales automatically with workload.

---

## 3. One Zone for Development

Single-AZ storage (~47% cheaper than Standard). Suitable for dev/test where AZ-level failure is acceptable.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: dev-onezone-efs
  org: my-org
  labels:
    environment: development
spec:
  availability_zone_name: us-east-1a
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
```

**Constraint:** Exactly one subnet, and it must be in the specified AZ (`us-east-1a`).

---

## 4. With Lifecycle Policies (IA + Primary Storage Class)

Automatic tiering to Infrequent Access and Archive, with warm-back to Standard on access. Uses `valueFrom` for `subnet_ids` and `security_group_ids`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: lifecycle-efs
  org: my-org
spec:
  transition_to_ia: AFTER_30_DAYS
  transition_to_archive: AFTER_90_DAYS
  transition_to_primary_storage_class: AFTER_1_ACCESS
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        fieldPath: status.outputs.security_group_id
```

**Cost impact:** IA ~92% cheaper than Standard; Archive ~96% cheaper. Per-access fees apply for IA/Archive. `AFTER_1_ACCESS` moves files back to Standard when read.

---

## 5. With Access Point for ECS Task

Access point enforces POSIX identity and restricts root directory. ECS task definition references `file_system_id` and `access_point_id`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: ecs-shared-efs
  org: my-org
spec:
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ecs-efs-sg
        fieldPath: status.outputs.security_group_id
  access_points:
    - name: ecs-app-data
      posix_user:
        uid: 1000
        gid: 1000
      root_directory:
        path: /app-data
        creation_info:
          owner_uid: 1000
          owner_gid: 1000
          permissions: "0755"
```

**ECS usage:** In the task definition volume configuration, use `file_system_id` from `status.outputs.file_system_id` and `access_point_id` from `status.outputs.access_point_ids.ecs-app-data`.

---

## 6. With Access Point for Lambda (valueFrom for access_point_arn)

Lambda file system config requires the access point **ARN**, not ID. Reference via `valueFrom` from another resource that consumes this EFS.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: lambda-efs
  org: my-org
spec:
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  security_group_ids:
    - value: sg-0123456789abcdef0
  access_points:
    - name: lambda-data
      posix_user:
        uid: 1001
        gid: 1001
      root_directory:
        path: /lambda-data
        creation_info:
          owner_uid: 1001
          owner_gid: 1001
          permissions: "0755"
```

**Lambda usage:** In the Lambda function's file system config:

```yaml
# In AwsLambda spec (or equivalent)
fileSystemConfig:
  arn:
    valueFrom:
      kind: AwsElasticFileSystem
      name: lambda-efs
      fieldPath: status.outputs.access_point_arns.lambda-data
  localMountPath: /mnt/efs
```

---

## 7. Production-Ready (Encrypted, Elastic, Lifecycle, Backup, Multiple Access Points, Encryption-in-Transit Policy)

Full production configuration combining encryption, elastic throughput, lifecycle policies, backup, multiple access points, and a resource policy enforcing TLS.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: prod-efs
  org: my-org
  labels:
    environment: production
    app: shared-storage
spec:
  encrypted: true
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: prod-efs-key
      fieldPath: status.outputs.key_arn
  throughput_mode: elastic
  transition_to_ia: AFTER_30_DAYS
  transition_to_archive: AFTER_90_DAYS
  transition_to_primary_storage_class: AFTER_1_ACCESS
  backup_enabled: true
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        fieldPath: status.outputs.security_group_id
  access_points:
    - name: ecs-app-data
      posix_user:
        uid: 1000
        gid: 1000
      root_directory:
        path: /app-data
        creation_info:
          owner_uid: 1000
          owner_gid: 1000
          permissions: "0755"
    - name: lambda-models
      posix_user:
        uid: 1001
        gid: 1001
      root_directory:
        path: /ml-models
        creation_info:
          owner_uid: 1001
          owner_gid: 1001
          permissions: "0755"
    - name: ec2-batch
      posix_user:
        uid: 1002
        gid: 1002
      root_directory:
        path: /batch-jobs
        creation_info:
          owner_uid: 1002
          owner_gid: 1002
          permissions: "0750"
  policy:
    Version: "2012-10-17"
    Statement:
      - Sid: EnforceEncryptionInTransit
        Effect: Deny
        Principal: "*"
        Action: "*"
        Resource: "*"
        Condition:
          Bool:
            aws:SecureTransport: "false"
```

**Summary:**
- Encryption at rest with customer-managed KMS key
- Elastic throughput for variable workloads
- Lifecycle: IA after 30 days, Archive after 90 days, warm-back on access
- Daily backups via AWS Backup
- Three access points for ECS, Lambda, and EC2
- Resource policy denies unencrypted NFS (clients must use TLS mount helper or NFS-over-TLS)

---

## CLI Flows

Validate manifest:

```bash
openmcf validate --manifest ./efs.yaml
```

Get outputs after deployment:

```bash
openmcf pulumi stack output file_system_id --stack my-org/project/prod
openmcf pulumi stack output dns_name --stack my-org/project/prod
openmcf pulumi stack output access_point_arns --stack my-org/project/prod
```

Mount from EC2:

```bash
sudo yum install -y amazon-efs-utils
sudo mount -t efs -o tls fs-0123456789abcdef0:/ /mnt/efs
```

For more architecture details and integration patterns, see [docs/README.md](docs/README.md).

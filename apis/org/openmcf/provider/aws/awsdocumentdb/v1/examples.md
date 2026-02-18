# AwsDocumentDb Examples

## Minimal manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: docdb-dev
spec:
  region: us-east-1
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[1].id
  masterPassword: ${secrets-group/docdb/MASTER_PASSWORD}
  skipFinalSnapshot: true
```

## Production cluster with encryption and backups
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: docdb-prod
spec:
  region: us-east-1
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: docdb-sg
        fieldPath: status.outputs.security_group_id
  engineVersion: "5.0.0"
  masterUsername: prodadmin
  masterPassword: ${secrets-group/docdb/PROD_MASTER_PASSWORD}
  instanceCount: 3
  instanceClass: db.r6g.xlarge
  storageEncrypted: true
  kmsKey:
    valueFrom:
      kind: AwsKmsKey
      name: docdb-encryption-key
      fieldPath: status.outputs.key_arn
  backupRetentionPeriod: 35
  preferredBackupWindow: "03:00-04:00"
  preferredMaintenanceWindow: "sun:05:00-sun:06:00"
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: docdb-prod-final-snapshot
  enabledCloudwatchLogsExports:
    - audit
    - profiler
```

## High-availability cluster with existing subnet group
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: docdb-ha
spec:
  region: us-east-1
  dbSubnetGroup:
    value: existing-db-subnet-group
  securityGroups:
    - value: sg-0123456789abcdef0
  engineVersion: "5.0.0"
  masterUsername: haadmin
  masterPassword: ${secrets-group/docdb/HA_MASTER_PASSWORD}
  instanceCount: 3
  instanceClass: db.r6g.large
  storageEncrypted: true
  backupRetentionPeriod: 14
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: docdb-ha-final-snapshot
  autoMinorVersionUpgrade: true
```

## Development cluster with minimal resources
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDocumentDb
metadata:
  name: docdb-dev-minimal
spec:
  region: us-east-1
  subnets:
    - value: subnet-12345678
    - value: subnet-87654321
  allowedCidrs:
    - "10.0.0.0/16"
  masterPassword: ${secrets-group/docdb/DEV_PASSWORD}
  instanceCount: 1
  instanceClass: db.t3.medium
  storageEncrypted: true
  backupRetentionPeriod: 1
  skipFinalSnapshot: true
```

## CLI flows
- Validate: `openmcf validate --manifest examples/aws/awsdocumentdb/v1/minimal.yaml`
- Pulumi deploy: `openmcf pulumi update --manifest examples/aws/awsdocumentdb/v1/minimal.yaml --stack <org/project/stack> --module-dir apis/org/openmcf/provider/aws/awsdocumentdb/v1/iac/pulumi`
- Terraform deploy: `openmcf tofu apply --manifest examples/aws/awsdocumentdb/v1/minimal.yaml --auto-approve`

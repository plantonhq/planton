# AwsNeptuneCluster Examples

## Minimal manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNeptuneCluster
metadata:
  name: neptune-dev
spec:
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  skipFinalSnapshot: true
  storageEncrypted: true
```

## Production-ready cluster with encryption and backups
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNeptuneCluster
metadata:
  name: neptune-prod
spec:
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
    - value: subnet-0a1b2c3d4e5f00003
  securityGroupIds:
    - value: sg-0123456789abcdef0
  engineVersion: "1.3.0.0"
  port: 8182
  instanceCount: 3
  instanceClass: db.r6g.xlarge
  storageType: iopt1
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  backupRetentionPeriod: 14
  preferredBackupWindow: "03:00-04:00"
  preferredMaintenanceWindow: "sun:05:00-sun:06:00"
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: neptune-prod-final-snapshot
  enabledCloudwatchLogsExports:
    - audit
    - slowquery
  copyTagsToSnapshot: true
```

## Neptune Serverless v2
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNeptuneCluster
metadata:
  name: neptune-serverless
spec:
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - value: sg-0123456789abcdef0
  instanceClass: db.serverless
  serverlessV2Scaling:
    minCapacity: 1.0
    maxCapacity: 16.0
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  backupRetentionPeriod: 7
  skipFinalSnapshot: true
```

## Using foreign-key references
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNeptuneCluster
metadata:
  name: neptune-app
spec:
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: neptune-sg
        fieldPath: status.outputs.security_group_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: network
      fieldPath: status.outputs.vpc_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: neptune-encryption-key
      fieldPath: status.outputs.key_arn
  iamRoles:
    - valueFrom:
        kind: AwsIamRole
        name: neptune-s3-loader
        fieldPath: status.outputs.role_arn
  instanceCount: 2
  instanceClass: db.r6g.large
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  backupRetentionPeriod: 7
  deletionProtection: true
  skipFinalSnapshot: false
  finalSnapshotIdentifier: neptune-app-final-snapshot
  enabledCloudwatchLogsExports:
    - audit
    - slowquery
```

## CLI flows
- Validate: `openmcf validate --manifest examples/aws/awsneptunecluster/v1/minimal.yaml`
- Pulumi deploy: `openmcf pulumi update --manifest examples/aws/awsneptunecluster/v1/minimal.yaml --stack <org/project/stack> --module-dir apis/org/openmcf/provider/aws/awsneptunecluster/v1/iac/pulumi`
- Terraform deploy: `openmcf tofu apply --manifest examples/aws/awsneptunecluster/v1/minimal.yaml --auto-approve`

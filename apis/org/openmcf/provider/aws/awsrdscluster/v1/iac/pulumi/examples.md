## Minimal manifest (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: aurora-mysql-cluster
spec:
  region: us-west-2
  # provide either 2+ subnet_ids or a db_subnet_group_name
  subnetIds:
    - subnet-aaaa
    - subnet-bbbb
  engine: aurora-mysql
  engineVersion: 8.0.mysql_aurora.3.05.2
  manageMasterUserPassword: true
  preferredMaintenanceWindow: mon:00:00-tue:01:00
  preferredBackupWindow: 00:00-01:00
  skipFinalSnapshot: true
```

## Aurora PostgreSQL with logs exports and KMS (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsCluster
metadata:
  name: aurora-pg-cluster
spec:
  region: us-west-2
  dbSubnetGroupName:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.db_subnet_group
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: db-ingress
        fieldPath: status.outputs.security_group_id
  engine: aurora-postgresql
  engineVersion: 14.6
  storageEncrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: aurora-key
      fieldPath: status.outputs.key_arn
  enabledCloudwatchLogsExports:
    - postgresql
    - upgrade
  preferredMaintenanceWindow: wed:02:00-wed:03:00
  backupRetentionPeriod: 7
  preferredBackupWindow: 02:00-03:00
  copyTagsToSnapshot: true
  skipFinalSnapshot: true
  dbClusterParameterGroupName: default.aurora-postgresql14
  parameters:
    - name: rds.force_ssl
      value: "1"
      applyMethod: immediate
```

## CLI

Preview:

```shell
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Update (apply):

```shell
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```



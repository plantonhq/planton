## Minimal manifest (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: example-db
spec:
  region: us-west-2
  # provide either 2+ subnet_ids or a db_subnet_group_name
  subnetIds:
    - value: subnet-aaaa
    - value: subnet-bbbb
  engine: postgres
  engineVersion: "14.10"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  username: appuser
  password: changeme
  port: 5432
  publiclyAccessible: false
  multiAz: false
```

## Using existing DB subnet group and security group (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: example-db
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
  engine: mysql
  engineVersion: "8.0.36"
  instanceClass: db.t3.small
  allocatedStorageGb: 50
  storageEncrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-key
      fieldPath: status.outputs.key_arn
  parameterGroupName: default.mysql8.0
  optionGroupName: default:mysql-8-0
  port: 3306
  publiclyAccessible: true
```

## CLI (OpenTofu)

```bash
openmcf tofu init --manifest ../hack/manifest.yaml
openmcf tofu plan --manifest ../hack/manifest.yaml
openmcf tofu apply --manifest ../hack/manifest.yaml --auto-approve
```

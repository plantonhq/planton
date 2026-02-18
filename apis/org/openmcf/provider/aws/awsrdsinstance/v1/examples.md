## Minimal manifest (YAML)
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: db-postgres-dev
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[*].id
    - valueFrom:
        kind: AwsVpc
        name: network
        fieldPath: status.outputs.private_subnets.[*].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: db-sg
        fieldPath: status.outputs.security_group_id
  engine: postgres
  engineVersion: "14.10"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  username: master
  password: ${secrets-group/db/MASTER_PASSWORD}
  port: 5432
```

## Publicly-accessible MySQL with existing subnet group
```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: db-mysql-public
spec:
  region: us-west-2
  dbSubnetGroupName:
    value: app-db-subnet-group
  securityGroupIds:
    - value: sg-0123456789abcdef0
  engine: mysql
  engineVersion: "8.0.35"
  instanceClass: db.m6g.large
  allocatedStorageGb: 100
  username: admin
  password: ${secrets-group/db/ADMIN_PASSWORD}
  port: 3306
  publiclyAccessible: true
```

## CLI flows
- Validate: `openmcf validate --manifest examples/aws/awsrdsinstance/v1/minimal.yaml`
- Pulumi deploy: `openmcf pulumi update --manifest examples/aws/awsrdsinstance/v1/minimal.yaml --stack <org/project/stack> --module-dir apis/project/planton/provider/aws/awsrdsinstance/v1/iac/pulumi`
- Terraform deploy: `openmcf tofu apply --manifest examples/aws/awsrdsinstance/v1/minimal.yaml --auto-approve`

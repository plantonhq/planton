# AWS RDS Instance

Deploys a single AWS RDS database instance supporting engines such as PostgreSQL, MySQL, MariaDB, Oracle, and SQL Server. The component handles DB subnet group creation, security group attachment, optional storage encryption via KMS, and Multi-AZ deployment. Either subnet IDs or an existing DB subnet group name must be provided for VPC placement.

## What Gets Created

When you deploy an AwsRdsInstance resource, Planton provisions:

- **DB Subnet Group** — created only when `subnetIds` are provided and `dbSubnetGroupName` is not set; groups the specified subnets for RDS networking
- **RDS DB Instance** — an `aws:rds:Instance` with the configured engine, version, instance class, storage, and networking settings, placed in the specified subnets with attached security groups

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **At least two private subnets** in different Availability Zones, or an existing DB subnet group name
- **Security groups** allowing inbound traffic on the database port (e.g., 5432 for PostgreSQL, 3306 for MySQL)
- **A KMS key ARN** if enabling customer-managed storage encryption
- **Master credentials** (username and password) for the database root user

## Quick Start

Create a file `rds-instance.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: my-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsRdsInstance.my-db
spec:
  region: us-west-2
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  engine: postgres
  engineVersion: "14.10"
  instanceClass: db.t3.micro
  allocatedStorageGb: 20
  username: dbadmin
  password: changeme123
```

Deploy:

```shell
planton apply -f rds-instance.yaml
```

This creates a single PostgreSQL 14.10 instance on a `db.t3.micro` with 20 GiB of storage, placed in two private subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the RDS instance will be created. Example: `us-west-2`, `eu-west-1`. | |
| `subnetIds` | `string[]` | Subnet IDs for the DB subnet group. Provide at least two private subnets for high availability. Required unless `dbSubnetGroupName` is set. | Minimum 2 items when used |
| `subnetIds[].value` | `string` | Direct subnet ID value | — |
| `subnetIds[].valueFrom` | `object` | Foreign key reference to an AwsSubnet resource | Default kind: `AwsSubnet`, field: `status.outputs.subnet_id` |
| `dbSubnetGroupName` | `string` | Name of an existing DB subnet group. Required unless `subnetIds` (>=2) is provided. Can reference another resource via `valueFrom`. | — |
| `engine` | `string` | Database engine identifier (e.g., `"postgres"`, `"mysql"`, `"mariadb"`, `"oracle-se2"`, `"sqlserver-ex"`). | Minimum length 1 |
| `engineVersion` | `string` | Engine version string (e.g., `"14.10"` for PostgreSQL, `"8.0.35"` for MySQL). | Minimum length 1 |
| `instanceClass` | `string` | DB instance class (e.g., `"db.t3.micro"`, `"db.m6g.large"`, `"db.r6g.xlarge"`). | Must start with `db.` |
| `allocatedStorageGb` | `int32` | Allocated storage size in GiB. | Must be greater than 0 |
| `username` | `string` | Master username for the database. | Minimum length 1 |
| `password` | `string` | Master password for the database. | Minimum length 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `securityGroupIds` | `string[]` | `[]` | Security group IDs to associate with the instance's network interface. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `storageEncrypted` | `bool` | `false` | Enables encryption at rest for the DB instance storage. |
| `kmsKeyId` | `string` | — | Customer-managed KMS key ARN or alias for storage encryption. Only used when `storageEncrypted` is `true`. If not set, the default AWS-managed RDS key is used. Can reference an AwsKmsKey resource via `valueFrom`. |
| `port` | `int32` | Engine default | Database port number. If not set, the engine default is used (e.g., 5432 for PostgreSQL, 3306 for MySQL). | 
| `publiclyAccessible` | `bool` | `false` | When `true`, the instance is assigned a public IP and is reachable from outside the VPC. |
| `multiAz` | `bool` | `false` | When `true`, deploys a standby replica in a different Availability Zone for automatic failover. |
| `parameterGroupName` | `string` | — | Name of a DB parameter group to associate with the instance for engine-specific tuning. |
| `optionGroupName` | `string` | — | Name of an option group to associate with the instance (applicable to certain engines like Oracle and SQL Server). |

## Examples

### Encrypted PostgreSQL with Security Group

A PostgreSQL instance with storage encryption and a security group for controlled access:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: app-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.AwsRdsInstance.app-db
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-postgres-access
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.medium
  allocatedStorageGb: 50
  storageEncrypted: true
  username: appuser
  password: s3cur3P@ssw0rd
  port: 5432
```

### Multi-AZ MySQL with Existing Subnet Group

A MySQL instance using an existing DB subnet group and Multi-AZ deployment for high availability:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: ha-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsRdsInstance.ha-mysql
spec:
  region: us-west-2
  dbSubnetGroupName: existing-db-subnet-group
  securityGroupIds:
    - sg-mysql-access
  engine: mysql
  engineVersion: "8.0.35"
  instanceClass: db.m6g.large
  allocatedStorageGb: 100
  storageEncrypted: true
  username: mysqladmin
  password: pr0dP@ssw0rd
  port: 3306
  multiAz: true
  parameterGroupName: custom-mysql80-params
```

### Full-Featured Production PostgreSQL

Production configuration with KMS encryption, Multi-AZ, parameter group, and security groups:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: prod-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsRdsInstance.prod-db
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-prod-db
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.r6g.xlarge
  allocatedStorageGb: 500
  storageEncrypted: true
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd1234-5678-90ab-cdef-example11111
  username: prodadmin
  password: pr0dDBp@ss!
  port: 5432
  multiAz: true
  parameterGroupName: prod-postgres15-tuned
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: ref-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsRdsInstance.ref-db
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: db-sg
        field: status.outputs.security_group_id
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.m6g.large
  allocatedStorageGb: 100
  storageEncrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: db-encryption-key
      field: status.outputs.key_arn
  username: dbadmin
  password: s3cur3P@ssw0rd
  port: 5432
  multiAz: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `rds_instance_id` | `string` | RDS DB instance resource identifier |
| `rds_instance_arn` | `string` | Amazon Resource Name of the DB instance |
| `rds_instance_endpoint` | `string` | Hostname of the DB instance endpoint (e.g., `my-db.abc123.us-east-1.rds.amazonaws.com`) |
| `rds_instance_port` | `int32` | Port on which the DB instance accepts connections |
| `rds_subnet_group` | `string` | Name of the DB subnet group associated with the instance. Set when `subnetIds` are provided or `dbSubnetGroupName` is specified. |
| `rds_security_group` | `string` | First security group ID associated with the instance. Set only when `securityGroupIds` are provided. |
| `rds_parameter_group` | `string` | Name of the parameter group associated with the instance. Set only when `parameterGroupName` is specified. |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the subnets for DB instance placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls inbound and outbound traffic to the DB instance
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides the customer-managed key for storage encryption

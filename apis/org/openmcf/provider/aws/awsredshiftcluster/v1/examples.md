# AWS Redshift Cluster Examples

## 1. Minimal — Single-Node Dev Cluster

A single `dc2.large` node with AWS-managed password rotation. Suitable for
development and experimentation. Final snapshot is skipped for easy teardown.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: dev-analytics
spec:
  nodeType: dc2.large
  numberOfNodes: 1
  databaseName: dev
  masterUsername: admin
  manageMasterPassword: true
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  encrypted: true
  skipFinalSnapshot: true
```

**What this creates:**
- 1 × `dc2.large` single-node cluster (leader + compute combined)
- Redshift subnet group spanning two subnets
- AWS-managed master password in Secrets Manager
- Encryption at rest with the default Redshift service key
- No final snapshot on deletion

## 2. Production — Multi-Node RA3 with Full Security

A two-node `ra3.xlplus` cluster with managed storage, customer-managed KMS
encryption, CloudWatch audit logging, enhanced VPC routing, and a defined
maintenance window.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedshiftCluster
metadata:
  name: prod-warehouse
spec:
  nodeType: ra3.xlplus
  numberOfNodes: 2
  databaseName: warehouse
  masterUsername: warehouse_admin
  manageMasterPassword: true
  masterPasswordSecretKmsKeyId:
    value: "<kms-key-arn-for-secrets-manager>"
  port: 5439

  # Networking
  subnetIds:
    - value: "<private-subnet-id-az1>"
    - value: "<private-subnet-id-az2>"
  vpcId:
    value: "<vpc-id>"
  securityGroupIds:
    - value: "<app-layer-sg-id>"
  enhancedVpcRouting: true

  # Encryption
  encrypted: true
  kmsKeyId:
    value: "<kms-key-arn-for-cluster>"

  # IAM
  iamRoles:
    - value: "<redshift-s3-role-arn>"
    - value: "<redshift-glue-role-arn>"
  defaultIamRoleArn:
    value: "<redshift-s3-role-arn>"

  # Snapshots
  automatedSnapshotRetentionPeriod: 7
  skipFinalSnapshot: false
  finalSnapshotIdentifier: prod-warehouse-final

  # Maintenance
  preferredMaintenanceWindow: "sat:03:00-sat:04:00"
  allowVersionUpgrade: true

  # Logging — CloudWatch
  logging:
    logDestinationType: cloudwatch
    logExports:
      - connectionlog
      - useractivitylog
      - userlog

  # Parameter Group — enforce SSL
  parameters:
    - name: require_ssl
      value: "true"
    - name: enable_user_activity_logging
      value: "true"
```

**What this creates:**
- 2 × `ra3.xlplus` nodes (1 leader + 2 compute, managed storage backed by S3)
- Redshift subnet group across two AZs
- Managed security group allowing traffic from the application SG on port 5439
- Customer-managed KMS encryption for cluster data and Secrets Manager secret
- Two IAM roles attached for S3/Glue access
- 7-day automated snapshot retention + final snapshot on deletion
- Weekly Saturday 03:00–04:00 UTC maintenance window
- Audit logs (connections, queries, DDL) streamed to CloudWatch Logs
- Custom parameter group enforcing SSL and user activity logging

## 3. CLI Workflows

### Validate a Manifest

```bash
openmcf validate -f manifest.yaml
```

### Deploy with Pulumi

```bash
# Preview changes
openmcf pulumi preview -f manifest.yaml --stack dev

# Apply changes
openmcf pulumi up -f manifest.yaml --stack dev
```

### Deploy with Terraform

```bash
# Initialize and plan
openmcf terraform init -f manifest.yaml
openmcf terraform plan -f manifest.yaml

# Apply
openmcf terraform apply -f manifest.yaml
```

### Destroy

```bash
# Pulumi
openmcf pulumi destroy -f manifest.yaml --stack dev

# Terraform
openmcf terraform destroy -f manifest.yaml
```

# AWS MWAA Environment

Deploys a managed Apache Airflow environment on Amazon MWAA with auto-scaling workers and webservers, per-module CloudWatch logging, customer-managed KMS encryption, and VPC-scoped networking. DAGs, plugins, and Python requirements are sourced from an S3 bucket. The source bucket, execution role, subnets, security groups, and KMS key all accept ValueFromRef wiring, so the environment composes with its network and storage dependencies in one InfraChart. Subnets are create-time only — changing them replaces the environment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MWAA Environment** -- a managed Airflow environment running the specified version with configurable environment class, worker auto-scaling, scheduler count, and webserver access mode, placed in two private subnets across different Availability Zones with the referenced security groups attached to its VPC endpoints
- **CloudWatch Log Groups** -- created automatically by MWAA when logging modules are enabled; one log group per module (DAG processing, scheduler, task, webserver, worker)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An S3 bucket** with versioning enabled containing DAG files, optional plugins.zip, and optional requirements.txt. The execution role must have read access to this bucket. Provide the bucket ARN directly or reference an AwsS3Bucket Infra Component via ValueFromRef.
- **An IAM execution role** with permissions for S3 (DAGs bucket), CloudWatch Logs, SQS (Celery backend), and any AWS services your DAGs interact with. Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **Exactly two private subnets** in distinct Availability Zones with no direct route to an internet gateway. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef. Subnets are create-time only -- changing them replaces the environment.
- **At least one security group** to attach to the MWAA VPC endpoints. Network ingress is composed, never embedded: the referenced AwsSecurityGroup must carry a self-referencing all-traffic ingress rule (MWAA's components communicate with each other through it), HTTPS (443) ingress from whatever should reach the Airflow UI, and outbound egress. Author those rules on the security group resource, where they stay shareable and auditable. Unlike subnets, the attached groups can be changed in place.
- **A KMS key** (optional) for encrypting environment data at rest. If omitted, MWAA uses the default `aws/airflow` service key.

## Deploy

### Console

Open the deployment store, find **AWS MWAA Environment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Private Airflow Environment** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMwaaEnvironment
metadata:
  name: data-pipelines
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  sourceBucketArn:
    value: "arn:aws:s3:::my-airflow-bucket"
  dagS3Path: "dags/"
  executionRoleArn:
    value: "arn:aws:iam::123456789012:role/mwaa-execution-role"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
  environmentClass: mw1.small
```

```shell
planton apply -f mwaa-environment.yaml
```

This creates a private-access Airflow environment with mw1.small capacity, auto-scaling workers (1-10), and the default AWS-managed encryption key. No CloudWatch logging is enabled. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the MWAA environment to an S3 bucket, IAM role, VPC, security group, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  sourceBucketArn:
    valueFrom:
      kind: AwsS3Bucket
      name: airflow-dags
      fieldPath: status.outputs.bucket_arn
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: mwaa-role
      fieldPath: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-az2
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: mwaa-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: airflow-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the S3 bucket, IAM role, subnets, security group, and KMS key first, then provisions the MWAA environment with the resolved values.

## Key Configuration

These are the most important decisions when configuring an MWAA environment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Environment class** -- Determines CPU and memory per Airflow component. `mw1.small` (1 vCPU, 2 GB) suits small workloads; `mw1.medium` (2 vCPU, 4 GB) suits production; `mw1.large` and above handle high-throughput DAG repositories. Class changes are applied in-place.

**Worker scaling** -- Set `minWorkers` and `maxWorkers` to control Celery worker auto-scaling based on task queue depth. Start with 1-10 for most workloads. Each additional worker incurs per-hour compute cost proportional to the environment class.

**Webserver access** -- Set `webserverAccessMode` to `PRIVATE_ONLY` (default) for VPC-only access via VPC endpoint, `PUBLIC_ONLY` for internet-accessible UI with IAM-based login, or `PUBLIC_AND_PRIVATE` for both. Private access is recommended for production and compliance environments.

**Logging** -- Enable individual logging modules (`loggingConfiguration`) for DAG processing, scheduler, task, webserver, and worker. Each module can have its own log level. Enable at least task and scheduler logs for production debugging.

**Encryption** -- Provide `kmsKeyArn` for a customer-managed key encrypting the metadata database, DAG logs, SQS queue, and web server logs. Changing the KMS key forces environment replacement.

**Deterministic deployments** -- Pin `pluginsS3ObjectVersion`, `requirementsS3ObjectVersion`, and `startupScriptS3ObjectVersion` to specific S3 object versions so every environment update installs exactly the artifacts you tested. Unpinned paths use the latest object version.

**Endpoint management** -- `endpointManagement` defaults to AWS-managed (`SERVICE`) VPC endpoints. `CUSTOMER` mode hands endpoint creation to you: the environment waits in CREATING until VPC endpoints exist against the `database_vpc_endpoint_service` and `webserver_vpc_endpoint_service` output values. Fixed at creation.

**Worker replacement** -- `workerReplacementStrategy` decides how environment updates treat running tasks: `GRACEFUL` waits for each worker's tasks to finish; `FORCED` replaces immediately and may interrupt them.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `sourceBucketArn` | `status.outputs.bucket_arn` |
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `environment_arn` | Amazon Resource Name | IAM policies, CloudWatch alarms, resource tagging |
| `environment_name` | Environment name | CLI commands, monitoring dashboards |
| `webserver_url` | Airflow web UI URL | Team access, bookmarks, CI/CD DAG validation |
| `airflow_version` | Running Airflow version | Compatibility verification |
| `service_role_arn` | AWS service role ARN | Audit, IAM policy references |
| `environment_class` | Effective environment class | Monitoring, capacity planning |
| `status` | Environment status | Health checks, deployment automation |
| `created_at` | Environment creation timestamp | Audit, lifecycle tracking |
| `database_vpc_endpoint_service` | Metadata-database endpoint service name | AwsVpcEndpoint targets under customer-managed endpoints |
| `webserver_vpc_endpoint_service` | Webserver endpoint service name | AwsVpcEndpoint targets under customer-managed endpoints (empty for public webservers) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic private Airflow** -- Private-access mw1.small environment with auto-scaling workers and default encryption. Suited for development and small teams getting started with managed Airflow. Start from the **Basic Private Airflow Environment** preset.

**Production encrypted with logging** -- mw1.medium environment with customer-managed KMS encryption, all five logging modules enabled, graceful worker replacement, and a defined maintenance window. Suited for production data pipeline orchestration. Start from the **Production Encrypted Airflow with Full Logging** preset.

**Public access with plugins** -- mw1.large environment with public webserver access, custom plugins.zip, Python requirements, a startup script, and aggressive worker scaling up to 25. Demonstrates the full breadth of MWAA extensibility. Start from the **Public Access with Plugins and Custom Packages** preset.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the source bucket for DAG files, plugins, and requirements
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the execution role for accessing AWS services from DAGs
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the two private subnets for MWAA network interfaces
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- carries the self-referencing ingress and HTTPS rules attached to the MWAA VPC endpoints
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for environment data encryption
- [**AWS VPC Endpoint**](/infra-catalog/aws-vpc-endpoint) -- consumes the endpoint-service outputs under customer-managed endpoint management
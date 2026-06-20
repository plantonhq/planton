---
title: "MWAA Environment"
description: "MWAA Environment deployment documentation"
icon: "package"
order: 100
componentName: "awsmwaaenvironment"
---

# AWS MWAA Environment

Deploys an Amazon Managed Workflows for Apache Airflow environment with DAGs sourced from S3, VPC-based networking across two Availability Zones, and optional managed security group creation. The component handles environment sizing, per-module CloudWatch logging, KMS encryption, and worker auto-scaling configuration.

## What Gets Created

When you deploy an AwsMwaaEnvironment resource, OpenMCF provisions:

- **MWAA Environment** — an `aws_mwaa_environment` with DAGs loaded from an S3 bucket, an execution role for AWS service access, and VPC endpoints in private subnets across two Availability Zones
- **Managed Security Group** — created only when `vpcId` is provided together with `securityGroupIds` or `allowedCidrBlocks`. Includes a self-referencing inbound rule (all traffic) for MWAA component intercommunication, HTTPS (port 443) ingress from each specified source security group and CIDR block, and full egress
- **Security Group Rules** — one ingress rule per source security group and one per CIDR block, all on port 443

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An S3 bucket** with versioning enabled, containing your DAG files (and optionally plugins.zip, requirements.txt, startup script)
- **An IAM execution role** with permissions for S3, CloudWatch Logs, SQS, and any AWS services your DAGs interact with
- **Two private subnets** in different Availability Zones (no direct route to an internet gateway)
- **A VPC ID** if using managed security group creation, or existing security groups via `associateSecurityGroupIds`
- **A KMS key ARN** if enabling customer-managed encryption

## Quick Start

Create a file `mwaa.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: my-airflow
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMwaaEnvironment.my-airflow
spec:
  region: us-west-2
  sourceBucketArn:
    value: arn:aws:s3:::my-airflow-dags
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/mwaa-execution-role
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  associateSecurityGroupIds:
    - value: sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f mwaa.yaml
```

This creates a private Airflow environment with the default `mw1.small` instance class, DAGs loaded from S3, and an existing security group attached directly.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the MWAA environment will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `sourceBucketArn` | `StringValueOrRef` | ARN of the S3 bucket containing DAGs, plugins, and requirements. Bucket must have versioning enabled. | Required |
| `sourceBucketArn.value` | `string` | Direct S3 bucket ARN value | — |
| `sourceBucketArn.valueFrom` | `object` | Foreign key reference to an AwsS3Bucket resource | Default field: `status.outputs.bucket_arn` |
| `dagS3Path` | `string` | Relative path within the S3 bucket to the DAG files folder. Must not start with `/`. | Required, no leading slash |
| `executionRoleArn` | `StringValueOrRef` | ARN of the IAM role MWAA assumes for S3, CloudWatch Logs, SQS, and DAG service access. | Required |
| `executionRoleArn.value` | `string` | Direct IAM role ARN value | — |
| `executionRoleArn.valueFrom` | `object` | Foreign key reference to an AwsIamRole resource | Default field: `status.outputs.role_arn` |
| `subnetIds` | `StringValueOrRef[]` | Private subnets where MWAA creates network interfaces. Must be in different Availability Zones. Changing subnets forces replacement. | Minimum 2 items |
| `subnetIds[].value` | `string` | Direct subnet ID value | — |
| `subnetIds[].valueFrom` | `object` | Foreign key reference to an AwsSubnet resource | Default field: `status.outputs.subnet_id` |

At least one of `vpcId` (for managed SG) or `associateSecurityGroupIds` must be provided for VPC endpoint security.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `airflowVersion` | `string` | Latest supported | Apache Airflow version (e.g., `2.10.1`). Minor upgrades apply in-place; major changes force replacement. |
| `airflowConfigurationOptions` | `map<string, string>` | `{}` | Airflow configuration overrides in `section.property` format (e.g., `core.default_timezone`). May contain sensitive values. |
| `environmentClass` | `string` | `mw1.small` | Compute capacity. One of: `mw1.micro`, `mw1.small`, `mw1.medium`, `mw1.large`, `mw1.xlarge`, `mw1.2xlarge`. |
| `minWorkers` | `int` | `1` | Minimum Celery workers for auto-scaling. Must be >= 1. |
| `maxWorkers` | `int` | `10` | Maximum Celery workers for auto-scaling. Must be >= `minWorkers`. |
| `minWebservers` | `int` | `2` | Minimum webserver instances. Range: 1-5 (1 only for `mw1.micro`). |
| `maxWebservers` | `int` | `2` | Maximum webserver instances. Range: 1-5. Must be >= `minWebservers`. |
| `schedulers` | `int` | `2` | Number of Airflow schedulers. Range: 2-5. More schedulers improve DAG parsing throughput. |
| `webserverAccessMode` | `string` | `PRIVATE_ONLY` | `PRIVATE_ONLY`: VPC-only access. `PUBLIC_ONLY`: internet-accessible with IAM login. |
| `endpointManagement` | `string` | `SERVICE` | `SERVICE`: AWS manages VPC endpoints. `CUSTOMER`: you manage endpoints. Changing forces replacement. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Source security groups allowed to reach MWAA endpoints on port 443. Requires `vpcId`. Triggers managed SG creation. |
| `allowedCidrBlocks` | `string[]` | `[]` | IPv4 CIDR ranges allowed to reach MWAA endpoints on port 443. Requires `vpcId`. Triggers managed SG creation. Must be unique, valid CIDR notation. |
| `associateSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Existing security groups attached directly to the MWAA environment. Use when you manage your own SG with self-referencing rules. |
| `vpcId` | `StringValueOrRef` | — | VPC in which to create the managed security group. Required when `securityGroupIds` or `allowedCidrBlocks` are provided. |
| `kmsKeyArn` | `StringValueOrRef` | AWS-managed key | KMS key ARN for encrypting environment data at rest. Changing forces replacement. |
| `pluginsS3Path` | `string` | — | Relative path in the S3 bucket to a `plugins.zip` file containing custom operators, hooks, and sensors. |
| `pluginsS3ObjectVersion` | `string` | Latest | S3 object version for `plugins.zip`. Pins to a specific version for deterministic deployments. |
| `requirementsS3Path` | `string` | — | Relative path in the S3 bucket to a `requirements.txt` file listing additional Python packages. |
| `requirementsS3ObjectVersion` | `string` | Latest | S3 object version for `requirements.txt`. Pins to a specific version for deterministic deployments. |
| `startupScriptS3Path` | `string` | — | Relative path in the S3 bucket to a startup shell script for OS-level setup at environment boot. |
| `startupScriptS3ObjectVersion` | `string` | Latest | S3 object version for the startup script. Pins to a specific version for deterministic deployments. |
| `weeklyMaintenanceWindowStart` | `string` | AWS-selected | Preferred maintenance window in `DAY:HH:MM` UTC format (e.g., `TUE:03:30`). |
| `workerReplacementStrategy` | `string` | — | `FORCED`: replaces workers immediately (may interrupt tasks). `GRACEFUL`: waits for running tasks to complete. |
| `loggingConfiguration` | `object` | — | Per-module CloudWatch Logs configuration. See logging fields below. |
| `loggingConfiguration.dagProcessingLogs.enabled` | `bool` | `false` | Enable DAG processing logs to CloudWatch. |
| `loggingConfiguration.dagProcessingLogs.logLevel` | `string` | `INFO` | Log level: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |
| `loggingConfiguration.schedulerLogs.enabled` | `bool` | `false` | Enable scheduler logs to CloudWatch. |
| `loggingConfiguration.schedulerLogs.logLevel` | `string` | `INFO` | Log level: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |
| `loggingConfiguration.taskLogs.enabled` | `bool` | `false` | Enable task execution logs to CloudWatch. |
| `loggingConfiguration.taskLogs.logLevel` | `string` | `INFO` | Log level: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |
| `loggingConfiguration.webserverLogs.enabled` | `bool` | `false` | Enable webserver logs to CloudWatch. |
| `loggingConfiguration.webserverLogs.logLevel` | `string` | `INFO` | Log level: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |
| `loggingConfiguration.workerLogs.enabled` | `bool` | `false` | Enable worker logs to CloudWatch. |
| `loggingConfiguration.workerLogs.logLevel` | `string` | `INFO` | Log level: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |

## Examples

### Basic Private Airflow

A minimal environment using an existing security group attached directly. No managed SG is created:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: dev-airflow
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMwaaEnvironment.dev-airflow
spec:
  region: us-west-2
  sourceBucketArn:
    value: arn:aws:s3:::dev-airflow-bucket
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/mwaa-execution
  subnetIds:
    - value: subnet-private-az1
    - value: subnet-private-az2
  associateSecurityGroupIds:
    - value: sg-mwaa-existing
```

### Production with KMS and Logging

Encrypted environment with all five log modules enabled, a weekly maintenance window, and graceful worker replacement:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: prod-airflow
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMwaaEnvironment.prod-airflow
spec:
  region: us-east-1
  airflowVersion: "2.10.1"
  sourceBucketArn:
    value: arn:aws:s3:::prod-airflow-bucket
  dagS3Path: dags/
  pluginsS3Path: plugins/plugins.zip
  pluginsS3ObjectVersion: "3"
  requirementsS3Path: requirements/requirements.txt
  requirementsS3ObjectVersion: "7"
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/prod-mwaa-execution
  subnetIds:
    - value: subnet-prod-az1
    - value: subnet-prod-az2
  associateSecurityGroupIds:
    - value: sg-prod-mwaa
  kmsKeyArn:
    value: arn:aws:kms:us-east-1:123456789012:key/prod-mwaa-key
  environmentClass: mw1.large
  minWorkers: 2
  maxWorkers: 20
  minWebservers: 2
  maxWebservers: 4
  schedulers: 3
  weeklyMaintenanceWindowStart: "TUE:03:30"
  workerReplacementStrategy: GRACEFUL
  airflowConfigurationOptions:
    core.default_timezone: "UTC"
    webserver.dag_default_view: "grid"
  loggingConfiguration:
    dagProcessingLogs:
      enabled: true
      logLevel: WARNING
    schedulerLogs:
      enabled: true
      logLevel: INFO
    taskLogs:
      enabled: true
      logLevel: INFO
    webserverLogs:
      enabled: true
      logLevel: WARNING
    workerLogs:
      enabled: true
      logLevel: INFO
```

### Managed Security Group with VPC

Creates a managed security group with source SGs and CIDR-based ingress. Use this pattern when MWAA endpoints need to accept connections from multiple known sources:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: team-airflow
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsMwaaEnvironment.team-airflow
spec:
  region: us-west-2
  sourceBucketArn:
    valueFrom:
      kind: AwsS3Bucket
      name: airflow-dags-bucket
      field: status.outputs.bucket_arn
  dagS3Path: dags/
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: mwaa-execution-role
      field: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: main-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: main-private-subnet-b
        fieldPath: status.outputs.subnet_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: main-vpc
      field: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: bastion-sg
        field: status.outputs.security_group_id
    - valueFrom:
        kind: AwsSecurityGroup
        name: cicd-sg
        field: status.outputs.security_group_id
  allowedCidrBlocks:
    - "10.0.0.0/16"
    - "172.16.0.0/12"
  environmentClass: mw1.medium
  minWorkers: 1
  maxWorkers: 10
  webserverAccessMode: PRIVATE_ONLY
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: mwaa-key
      field: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `environment_arn` | `string` | ARN of the MWAA environment, used in IAM policies and cross-service references |
| `environment_name` | `string` | Name of the MWAA environment |
| `webserver_url` | `string` | Airflow UI URL in the format `{id}.{region}.airflow.amazonaws.com`. Access depends on `webserverAccessMode`. |
| `airflow_version` | `string` | Effective Apache Airflow version running in the environment |
| `service_role_arn` | `string` | ARN of the AWS service role MWAA created for managing infrastructure |
| `environment_class` | `string` | Effective environment class (compute capacity) |
| `status` | `string` | Current environment status (e.g., `AVAILABLE`, `CREATING`, `UPDATING`) |
| `security_group_id` | `string` | ID of the managed security group. Only populated when `securityGroupIds` or `allowedCidrBlocks` triggered managed SG creation. |

## Related Components

- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — hosts the DAG files, plugins, requirements, and startup scripts
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the execution role for MWAA service access
- [AwsVpc](/docs/catalog/aws/vpc) — provides private subnets and VPC ID for networking and managed security group creation
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls traffic to MWAA VPC endpoints, used as source SGs or directly associated
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides the encryption key for environment data at rest

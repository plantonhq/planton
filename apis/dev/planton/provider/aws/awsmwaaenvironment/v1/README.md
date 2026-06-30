# AwsMwaaEnvironment

Amazon MWAA (Managed Workflows for Apache Airflow) environment resource for Planton. Provisions a fully managed Apache Airflow environment on AWS with configurable compute sizing, auto-scaling, encryption, logging, networking, and S3-based DAG deployment. MWAA handles scheduler, worker, and webserver infrastructure so teams can focus on authoring, scheduling, and monitoring data pipelines using Airflow DAGs.

## When to use

- You need a managed Apache Airflow environment on AWS without operating scheduler, worker, or webserver infrastructure.
- Data pipeline orchestration: ETL/ELT workflows, data lake ingestion, warehouse loading, data quality checks.
- ML pipeline management: model training, feature engineering, batch inference, experiment tracking.
- Cross-service orchestration: coordinating Lambda, Glue, EMR, Redshift, Step Functions, SageMaker, and other AWS services.
- You want a Celery-based worker pool with auto-scaling that responds to DAG task queue depth.
- Your workflows require versioned DAG deployment via S3 with rollback capability.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| S3 bucket with versioning enabled | Stores DAG files, plugins.zip, requirements.txt, and startup scripts. Must have a bucket policy granting MWAA read access. | `AwsS3Bucket` |
| IAM execution role | MWAA assumes this role to access S3 (DAGs bucket), CloudWatch Logs, SQS (Celery backend), and any AWS services your DAGs call. | `AwsIamRole` |
| VPC with 2 private subnets in different AZs | MWAA creates VPC endpoints for the webserver, scheduler, and metadata database. Subnets must be private (no direct internet gateway route). | `AwsVpc` |
| Security groups or CIDR ranges (optional) | Control which clients can reach the MWAA VPC endpoints on port 443. When provided with `vpcId`, a managed security group is created. | `AwsSecurityGroup` |
| KMS key (optional) | Customer-managed encryption for metadata database, DAG logs, SQS queue, and webserver logs. Without it, AWS uses the default `aws/airflow` service key. | `AwsKmsKey` |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMwaaEnvironment
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

### Airflow Configuration

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `airflowVersion` | string | no | latest supported | Apache Airflow version (`2.10.1`, `2.9.2`, `2.8.1`, etc.). Minor upgrades are in-place; major changes **ForceNew**. |
| `airflowConfigurationOptions` | map(string, string) | no | {} | Airflow config overrides in `section.property` format (e.g., `core.default_timezone`, `celery.worker_autoscale`). Values may contain secrets. |

### S3 Source

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `sourceBucketArn` | StringValueOrRef | **yes** | — | ARN of the S3 bucket containing DAGs, plugins, and requirements. Bucket must have versioning enabled. Default ref: `AwsS3Bucket → status.outputs.bucket_arn`. |
| `dagS3Path` | string | **yes** | — | Relative path to the DAG folder in S3 (e.g., `dags/`). Must not start with `/`. |
| `pluginsS3Path` | string | no | — | Relative path to `plugins.zip` in S3 (e.g., `plugins/plugins.zip`). |
| `pluginsS3ObjectVersion` | string | no | latest | Pins plugins.zip to a specific S3 object version for deterministic deployments. |
| `requirementsS3Path` | string | no | — | Relative path to `requirements.txt` in S3 (e.g., `requirements/requirements.txt`). |
| `requirementsS3ObjectVersion` | string | no | latest | Pins requirements.txt to a specific S3 object version. |
| `startupScriptS3Path` | string | no | — | Relative path to a startup shell script (e.g., `scripts/startup.sh`). For OS-level setup: system packages, env vars, auth config. Airflow 2.x+. |
| `startupScriptS3ObjectVersion` | string | no | latest | Pins the startup script to a specific S3 object version. |

### IAM

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `executionRoleArn` | StringValueOrRef | **yes** | — | ARN of the IAM role MWAA assumes. Needs S3, CloudWatch Logs, SQS, and any DAG-referenced service permissions. Default ref: `AwsIamRole → status.outputs.role_arn`. |

### VPC Networking

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `subnetIds` | list(StringValueOrRef) | **yes** (≥2) | — | Private subnets in 2 different AZs for MWAA network interfaces. **ForceNew**. Default ref: `AwsSubnet → status.outputs.subnet_id`. |
| `securityGroupIds` | list(StringValueOrRef) | no | [] | Source security groups. When set (with `vpcId`), creates a managed SG with self-referencing rule + HTTPS (443) ingress. Default ref: `AwsSecurityGroup → status.outputs.security_group_id`. |
| `allowedCidrBlocks` | list(string) | no | [] | IPv4 CIDRs allowed to reach MWAA endpoints on port 443. Same managed-SG behavior as `securityGroupIds`. Validated CIDR format. |
| `associateSecurityGroupIds` | list(StringValueOrRef) | no | [] | Existing SGs attached directly alongside the managed SG. Use when you already have a self-referencing SG. Default ref: `AwsSecurityGroup → status.outputs.security_group_id`. |
| `vpcId` | StringValueOrRef | conditional | — | VPC for the managed SG. Required when `securityGroupIds` or `allowedCidrBlocks` are provided. Default ref: `AwsVpc → status.outputs.vpc_id`. |

### Encryption

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `kmsKeyArn` | StringValueOrRef | no | `aws/airflow` service key | KMS key for encrypting metadata DB, DAG logs, SQS queue, and webserver logs. **ForceNew**. Default ref: `AwsKmsKey → status.outputs.key_arn`. |

### Environment Sizing

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `environmentClass` | string | no | `mw1.small` | Compute capacity: `mw1.micro` (0.5 vCPU, 1 GB), `mw1.small` (1 vCPU, 2 GB), `mw1.medium` (2 vCPU, 4 GB), `mw1.large` (4 vCPU, 8 GB), `mw1.xlarge` (8 vCPU, 16 GB), `mw1.2xlarge` (16 vCPU, 32 GB). |
| `minWorkers` | int32 | no | 1 | Minimum Celery workers for auto-scaling. ≥ 1. |
| `maxWorkers` | int32 | no | 10 | Maximum Celery workers for auto-scaling. ≥ 1. Must be ≥ `minWorkers`. |
| `minWebservers` | int32 | no | 2 | Minimum Airflow webservers. Range: 2–5 (1 for `mw1.micro`). |
| `maxWebservers` | int32 | no | 2 | Maximum Airflow webservers. Range: 2–5 (1 for `mw1.micro`). Must be ≥ `minWebservers`. |
| `schedulers` | int32 | no | 2 | Number of Airflow schedulers. Range: 2–5. More schedulers improve parsing and scheduling throughput. |

### Access & Networking

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `webserverAccessMode` | string | no | `PRIVATE_ONLY` | `PRIVATE_ONLY`: accessible only within VPC via VPC endpoint. `PUBLIC_ONLY`: internet-accessible with IAM login. |
| `endpointManagement` | string | no | `SERVICE` | `SERVICE`: AWS manages VPC endpoints. `CUSTOMER`: you manage endpoints yourself (advanced). **ForceNew**. |

### Logging

Nested message `loggingConfiguration` with 5 log modules:

| Module | Field | Description |
|---|---|---|
| DAG processing | `dagProcessingLogs` | Logs from the DAG parser (determines scheduling requirements). |
| Scheduler | `schedulerLogs` | Logs from the scheduler (triggers task instances). |
| Task | `taskLogs` | Stdout/stderr from individual DAG task runs. |
| Webserver | `webserverLogs` | Logs from the Airflow web UI and REST API. |
| Worker | `workerLogs` | Logs from Celery workers executing task code. |

Each module (`AwsMwaaEnvironmentLoggingModuleConfig`) has:

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | false | Whether logs for this module are delivered to CloudWatch Logs. |
| `logLevel` | string | `INFO` | Minimum severity: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. |

CloudWatch log groups are auto-created by MWAA: `/aws/mwaa/{environment-name}/{module-name}`.

### Maintenance

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `weeklyMaintenanceWindowStart` | string | no | AWS-selected | Preferred maintenance window in `DAY:HH:MM` UTC format (e.g., `TUE:03:30`, `SUN:00:00`). |

### Operations

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `workerReplacementStrategy` | string | no | — | `FORCED`: replaces workers immediately (faster, may interrupt tasks). `GRACEFUL`: waits for running tasks to complete (slower, no data loss). |

## Output fields reference

| Output | Type | Description |
|---|---|---|
| `environment_arn` | string | ARN of the MWAA environment. Used in IAM policies and cross-service references. |
| `environment_name` | string | Human-readable environment name. |
| `webserver_url` | string | Airflow web UI URL (`{id}.{region}.airflow.amazonaws.com`). Access depends on `webserverAccessMode`. |
| `airflow_version` | string | Effective Apache Airflow version running in the environment. |
| `service_role_arn` | string | ARN of the AWS service role created by MWAA for managing infrastructure. |
| `environment_class` | string | Effective environment class (compute capacity). |
| `status` | string | Current environment status: `AVAILABLE`, `CREATING`, `UPDATING`, `DELETING`, etc. |
| `security_group_id` | string | ID of the managed security group (if created from `securityGroupIds` or `allowedCidrBlocks`). |

## Examples

### Minimal private Airflow environment

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMwaaEnvironment
metadata:
  name: dev-airflow
spec:
  sourceBucketArn:
    value: arn:aws:s3:::dev-airflow-dags
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::111122223333:role/mwaa-execution-role
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  associateSecurityGroupIds:
    - value: sg-0mwaa001
```

### Production-ready with KMS + logging

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMwaaEnvironment
metadata:
  name: prod-data-pipelines
spec:
  airflowVersion: "2.10.1"
  sourceBucketArn:
    valueFrom:
      kind: AwsS3Bucket
      name: prod-airflow-bucket
      fieldPath: status.outputs.bucket_arn
  dagS3Path: dags/
  pluginsS3Path: plugins/plugins.zip
  requirementsS3Path: requirements/requirements.txt
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: prod-mwaa-execution-role
      fieldPath: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: production-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: production-private-subnet-b
        fieldPath: status.outputs.subnet_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: data-team-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: platform-encryption-key
      fieldPath: status.outputs.key_arn
  environmentClass: mw1.medium
  minWorkers: 2
  maxWorkers: 10
  schedulers: 2
  webserverAccessMode: PRIVATE_ONLY
  loggingConfiguration:
    dagProcessingLogs:
      enabled: true
      logLevel: INFO
    schedulerLogs:
      enabled: true
      logLevel: WARNING
    taskLogs:
      enabled: true
      logLevel: INFO
    webserverLogs:
      enabled: true
      logLevel: WARNING
    workerLogs:
      enabled: true
      logLevel: INFO
  weeklyMaintenanceWindowStart: "SUN:03:00"
```

## Related resources

| Resource | Relationship |
|---|---|
| `AwsS3Bucket` | Stores DAG files, plugins.zip, requirements.txt, and startup scripts. Must have versioning enabled. |
| `AwsIamRole` | Execution role assumed by MWAA for S3, CloudWatch Logs, SQS, and DAG service access. |
| `AwsVpc` | Provides private subnets and VPC ID for MWAA network interfaces and managed security group. |
| `AwsSecurityGroup` | Source security groups for managed ingress rules on port 443. |
| `AwsKmsKey` | Customer-managed KMS key for at-rest encryption of metadata DB, logs, and SQS queue. |
| `AwsCloudwatchLogGroup` | Destination for Airflow log modules (auto-created by MWAA, but can be pre-created for retention policies). |

## Cross-field validations

The spec enforces five cross-field validations at the protobuf level:

1. **max_workers ≥ min_workers** — `maxWorkers` must be ≥ `minWorkers` when both are specified.
2. **max_webservers ≥ min_webservers** — `maxWebservers` must be ≥ `minWebservers` when both are specified.
3. **dag_s3_path is relative** — `dagS3Path` must not start with `/`.
4. **vpc_id required for managed SG** — `vpcId` is required when `securityGroupIds` or `allowedCidrBlocks` are provided.
5. **Security coverage required** — At least one of `vpcId` (for managed SG) or `associateSecurityGroupIds` must be provided for MWAA VPC endpoint security.

## Deliberately omitted features

The following MWAA features are **not** covered by this v1 API. They may be added in future versions:

| Feature | Reason |
|---|---|
| Custom VPC endpoint creation (`CUSTOMER` endpoint management) | Requires pre-existing VPC endpoints with specific service names. Less than 5% adoption. |
| MWAA web login token generation | Runtime operation (`CreateWebLoginToken` API) — not an infrastructure concern. |
| DAG trigger / task management | Runtime Airflow operations via REST API — outside IaC scope. |
| Environment version rollback | Requires snapshot/restore capability not yet available in MWAA. |
| Cross-account S3 bucket references | Complex IAM trust policy setup; better handled as a pattern guide. |
| Airflow connections/variables as IaC | Environment-level secrets management; better modeled as a separate resource or startup script. |

## How it works

Planton provisions the MWAA environment via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (`api.proto`, `spec.proto`, `stack_outputs.proto`) and stack execution is orchestrated by the platform using the `AwsMwaaEnvironmentStackInput` (includes provider credentials and target resource).

When `securityGroupIds` or `allowedCidrBlocks` are provided alongside `vpcId`, the module creates a managed security group with:
- A self-referencing inbound rule (all traffic) for MWAA component intercommunication.
- HTTPS (port 443) ingress from source security groups and/or CIDR blocks.
- Full egress for outbound connectivity.

This managed security group is combined with any `associateSecurityGroupIds` and attached to the MWAA environment's network configuration.

## References

- [Amazon MWAA Documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/what-is-mwaa.html)
- [MWAA Networking](https://docs.aws.amazon.com/mwaa/latest/userguide/networking-about.html)
- [MWAA Environment Class Sizing](https://docs.aws.amazon.com/mwaa/latest/userguide/environment-class.html)
- [MWAA Pricing](https://aws.amazon.com/managed-workflows-for-apache-airflow/pricing/)
- [Apache Airflow Documentation](https://airflow.apache.org/docs/)

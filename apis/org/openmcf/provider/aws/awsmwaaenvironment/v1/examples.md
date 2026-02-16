# AwsMwaaEnvironment Examples

Realistic deployment examples for the AwsMwaaEnvironment resource. Each example is a complete manifest ready for customization.

---

## 1. Minimal private Airflow environment

The simplest viable MWAA environment. Private webserver access with a pre-existing security group attached directly.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: dev-airflow
spec:
  sourceBucketArn:
    value: arn:aws:s3:::dev-airflow-dags-bucket
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::111122223333:role/mwaa-dev-execution-role
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  associateSecurityGroupIds:
    - value: sg-0mwaadev00000001
```

**Key points:**
- Uses `associateSecurityGroupIds` to attach a pre-existing security group directly — no managed SG creation, so `vpcId` is not required.
- `environmentClass` defaults to `mw1.small` (1 vCPU, 2 GB) — adequate for development and testing.
- `webserverAccessMode` defaults to `PRIVATE_ONLY` — the Airflow UI is accessible only from within the VPC.
- `minWorkers` defaults to 1, `maxWorkers` defaults to 10. MWAA auto-scales based on task queue depth.
- AWS uses the latest supported Airflow version and the default `aws/airflow` KMS key.

---

## 2. Production-grade with KMS + full logging

Production environment with customer-managed encryption, all 5 log modules enabled, explicit sizing, and a scheduled maintenance window.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: prod-data-pipelines
spec:
  airflowVersion: "2.10.1"
  sourceBucketArn:
    value: arn:aws:s3:::prod-airflow-artifacts
  dagS3Path: dags/
  pluginsS3Path: plugins/plugins.zip
  pluginsS3ObjectVersion: "v3.2.1"
  requirementsS3Path: requirements/requirements.txt
  requirementsS3ObjectVersion: "v2.0.0"
  startupScriptS3Path: scripts/startup.sh
  executionRoleArn:
    value: arn:aws:iam::111122223333:role/mwaa-prod-execution-role
  subnetIds:
    - value: subnet-prod-useast1a
    - value: subnet-prod-useast1b
  vpcId:
    value: vpc-0prod123456789abc
  securityGroupIds:
    - value: sg-0datateam00000001
    - value: sg-0mlplatform0000002
  kmsKeyArn:
    value: arn:aws:kms:us-east-1:111122223333:key/mrk-prod-airflow-key
  environmentClass: mw1.medium
  minWorkers: 2
  maxWorkers: 15
  minWebservers: 2
  maxWebservers: 4
  schedulers: 3
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
  workerReplacementStrategy: GRACEFUL
```

**Key points:**
- `mw1.medium` (2 vCPU, 4 GB) handles medium-to-large DAG counts with fast scheduling.
- `pluginsS3ObjectVersion` and `requirementsS3ObjectVersion` pin S3 artifacts for deterministic deployments.
- `securityGroupIds` from two different teams create a managed SG with HTTPS ingress from both.
- KMS key encrypts metadata DB, DAG logs, SQS Celery queue, and webserver logs.
- All 5 log modules deliver to CloudWatch Logs at appropriate levels (INFO for execution, WARNING for control plane).
- `GRACEFUL` worker replacement ensures running tasks complete before workers are replaced during updates.
- Maintenance window is set to early Sunday morning UTC to minimize impact on business-hours workflows.

---

## 3. Public access with plugins and requirements

Internet-accessible Airflow UI for teams without VPN access. Includes custom plugins and Python dependencies.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: analytics-airflow
spec:
  airflowVersion: "2.9.2"
  sourceBucketArn:
    value: arn:aws:s3:::analytics-airflow-bucket
  dagS3Path: airflow/dags/
  pluginsS3Path: airflow/plugins/plugins.zip
  requirementsS3Path: airflow/requirements/requirements.txt
  startupScriptS3Path: airflow/scripts/install-system-deps.sh
  executionRoleArn:
    value: arn:aws:iam::444455556666:role/analytics-mwaa-role
  subnetIds:
    - value: subnet-analytics-az1
    - value: subnet-analytics-az2
  associateSecurityGroupIds:
    - value: sg-0analytics0000001
  environmentClass: mw1.small
  minWorkers: 1
  maxWorkers: 5
  webserverAccessMode: PUBLIC_ONLY
  loggingConfiguration:
    taskLogs:
      enabled: true
      logLevel: INFO
    workerLogs:
      enabled: true
      logLevel: INFO
```

**Key points:**
- `PUBLIC_ONLY` makes the Airflow UI accessible over the internet with IAM-based login (AWS SSO or federated credentials).
- DAG folder is nested under `airflow/dags/` — common pattern when the S3 bucket serves multiple purposes.
- Only task and worker logs are enabled — sufficient for debugging DAG execution without the cost of all 5 modules.
- Startup script installs OS-level system dependencies that `requirements.txt` cannot handle (e.g., `apt-get install libpq-dev`).
- `mw1.small` (1 vCPU, 2 GB) is adequate for analytics teams with <50 DAGs.

---

## 4. Custom Airflow configuration overrides

Environment with extensive Airflow configuration tuning for performance-sensitive workloads.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: tuned-airflow
spec:
  airflowVersion: "2.10.1"
  sourceBucketArn:
    value: arn:aws:s3:::tuned-airflow-dags
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::777788889999:role/mwaa-tuned-execution
  subnetIds:
    - value: subnet-tuned-az1
    - value: subnet-tuned-az2
  associateSecurityGroupIds:
    - value: sg-0tuned0000000001
  environmentClass: mw1.large
  minWorkers: 4
  maxWorkers: 25
  schedulers: 4
  airflowConfigurationOptions:
    core.default_timezone: "America/New_York"
    core.parallelism: "64"
    core.max_active_tasks_per_dag: "32"
    core.max_active_runs_per_dag: "8"
    core.dag_file_processor_timeout: "120"
    scheduler.parsing_processes: "4"
    scheduler.min_file_process_interval: "60"
    celery.worker_autoscale: "16,4"
    celery.worker_concurrency: "16"
    webserver.dag_default_view: "grid"
    webserver.default_dag_run_display_number: "50"
    webserver.page_size: "100"
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
    workerLogs:
      enabled: true
      logLevel: INFO
```

**Key points:**
- `mw1.large` (4 vCPU, 8 GB) provides capacity for hundreds of DAGs with complex parsing.
- `core.parallelism: "64"` allows up to 64 concurrent task instances across all DAGs.
- `celery.worker_autoscale: "16,4"` sets each worker's Celery concurrency range (max 16, min 4).
- `scheduler.parsing_processes: "4"` with 4 schedulers gives 16 parallel DAG-parsing processes.
- `core.default_timezone` sets the Airflow UI to Eastern Time without affecting UTC-based scheduling.
- `webserver.dag_default_view: "grid"` uses the modern grid view instead of the legacy tree view.
- See [MWAA configuration reference](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-env-variables.html) for the full list of allowed keys.

---

## 5. Managed security group with VPC

Environment demonstrating the managed security group pattern: source security groups and CIDR blocks create a managed SG with self-referencing rules and HTTPS ingress.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: platform-airflow
spec:
  sourceBucketArn:
    value: arn:aws:s3:::platform-airflow-dags
  dagS3Path: dags/
  executionRoleArn:
    value: arn:aws:iam::111122223333:role/platform-mwaa-role
  subnetIds:
    - value: subnet-platform-az1
    - value: subnet-platform-az2
  vpcId:
    value: vpc-0platform12345678
  securityGroupIds:
    - value: sg-0dataeng00000001
    - value: sg-0mlops000000002
    - value: sg-0devops00000003
  allowedCidrBlocks:
    - "10.100.0.0/16"
    - "10.200.0.0/16"
    - "172.16.0.0/12"
  environmentClass: mw1.medium
  minWorkers: 2
  maxWorkers: 10
  schedulers: 2
```

**Key points:**
- `vpcId` + `securityGroupIds` triggers managed security group creation.
- The managed SG gets:
  - Self-referencing inbound rule (all traffic, all ports) — required for MWAA internal communication between scheduler, workers, and webserver.
  - HTTPS (port 443) ingress from each of the 3 source security groups.
  - HTTPS (port 443) ingress from all 3 CIDR blocks.
  - Full egress (all traffic, all ports) for outbound connectivity.
- The managed SG ID is exported as the `security_group_id` output.
- No `associateSecurityGroupIds` — the managed SG is the sole security group attached to the environment.

---

## 6. Cross-resource valueFrom references — infrastructure chart pattern

Production environment wired to other OpenMCF resources using `valueFrom` foreign-key references. This is the recommended pattern for infrastructure charts where resources reference each other by name instead of hardcoding IDs.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMwaaEnvironment
metadata:
  name: prod-etl-airflow
spec:
  airflowVersion: "2.10.1"
  sourceBucketArn:
    valueFrom:
      kind: AwsS3Bucket
      name: prod-airflow-dags-bucket
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
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: data-engineering-sg
        fieldPath: status.outputs.security_group_id
    - valueFrom:
        kind: AwsSecurityGroup
        name: ml-platform-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: platform-encryption-key
      fieldPath: status.outputs.key_arn
  environmentClass: mw1.large
  minWorkers: 3
  maxWorkers: 20
  minWebservers: 2
  maxWebservers: 4
  schedulers: 3
  webserverAccessMode: PRIVATE_ONLY
  airflowConfigurationOptions:
    core.parallelism: "48"
    core.max_active_tasks_per_dag: "24"
    celery.worker_autoscale: "12,2"
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
  weeklyMaintenanceWindowStart: "SAT:04:00"
  workerReplacementStrategy: GRACEFUL
```

**Key points:**
- Every infrastructure reference uses `valueFrom` — no hardcoded subnet IDs, VPC IDs, security group IDs, KMS ARNs, or S3 bucket ARNs.
- `valueFrom` references resolve at deployment time from other resources' stack outputs.
- `fieldPath` values match the default ref paths defined in the protobuf annotations: `status.outputs.bucket_arn`, `status.outputs.role_arn`, `status.outputs.private_subnets.[N].id`, `status.outputs.vpc_id`, `status.outputs.security_group_id`, `status.outputs.key_arn`.
- This pattern enables environment promotion (dev → staging → prod) by changing only the referenced resource names.
- `mw1.large` (4 vCPU, 8 GB) with up to 20 workers handles production ETL workloads with hundreds of DAGs.
- `celery.worker_autoscale: "12,2"` sets per-worker concurrency (max 12, min 2), so the effective max concurrency is 12 × 20 = 240 parallel tasks.

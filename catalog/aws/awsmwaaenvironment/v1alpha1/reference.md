# AwsMwaaEnvironment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsMwaaEnvironmentSpec defines the desired state of an Amazon MWAA (Managed Workflows for Apache Airflow) environment.
MWAA is a fully managed orchestration service that provisions Apache Airflow environments,
handling scheduler, worker, and webserver infrastructure so teams can author, schedule,
and monitor data pipelines using Airflow DAGs stored in Amazon S3.

Network ingress is composed, never embedded: the environment attaches the
referenced security_group_ids directly, and the rules MWAA needs -- a
self-referencing all-traffic ingress rule for component intercommunication,
HTTPS (443) ingress for Airflow UI access, and outbound egress -- live on
those first-class AwsSecurityGroup nodes where they can be shared, audited,
and evolved independently of the environment.

Provisioning expectation: environment creation is exceptionally slow --
creates commonly run tens of minutes and the provider's default timeouts
allow 120 minutes for create and 90 for update/delete. A deploy that
appears stalled at the 40-minute mark is normal, not hung.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMwaaEnvironment
metadata:
  name: test-mwaa-environment
spec:
  region: us-west-2
  airflowVersion: "2.10.1"
  sourceBucketArn:
    value: "arn:aws:s3:::airflow-dags-bucket"
  dagS3Path: "dags/"
  executionRoleArn:
    value: "arn:aws:iam::123456789012:role/airflow-execution-role"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f60001"
    - value: "subnet-0a1b2c3d4e5f60002"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f60003"
  environmentClass: mw1.small
  minWorkers: 1
  maxWorkers: 5
  schedulers: 2
  webserverAccessMode: PRIVATE_ONLY
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.airflowVersion` | `string` |  |  |  |
| `spec.airflowConfigurationOptions` | `map<string, string>` |  |  |  |
| `spec.sourceBucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.dagS3Path` | `string` | yes |  |  |
| `spec.pluginsS3Path` | `string` |  |  |  |
| `spec.pluginsS3ObjectVersion` | `string` |  |  |  |
| `spec.requirementsS3Path` | `string` |  |  |  |
| `spec.requirementsS3ObjectVersion` | `string` |  |  |  |
| `spec.startupScriptS3Path` | `string` |  |  |  |
| `spec.startupScriptS3ObjectVersion` | `string` |  |  |  |
| `spec.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.environmentClass` | `string` |  |  |  |
| `spec.minWorkers` | `int32` |  |  |  |
| `spec.maxWorkers` | `int32` |  |  |  |
| `spec.minWebservers` | `int32` |  |  |  |
| `spec.maxWebservers` | `int32` |  |  |  |
| `spec.schedulers` | `int32` |  |  |  |
| `spec.webserverAccessMode` | `string` |  | `PRIVATE_ONLY` |  |
| `spec.endpointManagement` | `string` |  |  |  |
| `spec.loggingConfiguration` | `AwsMwaaEnvironmentLoggingConfiguration` |  |  |  |
| `spec.loggingConfiguration.dagProcessingLogs` | `AwsMwaaEnvironmentLoggingModuleConfig` |  |  |  |
| `spec.loggingConfiguration.dagProcessingLogs.enabled` | `bool` |  |  |  |
| `spec.loggingConfiguration.dagProcessingLogs.logLevel` | `string` |  |  |  |
| `spec.loggingConfiguration.schedulerLogs` | `AwsMwaaEnvironmentLoggingModuleConfig` |  |  |  |
| `spec.loggingConfiguration.schedulerLogs.enabled` | `bool` |  |  |  |
| `spec.loggingConfiguration.schedulerLogs.logLevel` | `string` |  |  |  |
| `spec.loggingConfiguration.taskLogs` | `AwsMwaaEnvironmentLoggingModuleConfig` |  |  |  |
| `spec.loggingConfiguration.taskLogs.enabled` | `bool` |  |  |  |
| `spec.loggingConfiguration.taskLogs.logLevel` | `string` |  |  |  |
| `spec.loggingConfiguration.webserverLogs` | `AwsMwaaEnvironmentLoggingModuleConfig` |  |  |  |
| `spec.loggingConfiguration.webserverLogs.enabled` | `bool` |  |  |  |
| `spec.loggingConfiguration.webserverLogs.logLevel` | `string` |  |  |  |
| `spec.loggingConfiguration.workerLogs` | `AwsMwaaEnvironmentLoggingModuleConfig` |  |  |  |
| `spec.loggingConfiguration.workerLogs.enabled` | `bool` |  |  |  |
| `spec.loggingConfiguration.workerLogs.logLevel` | `string` |  |  |  |
| `spec.weeklyMaintenanceWindowStart` | `string` |  |  |  |
| `spec.workerReplacementStrategy` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.airflowVersion

`string`

airflow_version is the Apache Airflow version for the environment.
Examples: "2.10.1", "2.9.2", "2.8.1". If omitted, AWS uses the latest supported version.
Minor version upgrades are applied in-place; major version changes force environment replacement.

### spec.airflowConfigurationOptions

`map<string, string>`

airflow_configuration_options overrides specific Apache Airflow configuration properties.
Keys use the Airflow "section.property" format, e.g., "core.default_timezone",
"webserver.dag_default_view", "celery.worker_autoscale".
Values may contain sensitive information (database URIs, API keys) -- treat as
confidential. The Terraform provider marks the whole map Sensitive and redacts it in
plan output; prefer Airflow connections/variables backed by Secrets Manager for real
credentials rather than pasting them here.
Machinery note: the platform's sensitive marker cannot attach to map
fields (and the coverage gate scopes to name-flagged fields), so the
Terraform provider's Sensitive treatment is the redaction this map gets.

### spec.sourceBucketArn

`string | valueFrom` · required

source_bucket_arn is the ARN of the S3 bucket containing DAGs, plugins, and requirements.
The bucket must have versioning enabled and a bucket policy granting MWAA access.
The execution role must have permissions to read from this bucket.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.dagS3Path

`string` · required

dag_s3_path is the relative path within the S3 bucket to the folder containing DAG files.
Example: "dags/" or "airflow/dags/". Must not start with "/".

- rule: {"required":true}

### spec.pluginsS3Path

`string`

plugins_s3_path is the relative path within the S3 bucket to a plugins.zip file.
The zip contains custom Airflow plugins (operators, hooks, sensors, macros).
Example: "plugins/plugins.zip".

### spec.pluginsS3ObjectVersion

`string`

plugins_s3_object_version pins the plugins.zip to a specific S3 object version.
Ensures deterministic deployments. If omitted, the latest version is used.

### spec.requirementsS3Path

`string`

requirements_s3_path is the relative path within the S3 bucket to a requirements.txt file.
Lists additional Python packages to install in the Airflow environment.
Example: "requirements/requirements.txt".

### spec.requirementsS3ObjectVersion

`string`

requirements_s3_object_version pins the requirements.txt to a specific S3 object version.
Ensures deterministic deployments. If omitted, the latest version is used.

### spec.startupScriptS3Path

`string`

startup_script_s3_path is the relative path within the S3 bucket to a startup shell script.
Runs at environment startup for OS-level setup (install system packages, set environment
variables, configure authentication) that requirements.txt cannot handle.
Available for Airflow 2.x+. Example: "scripts/startup.sh".

### spec.startupScriptS3ObjectVersion

`string`

startup_script_s3_object_version pins the startup script to a specific S3 object version.
Ensures deterministic deployments. If omitted, the latest version is used.

### spec.executionRoleArn

`string | valueFrom` · required

execution_role_arn is the ARN of the IAM role that MWAA assumes to access AWS resources.
This role needs permissions for S3 (DAGs bucket), CloudWatch Logs, SQS (Celery backend),
and any AWS services your DAGs interact with (e.g., Glue, EMR, Redshift, Lambda).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom` · required

subnet_ids are the private VPC subnets where MWAA creates network interfaces.
Requires exactly 2 subnets in different Availability Zones. Must be private subnets
(no direct route to an internet gateway). ForceNew: changing subnets forces replacement.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom` · required

security_group_ids are the security groups ATTACHED to the MWAA VPC
endpoints -- they define what can reach the environment. AWS requires at
least one. The referenced AwsSecurityGroup must carry a self-referencing
all-traffic ingress rule (MWAA components communicate with each other
through it), HTTPS (443) ingress from whatever should reach the Airflow
UI, and egress for outbound connectivity. Unlike subnets, the attached
groups can be changed in place after creation.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.kmsKeyArn

`string | valueFrom`

kms_key_arn is the KMS key ARN for encrypting environment data at rest
(metadata database, DAG logs, SQS queue, web server logs).
If omitted, AWS uses the default aws/airflow service key.
ForceNew: changing the KMS key forces environment replacement.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.environmentClass

`string`

environment_class determines the compute and memory capacity of the Airflow components.
"mw1.micro": 0.5 vCPU, 1 GB (dev/test, limited to 1 webserver).
"mw1.small" (default): 1 vCPU, 2 GB (small-medium workloads).
"mw1.medium": 2 vCPU, 4 GB (medium-large workloads).
"mw1.large": 4 vCPU, 8 GB (large workloads).
"mw1.xlarge": 8 vCPU, 16 GB (very large workloads).
"mw1.2xlarge": 16 vCPU, 32 GB (maximum capacity).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["mw1.micro","mw1.small","mw1.medium","mw1.large","mw1.xlarge","mw1.2xlarge"]}}

### spec.minWorkers

`int32`

min_workers is the minimum number of Celery workers for auto-scaling.
Range: >= 1. Workers process DAG tasks. MWAA scales between min_workers and max_workers
based on task queue depth. Default: 1.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.maxWorkers

`int32`

max_workers is the maximum number of Celery workers for auto-scaling.
Range: >= 1. Must be >= min_workers. Default: 10.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.minWebservers

`int32`

min_webservers is the minimum number of Airflow webservers.
Range: 2-5 (1 for mw1.micro). Default: 2.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":5,"gte":1}}

### spec.maxWebservers

`int32`

max_webservers is the maximum number of Airflow webservers.
Range: 2-5 (1 for mw1.micro). Must be >= min_webservers. Default: 2.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":5,"gte":1}}

### spec.schedulers

`int32`

schedulers is the number of Airflow schedulers.
Range: 2-5. Default: 2. More schedulers improve DAG parsing and scheduling throughput.
Note: stricter than AWS (which accepts any count); the 2-5 window is the
supported range for Airflow 2.x on standard environment classes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":5,"gte":2}}

### spec.webserverAccessMode

`string` · optional (explicit presence)

webserver_access_mode controls how the Airflow web UI is accessed.
"PRIVATE_ONLY" (default): accessible only within the VPC via VPC endpoint.
"PUBLIC_ONLY": accessible over the internet with IAM-based login.
"PUBLIC_AND_PRIVATE": reachable both over the internet and privately from
within the VPC.

- default: `PRIVATE_ONLY`
- rule: {"string":{"in":["PRIVATE_ONLY","PUBLIC_ONLY","PUBLIC_AND_PRIVATE"]}}

### spec.endpointManagement

`string`

endpoint_management controls who manages the VPC endpoints for the environment.
"SERVICE" (default): AWS creates and manages VPC endpoints automatically.
"CUSTOMER": you create and manage VPC endpoints yourself against the
database_vpc_endpoint_service and webserver_vpc_endpoint_service
outputs (advanced, <5% adoption).
ForceNew: changing this forces environment replacement.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SERVICE","CUSTOMER"]}}

### spec.loggingConfiguration

`AwsMwaaEnvironmentLoggingConfiguration`

logging_configuration controls per-module Airflow log delivery to CloudWatch Logs.
MWAA supports 5 log modules: DAG processing, scheduler, task, webserver, and worker.
Each module can be independently enabled with its own log level.
CloudWatch Logs groups are auto-created by MWAA in the format:
/aws/mwaa/{environment-name}/{module-name}

### spec.loggingConfiguration.dagProcessingLogs

`AwsMwaaEnvironmentLoggingModuleConfig`

dag_processing_logs controls logging for the DAG processing component.
The DAG processor parses DAG files to determine scheduling requirements.

### spec.loggingConfiguration.dagProcessingLogs.enabled

`bool`

enabled controls whether logs for this module are delivered to CloudWatch Logs.

### spec.loggingConfiguration.dagProcessingLogs.logLevel

`string`

log_level sets the minimum severity of log messages delivered.
"CRITICAL": only critical errors. "ERROR": errors and above.
"WARNING": warnings and above. "INFO" (default): informational and above.
"DEBUG": all messages including debug output (verbose, higher CloudWatch costs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","ERROR","WARNING","INFO","DEBUG"]}}

### spec.loggingConfiguration.schedulerLogs

`AwsMwaaEnvironmentLoggingModuleConfig`

scheduler_logs controls logging for the Airflow scheduler.
The scheduler triggers task instances based on DAG definitions and timing.

### spec.loggingConfiguration.schedulerLogs.enabled

`bool`

enabled controls whether logs for this module are delivered to CloudWatch Logs.

### spec.loggingConfiguration.schedulerLogs.logLevel

`string`

log_level sets the minimum severity of log messages delivered.
"CRITICAL": only critical errors. "ERROR": errors and above.
"WARNING": warnings and above. "INFO" (default): informational and above.
"DEBUG": all messages including debug output (verbose, higher CloudWatch costs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","ERROR","WARNING","INFO","DEBUG"]}}

### spec.loggingConfiguration.taskLogs

`AwsMwaaEnvironmentLoggingModuleConfig`

task_logs controls logging for task execution.
Task logs capture stdout/stderr from individual DAG task runs.

### spec.loggingConfiguration.taskLogs.enabled

`bool`

enabled controls whether logs for this module are delivered to CloudWatch Logs.

### spec.loggingConfiguration.taskLogs.logLevel

`string`

log_level sets the minimum severity of log messages delivered.
"CRITICAL": only critical errors. "ERROR": errors and above.
"WARNING": warnings and above. "INFO" (default): informational and above.
"DEBUG": all messages including debug output (verbose, higher CloudWatch costs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","ERROR","WARNING","INFO","DEBUG"]}}

### spec.loggingConfiguration.webserverLogs

`AwsMwaaEnvironmentLoggingModuleConfig`

webserver_logs controls logging for the Airflow webserver (UI).
The webserver serves the Airflow web interface and REST API.

### spec.loggingConfiguration.webserverLogs.enabled

`bool`

enabled controls whether logs for this module are delivered to CloudWatch Logs.

### spec.loggingConfiguration.webserverLogs.logLevel

`string`

log_level sets the minimum severity of log messages delivered.
"CRITICAL": only critical errors. "ERROR": errors and above.
"WARNING": warnings and above. "INFO" (default): informational and above.
"DEBUG": all messages including debug output (verbose, higher CloudWatch costs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","ERROR","WARNING","INFO","DEBUG"]}}

### spec.loggingConfiguration.workerLogs

`AwsMwaaEnvironmentLoggingModuleConfig`

worker_logs controls logging for Celery workers.
Workers execute the actual task code defined in DAGs.

### spec.loggingConfiguration.workerLogs.enabled

`bool`

enabled controls whether logs for this module are delivered to CloudWatch Logs.

### spec.loggingConfiguration.workerLogs.logLevel

`string`

log_level sets the minimum severity of log messages delivered.
"CRITICAL": only critical errors. "ERROR": errors and above.
"WARNING": warnings and above. "INFO" (default): informational and above.
"DEBUG": all messages including debug output (verbose, higher CloudWatch costs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","ERROR","WARNING","INFO","DEBUG"]}}

### spec.weeklyMaintenanceWindowStart

`string`

weekly_maintenance_window_start is the preferred start time for weekly maintenance.
Format: "DAY:HH:MM" in UTC (e.g., "TUE:03:30", "SUN:00:00").
During maintenance, MWAA may apply patches or updates. If omitted, AWS selects a window.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(MON|TUE|WED|THU|FRI|SAT|SUN):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.workerReplacementStrategy

`string`

worker_replacement_strategy controls how workers are replaced during environment updates.
"FORCED": replaces workers immediately (faster updates, may interrupt running tasks).
"GRACEFUL": waits for running tasks to complete before replacing workers (slower, no data loss).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FORCED","GRACEFUL"]}}

## Validation Rules

- `max_workers_gte_min_workers`: max_workers must be >= min_workers when both are specified
- `max_webservers_gte_min_webservers`: max_webservers must be >= min_webservers when both are specified
- `dag_s3_path_no_leading_slash`: dag_s3_path must be a relative path (must not start with '/')

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsMwaaEnvironment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.environment_arn` | `string` | environment_arn is the Amazon Resource Name of the MWAA environment, used in IAM policies and cross-service references. |
| `status.outputs.environment_name` | `string` | environment_name is the human-readable name of the environment. |
| `status.outputs.webserver_url` | `string` | webserver_url is the URL of the Airflow web UI. Format: "{random-id}.{region}.airflow.amazonaws.com". Access depends on webserver_access_mode (PRIVATE_ONLY requires VPC access, PUBLIC_ONLY is internet-accessible). |
| `status.outputs.airflow_version` | `string` | airflow_version is the effective Apache Airflow version running in the environment. |
| `status.outputs.service_role_arn` | `string` | service_role_arn is the ARN of the AWS service role created by MWAA for managing environment infrastructure (VPC endpoints, CloudWatch Logs, etc.). |
| `status.outputs.environment_class` | `string` | environment_class is the effective environment class (compute capacity). |
| `status.outputs.status` | `string` | status is the current status of the MWAA environment (e.g., "AVAILABLE", "CREATING", "UPDATING", "DELETING"). |
| `status.outputs.created_at` | `string` | created_at is the timestamp when the environment was created. |
| `status.outputs.database_vpc_endpoint_service` | `string` | database_vpc_endpoint_service is the VPC endpoint service name for the environment's Airflow metadata database. When endpoint_management is "CUSTOMER", create an AwsVpcEndpoint against this service name. |
| `status.outputs.webserver_vpc_endpoint_service` | `string` | webserver_vpc_endpoint_service is the VPC endpoint service name for the environment's Airflow webserver. When endpoint_management is "CUSTOMER", create an AwsVpcEndpoint against this service name. Empty when the webserver is PUBLIC_ONLY. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.sourceBucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)

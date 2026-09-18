variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsMwaaEnvironment specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # airflow_version is the Apache Airflow version for the environment.
    # Examples: "2.10.1", "2.9.2", "2.8.1". If omitted, AWS uses the latest supported version.
    # Minor version upgrades are applied in-place; major version changes force environment replacement.
    airflow_version = optional(string, "")

    # airflow_configuration_options overrides specific Apache Airflow configuration properties.
    # Keys use the Airflow "section.property" format, e.g., "core.default_timezone",
    # "webserver.dag_default_view", "celery.worker_autoscale".
    # Values may contain sensitive information (database URIs, API keys) -- treat as
    # confidential. The Terraform provider marks the whole map Sensitive and redacts it in
    # plan output; prefer Airflow connections/variables backed by Secrets Manager for real
    # credentials rather than pasting them here.
    # Machinery note: the platform's sensitive marker cannot attach to map
    # fields (and the coverage gate scopes to name-flagged fields), so the
    # Terraform provider's Sensitive treatment is the redaction this map gets.
    airflow_configuration_options = optional(map(string), {})

    # source_bucket_arn is the ARN of the S3 bucket containing DAGs, plugins, and requirements.
    # The bucket must have versioning enabled and a bucket policy granting MWAA access.
    # The execution role must have permissions to read from this bucket.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    source_bucket_arn = string

    # dag_s3_path is the relative path within the S3 bucket to the folder containing DAG files.
    # Example: "dags/" or "airflow/dags/". Must not start with "/".
    dag_s3_path = string

    # plugins_s3_path is the relative path within the S3 bucket to a plugins.zip file.
    # The zip contains custom Airflow plugins (operators, hooks, sensors, macros).
    # Example: "plugins/plugins.zip".
    plugins_s3_path = optional(string, "")

    # plugins_s3_object_version pins the plugins.zip to a specific S3 object version.
    # Ensures deterministic deployments. If omitted, the latest version is used.
    plugins_s3_object_version = optional(string, "")

    # requirements_s3_path is the relative path within the S3 bucket to a requirements.txt file.
    # Lists additional Python packages to install in the Airflow environment.
    # Example: "requirements/requirements.txt".
    requirements_s3_path = optional(string, "")

    # requirements_s3_object_version pins the requirements.txt to a specific S3 object version.
    # Ensures deterministic deployments. If omitted, the latest version is used.
    requirements_s3_object_version = optional(string, "")

    # startup_script_s3_path is the relative path within the S3 bucket to a startup shell script.
    # Runs at environment startup for OS-level setup (install system packages, set environment
    # variables, configure authentication) that requirements.txt cannot handle.
    # Available for Airflow 2.x+. Example: "scripts/startup.sh".
    startup_script_s3_path = optional(string, "")

    # startup_script_s3_object_version pins the startup script to a specific S3 object version.
    # Ensures deterministic deployments. If omitted, the latest version is used.
    startup_script_s3_object_version = optional(string, "")

    # execution_role_arn is the ARN of the IAM role that MWAA assumes to access AWS resources.
    # This role needs permissions for S3 (DAGs bucket), CloudWatch Logs, SQS (Celery backend),
    # and any AWS services your DAGs interact with (e.g., Glue, EMR, Redshift, Lambda).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    execution_role_arn = string

    # subnet_ids are the private VPC subnets where MWAA creates network interfaces.
    # Requires exactly 2 subnets in different Availability Zones. Must be private subnets
    # (no direct route to an internet gateway). ForceNew: changing subnets forces replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # security_group_ids are the security groups ATTACHED to the MWAA VPC
    # endpoints -- they define what can reach the environment. AWS requires at
    # least one. The referenced AwsSecurityGroup must carry a self-referencing
    # all-traffic ingress rule (MWAA components communicate with each other
    # through it), HTTPS (443) ingress from whatever should reach the Airflow
    # UI, and egress for outbound connectivity. Unlike subnets, the attached
    # groups can be changed in place after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = list(string)

    # kms_key_arn is the KMS key ARN for encrypting environment data at rest
    # (metadata database, DAG logs, SQS queue, web server logs).
    # If omitted, AWS uses the default aws/airflow service key.
    # ForceNew: changing the KMS key forces environment replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # environment_class determines the compute and memory capacity of the Airflow components.
    # "mw1.micro": 0.5 vCPU, 1 GB (dev/test, limited to 1 webserver).
    # "mw1.small" (default): 1 vCPU, 2 GB (small-medium workloads).
    # "mw1.medium": 2 vCPU, 4 GB (medium-large workloads).
    # "mw1.large": 4 vCPU, 8 GB (large workloads).
    # "mw1.xlarge": 8 vCPU, 16 GB (very large workloads).
    # "mw1.2xlarge": 16 vCPU, 32 GB (maximum capacity).
    environment_class = optional(string, "")

    # min_workers is the minimum number of Celery workers for auto-scaling.
    # Range: >= 1. Workers process DAG tasks. MWAA scales between min_workers and max_workers
    # based on task queue depth. Default: 1.
    min_workers = optional(number, 0)

    # max_workers is the maximum number of Celery workers for auto-scaling.
    # Range: >= 1. Must be >= min_workers. Default: 10.
    max_workers = optional(number, 0)

    # min_webservers is the minimum number of Airflow webservers.
    # Range: 2-5 (1 for mw1.micro). Default: 2.
    min_webservers = optional(number, 0)

    # max_webservers is the maximum number of Airflow webservers.
    # Range: 2-5 (1 for mw1.micro). Must be >= min_webservers. Default: 2.
    max_webservers = optional(number, 0)

    # schedulers is the number of Airflow schedulers.
    # Range: 2-5. Default: 2. More schedulers improve DAG parsing and scheduling throughput.
    # Note: stricter than AWS (which accepts any count); the 2-5 window is the
    # supported range for Airflow 2.x on standard environment classes.
    schedulers = optional(number, 0)

    # webserver_access_mode controls how the Airflow web UI is accessed.
    # "PRIVATE_ONLY" (default): accessible only within the VPC via VPC endpoint.
    # "PUBLIC_ONLY": accessible over the internet with IAM-based login.
    # "PUBLIC_AND_PRIVATE": reachable both over the internet and privately from
    # within the VPC.
    webserver_access_mode = optional(string)

    # endpoint_management controls who manages the VPC endpoints for the environment.
    # "SERVICE" (default): AWS creates and manages VPC endpoints automatically.
    # "CUSTOMER": you create and manage VPC endpoints yourself against the
    # database_vpc_endpoint_service and webserver_vpc_endpoint_service stack
    # outputs (advanced, <5% adoption).
    # ForceNew: changing this forces environment replacement.
    endpoint_management = optional(string, "")

    # logging_configuration controls per-module Airflow log delivery to CloudWatch Logs.
    # MWAA supports 5 log modules: DAG processing, scheduler, task, webserver, and worker.
    # Each module can be independently enabled with its own log level.
    # CloudWatch Logs groups are auto-created by MWAA in the format:
    # /aws/mwaa/{environment-name}/{module-name}
    logging_configuration = optional(object({
      # dag_processing_logs controls logging for the DAG processing component.
      # The DAG processor parses DAG files to determine scheduling requirements.
      dag_processing_logs = optional(object({
        # enabled controls whether logs for this module are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_level sets the minimum severity of log messages delivered.
        # "CRITICAL": only critical errors. "ERROR": errors and above.
        # "WARNING": warnings and above. "INFO" (default): informational and above.
        # "DEBUG": all messages including debug output (verbose, higher CloudWatch costs).
        log_level = optional(string, "")
      }))

      # scheduler_logs controls logging for the Airflow scheduler.
      # The scheduler triggers task instances based on DAG definitions and timing.
      scheduler_logs = optional(object({
        # enabled controls whether logs for this module are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_level sets the minimum severity of log messages delivered.
        # "CRITICAL": only critical errors. "ERROR": errors and above.
        # "WARNING": warnings and above. "INFO" (default): informational and above.
        # "DEBUG": all messages including debug output (verbose, higher CloudWatch costs).
        log_level = optional(string, "")
      }))

      # task_logs controls logging for task execution.
      # Task logs capture stdout/stderr from individual DAG task runs.
      task_logs = optional(object({
        # enabled controls whether logs for this module are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_level sets the minimum severity of log messages delivered.
        # "CRITICAL": only critical errors. "ERROR": errors and above.
        # "WARNING": warnings and above. "INFO" (default): informational and above.
        # "DEBUG": all messages including debug output (verbose, higher CloudWatch costs).
        log_level = optional(string, "")
      }))

      # webserver_logs controls logging for the Airflow webserver (UI).
      # The webserver serves the Airflow web interface and REST API.
      webserver_logs = optional(object({
        # enabled controls whether logs for this module are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_level sets the minimum severity of log messages delivered.
        # "CRITICAL": only critical errors. "ERROR": errors and above.
        # "WARNING": warnings and above. "INFO" (default): informational and above.
        # "DEBUG": all messages including debug output (verbose, higher CloudWatch costs).
        log_level = optional(string, "")
      }))

      # worker_logs controls logging for Celery workers.
      # Workers execute the actual task code defined in DAGs.
      worker_logs = optional(object({
        # enabled controls whether logs for this module are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_level sets the minimum severity of log messages delivered.
        # "CRITICAL": only critical errors. "ERROR": errors and above.
        # "WARNING": warnings and above. "INFO" (default): informational and above.
        # "DEBUG": all messages including debug output (verbose, higher CloudWatch costs).
        log_level = optional(string, "")
      }))
    }))

    # weekly_maintenance_window_start is the preferred start time for weekly maintenance.
    # Format: "DAY:HH:MM" in UTC (e.g., "TUE:03:30", "SUN:00:00").
    # During maintenance, MWAA may apply patches or updates. If omitted, AWS selects a window.
    weekly_maintenance_window_start = optional(string, "")

    # worker_replacement_strategy controls how workers are replaced during environment updates.
    # "FORCED": replaces workers immediately (faster updates, may interrupt running tasks).
    # "GRACEFUL": waits for running tasks to complete before replacing workers (slower, no data loss).
    worker_replacement_strategy = optional(string, "")
  })
}

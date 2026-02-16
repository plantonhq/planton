# AwsMwaaEnvironment — Technical Reference

Comprehensive technical documentation for the AwsMwaaEnvironment deployment component in OpenMCF. This document covers architecture, networking, security, encryption, logging, auto-scaling, configuration, cost, limits, common patterns, and the v2 roadmap.

---

## Table of Contents

1. [MWAA Architecture](#mwaa-architecture)
2. [Networking Model](#networking-model)
3. [Environment Classes and Sizing](#environment-classes-and-sizing)
4. [Cost Model](#cost-model)
5. [Security Model](#security-model)
6. [Logging Architecture](#logging-architecture)
7. [Auto-Scaling Behavior](#auto-scaling-behavior)
8. [Airflow Configuration Overrides](#airflow-configuration-overrides)
9. [S3 Artifacts](#s3-artifacts)
10. [Service Limits and Quotas](#service-limits-and-quotas)
11. [Common Patterns](#common-patterns)
12. [v2 Roadmap](#v2-roadmap)

---

## MWAA Architecture

Amazon MWAA provisions and manages a complete Apache Airflow environment with the following components:

### Scheduler

- The scheduler is the Airflow component responsible for parsing DAG files, determining task dependencies, and triggering task instances based on schedules, sensors, and upstream completion.
- MWAA runs 2–5 scheduler instances (configurable via `schedulers`). More schedulers improve DAG parsing throughput and scheduling latency for environments with hundreds of DAGs.
- Schedulers read DAG files from S3, parse them into a dependency graph, and write task instance records to the metadata database.
- Schedulers communicate with workers via an SQS-backed Celery broker (managed by AWS, not visible in the customer account).

### Workers

- Workers are Celery processes that execute the actual task code defined in DAGs.
- Each worker runs on Fargate-based compute (managed by AWS) sized according to the `environmentClass`.
- Workers pull tasks from the Celery queue (SQS), execute them, and report status back to the metadata database.
- MWAA auto-scales workers between `minWorkers` and `maxWorkers` based on the number of queued and running tasks.
- Each worker's concurrency (number of parallel tasks per worker) is controlled by the `celery.worker_autoscale` Airflow configuration option.

### Webserver

- The webserver hosts the Airflow web UI and REST API.
- MWAA runs 2–5 webserver instances (configurable via `minWebservers` / `maxWebservers`). The `mw1.micro` class is limited to 1 webserver.
- Access is controlled by `webserverAccessMode`:
  - `PRIVATE_ONLY` — accessible only within the VPC via a VPC endpoint. Requires VPN or bastion host for browser access.
  - `PUBLIC_ONLY` — accessible over the internet. Authentication is IAM-based (AWS SSO, federated credentials, or `CreateWebLoginToken` API).
- The webserver URL follows the pattern: `{random-id}.{region}.airflow.amazonaws.com`.

### Metadata Database

- MWAA provisions and manages an Aurora PostgreSQL-compatible metadata database.
- This database stores DAG definitions, task instance state, XCom data, variables, connections, and scheduler bookkeeping.
- The metadata database is not directly accessible to customers — it exists in AWS-managed infrastructure.
- Encrypted at rest using the KMS key specified in `kmsKeyArn` (or the default `aws/airflow` service key).

### Celery Backend (SQS)

- MWAA uses Amazon SQS as the Celery broker and result backend.
- The SQS queue is created and managed by AWS in the service account — not visible in the customer's account.
- Task messages are encrypted using the same KMS key as the rest of the environment.
- The scheduler enqueues task instances; workers dequeue and execute them.
- This is the mechanism behind MWAA's auto-scaling: AWS monitors the SQS queue depth to decide when to add or remove workers.

### Component Interaction Flow

```
DAG files (S3)
     │
     ▼
┌─────────────┐     parse      ┌──────────────────┐
│  Scheduler  │ ──────────────▶│  Metadata DB     │
│  (2-5)      │                │  (Aurora PgSQL)  │
└─────┬───────┘                └──────────────────┘
      │ enqueue tasks                    ▲
      ▼                                  │ update status
┌─────────────┐     execute    ┌─────────┴────────┐
│  SQS Queue  │ ──────────────▶│  Workers         │
│  (Celery)   │                │  (1-N, autoscale)│
└─────────────┘                └──────────────────┘
                                         │
                                         ▼
                               ┌──────────────────┐
                               │  AWS Services    │
                               │  (Glue, EMR,     │
                               │   Lambda, S3...) │
                               └──────────────────┘
                                         │
┌─────────────┐                          │
│  Webserver  │◀─── user/API ────────────┘
│  (2-5)      │     access
└─────────────┘
```

---

## Networking Model

### VPC Placement

MWAA environments are deployed within a customer VPC:

- MWAA creates Elastic Network Interfaces (ENIs) in the 2 private subnets specified by `subnetIds`.
- Subnets must be in 2 different Availability Zones for high availability.
- Subnets must be **private** — no direct route to an internet gateway.
- If DAGs need internet access (e.g., calling external APIs), the private subnets must have a NAT gateway route.
- **ForceNew:** Changing `subnetIds` forces complete environment replacement (20-40 minute operation).

### VPC Endpoints

MWAA creates and manages VPC endpoints for:

- **Webserver endpoint** — serves the Airflow UI and REST API.
- **Scheduler/worker endpoint** — internal communication between Airflow components.
- **Metadata database endpoint** — Aurora PostgreSQL access for schedulers, workers, and webservers.

When `endpointManagement` is `SERVICE` (default), AWS creates these VPC endpoints automatically. When set to `CUSTOMER`, you must pre-create VPC endpoints for the MWAA service (`airflow.api`, `airflow.env`, `airflow.ops`).

### Security Groups

The OpenMCF component supports three security group patterns:

**1. Managed security group from source SGs (`securityGroupIds`)**

Creates an EC2 security group in the specified VPC with:
- Self-referencing inbound rule (all traffic, all ports, protocol `-1`) — **required** for MWAA. The scheduler, workers, webserver, and metadata DB communicate through this self-referencing rule.
- HTTPS (port 443) ingress from each source security group — allows Airflow UI access.
- Full egress (all traffic) for outbound connectivity.

Requires `vpcId` to be set.

**2. Managed security group from CIDRs (`allowedCidrBlocks`)**

Same as above, but with HTTPS ingress rules based on IPv4 CIDR blocks instead of source security groups. Can be combined with `securityGroupIds`.

**3. Direct attachment (`associateSecurityGroupIds`)**

Existing security groups attached directly to the MWAA environment's network configuration. No managed SG creation. Use this when you already have a security group configured with the self-referencing pattern.

All three patterns can be combined. The managed SG (if created) is included alongside any `associateSecurityGroupIds`.

### Self-Referencing Security Group Pattern

This is the most important networking concept for MWAA. Unlike most AWS services where security groups control external access, MWAA requires a security group that allows **all traffic from itself**:

```
┌─ Security Group (sg-xxx) ──────────────────────────┐
│                                                     │
│  Inbound Rules:                                     │
│  ┌────────────────────────────────────────────────┐ │
│  │ Type: All Traffic                              │ │
│  │ Source: sg-xxx (self-referencing)               │ │
│  │ Purpose: MWAA component intercommunication     │ │
│  └────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────┐ │
│  │ Type: HTTPS (443)                              │ │
│  │ Source: sg-clients / CIDR blocks               │ │
│  │ Purpose: Airflow UI and REST API access        │ │
│  └────────────────────────────────────────────────┘ │
│                                                     │
│  Outbound Rules:                                    │
│  ┌────────────────────────────────────────────────┐ │
│  │ Type: All Traffic                              │ │
│  │ Destination: 0.0.0.0/0                         │ │
│  │ Purpose: Internet access (via NAT), AWS APIs   │ │
│  └────────────────────────────────────────────────┘ │
│                                                     │
└─────────────────────────────────────────────────────┘
```

Without the self-referencing rule, MWAA components cannot communicate and the environment fails to reach `AVAILABLE` status.

### DNS Resolution

- Both `enableDnsSupport` and `enableDnsHostnames` must be enabled on the VPC.
- MWAA VPC endpoints create Route 53 private hosted zone entries that resolve to the ENIs in the customer subnets.
- For `PRIVATE_ONLY` access, clients must be in the VPC (or connected via VPN/Direct Connect/peering) to resolve the webserver hostname.

---

## Environment Classes and Sizing

MWAA provides 6 environment classes that determine the compute and memory capacity of **each** Airflow component (scheduler, worker, webserver):

| Class | vCPU | Memory (GB) | Typical Use Case | Max Workers |
|---|---|---|---|---|
| `mw1.micro` | 0.5 | 1 | Dev/test, <10 simple DAGs. Limited to 1 webserver. | 10 |
| `mw1.small` | 1 | 2 | Small workloads, <50 DAGs with moderate complexity. | 25 |
| `mw1.medium` | 2 | 4 | Medium workloads, 50-200 DAGs, multi-team environments. | 25 |
| `mw1.large` | 4 | 8 | Large workloads, 200-500 DAGs, complex DAG logic. | 25 |
| `mw1.xlarge` | 8 | 16 | Very large workloads, 500-1000 DAGs, heavy task concurrency. | 25 |
| `mw1.2xlarge` | 16 | 32 | Maximum capacity, 1000+ DAGs, enterprise-scale pipelines. | 25 |

### Sizing Guidance

**Scheduler sizing** depends on:
- Number of DAGs and DAG file complexity (parsing time).
- Number of tasks per DAG and scheduling interval.
- Number of variables, connections, and XCom entries.
- Rule of thumb: `mw1.small` handles up to 50 DAGs; `mw1.medium` handles 50-200; `mw1.large`+ for 200+.

**Worker sizing** depends on:
- Task memory requirements (e.g., Pandas DataFrames, ML model loading).
- Task CPU requirements (e.g., data transformations, compression).
- External service call latency (workers waiting on HTTP/database calls still consume memory).
- Rule of thumb: If tasks load large datasets into memory, size up. If tasks mostly call external APIs, `mw1.small` workers with higher concurrency may suffice.

**Webserver sizing** depends on:
- Number of concurrent UI users.
- REST API call volume.
- DAG count (affects UI rendering time).
- Rule of thumb: `mw1.small` handles up to 10 concurrent users; `mw1.medium` for 10-50.

### Scheduler Count Guidance

| Schedulers | When to use |
|---|---|
| 2 | Default. Adequate for most environments with <200 DAGs. |
| 3 | 200-500 DAGs or when DAG files are complex (many imports, dynamic generation). |
| 4-5 | 500+ DAGs, or when scheduling latency must be minimized (SLA-critical pipelines). |

---

## Cost Model

MWAA pricing has several dimensions. All prices are approximate (us-east-1, as of 2025) and should be verified on the [AWS pricing page](https://aws.amazon.com/managed-workflows-for-apache-airflow/pricing/).

### Base Environment Cost

The environment itself has a per-hour charge based on the environment class:

| Class | Approx. Cost/Hour | Approx. Cost/Month (730h) |
|---|---|---|
| `mw1.micro` | $0.028 | ~$20 |
| `mw1.small` | $0.055 | ~$40 |
| `mw1.medium` | $0.110 | ~$80 |
| `mw1.large` | $0.220 | ~$161 |
| `mw1.xlarge` | $0.440 | ~$321 |
| `mw1.2xlarge` | $0.880 | ~$642 |

### Worker Cost

Workers are charged per worker-hour based on the environment class:

| Class | Approx. Worker Cost/Hour |
|---|---|
| `mw1.micro` | $0.016 |
| `mw1.small` | $0.033 |
| `mw1.medium` | $0.065 |
| `mw1.large` | $0.130 |
| `mw1.xlarge` | $0.260 |
| `mw1.2xlarge` | $0.520 |

**Important:** You pay for at least `minWorkers` workers at all times, even when idle. MWAA auto-scales up to `maxWorkers` based on demand.

### Additional Webserver Cost

Each additional webserver beyond the baseline (2) incurs a per-hour charge similar to the worker cost.

### Additional Scheduler Cost

Each additional scheduler beyond 2 incurs a per-hour charge.

### Other Costs

- **CloudWatch Logs** — charged per GB ingested and stored. With DEBUG logging on all 5 modules, costs can be significant.
- **S3 storage** — for DAG files, plugins, requirements. Typically negligible.
- **KMS** — $1/month per CMK + $0.03 per 10,000 API calls.
- **NAT Gateway** — if private subnets use NAT for internet access, data processing charges apply.
- **VPC endpoints** — when `endpointManagement` is `SERVICE`, AWS manages endpoints at no additional VPC endpoint charge.

### Cost Optimization Tips

1. Use `mw1.micro` or `mw1.small` for development environments. Delete dev environments when not in use.
2. Set `minWorkers: 1` for non-critical environments to minimize idle worker costs.
3. Use `WARNING` or `ERROR` log levels for control-plane modules (scheduler, webserver, DAG processing) to reduce CloudWatch Logs costs.
4. Schedule maintenance windows during low-activity periods to avoid scaling up replacement workers during peak hours.
5. Right-size `environmentClass` based on actual scheduler and worker utilization (CloudWatch metrics).

---

## Security Model

### IAM Execution Role

The execution role (`executionRoleArn`) is the IAM role that MWAA assumes to:
- Read DAG files, plugins, and requirements from S3.
- Write logs to CloudWatch Logs.
- Send and receive messages from the SQS Celery queue.
- Access any AWS services your DAGs call (Glue, EMR, Redshift, Lambda, SageMaker, etc.).

**Required permissions:**
```json
{
  "Effect": "Allow",
  "Action": [
    "s3:GetObject*",
    "s3:ListBucket"
  ],
  "Resource": [
    "arn:aws:s3:::your-dags-bucket",
    "arn:aws:s3:::your-dags-bucket/*"
  ]
}
```
```json
{
  "Effect": "Allow",
  "Action": [
    "logs:CreateLogStream",
    "logs:CreateLogGroup",
    "logs:PutLogEvents",
    "logs:GetLogEvents",
    "logs:GetLogRecord",
    "logs:GetLogGroupFields",
    "logs:GetQueryResults"
  ],
  "Resource": "arn:aws:logs:*:*:log-group:airflow-*"
}
```
```json
{
  "Effect": "Allow",
  "Action": [
    "sqs:ChangeMessageVisibility",
    "sqs:DeleteMessage",
    "sqs:GetQueueAttributes",
    "sqs:GetQueueUrl",
    "sqs:ReceiveMessage",
    "sqs:SendMessage"
  ],
  "Resource": "arn:aws:sqs:*:*:airflow-celery-*"
}
```

The role must also have a trust policy allowing `airflow.amazonaws.com` and `airflow-env.amazonaws.com` to assume it.

### KMS Encryption

When `kmsKeyArn` is specified, MWAA encrypts:
- The Aurora PostgreSQL metadata database (at-rest encryption).
- DAG logs stored in CloudWatch Logs.
- SQS messages (Celery task queue).
- Webserver session data.

The KMS key must grant `kms:GenerateDataKey*`, `kms:Decrypt`, `kms:DescribeKey`, and `kms:CreateGrant` to the MWAA service role and the execution role.

**ForceNew:** Changing the KMS key forces complete environment replacement.

### VPC Isolation

- `PRIVATE_ONLY` environments are completely isolated within the VPC. The webserver endpoint is only accessible from within the VPC or via VPC peering/VPN/Direct Connect.
- `PUBLIC_ONLY` environments expose the webserver over the internet but require IAM authentication (AWS SSO, federated credentials, or `CreateWebLoginToken` API).
- Even with `PUBLIC_ONLY`, workers and schedulers communicate exclusively through VPC endpoints — only the webserver is internet-facing.

### Data in Transit

- All MWAA internal communication (scheduler ↔ metadata DB, worker ↔ SQS, webserver ↔ metadata DB) uses TLS.
- The Airflow UI is served over HTTPS (port 443).
- S3 access uses HTTPS endpoints by default.

---

## Logging Architecture

### Log Modules

MWAA supports 5 independent log modules, each delivering to its own CloudWatch Logs log group:

| Module | Log Group Name | What It Captures |
|---|---|---|
| DAG Processing | `/aws/mwaa/{env-name}/DAGProcessing` | DAG file parsing: import errors, syntax errors, parse time, dynamic DAG generation. |
| Scheduler | `/aws/mwaa/{env-name}/Scheduler` | Scheduling decisions: task triggers, dependency resolution, pool assignments, SLA misses. |
| Task | `/aws/mwaa/{env-name}/Task` | Individual task execution: stdout/stderr from task code, operator logs, XCom pushes. |
| Webserver | `/aws/mwaa/{env-name}/Webserver` | Web UI activity: HTTP requests, authentication events, REST API calls, Flask errors. |
| Worker | `/aws/mwaa/{env-name}/Worker` | Celery worker lifecycle: task dequeue, execution start/end, heartbeats, resource usage. |

### Log Levels

Each module supports 5 log levels (from least to most verbose):

| Level | Description | Recommended For |
|---|---|---|
| `CRITICAL` | Only critical errors that prevent Airflow from functioning. | Rarely used standalone. |
| `ERROR` | Errors that affect specific tasks or DAGs but don't crash the system. | Alerting on production environments. |
| `WARNING` | Potential issues: deprecated features, slow parsing, nearing limits. | Control-plane modules (scheduler, webserver, DAG processing). |
| `INFO` | Normal operational information: task started, task completed, DAG parsed. | Execution modules (task, worker). Recommended default. |
| `DEBUG` | Verbose debugging: SQL queries, full HTTP request/response, internal state. | **Development only.** High CloudWatch Logs cost. |

### Recommended Log Configuration

**Development:**
```yaml
loggingConfiguration:
  taskLogs:
    enabled: true
    logLevel: DEBUG
  workerLogs:
    enabled: true
    logLevel: DEBUG
```

**Production:**
```yaml
loggingConfiguration:
  dagProcessingLogs:
    enabled: true
    logLevel: WARNING
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
```

### CloudWatch Logs Integration

- Log groups are **auto-created by MWAA** when logging is enabled. You do not need to pre-create them.
- However, you can pre-create log groups (e.g., via `AwsCloudwatchLogGroup`) to set custom retention policies, encryption, and subscription filters before MWAA starts writing.
- MWAA uses its service role (not the execution role) to create log groups and write log events.
- Log group retention is set to "Never Expire" by default. To control costs, configure retention policies (7, 14, 30, 60, 90 days) on the pre-created log groups.

---

## Auto-Scaling Behavior

### Worker Auto-Scaling

MWAA automatically scales workers between `minWorkers` and `maxWorkers`:

1. **Scale-up trigger:** When the number of queued tasks exceeds the current worker capacity, MWAA adds workers. New workers typically take 2-5 minutes to become available.
2. **Scale-down trigger:** When queued tasks are zero and running tasks are below capacity, MWAA removes excess workers after a cooldown period.
3. **Minimum guarantee:** At least `minWorkers` workers are always running, even when idle. You pay for idle workers.
4. **Per-worker concurrency:** Each worker can execute multiple tasks concurrently. The default is controlled by `celery.worker_concurrency` (defaults vary by `environmentClass`). Fine-tune with `celery.worker_autoscale` (e.g., `"16,4"` means max 16, min 4 concurrent tasks per worker).

**Scaling formula:**
```
Effective concurrency = number_of_workers × per_worker_concurrency
```

Example: 5 workers × 12 tasks/worker = 60 concurrent tasks.

### Webserver Auto-Scaling

MWAA scales webservers between `minWebservers` and `maxWebservers` based on request load:

- Webservers serve the Airflow UI and REST API.
- `mw1.micro` is limited to 1 webserver (no auto-scaling).
- For other classes, the default range is 2–2 (no auto-scaling unless `maxWebservers` > `minWebservers`).
- Scale-up responds to increased HTTP request volume and CPU utilization.

### Scheduler Scaling

Schedulers are **not** auto-scaled — the count is fixed at deployment time via the `schedulers` field (2–5). Changing the count requires an environment update (rolling restart).

---

## Airflow Configuration Overrides

The `airflowConfigurationOptions` map allows overriding Apache Airflow configuration properties. Keys use the `section.property` format.

### Commonly Used Configuration Keys

| Key | Default | Description |
|---|---|---|
| `core.default_timezone` | `utc` | Timezone for the Airflow UI. Does not affect scheduling (always UTC). |
| `core.parallelism` | varies by class | Max concurrent task instances across all DAGs. |
| `core.max_active_tasks_per_dag` | `16` | Max concurrent tasks within a single DAG. |
| `core.max_active_runs_per_dag` | `16` | Max concurrent DAG runs for a single DAG. |
| `core.dag_file_processor_timeout` | `50` | Seconds before a DAG file parse is killed. |
| `scheduler.parsing_processes` | varies | Number of processes for DAG file parsing. |
| `scheduler.min_file_process_interval` | `30` | Minimum seconds between re-parsing a DAG file. |
| `celery.worker_autoscale` | varies | `"max,min"` concurrency per worker (e.g., `"16,4"`). |
| `celery.worker_concurrency` | varies | Fixed concurrency per worker (overrides autoscale). |
| `webserver.dag_default_view` | `grid` | Default DAG view: `grid`, `graph`, `duration`, `gantt`, `landing_times`. |
| `webserver.default_dag_run_display_number` | `25` | Number of DAG runs shown in the UI. |
| `webserver.page_size` | `100` | Default page size for list views. |
| `email.email_backend` | — | Email backend for alerts (e.g., `airflow.providers.amazon.aws.utils.emailer.send_email`). |

### Restricted Configuration Keys

MWAA blocks certain configuration keys for security and operational reasons:

- `core.executor` — always CeleryExecutor on MWAA.
- `core.sql_alchemy_conn` — managed Aurora PostgreSQL connection.
- `celery.broker_url` — managed SQS broker.
- `celery.result_backend` — managed SQS/database backend.
- `webserver.secret_key` — managed by MWAA.
- `logging.*_handler` — log handlers are managed by MWAA.

See [MWAA configuration reference](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-env-variables.html) for the full list of allowed and blocked keys.

---

## S3 Artifacts

MWAA reads all Airflow artifacts from a single S3 bucket (`sourceBucketArn`).

### DAGs (`dagS3Path`)

- Contains Python files (`.py`) defining Airflow DAGs.
- MWAA syncs DAGs from S3 every 30 seconds (configurable via `scheduler.min_file_process_interval`).
- DAG files must be at the specified path (not in subdirectories, unless using Airflow 2.3+ which supports subdirectory scanning).
- The S3 bucket must have **versioning enabled**.

### Plugins (`pluginsS3Path`)

- A `plugins.zip` file containing custom Airflow plugins: operators, hooks, sensors, macros, blueprints.
- The zip must follow the Airflow plugins directory structure:
  ```
  plugins.zip
  ├── operators/
  │   └── custom_operator.py
  ├── hooks/
  │   └── custom_hook.py
  └── sensors/
      └── custom_sensor.py
  ```
- Use `pluginsS3ObjectVersion` to pin to a specific S3 object version for deterministic deployments.

### Requirements (`requirementsS3Path`)

- A `requirements.txt` file listing additional Python packages to install.
- Installed via `pip install` during environment initialization and updates.
- Example:
  ```
  apache-airflow-providers-amazon==8.0.0
  pandas==2.1.0
  boto3==1.28.0
  requests==2.31.0
  ```
- Use `requirementsS3ObjectVersion` for version pinning.
- **Important:** Package installation happens during environment creation/update. It can take 10-30 minutes for complex dependency trees.

### Startup Script (`startupScriptS3Path`)

- A shell script (`.sh`) that runs at environment startup.
- Use cases:
  - Install OS-level system packages (`apt-get install libpq-dev`).
  - Set environment variables for DAG code.
  - Configure authentication (e.g., gcloud auth for GCP access).
  - Install non-Python tools required by operators.
- Available for Airflow 2.x+.
- Use `startupScriptS3ObjectVersion` for version pinning.

---

## Service Limits and Quotas

Key MWAA service limits (default, as of 2025):

| Limit | Default | Adjustable |
|---|---|---|
| Environments per region per account | 10 | Yes (via Service Quotas) |
| DAGs per environment | 1,000 | Soft (depends on class) |
| Workers per environment (max) | 25 | Yes |
| Webservers per environment | 5 | No |
| Schedulers per environment | 5 | No |
| Airflow configuration options | 50 keys | No |
| plugins.zip size | 1 GB | No |
| requirements.txt line count | 200 | No |
| startup script size | 10 KB | No |
| S3 bucket per environment | 1 | No |
| Environment creation time | 20-40 min | N/A |
| Environment update time | 10-30 min | N/A |
| Environment deletion time | 10-20 min | N/A |

### Performance Limits by Environment Class

| Class | Max DAGs (practical) | Max Task Concurrency (practical) |
|---|---|---|
| `mw1.micro` | ~10 | ~5 |
| `mw1.small` | ~50 | ~50 |
| `mw1.medium` | ~200 | ~200 |
| `mw1.large` | ~500 | ~500 |
| `mw1.xlarge` | ~1,000 | ~1,000 |
| `mw1.2xlarge` | ~1,000+ | ~2,000+ |

These are practical guidelines — actual limits depend on DAG complexity, task duration, and scheduling patterns.

---

## Common Patterns

### Development Environment

```yaml
environmentClass: mw1.micro
minWorkers: 1
maxWorkers: 2
webserverAccessMode: PUBLIC_ONLY
loggingConfiguration:
  taskLogs:
    enabled: true
    logLevel: DEBUG
```

- Cheapest configuration (~$20/month base + minimal worker costs).
- `PUBLIC_ONLY` for easy browser access without VPN.
- `DEBUG` logging for development iteration.
- Delete when not in use to save costs.

### Staging Environment

```yaml
environmentClass: mw1.small
minWorkers: 1
maxWorkers: 5
webserverAccessMode: PRIVATE_ONLY
loggingConfiguration:
  taskLogs:
    enabled: true
    logLevel: INFO
  workerLogs:
    enabled: true
    logLevel: INFO
```

- Mirrors production access pattern (`PRIVATE_ONLY`) for testing.
- Lower worker limits to control costs.
- `INFO` logging for realistic debugging without DEBUG cost.

### Production Environment

```yaml
environmentClass: mw1.medium  # or mw1.large for heavy workloads
minWorkers: 2
maxWorkers: 15
minWebservers: 2
maxWebservers: 4
schedulers: 3
webserverAccessMode: PRIVATE_ONLY
kmsKeyArn: { ... }  # customer-managed key
loggingConfiguration:
  dagProcessingLogs:
    enabled: true
    logLevel: WARNING
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

- Customer-managed KMS key for compliance.
- All 5 log modules enabled with appropriate levels.
- `GRACEFUL` worker replacement to avoid interrupting running tasks during updates.
- Maintenance window during lowest activity period.
- `minWorkers: 2` ensures capacity for burst task scheduling.

### ML Pipeline Environment

```yaml
environmentClass: mw1.large  # tasks may load ML models into memory
minWorkers: 2
maxWorkers: 10
airflowConfigurationOptions:
  core.parallelism: "32"
  core.max_active_tasks_per_dag: "8"
  core.max_active_runs_per_dag: "4"
  celery.worker_autoscale: "4,1"
```

- `mw1.large` (8 GB per worker) for ML model loading and data processing.
- Lower per-worker concurrency (`"4,1"`) because ML tasks are memory-intensive.
- Limited `max_active_runs_per_dag` to prevent resource exhaustion from parallel training runs.

### Multi-Team Shared Environment

```yaml
environmentClass: mw1.xlarge
minWorkers: 5
maxWorkers: 25
schedulers: 4
airflowConfigurationOptions:
  core.parallelism: "100"
  core.max_active_tasks_per_dag: "16"
  scheduler.parsing_processes: "8"
```

- High parallelism for multiple teams' DAGs running concurrently.
- 4 schedulers with 8 parsing processes each = 32 parallel DAG parsers.
- 25 max workers provide headroom for burst loads across teams.

---

## v2 Roadmap

Features under consideration for v2 of the AwsMwaaEnvironment component:

| Feature | Description | Priority |
|---|---|---|
| Environment tagging | Expose `tags` field for custom AWS resource tags beyond the standard labels. | High |
| CloudWatch alarm integration | Automatically create alarms for key metrics (queue depth, scheduler heartbeat, failed tasks). | High |
| DAG S3 notifications | Configure S3 event notifications to trigger environment DAG sync on upload. | Medium |
| Airflow connections as IaC | Model Airflow connections/variables as part of the spec (via startup script generation). | Medium |
| CUSTOMER endpoint management | Full support for `CUSTOMER` mode with pre-created VPC endpoint references. | Low |
| Environment cloning | Create a new environment from an existing one with selective configuration overrides. | Low |
| Backup/restore | Snapshot metadata database and restore to a new environment. | Low |
| Cross-account S3 | Documented patterns and IAM policy generation for cross-account DAG buckets. | Low |

---

## References

- [Amazon MWAA Documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/what-is-mwaa.html)
- [MWAA Networking](https://docs.aws.amazon.com/mwaa/latest/userguide/networking-about.html)
- [MWAA Environment Class](https://docs.aws.amazon.com/mwaa/latest/userguide/environment-class.html)
- [MWAA Pricing](https://aws.amazon.com/managed-workflows-for-apache-airflow/pricing/)
- [MWAA Configuration Reference](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-env-variables.html)
- [MWAA Execution Role](https://docs.aws.amazon.com/mwaa/latest/userguide/mwaa-create-role.html)
- [MWAA Security Best Practices](https://docs.aws.amazon.com/mwaa/latest/userguide/security-best-practices.html)
- [Apache Airflow Documentation](https://airflow.apache.org/docs/)
- [Airflow Configuration Reference](https://airflow.apache.org/docs/apache-airflow/stable/configurations-ref.html)

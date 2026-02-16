# GcpCloudComposerEnvironment — Research & Design Documentation

Comprehensive research document covering Cloud Composer deployment, architecture, networking models, workload sizing, security, maintenance, recovery, cost optimization, and design decisions.

---

## 1. Cloud Composer Deployment Landscape

### 1.1 What is Cloud Composer?

Cloud Composer is Google Cloud's fully managed Apache Airflow service. It provisions and manages:
- **GKE cluster** — Kubernetes cluster running Airflow components
- **Cloud SQL** — PostgreSQL database for Airflow metadata
- **Cloud Storage** — GCS bucket for DAG files
- **Airflow components** — Scheduler, workers, web server, triggerer, DAG processor

You focus on writing and deploying DAGs; Google handles infrastructure provisioning, updates, patching, and scaling.

### 1.2 Composer Versions: 1.x vs 2.x vs 3

| Aspect | Composer 1.x | Composer 2.x | Composer 3 |
|--------|--------------|--------------|------------|
| **Status** | Deprecated | Supported | Latest |
| **Airflow versions** | 1.x, 2.x | 2.x | 2.x |
| **Networking** | VPC peering only | VPC peering or PSC | PSC (default) |
| **Private endpoints** | Limited | Full support | Full support |
| **Workload sizing** | Fixed | Configurable | Configurable |
| **DAG processor** | No | No | Yes (separate component) |
| **Resilience modes** | No | STANDARD/HIGH | STANDARD/HIGH |
| **Web server plugins** | Limited | Limited | Configurable (ENABLED/DISABLED) |
| **Private builds** | No | No | Yes (enablePrivateBuildsOnly) |

**Composer 1.x is deprecated** — Google recommends migrating to Composer 2.x or 3. This component excludes all Composer 1.x-specific fields.

### 1.3 When to Choose Cloud Composer

- **Managed Airflow** — You want Airflow without managing Kubernetes, databases, or storage
- **Data pipelines** — ETL/ELT workflows, data transformations, scheduled batch jobs
- **Workflow orchestration** — Complex multi-step workflows with dependencies
- **Integration** — Native integration with GCP services (BigQuery, Cloud Storage, Pub/Sub, etc.)
- **Compliance** — Need managed service with SLAs, encryption, and audit logs

### 1.4 When to Choose Alternatives

- **Self-managed Airflow** — You need full control over infrastructure and Airflow configuration
- **Cloud Workflows** — Simple, serverless workflows without complex dependencies
- **Cloud Functions + Cloud Scheduler** — Event-driven, lightweight workflows
- **Dataflow** — Stream processing or batch processing with Apache Beam
- **Cloud Run Jobs** — Containerized batch jobs without workflow orchestration

---

## 2. Architecture

### 2.1 Component Architecture

```
Cloud Composer Environment
├── GKE Cluster
│   ├── Scheduler Pod(s)
│   ├── Worker Pod(s) (autoscaling)
│   ├── Web Server Pod
│   ├── Triggerer Pod(s)
│   └── DAG Processor Pod(s) (Composer 3 only)
├── Cloud SQL (PostgreSQL)
│   └── Airflow metadata database
└── Cloud Storage
    └── DAG bucket (gs://{region}-{project}-composer-{env})
```

### 2.2 Networking Models

**Composer 2.x — VPC Peering:**
- Composer creates a VPC peering connection between your VPC and the Composer-managed VPC
- GKE nodes run in your VPC subnet
- Cloud SQL uses a private IP in a specified CIDR range
- GKE master uses a private IP in a specified CIDR range
- Requires IP ranges for master, Cloud SQL, and Composer internal components

**Composer 2.x — Private Service Connect:**
- Uses PSC endpoints instead of VPC peering
- Avoids VPC peering quota limits (25 peerings per network)
- Better for large-scale networks
- Requires PSC connection subnetwork

**Composer 3 — Private Service Connect (default):**
- Uses `composerNetworkAttachment` (PSC network attachment)
- Requires `composerInternalIpv4CidrBlock` (/20 range)
- Supports private environments (`enablePrivateEnvironment: true`)
- Supports private builds only (`enablePrivateBuildsOnly: true`)

### 2.3 Data Flow

1. **DAG deployment** — Upload DAG files to Cloud Storage bucket
2. **DAG parsing** — DAG processor (Composer 3) or scheduler parses DAG files
3. **Task scheduling** — Scheduler creates task instances based on DAG dependencies
4. **Task execution** — Workers execute tasks (operators, sensors, etc.)
5. **Metadata storage** — Task state, logs, and metadata stored in Cloud SQL
6. **UI access** — Web server provides Airflow UI for monitoring and management

---

## 3. Networking Models Deep Dive

### 3.1 VPC Peering (Composer 2.x)

**How it works:**
- Composer creates a VPC peering connection between your VPC and Composer's managed VPC
- GKE nodes are placed in your specified subnet
- Cloud SQL and GKE master use private IPs in specified CIDR ranges

**Requirements:**
- VPC network and subnetwork in the same region
- IP ranges for:
  - GKE master (default: `172.16.0.0/28`)
  - Cloud SQL (e.g., `10.0.0.0/24`)
  - Composer internal components (e.g., `10.1.0.0/24`)

**Limitations:**
- VPC peering quota: 25 peerings per network
- Subnet IP ranges: 400 per network
- Forwarding rules: 175 per network

**Use when:**
- Small to medium networks
- Direct VPC connectivity preferred
- Composer 2.x environment

### 3.2 Private Service Connect (Composer 2.x and 3)

**How it works:**
- Uses PSC endpoints for connectivity
- Avoids VPC peering quota consumption
- Better for large-scale networks

**Composer 2.x PSC:**
- Set `privateEnvironmentConfig.connectionType: PRIVATE_SERVICE_CONNECT`
- Specify `cloudComposerConnectionSubnetwork`

**Composer 3 PSC:**
- Use `nodeConfig.composerNetworkAttachment`
- Specify `nodeConfig.composerInternalIpv4CidrBlock` (/20 range)
- Enable `enablePrivateEnvironment: true` for no public IPs

**Use when:**
- Large-scale networks (avoiding peering quotas)
- Composer 3 environments (default)
- Multi-project or multi-organization setups

### 3.3 Private Endpoints

**Composer 2.x:**
- Set `privateEnvironmentConfig.enablePrivateEndpoint: true`
- Web server accessible only via private IP
- Requires VPC peering or PSC connectivity

**Composer 3:**
- Set `enablePrivateEnvironment: true`
- No public IP endpoints for web server
- Requires PSC networking

**Use when:**
- Compliance requirements (HIPAA, PCI-DSS, FedRAMP)
- Internal-only access to Airflow UI
- Network security policies

---

## 4. Workload Sizing Guidance

### 4.1 Component Roles

| Component | Purpose | Scaling |
|-----------|---------|---------|
| **Scheduler** | Parses DAGs, manages task scheduling, triggers task instances | Fixed count (typically 1) |
| **Workers** | Execute tasks defined in DAGs | Autoscaling (min/max) |
| **Web Server** | Airflow UI for monitoring and management | Fixed (always 1) |
| **Triggerer** | Monitors deferred tasks, resumes when conditions met | Fixed count (0 to disable) |
| **DAG Processor** | Parses DAG files independently (Composer 3) | Fixed count (typically 1) |

### 4.2 Sizing Recommendations

**Small Environment (`ENVIRONMENT_SIZE_SMALL`):**
- Scheduler: 0.5 CPU, 1.5 GB memory
- Workers: 0.5 CPU, 1.5 GB memory per worker, 1-3 workers
- Web Server: 0.5 CPU, 1.5 GB memory
- Triggerer: 0.5 CPU, 1.0 GB memory, 1 replica

**Medium Environment (`ENVIRONMENT_SIZE_MEDIUM`):**
- Scheduler: 2.0 CPU, 7.5 GB memory
- Workers: 2.0 CPU, 7.5 GB memory per worker, 2-10 workers
- Web Server: 2.0 CPU, 4.0 GB memory
- Triggerer: 1.0 CPU, 2.0 GB memory, 1-2 replicas

**Large Environment (`ENVIRONMENT_SIZE_LARGE`):**
- Scheduler: 4.0 CPU, 15.0 GB memory
- Workers: 4.0 CPU, 15.0 GB memory per worker, 3-20 workers
- Web Server: 4.0 CPU, 8.0 GB memory
- Triggerer: 2.0 CPU, 4.0 GB memory, 2-4 replicas

### 4.3 Worker Autoscaling

Workers support autoscaling based on task queue depth:
- **minCount** — Minimum number of workers (prevents scale-to-zero)
- **maxCount** — Maximum number of workers (budget ceiling)

**Best practices:**
- Set `minCount: 2` for production (high availability)
- Set `maxCount` based on peak workload and budget
- Monitor task queue depth and worker utilization
- Adjust based on DAG complexity and task duration

### 4.4 Storage Sizing

Storage is used for:
- **Scheduler** — DAG parsing cache, task state
- **Workers** — Task execution logs, temporary files
- **Web Server** — UI assets, session data
- **DAG Processor** — DAG parsing cache (Composer 3)

**Recommendations:**
- Scheduler: 5-10 GB
- Workers: 10-20 GB per worker
- Web Server: 2-5 GB
- DAG Processor: 1-5 GB

---

## 5. Security Model

### 5.1 CMEK Encryption

**What's encrypted:**
- GKE node disks
- Cloud SQL database
- Cloud Storage DAG bucket

**Requirements:**
- KMS key in the same region as the Composer environment
- Composer service account must have `cloudkms.cryptoKeyEncrypterDecrypter` permission
- Key must be in the same project or shared via IAM

**Use when:**
- Compliance requirements (HIPAA, PCI-DSS, FedRAMP)
- Customer-managed encryption keys required
- Regulatory mandates

### 5.2 Private Endpoints

**Composer 2.x:**
- `privateEnvironmentConfig.enablePrivateEndpoint: true`
- Web server accessible only via private IP
- Requires VPC peering or PSC

**Composer 3:**
- `enablePrivateEnvironment: true`
- No public IP endpoints
- Requires PSC networking

**Use when:**
- Internal-only access required
- Network security policies
- Compliance requirements

### 5.3 IP Allowlisting

Restrict web server access to specific IP ranges:
- Configure `webServerNetworkAccessControl.allowedIpRanges`
- CIDR notation (e.g., `10.0.0.0/8`, `203.0.113.0/24`)
- Optional descriptions for each range

**Use when:**
- Corporate network access only
- VPN or bastion host access
- Additional security layer

### 5.4 Service Account

Specify a custom service account for GKE nodes:
- Default: Compute Engine default service account
- Custom: Service account with least-privilege IAM roles
- Required permissions: Cloud SQL client, Cloud Storage access, etc.

---

## 6. Maintenance and Recovery

### 6.1 Maintenance Windows

Define when GCP may perform scheduled maintenance:
- **Start/end time** — RFC3339 format (e.g., `2026-01-01T02:00:00Z`)
- **Recurrence** — RFC5545 RRULE format (e.g., `FREQ=WEEKLY;BYDAY=TU,WE,TH`)
- **Minimum duration** — 12 hours

**Best practices:**
- Schedule during low-traffic periods
- Use weekly recurrence for predictable windows
- Avoid weekends if possible (for on-call teams)

### 6.2 Recovery Configuration

Enable scheduled snapshots for disaster recovery:
- **Snapshot location** — Cloud Storage bucket folder URI
- **Schedule** — Unix-cron format (e.g., `0 4 * * *` for daily at 4 AM)
- **Time zone** — IANA time zone (e.g., `America/Los_Angeles`)

**What's backed up:**
- Cloud SQL database (Airflow metadata)
- DAG files (already in Cloud Storage)
- Airflow configuration

**Use when:**
- Disaster recovery requirements
- Compliance mandates
- Point-in-time recovery needs

---

## 7. Cost Optimization Strategies

### 7.1 Environment Size

| Size | Approximate Cost | Use Case |
|------|------------------|----------|
| SMALL | ~$300/month | Development, testing |
| MEDIUM | ~$1,200/month | Small to medium production |
| LARGE | ~$4,800/month | Large-scale production |

### 7.2 Worker Autoscaling

- **Scale down during low-traffic periods** — Set appropriate `minCount` to avoid over-provisioning
- **Scale up for peak workloads** — Set `maxCount` based on peak demand
- **Monitor queue depth** — Adjust min/max based on actual usage

### 7.3 Storage Optimization

- **Right-size storage** — Don't over-provision storage for workers
- **Clean up logs** — Configure log retention policies
- **Archive old DAGs** — Move unused DAGs to cold storage

### 7.4 Network Costs

- **Use same-region resources** — Minimize cross-region network egress
- **Private networking** — Reduces public IP costs (if applicable)
- **VPC peering vs PSC** — PSC may have lower costs for large networks

### 7.5 Development vs Production

- **Use SMALL size for dev** — Reduces costs for non-production environments
- **Separate environments** — Isolate dev/test from production
- **Schedule-based scaling** — Scale down dev environments during off-hours (if supported)

---

## 8. 80/20 Scoping Rationale

### 8.1 What We Cover

| Feature | Included | Rationale |
|---------|----------|-----------|
| Composer 2.x and 3 | Yes | Current supported versions |
| VPC peering networking | Yes | Traditional networking model |
| Private Service Connect | Yes | Modern networking model, required for Composer 3 |
| Workload sizing | Yes | Critical for performance and cost |
| Software configuration | Yes | Image versions, packages, config overrides |
| Private networking | Yes | Security and compliance requirements |
| CMEK encryption | Yes | Enterprise compliance |
| Maintenance windows | Yes | Operational control |
| Recovery snapshots | Yes | Disaster recovery |
| Access control | Yes | Security requirements |
| Environment sizes | Yes | Capacity management |
| Resilience modes | Yes | High availability |

### 8.2 What We Exclude

| Feature | Excluded | Rationale |
|---------|----------|-----------|
| Composer 1.x fields | Yes | Deprecated by Google |
| DAG management | Yes | Application-level concern; managed via Cloud Storage |
| Airflow connections | Yes | Application-level configuration |
| Airflow variables | Yes | Application-level configuration |
| Custom Airflow plugins | Yes | Application-level concern |
| Monitoring/alerts | Yes | Managed via Cloud Monitoring |
| IAM bindings | Yes | Managed via GCP IAM |
| Cloud SQL instance details | Yes | Managed by Composer |
| Cloud Storage bucket details | Yes | Managed by Composer |
| GKE cluster details | Yes | Managed by Composer |

### 8.3 Deliberate Design Choices

**Composer 2.x and 3 only:** Composer 1.x is deprecated. Excluding Composer 1.x fields simplifies the component and avoids supporting deprecated features.

**Networking flexibility:** Support both VPC peering (Composer 2.x) and PSC (Composer 2.x and 3) to cover all networking models.

**Workload sizing:** Configurable workload sizing is critical for performance and cost optimization. Defaults are provided via `environmentSize`, but fine-grained control is available via `workloadsConfig`.

**Private networking:** Both Composer 2.x (`privateEnvironmentConfig`) and Composer 3 (`enablePrivateEnvironment`) are supported to cover all private networking scenarios.

**No DAG management:** DAGs are application-level concerns. They're deployed to Cloud Storage via CI/CD pipelines or manual uploads, not via infrastructure-as-code.

**No Airflow connections/variables:** These are application-level configuration managed via Airflow UI or API, not infrastructure configuration.

---

## 9. Immutable Fields

The following fields cannot be changed after creation; changing them requires recreating the environment:

| Field | Scope | Notes |
|-------|-------|-------|
| `region` | Environment | GCP region placement |
| `nodeConfig.network` | Environment | VPC network (Composer 2.x VPC peering) |
| `nodeConfig.subnetwork` | Environment | VPC subnetwork (Composer 2.x VPC peering) |
| `nodeConfig.composerNetworkAttachment` | Environment | PSC network attachment (Composer 3) |
| `privateEnvironmentConfig.connectionType` | Environment | VPC_PEERING vs PRIVATE_SERVICE_CONNECT |
| `kmsKeyName` | Environment | CMEK encryption key |

---

## 10. Best Practices for Production Deployments

1. **Use HIGH_RESILIENCE mode** — Multi-zone redundancy for increased availability (Composer 2.1.15+)
2. **Enable private endpoints** — Restrict web server access to private IPs
3. **Use CMEK encryption** — Customer-managed encryption keys for compliance
4. **Configure maintenance windows** — Schedule maintenance during low-traffic periods
5. **Enable recovery snapshots** — Scheduled snapshots for disaster recovery
6. **Right-size workloads** — Configure CPU, memory, and storage based on actual usage
7. **Set worker autoscaling** — Min/max counts based on workload patterns
8. **Use IP allowlisting** — Restrict web server access to corporate networks
9. **Monitor metrics** — Track scheduler/worker utilization, task queue depth, DAG parsing time
10. **Separate environments** — Isolate dev/test from production

---

## 11. Monitoring and Observability

### 11.1 Key Metrics

| Metric | Source | Alert Threshold |
|--------|--------|-----------------|
| Scheduler CPU utilization | Cloud Monitoring | > 80% sustained |
| Worker CPU utilization | Cloud Monitoring | > 80% sustained |
| Task queue depth | Cloud Monitoring | > 100 pending tasks |
| DAG parsing time | Cloud Monitoring | > 30s per DAG |
| Failed task rate | Cloud Monitoring | > 5% |
| Web server latency | Cloud Monitoring | > 1s p99 |

### 11.2 Dashboards

GCP provides built-in Composer monitoring dashboards in Cloud Console. For custom dashboards, use Cloud Monitoring with the `composer.googleapis.com` metric prefix.

---

## 12. Troubleshooting Common Issues

### 12.1 Environment Creation Failures

- **Service account permissions** — Ensure Composer service account has required IAM roles
- **Network configuration** — Verify VPC/subnetwork exists and IP ranges don't conflict
- **Quota limits** — Check VPC peering, subnet IP ranges, forwarding rules quotas
- **Organization policies** — Ensure no conflicting Org Policies

### 12.2 DAG Parsing Errors

- **Syntax errors** — Check DAG Python syntax
- **Import errors** — Ensure PyPI packages are installed
- **Configuration errors** — Verify Airflow config overrides

### 12.3 Task Execution Failures

- **Resource limits** — Increase worker CPU/memory if tasks are OOMKilled
- **Network connectivity** — Verify network access to external services
- **Service account permissions** — Ensure workers have required IAM permissions

---

## 13. Migration from Composer 1.x

If migrating from Composer 1.x:

1. **Export DAGs** — Download DAGs from Cloud Storage
2. **Export connections/variables** — Export via Airflow UI or API
3. **Create Composer 2.x/3 environment** — Use this component
4. **Import DAGs** — Upload to new Cloud Storage bucket
5. **Import connections/variables** — Import via Airflow UI or API
6. **Verify functionality** — Test DAGs in new environment
7. **Decommission Composer 1.x** — Delete old environment after migration

---

## 14. Integration Patterns

### Pattern 1: CI/CD DAG Deployment

```
GitHub Actions / Cloud Build
  → Build DAG files
  → Upload to Cloud Storage DAG bucket
  → Trigger Airflow DAG refresh
```

### Pattern 2: Multi-Environment Setup

```
Development Environment (SMALL)
  → Test DAGs
Production Environment (LARGE, HIGH_RESILIENCE)
  → Run production workloads
```

### Pattern 3: Hybrid Networking

```
Composer 2.x with VPC Peering
  → Direct VPC connectivity
  → Private endpoints
  → IP allowlisting
```

---

This research document provides comprehensive guidance for deploying and managing Cloud Composer environments. For component-specific details, see the [main README](../README.md) and [examples](../examples.md).

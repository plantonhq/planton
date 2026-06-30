# Deploying Kubernetes CronJobs: From Manual YAML to Production-Ready Automation

## The Scheduling Problem That Everyone Solves Differently

Every production system needs to run scheduled tasks: database backups at midnight, cache warming every five minutes, report generation at the start of each month, cleanup jobs at week's end. The Unix `cron` daemon solved this problem forty years ago, yet in the Kubernetes era, teams struggle with the same question: *what's the right way to deploy and manage scheduled workloads?*

The Kubernetes CronJob resource, stable since v1.21, provides a native, controller-based solution that mirrors the simplicity of traditional crontab while leveraging Kubernetes' declarative model. Yet the path from a simple `kubectl apply` to a production-ready, GitOps-managed, observable, and failure-resilient scheduling system is far from obvious. The native API's defaults are optimized for simplicity, not safety. Critical configuration fields are buried in deep nesting. Common deployment patterns can lead to catastrophic failure modes—from "thundering herd" database overloads to silent job failures with no trace evidence.

This document maps the complete deployment landscape for Kubernetes CronJobs, from anti-patterns to production-proven approaches, and explains how Planton's `CronJobKubernetes` resource provides an opinionated abstraction that makes production-ready scheduling the path of least resistance.

---

## The Maturity Spectrum: Five Levels of CronJob Deployment

### Level 0: The Anti-Pattern — Manual kubectl with No Reconciliation

**What it is:** An engineer writes a CronJob YAML manifest and runs `kubectl apply -f cronjob.yaml`. The job is scheduled. The engineer moves on.

**What it solves:** Gets a job running quickly for local development or one-off experimentation.

**What it doesn't solve:**
- **Configuration drift**: If anyone edits the CronJob in the cluster (`kubectl edit`), the state permanently diverges from the local file. There's no mechanism for detecting or preventing this drift.
- **No audit trail**: No record of who deployed what, when, or why.
- **Environment inconsistency**: Managing dev, staging, and production requires manually maintaining and applying separate YAML files.
- **No rollback**: If the new schedule breaks, there's no easy way to revert to the previous configuration.

**The critical flaw:** `kubectl apply` is an *imperative action*, not a *declarative system*. It pushes state once but provides no continuous reconciliation. For production infrastructure, this is fundamentally unreliable.

**Verdict:** Acceptable for learning and local testing. **Never use this for production deployments.**

---

### Level 1: The First Step Up — Templating with Helm or Kustomize

**What it is:** Using templating tools to manage CronJob configurations across multiple environments.

**Helm approach:** Package the CronJob (and related resources like Secrets and ConfigMaps) into a versioned Chart. Use `values.yaml` files to parameterize settings like `schedule`, `image.tag`, and `resources.limits`. Deploy with `helm install` or `helm upgrade`.

```yaml
# values.yaml
schedule: "*/5 * * * *"
image:
  repository: myapp/cache-warmer
  tag: v1.2.3
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 100m
    memory: 128Mi
```

**Kustomize approach:** Define a base CronJob manifest in a `base/` directory. Create environment-specific `overlays/` (e.g., `overlays/prod/`) that patch the base with production-specific changes (different schedule, higher resource limits). Apply with `kubectl apply -k overlays/prod/`.

**What it solves:**
- Multi-environment management through parameterization
- Versioned, repeatable deployments
- Reduced YAML duplication
- Better configuration organization
- Helm provides native rollback capability

**What it doesn't solve:**
- Still relies on *push-based* deployment: a human or CI pipeline must remember to run `helm upgrade` or `kubectl apply -k`
- No built-in drift detection or automatic remediation
- Limited audit trail (unless the CI system provides it)

**Verdict:** A solid foundation for teams with manual or CI-driven deployment pipelines. Still lacks continuous reconciliation.

---

### Level 2: The Paradigm Shift — Declarative GitOps

**What it is:** Using Git as the single source of truth for cluster state. In-cluster controllers (ArgoCD or Flux) continuously monitor a Git repository and automatically sync changes to the cluster.

**How it works:**
1. CronJob manifests (raw YAML, Helm charts, or Kustomize overlays) live in a Git repository
2. ArgoCD or Flux is deployed in the cluster and configured to watch specific Git paths
3. When a developer pushes a commit changing the CronJob's schedule, the GitOps controller detects the diff and applies it automatically
4. If someone manually edits the CronJob in the cluster, the controller detects drift and can auto-remediate by reverting to Git state

**What this transforms:**
- **Changes become pull requests**: Modifying a schedule is no longer a `kubectl` command run by one person; it's a code-reviewed PR with full context and approval
- **Perfect audit trail**: Every change is a Git commit with author, timestamp, and rationale
- **Automatic rollback**: Reverting a change is `git revert` followed by automatic redeployment
- **Multi-cluster consistency**: The same Git repo can be the source of truth for dev, staging, and prod clusters

**ArgoCD vs Flux:**
- **ArgoCD**: Application-centric with a powerful web UI. Ideal for teams that want visual representation of application state and drift detection.
- **Flux**: Toolkit-based and CLI-first. Preferred in organizations with strong CLI/automation culture.

**What it doesn't solve:**
- Still requires understanding of Kubernetes CronJob configuration
- Doesn't prevent you from deploying jobs with unsafe defaults (e.g., missing resource limits, wrong concurrency policies)
- Doesn't abstract away the API's complexity

**Verdict:** The gold standard for production Kubernetes deployments. **Every production CronJob should be deployed via GitOps.**

---

### Level 3: The Infrastructure-as-Code Integration — Unified Control Planes

**What it is:** Managing CronJobs alongside the broader infrastructure (databases, IAM roles, cloud storage, networking) using IaC tools.

**Use cases:**
- **Terraform**: Uses HCL and the `kubernetes` provider or `helm` provider to define CronJobs. Manages state in a backend (S3, GCS, Terraform Cloud). Changes are applied via `terraform apply`.
- **Pulumi**: Uses general-purpose programming languages (Python, Go, TypeScript) to define infrastructure. The CronJob definition is just code, allowing for loops, conditionals, and abstraction.
- **Crossplane**: A Kubernetes-native IaC tool that turns the cluster itself into a control plane for external resources. Defines cloud resources (RDS databases, S3 buckets) and Kubernetes resources (CronJobs) as CRDs in the same cluster.

**When this matters:**
- You're deploying a backup CronJob that needs to interact with an RDS database, write to an S3 bucket, and use IAM roles for authentication. Managing these as a *single, atomic stack* in Terraform or Pulumi is far more reliable than managing them separately.
- Your organization has standardized on Terraform or Pulumi for all infrastructure, and treating Kubernetes workloads as just another resource simplifies operations.

**The architectural divide: Push vs. Pull**
- **Terraform/Pulumi/Planton**: *Push-based*. An external system (CI pipeline, developer laptop) runs a command to push state to the cluster.
- **GitOps/Crossplane**: *Pull-based*. An in-cluster controller pulls desired state from a source (Git, CRDs) and continuously reconciles.

**The ideal pattern:** Use IaC tools to *generate* GitOps artifacts. For example, Planton or Pulumi can generate Kubernetes manifests that are then committed to a Git repository and deployed by ArgoCD. This combines the benefits of programmatic infrastructure definition with continuous reconciliation.

**Verdict:** Essential for teams managing both infrastructure and Kubernetes workloads. Best combined with GitOps for the actual deployment.

---

### Level 4: The Defensive API — Planton's Production-Safe Abstraction

**What it is:** Planton's `CronJobKubernetes` resource provides an opinionated API that **flattens the 80% of commonly used fields** and **enforces safe-by-default production settings**.

**The core problem with the native API:**
The native Kubernetes CronJob API has several fatal flaws for production use:
1. **Unsafe defaults**: `concurrencyPolicy: Allow` (the default) allows multiple jobs to run simultaneously, causing race conditions for stateful tasks. The default should be `Forbid`.
2. **Buried critical fields**: Essential settings like `startingDeadlineSeconds` (prevents "thundering herd" failures) are nested deep in `spec.` and easy to miss.
3. **No resource enforcement**: Jobs without `resources.limits` can exhaust node resources and crash other services. The API doesn't prevent this.
4. **Ambiguous time zones**: Historically, CronJobs used the controller-manager's local time zone, causing jobs to run at incorrect times.
5. **Silent failures**: The default `failedJobsHistoryLimit: 1` means failed job Pods are quickly garbage-collected, erasing all evidence for debugging.

**How Planton solves this:**

**1. Safe defaults that prevent common disasters:**
- `concurrencyPolicy: Forbid` (prevents concurrent runs)
- `timeZone: "Etc/UTC"` (explicit, portable schedules)
- `startingDeadlineSeconds: 600` (prevents job pileups after cluster downtime)
- `failedJobsHistoryLimit: 3` (keeps failed Pods for post-mortem analysis)
- `restartPolicy: Never` (clearer failure signals)
- `backoffLimit: 3` (finite, sensible retry limit)

**2. Flattened, 80/20 API:**
The native API requires deeply nested configuration:
```yaml
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - image: postgres:15
              resources:
                requests:
                  cpu: "1"
```

Planton flattens this:
```yaml
schedule: "0 0 * * *"
image: "postgres:15"
resources:
  requests:
    cpu: "1"
    memory: "2Gi"
```

**3. Mandated critical fields:**
The API *fails validation* if `resources` (requests and limits) are not provided. This single constraint prevents the most dangerous anti-pattern: resource-unbounded "noisy neighbor" jobs.

**4. Secure-by-default secret handling:**
Instead of encouraging secrets as environment variables (which leak into logs and child processes), Planton provides a top-level `secret_volumes` field that mounts secrets as read-only files—the secure pattern.

**Verdict:** This is the recommended approach for teams using Planton. It codifies production best practices into the API itself, preventing entire categories of operational failures before they happen.

---

## Production-Ready Configuration: What Really Matters

Based on analysis of real-world CronJob deployments, here are the configuration decisions that define production-readiness:

### The Concurrency Control Decision

| Policy | Behavior | When to Use | Critical Pitfall |
|--------|----------|-------------|------------------|
| **Allow** | Allows concurrent job runs | **Rare**: Only for truly stateless, idempotent tasks like HTTP pings | **This is the dangerous default**. Causes race conditions for any stateful work (backups, data processing). |
| **Forbid** | Skips new run if previous is still active | **90% of use cases**: Database backups, report generation, cleanup jobs | Subject to a rare controller-manager restart race condition that can cause duplicate runs. Application-level idempotency is mandatory. |
| **Replace** | Kills current job and starts a new one | **"Latest-only" tasks**: Fetching a config file where only the newest version matters | Resource-intensive if jobs are long-running. |

**The Forbid pitfall:** Even with `concurrencyPolicy: Forbid`, there's a documented race condition where a controller-manager restart can lose the "job is running" state in memory and launch a duplicate job. This has caused real-world financial losses. **Your jobs must be designed to be idempotent at the application level.**

### The "Thundering Herd" Prevention: startingDeadlineSeconds

**The failure scenario:**
1. A cluster is down for 3 hours for emergency maintenance
2. A CronJob scheduled to run every 5 minutes has 36 "missed" schedules
3. When the controller-manager restarts, it sees all 36 missed schedules
4. By default, it attempts to run **all 36 jobs simultaneously**
5. This "thundering herd" overwhelms the target database, causing a cascading failure

**The solution:**
```yaml
startingDeadlineSeconds: 600  # 10 minutes
```

The controller evaluates each missed job. If the current time is more than 600 seconds past the scheduled time, it skips the job. Only jobs within the valid window are run. This prevents the thundering herd.

**Planton default:** `600` (10 minutes). The native Kubernetes default is `null` (no limit), which is unsafe.

### The Debugging Requirement: failedJobsHistoryLimit

**The anti-pattern:**
Setting `failedJobsHistoryLimit: 0` (or leaving it at the default of `1`) means failed Pods are immediately (or quickly) garbage-collected. When a job fails at 3 AM and pages an engineer, there are:
- No Pods to `kubectl describe`
- No logs to `kubectl logs`
- No evidence to debug

It's a "ghost failure."

**The solution:**
```yaml
failedJobsHistoryLimit: 3
successfulJobsHistoryLimit: 1
```

Keep 3 failed job histories for debugging. Keep only 1 successful history to save cluster resources (you rarely need to debug success).

**Planton defaults:** `3` for failed, `1` for successful.

### The "Noisy Neighbor" Prevention: Resource Limits

**The anti-pattern:**
Deploying a CronJob without `resources.requests` or `resources.limits`. The Pod is scheduled with `BestEffort` QoS. During execution (e.g., a memory-intensive data processing task), it consumes all available memory on the node. The Kubelet begins OOM-killing other Pods, including mission-critical services.

**The solution:**
```yaml
resources:
  requests:
    cpu: "1"
    memory: "2Gi"
  limits:
    cpu: "2"
    memory: "4Gi"
```

This ensures:
- The Pod is scheduled only on nodes with sufficient resources (`requests`)
- The Pod is hard-capped and cannot exceed limits, protecting other workloads (`limits`)

**Planton requirement:** The API *fails validation* if resources are not provided. This forces users into safe behavior.

### The Secret Security Pattern: Volumes vs Environment Variables

**The insecure pattern:**
```yaml
env:
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-secret
        key: password
```

**Why this is dangerous:**
- Applications often log their full environment on startup: "Environment: DB_HOST=..., DB_PASSWORD=secret123, ..."
- Environment variables are inherited by all child processes (shell scripts, CLIs). A simple `curl` command can leak the secret to logs.
- Environment variables are visible in `kubectl describe pod` output.

**The secure pattern:**
```yaml
volumes:
  - name: secrets
    secret:
      secretName: db-secret
volumeMounts:
  - name: secrets
    mountPath: /etc/secrets
    readOnly: true
```

The application explicitly reads `/etc/secrets/password`. The secret is:
- Not logged automatically
- Not inherited by child processes
- Not easily inspectable
- Mounted on an in-memory tmpfs (never written to disk)

**Planton API:** Provides a top-level `secret_volumes` field to make this the easy path.

### The Time Zone Clarity: timeZone

**The historical problem:**
Before Kubernetes 1.27, CronJobs used the time zone of the kube-controller-manager process. This was ambiguous and varied by cluster configuration, causing jobs to run at incorrect times.

**The solution (stable in 1.27+):**
```yaml
timeZone: "Etc/UTC"  # or "America/New_York", "Asia/Tokyo"
```

This makes the schedule explicit and portable.

**Planton default:** `"Etc/UTC"`. Explicit, unambiguous, and the de facto standard for distributed systems.

---

## Common Anti-Patterns and How to Avoid Them

| Anti-Pattern | Failure Mode | Prevention |
|--------------|--------------|------------|
| **Missing `startingDeadlineSeconds`** | After cluster downtime, all missed schedules run at once, overwhelming target systems | Always set to 2-3x the schedule interval (e.g., 600 for a 5-minute job) |
| **Jobs without resource limits** | Batch job consumes all node memory/CPU, causing OOM kills of other services | Mandate resources in API validation. Set both requests and limits. |
| **`concurrencyPolicy: Allow` for stateful jobs** | Multiple backup jobs run simultaneously, causing corruption or locking issues | Use `Forbid` for all stateful tasks. Design jobs to be idempotent. |
| **Secrets as environment variables** | Credentials leak into application logs, child process environments, and kubectl output | Mount secrets as read-only volumes at `/etc/secrets` |
| **`failedJobsHistoryLimit: 0`** | Failed Pods are immediately deleted; no logs or evidence for debugging | Set to at least 3 to retain debugging information |
| **Over-reliance on `concurrencyPolicy: Forbid`** | Rare controller restart causes race condition and duplicate jobs | Design application logic to be idempotent; don't trust Forbid as a guarantee |
| **Missing or infinite `backoffLimit`** | Misconfigured job (typo in command, wrong image) enters infinite retry loop | Set finite limit (e.g., 3) to fail fast and alert |

---

## When NOT to Use Kubernetes CronJobs

The native CronJob resource is powerful but not universal. Here's when to use alternatives:

### Use Argo Workflows or Tekton Instead
**When:** You have a multi-step pipeline with dependencies (a Directed Acyclic Graph).

**Example:** "Run Task A (extract data), and if it succeeds, run Task B (transform) and Task C (load) in parallel."

**Why CronJobs fail here:** A CronJob executes a single task in a single Pod. Orchestrating 50 separate CronJobs with staggered schedules to simulate dependencies is brittle and unmaintainable. One team replaced 50 CronJobs with a single Argo Workflow and saw significant reliability improvements.

### Use AWS EventBridge or Google Cloud Scheduler Instead
**When:** The task does *not* need access to resources inside the Kubernetes cluster's private network.

**Example:** "Call a public API every 5 minutes" or "Trigger a Lambda function hourly."

**Why CronJobs are overkill:** If the task is just an HTTP call, a serverless cloud scheduler is simpler, cheaper, and more reliable. Use CronJobs *only* for tasks that must run inside the cluster (e.g., querying a database on the cluster's private network).

### Use Apache Airflow Instead
**When:** You're building data engineering pipelines (ETL/ELT) that need backfilling, complex dependencies, and a data-centric UI.

**Example:** "Extract from 3 databases, transform with 5 steps, load to a data warehouse, and if any step fails, send a Slack alert and retry with exponential backoff."

**Why CronJobs fail here:** CronJobs are for infrastructure operations (backups, cleanups, log rotation). Airflow is for data pipelines. While Airflow can run on Kubernetes (using `KubernetesPodOperator`), it's a heavyweight system built for a different persona (Data Engineers, not SREs).

| Tool | Use When | Kubernetes-Native | Key Strength |
|------|----------|-------------------|--------------|
| **Kubernetes CronJob** | Simple, recurring infrastructure tasks | ✅ Yes | Simple, reliable, native |
| **Argo Workflows** | CI/CD, multi-step jobs with dependencies | ✅ Yes | Powerful DAG execution |
| **Tekton** | CI/CD pipelines | ✅ Yes | Composable, CI-focused |
| **Cloud Schedulers** | Simple tasks not needing cluster access | ❌ No | Extremely cheap, simple, reliable |
| **Apache Airflow** | Data engineering pipelines | Runs *on* K8s | Backfilling, complex dependencies |

---

## Observability: Monitoring and Alerting for CronJobs

### The Logging Challenge
CronJob Pods are ephemeral. Relying on `kubectl logs` is a losing strategy—the Pod will be gone (especially after a failure). **All logs must be shipped to a central system** (Loki, ElasticSearch, CloudWatch) by a node-level agent (FluentBit, Vector).

This connects directly to `failedJobsHistoryLimit`. If set to `0`, the Pod is deleted immediately upon failure. The log-scraping agent may not have had time to collect the logs. Setting `failedJobsHistoryLimit: 3` keeps the failed Pod around, giving the agent time to scrape logs and ensuring you have debugging data.

### The Metrics Strategy
There are two categories of metrics:

**1. Job Metadata (via kube-state-metrics)**

The `kube-state-metrics` service watches the Kubernetes API and exposes Prometheus metrics about object state:

- `kube_job_status_failed > 0`: Fires immediately if any job fails
- `time() - kube_cronjob_next_schedule_time > 3600`: Fires if the next scheduled run is in the past (controller is failing)
- `time() - max(kube_job_status_start_time{...status="Succeeded"}) > 7200`: Fires if a job hasn't successfully completed in its expected interval (for an hourly job, alert if no success in 2 hours)

**2. Application Metrics (via Prometheus PushGateway)**

Prometheus is a *pull-based* system and cannot reliably scrape ephemeral CronJob Pods. The solution:
1. Deploy the Prometheus PushGateway as a persistent service in the cluster (a "metrics mailbox")
2. The CronJob, as its last step, pushes custom metrics (e.g., `rows_processed=1000`, `backup_size_gb=50`) to the PushGateway's HTTP endpoint
3. Prometheus scrapes the (persistent) PushGateway, not the (ephemeral) job Pod

This is the only robust pattern for collecting custom application metrics from batch workloads.

---

## How Planton Makes This Simple

The `CronJobKubernetes` resource in Planton is designed to encode all of these production lessons into the API itself.

### Example 1: Simple Scheduled Task (Cache Warmer)

```yaml
kind: CronJobKubernetes
metadata:
  name: cache-warmer
spec:
  schedule: "*/5 * * * *"  # Every 5 minutes
  image: "alpine:3.18"
  command: ["/bin/sh", "-c"]
  args:
    - "curl -s http://my-service.prod.svc.cluster.local/api/v1/warm-cache"
  resources:
    requests:
      cpu: "50m"
      memory: "64Mi"
    limits:
      cpu: "100m"
      memory: "128Mi"
  # All policies use safe defaults:
  # concurrencyPolicy: Forbid
  # timeZone: Etc/UTC
  # startingDeadlineSeconds: 600
  # failedJobsHistoryLimit: 3
```

### Example 2: Database Backup (Resource-Intensive, Secure Secrets)

```yaml
kind: CronJobKubernetes
metadata:
  name: postgres-backup
spec:
  schedule: "0 0 * * *"  # Midnight daily
  image: "postgres:15"
  command: ["/bin/sh", "-c"]
  args:
    - |
      export PGPASSWORD=$(cat /etc/secrets/db-password)
      pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > /backups/backup-$(date +%F).sql
  resources:
    requests:
      cpu: "1"      # 1 vCPU
      memory: "2Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  config_map_names:
    - "db-backup-config"  # Provides DB_HOST, DB_USER, DB_NAME as env vars
  secret_volumes:
    - secret_name: "db-backup-password"
      mount_path: "/etc/secrets"  # Mounts db-password key to /etc/secrets/db-password
  pvc_mounts:
    - pvc_name: "db-backup-pvc"
      mount_path: "/backups"
  policy:
    timezone: "America/New_York"  # Job runs at midnight New York time
    concurrency_policy: "Forbid"  # Ensure only one backup runs at a time
    retry_limit: 2
    history_limit_failed: 5  # Keep more history for critical jobs
```

### What's Different from Raw Kubernetes YAML?

**1. Validation prevents disasters:**
- If you forget `resources`, the API rejects it. You cannot deploy a "noisy neighbor" job.

**2. Safe defaults prevent common failures:**
- You don't need to remember `startingDeadlineSeconds`. It's set by default.
- `concurrencyPolicy: Forbid` is the default, not `Allow`.
- Time zone is explicit (`Etc/UTC`), not ambiguous.

**3. Security is the easy path:**
- `secret_volumes` is a first-class field. Mounting secrets securely is simpler than using environment variables.

**4. The API is learnable:**
- Essential fields are at the top level (`schedule`, `image`, `resources`)
- Advanced fields are organized under `advanced_scheduling` and `advanced_container`
- You only go deeper when you actually need the complexity

---

## Namespace Management: Isolation vs. Multi-tenancy

One of the critical architectural decisions for Kubernetes CronJobs is namespace management: should each CronJob get its own dedicated namespace, or should multiple CronJobs share a common namespace? This decision impacts resource isolation, RBAC boundaries, operational complexity, and cleanup patterns.

### The Two Patterns

#### Pattern 1: Dedicated Namespace (Isolation)

Each CronJob is deployed into its own namespace, created and managed by the CronJob deployment.

**Configuration:**
```yaml
spec:
  namespace:
    value: my-backup-job
  create_namespace: true
```

**Characteristics:**
- **Isolation**: The CronJob has complete namespace isolation
- **Clean boundaries**: RBAC policies can be scoped to just this namespace
- **Easy cleanup**: Deleting the namespace removes all related resources
- **Resource quotas**: Can apply namespace-level quotas without affecting other workloads

**When to use:**
- CronJobs with sensitive operations requiring RBAC isolation
- Jobs that benefit from their own resource quota
- Development and testing environments where quick cleanup is important
- Jobs that are logically independent and have distinct lifecycles

**Trade-offs:**
- More namespaces to manage and monitor
- Higher operational overhead (each namespace appears in lists, dashboards)
- Potential for namespace sprawl in large organizations

#### Pattern 2: Shared Namespace (Multi-tenancy)

Multiple CronJobs are deployed into a common namespace, pre-created and managed separately.

**Configuration:**
```yaml
spec:
  namespace:
    value: batch-jobs
  create_namespace: false
```

**Characteristics:**
- **Consolidation**: All batch jobs grouped logically in one namespace
- **Simplified monitoring**: Single namespace to watch for job status
- **Shared quotas**: All jobs share the namespace's resource quota
- **Centralized RBAC**: One set of RBAC rules applies to all jobs

**When to use:**
- Many related CronJobs (e.g., all nightly batch processing)
- GitOps workflows where namespaces are managed separately
- Organizations with strict namespace governance
- Jobs that share common ConfigMaps or Secrets

**Trade-offs:**
- Less isolation (all jobs see each other)
- Shared resource quotas can lead to contention
- Cleanup requires careful labeling to avoid removing other jobs' resources
- RBAC becomes more complex (need to grant permissions selectively)

### Design Rationale: The `create_namespace` Flag

The `create_namespace` boolean flag provides explicit control over namespace lifecycle:

**`create_namespace: true` (Default for new deployments):**
- The deployment module creates and manages the namespace
- Namespace gets appropriate labels matching the CronJob metadata
- Destroying the CronJob also destroys the namespace
- Enables the "cattle" approach: treat each CronJob as disposable

**`create_namespace: false` (For shared namespaces):**
- The deployment module references an existing namespace
- Namespace must exist before deployment (deployment fails otherwise)
- Destroying the CronJob leaves the namespace intact
- Enables the "pets" approach: careful curation of a stable namespace

### Best Practices

#### For Isolated CronJobs

1. **Use descriptive namespace names**: Name the namespace after the job's purpose (e.g., `db-backup-prod`, `cache-warmer-staging`)
2. **Apply resource quotas**: Prevent runaway jobs from exhausting cluster resources
3. **Label consistently**: Use consistent labeling (e.g., `app`, `team`, `environment`) for monitoring and cost allocation
4. **Document ownership**: Add annotations indicating the team or service owning the namespace

Example:
```yaml
metadata:
  name: monthly-report-generator
spec:
  namespace:
    value: monthly-reports
  create_namespace: true
```

#### For Shared Namespaces

1. **Pre-create with governance**: Create the namespace separately using a centralized IaC module with appropriate RBAC and quotas
2. **Use strict labeling**: Every resource (CronJob, Secret, ConfigMap) must have labels identifying its owning job
3. **Monitor namespace-level metrics**: Watch for resource contention and quota exhaustion
4. **Implement cleanup policies**: Use Kubernetes TTL controllers or external tools to clean up completed jobs

Example namespace setup:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: batch-jobs
  labels:
    purpose: scheduled-batch
    team: platform
  annotations:
    scheduler: cron
```

Example CronJob referencing it:
```yaml
metadata:
  name: nightly-aggregation
  labels:
    job-type: aggregation
    schedule: nightly
spec:
  namespace:
    value: batch-jobs
  create_namespace: false
```

### GitOps Considerations

In GitOps workflows (ArgoCD, Flux), namespace management follows a specific pattern:

**Recommended approach:**
1. Create a separate Git repository or directory for namespace definitions
2. Deploy namespaces first using a dedicated ArgoCD Application or Flux Kustomization
3. Set CronJob `create_namespace: false` and reference the pre-created namespaces
4. This ensures namespaces are created in the correct order and can be managed independently

**Directory structure:**
```
gitops-repo/
  infrastructure/
    namespaces/
      batch-jobs-namespace.yaml
      monitoring-namespace.yaml
  applications/
    cronjobs/
      nightly-backup.yaml  # create_namespace: false
      weekly-cleanup.yaml  # create_namespace: false
```

**ArgoCD Application ordering:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: namespaces
spec:
  # Deployed first
  source:
    path: infrastructure/namespaces
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cronjobs
spec:
  # Deployed after namespaces
  source:
    path: applications/cronjobs
  syncPolicy:
    syncOptions:
      - CreateNamespace=false  # Fail if namespace doesn't exist
```

### Security and RBAC Implications

Namespace choice has direct security implications:

**Dedicated namespaces:**
- **Principle of least privilege**: ServiceAccounts can be scoped to just one namespace
- **Blast radius containment**: Compromise of one job doesn't affect others
- **Audit clarity**: Namespace-level audit logs show only one job's activity

**Shared namespaces:**
- **Complex RBAC**: Must use RoleBindings with fine-grained resource matching
- **Secret sprawl risk**: All jobs can potentially list Secrets in the namespace
- **Network policies**: Harder to implement pod-to-pod isolation

**Recommendation:** For security-sensitive operations (database access, API key usage, PII processing), always use dedicated namespaces with `create_namespace: true`.

### Cost and Resource Management

Namespace strategy affects resource management:

**Dedicated namespaces:**
- Apply ResourceQuotas per job to prevent resource hogging
- Easier to track costs with namespace-level tags
- Can use LimitRanges to enforce pod-level resource bounds

**Shared namespaces:**
- Single quota shared by all jobs (can lead to contention)
- Harder to attribute costs to individual jobs
- Requires careful label-based monitoring for cost allocation

**Example ResourceQuota for dedicated namespace:**
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: backup-job-quota
  namespace: db-backup-prod
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
    persistentvolumeclaims: "1"
```

### Migration Strategy

If you need to migrate between patterns:

**From shared to dedicated:**
1. Deploy new CronJob with `create_namespace: true` and different name
2. Verify it works correctly
3. Suspend the old CronJob (`suspend: true`)
4. After one cycle, delete the old CronJob
5. The old namespace remains (since it was shared)

**From dedicated to shared:**
1. Pre-create the shared namespace
2. Deploy new CronJob with `create_namespace: false` referencing shared namespace
3. Verify it works correctly
4. Delete old CronJob (its dedicated namespace gets deleted automatically)
5. Update Git/IaC to prevent accidental recreation

### Summary: Choosing the Right Pattern

| Factor | Dedicated Namespace | Shared Namespace |
|--------|-------------------|------------------|
| **Isolation** | ✅ High | ❌ Low |
| **RBAC Simplicity** | ✅ Simple | ❌ Complex |
| **Operational Overhead** | ❌ Higher | ✅ Lower |
| **GitOps Compatibility** | ⚠️ Moderate | ✅ High |
| **Resource Quotas** | ✅ Per-job control | ⚠️ Shared quota |
| **Cost Attribution** | ✅ Easy | ⚠️ Requires labels |
| **Cleanup** | ✅ Automatic | ⚠️ Manual/complex |
| **Multi-tenancy** | ❌ Not applicable | ✅ Ideal |

**General recommendation:**
- **Production, security-sensitive jobs**: Use dedicated namespaces (`create_namespace: true`)
- **Related batch jobs in a mature GitOps environment**: Use shared namespaces (`create_namespace: false`)
- **Development/testing**: Use dedicated namespaces for easy cleanup
- **When in doubt**: Start with dedicated namespaces, migrate to shared only when operational overhead becomes a problem

The `create_namespace` flag makes this choice explicit and reversible, allowing teams to adapt their namespace strategy as their operational maturity evolves.

---

## Conclusion: Production-Ready by Default

Kubernetes CronJobs are a powerful primitive for scheduled automation, but the native API's defaults and complexity make it easy to deploy unreliable, insecure, or resource-exhausting workloads. The deployment landscape ranges from anti-patterns (manual `kubectl apply`) to production-proven approaches (GitOps with ArgoCD/Flux), with IaC tools providing unified infrastructure management.

The key insight is that **production-readiness is not about adding features; it's about preventing failures before they happen**. The most common CronJob disasters—thundering herds, resource exhaustion, silent failures, security leaks—are all *preventable* with correct configuration. Yet the native API makes it easy to forget critical fields or accept dangerous defaults.

Planton's `CronJobKubernetes` resource solves this by **codifying production best practices into the API itself**. Safe defaults prevent the common failure modes. Mandatory validation prevents resource unbounded jobs. A flattened, 80/20 API makes the correct path the easy path. The result: teams can deploy production-grade scheduled workloads without needing to become Kubernetes CronJob experts.

**Deployment recommendation:** Use Planton to define your CronJob resources, deploy them via GitOps (ArgoCD or Flux), monitor with kube-state-metrics and Prometheus PushGateway, and design your jobs to be idempotent. This combination provides the reliability, auditability, and operational excellence required for production systems.

**Remember:** A CronJob is only as reliable as its configuration, its deployment process, and its observability. Start with safe defaults, deploy declaratively, monitor continuously, and design for failure.


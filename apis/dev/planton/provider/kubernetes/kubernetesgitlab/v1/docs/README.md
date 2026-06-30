# Deploying GitLab on Kubernetes: A Strategic Guide

## Introduction: The Cloud-Native GitLab Paradox

For years, the conventional wisdom around deploying GitLab on Kubernetes was clear: "It's complicated, and you probably shouldn't." GitLab is a distributed DevOps platform comprising over a dozen microservices—web servers, background job processors, Git storage, container registries, and more. Each component has distinct operational requirements, and several are inherently stateful. The natural instinct when approaching Kubernetes is to containerize everything: "If it runs in a container, it should run in Kubernetes."

But here's the paradox: **the only officially supported production architecture for GitLab on Kubernetes is one where the most critical components *don't* run in Kubernetes at all.**

This isn't a limitation of Kubernetes or a shortcoming of GitLab's engineering. It's a pragmatic architectural decision driven by a single, critical constraint: **Gitaly**, GitLab's Git repository storage service, requires high-performance local storage and predictable infrastructure that Kubernetes pods—with their ephemeral, network-attached storage and automated lifecycle management—cannot reliably provide at production scale.

This document explains the deployment landscape for GitLab on Kubernetes, presents the evolution from naive approaches to production-ready architectures, and clarifies what Planton supports and why.

## The Maturity Spectrum: From Anti-Pattern to Production

### Level 0: The Anti-Pattern — Manual StatefulSets

**What it is:** Creating raw Kubernetes manifests with simple Deployments for stateless services (webservice, sidekiq) and StatefulSets for stateful components (PostgreSQL, Redis, Gitaly).

**Why it fails:**

GitLab is not a simple three-tier web application. It's a complex distributed system with precise orchestration requirements:

- **Gitaly as a Single Point of Failure:** Gitaly, which stores all Git repositories, is by design a single point of failure when running standalone. In a standard Kubernetes deployment, pod rotation during upgrades, node maintenance, or evictions causes the Gitaly pod to terminate and restart, disrupting all Git operations for every user.
- **The HA Solution Doesn't Work in Kubernetes:** GitLab's high-availability solution for Gitaly is called Gitaly Cluster (implemented using Praefect). However, Gitaly Cluster is explicitly documented as "unsuited to run in Kubernetes" due to its design constraints around storage and consistency guarantees.
- **Lifecycle Conflict:** Traditional VM-based deployments use graceful reloads to update Gitaly without disruption. The Kubernetes pod lifecycle doesn't provide this capability, creating a fundamental mismatch.

The simple StatefulSet approach is viable for development or proof-of-concept deployments where downtime during pod restarts is acceptable. For production, it's an architectural dead-end.

**Verdict:** Anti-pattern for production use.

---

### Level 1: The Default — Official Helm Chart (All-in-One)

**What it is:** Running `helm install gitlab gitlab/gitlab` with default settings. This deploys the entire GitLab stack—webservice, sidekiq, gitaly, postgresql, redis, and minio—all inside your Kubernetes cluster using the official Helm chart.

**Why it exists:**

The official `gitlab/gitlab` Helm chart is an "umbrella chart" that orchestrates over a dozen sub-charts. The default configuration is designed for **trial purposes only**. It provides a complete, working GitLab instance in minutes, which is invaluable for evaluation, development, and testing.

**The production problem:**

The in-cluster stateful components are explicitly labeled as non-production:

- **PostgreSQL:** Uses the Bitnami PostgreSQL chart. Not load-tested by GitLab for production workloads.
- **Redis:** Uses the Bitnami Redis chart. Not production-ready.
- **MinIO:** Provides S3-compatible object storage in-cluster. Suitable only for trials.
- **Gitaly:** Runs as a StatefulSet with the same SPoF issues described in Level 0.

The chart documentation states clearly: *"Do not use the defaults for production."*

**Verdict:** Excellent for dev/test environments. Not suitable for production.

---

### Level 2: The Hybrid — Cloud-Native Hybrid Architecture (Production Standard)

**What it is:** The only officially supported production architecture. This is a "hybrid" deployment where:

- **Stateless components run in Kubernetes:** Webservice (Rails/Puma), Sidekiq (background jobs), GitLab Shell (SSH), Registry (container images), and KAS (agent server) are deployed as scalable, cloud-native workloads.
- **Stateful components run externally:** PostgreSQL, Redis, object storage, and Gitaly all run *outside* the Kubernetes cluster.

**How it works:**

The Helm chart is configured to *disable* all in-cluster stateful services and connect to external dependencies:

| Dependency | External Solution | Configuration |
|------------|-------------------|---------------|
| **PostgreSQL** | AWS RDS, GCP Cloud SQL, Azure Database | Set `postgresql.install: false`, provide `global.psql.host` |
| **Redis** | AWS ElastiCache, GCP Memorystore, Azure Cache | Set `redis.install: false`, provide `global.redis.host` |
| **Object Storage** | AWS S3, GCS, Azure Blob Storage | Set `global.minio.enabled: false`, provide S3 credentials |
| **Gitaly** | Dedicated VMs with local SSD/NVMe storage | Set `global.gitaly.enabled: false`, provide external Gitaly endpoints |

**Why this architecture?**

1. **Gitaly's Requirements:** Gitaly requires extremely high I/O performance (CPU, memory, fast local disk). Running it on dedicated VMs with local SSD/NVMe storage provides the performance needed for production Git operations.
2. **High Availability:** Managed database and cache services (RDS Multi-AZ, ElastiCache with replicas) provide production-grade HA without operational overhead.
3. **Durability:** Object storage services provide 99.999999999% (11 nines) durability for artifacts, LFS files, container images, and backups.
4. **Separation of Concerns:** Kubernetes manages ephemeral, stateless workloads excellently. Cloud providers manage databases, caches, and storage excellently. Use each for what it does best.

**Trade-offs:**

- **Complexity:** You must provision and configure external infrastructure before deploying GitLab.
- **Cost:** A production deployment requires the Kubernetes cluster *plus* managed services. This is significantly more expensive than a single large VM running GitLab Omnibus.
- **Operational Overhead:** You're responsible for managing external dependencies (backups, upgrades, monitoring).

**When to choose this:**

- You need production-grade reliability and performance
- Your organization is above 100-500 concurrent users
- You require horizontal scalability
- You're comfortable managing cloud infrastructure

**Verdict:** The only production-ready architecture for GitLab on Kubernetes.

---

### Level 3: Enterprise HA — Gitaly Cluster (Praefect) on External VMs

**What it is:** An enhancement of the Cloud-Native Hybrid architecture that adds true high availability for Git repository storage using Gitaly Cluster (Praefect).

**How it differs:**

In the standard hybrid model (Level 2), Gitaly runs on a single external VM or a small set of VMs, but it remains a potential single point of failure. Gitaly Cluster addresses this:

- **Praefect:** A reverse proxy and transaction manager that sits in front of multiple Gitaly nodes
- **Replication:** Every Git write operation is replicated to multiple Gitaly nodes (typically 3)
- **Automatic Failover:** If a Gitaly node fails, Praefect automatically routes requests to healthy replicas
- **Load Balancing:** Read operations are distributed across Gitaly nodes

**The critical constraint:**

Gitaly Cluster (Praefect) is explicitly **not supported for production when running in Kubernetes**. The entire Praefect stack—3 Praefect nodes, 3+ Gitaly nodes, and a separate PostgreSQL database for Praefect metadata—must run on dedicated VMs.

**Architecture:**

```
┌─────────────────────────────────────┐
│     Kubernetes Cluster              │
│  ┌──────────────────────────────┐   │
│  │  Webservice, Sidekiq, etc.   │───┼──→ External Praefect Load Balancer
│  └──────────────────────────────┘   │          │
└─────────────────────────────────────┘          │
                                                  ↓
                               ┌─────────────────────────────────┐
                               │  Praefect Cluster (VMs)         │
                               │  ┌─────────────────────────┐    │
                               │  │  Praefect Nodes (3x)    │    │
                               │  │  Gitaly Nodes (3x)      │    │
                               │  │  Praefect PostgreSQL    │    │
                               │  └─────────────────────────┘    │
                               └─────────────────────────────────┘
```

**When to choose this:**

- You're running GitLab for thousands of concurrent users
- Git repository availability is mission-critical
- Your organization requires 99.9%+ uptime SLAs
- You have the infrastructure and budget for a dedicated Gitaly cluster

**Verdict:** Production-ready for enterprise scale, but requires significant infrastructure investment.

---

## The GitLab Operator: Is It Different?

GitLab provides an official Kubernetes Operator, and it's natural to ask: "Should I use the Operator or the Helm chart?"

**Short answer:** The Operator is not a fundamentally different deployment method. It's automation *on top of* the Helm chart.

**What the Operator does:**

- Provides a Custom Resource Definition (CRD) for GitLab instances
- Automates the upgrade process ("Day 2 operations")
- Orchestrates the correct sequencing of Helm chart updates

**What the Operator doesn't do:**

- Manage stateful dependencies (you still need external PostgreSQL, Redis, Gitaly, etc.)
- Eliminate the complexity of production configuration
- Change the underlying Cloud-Native Hybrid architecture

**The Operator wraps the Helm chart.** It uses the Helm chart under the hood to provision resources. As a result, it inherits all the same architectural constraints and configuration requirements.

**Status and maturity:**

As of mid-2024, the GitLab Operator is still considered experimental or suitable only for specific scenarios (particularly OpenShift environments). Work is underway to remove its "experimental" status, but the Helm chart remains the de facto standard for production deployments.

**Verdict:** Use the Operator if you need automated upgrade orchestration and are comfortable with its experimental status. Otherwise, the Helm chart is the proven, stable path.

---

## Deployment Tool Comparison

| Method | Primary Tool | Maturity | Production Support | Key Characteristic |
|--------|--------------|----------|-------------------|-------------------|
| **Manual StatefulSets** | kubectl, YAML | N/A | ❌ Anti-Pattern | Fails to manage GitLab's distributed complexity; Gitaly is a SPoF |
| **Official Helm Chart** | helm | ✅ GA | ✅ Yes (Hybrid Only) | De facto standard. Umbrella chart; default is non-production |
| **GitLab Operator** | kubectl (CRD) | ⚠️ Experimental | ✅ Yes (Hybrid Only) | Automates Day 2 upgrades; wraps the Helm chart |
| **Bitnami Chart** | helm | ✅ GA | ❌ Not by GitLab | Simpler but not officially supported for production |
| **Cloud Marketplace** | Cloud Console | ✅ GA | ✅ Yes (by Vendor) | Automated Helm deployment pre-configured for hybrid model |

---

## The Essential Configuration: 80/20 Analysis

For a production GitLab deployment using the Cloud-Native Hybrid architecture, what do you actually need to configure?

The ironic truth: **most of your configuration is about turning things off.** The default Helm chart installs everything in-cluster. Your job is to disable those defaults and provide connection details to external services.

### The Essential 20%

These are the fields that 80% of production users need:

**Global Settings:**
- `global.hosts.domain`: Your GitLab domain (e.g., `gitlab.company.com`)
- `global.edition`: `ce` (Community Edition) or `ee` (Enterprise Edition)
- `gitlab.migrations.initialRootPassword.secret`: Kubernetes secret with the initial admin password

**Disable In-Cluster Defaults:**
- `postgresql.install: false`
- `redis.install: false`
- `global.minio.enabled: false`
- `global.gitaly.enabled: false`

**Connect to External Services:**
- **Database:** `global.psql.host`, `global.psql.password.secret`
- **Redis:** `global.redis.host`, `global.redis.password.secret`
- **Object Storage:** `global.appConfig.object_store.connection.secret` (S3/GCS credentials)
- **Gitaly:** `global.gitaly.external` (array of external Gitaly VM endpoints)

**Ingress & TLS:**
- `global.ingress.tls.secretName`: Kubernetes secret with TLS certificate
- `global.ingress.configureCertmanager`: Set to `true` to use cert-manager for automated certificates

**Resource Allocation:**
- `gitlab.webservice.resources`: CPU/memory requests and limits
- `gitlab.sidekiq.resources`: CPU/memory requests and limits

### Example: Production Configuration

This values.yaml demonstrates the essential configuration for a production deployment:

```yaml
# Production: Cloud-Native Hybrid
global:
  hosts:
    domain: gitlab.mycompany.com
  ingress:
    configureCertmanager: false  # Using pre-provisioned certificate
    tls:
      enabled: true
      secretName: gitlab-tls-cert
  edition: 'ee'
  
  # Reference pre-created secrets
  gitlab:
    license:
      secret: gitlab-ee-license
  gitaly:
    authToken:
      secret: gitlab-gitaly-token
  psql:
    password:
      secret: gitlab-db-password
  redis:
    password:
      secret: gitlab-redis-password
  appConfig:
    object_store:
      connection:
        secret: gitlab-s3-credentials

gitlab:
  migrations:
    initialRootPassword:
      secret: gitlab-initial-root-password

# --- DISABLE ALL IN-CLUSTER STATEFUL SERVICES ---

# 1. Disable in-cluster PostgreSQL
postgresql:
  install: false
# Configure external PostgreSQL (e.g., AWS RDS)
global:
  psql:
    host: gitlab-db.abc123.us-east-1.rds.amazonaws.com
    port: 5432
    database: gitlabhq_production
    username: gitlab

# 2. Disable in-cluster Redis
redis:
  install: false
# Configure external Redis (e.g., AWS ElastiCache)
global:
  redis:
    host: gitlab-redis.xyz456.ng.0001.use1.cache.amazonaws.com
    port: 6379

# 3. Disable in-cluster MinIO
global:
  minio:
    enabled: false
# (Object storage configured via appConfig.object_store.connection.secret)

# 4. Disable in-cluster Gitaly
global:
  gitaly:
    enabled: false
# Configure external Gitaly (running on a VM)
global:
  gitaly:
    external:
      - name: 'default'
        hostname: 'gitaly-vm-1.mycompany.com'
        port: 8075

# --- RESOURCE ALLOCATION ---
gitlab:
  webservice:
    minReplicas: 2
    maxReplicas: 5  # Enable HPA
    resources:
      requests: { cpu: "1", memory: "2Gi" }
      limits: { cpu: "2", memory: "4Gi" }
  sidekiq:
    minReplicas: 2
    maxReplicas: 5  # Enable HPA
    resources:
      requests: { cpu: "500m", memory: "1Gi" }
      limits: { cpu: "1", memory: "2Gi" }
```

---

## Infrastructure Dependencies: The Strategic Choices

### Database: PostgreSQL

**In-Cluster Option:**
The Helm chart can deploy PostgreSQL using the Bitnami chart. This is suitable only for development and testing.

**Production Path:**
Use a managed database service:
- **AWS:** Amazon RDS for PostgreSQL
- **GCP:** Cloud SQL for PostgreSQL
- **Azure:** Azure Database for PostgreSQL

**Why external?**
- Automated backups with point-in-time recovery
- Multi-AZ high availability
- Automated patching and upgrades
- Performance insights and monitoring
- Read replicas for scaling

---

### Cache & Queues: Redis

**In-Cluster Option:**
The Helm chart can deploy Redis using the Bitnami chart. This is suitable only for development and testing.

**Production Path:**
Use a managed cache service:
- **AWS:** Amazon ElastiCache for Redis
- **GCP:** Cloud Memorystore for Redis
- **Azure:** Azure Cache for Redis

**Critical constraint:** GitLab does **not** support Redis Cluster mode. Use Redis Standalone with HA (Primary/Replica setup).

**Why external?**
- Automatic failover with replicas
- Automated backups
- Automatic patching
- At-rest and in-transit encryption

---

### Object Storage

**In-Cluster Option:**
The Helm chart can deploy MinIO for S3-compatible storage. This is suitable only for development and testing.

**Production Path:**
Use cloud object storage:
- **AWS:** Amazon S3
- **GCP:** Google Cloud Storage
- **Azure:** Azure Blob Storage

GitLab stores the following in object storage:
- CI/CD artifacts
- Git LFS objects
- File uploads (issue attachments, avatars, etc.)
- Container registry images
- Terraform state files
- Packages (npm, maven, PyPI, etc.)

**Why external?**
- 99.999999999% (11 nines) durability
- Unlimited scalability
- Versioning and lifecycle policies
- Cross-region replication for DR
- Native integration with CDNs

---

### Git Repository Storage: Gitaly

**In-Cluster Option:**
The Helm chart can deploy Gitaly as a StatefulSet. This creates a single point of failure and is experimental for production use.

**Production Path:**
Run Gitaly on dedicated VMs with high-performance local storage:
- **Storage:** Local SSD or NVMe drives (not network-attached)
- **Sizing:** Based on repository data size and I/O requirements
- **Operating System:** Linux (Ubuntu, Debian, RHEL)

**For Enterprise HA:**
Deploy Gitaly Cluster (Praefect) on VMs:
- 3x Praefect nodes (load balancer + transaction coordinator)
- 3x Gitaly nodes (each with local high-speed storage)
- 1x PostgreSQL database (for Praefect metadata)

**Why external?**
- Gitaly requires predictable, high-performance local storage
- Kubernetes persistent volumes are typically network-attached, creating I/O bottlenecks
- Gitaly Cluster (Praefect) is explicitly unsupported in Kubernetes
- VM-based deployment allows for graceful updates without Git operation disruption

---

## GitLab Runner Integration: Cloud-Native CI/CD

A GitLab deployment is incomplete without its CI/CD execution engine: GitLab Runner.

### Deployment Method: GitLab Runner Helm Chart

The official, production-ready method is the `gitlab/gitlab-runner` Helm chart. This is a separate chart from the main GitLab chart.

**Why separate?**
- Runners are often deployed to multiple clusters or namespaces
- Runner scaling is independent of the GitLab server
- Different teams may manage runners with different configurations

### The Kubernetes Executor: Pod-per-Job Architecture

The key to cloud-native CI/CD is the **Kubernetes executor**. When configured with this executor:

1. The GitLab Runner pod acts as a **controller**, not an executor
2. For every CI/CD job, the Runner creates a new, ephemeral pod
3. The job executes inside this pod
4. When the job completes, the pod is destroyed

**Benefits:**
- **Isolation:** Each job runs in its own environment
- **Security:** Jobs cannot interfere with each other
- **Scalability:** Limited only by cluster capacity
- **Resource Efficiency:** Pods are created on-demand and destroyed when idle

### Autoscaling: The Cluster Autoscaler Pattern

Don't use HorizontalPodAutoscaler for the Runner manager pod. Instead, rely on the **Kubernetes Cluster Autoscaler**:

1. Developer triggers 10 parallel CI jobs
2. Runner manager receives jobs and attempts to create 10 job pods
3. Cluster nodes are at capacity; job pods enter "Pending" state
4. Cluster Autoscaler detects pending pods
5. Autoscaler provisions new nodes
6. Job pods schedule and execute
7. When jobs complete and nodes are idle, Autoscaler scales down

This creates a perfectly elastic CI/CD system that scales from zero to hundreds of concurrent jobs and back down automatically.

### Security: RBAC and Least Privilege

**Runner Manager RBAC:**
The Runner manager pod needs permissions to create, list, watch, and delete pods, ConfigMaps, and Secrets in the namespace where jobs run.

**Job Pod RBAC (Critical):**
By default, job pods inherit the Runner's service account, giving every CI job the power to create pods—a security nightmare.

**Best Practice:**
- Configure the Runner with a minimal service account
- For deployment jobs that need Kubernetes API access, use the `KUBERNETES_SERVICE_ACCOUNT_OVERWRITE` CI/CD variable to assign a different, more privileged service account only to those specific jobs

This implements the principle of least privilege: a lint job has no Kubernetes permissions, but a deployment job has the permissions it needs.

---

## Production Operations Best Practices

### High Availability

**Stateless Components:**
Scale horizontally using Kubernetes HorizontalPodAutoscaler (HPA) based on CPU or memory metrics. Webservice and Sidekiq are designed for this.

**Stateful Components:**
HA is provided externally:
- **PostgreSQL:** Use cloud provider's Multi-AZ or read replica features
- **Redis:** Use Primary/Replica configuration with automatic failover
- **Gitaly:** Use Gitaly Cluster (Praefect) on VMs for true Git storage HA

---

### Backup and Disaster Recovery

GitLab data is distributed across three storage systems, each requiring a different backup strategy:

1. **PostgreSQL Database**
   - Use cloud provider's automated snapshot tools (e.g., RDS automated backups)
   - Enable point-in-time recovery
   - Test restore procedures regularly

2. **Object Storage**
   - **Not included in GitLab's backup utility**
   - Enable versioning and lifecycle policies
   - Configure cross-region replication for geographic redundancy

3. **Git Repositories (Gitaly)**
   - Use Gitaly's "server-side backup" feature
   - Backups stream directly to object storage
   - Run from the `toolbox` pod using the `gitlab-backup` utility

**Full DR Plan:** Back up both the GitLab backup bucket and the primary object storage buckets.

---

### Security Hardening

**TLS/SSL:**
- Terminate TLS at the Ingress controller
- Use cert-manager for automated certificate acquisition and renewal
- Enforce HTTPS for all connections

**Secrets Management:**
- **Don't rely on base64 encoding alone** (default Kubernetes Secrets)
- Use an external secrets manager: HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager
- Integrate with **External Secrets Operator (ESO)** to sync external secrets into Kubernetes
- Rotate secrets regularly

**Network Policies:**
- Implement Kubernetes NetworkPolicy resources
- Restrict pod-to-pod communication to only what's necessary
- Example: Only webservice and sidekiq pods should access Redis

**RBAC:**
- Apply principle of least privilege to service accounts
- Regularly audit permissions
- Use separate service accounts for different GitLab components

---

### Common Pitfalls

1. **Resource Underprovisioning**
   - **Most common failure mode**
   - Gitaly and Sidekiq are CPU and memory intensive
   - Underprovisioning causes OOMKills, slow performance, cascading failures
   - Start with generous allocations and tune based on monitoring

2. **Storage Performance**
   - Using standard network-attached persistent volumes for Gitaly creates I/O bottlenecks
   - This is the primary reason Gitaly should run on VMs with local SSD/NVMe storage

3. **Misconfigured Object Storage**
   - Forgetting to configure object storage causes failures for CI artifacts, uploads, container registry
   - Incorrect permissions (IAM policies) are a frequent issue
   - Test object storage connectivity before deploying GitLab

---

### Cost Considerations

**Reality check:** A production GitLab deployment on Kubernetes in the Cloud-Native Hybrid model is **expensive**.

**Cost components:**
- Kubernetes cluster (multiple nodes with adequate CPU/memory)
- Managed PostgreSQL (with Multi-AZ HA)
- Managed Redis (with replicas)
- Object storage (can become significant with large CI artifact volumes)
- Dedicated Gitaly VMs (with high-performance local storage)
- For HA: Full Gitaly Cluster (6+ VMs plus a PostgreSQL instance)

**Cost comparison:**
- A single large VM running GitLab Omnibus (the traditional deployment method) is significantly cheaper for small-to-medium deployments
- The hybrid Kubernetes architecture provides value through scalability, reliability, and separation of concerns—but not through cost savings

**When Kubernetes makes sense:**
- Your team is growing beyond 500 concurrent users
- You require horizontal scalability
- You need geographic distribution
- Your organization already operates Kubernetes at scale

---

## Planton's Approach

Planton's `GitlabKubernetes` API is designed **hybrid-first**. This is not a limitation—it's a deliberate design choice that aligns with GitLab's official production guidance.

### What Planton Provides

- **API Abstraction:** A Protobuf-defined API that captures the essential 80/20 configuration
- **Hybrid Model by Default:** The API assumes external dependencies are the primary path
- **Connection Configuration:** The API configures GitLab to connect to your external PostgreSQL, Redis, object storage, and Gitaly—it does not provision those services for you
- **Production Patterns:** Sensible defaults for resource allocation, scaling, and security

### What Planton Doesn't Do

- **Provision External Infrastructure:** You are responsible for provisioning your RDS instance, ElastiCache cluster, S3 buckets, and Gitaly VMs
- **Manage Day 2 Operations:** Backups, upgrades, and monitoring of external dependencies are your responsibility
- **Support In-Cluster Stateful Defaults:** The API does not encourage the "trial-only" all-in-one deployment for production use

### Why This Design?

Because attempting to abstract away the hybrid architecture—to make it appear simpler than it is—would be dishonest and would lead to production failures. The Cloud-Native Hybrid model is the reality of production GitLab on Kubernetes. Planton's role is to make that model easier to configure and deploy correctly, not to hide its inherent complexity.

---

## Conclusion: Pragmatic Cloud-Native Architecture

Deploying GitLab on Kubernetes is not about achieving a "pure" cloud-native deployment where everything runs in containers. It's about pragmatically leveraging Kubernetes for what it excels at—managing ephemeral, stateless workloads—while using purpose-built infrastructure (managed databases, object storage, dedicated VMs) for stateful components that have specific operational requirements.

The Cloud-Native Hybrid architecture is not a compromise or a workaround. It's the correct production pattern, validated by GitLab's engineering team and deployed at scale by enterprises worldwide.

If you're evaluating whether to deploy GitLab on Kubernetes:

- **Choose Kubernetes if:** You need to scale beyond 500 concurrent users, require horizontal scalability, and have the infrastructure expertise to manage the hybrid model.
- **Choose GitLab Omnibus on VMs if:** You're a small-to-medium team (under 500 users), prefer operational simplicity, or want to minimize infrastructure costs.

There's no universally "right" choice—only the right choice for your organization's scale, expertise, and requirements.

---

## Further Reading

- **GitLab Helm Chart Documentation:** [https://docs.gitlab.com/charts/](https://docs.gitlab.com/charts/)
- **Reference Architectures:** [https://docs.gitlab.com/administration/reference_architectures/](https://docs.gitlab.com/administration/reference_architectures/)
- **Gitaly on Kubernetes:** [https://docs.gitlab.com/administration/gitaly/kubernetes/](https://docs.gitlab.com/administration/gitaly/kubernetes/)
- **GitLab Runner on Kubernetes:** [https://docs.gitlab.com/runner/install/kubernetes/](https://docs.gitlab.com/runner/install/kubernetes/)


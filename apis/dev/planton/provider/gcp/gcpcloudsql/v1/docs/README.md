# Deploying Google Cloud SQL: From Manual Clicks to Production-Grade Automation

## Introduction

For years, the conventional wisdom around managed databases was simple: "They're just MySQL or PostgreSQL in the cloud—point, click, done." Yet production deployments tell a different story. A single misconfigured network setting can leave your database unreachable from Cloud Run. Forgetting to enable high availability means a zone failure takes down your entire application. Using the wrong machine tier violates your SLA and costs you customer trust.

Google Cloud SQL is GCP's fully managed relational database service for MySQL, PostgreSQL, and SQL Server. It promises to eliminate the operational burden of running traditional RDBMSs—no patching, no backup management, no replication configuration. But "fully managed" doesn't mean "zero configuration." The path from a working dev database to a resilient, secure, production-grade instance requires understanding a spectrum of deployment approaches and making informed architectural choices.

This guide explores **how to deploy Cloud SQL**, from anti-patterns to production-ready automation. We examine the tools developers actually use—from Console wizards to Infrastructure-as-Code frameworks—and explain what Planton supports and why.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern—Manual Console with Public IP and 0.0.0.0/0

**What it is:** Creating an instance through the GCP Console, enabling a public IP address, and adding `0.0.0.0/0` to the "Authorized Networks" list.

**Why developers do this:** It works immediately. Every client—your laptop, CI/CD, serverless functions—can connect without VPC configuration or proxy setup.

**Why it's dangerous:** Your production database is exposed to the entire internet, protected only by a password. IP whitelisting with `0.0.0.0/0` is not security; it's a critical vulnerability. GCP's own Recommender API actively flags this configuration and recommends disabling it.

**Verdict:** Acceptable only for throwaway demos. Never for staging or production.

### Level 1: The "Click Ops" Pattern—Manual Console with Authorized Networks

**What it is:** Using the Console wizard but properly whitelisting specific CIDR ranges (developer IPs, application server IPs) instead of `0.0.0.0/0`.

**Why it's better:** You've eliminated the worst security risk. Your database is no longer open to the world.

**Why it's still problematic:**
- **Brittle:** Developer IPs change (home, office, VPN). Each change requires a manual Console update.
- **Not repeatable:** Disasters happen. If you need to recreate the instance, you're clicking through the wizard again, hoping you remember all the settings.
- **No versioning:** Configuration lives in the Console, not in Git. You can't review changes, roll back mistakes, or enforce policy.

**Verdict:** A step up from Level 0, but fundamentally incompatible with modern DevOps practices. Use this only for early prototyping.

### Level 2: The Scripting Era—gcloud CLI and Ansible

**What it is:** Using `gcloud sql instances create` with flags for configuration, or writing Ansible playbooks with the `google.cloud` collection.

**What this unlocks:**
- **Repeatability:** Your configuration is now codified in scripts or playbooks, stored in version control.
- **Composability:** Ansible's modular design shines here. The `gcp_sql_instance` module creates the instance, `gcp_sql_database` creates logical databases within it, and `gcp_sql_user` manages credentials. This separation of concerns is a foundational pattern in modern IaC.

**Limitations:**
- **State management:** Ansible (by default) and `gcloud` are stateless. If you run the same script twice, you get errors or undefined behavior. They don't know what already exists.
- **Secret management:** Handling the root password securely in scripts requires external tools (Vault, GCP Secret Manager) and careful plumbing.

**Verdict:** A significant maturity leap. Ansible is production-viable for teams already using it for application deployment and configuration. For pure infrastructure provisioning, the next level is more powerful.

### Level 3: The Modern Standard—Terraform and Pulumi

**What it is:** Declaring infrastructure as code using Terraform's HCL or Pulumi's general-purpose programming languages (Python, TypeScript, Go).

**Why this is production-grade:**

**Stateful and Declarative:** Both tools maintain a state file that maps your declared configuration to the actual GCP resources. You declare the desired end state; the tool figures out what to create, update, or delete. This eliminates the "what if I run it twice?" problem.

**Battle-Tested Abstractions:** The Terraform `google_sql_database_instance` resource and Pulumi's `gcp.sql.DatabaseInstance` are mature, widely used, and cover the full Cloud SQL API surface. Community modules (like `terraform-google-modules/terraform-google-sql-db`) encapsulate best practices, including tricky configurations like VPC peering for Private IP.

**Secret Management:** Terraform supports "write-only" attributes. The `google_sql_user` resource has a `password_wo` (write-only) field. You can *set* a password on user creation, but Terraform will never read it back or store it in the state file. This is a critical security pattern. (Note: The instance's `root_password` does not have a write-only variant, which is intentional—the best practice is to set a strong root password once and *never use it*, creating all operational users via `google_sql_user` with `password_wo`.)

**Multi-Environment Patterns:** Terraform uses directory-based environments (`env/dev`, `env/staging`, `env/prod`) or remote state backends. Pulumi has first-class "Stacks" built in. Both enable identical code to deploy different configurations to different environments.

**The OpenTofu Nuance:** OpenTofu is a community-driven, open-source fork of Terraform 1.6, created after HashiCorp's license change in 2023. Crucially, OpenTofu uses the **exact same `hashicorp/google` provider** that Terraform uses. Any analysis of Terraform's Cloud SQL resource schema applies identically to OpenTofu. For teams prioritizing open-source licensing, OpenTofu is Terraform without the licensing concern.

**Terraform vs. Pulumi:**

| Criterion | Terraform / OpenTofu | Pulumi |
|-----------|---------------------|--------|
| **Language** | HCL (declarative DSL) | Python, TypeScript, Go, C# |
| **State Security** | Secrets visible in state file (requires external encryption) | Secrets encrypted by default in state |
| **Multi-Environment** | Directory-based or workspaces | First-class Stacks |
| **Ecosystem** | Industry standard; largest community | Growing; excellent for teams preferring code over DSL |
| **GCP Support** | Official joint support from Google and HashiCorp | Third-party, but production-quality |

**Verdict:** This is the gold standard. Choose Terraform/OpenTofu for maximum ecosystem support and battle-tested patterns. Choose Pulumi if your team prefers writing infrastructure in real programming languages with loops, functions, and type safety.

### Level 4: The Kubernetes-Native Approach—Config Connector and Crossplane

**What it is:** Managing Cloud SQL instances as Kubernetes Custom Resources (CRDs), applying YAML manifests via `kubectl` or GitOps tools (Flux, Argo CD).

**Why this exists:** For organizations where Kubernetes is the operational hub, this model is philosophically elegant. Your infrastructure (a Cloud SQL instance) and your workloads (Deployments, Services) are managed the same way: declarative YAML, stored in Git, applied to a cluster, reconciled by controllers.

**How it works:**
- **Config Connector:** GCP's native solution. Install it in your GKE cluster. Define an `SQLInstance` CRD. Config Connector's controller watches for these resources and calls the GCP APIs to create/update/delete the actual Cloud SQL instance.
- **Crossplane:** A cloud-agnostic CNCF project. It uses the same CRD pattern but supports multiple clouds. It also has a "Composition" model where developers can request abstract resources (a `PostgreSQLInstance` "claim") and platform teams define how that maps to a concrete implementation (a `CloudSQLInstance` on GCP, or an RDS instance on AWS).

**API Design Insight:** The CRD specifications from Config Connector and Crossplane are some of the best-designed Cloud SQL APIs available. They use a clean, nested `spec.settings` structure (mirroring the GCP API) and demonstrate secure patterns like `spec.rootPassword.valueFrom.secretKeyRef` for secrets.

**Limitations:**
- **Operational complexity:** You need a running, healthy Kubernetes cluster to manage your database. If the cluster is down, you can't provision or modify Cloud SQL instances (though existing instances keep running—they're managed by GCP, not K8s).
- **Not universal:** This pattern only makes sense for teams already operating on Kubernetes. For everyone else, it's over-engineering.

**Verdict:** Production-grade for Kubernetes-centric organizations. Overkill for most teams. The key lesson here is the **API design patterns** these tools use—Planton should emulate their nested, composable structure.

### Level -1: The Deprecated Path—Cloud Deployment Manager

**What it was:** GCP's original native IaC service, using YAML with Jinja or Python templates.

**Status:** **Deprecated.** Cloud Deployment Manager reaches end-of-support on **March 31, 2026**. Google's official successor is "Infrastructure Manager," a managed service for running **Terraform**.

**The strategic signal:** By deprecating its own tool in favor of a managed Terraform service, Google has effectively endorsed the Terraform provider ecosystem as the de facto standard for GCP infrastructure provisioning.

**Verdict:** Do not use for any new deployments. If you're on CDM, migrate to Terraform, OpenTofu, or Infrastructure Manager.

## Production Essentials: What Separates Dev from Prod

A development Cloud SQL instance and a production instance are not just different in size—they're architecturally different. Here's what must change:

### High Availability: Regional (Multi-Zone) Architecture

**Zonal (Default):** A single instance in a single zone. If that VM fails, or if the entire zone has an outage, your database goes offline. Recovery is a **manual** process (restore from backup), and downtime is measured in hours.

**Regional (High Availability):** This provisions a *primary* instance in one zone and a *standby* replica in a different zone within the same region. Data is written **synchronously** to a regional persistent disk that replicates across both zones. A heartbeat system monitors the primary. On failure, Cloud SQL automatically promotes the standby, reassigns the instance's IP address, and completes failover in approximately **60 seconds**.

**Trade-offs:**
- **Pro:** Near-zero-downtime recovery from zone failures.
- **Con:** Slightly higher write latency (synchronous replication) and double the instance cost (you pay for the standby).

**For production OLTP workloads, Regional (HA) is non-negotiable.** The cost of a zone outage—lost revenue, customer trust, SLA penalties—dwarfs the cost of the standby instance.

**Configuration:** Set `settings.availability_type = REGIONAL` in your API or Terraform.

### Backups and Point-in-Time Recovery (PITR)

**Automated Backups:** Cloud SQL takes daily backups within a configured 4-hour window. These protect against hardware failures and major corruption.

**Point-in-Time Recovery (PITR):** This is the ability to restore your database to a specific *second* in time, such as 09:30:15 AM, right before someone ran `DROP TABLE customers;`.

**How PITR works:** Cloud SQL retains transaction logs (binary logs for MySQL, write-ahead logs for PostgreSQL). Combined with a daily backup, these logs allow reconstruction of the database state at any moment within the retention window (1-7 days for Enterprise, up to 35 days for Enterprise Plus).

**Prerequisites for PITR:**
1. **Automated backups** must be enabled.
2. **Binary/transaction logging** must be enabled.

**Configuration:** In your API, expose two separate booleans:
- `backup_configuration.enabled` (for daily backups)
- `backup_configuration.point_in_time_recovery_enabled` (for PITR)

**Why this matters:** Backups protect against disasters. PITR protects against human error and targeted attacks. Production databases need both.

### Network Security: The "Smart Hybrid" Pattern

Networking is the most complex and commonly misconfigured aspect of Cloud SQL. Here's the evolution:

**Anti-Pattern: Public IP + Authorized Networks**
- What it is: Enable a public IP, whitelist specific CIDR ranges.
- Problem: Brittle (IPs change), insecure (relies on IP spoofing protection), violates least-privilege principles.

**"Classic Security": Private IP Only (VPC Peering)**
- What it is: Disable public IP, enable Private IP, link the instance to a VPC via "Private Services Access."
- Pros: Excellent security for in-VPC resources. All traffic stays on Google's private network—no internet exposure.
- Cons: **Inflexibility.** This configuration breaks all access from outside the VPC, including:
  - GKE (if not in the same VPC or using VPC-native clusters with IP aliasing)
  - Cloud Run and other serverless platforms
  - Developer laptops (requiring complex bastion hosts or IAP tunnels)
- Gotcha: VPC Peering is **not transitive**. If VPC-A peers with the Cloud SQL service network, and VPC-B peers with VPC-A, VPC-B **cannot** access the Cloud SQL instance. This is a common networking trap.

**The "Smart Hybrid" Pattern (Recommended):**
1. Enable **Private IP** (for low-latency, in-VPC access).
2. Enable **Public IP**.
3. Set **Authorized Networks** to an **empty list** (`[]`).

**How this works:**
- Resources inside your VPC (like GCE VMs in the same network) connect to the Private IP address—fast, secure, no internet.
- Resources outside your VPC (GKE, Cloud Run, developer machines, CI/CD) connect via the **Cloud SQL Proxy**.
- The Cloud SQL Proxy is a client-side tool that creates an IAM-authorized, TLS 1.3-encrypted tunnel to the instance. It connects to the public endpoint but authenticates using **IAM credentials**, not IP whitelisting.
- Because Authorized Networks is empty, the public endpoint is effectively secure—only IAM-authorized clients using the Proxy can connect.

**Configuration:**
```
ip_configuration:
  ipv4_enabled: true
  private_network: "projects/my-project/global/networks/my-vpc"
  authorized_networks: []  # Empty—no IP whitelisting
```

**Verdict:** This pattern provides maximum security **and** maximum flexibility. It's the production best practice.

### Observability: Enable Query Insights

Standard Cloud Monitoring metrics (CPU, memory, disk I/O) are automatically available. But the most valuable observability feature is **Query Insights**.

**What it does:**
- Identifies slow or problematic queries in production.
- Traces query load back to the application making the request.
- Provides query execution plans (in Enterprise Plus edition).

**Critical detail:** Query Insights is **opt-in**. It must be explicitly enabled.

**Configuration:** Add a boolean field: `settings.query_insights_enabled = true`.

**Why this matters:** Without query insights, troubleshooting production performance issues is guesswork. With it, you have application-level tracing from query to source code.

### Instance Sizing: Shared-Core vs. Dedicated-Core

**Shared-Core Tiers (`db-f1-micro`, `db-g1-small`):**
- What: vCPUs shared with other tenants.
- Use case: **Development and testing only.**
- Problem: Not covered by the Cloud SQL SLA. Not suitable for production load.

**Dedicated-Core Tiers (`db-n1-standard-1`, `db-n1-standard-2`, etc.):**
- What: Guaranteed, dedicated vCPUs and memory.
- Use case: **All production workloads.**
- Predictable performance, covered by SLA.

**Editions:**
- **Enterprise:** The standard tier. 99.95% SLA for Regional (HA) instances.
- **Enterprise Plus:** Premium tier. 99.99% SLA, 4x read performance (for SQL Server), and advanced monitoring. Planned maintenance downtime reduced from ~60 seconds to <10 seconds.

**Recommendation:** Use shared-core only for dev. Use Enterprise for standard production. Use Enterprise Plus for mission-critical, high-transaction workloads.

## What Planton Supports and Why

Planton's `GcpCloudSql` API is designed to orchestrate **Terraform** (or OpenTofu) to provision Cloud SQL instances. Here's why:

### Why Terraform (and OpenTofu)?

1. **Industry Standard:** Terraform is the most widely used, battle-tested IaC tool for GCP. The `hashicorp/google` provider is jointly maintained by Google and HashiCorp.
2. **Google's Endorsement:** By deprecating Cloud Deployment Manager and replacing it with a managed Terraform service (Infrastructure Manager), Google has signaled that Terraform is the de facto standard.
3. **Mature Ecosystem:** Community modules, extensive documentation, and years of production use mean edge cases and "gotchas" are well-documented and solved.
4. **OpenTofu Compatibility:** OpenTofu is a drop-in, open-source replacement for Terraform. It uses the exact same provider. Supporting Terraform automatically means supporting OpenTofu.

### API Design Philosophy: The 80/20 Principle

Most IaC tools expose every API parameter, leading to sprawling, 50+ field resource definitions. Planton takes a different approach: **focus on the 20% of configuration that 80% of users need**.

**Essential Fields (The 80%):**
- `name`: Instance ID
- `database_version`: Engine and version (e.g., `MYSQL_8_0`, `POSTGRES_15`)
- `region`: GCP region
- `project`: GCP Project ID
- `settings.tier`: Machine type (e.g., `db-f1-micro`, `db-n1-standard-2`)

With just these five fields, you can deploy a working Cloud SQL instance for development.

**Production-Ready Fields (The 20%):**
- `deletion_protection`: Prevent accidental deletion
- `settings.edition`: `ENTERPRISE` or `ENTERPRISE_PLUS`
- `settings.availability_type`: `ZONAL` or `REGIONAL` (HA toggle)
- `settings.disk_size_gb` and `settings.disk_auto_resize`
- `settings.backup_configuration`: `enabled`, `start_time`, `point_in_time_recovery_enabled`
- `settings.ip_configuration`: `ipv4_enabled`, `private_network`, `authorized_networks`
- `settings.maintenance_window`: Schedule maintenance for off-peak hours
- `settings.query_insights_enabled`: Enable advanced monitoring

**Rare/Advanced Fields:**
- `settings.database_flags`: Custom tuning (e.g., `slow_query_log: on` for MySQL, `log_min_duration_statement: 1000` for PostgreSQL).

**Modular Resource Model:**
The research validates a key architectural decision: **separate resources for instances, databases, and users.**

- **GcpCloudSql:** Manages the instance (the VM, storage, networking).
- **GcpCloudSqlDatabase:** Manages logical databases (schemas) within that instance.
- **GcpCloudSqlUser:** Manages users and credentials, with secure `password_wo` (write-only) support.

This mirrors the pattern used by Terraform, Ansible, and Pulumi. It's more flexible, more composable, and scales better than a monolithic "do everything" resource.

## MySQL vs. PostgreSQL: Choosing Your Engine

### Version Selection

Cloud SQL supports multiple major versions of both MySQL and PostgreSQL. **Recommendation:** Always select the version marked as "default" in GCP documentation for new instances (e.g., `MYSQL_8_0`, `POSTGRES_15` or `POSTGRES_17` as of late 2024).

**End-of-Life Warning:** Many older versions (PostgreSQL 9.6, 10, 11, 12 and MySQL 5.6, 5.7) are entering "Extended Support" on February 1, 2025. Starting May 1, 2025, instances running these EOL versions will incur **additional charges**. Plan your upgrades accordingly.

### MySQL vs. PostgreSQL: Feature Comparison

| Feature | MySQL | PostgreSQL |
|---------|-------|------------|
| **Database Model** | Purely Relational | Object-Relational |
| **Best For** | High-frequency **read** operations | High-frequency **write** operations, complex queries |
| **Index Types** | B-tree, R-tree | B-tree, Hash, Partial, Expression, GIN |
| **Advanced Data Types** | JSON | JSONB, UUID, Arrays, Geometric |
| **ACID Compliance** | Yes (with InnoDB engine) | Yes (always) |
| **Concurrency** | MVCC (InnoDB) | Full MVCC |
| **Ease of Use** | Simpler, easier for beginners | More complex, powerful, extensible |

**Recommendation:**
- **MySQL:** Web applications, content management systems, e-commerce platforms. Best for read-heavy workloads.
- **PostgreSQL:** Analytical workloads, applications with complex queries, systems requiring advanced data types (JSONB for semi-structured data, UUIDs for distributed systems).

### Common Database Flags

Database flags are the API-driven "escape hatch" for tuning a managed database.

**MySQL:**
- `slow_query_log: on` — Enable slow query logging
- `log_output: FILE` — Route logs to Cloud Logging (required for slow query log to be useful)
- `time_zone: UTC` — Set default time zone
- `performance_schema: on` — Enable detailed performance metrics

**PostgreSQL:**
- `log_min_duration_statement: 1000` — Log queries running longer than 1000ms
- `log_statement: 'ddl'` — Log all schema changes (CREATE, ALTER, DROP)
- `auto_explain.log_min_duration: '5s'` — Automatically log execution plans for queries running >5 seconds

## When Not to Use Cloud SQL: Strategic Alternatives

Cloud SQL is the right choice for 90% of traditional relational database workloads. But it's not universal.

### Cloud SQL vs. Self-Managed on GCE

**Self-Managed on GCE:**
- **Control:** Full control over OS, database software, configuration.
- **Responsibility:** You handle all patching, backups, replication, HA, security.
- **Use case:** Legacy applications with very specific, non-standard configurations. Rarely the right choice for new applications.

### Cloud SQL vs. Cloud Spanner

**Cloud Spanner** is not a bigger Cloud SQL. It's an architecturally different database.

| | Cloud SQL | Cloud Spanner |
|----------|-----------|---------------|
| **Database Model** | Traditional RDBMS (MySQL, PostgreSQL, SQL Server) | Globally distributed, horizontally scalable RDBMS |
| **Scalability** | **Vertical** (scale up to a larger machine) | **Horizontal** (scale out to more nodes) |
| **Availability SLA** | 99.95% (Regional HA) or 99.99% (Enterprise Plus) | 99.999% (five nines) |
| **Best For** | Traditional web apps, e-commerce, CMS | Planet-scale data (>10TB), global consistency, massive transaction volume (>100K r/w/sec) |
| **Use Cases** | Standard line-of-business applications | Financial systems, global logistics, gaming leaderboards |

**Recommendation:**
- **Cloud SQL:** Your default for traditional relational workloads.
- **Cloud Spanner:** When your data is >10TB, your application is globally distributed, and you need global consistency with five-nines availability.

## Conclusion

The journey from "just a managed MySQL instance" to a production-grade, resilient, secure Cloud SQL deployment is a journey through architectural choices. A single toggle—`availability_type: REGIONAL`—transforms a fragile, single-zone instance into a near-zero-downtime, multi-zone system. A properly configured network (the Smart Hybrid pattern) balances security and flexibility, allowing in-VPC resources to connect privately while enabling serverless platforms and developers to use the IAM-authorized Cloud SQL Proxy.

Google's deprecation of Cloud Deployment Manager in favor of a managed Terraform service sends a clear strategic signal: the Terraform provider ecosystem is the industry standard for GCP infrastructure. Planton's choice to orchestrate Terraform (and OpenTofu) for Cloud SQL provisioning aligns with this reality, leveraging a mature, battle-tested toolchain.

By focusing on the 80/20 principle—providing first-class support for the fields that matter most (HA, backups, networking, machine sizing) while still exposing an escape hatch for advanced tuning (database flags)—Planton's `GcpCloudSql` API aims to be both simple for common cases and powerful for production-grade deployments.

Cloud SQL is not magic. It's a managed service that eliminates the *operational* burden of database administration. But it still requires thoughtful *configuration* to achieve production-grade resilience, security, and performance. This guide is your roadmap from "click to create" to "deploy with confidence."


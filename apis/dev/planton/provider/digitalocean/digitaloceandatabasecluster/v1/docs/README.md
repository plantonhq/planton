# DigitalOcean Managed Database Clusters

## Introduction

The promise of cloud databases has always been simple: offload the complex, time-consuming work of database administration so your team can focus on building applications, not managing infrastructure. DigitalOcean Managed Databases delivers on this promise with a focused value proposition—**simplicity and predictable pricing**—targeting startups and small-to-medium businesses that lack dedicated database administrators and value reduced operational overhead over granular control.

Unlike hyperscalers like AWS and Google Cloud, which offer overwhelming feature sets that often require specialized expertise to navigate, DigitalOcean provides the core 80% of required features in a streamlined, developer-friendly interface. But the real financial differentiator often goes unnoticed: **bandwidth costs**. While AWS and Google Cloud charge up to $0.15/GB for data egress, DigitalOcean charges a flat $0.01/GB, with all VPC-internal traffic completely free. For data-heavy applications—analytics platforms, public-facing APIs, content-heavy sites—this translates to dramatically lower total cost of ownership, even when raw compute prices appear comparable.

This document explores the deployment and management landscape for DigitalOcean Managed Databases, the production standards that separate development clusters from mission-critical infrastructure, and the critical platform-specific considerations that every architect must understand before committing workloads.

## The Deployment Spectrum: From Click-Ops to Declarative Automation

DigitalOcean provides a full spectrum of interaction models for provisioning and managing database clusters. Understanding this progression is essential for choosing the right tool for your team's maturity and requirements.

### Level 0: Manual Provisioning via the Control Panel

**What it is:** The DigitalOcean Control Panel is a web-based GUI offering a wizard-driven provisioning process. Users select a datacenter region, database engine, version, cluster configuration (node plan, node count, storage), and the cluster is created with a few clicks.

**When it's appropriate:** This method is ideal for initial service exploration, rapid prototyping, and teams without established automation practices.

**The anti-patterns:** The simplicity of the UI can lead to dangerous production mistakes:

- **Single-Node Clusters**: The UI allows selecting `node_count = 1`. While cheaper, this configuration is **not highly available**. A node failure results in downtime while the service reprovisions from backup. This is acceptable only for development or testing.
- **Open Firewalls**: Rushing through the "Trusted Sources" configuration and adding `0.0.0.0/0` exposes the database to the entire public internet, inviting brute-force attacks.
- **Confusing Dev Databases**: DigitalOcean's App Platform offers "dev databases"—a limited, non-scalable PostgreSQL option intended only for development within App Platform. These are not production-grade Managed Databases.

**Verdict**: Use the Control Panel for learning and exploration, but never for production infrastructure. Manual provisioning doesn't scale, creates configuration drift, and lacks auditability.

### Level 1: Scripted Automation (CLI, API, SDKs)

**What it is:** DigitalOcean provides imperative tools for scripting database lifecycle management:

- **doctl**: The official command-line interface. Commands like `doctl databases create`, `doctl databases resize`, and `doctl databases firewalls` provide full control over clusters via shell scripts.
- **API v2**: The underlying RESTful API (`POST /v2/databases`, `PUT /v2/databases/{id}/resize`) that powers all tools.
- **SDKs**: Official libraries for Go (`godo`) and Python (`PyDo`) that wrap the API in native language constructs.

**When it's appropriate:** This level is suitable for custom automation, CI/CD pipeline integration, or teams that prefer scripting over declarative state management.

**The limitation:** Imperative tools are stateful in the operator's head, not in code. If a script runs twice, you may create duplicate resources. If a team member manually changes a setting, the script doesn't know about it. There's no built-in drift detection or plan previews.

**Verdict**: A step up from manual provisioning for automation, but not production-grade infrastructure management. The lack of state tracking and drift detection makes it unsuitable for managing complex, multi-environment infrastructure.

### Level 2: Configuration Management (Ansible)

**What it is:** Ansible is a push-based configuration management tool with robust modules for provisioning cloud resources. DigitalOcean offers two Ansible collections:

- `community.digitalocean`: The older, community-driven collection with the `digital_ocean_database` module.
- `digitalocean.cloud`: The newer, official collection with the `database_cluster` module.

These modules provide declarative-style configuration (you declare the desired state in a YAML playbook), and Ansible reconciles it.

**When it's appropriate:** Teams already invested in Ansible for configuration management may find this a natural extension for infrastructure provisioning.

**The limitation:** While Ansible supports declarative syntax, it lacks the sophisticated state management, dependency graphing, and plan/preview capabilities of purpose-built Infrastructure as Code (IaC) tools. Ansible is also typically run in "push" mode, which can create race conditions in multi-user environments.

**Verdict**: A pragmatic choice for Ansible-native teams, but falls short of the robustness and developer experience provided by dedicated IaC tools.

### Level 3: Infrastructure as Code (Terraform, Pulumi, Crossplane) — The Production Standard

**What it is:** Declarative IaC tools represent the most advanced, robust, and repeatable method for managing cloud infrastructure. You define the desired state in configuration files (or code), and the tool automatically reconciles the actual state to match, with built-in drift detection, plan previews, and state management.

DigitalOcean is fully supported by the industry's leading IaC platforms:

**Terraform and OpenTofu**

The `digitalocean/digitalocean` Terraform provider is comprehensive, production-ready, and the foundation for DigitalOcean's IaC ecosystem. OpenTofu (an open-source fork of Terraform) is fully compatible with the same provider.

Key resources include:

- `digitalocean_database_cluster`: The core resource for provisioning clusters.
- `digitalocean_database_firewall`: Manages "Trusted Sources" network access rules.
- `digitalocean_database_user`: Manages individual database users within a cluster.
- `digitalocean_database_db`: Creates individual databases within a cluster.
- `digitalocean_database_connection_pool`: Manages PgBouncer connection pools for PostgreSQL.
- `digitalocean_database_replica`: Provisions read-only replicas for horizontal scaling.
- Engine-specific virtual resources (e.g., `digitalocean_database_valkey_config`) for advanced tuning.

This granular, composable design is the hallmark of production IaC. The firewall is a separate resource, not a nested field, because it may be managed by a different team (security) than the database (platform). Users and databases are separate because they have different lifecycles and ownership.

**Pulumi**

Pulumi offers the same infrastructure management capabilities but allows developers to use general-purpose programming languages (TypeScript, Python, Go, C#) instead of HCL.

A critical architectural detail: The `pulumi/digitalocean` provider is a **bridged provider**. Pulumi's "Terraform bridge" automatically generates Pulumi providers from existing Terraform providers. This means:

- The Pulumi DigitalOcean provider is a wrapper around the `digitalocean/digitalocean` Terraform provider.
- It has **identical resource coverage, features, and limitations**.
- A bug in the Terraform `digitalocean_database_cluster` resource will also exist in Pulumi's `DatabaseCluster` resource.
- The choice between Terraform and Pulumi is one of **developer experience** (HCL vs. programming languages), not technical capability.

**Crossplane**

Crossplane extends the Kubernetes API to manage external cloud resources. A developer provisions a DigitalOcean database by applying a YAML manifest (`kubectl apply -f my-db.yaml`). A Crossplane controller running in the cluster then calls the DigitalOcean API.

**Critical version note**: The original `crossplane-contrib/provider-digitalocean` is **archived and unmaintained**. The correct, official provider is `crossplane-contrib/provider-upjet-digitalocean`, built using Upjet (Crossplane's framework for generating providers from Terraform providers).

**Why IaC is the production standard:**

1. **Version Control**: Infrastructure is defined in Git, providing full audit history and rollback capability.
2. **Drift Detection**: Tools continuously monitor for manual changes and can automatically reconcile them.
3. **Plan/Preview**: See exactly what will change before applying it.
4. **Multi-Environment Management**: Native support for dev/staging/prod environments (Terraform Workspaces, Pulumi Stacks).
5. **Team Collaboration**: State locking prevents concurrent modifications.
6. **Reproducibility**: The same configuration always produces the same infrastructure.

**Verdict**: IaC is non-negotiable for production infrastructure. The choice among Terraform, Pulumi, and Crossplane depends on your team's existing ecosystem and preferred workflow. All three are fully capable and production-ready for DigitalOcean Managed Databases.

## Production Essentials: The Non-Negotiable Standards

Provisioning a managed database does not absolve you of all architectural responsibilities. Production-readiness requires specific configurations and practices.

### High Availability: The Three-Node Mandate

**The Requirement**: Production workloads cannot tolerate downtime from a single node failure. Databases must be deployed with standby nodes.

**The Configuration**:

- `node_count = 1` (Primary only): **Not highly available**. Suitable only for development or testing.
- `node_count = 2` (Primary + 1 Standby): The **minimum HA configuration**.
- `node_count = 3` (Primary + 2 Standbys): The **most resilient HA configuration**.

In a multi-node cluster, DigitalOcean continuously replicates data from the primary to standby nodes. If the primary fails, the service automatically promotes a standby to primary within seconds.

**The Client-Side Responsibility**: Automatic failover is only half the story. Applications must implement robust reconnection and retry logic. Platform maintenance, updates, and failover events cause brief (5-10 second) connection drops. If your application doesn't automatically retry, these "hiccups" become user-facing crashes, even though the database itself recovered.

### Network Security: VPC-First, Tag-Based Firewalls

**The Default**: All new database clusters are provisioned inside a VPC by default, isolated from the public internet.

**The Best Practices**:

1. **Use Private Connection Strings**: For applications in the same DigitalOcean region, always use the private connection string. This routes traffic over DigitalOcean's internal network, which is more secure, has lower latency, and incurs no bandwidth charges.

2. **Lock Down "Trusted Sources" (Firewalls)**: This is the primary access control mechanism.

   **Anti-Pattern**: Adding `0.0.0.0/0` as a trusted source. This exposes the database to the entire internet.

   **Anti-Pattern**: Adding individual Droplet IP addresses. This breaks in auto-scaling environments and is difficult to manage.

   **Best Practice 1 (VPC CIDR)**: Add the VPC's CIDR range (e.g., `10.108.0.0/20`) as a single trusted source. This grants access to all current and future resources within that private network and counts as only one rule (the firewall has a 100-rule limit).

   **Best Practice 2 (Tag-Based)**: Assign DigitalOcean tags (e.g., `prod-webserver`) to resources that need database access. Create a firewall rule with `type = "tag"` and `value = "prod-webserver"`. This is the most flexible and scalable method, as the firewall automatically updates as new, tagged resources come online.

### PostgreSQL: The Connection Limit Crisis and PgBouncer Mandate

DigitalOcean Managed PostgreSQL has **severely limited** direct backend connection limits tied to node RAM:

- 1 GiB RAM: **22 connections**
- 2 GiB RAM: **47 connections**
- 4 GiB RAM: **97 connections**
- 8 GiB RAM: **197 connections**

These limits are easily exhausted by modern applications, especially microservices, serverless functions, or high-concurrency web applications.

**The Mandatory Solution**: DigitalOcean provides a **managed PgBouncer** instance with every PostgreSQL cluster. PgBouncer maintains a pool of connections to the database and funnels thousands of incoming client connections into this efficient pool.

**Production Mandate**: All production applications must be configured to connect via the PgBouncer connection string, not the direct database connection string. This is managed in IaC via the `digitalocean_database_connection_pool` resource.

Failure to use PgBouncer will result in "too many connections" errors and application failures.

### MySQL: The Hotel California Problem

A severe, business-critical limitation exists for DigitalOcean Managed MySQL that creates extreme vendor lock-in.

**The Issue**: The industry-standard tool for fast, non-blocking physical backups and migrations of large MySQL databases is **Percona XtraBackup**. This tool requires the MySQL `BACKUP_ADMIN` privilege. DigitalOcean **explicitly blocks this privilege** on its managed service.

**The Consequence**: This leaves `mysqldump` (a slow, logical backup tool) as the only official method to export data. For large databases, this is catastrophic:

- An 800GB database took **five days** for a complete mysqldump and restore.
- Migrating a large production MySQL database off DigitalOcean requires at least five days of total application downtime.
- This is a commercially non-viable proposition for most businesses.

**The Risk**: This limitation became apparent to one user only when attempting to sell their SaaS company—the buyer could not migrate the data. The user described it as "The Hotel California of Managed Services: You can check in, but you can never leave."

**Recommendation**: If you anticipate operating a large (>500GB) MySQL database and value migration flexibility, seriously consider alternative platforms or self-hosted solutions where you retain full administrative privileges. Planton users must be aware of this risk before committing MySQL workloads to DigitalOcean.

### The Redis-to-Valkey Migration: Future-Proof Your Configuration

As of June 30, 2025, DigitalOcean is discontinuing "Managed Caching" (Redis) and replacing it with **Managed Valkey**—a high-performance, open-source, Redis-compatible datastore that serves as a drop-in replacement.

All existing Redis clusters will be automatically migrated, retaining all data. However, **all new caching clusters should target the `valkey` engine**, not `redis`.

**Planton Implementation Note**: The `spec.proto` currently defines `redis = 3` in the `DigitalOceanDatabaseEngine` enum. This should be updated to `valkey = 3` to align with DigitalOcean's platform evolution and prevent future deprecation warnings.

**Valkey Limitations**:
- No managed daily backups (PITR)
- No read-only replicas
- No forking (cloning from backup)

Valkey supports its own persistence methods (RDB snapshots, AOF append-only file) configured via the `digitalocean_database_valkey_config` virtual resource.

### Storage Autoscaling: The Critical IaC Gap

DigitalOcean recently introduced **Storage Autoscaling**, a crucial production feature that monitors disk utilization and automatically provisions additional storage when a threshold is exceeded. This prevents "disk full" errors and eliminates the need to overprovision storage "just in case."

**The Problem**: Extensive analysis of the `digitalocean/digitalocean` Terraform provider documentation reveals **no declarative argument** to enable this feature. This implies Storage Autoscaling is currently API/UI-only and **cannot be managed via IaC**.

This is a significant gap in the automation ecosystem. Production teams relying on Terraform or Pulumi cannot currently enable this critical feature declaratively.

**Planton Opportunity**: This represents a significant value-add opportunity. Planton should track this feature gap and, if/when DigitalOcean adds IaC support, immediately incorporate it into the `DigitalOceanDatabaseClusterSpec`. Until then, users must enable Storage Autoscaling manually via the Control Panel after IaC provisioning.

### MongoDB: The Sharding Limitation

All DigitalOcean Managed MongoDB clusters are provisioned as **replica sets** by default—the standard production-ready configuration for high availability.

**The Gap**: DigitalOcean does **not support managed sharding**. Sharding is MongoDB's method for horizontal scaling (distributing data across multiple replica sets) to support very large datasets and high-throughput writes.

Users who outgrow the vertical limits of a single replica set are forced to "roll their own" sharding—a highly complex, manual process involving provisioning multiple replica sets, config servers, and running custom `mongos` query router instances. This largely defeats the purpose of a "managed" service.

Specialized competitors like MongoDB Atlas offer "auto-sharding" as a core feature. If your workload requires MongoDB sharding, DigitalOcean may not be the right platform.

## Cost Optimization: The Bandwidth Advantage

DigitalOcean's transparent, predictable pricing model is a key differentiator, especially for data-heavy applications.

### The Three Cost Drivers

1. **Node Size (`size_slug`)**: Plans with dedicated vCPUs cost more than "Basic" plans with shared vCPUs.
2. **Node Count**: HA clusters with `node_count = 3` cost **three times** the base node price.
3. **Storage**: Each `size_slug` includes base storage. Additional storage is billed at $0.215/GiB/month.

### The Bandwidth Differentiator

All traffic **within a VPC in the same datacenter is free**. By placing application servers (Droplets, Kubernetes nodes) in the same VPC as the database and using the private connection string, all application-to-database traffic becomes free and more secure.

For outbound traffic, DigitalOcean charges $0.01/GB (versus AWS's $0.15/GB). For applications with high data egress—analytics, APIs, content delivery—the total cost of ownership can be dramatically lower than on hyperscalers.

### Right-Sizing and Scaling Strategies

- **Start small**: Provision the smallest plan that meets initial needs.
- **Monitor metrics**: Use DigitalOcean's built-in CPU, RAM, and disk monitoring.
- **Vertical scaling**: When metrics show sustained high utilization, scale up to the next `size_slug`. **Note**: This is a one-way operation—you cannot scale down.
- **Horizontal scaling**: For PostgreSQL and MySQL, add read-only replicas (billed separately, starting at $15/month) to distribute read query load.
- **Storage autoscaling** (when IaC-supported): Avoid overprovisioning by letting the platform automatically scale storage as consumption grows.

### No Reserved Instances

DigitalOcean does **not offer Reserved Instances or Savings Plans**. Unlike AWS (which offers up to 72% discounts for 1-3 year commitments), DigitalOcean is purely pay-as-you-go.

This provides maximum flexibility for startups and SMBs with evolving workloads. However, for large, stable, long-term workloads, AWS Reserved Instances may be more cost-effective. The choice is one of flexibility versus locked-in discounts.

## Secret Management: Terraform vs. Pulumi

Handling sensitive information—API tokens and database credentials—reveals a significant philosophical difference between IaC tools.

### Provider Authentication (API Token)

- **Terraform**: Standard practice is to pass the token via an environment variable (`DIGITALOCEAN_TOKEN`), keeping it out of code.
- **Pulumi**: Also supports environment variables, but offers an integrated solution: `pulumi config set --secret digitalocean:token <value>`. This encrypts the token in the stack configuration file using the Pulumi backend or a self-managed secret provider (AWS KMS, HashiCorp Vault).

### Database Credentials

The `database_cluster` resource generates a password as an output. This credential must be securely passed to applications.

**Terraform Best Practices**:

1. **Remote Backend**: The password is stored in **plain text** in `terraform.tfstate`. This file **must** be stored in a secure, encrypted remote backend (DigitalOcean Spaces, AWS S3, HashiCorp Consul), never locally.
2. **Sensitive Outputs**: Mark password exports with `sensitive = true` to prevent them from appearing in logs.

**Pulumi Best Practices**:

1. **First-Class Secrets**: Pulumi automatically marks resource outputs like `password` as secrets.
2. **Automatic Encryption**: When Pulumi saves its state, all secrets are **automatically encrypted**. The plain text value never touches disk in the state file.
3. **Superior Posture**: This default encryption is a superior security posture compared to Terraform, which places the burden of securing the state file entirely on the user.

**The Right Way (Application Integration)**:

The most secure pattern decouples infrastructure provisioning from application secret consumption:

1. **Kubernetes Secrets**: The IaC tool provisions the database user, then uses the password output to create a `kubernetes_secret` resource directly in the target cluster. The application pod mounts this secret as an environment variable.
2. **HashiCorp Vault (Dynamic Secrets)**: The IaC tool provisions the database and configures Vault's Database Secrets Engine with the credentials. Applications authenticate to Vault (e.g., via the Vault CSI driver) to request dynamic, short-lived credentials that are automatically rotated and revoked. This eliminates the problem of static, long-lived database credentials entirely.

## Planton's Choice: Universal IaC Abstraction

Planton supports DigitalOcean Managed Databases through a **universal, cloud-agnostic API** that abstracts the underlying IaC provider (Terraform, Pulumi, or custom Golang modules).

The `DigitalOceanDatabaseClusterSpec` (defined in `spec.proto`) follows the **80/20 principle**: it exposes only the essential fields required for 80% of use cases, keeping the API simple, predictable, and maintainable.

**Essential Fields (Always Required)**:

- `cluster_name`: A human-readable cluster identifier.
- `engine`: The database engine (`POSTGRES`, `MYSQL`, `VALKEY`, `MONGODB`).
- `engine_version`: The major version (e.g., `"15"`, `"8"`).
- `region`: The DigitalOcean datacenter region.
- `size_slug`: The node plan identifier (e.g., `db-s-2vcpu-4gb`).
- `node_count`: The number of nodes (1-3).

**Production-Ready Fields (Optional but Recommended)**:

- `vpc`: Reference to a DigitalOcean VPC for custom network topology.
- `storage_gib`: Custom storage size if more than the `size_slug` default is needed.
- `enable_public_connectivity`: Whether to enable public network access (defaults to `false` for security).

**What's Intentionally Excluded**:

- **Firewall Rules**: Managed via a separate `DigitalOceanDatabaseFirewall` resource (following the Terraform provider pattern of granular, composable resources).
- **Users and Databases**: Managed via separate `DigitalOceanDatabaseUser` and `DigitalOceanDatabaseDb` resources with different lifecycles.
- **Engine-Specific Tuning**: Handled via separate "virtual" resources (e.g., `DigitalOceanDatabaseValkeyConfig`) to keep the core API clean and extensible.
- **Advanced Features**: Maintenance windows, backup restoration (a different action—"fork" vs. "create"), and read replicas are deferred to avoid API bloat.

This design ensures the API remains focused, understandable, and resilient to DigitalOcean platform evolution.

## Conclusion: Simplicity with Eyes Wide Open

DigitalOcean Managed Databases delivers on its core promise: a simplified, cost-predictable database service that eliminates operational overhead for startups and SMBs. The platform's transparent pricing, generous bandwidth allowances, and intuitive developer experience make it a compelling alternative to hyperscaler complexity.

However, simplicity has trade-offs. The PostgreSQL connection limits demand disciplined use of PgBouncer. The MySQL backup privilege restrictions create severe vendor lock-in for large databases. The absence of managed MongoDB sharding limits horizontal scale. The current Storage Autoscaling IaC gap requires manual intervention.

These are not fatal flaws—they are design choices optimized for the service's target audience. DigitalOcean has made a deliberate decision to provide the essential 80% of features with exceptional ease of use, rather than the exhaustive 100% with overwhelming complexity.

For teams building on DigitalOcean, the key is to enter with eyes wide open: understand the limitations, design around them, and leverage Infrastructure as Code to manage your database infrastructure with the same rigor, repeatability, and auditability as your application code.

With the right architecture, the right IaC tooling, and an understanding of the platform's constraints, DigitalOcean Managed Databases can be a robust, cost-effective foundation for production applications.


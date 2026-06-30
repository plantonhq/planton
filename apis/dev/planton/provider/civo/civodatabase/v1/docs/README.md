# Deploying Managed Databases on Civo: A Practical Guide

## Introduction

The decision to use a managed database service versus self-managing databases on compute instances has always been a study in trade-offs. The conventional wisdom for years was straightforward: if you need full control, self-manage; if you want simplicity, pay a premium for a managed service and accept vendor lock-in.

Civo Database challenges this binary by targeting a different value proposition entirely: **radical simplicity with transparent, predictable pricing**. Unlike hyperscaler database services (AWS RDS, GCP Cloud SQL) that layer complexity upon complexity, Civo takes a developer-first approach. The service bundles CPU, RAM, and NVMe storage into fixed tiers and—most notably—charges **no separate fees for egress or I/O operations**. What you see on the pricing page is exactly what you pay.

This document explores the landscape of deployment methods for Civo's managed MySQL and PostgreSQL databases, from manual console clicks to declarative Infrastructure as Code, and explains why Planton provides a critical, missing piece in the Civo cloud-native ecosystem.

## The Database Deployment Spectrum on Civo

When deploying databases on Civo, you essentially have two paths: **managed** (CivoDatabase service) or **self-managed** (on Civo Compute instances or Kubernetes clusters). Understanding when to choose each is fundamental.

### Level 0: Self-Managed on Compute Instances

**What it is:** You provision a Civo Compute instance (or Kubernetes pod) and manually install and configure PostgreSQL or MySQL using standard OS package managers or container images.

**When it's appropriate:**
- You need a database engine or version Civo doesn't offer (e.g., MongoDB, or a specific PostgreSQL fork)
- You require deep OS-level customization, custom extensions (like PostGIS), or non-standard configurations
- You're building a portable, cloud-agnostic application on Kubernetes using database operators

**The operational burden:** This approach requires significant database administration expertise. You're responsible for:
- OS and database patching and security updates
- Backup automation and disaster recovery planning
- High availability configuration (replication, failover)
- Performance monitoring and tuning
- Storage management and scaling

**Verdict:** This is the **20% solution** for specialized workloads that can justify the high operational overhead. For most applications, it's unnecessary complexity.

### Level 1: The Civo Database Service

**What it is:** A fully managed Database-as-a-Service (DBaaS) that automates provisioning, patching, backups, and failure detection. Civo handles the "heavy lifting" so you can focus on your application.

**Supported engines:**
- **PostgreSQL** (actively maintained versions, with public beta access to new releases like PostgreSQL 17)
- **MySQL** (currently supported versions, with deprecation notices for older releases)

**Instance tiers:** Civo simplifies sizing by bundling vCPU, RAM, and NVMe SSD storage into fixed tiers. Unlike AWS RDS where you configure instance type and storage independently, Civo's approach is intentionally coupled:

| Size Tier | Size Slug | vCPU | RAM | Storage | Price/month |
|-----------|-----------|------|-----|---------|-------------|
| Small | `g3.db.small` | 2 | 4 GB | 40 GB | $43.45 |
| Medium | `g3.db.medium` | 4 | 8 GB | 80 GB | $86.91 |
| Large | `g3.db.large` | 6 | 16 GB | 160 GB | $173.81 |
| Extra Large | `g3.db.xlarge` | 8 | 32 GB | 320 GB | $347.62 |
| 2x Extra Large | `g3.db.2xlarge` | 10 | 64 GB | 640 GB | $695.24 |

**High availability:** Civo supports up to 5 additional replica nodes (6 total nodes in a cluster). The service manages automatic failover using a DNS-based mechanism. Applications should connect using the `dns_endpoint` (not the static master endpoint) to benefit from automatic failover.

**The pricing advantage:** This is where Civo truly differentiates itself. A comparable PostgreSQL instance on DigitalOcean costs $60/month versus Civo's $40/month. More importantly, Civo's pricing is **all-inclusive**—no surprise charges for data transfer (egress), no additional IOPS fees. This predictability is a massive advantage over hyperscalers where these "hidden" costs can unexpectedly inflate monthly bills.

**Critical limitation—no Point-in-Time Recovery (PITR):** Civo provides automated daily backups, giving you a 24-hour recovery granularity. However, unlike AWS RDS or GCP Cloud SQL, there is **no native PITR** capability to restore to a specific minute. For production workloads with high transactional value, you should augment the native backups with a custom solution (discussed in the Production Essentials section).

**Verdict:** This is the **80% solution**. For the vast majority of web applications, APIs, and microservices on Civo, the managed database service is the right choice. It balances cost, simplicity, and production readiness.

## The IaC Landscape for CivoDatabase

Moving beyond manual console clicks to production-grade infrastructure requires declarative configuration and state management. The maturity of IaC tooling for Civo varies significantly.

### Terraform: The De Facto Standard

**Maturity:** Production-ready. The official `civo/civo` Terraform provider is actively maintained by Civo and serves as the reference implementation.

**Resource:** The `civo_database` resource abstracts the multi-step API calls (create database, attach firewall, configure replicas) into a single declarative block.

**Key pattern—Security by design:** A production CivoDatabase deployment is never a single resource; it's a composite of three linked resources:

1. **`civo_network`** (private network): Ensures database isolation from the public internet
2. **`civo_firewall`**: Controls access with least-privilege rules
3. **`civo_database`**: The database itself, linked to the network and firewall

Example configuration:

```hcl
# 1. Define isolated private network
resource "civo_network" "db_net" {
  label  = "prod-network"
  region = "LON1"
}

# 2. Define firewall with strict ingress rules
resource "civo_firewall" "db_fw" {
  name       = "prod-db-firewall"
  network_id = civo_network.db_net.id
  region     = "LON1"

  # Allow PostgreSQL only from within private network
  ingress_rule {
    label    = "allow-pg-internal"
    protocol = "tcp"
    port     = "5432"
    cidr     = civo_network.db_net.cidr
    action   = "allow"
  }
}

# 3. Define database, linking network and firewall
resource "civo_database" "prod_db" {
  name        = "production-db"
  engine      = "postgresql"
  version     = "16"
  size        = "g3.db.medium"
  nodes       = 3  // 1 master + 2 replicas for HA
  region      = "LON1"
  
  network_id  = civo_network.db_net.id
  firewall_id = civo_firewall.db_fw.id
}
```

**State management best practice:** The Terraform state file contains sensitive database credentials (username, password). Never store `terraform.tfstate` locally or in version control. Use a secure remote backend such as Civo Object Store (S3-compatible).

**Verdict:** Terraform is the **recommended production tool** for teams already using HashiCorp tooling.

### Pulumi: Multi-Language IaC

**Maturity:** Production-ready with caveats.

**Provider architecture:** The `pulumi-civo` provider is a **bridged provider**—it's programmatically generated from the Terraform provider, not a native Pulumi implementation. This means it inherits the Terraform provider's schema and behavior.

**Advantages:**
- Supports multiple languages (TypeScript, Python, Go, .NET)
- Superior secret management: Pulumi encrypts sensitive values in state by default
- Better developer experience for secret propagation (can create database and Kubernetes secret in same program)

**Example pattern (TypeScript):**

```typescript
import * as civo from "@pulumi/civo";
import * as k8s from "@pulumi/kubernetes";

const db = new civo.Database("prod-db", {
  name: "production-db",
  engine: "postgresql",
  version: "16",
  size: "g3.db.medium",
  nodes: 3,
  networkId: network.id,
  firewallId: firewall.id,
});

// Immediately pipe credentials to Kubernetes secret
const dbSecret = new k8s.core.v1.Secret("db-credentials", {
  stringData: {
    hostname: db.dnsEndpoint,
    port: db.port.apply(p => p.toString()),
    username: db.username,
    password: db.password,
  },
});
```

**Verdict:** Pulumi is **recommended for teams preferring general-purpose programming languages** and needing seamless Kubernetes integration.

### The Critical Gap: Crossplane and Kubernetes-Native IaC

**Current state:** The community `provider-civo` for Crossplane supports only `CivoKubernetes` and `CivoInstance` resources. It does **not support CivoDatabase**.

**What this means:** The primary use case for Crossplane—managing an application's database dependency from within Kubernetes—is currently **impossible** on Civo. There is no way to simply `kubectl apply -f database.yaml` and have a Civo database provisioned as a Kubernetes Custom Resource.

**The Planton opportunity:** This gap creates a perfect, well-defined opportunity. By providing a `civo.planton.dev/v1` database resource, Planton becomes **the first and only solution** to provide true, Kubernetes-native, declarative management for Civo databases.

This isn't just "another IaC tool"—it's filling a critical missing piece in the Civo cloud-native ecosystem.

### Other Tools: Ansible and Civo CLI

**Ansible:** There is **no dedicated `civo_database` module** in Ansible collections. Users must resort to brittle workarounds (shelling out to `civo` CLI or manually calling the REST API with `ansible.builtin.uri`). Not recommended for production.

**Civo CLI:** The `civo database create` command is useful for quick prototyping and scripting, but it's imperative (not declarative) and has no state management. Not suitable for managing production infrastructure.

## Production Essentials

### High Availability Architecture

**Replica configuration:** Set `replicas` to 1 or more in your IaC configuration (this translates to 3+ total nodes: 1 master + replicas).

**Failover mechanism:** Civo manages automatic failover via DNS. The `dns_endpoint` CNAME/A-record is automatically repointed to the new master upon failure.

**Application requirements:**
- **Always** connect using the `dns_endpoint`, never the static `endpoint` or `private_ipv4`
- Implement robust connection retry logic to handle brief DNS TTL windows during cutover
- Test failover scenarios in staging environments

### Disaster Recovery and Backup Strategy

**Native backups:** Civo provides automated daily backups (24-hour granularity).

**Critical production recommendation:** **Do not rely solely on native backups.** For production workloads, implement a custom backup solution:

**Recommended pattern** (Kubernetes CronJob + pg_dump + Civo Object Storage):

1. Create a Kubernetes CronJob in your Civo cluster
2. Run a container that executes `pg_dump` (PostgreSQL) or `mysqldump` (MySQL) against the database
3. Upload the resulting SQL file to Civo Object Storage (S3-compatible)
4. Configure retention policies in object storage

This approach gives you:
- Full control over backup frequency (e.g., hourly vs. daily)
- Flexible retention policies
- Auditable, portable backups in standard SQL format
- No dependency on Civo's black-box backup mechanism

**Restore process:** Standard SQL tools (`psql` or `mysql` client):

```bash
psql -h [civo_host] -U [civo_username] [civo_database] < [dump_file].sql
```

### Network Security Architecture (Defense in Depth)

**First principle:** The database must be **unreachable from the public internet**.

**Layer 1—Network Isolation:**
- **Always** deploy into a `civo_network` (private network)
- The database receives only a private IP address
- Cannot be accessed from outside the Civo private network

**Layer 2—Firewall ACL:**
- **Always** associate a `civo_firewall` with the database
- Configure ingress rules using **least privilege principle**
- Allow database port (e.g., `tcp/5432` for PostgreSQL) **only** from specific CIDR blocks (Kubernetes node range or application instance range)

**Anti-pattern to NEVER use:**

```hcl
# ❌ NEVER DO THIS
ingress_rule {
  cidr = "0.0.0.0/0"  # Exposes database to entire network!
}
```

**Best practice pattern:**

```hcl
# ✅ Restrict to Kubernetes nodes only
ingress_rule {
  protocol = "tcp"
  port     = "5432"
  cidr     = civo_kubernetes_cluster.app_cluster.node_cidr
  action   = "allow"
}
```

### Integration with Civo Kubernetes

**Credential management pattern:**

1. IaC provider (Planton) creates the `civo_database`
2. Provider automatically creates a Kubernetes `Secret` in the cluster with connection details:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  hostname: <base64_encoded_dns_endpoint>
  port: <base64_encoded_port>
  username: <base64_encoded_username>
  password: <base64_encoded_password>
  dbname: <base64_encoded_database_name>
```

3. Application `Deployment` consumes credentials via environment variables:

```yaml
env:
  - name: DB_HOSTNAME
    valueFrom:
      secretKeyRef:
        name: db-credentials
        key: hostname
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-credentials
        key: password
```

This pattern ensures secrets are never hardcoded and flow securely from Civo API → Kubernetes Secret → Application.

### Monitoring and Performance

**Built-in monitoring:** Civo does **not provide** an integrated monitoring dashboard for databases.

**Recommended tooling:**

1. **Percona Monitoring and Management (PMM):** Available in Civo Marketplace. Provides query analytics, performance dashboards, and alerting for MySQL and PostgreSQL. This is the **explicitly recommended** solution for deep database monitoring on Civo.

2. **Prometheus + Grafana:** Deploy a Prometheus operator on your Civo Kubernetes cluster and configure database exporters (`postgres_exporter` or `mysqld_exporter`) to scrape metrics.

**Scaling strategies:**

- **Vertical scaling (scale-up):** Modify the `size_slug` (e.g., `g3.db.medium` → `g3.db.large`). Requires downtime as the instance is reprovisioned.
- **Horizontal scaling (scale-out):** Increase `replicas` count. Preferred method for read-heavy workloads; also improves high availability.

## The Planton Approach

Planton's `CivoDatabase` resource embodies the **80/20 principle**: expose only the fields that 80% of users need, while keeping the API clean and opinionated.

### Essential Configuration Fields

Based on comprehensive analysis of the Civo API, Terraform provider, and production patterns, the Planton API includes these essential fields:

1. **`db_instance_name`**: Unique identifier for the database
2. **`engine`**: Database engine (MySQL or PostgreSQL)
3. **`engine_version`**: Version string (e.g., "16" for PostgreSQL 16)
4. **`size_slug`**: Instance tier (e.g., `g3.db.medium`)
5. **`region`**: Civo region (e.g., `LON1`, `NYC1`)
6. **`network_id`**: Private network ID (security-critical, required)
7. **`firewall_ids`**: Firewall ID(s) for access control (security-critical)
8. **`replicas`**: Number of replica nodes (0-indexed: 0 = master only, 2 = master + 2 replicas)

### What's Intentionally Excluded

**Custom storage sizes:** Not supported by Civo. Storage is bundled with `size_slug` and cannot be configured independently.

**Resource tagging:** Civo's database API does not support tags (unlike compute instances).

### Example Configurations

**Development (minimal cost, no HA):**

```yaml
db_instance_name: "planton-dev-db"
engine: postgres
engineVersion: "16"
size_slug: "g3.db.small"
region: LON1
network_id: "a1b2c3d4-dev-net-id"
firewall_ids: ["f1e2d3c4-dev-fw-id"]
replicas: 0  # Single master node
```

**Staging (moderate cost, basic HA):**

```yaml
db_instance_name: "planton-staging-db"
engine: postgres
engineVersion: "16"
size_slug: "g3.db.medium"
region: LON1
network_id: "b2c3d4e5-staging-net-id"
firewall_ids: ["e2d3c4b5-staging-fw-id"]
replicas: 1  # 1 master + 1 replica = 2 nodes
```

**Production (HA + read scaling):**

```yaml
db_instance_name: "planton-prod-db"
engine: postgres
engineVersion: "16"
size_slug: "g3.db.large"
region: LON1
network_id: "c3d4e5f6-prod-net-id"
firewall_ids: ["d3c4b5a6-prod-fw-id"]
replicas: 2  # 1 master + 2 replicas = 3 nodes
```

### Kubernetes-Native Workflow

With Planton, provisioning a Civo database becomes a standard Kubernetes operation:

```bash
kubectl apply -f civo-database.yaml
```

The Planton controller:
1. Parses the custom resource definition
2. Calls the Civo API to create the database (via `civogo` Go SDK)
3. Handles multi-step provisioning (create → attach firewall → configure replicas)
4. Writes connection credentials to a Kubernetes Secret
5. Reports status back to the custom resource

This fills the **critical gap** left by the community Crossplane provider and provides a cloud-native IaC experience for Civo databases.

## Civo's Unique Position in the Market

### Strengths

**Cost predictability:** The all-inclusive pricing model with no egress or IOPS fees is genuinely unique. For budget-conscious teams or those burned by surprise hyperscaler bills, this is transformative.

**Simplicity:** The bundled instance tiers, minimal API surface, and developer-first design make Civo databases extremely fast to deploy and understand.

**Competitive pricing:** Demonstrably cheaper than DigitalOcean for comparable resources, and far more transparent than hyperscalers.

### Limitations

**No Point-in-Time Recovery (PITR):** Significant for production workloads with low tolerance for data loss. Requires custom backup solutions.

**Bundled storage:** Cannot scale storage independently of compute. A database with large storage needs but low CPU requirements must overprovision (and overpay).

**Nascent ecosystem:** Beyond Terraform/Pulumi, the IaC ecosystem is limited. No native Ansible modules, and the Crossplane provider doesn't support databases—creating the opportunity for Planton.

## Conclusion

The landscape of database deployment on Civo presents a clear strategic choice: for the vast majority of applications, Civo's managed database service offers the right balance of simplicity, cost, and production readiness. Self-managed databases remain viable only for specialized use cases requiring deep customization or extreme portability.

The IaC tooling landscape, however, has a critical gap. While Terraform and Pulumi provide mature, production-ready solutions, the cloud-native community lacks a Kubernetes-native way to manage Civo databases. The Crossplane provider doesn't support them, and manual configuration is error-prone.

Planton's `CivoDatabase` resource fills this gap. By providing a Kubernetes Custom Resource backed by intelligent API abstraction, it enables teams to declare their entire infrastructure—including databases—using standard Kubernetes manifests. This isn't just convenience; it's bringing the declarative, GitOps-friendly workflow that teams already use for applications to the infrastructure layer.

For Civo users building cloud-native applications on Kubernetes, Planton becomes the **first and only** way to achieve this vision. That's not marketing—it's architectural necessity, grounded in the current state of the ecosystem.


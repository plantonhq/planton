# Deploying Temporal on Kubernetes: From Chaos to Clarity

## Introduction

For years, the conventional wisdom was straightforward: "Temporal is complex, running it yourself is even more complex, just use Temporal Cloud." While Temporal Cloud remains an excellent choice for many teams, this advice obscured a more nuanced truth—**self-hosted Temporal on Kubernetes is not only viable but can be production-grade, provided you understand the architecture and avoid common pitfalls.**

The challenge has never been whether Temporal _can_ run on Kubernetes. The challenge has been navigating the bewildering array of deployment methods, database choices, and configuration complexity that can make or break a production deployment. Choose the wrong database backend? You've permanently capped your throughput. Use the default Helm chart settings? You've deployed a development toy, not a production system. Skip schema management? Your cluster won't even start.

This guide synthesizes production experience and deployment patterns to illuminate the path from development to production-grade Temporal on Kubernetes. We'll explore the deployment method landscape, examine why certain approaches fail, compare database backends with honest trade-off analysis, and explain the architectural decisions behind Planton's `TemporalKubernetes` resource.

## The Deployment Maturity Spectrum

Understanding how to deploy Temporal well requires first understanding how _not_ to deploy it. Let's examine the spectrum from anti-patterns to production-ready solutions.

### Level 0: The StatefulSet Fallacy

The reasoning often goes: "Temporal is a stateful application, therefore it needs a Kubernetes StatefulSet." This is a fundamental architectural misunderstanding.

**Why this fails:**

Temporal's architecture consists of multiple cooperating services (frontend, history, matching, worker, web UI) that are **stateless compute**. These services store no state locally—all workflow state persists externally in a shared database (Cassandra, PostgreSQL, or MySQL). StatefulSets provide stable pod identities and persistent storage volumes, neither of which Temporal services need or want.

Using StatefulSets for Temporal services introduces harmful constraints: ordered pod creation/deletion (OrderedReady), stable network identities that serve no purpose, and persistent volume claims that will never be used. The correct Kubernetes resource for all core Temporal services is the **Deployment**, which provides the stateless, rolling update semantics these services require.

The only component that legitimately uses StatefulSets is the database itself (if running in-cluster). However, this leads to our next anti-pattern.

### Level 1: The "Batteries-Included" Trap

The official `temporalio/helm-charts` repository provides a convenient, batteries-included setup: Temporal services, bundled Cassandra, bundled Elasticsearch, bundled Prometheus/Grafana. For local development, this is excellent. For production, this is **explicitly and unambiguously discouraged** by both the Temporal team and production users.

The Temporal team's own guidance is blunt: *"we feel it is not possible to offer a helm deployment of a production quality Cassandra because it's way too complex."* This admission reveals a profound truth—complex, stateful distributed systems like databases are poor candidates for simplistic Helm deployments.

**The production reality:**

Every production deployment follows the same pattern: disable all bundled dependencies (`cassandra.enabled: false`, `elasticsearch.enabled: false`, `prometheus.enabled: false`) and connect to external, production-hardened instances. The batteries are included, but they're for toys, not production machinery.

### Level 2: Helm for Services, External for State

This is the **de facto production standard**. Use the official Helm chart exclusively for the stateless Temporal services, while all persistence and observability dependencies run externally as managed services or dedicated clusters.

**Why this works:**

- **Database as a Service**: Amazon RDS (PostgreSQL), Google Cloud SQL, AWS Keyspaces (Cassandra), or Azure Database provide production-grade databases with automated backups, high availability, and expert management without in-cluster complexity.
- **Managed Observability**: External Elasticsearch clusters (like Elastic Cloud or Amazon OpenSearch) scale independently and don't compete with Temporal for cluster resources.
- **Operational Focus**: Teams focus on running Temporal, not becoming database administrators or Elasticsearch experts.

**Where this falls short:**

While Helm excels at "Day 1" (installation), it fails at "Day 2" (operations):

1. **No Schema Management**: The Helm chart cannot run `temporal-sql-tool` or `temporal-cassandra-tool` to initialize or migrate database schemas. This is a manual, imperative step that breaks declarative GitOps workflows.
2. **No Upgrade Orchestration**: `helm upgrade` is unaware of Temporal's internal state. It cannot perform graceful, service-by-service upgrades that respect workflow execution.
3. **Configuration Complexity**: The `values.yaml` is vast. Production teams end up using `helm template` to render raw manifests, then layering `kustomize` patches for "last-mile" configuration—a brittle, error-prone workflow.

### Level 3: Kubernetes Operators and IaC Abstractions

The emerging production pattern involves tools that solve Helm's Day 2 limitations through intelligent automation.

**Kubernetes Operators** (like the community `temporal-operator`) are in-cluster controllers that manage the full lifecycle. They can:
- Automatically run schema tools during installation and upgrades
- Perform state-aware, graceful cluster upgrades service-by-service
- Watch external resources and react to configuration changes

**IaC Abstractions** (like Planton's `TemporalKubernetes`) combine infrastructure-as-code principles with Kubernetes-native resource management. The typical workflow:

1. **Terraform/Pulumi**: Provision cloud infrastructure (VPC, Kubernetes cluster, RDS database, OpenSearch cluster) and write connection details to Kubernetes Secrets.
2. **GitOps (ArgoCD/Flux)**: Monitor Git repositories for Kubernetes manifests, including `TemporalKubernetes` resources.
3. **Custom Controllers**: Reconcile the declarative resource, deploy Temporal services, configure database connections from secrets, and orchestrate schema management via Kubernetes Jobs.

This level provides:
- **Declarative simplicity**: Define what you want, not how to build it
- **Automated lifecycle**: Schema setup, upgrades, and configuration drift correction
- **Production defaults**: Sensible configurations that prevent common pitfalls
- **Abstraction**: Hide Helm's complexity behind a clean, typed API

## The Database Decision: Your Architecture's Foundation

The choice of persistence backend is the single most consequential architectural decision for a Temporal deployment. This choice fundamentally determines your cluster's scalability, operational complexity, and cost trajectory.

### Understanding the Scalability Challenge

Temporal is "chatty by design"—it generates a high volume of database writes during normal workflow execution. At even moderate loads (~100 RPS), traditional SQL databases become bottlenecks. Database latency directly impacts workflow throughput; the entire cluster's performance is bounded by database performance.

### The Three Paths

#### Path 1: PostgreSQL/MySQL (The Familiar Path)

**When to choose:**
- Development and staging environments
- Production workloads with low-to-medium throughput requirements
- Organizations with strong PostgreSQL/MySQL expertise
- Access to high-quality managed database services (Amazon RDS, Google Cloud SQL)

**Strengths:**
- Operational simplicity and widespread expertise
- Excellent managed service availability
- Lower initial operational overhead

**The bottleneck:**
- **Does not scale horizontally**. Temporal performance is fundamentally capped by single-node database capacity.
- Under high load, becomes the cluster bottleneck regardless of Temporal service replica counts.
- Vertical scaling (larger instances) is limited and expensive.

**The hidden complexity:**
At scale, SQL stops being simple. To achieve horizontal scaling with MySQL requires **Vitess**—a complex database clustering system that provides sharding. At this point, you're choosing between "Cassandra's native horizontal scaling" versus "MySQL with Vitess complexity."

#### Path 2: Cassandra (The Scale Path)

**When to choose:**
- High-throughput production workloads
- Workloads that will grow and need horizontal scalability
- Teams comfortable with eventual consistency models
- Organizations willing to invest in Cassandra operations or use managed services (AWS Keyspaces)

**Strengths:**
- **Native horizontal scalability**: Add nodes to scale out linearly
- **Seamless replication**: Built-in multi-datacenter replication
- **Production-hardened**: Battle-tested at scale (originally developed at Uber for Temporal's predecessor)
- **Highest throughput**: Benchmarks consistently show Cassandra providing the best persistence performance for Temporal

**Challenges:**
- Higher operational overhead for self-managed deployments
- Fewer mature managed service options compared to SQL
- Steeper learning curve for teams unfamiliar with distributed databases

**The truth:**
If you're building for scale, Cassandra isn't more complex—it's more honest about the complexity that scale requires. PostgreSQL at high scale (with read replicas, connection pooling, aggressive tuning) isn't simpler; it's just hiding inevitable bottlenecks.

#### Path 3: The Hybrid Stack (Cassandra + Elasticsearch)

The optimal high-throughput architecture separates concerns:
- **Cassandra for persistence**: Handles high-volume workflow state writes
- **Elasticsearch for visibility**: Handles advanced visibility queries and custom search attributes

This hybrid model solves scalability bottlenecks for _both_ state persistence and observability queries. Benchmarks show this combination provides the best throughput and scalability for large-scale Temporal deployments.

### Database Backend Comparison

| Feature | Cassandra | PostgreSQL | MySQL | MySQL + Vitess |
|:--------|:----------|:-----------|:------|:---------------|
| **Horizontal Scalability** | ✅ Native | ❌ Single-node | ❌ Single-node | ✅ Via sharding |
| **Replication** | ✅ Seamless, built-in | ⚠️ Complex setup | ⚠️ Complex setup | ✅ Via Vitess |
| **Operational Overhead** | 🟡 High | 🟢 Low (High for HA) | 🟢 Low (High for HA) | 🔴 Very High |
| **Managed Service Options** | ⚠️ Limited (AWS Keyspaces) | ✅ Widely available | ✅ Widely available | ⚠️ Niche (PlanetScale) |
| **Persistence Performance** | 🟢 Highest | 🟡 Medium (bottlenecks) | 🟡 Medium (bottlenecks) | 🟢 High |
| **Visibility Performance** | N/A | 🟡 Low-Medium | 🟡 Low-Medium | 🟢 High |
| **Recommended Use Case** | High-throughput production | Dev, staging, low-medium prod | Dev, staging, low-medium prod | High-scale, SQL-native orgs |

## The Visibility Layer: Beyond Basic Queries

Temporal provides two levels of visibility into workflow state:

**Basic visibility** uses the persistence database to list and filter workflows by standard attributes (workflow type, status, start time). This is adequate for development and small deployments.

**Advanced visibility** requires a separate visibility store and enables custom search attributes—arbitrary key-value pairs that can be used to build sophisticated workflow queries (e.g., "find all pending orders for customer X in region Y").

### The Visibility Bottleneck

Just as persistence writes can bottleneck on SQL databases, visibility _queries_ can bottleneck on the same database. The query patterns are different (complex filtering vs. simple reads/writes), but the single-node limitation remains.

### Elasticsearch for Visibility

Elasticsearch is purpose-built for the exact query and indexing patterns that advanced visibility requires. For production deployments with advanced visibility needs, the recommended architecture is:

- **Separate visibility from persistence**: Configure Temporal with distinct stores (e.g., PostgreSQL for persistence, Elasticsearch for visibility, or Cassandra for persistence, Elasticsearch for visibility)
- **External Elasticsearch cluster**: Use managed services (Elastic Cloud, Amazon OpenSearch) or dedicated Elasticsearch clusters

This architectural separation is a first-class concept in Temporal's configuration and should be reflected in any declarative API.

## Planton's Approach: External-First, Production-Default

The `TemporalKubernetes` resource is designed around lessons learned from production deployments, with a philosophy of **external-first** and **safe defaults**.

### Core Design Principles

**1. External Dependencies by Default**

The API assumes you're connecting to external, production-grade databases. Bundled in-cluster databases are opt-in for development environments only (via `devMode` or explicit configuration for Cassandra replicas).

**2. Separate Persistence and Visibility**

The API treats `database.backend` and visibility configuration as distinct concerns, natively supporting the Cassandra + Elasticsearch hybrid pattern without configuration gymnastics.

**3. Automated Schema Management**

The controller automatically runs database schema setup and migrations via Kubernetes Jobs, transforming a manual, imperative step into a declarative, automated one. This is non-negotiable for a production-grade IaC tool.

**4. Abstraction Over Complexity**

Rather than exposing the full complexity of Helm's `values.yaml`, the API provides high-level fields that abstract common patterns:
- `ingress.frontend.enabled: true` generates Service and LoadBalancer/Ingress resources
- `ingress.webUi.enabled: true` handles Web UI exposure
- `externalElasticsearch` configuration sets up all necessary visibility configuration

**5. Safe, Production-Oriented Defaults**

The API prevents common pitfalls through sensible defaults:
- High availability replica counts for core services
- Appropriate resource requests/limits
- Secure-by-default configurations where possible

### Configuration Philosophy: The 80/20 Principle

The API focuses on the 20% of configuration that 80% of users need. Essential fields (database backend, connection details, ingress hostnames) are explicit and required. Advanced tuning (per-service replicas, resource limits, custom annotations) is available but optional.

**Minimal production configuration:**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-prod
spec:
  database:
    backend: postgresql
    externalDatabase:
      host: postgres.example.com
      port: 5432
      username: temporal
      password: "${DB_PASSWORD}"
  ingress:
    frontend:
      enabled: true
      grpcHostname: temporal-grpc.example.com
    webUi:
      enabled: true
      hostname: temporal-ui.example.com
```

This minimal configuration deploys a production-ready Temporal cluster with:
- External PostgreSQL for persistence and visibility
- Automated schema setup
- Exposed frontend (gRPC) via LoadBalancer
- Exposed Web UI via Ingress
- High-availability replicas for core services
- Prometheus integration (if cluster has Prometheus Operator)

### Beyond Basic Deployment: Observability and Security

For enhanced visibility, add external Elasticsearch:

```yaml
spec:
  externalElasticsearch:
    host: elasticsearch.example.com
    port: 9200
    user: elastic
    password: "${ES_PASSWORD}"
```

The controller configures Temporal to use Elasticsearch for advanced visibility queries while continuing to use PostgreSQL for persistence—the hybrid pattern we discussed earlier.

## Common Pitfalls and How We Avoid Them

### Pitfall 1: The numHistoryShards Trap

The number of history shards (`numHistoryShards`) is an **immutable** Day 0 decision. This value determines the maximum parallelism and throughput of the cluster. The Helm chart default is low (typically 4-8).

If you deploy with the default, you've permanently capped your cluster's scalability. Increasing shard count requires creating a new cluster and migrating all data—a complex, manual operation.

**Planton's solution**: The controller uses a high, safe default (e.g., 1024 or higher) to prevent this permanent scalability ceiling.

### Pitfall 2: Forgetting Schema Management

The most common deployment failure: connecting to an external database without initializing the schema. The Temporal services fail to start, logs show database errors, and the deployment is broken.

**Planton's solution**: Automated schema management via Kubernetes Jobs that run `temporal-sql-tool` or `temporal-cassandra-tool` during initial deployment and upgrades.

### Pitfall 3: Choosing SQL Without an Exit Strategy

Teams choose PostgreSQL for its simplicity, deploy to production, scale to moderate load, and hit the single-node bottleneck. At this point, migrating to Cassandra is a major infrastructure project.

**Planton's solution**: The API makes it equally simple to choose Cassandra or PostgreSQL from Day 0. We provide honest guidance about when each is appropriate, removing the "default to SQL" bias.

### Pitfall 4: Treating Workers as an Afterthought

The `TemporalKubernetes` resource deploys the Temporal _server_. Your workers are separate applications that connect to the server. Deploying new worker code that breaks determinism (the contract that allows Temporal to replay workflows) causes in-flight workflows to fail.

**The production solution** is Worker Versioning—running multiple versions of workers simultaneously so running workflows stay pinned to old code while new workflows use new code. This requires sophisticated deployment patterns (sometimes called "rainbow deployments").

While `TemporalKubernetes` solves the server deployment problem, be aware that production-grade worker deployment requires additional strategies and potentially a companion resource for worker lifecycle management.

## The Path Forward

The landscape of deploying Temporal on Kubernetes has matured significantly. What was once a bewildering maze of configuration options, database choices, and operational pitfalls is becoming a well-paved road—provided you have the right map.

The shift from "Helm for everything" to "Helm for services, external for state" to "IaC abstractions with automated lifecycle" represents the natural evolution of production Kubernetes workloads. As teams gained experience, they identified gaps, built better abstractions, and codified best practices.

Planton's `TemporalKubernetes` resource embodies these lessons: external-first for production readiness, automated schema management for Day 2 operations, clean abstractions for complex configuration, and honest defaults that prevent common pitfalls.

Whether you're deploying your first Temporal cluster or migrating from manual Helm configurations, the key insight remains: **complexity doesn't disappear, but thoughtful architecture can make it manageable**. Choose your database for the scale you'll reach, not the scale you're at today. Design for external dependencies from the start. Automate what can be automated. And most importantly, learn from the production war stories of teams who've run Temporal at scale.

The Temporal Workflow Engine is powerful. Running it on Kubernetes is viable. And with the right approach, it can be production-grade from Day 0.

---

## Further Reading

- **Database Schema Management**: Deep dive into automated schema setup, migrations, and the Kubernetes Job patterns used by Planton's controller
- **High Availability Patterns**: Multi-AZ deployments, pod anti-affinity, and disaster recovery strategies for self-hosted Temporal
- **Monitoring and Alerting**: Essential metrics, Grafana dashboards, and alerting rules for production Temporal clusters
- **Security**: mTLS configuration, network policies, secrets management, and encryption patterns
- **Worker Deployment Best Practices**: Strategies for safe worker deployments, versioning, and the determinism contract

*Note: These guides are planned documentation that will provide comprehensive implementation details for advanced topics.*


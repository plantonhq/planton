# Deploying Keycloak on Kubernetes: From Anti-Patterns to Production Excellence

## Introduction: The Stateful Challenge

For years, the conventional wisdom around deploying Keycloak—the industry-standard open-source identity and access management platform—on Kubernetes was "just use a Deployment with multiple replicas for high availability." This advice, while intuitive for stateless applications, represents one of the most common and catastrophic anti-patterns for Keycloak deployments.

The problem? **Keycloak is inherently stateful**. It relies on an embedded Infinispan cache to store critical runtime data including user authentication sessions, offline tokens, and login failure tracking. When you deploy multiple replicas using a standard `kind: Deployment`, each pod boots as a standalone instance with no awareness of its siblings. The result is a "split-brain" scenario where a user authenticates against pod-A, but their next request lands on pod-B, which has no knowledge of their session, forcing them to re-authenticate and completely breaking the Single Sign-On promise that Keycloak exists to deliver.

This document explores the landscape of Keycloak deployment methods on Kubernetes, from common mistakes through intermediate approaches to production-ready solutions. More importantly, it explains **why** Planton has made specific architectural choices that prioritize open-source sustainability, operational simplicity, and Day 2 lifecycle management over mere Day 1 installation convenience.

## The Evolution: From Naive Deployments to Operator-Driven Lifecycle Management

### Level 0: The Anti-Pattern — kind: Deployment

**What it attempts to solve:** Quick deployment with built-in pod restart and basic replication.

**What it breaks:** Everything that makes Keycloak useful in production.

When you create a `kind: Deployment` with `replicas: 3`, Kubernetes spins up three pods simultaneously, each initializing its own Infinispan cache independently. These pods never form a cluster. From the user's perspective, this manifests as random authentication failures. A user logs in, their session is cached in pod-A's memory. The load balancer routes their next request to pod-B, which has never seen this user, so it forces a re-login. Then the user hits pod-C on the third request, triggering yet another authentication challenge.

This isn't a minor inconvenience—it completely negates the value of an SSO solution. Users experience the application as fundamentally broken, and debugging logs will show no obvious errors because each pod is functioning correctly in isolation. The system is working as designed; the design is just wrong.

**Verdict:** Never use a Deployment for multi-replica Keycloak. This approach is only acceptable for single-replica development environments where no clustering is required.

### Level 1: The Foundation — kind: StatefulSet

**What it solves:** Provides the stable network identities and ordered deployment that Keycloak's clustering protocols require.

**What it doesn't solve:** The operational complexity of upgrades, database migrations, and realm management.

A `StatefulSet` gives each pod a stable, predictable hostname (`keycloak-0`, `keycloak-1`, etc.) and ensures that pod-0 reaches the "Ready" state before pod-1 begins to start. This is non-negotiable for Keycloak's clustering to function, as the JGroups-based discovery protocols rely on stable identities to allow nodes to find and communicate with each other.

Modern Keycloak deployments default to **jdbc-ping** for cluster discovery. Instead of relying on Kubernetes-specific mechanisms like DNS queries, each Keycloak node writes its presence (IP address, port) to a shared table (typically named `JGROUPSPING`) in the main Keycloak database. This architectural shift makes the database not just a data store, but the **single source of truth** for cluster coordination. It radically simplifies deployment—nodes only need to reach the same external database—but it also makes database high availability the most critical component of your entire infrastructure.

With a properly configured `StatefulSet`, Keycloak pods will form a coherent cluster, session data will be replicated across nodes via Infinispan, and users will experience seamless SSO as their requests move between pods.

However, a `StatefulSet` is still just a primitive Kubernetes resource. It deploys Keycloak, but it doesn't actively **manage** it. Performing a rolling upgrade, handling database schema migrations, automating backups, or managing realm configurations as code—these Day 2 operational tasks remain manual and error-prone.

**Verdict:** StatefulSets are the foundational requirement for any production Keycloak deployment on Kubernetes, but they're only the starting point, not the destination.

### Level 2: The Installer — Helm Charts

**What they solve:** Parameterized, repeatable Day 1 installation with optional dependency bundling (like PostgreSQL).

**What they don't solve:** Day 2 lifecycle management, declarative realm configuration, or zero-downtime upgrade guarantees.

Helm charts are templated YAML generators. They excel at the "Day 1" problem: getting Keycloak installed into your cluster with all the right configuration knobs exposed in a single `values.yaml` file. For teams that need to quickly spin up Keycloak for testing or development, Helm provides significant value through parameterization and the ability to bundle dependencies.

The Helm landscape for Keycloak has undergone a critical shift in 2024-2025. For years, the **Bitnami Helm chart** (`bitnami/keycloak`) was the de facto community standard due to its polish and comprehensive feature set. However, following VMware's acquisition by Broadcom, Bitnami announced that as of **August 28, 2025**, its container images will no longer be publicly maintained. Security updates and new releases will require a commercial "Bitnami Secure Images" subscription. While the Helm charts themselves remain open-source, they'll point to either unmaintained legacy images or paywalled images.

For open-source projects and teams prioritizing sustainability, this creates an existential dependency risk. Building infrastructure on Bitnami charts today means facing a forced choice in mid-2025: run on unpatched, insecure images, or purchase a Broadcom subscription.

The **Codecentric Helm chart** (`codecentric/keycloak` and `codecentric/keycloakx`) has emerged as the community-driven alternative. Its primary advantage is that it uses the official, Apache 2.0-licensed `quay.io/keycloak/keycloak` container image. It correctly deploys Keycloak as a `StatefulSet` and offers extensive configuration flexibility. It's a viable, open-source-compliant solution for teams that understand its limitations.

The fundamental limitation of **any** Helm chart is that it's an installer, not a lifecycle manager. When you run `helm upgrade`, you're triggering a new templating pass and applying the changes to Kubernetes resources, but the chart has no deep understanding of Keycloak's stateful semantics. It doesn't know how to gracefully perform a rolling restart of a distributed cache cluster, when to wait for database schema migrations to complete, or how to verify that session replication is functioning before marking a pod as "Ready." These capabilities require encoding domain-specific operational knowledge—the hallmark of the Operator pattern.

**Verdict:** Helm charts are valuable for initial deployments and environments where manual Day 2 operations are acceptable. For production systems requiring automated, zero-downtime lifecycle management, they're a necessary but insufficient tool.

### Level 3: The Lifecycle Manager — Kubernetes Operators

**What they solve:** Full application lifecycle management, encoding the operational knowledge of a human expert into automated controllers.

**What they require:** Additional infrastructure (the Operator itself must be deployed and maintained) and a willingness to embrace opinionated defaults.

A Kubernetes Operator is a custom controller that watches for high-level Custom Resource Definitions (CRDs) and translates them into low-level Kubernetes primitives while continuously reconciling the actual state with the desired state. The **Official Keycloak Operator** introduces the `kind: Keycloak` CRD, which abstracts away `StatefulSets`, `Services`, `Ingresses`, and configuration complexity.

Instead of managing dozens of YAML fields across multiple resources, you declare a single `Keycloak` object with essential production fields like `replicas`, `hostname`, `database`, and `tls`. The Operator controller generates and manages the underlying infrastructure using production best practices: it defaults to `jdbc-ping` for cluster discovery, configures proper liveness/readiness probes, handles graceful rolling upgrades, and even automates the creation of Prometheus `ServiceMonitor` resources for observability.

But the Operator's true value proposition is **declarative realm management**. The `KeycloakRealm` CRD allows you to define realms, clients, roles, and identity providers as YAML resources and apply them with `kubectl`. This enables GitOps workflows where your identity configuration is version-controlled, peer-reviewed, and deployed through the same CI/CD pipelines as your application code.

The trade-off? You're adding another moving part to your cluster (the Operator itself), and you're accepting the Operator's opinionated decisions about how Keycloak should run. For teams with highly custom deployment requirements, overriding defaults often requires diving into the `additionalOptions` field and understanding the underlying Keycloak server configuration parameters.

**Verdict:** For production environments prioritizing automation, zero-downtime upgrades, and infrastructure-as-code, the Operator pattern is architecturally superior. It solves the Day 2 problem that Helm charts ignore.

## Comparative Analysis: Understanding Your Options

The following table synthesizes the key decision criteria across the three viable deployment solutions (the Bitnami chart is excluded due to its licensing trajectory):

| **Feature** | **Official Keycloak Operator** | **Codecentric Helm Chart** | **Bitnami Helm Chart** |
|-------------|-------------------------------|---------------------------|------------------------|
| **Deployment Method** | Operator Controller (CRD) | Helm (StatefulSet) | Helm (StatefulSet) |
| **Container Image** | Official `quay.io/keycloak` | Official `quay.io/keycloak` | Custom `bitnami/keycloak` |
| **Image License** | Apache 2.0 | Apache 2.0 | **Commercial Subscription Required (Post-Aug 2025)** |
| **Lifecycle Management** | Full Day 2 (Upgrades, Healing) | Day 1 Install Only | Day 1 Install Only |
| **Declarative Realm Management** | **Yes** (via `KeycloakRealm` CRD) | No | No |
| **Observability Integration** | Auto-creates `ServiceMonitor` for Prometheus | Manual Configuration | Manual Configuration |
| **Production Upgrade Guarantees** | Built-in graceful rolling updates | Manual `helm upgrade` with careful tuning | Manual `helm upgrade` with careful tuning |
| **Long-Term Open Source Viability** | **High** | **High** | **None** (Paywalled images) |

### Licensing Deep Dive: Why Open Source Sustainability Matters

Keycloak itself is licensed under **Apache 2.0**—a permissive open-source license with no restrictions. The official container images (`quay.io/keycloak/keycloak`) and the official Operator are also Apache 2.0. This creates a fully open-source, vendor-neutral deployment stack.

The Bitnami licensing change represents a cautionary tale about dependency risk. For years, Bitnami built goodwill by providing high-quality, well-maintained charts and images. But a corporate acquisition changed the business model overnight, creating a "licensing time bomb" set to detonate in mid-2025. Any infrastructure built on Bitnami charts will face a forced migration or subscription requirement.

For Planton—an open-source IaC framework—building abstractions on Bitnami would expose every user to this risk. The only sustainable path is to standardize on the official Apache 2.0 stack or community-driven alternatives like Codecentric that explicitly commit to open-source principles.

## Planton's Approach: Emulating the Operator's Declarative Surface

Planton's `KeycloakKubernetes` resource is designed to mirror the clean, declarative API surface of the Official Keycloak Operator's CRD, not the sprawling `values.yaml` of a Helm chart.

### Why This Matters: Day 2 vs. Day 1 Philosophy

Helm charts are optimized for the "Day 1" problem: parameterizing installation. They expose hundreds of configuration fields because they're trying to serve every possible use case through `values.yaml` overrides. This creates cognitive overload—users must understand Keycloak internals, Kubernetes networking, and Helm templating to achieve a production deployment.

Operators are optimized for the "Day 2" problem: ongoing lifecycle management. They hide the 80% of complexity that represents implementation details (which port Infinispan uses, how JGroups is configured, which probes to use) and expose only the 20% of essential fields that define your deployment's identity (replicas, hostname, database, TLS).

Planton follows this philosophy: **the abstraction should make strong, production-ready decisions on behalf of the user, exposing only the fields that genuinely vary between environments.**

### The "80/20" API Surface

The minimal, production-ready configuration for `KeycloakKubernetes` includes:

1. **`replicas` (int):** How many Keycloak nodes to run. Three or more is recommended for high availability.

2. **`adminCredentialsSecret` (string):** The name of a Kubernetes Secret containing the initial admin username and password. Credentials must never be in plain text.

3. **`database` (object):** The most critical configuration block. If omitted, triggers a non-production embedded H2 database. For production:
   - `type`: `postgres`, `mysql`, `mariadb`, or `mssql`
   - `host`: The database server hostname
   - `port`: The database server port
   - `databaseName`: The name of the database (e.g., `keycloak`)
   - `credentialsSecret`: The name of a Kubernetes Secret with database username and password

4. **`hostname` (string):** The public-facing URL for Keycloak (e.g., `sso.mycompany.com`). Required for production mode.

5. **`tls` (object):**
   - `secretName`: The name of a `kubernetes.io/tls` Secret containing the certificate for the hostname. Omitting this implies HTTP (non-production only).

6. **`computeResources` (object):** CPU and memory requests/limits. Highly recommended for production to ensure pod stability.

### Example: Production Configuration

Here's what a production-ready `KeycloakKubernetes` resource looks like:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KeycloakKubernetes
metadata:
  name: keycloak-prod
spec:
  replicas: 3
  adminCredentialsSecret: keycloak-prod-admin-creds
  database:
    type: postgres
    host: "keycloak-prod-db.c1a2b3d4.us-east-1.rds.amazonaws.com"
    port: 5432
    databaseName: keycloak_prod
    credentialsSecret: keycloak-prod-db-creds
  hostname: "sso.mycompany.com"
  tls:
    secretName: sso-prod-tls
  computeResources:
    requests:
      cpu: "1"
      memory: "2Gi"
    limits:
      cpu: "2"
      memory: "4Gi"
  # Advanced features (optional):
  hostnameAdmin: "sso-admin.mycompany.com"
  monitoring:
    enabled: true
```

This configuration:
- Deploys three replicas for high availability
- Connects to an external, HA-managed PostgreSQL database (AWS RDS in this example)
- Exposes Keycloak at `sso.mycompany.com` with TLS
- Exposes the admin console at a separate, more restrictive hostname for security
- Automatically creates Prometheus metrics endpoints and `ServiceMonitor` resources

The fields you **don't** see are equally important: the API doesn't expose JGroups configuration, Infinispan cache tuning, probe definitions, or port mappings. These are implementation details that Planton handles internally using production best practices.

## Production Best Practices

### High Availability: It's About the Database, Not Just the Application

Most teams instinctively focus on application-layer HA—running multiple Keycloak replicas—while underinvesting in database HA. This is backwards.

Because modern Keycloak deployments use `jdbc-ping`, the database is not just the persistent data store—it's also the **cluster coordination mechanism**. If the database becomes unavailable, the entire Keycloak cluster loses the ability to discover nodes, form quorum, or replicate sessions. Your three-replica Keycloak deployment will fail just as hard as a single replica would.

**Recommended database HA patterns:**

- **Cloud-managed (preferred):** AWS RDS with Multi-AZ failover, Google Cloud SQL, or Azure Database for PostgreSQL. These services provide automated failover, backups, and scaling with minimal operational overhead.

- **Kubernetes-native (for on-premises):** Use a dedicated PostgreSQL operator like **CloudNativePG** or **Patroni**. These tools manage a distributed, highly-available PostgreSQL cluster within Kubernetes, handling automated failover, replication, and connection pooling. Do not use a single-pod PostgreSQL deployment for production.

### Security: TLS, Admin Access Control, and Secret Management

1. **TLS Termination:** All production traffic must be HTTPS. The most common pattern is **edge termination**—TLS is terminated at the Ingress controller, which communicates with Keycloak pods over HTTP within the cluster. This requires configuring Keycloak with `http.enabled=true` and `proxy.headers.xforwarded=true` so it understands it's behind a reverse proxy.

2. **Admin Console Isolation:** A critical yet often overlooked security practice is exposing the admin console on a separate hostname (e.g., `sso-admin.mycompany.com` vs. `sso.mycompany.com`). This allows you to apply strict access controls—IP whitelisting, VPN requirements, or mutual TLS—to administrative endpoints without impacting user-facing authentication flows.

3. **Secret Management:** All credentials (admin passwords, database passwords, TLS certificates) must be stored in Kubernetes Secrets and referenced by name. Never store credentials in plain text in ConfigMaps or resource specifications.

### Backup and Disaster Recovery: The Database Is the Source of Truth

Keycloak's disaster recovery strategy revolves around the database, which contains all users, realms, clients, roles, and persistent sessions.

**What to backup:**
1. **Database dumps:** Use standard tools like `pg_dump` for PostgreSQL. This should be automated with a Kubernetes `CronJob` that runs daily and stores dumps to an external object store (S3, GCS, etc.).

2. **Realm configuration exports:** Use the `kc.sh export` command to dump realm configurations as JSON files for version control and GitOps workflows.

**What NOT to rely on:** The "Partial Export" button in the Keycloak Admin Console is **not a backup**. The official documentation explicitly warns that this feature masks all secrets, omits users, and is unsuitable for disaster recovery or server migrations.

### Observability: Metrics, Logs, and Health Monitoring

Keycloak exposes comprehensive Prometheus-format metrics via the `/metrics` endpoint, including JVM statistics, cache hit rates, and per-endpoint performance data. When monitoring is enabled, Planton automatically creates a Prometheus `ServiceMonitor` resource for seamless integration with the Prometheus Operator.

Keycloak logs are emitted in structured JSON format, making them easily parseable by log aggregation tools like Loki, Fluent Bit, or Elasticsearch.

## Conclusion: Choosing Lifecycle Management Over Installation Convenience

The Keycloak deployment landscape has matured from "just run a Deployment" anti-patterns to sophisticated, production-grade lifecycle management patterns. The key insight is that **deploying Keycloak is easy; operating it reliably is hard**.

Helm charts solve the Day 1 problem elegantly, but they leave Day 2 operations—upgrades, scaling, healing, and configuration management—as manual, error-prone tasks. The Operator pattern, by contrast, encodes operational expertise into automated controllers, enabling zero-downtime upgrades, declarative realm management, and self-healing infrastructure.

Planton's `KeycloakKubernetes` resource follows this philosophy: it provides a clean, minimal API surface modeled on the Official Keycloak Operator's CRD, hiding implementation complexity while exposing only the fields that genuinely vary between environments. This approach prioritizes long-term operational excellence over short-term installation convenience.

By standardizing on the Apache 2.0-licensed official Keycloak stack and avoiding licensing risks like the Bitnami paywall, Planton ensures that your identity infrastructure remains open-source, sustainable, and free from vendor lock-in.

**The modern paradigm for Keycloak on Kubernetes isn't about choosing between Helm and Operators—it's about recognizing that production infrastructure requires lifecycle management, not just installation tooling.** Planton delivers that guarantee through a declarative, production-ready API that lets you focus on your application's identity needs, not the operational complexity of distributed state management.


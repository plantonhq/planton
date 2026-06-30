# Deploying OpenFGA on Kubernetes: The Authorization System That Scales Like Google's

## Introduction: From Zanzibar's Shadow to Production Reality

For years, Google's internal authorization system, Zanzibar, existed as something of a legend—a mythical architecture that powered fine-grained permissions across YouTube, Google Drive, and Google Cloud at unprecedented scale. The conventional wisdom was clear: this level of sophistication was reserved for hyperscalers with infinite engineering resources.

OpenFGA changes that narrative. As a CNCF-incubated open-source implementation inspired by Zanzibar's principles, it brings Relationship-Based Access Control (ReBAC) to any development team. But bringing sophisticated authorization to production requires more than just running a container—it demands understanding the deployment landscape, making informed architectural choices, and avoiding the subtle pitfalls that transform a powerful authorization engine into a liability.

This document maps the terrain of deploying OpenFGA on Kubernetes, from anti-patterns to production-ready solutions. If you're building multi-tenant SaaS, implementing complex resource hierarchies, or simply exhausting the limits of traditional RBAC, OpenFGA offers a path forward. The question isn't whether to deploy it, but how to deploy it correctly.

## What OpenFGA Solves: The Authorization Complexity Wall

Traditional Role-Based Access Control (RBAC) works beautifully—until it doesn't. The breaking point arrives when you need to answer questions like:

- "Can Alice view this document **because** she's a member of the team that owns the parent folder?"
- "Which resources can Bob access across our entire multi-tenant platform?"
- "Does Carol have permission **through** her organizational hierarchy?"

These are graph problems masquerading as permission checks. OpenFGA excels precisely where traditional RBAC becomes unwieldy:

- **Multi-tenant isolation**: Enforce strict data boundaries between tenants without complex query filters
- **Resource-level permissions**: Implement "Google Doc-style" sharing where permissions are per-resource, not per-role
- **Hierarchical structures**: Model parent-child relationships where permissions inherit and cascade
- **Dynamic relationships**: Grant and revoke access based on runtime relationships, not static role assignments

OpenFGA's model is deceptively simple: define **types** (like `document` or `organization`), **relations** (like `viewer` or `member`), and store **relationship tuples** (facts like "user:anne is editor of document:roadmap"). The system then answers one fundamental question with sub-100ms latency: "Does user U have relation R with object O?"

### OpenFGA vs. The Authorization Ecosystem

Understanding where OpenFGA fits requires appreciating what it **isn't**:

**vs. Traditional RBAC**: RBAC asks "What role am I?" OpenFGA asks "Am I related to this resource in the required way?" RBAC is a subset of what OpenFGA can model.

**vs. OPA (Open Policy Agent)**: This comparison matters. OPA is a general-purpose **policy engine** for Attribute-Based Access Control (ABAC). It evaluates Rego policies against arbitrary JSON context. OpenFGA is a specialized **relationship engine** for ReBAC. It queries a graph of stored relationship tuples.

The recommended pattern for sophisticated systems is to use both in complementary roles:
- **Use OPA** for infrastructure policies and context-aware rules: "Can this user access this resource **if** the request is from an allowed IP **and** it's during business hours?"
- **Use OpenFGA** for application-level permissions based on relationships: "Can this user access this document **if** they own the project containing it?"

## The Kubernetes Deployment Landscape: From Anti-Patterns to Production

Deploying OpenFGA on Kubernetes presents a solved problem—provided you navigate the landscape correctly. The maturity spectrum spans from fundamentally broken approaches to battle-tested, production-ready solutions.

### Level 0: The Anti-Pattern (Simple Deployment + In-Memory Store)

The path of least resistance—and maximum regret—is deploying OpenFGA as a simple Kubernetes Deployment using the default in-memory datastore.

**Why it's tempting**: It "just works" for initial exploration. No database setup, no schema migrations, instant gratification.

**Why it fails**: The in-memory store is ephemeral. Every pod restart—voluntary or involuntary—erases **all** authorization models and relationship tuples. You've built an authorization system with selective amnesia.

**Verdict**: Useful for a 30-minute local test. Catastrophic for anything beyond that. The in-memory store is not a deployment strategy; it's a cautionary tale.

### Level 1: Manual Kubernetes Manifests (Heroic Effort, Brittle Outcome)

A step up from chaos is handcrafting the complete manifest stack: Deployment, Service, Secret (for datastore credentials and API keys), ConfigMap, and—critically—a Kubernetes Job to run the `openfga migrate` command before server pods start.

**What it solves**: Full control over every resource. No abstraction layers obscuring what's happening.

**What it costs**: Operational complexity, steep learning curve, error-prone updates, and the burden of maintaining intricate YAML across environments and OpenFGA versions.

**Verdict**: Viable for teams with deep Kubernetes expertise and a high tolerance for manual orchestration. Most teams have better ways to spend their engineering time.

### Level 2: The Official Helm Chart (Production Standard)

The **officially recommended and production-ready method** is deploying OpenFGA via the official Helm chart maintained in the `openfga/helm-charts` repository and published on Artifact Hub.

**Why it's the standard**:

1. **Automated schema management**: The chart solves the hardest problem—ensuring the database schema is initialized and migrated **before** server pods start. This is orchestrated via pre-install/pre-upgrade hooks running the `openfga migrate` command as a Kubernetes Job.

2. **First-class datastore integration**: The `values.yaml` provides clean interfaces for both bootstrapping in-cluster databases (Bitnami PostgreSQL/MySQL for development) and connecting to external managed databases (the production pattern).

3. **Production features out-of-the-box**: Configurable replica counts, ingress support, authentication (preshared keys), and monitoring (Prometheus metrics).

4. **Active maintenance**: Officially supported by the OpenFGA project with regular updates aligned to OpenFGA releases.

**A note on deprecated alternatives**: An older community chart (`alexandrebrg/openfga-helm`) exists but is deprecated. Always use the official chart for new deployments.

**Verdict**: This is the de facto standard. Attempting to build a custom operator or manually script manifests duplicates solved problems and introduces unnecessary risk.

### Level 3: The Kubernetes Operator (Emerging, Not Ready)

A Kubernetes Operator (`ZEISS/openfga-operator`) exists in the ecosystem. The operator pattern offers powerful lifecycle management capabilities—when mature.

**Current reality**: As of this writing, the operator is at `v0.0.1` with sparse documentation. Critical operational details—how it handles schema migrations, secret management, high availability configurations—remain unclear from available materials.

**Verdict**: Adopting this operator as a foundation for production infrastructure introduces unacceptable dependency risk. Monitor its maturity, but the official Helm chart remains the only battle-tested deployment method.

### Container Images: Official vs. Hardened

- **Official image**: `openfga/openfga` (Docker Hub) is Apache-2.0 licensed and should be the default.
- **Hardened alternative**: Chainguard provides minimal, distroless-style images built daily with significantly reduced attack surface (no shell, no package manager).

Production-conscious infrastructure should default to the official image but expose `image.repository` and `image.tag` fields to allow substituting hardened images.

## The Datastore: Not a Cache, But the Source of Truth

Here's a critical architectural distinction: In Google's Zanzibar, the authorization system acts as a replicated **cache**—the source of truth lives elsewhere in application services. OpenFGA's design acknowledges this pattern is impractical for most teams.

**In typical OpenFGA deployments, the datastore IS the source of truth.** This fundamentally elevates its importance. The datastore isn't supporting infrastructure; it's business-critical stateful storage. Its integrity, high availability, and disaster recovery strategy are primary architectural concerns.

### PostgreSQL vs. MySQL: A Clear Preference

OpenFGA officially supports both PostgreSQL and MySQL as production datastores. Both work. But the documentation, advanced features (streaming replication, read-replica configuration), and community focus reveal a clear preference for **PostgreSQL**.

**Recommendation**: Default to PostgreSQL. Support MySQL for compatibility, but make PostgreSQL the path of least resistance.

### Connection Patterns: Development vs. Production

**Development/Test**: The Helm chart can bootstrap a Bitnami PostgreSQL or MySQL instance **within** the cluster. This is convenient for local development and CI pipelines. It is **not** a production pattern.

**Production**: The mandated pattern is an external, managed database service:
- AWS RDS (Multi-AZ for high availability)
- Google Cloud SQL (with automated backups and point-in-time recovery)
- Azure Database for PostgreSQL

OpenFGA has **no native backup or disaster recovery logic**. It is 100% reliant on the underlying datastore. This isn't a limitation—it's a design decision that delegates data durability to specialized database services.

### The Security Anti-Pattern: Connection Strings with Embedded Passwords

Documentation examples often show the full datastore URI, including the password, as a single environment variable:

```
postgres://postgres:password@postgres:5432/postgres?sslmode=disable
```

Storing this in a ConfigMap or directly in an IaC resource spec is a **severe security anti-pattern**—credentials exposed in plain text.

**The secure pattern**: OpenFGA configuration allows the password to be provided via a separate environment variable (`OPENFGA_DATASTORE_PASSWORD`) that overrides any password in the URI. The Helm chart supports referencing existing Kubernetes Secret objects.

**Implication for API design**: An IaC API must **not** expose a single `datastore_uri` field. Instead, enforce security by decomposing configuration into a structured object:

- `datastore.engine`: `"postgres"` or `"mysql"`
- `datastore.host`: Database hostname
- `datastore.port`: Database port
- `datastore.database`: Database name
- `datastore.user`: Database user
- `datastore.password_secret_ref`: Reference to a Kubernetes Secret

The controller fetches the secret and injects the value securely into the `OPENFGA_DATASTORE_PASSWORD` environment variable.

### High Availability and Read Scaling

OpenFGA server processes are stateless. High availability comes from running multiple replicas (minimum 3 in production) that all connect to a single, shared, highly-available database.

For read-heavy workloads, OpenFGA supports explicit read-replica configuration via the `--datastore-secondary-uri` flag:

- **Writes** and **high-consistency reads** (requests using the `HIGHER_CONSISTENCY` flag) route to the primary database
- **Standard-consistency reads** (the default) route to the read replica, reducing primary database load

An IaC API should expose an optional `datastore.read_replica_credentials` block for this production optimization pattern.

### Common Datastore Pitfalls

1. **Unconfigured connection pooling**: Failing to tune `OPENFGA_DATASTORE_MAX_OPEN_CONNS` and `OPENFGA_DATASTORE_MAX_IDLE_CONNS` leads to connection exhaustion under load.

2. **MySQL URI missing `?parseTime=true`**: This query parameter is required by the MySQL driver and its absence causes cryptic errors.

3. **Network latency**: Deploying OpenFGA servers in a different datacenter or network zone from the database introduces latency on every authorization check. Colocation is critical.

## Production Architecture: Beyond Just Running Pods

A production-grade deployment requires more than functional pods—it demands resilience, security, and observability.

### Scaling Horizontally with HPA

As a stateless service, OpenFGA is an ideal candidate for Kubernetes HorizontalPodAutoscaler (HPA). Configure the HPA to monitor CPU utilization and automatically scale pod count to match demand.

**Recommended configuration**:
- Minimum replicas: 3 (for high availability)
- Maximum replicas: 10 (or higher based on load testing)
- Target CPU utilization: 80%

### High Availability: Replicas, PDBs, and Anti-Affinity

Running multiple replicas isn't enough. True high availability requires:

1. **PodDisruptionBudgets (PDBs)**: Protect the deployment from voluntary disruptions (node drains, cluster maintenance) by ensuring a quorum of pods always remains available.

2. **Pod Anti-Affinity Rules**: Spread pods across failure domains:
   - Across nodes (using `kubernetes.io/hostname`)
   - Across availability zones (using `topology.kubernetes.io/zone`)

This ensures that a single node failure or zone outage doesn't compromise the authorization system.

### Security: Authentication, Network Policies, and TLS

**API Authentication**: Running an unauthenticated OpenFGA server in production is a severe security risk. The recommended method is **preshared keys** (bearer tokens). The IaC API must support this via a secret reference (e.g., `authn.preshared.keys_secret_ref`).

**Network Policies**: For a zero-trust security posture, automatically create Kubernetes NetworkPolicy resources:
- **Ingress**: Allow traffic only from authorized application namespaces and the Prometheus namespace (for metrics scraping)
- **Egress**: Allow traffic only to the datastore (e.g., PostgreSQL port 5432) and, if configured, an OpenTelemetry collector

**Transport Security (TLS)**:
- **External (Ingress)**: Terminate TLS at a Kubernetes Ingress or Gateway—the standard pattern for HTTPS access
- **Internal (mTLS)**: OpenFGA can serve gRPC and HTTP over TLS directly for mutual TLS between services—an advanced but powerful pattern

**Disable the Playground**: The OpenFGA Playground is a useful development tool. It's also a security risk in production. Always set `playground.enabled: false` for production deployments.

### Monitoring: Prometheus and OpenTelemetry

OpenFGA is built with first-class observability:

**Prometheus Metrics**: OpenFGA natively exposes a Prometheus `/metrics` endpoint (default port 2112). For zero-configuration monitoring, an IaC controller should detect if the Prometheus Operator is installed in the cluster and automatically create a ServiceMonitor resource.

**OpenTelemetry Tracing**: Native support for OpenTelemetry provides deep visibility into authorization check latency and resolution paths. Expose a simple configuration field for the OTel collector endpoint to enable this.

### Performance Tuning: Caching and Consistency Trade-offs

OpenFGA includes a built-in in-memory cache for Check requests to minimize latency. This creates a consistency choice:

- **`MINIMIZE_LATENCY` (default)**: Serves requests from the cache, which may be slightly stale
- **`HIGHER_CONSISTENCY`**: Bypasses the cache and queries the database directly for read-after-write consistency

Most applications use the default. Mission-critical checks (e.g., "Does this user have permission to transfer funds?") might explicitly request `HIGHER_CONSISTENCY`.

## The Planton Choice: A Secure-by-Default Wrapper

Planton's `OpenFgaKubernetes` API is designed as a **declarative, type-safe wrapper** around the official OpenFGA Helm chart. This isn't reinventing the wheel—it's making the wheel accessible while enforcing security best practices.

### Design Principles

1. **Leverage the official Helm chart**: The chart solves complex orchestration (schema migrations, lifecycle hooks) in a battle-tested way. Planton renders and applies this chart behind the scenes.

2. **Enforce secure-by-default configuration**: The API **omits** insecure patterns (like a single `datastore_uri` with embedded passwords) and **mandates** secure patterns (structured credentials with secret references).

3. **Apply the 80/20 principle**: Expose the 20% of configuration fields that 80% of users need at the top level. Hide expert-level tuning parameters in an `advanced` block.

4. **Automate production necessities**: The controller automatically generates PodDisruptionBudgets, Pod Anti-Affinity rules, and ServiceMonitor resources (when Prometheus Operator is detected)—features users shouldn't have to remember to configure.

### The 80/20 Configuration Model

**Essential fields** (top-level, always visible):
- `replica_count`: Number of server pods
- `image.repository` and `image.tag`: Container image configuration
- `datastore.engine`: `"postgres"`, `"mysql"`, or `"memory"` (dev only)
- `datastore.credentials`: Structured object with `host`, `port`, `database`, `user`, `password_secret_ref`
- `datastore.read_replica_credentials`: Optional read-replica configuration
- `authn.preshared.keys_secret_ref`: API authentication secret reference
- `playground.enabled`: Developer playground toggle (must be `false` in production)
- `ingress`: Standard Kubernetes ingress configuration
- `monitoring.service_monitor.enabled`: Prometheus integration toggle
- `resources`: CPU and memory requests/limits
- `autoscaling.hpa`: HorizontalPodAutoscaler configuration

**Advanced fields** (hidden under `advanced` block):
- `datastore.tuning`: Connection pool tuning (`max_open_conns`, `max_idle_conns`)
- `scheduling`: Custom affinity, tolerations, node selectors (overriding sane defaults)
- `tracing.otel.exporter_endpoint`: OpenTelemetry collector endpoint

### Configuration Profiles by Environment

**Development**:
```yaml
replica_count: 1
datastore:
  engine: memory
authn:
  preshared:
    keys_secret_ref: ""  # Auth disabled
playground:
  enabled: true
resources:
  requests:
    cpu: 100m
    memory: 128Mi
```

**Production**:
```yaml
autoscaling:
  hpa:
    enabled: true
    min_replicas: 3
    max_replicas: 10
    target_cpu_utilization: 80
datastore:
  engine: postgres
  credentials:
    host: production-postgres.us-east-1.rds.amazonaws.com
    port: 5432
    database: openfga
    user: openfga
    password_secret_ref: openfga-prod-db-creds#password
  read_replica_credentials:
    host: production-postgres-replica.us-east-1.rds.amazonaws.com
    # ... (same structure as credentials)
authn:
  preshared:
    keys_secret_ref: openfga-prod-api-keys#bearer-tokens
playground:
  enabled: false
ingress:
  enabled: true
  host: openfga.example.com
  tls_secret_name: openfga-prod-tls
monitoring:
  service_monitor:
    enabled: true
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 2Gi
```

The controller implicitly generates PodDisruptionBudgets and Pod Anti-Affinity rules for production profiles.

## Operational Best Practices: The Day-2 Reality

Deploying OpenFGA is the beginning, not the end. Operational maturity requires understanding the authorization model lifecycle and testing strategies.

### Authorization Models: Immutable and Versionable

Authorization Models in OpenFGA are **immutable**. When you "update" a model, OpenFGA creates a new model with a new `authorization_model_id`. This isn't a limitation—it's a feature enabling safe, GitOps-driven workflows:

1. **Store models in Git**: Authorization models (in OpenFGA DSL format) live in version control
2. **Test in CI**: Use the OpenFGA CLI's `fga model test` command to validate models against test assertions before deployment
3. **Deploy via CD**: On merge, a CI/CD pipeline writes the new model to OpenFGA, receiving a new `authorization_model_id`
4. **Deliberate rollout**: Production services **pin** the `authorization_model_id` they use. Deploying a new model doesn't break production—services upgrade to the new model ID deliberately

This pattern prevents the "oops, I broke production authorization" scenario.

### Testing Authorization Rules

OpenFGA provides a robust testing framework via `.fga.yaml` files that define:
- The authorization model to test
- A set of relationship tuples to seed the test environment
- Assertions to validate (e.g., "user:anne **should** have `viewer` relation with `document:roadmap`")

This enables treating authorization logic as testable code, not untestable configuration.

### Multi-Environment Strategies

Teams deploying OpenFGA face an architectural decision: physical isolation (separate deployments per environment) or logical isolation (shared deployment, separate Stores)?

**Pattern A: Physical Isolation**  
Provision separate `OpenFgaKubernetes` resources for dev, staging, and production—each with its own database. Maximum isolation, higher cost and operational overhead.

**Pattern B: Logical Isolation (Recommended for Platform Teams)**  
Provision one highly-available `OpenFgaKubernetes` resource. Use OpenFGA's Store API to create multiple logically-isolated Stores (`dev_store`, `staging_store`, `prod_store_a`). Since a Store is the container for models and tuples, this provides complete logical separation on shared, cost-optimized infrastructure.

Pattern B is more cloud-native and cost-effective for platform teams supporting multiple applications or environments.

## Integration Ecosystem: SDKs and Adoption Patterns

OpenFGA's value is realized when applications integrate it seamlessly. The project maintains official SDKs for Go, Python, Java, Node.js, and .NET—providing idiomatic client libraries across the language ecosystem.

### Common Integration Patterns

**API Middleware**: The standard pattern is middleware in the application's API layer that intercepts requests, extracts user and resource context, and performs an OpenFGA Check call before business logic execution.

**Shadow Mode**: A powerful migration strategy for legacy authorization systems. Configure the application to call **both** the old authorization logic **and** OpenFGA, logging discrepancies without enforcing the new system. This validates a new authorization model with real production traffic, risk-free.

**API Gateway Integration**: Some architectures enforce authorization at the edge by integrating OpenFGA Check calls directly into an API Gateway, preventing unauthorized requests from reaching application services.

## Conclusion: Authorization as a First-Class Infrastructure Component

The maturity of OpenFGA's Kubernetes deployment ecosystem—particularly the official Helm chart—signals a broader shift: fine-grained authorization is no longer a proprietary feature of hyperscalers. It's accessible, production-ready, and deployable by any team.

The deployment landscape has clear winners and clear pitfalls. The in-memory store is a trap. Manual manifests are heroic but brittle. The official Helm chart is the production standard. Planton's approach—wrapping this chart with a secure-by-default, 80/20 API—respects what's been solved while adding the guardrails that prevent common security and operational mistakes.

Deploying OpenFGA correctly means treating the datastore as business-critical infrastructure, enforcing authentication, testing authorization models as code, and building observability from day one. When approached with this rigor, OpenFGA delivers on its promise: authorization that scales like Google's, without Google's infrastructure burden.

For deeper implementation guides on specific topics covered here, see:
- [Datastore Configuration Deep Dive](./datastore-guide.md) *(coming soon)*
- [Authorization Model Management Patterns](./model-management.md) *(coming soon)*
- [Production Security Checklist](./security-checklist.md) *(coming soon)*


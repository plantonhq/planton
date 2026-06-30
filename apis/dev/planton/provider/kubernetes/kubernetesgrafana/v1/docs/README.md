# Grafana on Kubernetes: The Evolution from Quick Start to Production

## Introduction

Grafana has become the de facto standard for observability dashboards in modern infrastructure. Its appeal is immediate: spin up a container, open a browser, drag some metrics onto a panel, and you have a beautiful, customizable dashboard in minutes. This "getting started" simplicity is Grafana's superpower—and its most dangerous trap.

The challenge isn't deploying Grafana to Kubernetes. The challenge is deploying it *correctly*—in a way that doesn't result in the "Day 1 success, Day 2 catastrophe" scenario that has frustrated countless teams. You spend hours crafting the perfect dashboards. Then a pod restart vaporizes everything. Panic ensues. Dashboards are recreated from memory. The cycle repeats.

This pattern is so common it deserves a name: **The Ephemeral Dashboard Disaster**.

The root cause? A fundamental mismatch between Grafana's default stateful behavior (it stores everything in an embedded SQLite database) and Kubernetes' default stateless pod management (containers are disposable cattle, not pets). Production-grade Grafana on Kubernetes requires bridging this gap through deliberate architectural choices about persistence, high availability, and lifecycle management.

This document surveys the deployment landscape, presents the evolution from anti-patterns to production-ready solutions, and explains why Planton chose a Kubernetes-native, operator-driven approach for the `GrafanaKubernetes` API.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern (Stateless Deployment)

**What it is:** A standard Kubernetes `Deployment` with no persistence configuration. Grafana uses its default embedded SQLite database, which lives in the container's ephemeral filesystem.

**What it solves:** Absolutely nothing beyond a demo. It gets Grafana running *right now* with zero configuration.

**The catastrophe:** Every time the pod restarts—due to a node failure, cluster upgrade, or simple redeployment—the container is destroyed. The SQLite database file vanishes. All dashboards, data sources, users, and settings are **permanently lost**. This is the #1 source of Grafana-related frustration in the Kubernetes ecosystem.

**Verdict:** Never use this in any environment where your work matters. This anti-pattern violates the most basic requirement of a production service: durability.

### Level 1: Deployment with PVC (Single-Instance Persistence)

**What it is:** A `Deployment` (replica: 1) with a `PersistentVolumeClaim` (PVC) attached. The SQLite database file is stored on a durable volume (e.g., AWS EBS, GCE PD).

**What it solves:** The Ephemeral Dashboard Disaster. Your dashboards survive pod restarts, node failures, and redeployments. This is the minimal viable production configuration for a **single-replica, non-HA** instance.

**What it doesn't solve:** High availability. With `replicas: 1`, your Grafana instance is a single point of failure. If the pod crashes, dashboards are down until Kubernetes reschedules it.

**Verdict:** Acceptable for dev/staging environments and small teams where a few minutes of downtime during pod restarts is tolerable. Not suitable for production systems that demand 99.9%+ uptime.

### Level 2: The StatefulSet Trap (Don't Go Here)

**What it is:** A `StatefulSet` with `replicas: 2` and `volumeClaimTemplates`, attempting to achieve HA through multiple Grafana instances.

**Why it seems logical:** StatefulSets are designed for stateful applications. Grafana is stateful (it has a database). Ergo, use a StatefulSet. Right?

**The catastrophic flaw:** A StatefulSet with `replicas: 2` creates two pods (`grafana-0`, `grafana-1`) with **two separate, independent PVCs**. This means `grafana-0` has its own SQLite database, and `grafana-1` has a completely different, empty SQLite database. You haven't created a HA cluster—you've created two isolated Grafana instances that don't know about each other. A dashboard created on `grafana-0` doesn't exist on `grafana-1`.

**Verdict:** This is an architectural dead-end. StatefulSets are the **wrong tool** for production Grafana HA.

### Level 3: The ReadWriteMany PVC Trap (Also Don't Go Here)

**What it is:** A `Deployment` with `replicas: 2` and a *single*, shared PVC using `ReadWriteMany` (RWX) access mode (e.g., NFS, AWS EFS).

**Why it seems logical:** Multiple pods, one shared database file. Everyone sees the same dashboards.

**The catastrophic flaw:** SQLite **does not support concurrent access**. Even with an RWX volume, SQLite's database file cannot be safely read and written by multiple processes simultaneously. This architecture will immediately fail with "database is locked" errors and can lead to data corruption.

**Verdict:** Another architectural dead-end. The problem isn't the volume; it's the application.

### Level 4: The Production Solution (Deployment + External Database)

**What it is:** A standard `Deployment` with `replicas: 2` (or more) for high availability. Grafana is configured to use an **external, shared database**—typically PostgreSQL or MySQL (e.g., AWS RDS, Google Cloud SQL, or a cluster-local PostgreSQL operator).

**The paradigm shift:** In this architecture, the Grafana pods themselves become **100% stateless**. All state (dashboards, users, config) is stored in the external database. The pods are interchangeable, disposable "cattle" that simply act as query and rendering engines.

**What it solves:** Everything. This is the only viable, production-grade architecture for Grafana HA on Kubernetes. It provides:
- **High availability:** Multiple replicas behind a load balancer. If one pod crashes, others continue serving traffic.
- **Zero data loss:** All state is in the durable, backed-up external database.
- **Horizontal scalability:** Add more replicas to handle increased load.

**Configuration highlights:**
- The `persistence.enabled` flag (in Helm charts) is set to `false`. No PVC is attached to the pods.
- Grafana's `grafana.ini` config file is set to point to the external database (via the `[database]` section).
- Database credentials are managed securely using Kubernetes Secrets.

**Verdict:** This is the gold standard. If you need production-grade uptime and performance, this is the only correct architecture.

## Provisioning Methods: From Manual to Declarative

Kubernetes is an API-driven, declarative platform. The best way to deploy Grafana should align with this philosophy. Let's survey the options.

### Method 1: Manual kubectl Manifests

**What it is:** Writing YAML files by hand for Deployment, Service, ConfigMap, Ingress, etc., and applying them with `kubectl apply`.

**Verdict:** This is the 2018 way. It's brittle, error-prone, and doesn't scale. Avoid this except for learning exercises.

### Method 2: Helm Charts (The "DIY Production" Approach)

**Official Chart:** `grafana/grafana` (maintained by Grafana Labs)

**What it provides:** A comprehensive, highly configurable Helm chart with an extensive `values.yaml` file. You can control every aspect of the deployment: persistence, ingress, resources, `grafana.ini` config, dashboard/datasource provisioning, and more.

**Key features:**
- **grafana.ini as YAML:** The chart converts YAML in `values.yaml` into the traditional `.ini` format, allowing declarative control over advanced features like external databases, authentication (LDAP, OAuth), and HA settings.
- **Sidecar pattern:** An optional sidecar container watches for ConfigMaps labeled with `grafana_dashboard: "1"`. This enables a "Dashboard-as-Code" workflow where dashboards stored in Git as ConfigMaps are automatically provisioned.

**When to use it:** Best for teams that need a standalone, highly-customized Grafana instance and are comfortable managing a detailed `values.yaml` file. This is the "power user" option.

**Bitnami Variant:** `bitnami/grafana`

The Bitnami chart offers hardened, minimal, non-root container images built on Photon OS. It's security-first, making it mandatory in some restricted environments (e.g., OpenShift). However, it uses a **different `values.yaml` schema** than the official chart, which can create friction since community documentation overwhelmingly references the official schema.

**Note:** As of 2025, Bitnami is transitioning production-ready charts and images to a commercial "Bitnami Secure Images" offering, though the open-source code remains available.

### Method 3: Integrated Stack (kube-prometheus-stack)

**What it is:** The `kube-prometheus-stack` Helm chart bundles a complete monitoring solution: Prometheus, Alertmanager, and Grafana. It *includes* the official `grafana/grafana` chart as a sub-chart dependency.

**The "batteries-included" value:** When enabled, Grafana is pre-configured with:
- A Prometheus data source pointing to the Prometheus instance the stack just installed
- Dozens of essential, pre-built dashboards for cluster monitoring (nodes, pods, API server, etc.)

**When to use it:** This is the **de facto standard** for the 80% use case: teams whose primary goal is comprehensive, turnkey cluster monitoring with minimal configuration. It's wildly popular for good reason.

**Limitation:** It's designed for a single, cluster-scoped Grafana instance. It's less suitable for platform teams that need to manage multiple Grafana instances (e.g., one per development team) or want to treat Grafana as a general-purpose visualization platform beyond cluster monitoring.

### Method 4: Grafana Operator (The Kubernetes-Native Approach)

**What it is:** The `grafana/grafana-operator` (maintained by Grafana Labs) is a Kubernetes Operator that extends the Kubernetes API with Custom Resource Definitions (CRDs) for managing Grafana instances and their resources.

**The paradigm shift: Server-side vs. Client-side**

- **Helm:** A client-side tool. You run `helm upgrade`, it templates YAML and applies it to the cluster, then exits. If someone manually deletes the Grafana Service, Helm won't know or care until the next `helm upgrade`. It's *imperative* and fire-and-forget.

- **Operator:** A server-side, in-cluster controller. You `kubectl apply` a high-level `Grafana` Custom Resource (CR). The Operator, which runs continuously in the cluster, sees this CR and works to make reality match the desired state. It creates the Deployment, Service, ConfigMaps. If someone deletes the Service, the Operator's control loop immediately detects the drift and recreates it. This is *declarative* and continuously reconciled.

The Operator model is a superset of Helm's capabilities. It moves management logic from a client-side binary to a server-side, cluster-aware controller—the standard "Kubernetes-native" pattern for complex applications.

**CRDs provided:**
- **Grafana:** Defines the Grafana instance (deployment, service, persistence, ingress, config)
- **GrafanaDashboard:** A first-class Kubernetes resource representing a dashboard. The JSON can be embedded, referenced from a ConfigMap, or pulled from grafana.com.
- **GrafanaDataSource:** A first-class Kubernetes resource representing a data source (e.g., Prometheus, Loki)
- **GrafanaFolder, GrafanaAlertRuleGroup, GrafanaContactPoint:** Additional resources for comprehensive "Grafana-as-Code" management

**When to use it:** This is the architecturally superior choice for:
1. **Platform teams** building an observability platform or multi-tenant environment
2. **Organizations that demand true, declarative, GitOps-driven management** of Grafana instances and all their resources (dashboards, datasources)
3. **Projects building an abstraction layer** on top of Grafana—like Planton

**Production readiness:** The Operator is actively developed, officially promoted by Grafana Labs, and considered production-ready. It integrates seamlessly with GitOps tools like ArgoCD and Flux.

## Dashboard-as-Code: How to Manage Dashboards Declaratively

One of Grafana's greatest strengths—its web-based dashboard editor—is also a challenge for production: point-and-click configuration is inherently undocumented, non-reproducible, and non-auditable. "Dashboard-as-Code" solves this by managing dashboard JSON in Git and provisioning it declaratively.

### Pattern 1: Helm Sidecar (ConfigMap Watcher)

**How it works:** Enable `sidecar.dashboards.enabled: true` in the Helm chart's `values.yaml`. Then create ConfigMaps containing dashboard JSON with a specific label (e.g., `grafana_dashboard: "1"`). A sidecar container in the Grafana pod watches for these labeled ConfigMaps and automatically provisions the dashboards.

**Pros:** Simple, built into the most popular Helm charts (official chart and kube-prometheus-stack).

**Cons:** It's an "indirect" pattern. The dashboard is just data inside a generic ConfigMap, identified by a "magic" label. Syncing can be slow, sometimes requiring a pod restart.

### Pattern 2: Grafana Operator (CRD Watcher)

**How it works:** Create a `GrafanaDashboard` manifest. This is a first-class Kubernetes resource. The Operator watches for these CRs and uses the Grafana API to create/update the dashboards.

**Pros:** This is a true, declarative, Kubernetes-native model. It integrates *perfectly* with GitOps tools like ArgoCD and Flux, which are designed to manage first-class Kubernetes resources, not labeled ConfigMaps.

**Cons:** Requires the Operator to be installed and running.

### Pattern 3: Inline in values.yaml (Anti-pattern)

**How it works:** Paste dashboard JSON directly into the Helm chart's `values.yaml`.

**Verdict:** This is acceptable for a single, simple, default data source. For dashboards, it's an unmanageable anti-pattern. The `values.yaml` file becomes thousands of lines long, unreadable, and brittle.

**Recommendation:** Pattern 2 (Operator/CRD) is the modern, robust, and most maintainable pattern. Pattern 1 (Sidecar) is a viable and extremely common alternative, especially for users of kube-prometheus-stack.

## Planton's Architectural Choice

Planton is an open-source, API-driven, multi-cloud IaC framework with protobuf-defined APIs. The `GrafanaKubernetes` API must align with this declarative, Kubernetes-native philosophy.

### Why the Grafana Operator?

**1. Philosophical alignment:** The Operator's model—managing infrastructure via first-class Kubernetes CRDs—is *identical* to Planton's approach. Both treat configuration as declarative, API-driven resources.

**2. Delegation of complexity:** By designing `GrafanaKubernetes` as a high-level facade that drives the underlying `Grafana` CRD, Planton delegates the complex, "Day 2" reconciliation logic to the official, community-vetted Grafana Operator. This is more robust and maintainable than embedding Helm chart templating logic.

**3. True GitOps support:** The Operator's CRDs (Grafana, GrafanaDashboard, GrafanaDataSource) integrate seamlessly with GitOps tools like ArgoCD and Flux. This enables a complete, declarative workflow where the Grafana *instance* and all its *resources* (dashboards, datasources) are managed in Git.

**4. Multi-tenancy and scalability:** For platform teams managing multiple Grafana instances (e.g., one per development team), the Operator's declarative, per-instance CRs are far more scalable than managing multiple Helm releases.

### What the GrafanaKubernetes API Provides

The `GrafanaKubernetes` protobuf API abstracts the Operator's `Grafana` CRD, providing a simplified, opinionated, production-ready-by-default interface. It focuses on the 80% use case with fields like:

- `replicas`: Control HA (1 for dev/staging, 3 for production)
- `adminPasswordSecretRef`: Secure credential management (no plain-text passwords)
- `ingress`: Network access (hostname, TLS certificate reference)
- `resources`: CPU/memory requests and limits
- `persistence`: For single-replica instances (PVC for SQLite)
- `externalDatabase`: For multi-replica HA (PostgreSQL/MySQL connection config)
- `defaultDataSource`: A simple "80% solution" for provisioning the most common data source (typically Prometheus)

The API enforces mutual exclusivity between `persistence` and `externalDatabase`, guiding users toward the correct architecture based on their HA requirements.

**What it omits:** Complex, edge-case features (LDAP/OAuth, custom plugins, full `grafana.ini` passthrough) are intentionally excluded from the minimal API to maintain simplicity and stability.

### Separation of Concerns: Instance vs. Resources

A mature "Grafana-as-Code" ecosystem separates the management of the **Grafana instance** (the application, its config, ingress) from its **resources** (dashboards, datasources). The `GrafanaKubernetes` API is responsible *only* for the instance. It does not manage lists of dashboards or data sources (beyond a single default datasource).

This decoupling prevents a bloated, unmaintainable API and avoids conflict with established GitOps patterns. Planton subsequently provides separate `GrafanaDashboard` and `GrafanaDataSource` protobuf specs that map 1:1 to the Operator's CRDs, enabling a complete, decoupled, API-driven solution.

## Production Best Practices

### Persistence and HA: The External Database Model

For production HA, the non-negotiable prerequisite is an external, shared database (PostgreSQL or MySQL). This can be a managed service (AWS RDS, Google Cloud SQL) or a self-hosted instance (via a PostgreSQL operator).

**Configuration checklist:**
- Set `persistence.enabled: false` (no PVC for SQLite)
- Configure `grafana.ini` (or the Operator's `config`) to connect to the external database
- Store database credentials in a Kubernetes Secret and reference them (never plain-text)
- Backup and recovery strategy focuses on the external database, not Kubernetes volumes

### Security: Layered Approach

**1. Admin password management:**
- **Anti-pattern:** Hardcoding `adminPassword: "my-password"` in `values.yaml` (leaks the password in Git and ConfigMaps)
- **Best practice:** Pre-create a Kubernetes Secret. The Grafana config references this Secret. This pattern is GitOps-friendly and allows the Secret's content to be managed by external systems (Vault, ExternalSecrets Operator).

**2. Database credentials ($__file provider):**
Grafana's config supports a special `$__file{...}` syntax. The recommended pattern:
1. Store the database password in a Kubernetes Secret
2. Mount this Secret as a file in the Grafana pod (e.g., `/etc/secrets/db-password`)
3. Configure `grafana.ini` to reference the file path: `password: "$__file{/etc/secrets/db-password}"`

This keeps the configuration declarative and auditable while keeping secrets secure.

**3. Ingress and TLS:**
All production Grafana instances must use HTTPS. Configure the Ingress resource with a `tls:` block referencing a Secret containing the certificate and key. This is typically automated using `cert-manager` with Let's Encrypt or an internal CA.

### Monitoring Grafana (Metamonitoring)

The monitoring system itself must be monitored. A broken Grafana instance creates a false sense of security.

**Common pitfalls:**
- Losing dashboards (solved by persistence)
- Resource exhaustion (solved by setting resource requests/limits)
- Misconfigured data sources
- Database connection failures (HA only)

**Key metrics to watch:**
- `kube_pod_container_status_restarts_total` (for Grafana pods)
- `container_cpu_usage_seconds_total` and `container_memory_working_set_bytes` (against requests/limits)
- Grafana's own Prometheus metrics (dashboard load times, query errors, datasource health)

### GitOps Workflows (ArgoCD / Flux)

**Operator-based pattern (recommended):**
1. **Bootstrap:** An ArgoCD/Flux Application installs the Grafana Operator (once)
2. **Instance:** A second Application syncs the `GrafanaKubernetes` (Planton) resource from Git, which creates the underlying `Grafana` CRD
3. **Resources:** The same or a third Application syncs all `GrafanaDashboard` and `GrafanaDataSource` CRDs from a `/dashboards` directory in Git

This is the most robust, scalable, and philosophically-aligned solution for platform teams. It allows the Grafana instance and all its resources to be managed declaratively in one Git repository by the same GitOps tools.

## Conclusion: State Management as the Central Challenge

The evolution from "quick start" to production-ready Grafana on Kubernetes is fundamentally a story about **state management**. Grafana's default stateful behavior (embedded SQLite) is mismatched with Kubernetes' default stateless pod model. Production deployment isn't about choosing the right controller (Deployment vs. StatefulSet); it's about understanding Grafana's conditional state model:

- **Single-replica:** The pods are stateful. Use a Deployment with a PVC.
- **Multi-replica (HA):** Externalize the state to a shared database. The pods become stateless. Use a Deployment with no PVC.

The provisioning method—Helm vs. Operator—reflects a choice between client-side, imperative tools and server-side, declarative, Kubernetes-native controllers. For an API-driven platform like Planton, the Operator is the clear architectural choice. It provides continuous reconciliation, native GitOps integration, and true "Day 2" lifecycle management.

By designing the `GrafanaKubernetes` API as a high-level facade over the Grafana Operator's CRDs, Planton provides a simplified, opinionated, production-ready-by-default experience while delegating complexity to a battle-tested, community-maintained controller. This is how you deploy Grafana on Kubernetes the right way—once.


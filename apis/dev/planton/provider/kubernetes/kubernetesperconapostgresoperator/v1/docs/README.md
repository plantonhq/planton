# Deploying PostgreSQL Operators on Kubernetes: The Control Plane Perspective

## Introduction: The Operator vs. The Database

When teams first explore PostgreSQL on Kubernetes, they often conflate two distinct layers:

1. **The Operator** (the control plane): Software that manages PostgreSQL clusters
2. **The Database** (the data plane): The actual PostgreSQL instances running your workloads

This distinction matters because deploying the operator is fundamentally different from deploying databases. The operator is your database orchestrator—a piece of infrastructure software that extends Kubernetes with PostgreSQL-specific intelligence. Once installed, it watches for your custom resource definitions (CRDs) like `PerconaPGCluster` and autonomously creates, manages, and maintains PostgreSQL clusters on your behalf.

**Think of it this way**: The operator is the database administrator that never sleeps. You deploy the operator once per cluster (or namespace), and it manages dozens or hundreds of PostgreSQL databases throughout their lifecycles.

This guide focuses exclusively on deploying the **Percona Operator for PostgreSQL**—the control plane component. If you're looking for guidance on deploying PostgreSQL databases themselves, that's handled via the `PerconaPGCluster` custom resource after the operator is installed.

---

## Why Percona? The Licensing Clarity Advantage

The PostgreSQL operator ecosystem is crowded, with several production-grade options. Percona's strategic position hinges on one decisive advantage: **unambiguous open-source licensing**.

### The Container Image License Trap

Here's a critical lesson many teams learn the hard way: **open-source code doesn't guarantee open-source container images**.

Consider Crunchy Data's PGO (Postgres Operator). Its source code is Apache 2.0—fully open. But the official container images are governed by the "Crunchy Data Developer Program," which explicitly restricts production use without a paid commercial subscription. Your CI/CD pipeline will happily pull the images from the registry, your operator will deploy successfully, and you won't discover the licensing violation until much later.

### Percona's Solution: Fork and Liberate

Percona's operator is explicitly **forked from Crunchy PGO**. This isn't hidden—it's their value proposition. They took Crunchy's mature, battle-tested codebase and rebuilt the entire container image stack under truly permissive terms:

- **Operator source code**: Apache 2.0
- **Container images**: Freely available on Docker Hub with **no production restrictions**
- **PostgreSQL distribution**: PostgreSQL License (permissive, MIT-like)

This is a deliberate strategy to "liberate" the mature PGO architecture from commercial constraints. The result is a production-grade operator with the licensing clarity essential for open-source IaC frameworks.

**Critical nuance**: Percona's Red Hat Certified images (used for OpenShift OLM installations) are built on Red Hat Universal Base Images (UBI) and carry Red Hat's EULA restrictions. For pure open-source deployments, use the standard Docker Hub images at `docker.io/percona/kubernetes-percona-postgres-operator`.

---

## The Deployment Methods Spectrum

Deploying the Percona PostgreSQL Operator follows a familiar Kubernetes pattern, progressing from manual to fully declarative approaches.

### Method 1: kubectl and Raw Manifests (Quick Start Only)

**How it works**: Percona publishes a `bundle.yaml` that packages the operator Deployment, CRDs, ServiceAccount, and RBAC into a single file.

```bash
kubectl create namespace postgres-operator
kubectl apply --server-side \
  -f https://raw.githubusercontent.com/percona/kubernetes-percona-postgres-operator/v2.7.0/deploy/bundle.yaml \
  -n postgres-operator
```

**Verdict**: Effective for quick prototyping or evaluation, but a **lifecycle management anti-pattern** for production. Upgrading requires manually diffing new CRDs and deployment specs. Configuration changes require downloading, patching, and reapplying the bundle. This is operational toil masquerading as simplicity.

### Method 2: Helm (Recommended)

**How it works**: Percona maintains official Helm charts at `https://percona.github.io/percona-helm-charts/`. Critically, there are **two separate charts**:

1. **`percona/pg-operator`**: Installs the operator controller, CRDs, and RBAC. **This is what Planton's API targets.**
2. **`percona/pg-db`**: A helper chart that creates a `PerconaPGCluster` custom resource. This is for convenience—not the operator itself.

```bash
helm repo add percona https://percona.github.io/percona-helm-charts/
helm repo update
helm install my-operator percona/pg-operator \
  --namespace postgres-operator \
  --create-namespace \
  --set watchAllNamespaces=true \
  --set resources.requests.cpu=500m \
  --set resources.requests.memory=512Mi
```

**Why Helm is superior**:
- Manages the full lifecycle: install, upgrade, rollback, uninstall
- Templating enables environment-specific configuration via `values.yaml`
- Handles RBAC and CRD dependencies correctly
- Provides semantic versioning and release management

**Verdict**: This is the production-grade approach. IaC tools (Terraform, Pulumi) should wrap Helm, not reimplement raw YAML.

### Method 3: Kustomize (Overlay Pattern)

**How it works**: Use the raw `bundle.yaml` as a base and apply overlays to customize the operator Deployment (e.g., resource limits, node selectors, tolerations).

```yaml
# kustomization.yaml
resources:
  - https://raw.githubusercontent.com/percona/kubernetes-percona-postgres-operator/v2.7.0/deploy/bundle.yaml

patches:
  - target:
      kind: Deployment
      name: kubernetes-percona-postgres-operator
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/resources
        value:
          requests:
            cpu: 500m
            memory: 512Mi
```

**Verdict**: Useful for teams already invested in Kustomize for GitOps workflows. Less intuitive for configuration management than Helm's `values.yaml`, but functionally equivalent.

### Method 4: Operator Lifecycle Manager (OpenShift)

**How it works**: For OpenShift users, the operator is a **Red Hat Certified Operator** available on OperatorHub. OLM automates installation, RBAC setup, and version upgrades.

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: kubernetes-percona-postgres-operator
  namespace: postgres-operator
spec:
  channel: stable
  name: kubernetes-percona-postgres-operator
  source: certified-operators
  sourceNamespace: openshift-marketplace
```

**Verdict**: The standard pattern for OpenShift. OLM provides strong lifecycle guarantees but is platform-specific. Remember: use this only on OpenShift; for other platforms, stick with Helm.

### Method 5: IaC Integration (Terraform, Pulumi)

**The correct pattern**: Use the `helm_release` resource (Terraform) or `helm.v3.Release` (Pulumi) to deploy the official `percona/pg-operator` chart. This combines Terraform's state management with Helm's package management.

**Terraform example**:

```hcl
resource "helm_release" "percona_operator" {
  name       = "percona-operator"
  repository = "https://percona.github.io/percona-helm-charts/"
  chart      = "pg-operator"
  version    = "2.7.0"
  namespace  = "postgres-operator"
  create_namespace = true

  set {
    name  = "watchAllNamespaces"
    value = "true"
  }

  set {
    name  = "resources.requests.cpu"
    value = "500m"
  }
}
```

**The anti-pattern**: Hardcoding the entire 400+ line `bundle.yaml` into your IaC code (as demonstrated in some Percona blog posts). This is brittle and breaks on every patch release. **Never do this.**

---

## The 80/20 Configuration Rule

The `percona/pg-operator` Helm chart has dozens of configuration options. But production deployments consistently require only a small subset. Here's the essential configuration surface:

### Essential Fields (Always Configure)

**1. Tenancy Scope**

The most critical deployment-time decision: which namespaces should the operator watch?

- `watchAllNamespaces: true` → **Cluster-wide mode**: One operator manages all databases in all namespaces
- `watchAllNamespaces: false` (default) → **Namespace-scoped**: Operator only watches its own namespace
- `watchNamespace: "ns-a,ns-b"` → **Multi-namespace mode**: Operator watches a specific list

**Why this matters**: This determines the operator's RBAC permissions (ClusterRole vs. Role) and architectural footprint. Cluster-wide is operationally simpler but creates a single point of failure. Namespace-scoped provides isolation but creates management overhead.

**2. Resource Limits**

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1
    memory: 1Gi
```

**Why this matters**: The chart's default is `{}` (empty), which results in a **BestEffort QoS pod**—the first to be evicted under node pressure. For a production control plane component, this is unacceptable. Set requests and limits to achieve **Guaranteed** or **Burstable** QoS.

**3. Image Specification**

```yaml
image:
  registry: docker.io
  repository: percona/kubernetes-percona-postgres-operator
  tag: 2.7.0
```

**Why this matters**: Version pinning ensures reproducible deployments. Private registry support is essential for air-gapped or security-conscious environments.

### Common Customizations (Production)

**4. Telemetry**

```yaml
disableTelemetry: true
```

Most enterprises disable telemetry for privacy and security compliance.

**5. Scheduling**

```yaml
nodeSelector:
  node-pool: infra-services

tolerations:
  - key: infra-only
    operator: Exists
    effect: NoSchedule
```

Operators are control plane components. In production, they're typically scheduled onto dedicated "infra" or "system" node pools using taints and node selectors.

### Configuration You Can Skip

For a minimal API, these fields are rarely customized:

- `imagePullPolicy` (defaults to `Always`)
- `logStructured` / `logLevel` (for debugging)
- `podAnnotations` (for Prometheus/Istio integrations)
- `rbac.create` (defaults to `true`; only set to `false` in highly controlled environments)

---

## Production Best Practices

### Prerequisites (Don't Skip These)

The most common Day 1 failure: **missing or misconfigured storage**.

The operator creates `PersistentVolumeClaims` for database storage. This requires:

1. A functioning **Container Storage Interface (CSI) driver** for your platform
2. A default or specified **StorageClass**

**Platform-specific gotchas**:

- **GKE**: GCE Persistent Disk CSI driver is typically pre-installed. Usually "just works."
- **EKS**: Amazon EBS CSI driver is **not installed by default**. You must add it as a cluster prerequisite, or all PVCs will remain stuck in `Pending` state.
- **AKS**: Requires Azure Disk CSI driver and a `StorageClass` like `managed-csi-premium`.

### Multi-Tenancy Patterns

**Cluster-Wide (One Operator)**

- **Pros**: Single control plane, minimal resource overhead, centralized upgrades
- **Cons**: Single point of failure, noisy neighbor risk (one misconfigured CRD can overload the operator's reconciliation queue)

**Namespace-Scoped (Operator-per-Team)**

- **Pros**: Complete isolation, limited blast radius, tightly scoped RBAC
- **Cons**: High resource overhead (potentially dozens of operator pods), management fragmentation

**Hybrid (Operator-per-Environment)**

- **Example**: One operator for production namespaces, another for dev/staging
- **Pros**: Balances isolation with operational efficiency
- **Verdict**: Often the best compromise for large organizations

### Monitoring the Operator (A Notable Gap)

The operator excels at exporting metrics **from PostgreSQL clusters** via Percona Monitoring and Management (PMM). However, the operator _itself_ lacks native Prometheus metrics endpoints.

**Best practice**: Monitor the operator via:

1. **Kubernetes metrics**: Track pod restarts, CPU/memory usage via `kube-state-metrics`
2. **Log aggregation**: Scrape operator logs (Fluentd, Promtail) and alert on errors
3. **CRD status**: Alert on `PerconaPGCluster` resources stuck in `Pending` or failed states

### Upgrade Strategy

**Critical rule**: **Always upgrade CRDs before upgrading the operator Deployment.**

The CRDs are backward-compatible across the last three minor versions, so updating them first is safe. Upgrading the operator without updating CRDs can cause validation failures or reconciliation errors.

**Process (Helm)**:

```bash
# 1. Update Helm repo
helm repo update

# 2. Fetch and apply new CRDs manually (safeguard)
kubectl apply -f https://raw.githubusercontent.com/percona/kubernetes-percona-postgres-operator/v2.7.1/deploy/crd.yaml

# 3. Upgrade the operator
helm upgrade my-operator percona/pg-operator \
  --version 2.7.1 \
  --namespace postgres-operator
```

**Does this disrupt database workloads?** No. Upgrading the operator performs a rolling update of the operator Deployment. PostgreSQL instances continue running normally. During the brief changeover, the control plane (automated failover, reconciliation) is temporarily unavailable, but running databases are unaffected.

### Common Anti-Patterns

**1. Installing pg-db before pg-operator**

The `percona/pg-db` Helm chart creates a `PerconaPGCluster` custom resource. If you install this before the operator, the CRD doesn't exist, or the resource sits there unreconciled. Always install `pg-operator` first.

**2. Missing CSI driver (especially on EKS)**

PVCs stuck in `Pending` with no error? Check your CSI driver. On EKS, install the Amazon EBS CSI driver as a prerequisite.

**3. Forgetting to set resource limits**

The operator defaults to `{}` for resources, resulting in BestEffort QoS. This is unstable for production control planes. Always set requests and limits.

**4. Mismatching operator and CRD versions**

Upgrading the operator image without upgrading CRDs causes compatibility issues. Always update CRDs first.

---

## The Competitive Landscape

Understanding Percona's position requires context. Here's how it compares to the major alternatives:

### Percona vs. Crunchy Data PGO

**Similarity**: Percona is a fork of Crunchy PGO, so the architecture (Patroni for HA, pgBackRest for backups) is nearly identical.

**Difference**: Licensing. Crunchy's source code is Apache 2.0, but the official container images require a commercial subscription for production use. Percona removes this restriction—Docker Hub images are unrestricted.

**Verdict**: Percona is "liberated PGO" for teams that want the mature PGO architecture without commercial entanglements.

### Percona vs. CloudNativePG

**Similarity**: Both are fully open source (Apache 2.0 code and images).

**Difference**: Architecture. CloudNativePG is a "pure" Kubernetes-native design that doesn't use Patroni. It implements HA and failover by interacting directly with the Kubernetes API. Percona uses the battle-tested Patroni.

**Verdict**: CloudNativePG is ideal for teams prioritizing Kubernetes-native tooling. Percona is ideal for teams who trust Patroni's maturity and want integrated monitoring (PMM).

### Percona vs. Zalando Postgres Operator

**Similarity**: Both are fully open source and both use Patroni for HA.

**Difference**: Vendor support and monitoring. Zalando is community-driven (though production-proven at Zalando.com). Percona is backed by a commercial database vendor and differentiates with out-of-the-box PMM integration.

**Verdict**: Zalando is battle-tested at scale with integrated PgBouncer. Percona offers enterprise support options and turnkey monitoring.

### Percona vs. StackGres

**Difference**: Licensing. StackGres is AGPL-3.0, a viral copyleft license. This is a non-starter for most corporate or mixed-source projects.

**Verdict**: StackGres is powerful but legally incompatible with many commercial use cases.

### Percona vs. KubeDB

**Difference**: Commercial model. KubeDB is "open-core." Essential features (backups, TLS, pooling) are gated behind the paid Enterprise Edition. It requires a `global.license` key to function.

**Verdict**: KubeDB is a commercial product, not a true open-source alternative.

---

## Why Planton Supports Percona

After evaluating the landscape, we integrated the **Percona Operator for PostgreSQL** into Planton's IaC framework for the following reasons:

### 1. Unambiguous Open Source

Percona's entire stack—operator code, container images, and the PostgreSQL distribution it deploys—is fully open source with no production restrictions. This aligns with Planton's open-source philosophy and avoids the legal landmines present in competitors.

### 2. Production-Proven Architecture

By forking Crunchy PGO, Percona inherits a mature, battle-tested architecture. The operator is stable at version 2.7.0, tested on all major cloud platforms (GKE, EKS, AKS), and supports PostgreSQL 12–16.

### 3. Integrated Monitoring

The seamless PMM (Percona Monitoring and Management) integration is a significant differentiator. Most operators require separate, manual monitoring setup. Percona provides a production-grade observability stack out of the box.

### 4. Ecosystem Compatibility

As a fork of PGO, Percona benefits from the upstream's continuous innovation while providing a truly open distribution. Teams familiar with PGO can migrate to Percona without architectural disruption.

---

## The Planton API

Our `KubernetesPerconaPostgresOperator` API (see `spec.proto`) abstracts the Helm chart complexity into a minimal, production-ready Protobuf specification. The API exposes only the essential 20% of configuration that 80% of users need:

- **Namespace**: Where to deploy the operator
- **Tenancy mode**: Namespace-scoped, cluster-wide, or multi-namespace
- **Image specification**: Registry, repository, tag (for version pinning and private registries)
- **Resource requirements**: CPU/memory requests and limits
- **Scheduling**: Node selectors and tolerations for dedicated node pools
- **Telemetry**: Opt-out flag for enterprise compliance

The underlying Pulumi and Terraform modules translate this minimal API into a full Helm release, applying sensible defaults and production best practices automatically.

---

## Conclusion: The Control Plane You Can Trust

Deploying PostgreSQL on Kubernetes is now a mature, well-supported pattern. But the quality of your database operations depends entirely on the quality of your operator—the control plane that manages those databases.

The Percona Operator for PostgreSQL stands out in a crowded field because of one simple truth: **licensing clarity matters**. Open-source code with proprietary images is a trap. AGPL licenses are viral and risky. Commercial "open-core" products hide costs behind feature gates.

Percona delivers on the full promise of open source: Apache 2.0 code, unrestricted container images, and production-grade capabilities without license keys or subscription walls. It's the operator you can deploy with confidence, knowing you're building on a solid, legally unencumbered foundation.

Our IaC modules abstract the deployment complexity, giving you a production-ready PostgreSQL control plane through a simple, declarative API. Install the operator once, and let it handle the lifecycle management of your PostgreSQL clusters. That's infrastructure automation done right.

Welcome to worry-free PostgreSQL operations. 🐘🔧☸️

---

## Further Reading

- [Percona Operator Official Documentation](https://docs.percona.com/percona-operator-for-postgresql/2.0/) - Upstream operator docs
- [Percona Helm Charts Repository](https://github.com/percona/percona-helm-charts) - Official Helm charts and values
- [Comparing PostgreSQL Operators](https://www.percona.com/blog/run-postgresql-in-kubernetes-solutions-pros-and-cons/) - Percona's perspective on the competitive landscape
- [PostgreSQL on Kubernetes: The Modern Landscape](../../../workload/postgreskubernetes/v1/docs/README.md) - Companion guide on deploying PostgreSQL databases


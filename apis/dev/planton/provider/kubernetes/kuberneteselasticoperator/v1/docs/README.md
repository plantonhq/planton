# Deploying Elastic Cloud on Kubernetes (ECK): From Quick Starts to Production

## Introduction: The Operator Pattern Changes Everything

For years, running Elasticsearch on Kubernetes was considered an anti-pattern. The conventional wisdom went like this: stateful workloads like Elasticsearch need stable storage, predictable networking, and careful orchestration—all things Kubernetes was supposedly bad at. Teams would run Elasticsearch on VMs, carefully tuning each node, managing certificates manually, and scripting complex upgrade procedures.

Then Kubernetes matured. StatefulSets brought stable identity and ordered deployment. The Operator pattern emerged, encoding operational knowledge directly into Kubernetes controllers. And Elastic built ECK—the **Elastic Cloud on Kubernetes operator**—which automates the full lifecycle of the Elastic Stack on Kubernetes.

Today, ECK represents a paradigm shift: running production Elasticsearch on Kubernetes is not just viable, it's often **easier and more reliable** than traditional approaches. The operator handles certificate rotation, rolling upgrades, scale-out operations, and cross-cluster replication—all the tasks that once required runbooks and midnight maintenance windows.

This document explores the landscape of ECK deployment methods, from quick-start patterns to production-ready approaches, and explains why Planton defaults to Helm-based installation with a clear path to GitOps maturity.

## Understanding ECK: The Operator

Before diving into deployment methods, it's important to understand what ECK actually is. ECK is a Kubernetes Operator that extends the Kubernetes API with Custom Resource Definitions (CRDs) for Elastic Stack components: Elasticsearch clusters, Kibana, APM Server, Enterprise Search, Beats, Elastic Agent, and Logstash.

When you create an Elasticsearch custom resource in Kubernetes, the ECK operator (running as a controller pod) reconciles your desired state with reality. It provisions StatefulSets, manages persistent volumes, generates TLS certificates, configures secure inter-node communication, and orchestrates rolling upgrades—all automatically.

**ECK sits in the middle of a spectrum:**

- **Elastic Cloud (SaaS)**: Fully managed by Elastic. Zero operational burden, but limited infrastructure control and customization.
- **ECK on Kubernetes**: Automated lifecycle management via the operator. You manage the Kubernetes cluster; ECK manages the Elastic Stack.
- **Self-Managed on VMs**: Manual deployment and operation. Maximum control, maximum operational complexity.

ECK brings the automation of managed services to your own infrastructure. For teams already invested in Kubernetes, it's the sweet spot: you retain control over infrastructure, data residency, and configuration while delegating operational toil to the operator.

### Licensing: Open and Flexible

ECK is distributed under the Elastic License v2 with two tiers:

- **Basic (free)**: Includes all core features—security (TLS, authentication, RBAC), monitoring, alerting, Canvas, Maps, and Kibana Spaces. This is production-ready and suitable for most use cases.
- **Enterprise (paid)**: Unlocks advanced features like machine learning, cross-cluster replication, searchable snapshots (frozen tier), and official Elastic support with SLAs.

You start with Basic by default. If you need enterprise features, you can activate a 30-day trial or apply a commercial license. The operator automatically propagates licensing to all managed Elastic Stack components.

## The Deployment Landscape: A Maturity Model

ECK can be deployed to Kubernetes clusters in multiple ways. These aren't just different tools achieving the same end—they represent different levels of operational maturity and investment in automation.

### Level 0: The Quick Start (kubectl + YAML)

**The approach:** Download Elastic's all-in-one YAML manifests (`crds.yaml` and `operator.yaml`) and apply them with `kubectl`.

```bash
kubectl create -f https://download.elastic.co/downloads/eck/3.2.0/crds.yaml
kubectl apply -f https://download.elastic.co/downloads/eck/3.2.0/operator.yaml
```

**What it solves:** This installs ECK in under a minute. It's perfect for local development, proof-of-concepts, or air-gapped environments where you want to understand exactly what's being applied.

**What it doesn't solve:**
- **No parameterization**: You can't easily configure the operator (e.g., resource limits, namespace scoping) without editing the YAML.
- **Manual upgrades**: To upgrade ECK, you download new manifests and reapply. There's no built-in version tracking or rollback.
- **No repeatability**: Applying the same manifests across multiple clusters means manually tracking versions and changes.

**Verdict:** Great for experimentation and learning. For production, you'll quickly want something more structured.

### Level 1: The Package Manager (Helm)

**The approach:** Use Elastic's official Helm chart to install ECK with parameterized configuration.

```bash
helm repo add elastic https://helm.elastic.co
helm install kubernetes-elastic-operator elastic/eck-operator \
  --namespace elastic-system --create-namespace \
  --set replicas=2 \
  --set resources.requests.cpu=500m \
  --set resources.requests.memory=256Mi
```

**What it solves:**
- **Configuration as code**: Override defaults via values files or `--set` flags. Set resource limits, enable metrics, configure namespace scoping.
- **Version management**: `helm upgrade` handles operator updates. `helm rollback` provides a safety net.
- **Repeatable deployments**: Package the values file in version control and apply consistently across environments.
- **Cluster-wide or namespace-scoped**: The chart supports both modes out-of-the-box via profiles.

**What it doesn't solve:**
- **Not continuous**: Helm applies configuration once. If someone manually changes the operator deployment, Helm won't revert it until the next `helm upgrade`.
- **Imperative workflows**: You run `helm install/upgrade` explicitly. It doesn't fit naturally into continuous delivery pipelines without additional orchestration.

**Verdict:** This is the **standard for most production deployments**. Helm provides the right balance of simplicity, repeatability, and flexibility. It's especially powerful when combined with CI/CD pipelines or when deploying to multiple clusters with environment-specific values files.

### Level 2: The Enterprise Catalog (Operator Lifecycle Manager)

**The approach:** Install ECK via OperatorHub on OpenShift or OLM-enabled clusters by creating a Subscription resource.

**What it solves:**
- **Centralized operator governance**: Platform teams can approve, version, and distribute operators to tenants via a catalog.
- **Automated upgrades**: Subscribe to a channel (e.g., "stable") and let OLM handle operator updates.
- **Multi-tenancy**: Control which namespaces can use which operators via OperatorGroups.

**What it doesn't solve:**
- **Complexity overhead**: Requires OLM installed and configured. Configuration is passed via ConfigMaps referenced by the Subscription.
- **Limited ecosystem**: OLM is most mature on OpenShift. On generic Kubernetes, it's less common.

**Verdict:** Ideal for **OpenShift environments** or large enterprises that already use OLM for operator lifecycle management. If you're not already in that ecosystem, the complexity overhead isn't justified for most teams.

### Level 3: The GitOps Paradigm (Argo CD / Flux)

**The approach:** Store ECK installation manifests (or Helm Chart references) in Git. Let Argo CD or FluxCD continuously reconcile your cluster to match the Git state.

```yaml
# argocd-application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kubernetes-elastic-operator
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://helm.elastic.co
    chart: eck-operator
    targetRevision: 3.2.0
    helm:
      releaseName: kubernetes-elastic-operator
      values: |
        replicas: 2
        resources:
          requests:
            cpu: 500m
            memory: 256Mi
  destination:
    server: https://kubernetes.default.svc
    namespace: elastic-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

**What it solves:**
- **Continuous reconciliation**: Argo/Flux continuously monitors Git and the cluster. Manual changes are reverted automatically (self-healing).
- **Audit trail**: Every change goes through Git commits. You can see who changed what, when, and why via commit messages.
- **Multi-cluster at scale**: Deploy ECK to dozens of clusters from a single Git repository using ApplicationSets or Kustomize overlays.
- **Declarative rollback**: Revert a bad change by reverting the Git commit. Argo/Flux applies it automatically.

**What it doesn't solve:**
- **Initial investment**: You need Argo CD or Flux running first (the bootstrapping problem).
- **Secret management**: Sensitive values (like enterprise license keys) require integration with tools like Sealed Secrets, SOPS, or External Secrets Operator.
- **Not for ephemeral clusters**: If you spin up clusters for short-lived tests, the GitOps overhead may not be worth it.

**Verdict:** GitOps represents **operational maturity**. For long-lived, multi-team clusters where consistency, auditability, and drift prevention are critical, this is the gold standard. It's especially powerful when combined with the App-of-Apps pattern: a base application installs ECK, and tenant applications deploy their Elasticsearch clusters.

### Level 4: The Infrastructure-as-Code Integration (Terraform / Pulumi)

**The approach:** Use Terraform or Pulumi to provision both the Kubernetes cluster *and* install ECK as part of cluster initialization.

**Terraform example:**

```hcl
resource "helm_release" "eck_operator" {
  name             = "kubernetes-elastic-operator"
  repository       = "https://helm.elastic.co"
  chart            = "eck-operator"
  namespace        = "elastic-system"
  create_namespace = true
  version          = "3.2.0"

  set {
    name  = "replicas"
    value = "2"
  }
}
```

**What it solves:**
- **End-to-end infrastructure codification**: One Terraform plan provisions your cloud infrastructure, Kubernetes cluster, and all cluster addons (including ECK).
- **Dependency management**: Terraform ensures ECK installs only after the cluster is ready.
- **State tracking**: Terraform state tracks what's deployed. Drift detection is built-in.

**What it doesn't solve:**
- **Not continuous**: Like Helm, Terraform applies changes on `terraform apply`. It doesn't self-heal if someone manually modifies the operator.
- **CRD ordering challenges**: Applying CRDs before the operator requires careful use of `depends_on` (though Helm charts typically handle this).

**Verdict:** Ideal for **platform teams provisioning ephemeral or multi-environment clusters**. If you already use Terraform to manage cloud infrastructure, extending it to install ECK ensures consistency. For long-lived clusters, consider combining this with GitOps: Terraform provisions the cluster and installs Argo CD; Argo CD manages ECK and applications.

## Comparison: Which Method When?

| Method | Best For | Pros | Cons |
|--------|----------|------|------|
| **kubectl + YAML** | Dev, PoCs, learning | Instant start, zero dependencies | No configuration, manual upgrades |
| **Helm** | Most production use cases | Repeatable, configurable, simple | Not continuous, manual execution |
| **OLM** | OpenShift, enterprise catalogs | Centralized governance, auto-updates | Complexity, ecosystem limited |
| **GitOps (Argo/Flux)** | Multi-team, long-lived clusters | Self-healing, audit trail, multi-cluster | Bootstrapping complexity, secret management |
| **Terraform/Pulumi** | Platform provisioning, IaC-first teams | End-to-end codification, state tracking | Not continuous, CRD ordering |

**Key insight:** These methods aren't mutually exclusive. Many organizations use **Terraform to provision clusters**, **Helm (via GitOps or Terraform) to install ECK**, and **GitOps to manage Elastic Stack resources** (Elasticsearch clusters, Kibana, etc.). The right combination depends on your operational maturity and team structure.

## Planton's Approach: Helm with Clear Abstractions

Planton defaults to **Helm-based ECK installation** for the operator itself, with first-class support for GitOps workflows.

### Why Helm?

1. **Simplicity meets power**: Helm is widely adopted, well-understood, and officially supported by Elastic. It provides parameterization without requiring additional controllers.

2. **The 80/20 principle**: Most teams need to configure a handful of settings—namespace, resource limits, replica count. Helm's values files handle this cleanly without overwhelming users with every possible flag.

3. **GitOps compatibility**: Helm Charts integrate seamlessly with Argo CD and Flux. You get the benefits of package management *and* continuous reconciliation.

4. **No vendor lock-in**: Unlike OLM (which is OpenShift-centric), Helm works on any Kubernetes distribution. Unlike Terraform (which requires state management), Helm is declarative and stateless.

### What Planton Abstracts

The `KubernetesElasticOperator` API in Planton focuses on the essential configuration needed to deploy the ECK operator:

- **Operator namespace and scope**: Whether the operator watches all namespaces (cluster-wide) or specific namespaces (multi-tenant).
- **High availability**: Replica count and resource requests/limits for the operator pod.
- **Custom settings**: Telemetry opt-out, metrics exposure, log verbosity.

**What's intentionally omitted:** Low-level Kubernetes details like pod security contexts, node selectors, or webhook certificates. These follow best practices by default and can be overridden via Helm values if needed, but they're not exposed in the primary API to keep it focused.

### The Path to GitOps

While Planton installs ECK via Helm by default, it's designed to fit into GitOps workflows:

1. **Initial deployment**: Use Planton (via Pulumi or Terraform) to provision the cluster and install ECK.
2. **Transition to GitOps**: Export the ECK Helm Chart reference to a Git repository managed by Argo CD or Flux.
3. **Continuous management**: Let GitOps handle ongoing operator updates, configuration drift prevention, and multi-cluster synchronization.

This "crawl, walk, run" progression lets teams start simple and graduate to GitOps maturity without replatforming.

## Production Considerations

Once ECK is deployed, the next challenge is running production Elasticsearch clusters. A few key points:

### High Availability
- **Operator HA**: Run 2 replicas with leader election enabled (ECK's default). Place them on separate nodes.
- **Elasticsearch HA**: Deploy 3 master-eligible nodes across availability zones. Use pod anti-affinity to avoid co-location.

### Resource Management
- **Right-size the operator**: ECK is lightweight. 0.5 CPU and 256Mi memory is typically sufficient.
- **Elasticsearch pods**: Allocate adequate memory (Elasticsearch is memory-hungry). Set JVM heap to ~50% of pod memory.

### Security
- ECK is **secure by default**: TLS is enabled automatically for inter-node communication. The `elastic` user password is generated and stored in a Kubernetes Secret.
- Integrate **cert-manager** for custom certificates if needed.
- Enable network policies to restrict traffic to Elasticsearch pods.

### Monitoring
- Enable operator metrics (set `metrics-port` in ECK config) for Prometheus scraping.
- Use **Stack Monitoring**: Deploy a dedicated monitoring Elasticsearch cluster to collect metrics from production clusters.

### Backup and Disaster Recovery
- Configure **snapshot repositories** (e.g., S3, GCS) via Elasticsearch settings.
- Use **Snapshot Lifecycle Management (SLM)** to automate periodic backups.
- Test restore procedures regularly.

### Licensing
- The **Basic license** (free) includes security, monitoring, and alerting—sufficient for most production workloads.
- Activate **Enterprise** if you need machine learning, cross-cluster replication, or searchable snapshots.
- Monitor license expiration via the `elastic-licensing` ConfigMap.

## Beyond Installation: What ECK Manages

It's worth emphasizing: installing the **ECK operator** is distinct from deploying **Elasticsearch clusters**. The operator is the lifecycle manager; the clusters are what it manages.

Once ECK is installed, you create Elastic Stack resources via CRDs:

```yaml
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: production
  namespace: elastic
spec:
  version: 8.10.0
  nodeSets:
  - name: masters
    count: 3
    config:
      node.roles: ["master"]
  - name: data
    count: 5
    config:
      node.roles: ["data", "ingest"]
```

ECK handles:
- Provisioning StatefulSets and persistent volumes
- Generating and rotating TLS certificates
- Orchestrating rolling upgrades (one node at a time, waiting for green status)
- Scaling clusters up or down safely
- Integrating Kibana, APM Server, Beats, and Elastic Agent

This is the power of the operator pattern: you declare intent, and ECK handles the operational complexity.

## Conclusion: From Quick Start to Operational Maturity

The evolution from `kubectl apply` to GitOps mirrors the broader maturity of Kubernetes itself. What started as a platform for stateless applications now reliably runs complex stateful systems like Elasticsearch—*because operators like ECK encode decades of operational knowledge into software*.

**Planton's choice—Helm-based installation with GitOps compatibility—reflects a pragmatic philosophy:**

- Start simple: Helm is approachable and widely understood.
- Scale gracefully: GitOps fits naturally as operational needs grow.
- Stay open: No lock-in to proprietary platforms or tooling.

The goal isn't to dictate a single "right" way to deploy ECK. It's to provide sensible defaults for the common case while offering flexibility for advanced scenarios. Whether you're running a single Elasticsearch cluster for application logs or managing dozens of clusters across multiple regions, ECK on Kubernetes—deployed thoughtfully—provides the automation and reliability that once seemed impossible.

**Next steps:**
- Explore the [ECK Operator Configuration Guide](./operator-configuration.md) for detailed tuning options
- Read [Running Production Elasticsearch on ECK](./production-patterns.md) for cluster topology best practices
- See [Multi-Tenant ECK Deployments](./multi-tenancy.md) for namespace isolation strategies

---

*This document is based on ECK 3.2.0 and Kubernetes 1.29+. For the latest official guidance, always consult the [Elastic Cloud on Kubernetes documentation](https://www.elastic.co/docs/deploy-manage/deploy/cloud-on-k8s).*


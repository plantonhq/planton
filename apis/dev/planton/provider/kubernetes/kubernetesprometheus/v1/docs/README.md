# Deploying Prometheus on Kubernetes: From Anti-Patterns to Production

## Introduction

"Just deploy Prometheus as a simple Deployment, it's just another container, right?" This misconception has derailed countless monitoring initiatives. Prometheus is fundamentally a **stateful time-series database**, not a stateless application. The journey from a basic setup that works in development to a production-grade monitoring stack that scales organizationally and technically requires understanding a critical evolution in deployment methods.

Prometheus on Kubernetes has matured from imperative, manual configurations to declarative, operator-driven ecosystems. This document explores that evolution, explains what deployment methods exist across the industry, and reveals why Planton chose the **kube-prometheus-stack** as its foundation for production monitoring.

## The Evolution: From Anti-Patterns to Production Solutions

### Level 0: The Deployment Anti-Pattern

**What it is**: Using a standard Kubernetes `Deployment` (kind: Deployment) to run Prometheus.

**Why it fails**: The Deployment controller is designed for stateless applications—pods that can be terminated and replaced without data loss. Prometheus's value lies entirely in its Time-Series Database (TSDB), which requires persistent, stable storage.

**The failure modes**:
- **With replicas > 1 and a single PVC**: The deployment will fail to start. Cloud provider storage classes (AWS EBS, GCE Persistent Disk) provide ReadWriteOnce (RWO) volumes that cannot be mounted by multiple pods simultaneously.
- **Without a PVC**: The deployment uses an `emptyDir` volume. All historical metric data is permanently lost upon every pod restart, crash, or node failure.

**Verdict**: Suitable only for brief local testing or "hello world" tutorials. This is a critical anti-pattern for any environment requiring data persistence.

### Level 1: The StatefulSet Trap

**What it is**: Deploying Prometheus using a StatefulSet with persistent volumes.

**Why it's better**: StatefulSets correctly provide stable, unique pod identities (e.g., `prometheus-0`) and corresponding unique, stable persistent volumes. The storage problem is solved.

**Why it's still insufficient**: While StatefulSets solve stable storage, they **completely fail to address configuration lifecycle management** in a dynamic Kubernetes environment.

**The trap**:
- **Configuration toil**: The `prometheus.yml` configuration is managed as a central ConfigMap. To add a new application scrape target, an operator must manually edit this ConfigMap and then trigger a rolling restart of the StatefulSet pods.
- **The "HA" fallacy**: Setting `replicas: 2` does **not** provide High Availability. It creates two identical Prometheus pods, each scraping the exact same targets and storing duplicate data. When you query for metrics, you receive two identical time series, resulting in confusing and incorrect graphs. This necessitates a complex query-layer solution like Thanos to perform deduplication.
- **Organizational bottleneck**: Manual configuration is untenable at scale. It creates a centralized bottleneck where all development teams must submit pull requests to modify a single infrastructure ConfigMap.

**Verdict**: The core issue is not the StatefulSet itself, but the attempt to imperatively manage static configuration in an environment defined by declarative, dynamic resources.

### Level 2: The Helm Chart (Without Operator)

**What it is**: Using the standalone `prometheus-community/prometheus` Helm chart, which deploys Prometheus as a StatefulSet without the Prometheus Operator.

**Capabilities**: This chart deploys Prometheus, Alertmanager, node-exporter, and kube-state-metrics. It includes default scrape configs for Kubernetes components.

**The limitation**: To add a new application, users must centrally edit the Helm release's `values.yaml` file, manually adding scrape jobs to the `extraScrapeConfigs` block using raw `prometheus.yml` syntax. A `helm upgrade` must then be performed, triggering pod restarts.

**Verdict**: This reverts to the "manual ConfigMap" anti-pattern. It creates an organizational bottleneck unsuitable for dynamic, multi-team environments. This method manages Prometheus _on_ Kubernetes, but does not make it _part of_ Kubernetes.

### Level 3: The Production Solution — The Prometheus Operator

**What it is**: The **Prometheus Operator** extends the Kubernetes API with Custom Resource Definitions (CRDs) that make Prometheus a first-class, Kubernetes-native service.

**The paradigm shift**: Instead of imperatively editing configuration files, you declaratively define monitoring intent using Kubernetes objects.

**Key CRDs**:
- **Prometheus**: Defines a desired Prometheus deployment (version, replicas, retention, storage)
- **Alertmanager**: Defines a desired Alertmanager cluster
- **ServiceMonitor**: Declaratively specifies how to monitor a set of Kubernetes Services
- **PodMonitor**: Declaratively specifies how to monitor Pods directly
- **PrometheusRule**: Declaratively defines alerting and recording rules
- **AlertmanagerConfig**: Declaratively configures Alertmanager routing and receivers

**The workflow transformation**:

**Old Way (Imperative)**:
1. Manually create a StatefulSet and ConfigMap
2. To add a target, manually edit the ConfigMap
3. Manually restart the StatefulSet

**New Way (Declarative)**:
1. Create a `kind: Prometheus` object in YAML defining high-level needs (e.g., `replicas: 2`, `retention: 30d`)
2. The Prometheus Operator controller watches for this object and reconciles the desired state by creating and managing the underlying StatefulSet, Secret (containing the generated config), Service, and RBAC

**The service discovery breakthrough**: To monitor a new application:
1. The application team deploys their app with a Service
2. In their own namespace, they deploy a `kind: ServiceMonitor` object
3. This ServiceMonitor uses a label selector to find their Service and specifies which port and path to scrape
4. The Prometheus Operator detects the new ServiceMonitor, automatically regenerates the `prometheus.yml` configuration, and gracefully reloads Prometheus

This pattern enables **declarative, GitOps-driven, and decentralized management** of monitoring targets. Application teams can manage their own monitoring configuration via standard Kubernetes RBAC, eliminating the platform team as a bottleneck.

### Level 4: The Complete Ecosystem — kube-prometheus-stack

**What it is**: The `kube-prometheus-stack` Helm chart packages the Prometheus Operator with a comprehensive, battle-tested, end-to-end monitoring ecosystem.

**Nomenclature clarity** (a common source of confusion):
- **prometheus-operator**: The controller itself (the Go binary that implements reconciliation logic)
- **kube-prometheus**: The upstream project that bundles the operator with Grafana dashboards, ServiceMonitors for core Kubernetes components, and a full library of default PrometheusRules (alerts)
- **kube-prometheus-stack**: The **Helm chart** (in the prometheus-community repository) that packages the entire kube-prometheus collection for easy deployment

**What it includes out-of-the-box**:
- The Prometheus Operator
- A Prometheus instance (managed by the Operator)
- An Alertmanager cluster (managed by the Operator)
- Grafana (pre-configured with Prometheus data source and dozens of default dashboards)
- Essential exporters: `kube-state-metrics` and `prometheus-node-exporter`
- Default ServiceMonitors to automatically scrape all core Kubernetes components (kubelet, API server, etc.)
- A comprehensive library of production-ready alerting rules

**Verdict**: This is the **de-facto industry standard** for production Prometheus on Kubernetes. It provides a complete, end-to-end, and battle-tested solution that is declarative, Kubernetes-native, and scales organizationally.

## Comparative Analysis: Operator vs. Standalone

The choice is not between two equivalent options, but between a modern, Kubernetes-native approach and a legacy, static one.

| Feature | kube-prometheus-stack (Operator) | prometheus-community/prometheus (Standalone) |
|:---|:---|:---|
| **Management Model** | **Declarative (Operator Pattern)** | **Imperative (Helm + values.yaml)** |
| **Service Discovery** | **CRD-based (Kubernetes-native)**<br/>ServiceMonitor, PodMonitor | **Static Config (YAML-native)**<br/>extraScrapeConfigs |
| **Configuration Workflow** | **Decentralized, GitOps-friendly, RBAC-able**<br/>App teams create their own ServiceMonitor CRDs | **Centralized, Monolithic, Bottleneck**<br/>All teams must edit one central values.yaml file |
| **Rule Management** | Declarative PrometheusRule CRD | Via serverFiles.rules in values.yaml (static files) |
| **Production Adoption** | **De-facto industry standard** | Niche / Simple / Legacy use cases |
| **Included Components** | **Full End-to-End Stack**<br/>Operator, Prometheus, Alertmanager, Grafana, Exporters, Default Rules/Dashboards | **Core Components**<br/>Prometheus, Alertmanager, Exporters |
| **Licensing** | **100% Open Source (Apache-2.0)** | **100% Open Source (Apache-2.0)** |

The standalone chart manages Prometheus _on_ Kubernetes; the Operator chart makes Prometheus _part of_ Kubernetes.

## The Hybrid Production Pattern: Self-Hosted + Managed Storage

While self-hosting provides complete control and data sovereignty, long-term storage at scale is challenging. The most common production pattern is a **hybrid approach**:

**Self-hosted monitoring with managed long-term storage**:
1. Deploy a lightweight kube-prometheus-stack with short retention (e.g., 24-48 hours)
2. Configure the `remoteWrite` setting to stream all metrics to a managed service:
   - **AWS Managed Service for Prometheus (AMP)**
   - **Google Cloud Managed Service for Prometheus (GMP)**
   - **Grafana Cloud (Mimir)**
3. This provides:
   - Fast, local querying and alerting for recent data
   - Fully managed, globally-scalable long-term storage (months to years)
   - Predictable operational burden

**Trade-offs**:
- **Self-hosted only**: Complete control, potentially lower cost at steady high volume, but high operational burden (scaling, HA, storage management, expertise in Thanos/federation)
- **Fully managed**: Zero infrastructure management, automatic scaling and HA, but unpredictable and potentially very expensive costs (based on samples ingested, which can be problematic with high cardinality)
- **Hybrid**: Best of both worlds for most production environments

Any production-grade deployment tool must support `remoteWrite` configuration as a first-class feature.

## Planton's Choice: kube-prometheus-stack

Planton deploys the **kube-prometheus-stack Helm chart** as the underlying implementation for its `PrometheusKubernetes` resource.

**Why kube-prometheus-stack**:
1. **Production-proven**: This is the undisputed industry standard for production Prometheus on Kubernetes
2. **Declarative and Kubernetes-native**: Enables GitOps workflows and decentralized monitoring management via CRDs
3. **Complete ecosystem**: Provides a full end-to-end monitoring stack (Prometheus, Alertmanager, Grafana, exporters, default dashboards, and alerts) out-of-the-box
4. **100% open source**: The entire stack and all components are licensed under the permissive Apache-2.0 license
5. **Organizational scalability**: The ServiceMonitor/PodMonitor pattern is the only method that scales in multi-team clusters, removing the platform team as a bottleneck

**The abstraction principle**: Planton's protobuf API provides a high-level, opinionated schema that translates into the kube-prometheus-stack chart's `values.yaml`. The API focuses on the **"80% essentials"** that most users need, while providing escape hatches for advanced features.

**Essential configuration** (the "80%"):
- Replicas (1 for staging, 2 for production HA)
- Retention period (e.g., `7d`, `30d`)
- Persistent storage (enabled/disabled, size, storage class)
- Component toggles (Alertmanager, Grafana)
- Ingress configuration (hostnames, annotations for cert-manager, authentication)

**Advanced configuration** (the "20%", exposed in `AdvancedSpec`):
- `remoteWrite` for long-term storage (streaming to AMP, GMP, Grafana Cloud)
- `thanos` sidecar for self-hosted HA and long-term storage
- `globalImageRegistry` override for air-gapped or enterprise security environments

## Production Best Practices

### High Availability: Redundancy vs. True HA

Setting `replicas: 2` provides **redundancy**, not **High Availability**. This creates two identical Prometheus pods, each scraping the same targets and storing duplicate data.

**True HA** is achieved on the query path using a global query layer like **Thanos**:
1. Deploy `replicas: 2` for redundancy
2. Configure `externalLabels` to include a unique `replica` label (e.g., `replica: "$(POD_NAME)"`)
3. Enable the Thanos sidecar in each Prometheus pod to upload TSDB blocks to object storage (e.g., S3) and expose a gRPC Store API
4. Deploy the **Thanos Querier** component in a central cluster
5. Configure Grafana to use Thanos Querier as its data source
6. Thanos Querier automatically discovers all Prometheus sidecars, fetches metrics from both redundant instances, and **deduplicates** them in real-time based on the `replica` label

This provides a single, correct, highly-available global view of all metrics, resilient to the failure of a single Prometheus pod.

### Storage Management

**Sizing formula**:
```
needed_disk_space = retention_time_seconds * ingested_samples_per_second * bytes_per_sample
```
- `bytes_per_sample`: Averages 1-2 bytes
- `ingested_samples_per_second`: Query `rate(prometheus_tsdb_head_samples_appended_total[5m])` on a running instance

**Best practices**:
- **Local retention**: 30-60 days maximum. Anything longer must use `remoteWrite` or Thanos sidecar to offload to object storage
- **Size-based safeguard**: Set `retentionSize` (e.g., `80Gi`) as a circuit breaker at 80-85% of total allocated disk space to prevent disk-full crashes
- **StorageClass selection**: Use high-IOPS storage:
  - **AWS**: Use `gp3` volumes (decouples IOPS from size for better performance at lower cost)
  - **Azure**: Use Premium SSD (`managed-csi-premium`)
  - **On-prem**: Use high-IOPS block storage (e.g., Ceph RBD, dedicated SAN)

### Resource Sizing

Prometheus memory usage is driven by time series cardinality. CPU usage is driven by scrape load, rule evaluation, and query complexity.

**Production sizing guidelines (per-replica)**:

| Cluster Size | Nodes | Pods | Samples/Sec (Est.) | CPU (Req/Limit) | Memory (Req/Limit) | PVC Size (30d) |
|:---|:---|:---|:---|:---|:---|:---|
| **Small (Staging)** | < 10 | < 100 | ~10,000 | 500m / 1 | 2Gi / 4Gi | 20-50Gi |
| **Medium (Prod)** | < 50 | < 500 | ~50,000 | 2 / 4 | 8Gi / 16Gi | 100-250Gi |
| **Large (Prod)** | > 100 | > 1000 | ~150,000+ | 4 / 8 | 16Gi / 32Gi | 250Gi+ (or requires long-term storage) |

**Anti-pattern**: Under-provisioning resources. A production instance scraping even a small cluster can easily consume 6-13GiB of memory. Medium-sized clusters require 8-16GiB of memory requests per replica.

### ServiceMonitor vs. PodMonitor

- **ServiceMonitor** (default, recommended): Dynamically discovers scrape targets from the Endpoints of a Kubernetes Service. Use this whenever your application is exposed via a Service.
- **PodMonitor**: Bypasses the Service abstraction and scrapes Pods directly via pod labels. Use only when a Service does not exist or does not make sense (e.g., DaemonSets like node-exporter, scraping kubelet on the host).

### Common Pitfalls

#### Pitfall 1: The Cardinality Explosion

**The #1 cause of Prometheus failure at scale.**

**Definition**: Cardinality is the total number of unique time series. A time series is defined by its metric name plus its unique set of key-value labels.

**Cause**: Instrumenting applications with labels that have unbounded, high-cardinality values.

**Anti-patterns** (never use these as labels):
- `user_id`, `session_id`, `request_id`, `email_address`, `ip_address`
- Full URL paths or any other unique identifier

**Symptoms**: Rapidly increasing memory usage, slow or failing queries, eventual OOMKilled events as the pod's memory limit is breached.

#### Pitfall 2: OOMKilled (Exit Code 137)

When a Prometheus pod is OOMKilled, the cause is almost always one of three issues:
1. **High cardinality** (most likely culprit—see above)
2. **Insufficient resource limits**: The pod's `resources.limits.memory` is set too low for the workload (production often needs 8GiB+)
3. **Expensive queries**: A single complex PromQL query (e.g., `rate(my_metric[30d])`) can load massive amounts of data into memory, causing a spike that kills the pod

#### Pitfall 3: Scrape Interval Anti-Patterns

**Anti-pattern**: Setting a globally low scrape interval (e.g., `5s` or `10s`). This dramatically increases ingestion load, network traffic, and storage requirements.

**Best practice**: Use a sensible default of `30s` or `60s`. Apply faster intervals (e.g., `15s`) only for critical components by specifying it on their individual ServiceMonitor or PodMonitor.

**Querying rule**: When using the `rate()` function in PromQL, the time range vector (e.g., `[5m]`) should be **at least 4x the scrape interval**. This ensures `rate()` always has at least two data points to calculate from, preventing graphs that intermittently drop to zero.

## Licensing: 100% Open Source

Planton requires a 100% open-source deployment path. The entire recommended stack is permissively licensed:

| Component | License | Source |
|:---|:---|:---|
| Prometheus Server | Apache-2.0 | prometheus/prometheus |
| Prometheus Operator | Apache-2.0 | prometheus-operator/prometheus-operator |
| kube-prometheus | Apache-2.0 | prometheus-operator/kube-prometheus |
| Community Helm Charts | Apache-2.0 | prometheus-community/helm-charts |
| Alertmanager | Apache-2.0 | prometheus/alertmanager |
| kube-state-metrics | Apache-2.0 | kubernetes/kube-state-metrics |
| node-exporter | Apache-2.0 | prometheus/node_exporter |
| prometheus-adapter | Apache-2.0 | kubernetes-sigs/prometheus-adapter |

**Container image sources**:
- **quay.io**: Hosts primary Prometheus, Operator, and Thanos images
- **registry.k8s.io**: Official Kubernetes registry, hosts kube-state-metrics and prometheus-adapter

**Enterprise considerations**: The kube-prometheus-stack chart supports the `global.imageRegistry` value, allowing organizations to mirror official images to their private registry (ECR, Artifactory, etc.) for air-gapped or high-security environments.

## Conclusion: A Paradigm Shift

The evolution from manual StatefulSets to the Prometheus Operator represents a fundamental paradigm shift. Monitoring is no longer an external tool _running on_ Kubernetes—it is a truly Kubernetes-native service, managed declaratively via CRDs and scaled organizationally through decentralized, GitOps-driven workflows.

Planton embraces this paradigm by building on the battle-tested kube-prometheus-stack. The result is a production-grade monitoring solution that is open source, declarative, and ready to scale from development clusters to large, multi-team production environments.

For comprehensive implementation details, operator configuration deep dives, and Thanos integration guides, explore the additional documentation in this directory.


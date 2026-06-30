# Deploying Apache Solr Operator on Kubernetes: From Anti-Patterns to Production-Ready Operators

## Introduction

For years, the conventional wisdom was clear: **don't run stateful search workloads like Solr on Kubernetes**. The reasoning seemed sound—Solr's complex distributed architecture with ZooKeeper coordination, shard placement, and stateful storage requirements appeared fundamentally incompatible with Kubernetes' ephemeral, containerized world. Yet today, organizations like Bloomberg run over a thousand SolrCloud clusters on Kubernetes, spanning hundreds of machines.

What changed? The answer lies not in Solr becoming simpler or Kubernetes gaining magic stateful powers, but in the emergence of **Kubernetes operators** that encode Solr's operational complexity into the control plane. The Apache Solr Operator represents this paradigm shift—transforming what was once a deployment nightmare into a manageable, production-ready solution.

This document explores the deployment landscape for Solr on Kubernetes, from anti-patterns to avoid through intermediate approaches, culminating in operator-based solutions. We'll examine why Planton chose the Apache Solr Operator as the default deployment method, and where to learn more about production operations.

## The Solr on Kubernetes Challenge

Before diving into deployment methods, it's crucial to understand what makes SolrCloud uniquely challenging on Kubernetes:

**ZooKeeper Dependency**: SolrCloud requires Apache ZooKeeper for cluster coordination and state management. Without a reliable ZooKeeper ensemble maintaining quorum, the entire SolrCloud becomes inaccessible for reads and writes. Running ZooKeeper "embedded" within Solr processes—while convenient for development—is strongly discouraged for production. Any node hosting embedded ZooKeeper that goes down takes the cluster coordinator with it.

**Sharding and Replication**: Collections in SolrCloud are split into shards, with multiple replicas per shard for availability. Kubernetes StatefulSets provide stable pod identities and persistent storage, but naive rolling updates can inadvertently take down multiple replicas of the same shard simultaneously, causing data unavailability.

**Data Locality and Persistence**: Solr nodes store index data on disk. In Kubernetes, this means PersistentVolumes that must survive pod rescheduling. The system must ensure correct volume mounting, prevent accidental deletion during scale-down, and coordinate carefully when pods move between nodes.

**Scaling Complexity**: Adding Solr pods doesn't automatically redistribute data. Scaling out requires explicit API calls to add replicas or reshuffle shards. Scaling in demands moving or deleting replicas before removing pods—otherwise you risk losing unique data. This orchestration cannot be handled by Kubernetes primitives alone.

These challenges explain why simple deployment approaches fall short and why an operator-based solution emerged as the industry standard.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern – Manual StatefulSets

**What it is**: Manually creating Kubernetes StatefulSets and Services for Solr and ZooKeeper, managing configuration through raw YAML files.

**The appeal**: Seems straightforward—just containerize Solr and deploy it like any stateful application.

**Why it fails**: This approach is fragile and error-prone. SolrCloud requires tight coordination that manual StatefulSets don't provide: ensuring ZooKeeper availability before Solr starts, handling ordered restarts during updates, managing shard placement, and coordinating recovery. You're left implementing operational logic through custom scripts or hoping for the best.

Using Solr's embedded ZooKeeper (bundled with Solr for convenience) in production is particularly dangerous. If that single Solr node dies, your entire cluster loses its coordinator. The official Solr documentation explicitly warns: **embedded ZooKeeper is meant only for development—always use an external ZooKeeper ensemble in production**.

**Verdict**: Avoid entirely. The operational burden and failure modes make this approach unsuitable for anything beyond proof-of-concept testing.

### Level 1: Helm Charts Without Operators

**What it is**: Using Helm charts (Apache's official charts or third-party offerings like Bitnami) to deploy SolrCloud with templated StatefulSets and Services.

**What it solves**: Helm simplifies initial deployment by templating Kubernetes resources. Charts can bootstrap a SolrCloud cluster including a ZooKeeper ensemble with minimal YAML editing. They handle basic configuration like resource limits, storage classes, and service exposure.

**Limitations**: While Helm charts improve deployment consistency, they don't solve lifecycle management. Scaling still relies on Kubernetes rolling updates that don't understand Solr's shard topology. Upgrades can inadvertently violate replication requirements. Complex operations like backup/restore, shard rebalancing, or controlled rolling restarts require manual intervention or custom scripting.

The Apache Solr project provides official Helm charts, but notably, these charts are designed to work **in conjunction with the Solr Operator's CRDs** rather than as standalone installations. Third-party charts that deploy pure StatefulSets exist but lack the deeper integration needed for production operations.

**Verdict**: Acceptable for staging environments or when paired with manual operational expertise. Insufficient for production without significant custom tooling.

### Level 2: The Production Solution – Kubernetes Operators

**What it is**: Using a Kubernetes operator—specifically the **Apache Solr Operator**—to automate the full lifecycle of SolrCloud clusters through custom resources and controllers.

**What it solves**: Everything. Operators encode domain-specific knowledge into Kubernetes controllers that continuously reconcile desired state with actual state. The Apache Solr Operator manages:

- **Cluster Creation and Updates**: Define a `SolrCloud` custom resource with desired configuration, and the operator creates StatefulSets, Services, ConfigMaps, and coordinates startup ordering.
- **Safe Rolling Updates**: The operator's managed update strategy ensures only a subset of replicas restart at a time, respecting Pod Disruption Budgets and shard availability.
- **ZooKeeper Integration**: Connect to external ZooKeeper or let the operator provision a ZooKeeper ensemble automatically via dependency operators.
- **Scaling with Shard Rebalancing**: Starting with version 0.8.0, the operator can automatically redistribute collection shards when nodes are added or removed, using Solr's Collections API under the hood.
- **Backup and Restore**: The `SolrBackup` CRD enables scheduling and managing Solr backups as Kubernetes objects.
- **Monitoring Integration**: Deploy the Solr Prometheus Exporter via CRD to expose metrics for Prometheus/Grafana monitoring.
- **Advanced Configuration**: TLS certificates, custom JVM options, node affinity, ingress setup—all exposed through the operator's API.

**Production Evidence**: Bloomberg's search team runs over a thousand SolrCloud clusters on Kubernetes using this operator. Originally developed at Bloomberg, it was donated to Apache and is now the **official Kubernetes operator for SolrCloud**. Active development, production-proven features, and real-world scale validation make this the mature choice.

**Verdict**: This is the recommended approach for production Solr deployments on Kubernetes. The official Apache recommendation aligns: use the Solr Operator for production.

## Apache Solr Operator: The Standard

### Why the Apache Solr Operator?

**100% Open Source**: Licensed under Apache 2.0, with both source code and container images freely available. No proprietary components, no license fees, no restricted features. You can use, modify, and distribute it without constraints.

**Production-Ready Maturity**: While still in beta API (v1beta1), the operator has been battle-tested at massive scale. Bloomberg's deployment spans hundreds of machines and thousands of Solr nodes. The project shows healthy activity with regular releases, active GitHub community, and comprehensive documentation.

**Comprehensive Feature Set**: Unlike other stateful operators that handle only basic deployment, the Solr Operator covers the full operational lifecycle:
- Automated ZooKeeper provisioning or external connection
- Persistent and ephemeral storage options
- Horizontal Pod Autoscaler (HPA) integration
- Automatic shard rebalancing on scale events
- Backup/restore as first-class Kubernetes objects
- Prometheus metrics integration
- Ingress and custom networking configuration
- TLS certificate management
- Pod anti-affinity and topology awareness

**No Viable Alternatives**: The Kubernetes operator ecosystem for Solr is narrow. DataStax (known for Cassandra tooling) does not provide a dedicated Solr operator. Their Cass Operator focuses on Cassandra and doesn't support managing SolrCloud. Third-party operators are essentially non-existent—most users either adopt the Apache operator or build custom Helm-based solutions with significant operational burden.

### Licensing Clarity

A critical consideration for production deployments is licensing transparency:

- **Apache Solr**: Apache License 2.0—permissive, free to use in any environment
- **Apache Solr Operator**: Apache License 2.0—same permissive terms
- **Container Images**: Official images include Apache 2.0 license information in `/etc/licenses`
- **ZooKeeper**: Also Apache License 2.0 (used via Pravega ZooKeeper Operator or external deployment)

There are no "open core" dynamics here. All features—backup, metrics, autoscaling, security—are available in the open source version. Commercial support is available from companies like Lucidworks and DataStax (through their Luna program), but purchasing support doesn't change the software license. You pay for expertise and SLAs, not for software rights.

### Key Design Philosophy

The Apache Solr Operator embraces **GitOps-friendly declarative management**. Define your desired SolrCloud state in a custom resource, apply it to the cluster, and the operator continuously reconciles reality with your specification. Changes to Solr version, replica count, storage size, or configuration trigger controlled rolling updates. Delete the custom resource, and the operator handles graceful cluster teardown (with configurable retention policies for data volumes).

This approach aligns with modern Kubernetes operations: treat infrastructure as code, leverage automated reconciliation loops, and reduce manual toil. The operator becomes an extension of Kubernetes' control plane, giving you the same operational patterns for Solr that you use for other workloads.

## Production Best Practices at a Glance

Deploying SolrCloud on Kubernetes successfully requires attention to several key areas:

**ZooKeeper Deployment**: Use a **3-node external ZooKeeper ensemble** with persistent storage. This can be a separate Helm-deployed ZK cluster or managed via a ZooKeeper operator. The Solr Operator can provision ZK automatically, but many teams prefer managing it separately for greater control and reliability. Never use embedded ZooKeeper in production.

**Persistent Storage**: Configure PersistentVolumes with fast SSDs or NVMe drives. Set `reclaimPolicy: Retain` to prevent accidental data loss during scale-down. Size volumes generously to accommodate index growth and merge operations. Avoid NFS for active indexes—high latency degrades performance.

**High Availability**: Run at least **replication factor 2** (preferably 3) for critical collections. Use pod anti-affinity to spread replicas across nodes and zones. Configure Pod Disruption Budgets (automatically created by the operator) to prevent simultaneous pod evictions during maintenance.

**Resource Sizing**: Allocate sufficient memory for JVM heap (typically 6-8GB for production) plus extra for OS page cache. Request 2-4 CPU cores per Solr pod. Monitor heap usage and GC metrics—long GC pauses degrade query performance.

**Monitoring and Alerting**: Deploy the Solr Prometheus Exporter to scrape metrics. Watch query latency (95th/99th percentile), cache hit rates, JVM heap usage, GC times, and replication lag. Aggregate logs centrally and alert on error patterns like OOM warnings or ZooKeeper connection losses.

**Backup Strategy**: Use the `SolrBackup` CRD to schedule regular backups to persistent volumes or cloud storage (S3, GCS). Test restores regularly. For disaster recovery, consider cross-datacenter replication (CDCR) to mirror SolrClouds across regions.

**Common Pitfalls to Avoid**:
- Running without replication (single point of failure)
- Neglecting ZooKeeper monitoring and capacity planning
- Using default JVM heap without tuning for workload
- Disabling pod anti-affinity and losing shard availability
- Scaling down without moving replicas first
- Upgrading Solr, operator, and Kubernetes simultaneously (change one thing at a time)

For a comprehensive guide to production operations, see the Apache Solr Operator documentation and the official Solr Reference Guide's deployment section.

## The 80/20 Configuration Principle

When designing the Planton API for KubernetesSolrOperator, we apply the **80/20 principle**: focus on the 20% of configuration parameters that 80% of users need to make informed decisions.

**Essential Configuration Fields**:
1. **Cluster Identity**: Name and namespace for the SolrCloud instance
2. **Solr Version**: The Solr Docker image tag (e.g., "8.11.2", "9.3.0")
3. **Replica Count**: Number of Solr nodes (pods) in the cluster
4. **Storage**: Persistent vs. ephemeral, storage size per node, storage class
5. **ZooKeeper Connection**: External connection string or automatic provisioning (with ZK replica count)
6. **Resource Limits**: CPU and memory requests/limits per Solr pod
7. **JVM Heap Size**: Solr's Java heap allocation (-Xms/-Xmx)
8. **External Exposure**: Whether to create Ingress for external access and domain configuration

**Advanced Settings (Defaulted or Omitted)**:
- Custom SolrPod environment variables, sidecars, tolerations
- Exotic shard placement policies
- TLS keystore/truststore details (assume ingress-terminated TLS)
- Solr Prometheus Exporter custom metrics configuration
- Backup scheduling (handled separately from deployment)
- Multi-zone topology spread constraints (basic anti-affinity covers most cases)

This minimal API surface makes KubernetesSolrOperator approachable without sacrificing flexibility. Power users needing advanced tuning can access the full Solr Operator CRD directly, but the Planton abstraction covers the typical production scenario.

## Planton's Choice

**Default: Apache Solr Operator** deployed via Helm, configured through the Planton abstraction layer.

**Justification**:
- **Open Source Philosophy**: Aligns with Planton's commitment to 100% open source tooling—no vendor lock-in, no license fees.
- **Production Proven**: Bloomberg's massive scale deployment and active Apache community provide confidence in stability and support.
- **Comprehensive Automation**: The operator handles the full lifecycle—deployment, scaling, upgrades, backup, monitoring—reducing operational toil.
- **No Viable Alternatives**: The Apache Solr Operator is the de facto standard. Other approaches (manual Helm, custom scripts) require significantly more effort for equivalent functionality.
- **Future-Proof**: Active development with a roadmap toward v1.0.0, backward-compatible upgrade paths, and alignment with Kubernetes evolution.

By defaulting to the Apache Solr Operator, Planton provides a production-ready SolrCloud deployment experience while abstracting away unnecessary complexity. The protobuf-based API exposes the essential 80% of configuration, with the operator handling the operational 20% that used to require manual intervention.

## Where to Learn More

**Apache Solr Operator Documentation**: [https://apache.github.io/solr-operator/](https://apache.github.io/solr-operator/)
Comprehensive guides covering installation, CRD reference, upgrade procedures, and advanced configuration.

**Apache Solr Reference Guide**: [https://solr.apache.org/guide/](https://solr.apache.org/guide/)
Official Solr documentation including ZooKeeper ensemble setup, SolrCloud concepts, and performance tuning.

**Sematext Solr Operator Autoscaling Tutorial**: [https://sematext.com/solr-operator-autoscaling-tutorial/](https://sematext.com/solr-operator-autoscaling-tutorial/)
Practical walkthrough of deploying Solr with the operator, including HPA integration and shard rebalancing.

**Bloomberg Tech Talk - Running Solr at Scale**: [https://www.youtube.com/watch?v=MD6NXTrA3xo](https://www.youtube.com/watch?v=MD6NXTrA3xo)
Insight into how Bloomberg operates over 1000 SolrClouds on Kubernetes using the operator.

**Lucidworks Apache Solr Support Policy**: [https://doc.lucidworks.com/policies/8qwsme/](https://doc.lucidworks.com/policies/8qwsme/)
Commercial support options if you need guaranteed SLAs and expert assistance.

## Conclusion

The journey from "don't run Solr on Kubernetes" to "run thousands of Solr clusters on Kubernetes" represents a fundamental shift in how we approach stateful workloads in containerized environments. The Apache Solr Operator embodies this shift—transforming Solr's operational complexity from a manual burden into automated, declarative infrastructure.

For Planton, choosing the Apache Solr Operator as the default deployment method means users get production-ready SolrCloud clusters without navigating the deployment landscape themselves. The operator handles the hard parts—ZooKeeper coordination, safe rolling updates, shard rebalancing, backup orchestration—while the Planton API simplifies the configuration surface to what actually matters for most deployments.

The result is a deployment experience that respects Solr's complexity without overwhelming operators with it. That's the promise of Kubernetes operators, fully realized for SolrCloud.


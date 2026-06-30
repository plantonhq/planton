# Deploying PostgreSQL Operators on Kubernetes

## Introduction

"Just use a managed database service." That was the prevailing wisdom for years when teams asked about running PostgreSQL in production. The conventional answer: databases are stateful, complex, and operationally demanding—let AWS, GCP, or Azure handle it.

But something shifted. As Kubernetes matured and organizations committed to multi-cloud or on-premises strategies, the question evolved from "should we run databases on Kubernetes?" to "how do we run them well?" The answer lies in **Kubernetes Operators**—controllers that codify operational expertise into software.

For PostgreSQL specifically, operators have reached production maturity. They handle the hard problems: automated failover, streaming replication, point-in-time recovery, rolling upgrades, and connection pooling. What once required a dedicated DBA team now runs autonomously in your cluster.

This document explores the landscape of deploying PostgreSQL operators on Kubernetes, with a focus on the **Zalando Postgres Operator**—one of the most battle-tested solutions in the ecosystem. We'll examine deployment methods from anti-patterns to production-ready approaches, understand the operator's architecture and capabilities, and explain why it remains a strong choice for teams building database services on Kubernetes.

The Planton `KubernetesZalandoPostgresOperator` resource provides a declarative API for deploying and managing the Zalando Postgres Operator across Kubernetes clusters.

## The Deployment Landscape

Deploying a PostgreSQL operator isn't fundamentally different from deploying any other Kubernetes workload, but the stakes are higher. Get it wrong, and you're managing stateful infrastructure with potential data loss exposure. Get it right, and you have a self-healing database platform.

### Level 0: The Manual Path

**Approach:** Apply raw YAML manifests with `kubectl apply -f` commands, pulling files directly from the operator's GitHub repository.

**What it solves:** Complete transparency. Every resource (CRDs, RBAC, ConfigMaps, Deployments) is visible and under your control. No hidden abstractions, no external tooling dependencies.

**What it doesn't solve:** Operational continuity. Upgrades become manual archaeology—which manifests changed between versions? Did you remember to apply CRD updates before the operator deployment? Configuration changes require editing raw YAML and reapplying, creating drift between environments.

**Verdict:** Fine for initial exploration or demos. Unsustainable for production where consistency and repeatability matter.

### Level 1: Kustomize Overlays

**Approach:** Use the operator's provided Kustomization or build your own, applying with `kubectl apply -k`.

**What it solves:** Bundled deployment of correct manifest versions. Kustomize overlays let you patch namespaces, resource limits, or other fields without editing base files. GitOps-friendly structure.

**What it doesn't solve:** Limited templating power. CRD installation ordering isn't automatic (you might still apply CRDs separately). Major customizations require complex patches.

**Verdict:** A step up from raw manifests. Good for teams already using Kustomize, but still requires careful orchestration of multi-step installations.

### Level 2: Helm Charts

**Approach:** Deploy via the official Helm chart (or community charts), configuring through a `values.yaml` file.

**What it solves:** Streamlined installation with templating built in. Configuration via structured values rather than YAML patches. Helm manages CRD lifecycle (with caveats). Repeatable across environments—same chart, different values files for dev/staging/prod.

**What it doesn't solve:** Another tool in the stack (Helm itself requires learning). Not all operator settings may be exposed through chart values (though official charts are comprehensive). CRD upgrades require special handling (Helm historically doesn't auto-upgrade CRDs).

**Verdict:** Production-appropriate for most teams. The official Zalando Helm chart has been used successfully in large-scale deployments. Balance between simplicity and flexibility.

### Level 3: GitOps (ArgoCD/Flux)

**Approach:** Declare the operator installation in Git (either as Helm charts or raw manifests), let ArgoCD or Flux CD continuously reconcile the desired state.

**What it solves:** Everything-as-code philosophy. Automated drift detection and remediation. Audit trail via Git history. Multi-cluster deployments from a single source of truth. ArgoCD sync waves can handle CRD ordering (apply CRDs before the operator deployment).

**What it doesn't solve:** Initial setup complexity—GitOps pipelines require configuration themselves. Must handle CRD lifecycle carefully (ArgoCD won't auto-prune CRDs by default, which is usually good). Requires team commitment to Git-driven operations.

**Verdict:** The gold standard for production Kubernetes. Once established, GitOps provides unmatched operational consistency. Particularly valuable for teams managing multiple clusters or environments.

**Pitfall to avoid:** ArgoCD's auto-prune can be dangerous if misconfigured—you don't want it accidentally deleting PostgreSQL cluster CRs during cleanup. Use annotations and sync policies deliberately.

### Level 4: Infrastructure-as-Code Abstractions

**Approach:** Manage operator deployment through Terraform, Pulumi, Crossplane, or similar IaC tools.

**What it solves:** Integration with broader infrastructure management. Enforce dependencies (e.g., operator must be present before database clusters). Fits teams already standardized on Terraform or Pulumi for all infrastructure. Planton's approach of wrapping the operator in a `KubernetesZalandoPostgresOperator` CRD exemplifies this pattern—operators become composable infrastructure components.

**What it doesn't solve:** Added abstraction layer means debugging complexity (chase issues through IaC state, Kubernetes resources, and operator logs). Module/provider version compatibility lag (IaC modules might not expose the latest operator features immediately).

**Verdict:** Powerful for platform teams building opinionated infrastructure layers. The abstraction pays off when you're managing dozens of operators across many clusters. However, requires discipline to avoid becoming a black box.

**Pitfall to avoid:** Race conditions on first install. Applying a PostgreSQL CRD and a cluster CR in the same Terraform/Pulumi apply can fail if the CRD isn't established yet. Use explicit `depends_on` or phased deployments.

## Production Patterns: What Actually Works

The deployment method (Helm, GitOps, IaC) is less important than the operational patterns you establish. Here's what matters in production:

### Multi-Tenancy Decisions

**Single Operator, Cluster-Wide:** One operator deployment watches all namespaces. Uses `ClusterRole` RBAC, manages PostgreSQL clusters across teams/environments. Easier to upgrade (one deployment), but failure affects all clusters. Use `teamId` prefixes to prevent naming collisions.

**Multiple Operators, Isolated:** One operator per namespace or per environment. Strong isolation—dev team's operator issues don't impact production. Allows environment-specific configuration (e.g., backups disabled in dev). Trade-off: more operator instances to maintain.

Most organizations start with single cluster-wide operator and split only when scale or isolation requirements demand it.

### The 80/20 of Operator Configuration

Not all configuration is created equal. The vast majority of deployments need only these essentials:

**Operator-level (OperatorConfiguration CRD):**
- `docker_image`: Spilo image version (determines default PostgreSQL version)
- `watched_namespace`: `"*"` for cluster-wide, specific namespace for isolation
- `enable_pod_antiaffinity`: `true` in production (spreads replicas across nodes)
- `enable_pod_disruption_budget`: `true` (safeguards against mass evictions)
- Resource requests/limits for the operator pod itself

**Backup configuration (per-cluster or global via environment variables):**
- S3-compatible bucket and endpoint (`WAL_S3_BUCKET`, `AWS_ENDPOINT`)
- Credentials (via Kubernetes Secrets referenced as `secretKeyRef`)
- Backup schedule (`BACKUP_SCHEDULE` cron expression)
- WAL-G enable flags (`USE_WALG_BACKUP=true`, `USE_WALG_RESTORE=true`)

**PostgreSQL cluster configuration (per cluster CR):**
- Team ID and cluster name
- Number of instances (replicas)
- Volume size and storage class
- Resource requests/limits
- Users and databases
- Optional: Connection pooler, TLS secrets

Everything else—Teams API integration, custom sidecar images, etcd for Patroni DCS, advanced RBAC—is edge-case configuration. The defaults are sensible.

### Backup and Disaster Recovery: The S3-Compatible Approach

The Zalando operator uses **WAL-G** for continuous archiving and point-in-time recovery. This is production-proven:

**Continuous WAL archiving** pushes write-ahead logs to S3-compatible storage in real-time. **Scheduled base backups** (typically nightly) provide restore anchors. Together, they enable recovery to any point in time.

**Why S3-compatible?** Because it's ubiquitous. AWS S3, Google Cloud Storage (with interop mode), Azure Blob, Cloudflare R2, MinIO—all work. The operator doesn't care; it just needs an S3 API endpoint and credentials.

**Cloudflare R2 deserves special mention:** Zero egress fees make it ideal for backups. AWS S3 charges significant egress when you pull a 100GB backup for disaster recovery. R2 doesn't. Storage costs are also lower (\$0.015/GB-month vs S3's \$0.023/GB-month). The trade-off: R2 lacks tiered storage classes like Glacier for long-term archival. But for active PITR windows (7-30 days), R2 is compelling.

**Configuration is environment variables:** Set `WAL_S3_BUCKET`, `AWS_ENDPOINT` (for non-AWS), `AWS_ACCESS_KEY_ID` (via Secret), and `BACKUP_SCHEDULE`. WAL-G handles the rest—compressing WAL segments, uploading base backups, managing retention (though S3 lifecycle policies are more reliable than WAL-G's retention logic).

**Point-in-time recovery** works by creating a new cluster CR with a `clone` section specifying the source cluster's UID and a timestamp. The operator fetches the nearest base backup before that timestamp and replays WAL to the exact moment. This is how you test DR scenarios or roll back from bad migrations.

### Security and RBAC

The operator needs broad permissions (managing StatefulSets, Secrets, Services cluster-wide), so deploy it in a protected namespace accessible only to platform admins. PostgreSQL pods themselves run as non-root (UID 1000) with an `fsGroup` for volume ownership.

Backup credentials deserve special care. Use **Kubernetes Secrets**, not ConfigMaps. Limit IAM permissions to specific buckets. If running on AWS EKS, use IRSA (IAM Roles for Service Accounts) instead of long-lived access keys.

For production, enable **delete protection annotations**. The operator can require explicit annotations on a cluster CR before allowing deletion—prevents accidental `kubectl delete` from wiping databases.

### High Availability Notes

**At the database level:** Patroni (built into Spilo images) handles automated failover using Kubernetes API objects as its distributed consensus store. If the primary pod fails, Patroni promotes a replica within ~30 seconds. The operator isn't in the failover path—this continues even if the operator is down.

**At the operator level:** The operator itself typically runs as a single instance. Kubernetes will restart it automatically if it crashes. This is fine—the operator's job is orchestration (scaling, upgrades, configuration sync), not live failover handling. Patroni doesn't need the operator for HA.

## Alternatives: CloudNativePG, Crunchy, Percona

The Kubernetes PostgreSQL operator landscape is mature. Zalando's operator isn't alone:

### CloudNativePG (CNPG)

A CNCF Sandbox project, Apache 2.0 licensed. **Key difference:** Built-in failover logic (no Patroni), lighter weight. Strong Kubernetes-native focus (e.g., uses PVC snapshots for backups). Backed by EnterpriseDB.

**When to choose:** If you want CNCF governance and a pure-Go operator. If you prefer avoiding Patroni's DCS dependencies (CNPG uses Kubernetes leases directly).

**Why Zalando might still win:** Patroni's years of production hardening. Spilo's extensive extension library (PostGIS, TimescaleDB, pg_partman pre-installed). Established community knowledge base.

### Crunchy Postgres Operator (PGO)

Backed by Crunchy Data (commercial Postgres support vendor). Uses **pgBackRest** for backups instead of WAL-G—arguably more enterprise-proven for very large databases. Historically had some image licensing requirements (developer program click-through).

**When to choose:** If you need vendor support contracts. If you prefer pgBackRest's dedicated backup management over WAL-G.

**Why Zalando might still win:** Fully open with no image license strings. Simpler backup configuration (environment variables vs pgBackRest repo pods).

### Percona Operator for PostgreSQL

Percona's offering, also uses Patroni + pgBackRest. Apache 2.0 licensed. Part of Percona's broader database portfolio (they do MySQL, MongoDB too).

**When to choose:** If you're already in the Percona ecosystem. If you want their monitoring (PMM) integration.

**Why Zalando might still win:** Longer production track record (Zalando's operator has been managing hundreds of clusters at scale since 2018). More comprehensive extension set in Spilo.

### The Zalando Advantage

**Battle-tested at scale:** Zalando runs this operator for thousands of PostgreSQL clusters in production. That operational burn-in is invaluable—edge cases have been discovered and fixed.

**Patroni's robustness:** Using Patroni means leveraging the most mature PostgreSQL HA solution in the ecosystem. Split-brain scenarios, network partitions, sync replication modes—Patroni handles them elegantly.

**Feature completeness:** Streaming replication, PITR, cloning, connection pooling (PgBouncer), volume resizing, TLS, in-place major version upgrades. It's all there.

**Open source purity:** MIT licensed code, no image licensing gotchas. No community vs enterprise edition splits.

**Extension-rich Spilo image:** PostGIS, TimescaleDB, pgAudit, pg_cron, and more—pre-installed. For platform teams building database-as-a-service, this saves significant effort.

All major operators (Zalando, CNPG, Crunchy, Percona) now support the essential features (HA, backups, self-healing, upgrades). The choice often comes down to specific preferences: Patroni vs built-in failover, WAL-G vs pgBackRest, vendor support vs pure open source.

## Planton's Choice: Zalando as the Foundation

For a platform like Planton—building multi-cloud IaC abstractions over Kubernetes—the Zalando Postgres Operator provides a rock-solid foundation for offering PostgreSQL as a service.

**Why it fits:**

1. **Production-proven reliability:** When you're abstracting infrastructure for others to consume, you need components with extensive burn-in. Zalando's operator has it.

2. **Comprehensive feature set:** Platform users need HA, backups, pooling, and upgrade automation. Zalando delivers all of it without gaps.

3. **Operational simplicity:** Backup configuration via environment variables, straightforward RBAC, clear upgrade paths—platform operators appreciate this.

4. **Extension versatility:** Different teams need different extensions. Spilo's pre-packaged approach (PostGIS for geo teams, TimescaleDB for time-series workloads) reduces friction.

5. **Open source alignment:** No licensing concerns means Planton can package and distribute this freely, whether for public cloud, private cloud, or on-premises deployments.

The abstraction layer—Planton's `KubernetesZalandoPostgresOperator` API—wraps deployment complexity (Helm charts, RBAC, namespace setup) into a declarative interface. Teams using Planton don't need to understand operator internals; they declare intent via protobuf APIs, and the platform handles the rest.

This is the promise of modern infrastructure platforms: take battle-tested open source components (like Zalando's operator), wrap them in ergonomic APIs, and deliver them consistently across clouds.

## Conclusion

Running PostgreSQL on Kubernetes stopped being a question of "if" years ago. The question is "how well?"

The shift from manual database administration to operator-managed automation represents a paradigm change. Operations that once required runbooks and escalations—failover, backup verification, replica scaling, major version upgrades—now run autonomously. The DBA role hasn't disappeared; it's been codified into Kubernetes controllers.

The Zalando Postgres Operator embodies this evolution. Built from production experience at a large-scale e-commerce company, refined by an active open source community, and hardened across thousands of deployments, it's not experimental technology. It's the infrastructure many organizations quietly run in production today.

For teams building platforms like Planton, or organizations standardizing their database infrastructure on Kubernetes, the Zalando operator offers the right balance: comprehensive features without unnecessary complexity, open source values without vendor lock-in, and a track record that speaks louder than marketing claims.

The real test of infrastructure isn't how it performs during normal operation—it's how it handles failure. When a node dies, when a network partition occurs, when someone accidentally deletes a critical table and needs a point-in-time restore—that's when architectural choices matter. The Zalando Postgres Operator, with Patroni's failover logic and WAL-G's recovery capabilities, has passed this test repeatedly.

Running databases on Kubernetes is no longer the bold frontier. It's the expected baseline. Choose your operator wisely.

---

*For implementation details on configuring the operator, setting up backup encryption, or designing multi-region disaster recovery strategies, refer to the [Zalando Postgres Operator official documentation](https://opensource.zalando.com/postgres-operator/).*


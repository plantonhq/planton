# Deploying Harbor on Kubernetes: A Guide to Production-Ready Architecture

## Introduction

For years, the conventional wisdom around deploying stateful applications like container registries on Kubernetes was straightforward: don't. The complexity of managing databases, caches, and persistent storage—coupled with the risk of catastrophic data loss—made managed services the only sane choice for production.

Harbor changed that calculation, but not in the way you might expect.

The Harbor registry is architected as a deliberately **stateless** application. Every critical piece of state—metadata (PostgreSQL), cache (Redis), and artifact blobs (Object Storage)—is externalized to purpose-built backends. The Harbor components themselves (Core API, Registry, Portal, Jobservice) are pure compute, designed to be horizontally scaled, load-balanced, and treated as cattle, not pets.

This architectural decision has profound implications: deploying Harbor on Kubernetes isn't about managing a complex stateful system. It's about orchestrating a set of stateless services that *delegate* state management to the right tools for the job. Do this correctly, and you get a production-grade, highly available container registry. Get it wrong, and you end up with the single most common anti-pattern in the Harbor ecosystem: the "all-in-one" deployment that bundles PostgreSQL and Redis as in-cluster StatefulSets and uses filesystem PVCs for artifact storage.

This guide explains the maturity spectrum of Harbor deployment methods, from anti-patterns to production-ready solutions, and clarifies what Planton supports and why.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern (Bundled All-in-One)

**What it is:** Running `helm install harbor harbor/harbor` with default settings.

**What you get:** A single-command installation that deploys Harbor with bundled PostgreSQL and Redis as in-cluster StatefulSets, using ReadWriteOnce (RWO) PersistentVolumeClaims for filesystem storage.

**Why it fails in production:**

1. **Data Loss Risk**: The lifecycle of your stateful components (database, cache, artifacts) becomes tied to the Helm release lifecycle. A `helm uninstall` or a failed `helm upgrade` can accidentally delete StatefulSets and their associated PVCs, permanently destroying all registry metadata and artifacts.

2. **No High Availability**: The default `filesystem` storage backend provisions an RWO volume, which can only be mounted by a single pod on a single node. This makes it impossible to scale the Registry component beyond one replica, eliminating any possibility of HA.

3. **No Automated Backups**: The bundled database has no automated backup, no Point-in-Time Recovery (PITR), and no failover strategy. Your entire registry metadata lives on a single PostgreSQL pod.

**Verdict:** Acceptable for local development or "hello world" testing. A production anti-pattern that risks data loss and cannot scale.

---

### Level 1: Externalized Dependencies (Basic Production)

**What it is:** Using the Helm chart with external PostgreSQL, external Redis, and object storage (S3, GCS, Azure Blob).

**Configuration changes:**
```yaml
database:
  type: external
  external:
    host: harbor-db.example.rds.amazonaws.com
    port: 5432
    username: harbor
    coreDatabase: harbor_db
    existingSecret: harbor-db-creds

redis:
  type: external
  external:
    host: harbor-redis.cache.amazonaws.com
    port: 6379
    existingSecret: harbor-redis-creds

persistence:
  imageChartStorage:
    type: s3
    s3:
      region: us-west-2
      bucket: my-harbor-artifacts
      existingSecret: harbor-s3-creds
```

**What you get:** A deployment where Harbor's stateless components (Core, Registry, Portal, Jobservice) run in Kubernetes, but all state is managed by purpose-built external services.

**Why this works:**

- **Decoupled Lifecycle**: The database, cache, and artifact storage exist independently of the Kubernetes cluster. You can destroy and recreate the cluster without losing any registry data.

- **Managed Resilience**: Using managed services like AWS RDS (PostgreSQL) and ElastiCache (Redis) gives you automated backups, multi-AZ replication, PITR, and transparent failover—features that would be complex and error-prone to implement in-cluster.

- **Shared Storage Backend**: Object storage (S3, GCS, Azure Blob) is ReadWriteMany by design. All Registry pod replicas can read and write simultaneously, enabling horizontal scaling.

**Trade-offs:**
- **Cloud Provider Lock-in**: Requires managed cloud services (RDS, ElastiCache, S3).
- **Higher Cost**: Managed services are more expensive than self-hosted alternatives.
- **Network Dependencies**: Harbor pods must have network access to external cloud endpoints.

**Verdict:** The baseline production pattern. This is the minimum viable architecture for a resilient, production-ready Harbor deployment.

---

### Level 2: Cloud-Agnostic In-Cluster HA (Advanced Production)

**What it is:** Using Kubernetes-native operators for PostgreSQL (e.g., CloudNativePG, Crunchy Data) and Redis to provide "RDS-like" managed services inside the cluster, while still using object storage for artifacts.

**Why this exists:** Teams that need cloud-agnostic deployments, want to avoid vendor lock-in, or operate in environments where external managed services aren't available (on-prem, air-gapped).

**Configuration approach:**
```yaml
database:
  type: external
  external:
    host: harbor-postgres-rw.postgres-operator.svc.cluster.local
    port: 5432
    username: harbor
    coreDatabase: harbor_db
    existingSecret: harbor-db-creds  # Managed by CloudNativePG

redis:
  type: external
  external:
    host: harbor-redis-ha.redis-operator.svc.cluster.local
    port: 6379
    existingSecret: harbor-redis-creds

persistence:
  imageChartStorage:
    type: s3  # Still requires object storage
    s3:
      region: us-west-2
      bucket: my-harbor-artifacts
      existingSecret: harbor-s3-creds
```

**What you get:** Database and cache HA within the cluster, declaratively managed by operators that automate failover, backups, and replication.

**Why this works:**

- **Cloud Agnostic**: PostgreSQL and Redis are managed as Kubernetes CRDs. The same deployment works on AWS, GCP, Azure, or on-prem.
- **Declarative HA**: Operators like CloudNativePG provide automated HA, automated backups to object storage, and PITR—matching many RDS features.
- **Data Stays in Cluster**: For regulatory or operational reasons, some teams prefer keeping all data within the Kubernetes boundary.

**Trade-offs:**
- **Operational Complexity**: You're now operating database and cache infrastructure. This requires cluster-admin privileges and expertise.
- **Still Requires Object Storage**: Even in this model, artifact storage *must* be object storage (S3, GCS, etc.) for HA. Filesystem PVCs remain an anti-pattern.

**Verdict:** A sophisticated production pattern for teams with Kubernetes expertise and requirements for cloud-agnostic or air-gapped deployments. More complex to operate than Level 1, but provides greater control and portability.

---

### Level 3: Operator-Managed Lifecycle (The Declarative Future)

**What it is:** Using the official [goharbor/harbor-operator](https://github.com/goharbor/harbor-operator) to manage Harbor as a Kubernetes Custom Resource Definition (CRD).

**Configuration approach:**
```yaml
apiVersion: goharbor.io/v1beta1
kind: Harbor
metadata:
  name: production-harbor
spec:
  externalURL: https://harbor.mycompany.com
  expose:
    core:
      ingress:
        host: harbor.mycompany.com
  database:
    kind: PostgreSQL
    spec:
      external:
        host: harbor-db.rds.amazonaws.com
        port: 5432
        username: harbor
        passwordRef: harbor-db-secret
  redis:
    kind: Redis
    spec:
      external:
        addr: harbor-redis.cache.amazonaws.com:6379
        passwordRef: harbor-redis-secret
  storage:
    kind: S3
    spec:
      bucket: my-harbor-artifacts
      region: us-west-2
      credentialsRef: harbor-s3-secret
```

**What you get:** A declarative, CRD-based API that manages the entire Harbor application lifecycle—installation, upgrades, reconciliation, and (future roadmap) automated backup/restore and auto-scaling.

**Why this exists:** Helm is excellent for Day 1 (installation), but it "falls short" for Day 2 operations (ongoing reconciliation, state management, and operational tasks). The operator provides a state-reconciling, declarative control plane for Harbor.

**Current maturity:** The operator is newer and less widely adopted than the Helm chart. It represents the strategic direction of the Harbor project but is not yet the default for most production deployments.

**Verdict:** The future of declarative Harbor management. Worth evaluating for new deployments, especially if you value operator-based Day 2 lifecycle management over Helm's imperative upgrade model.

---

## Deployment Method Comparison

| Method | Primary Use Case | Maturity | Production Readiness | Key Constraint |
|:-------|:----------------|:---------|:--------------------|:---------------|
| **Bundled All-in-One** (Default Helm) | Dev/Testing | High | ❌ Not production-ready | Risk of data loss; no HA |
| **Externalized Dependencies** (Helm + external services) | Production (Cloud) | High | ✅ Production-ready | Requires managed cloud services |
| **In-Cluster HA Operators** (Helm + CloudNativePG/Crunchy) | Production (Cloud-agnostic) | Medium | ✅ Production-ready | High operational complexity |
| **Harbor Operator** (CRD-based) | Day 2 Lifecycle Management | Medium | ⚠️ Emerging | Less mature; fewer adopters |
| **Raw YAML Manifests** | N/A | N/A | ❌ Anti-pattern | Intractable complexity |

---

## Special Considerations

### The Object Storage Redirect Behavior

A critical, non-obvious configuration for production Harbor deployments is the **storage redirect** behavior.

By default, when a client requests an image layer (a "blob"), the Harbor Registry component issues an **HTTP 302 redirect**, telling the client to download the blob directly from the S3 or GCS URL. This is highly performant—Harbor never proxies multi-gigabyte images, and object storage serves content at massive scale.

However, in high-security or air-gapped environments, this redirect is a problem. It forces clients (CI runners, developer machines, Kubernetes nodes) to have direct network access and firewall exceptions to the cloud provider's object storage endpoints.

**The solution:** Set `persistence.imageChartStorage.disableredirect: true` in the Helm chart. This forces Harbor to *proxy* all artifact downloads. Clients only ever connect to the Harbor `externalURL`. This is slower and more resource-intensive for Harbor but is mandatory for many secure environments.

A production IaC framework must expose this boolean toggle prominently.

---

### Non-Root Images: The OpenShift and Security-First Pattern

The official `goharbor/harbor-helm` chart uses container images that default to running as the `root` user. In security-hardened environments like Red Hat OpenShift, containers are forbidden from running as root by default using Security Context Constraints (SCCs).

The **Bitnami Harbor chart** (`bitnami/harbor`) solves this at the source by using Bitnami's non-root container images. This is a security-first posture that makes Harbor deployable on OpenShift and other restricted environments without complex `securityContext` overrides.

For platform-agnostic IaC frameworks that must support diverse security policies, adopting the Bitnami chart's non-root pattern is a critical consideration.

---

### Ingress vs. Gateway API: The Future of Traffic Management

Today, Harbor is exposed via the standard Kubernetes **Ingress** API, with TLS termination handled by ingress controllers like NGINX, Traefik, or Istio Gateway.

The new **Kubernetes Gateway API** is the official successor to Ingress. It provides:
- More robust, role-oriented traffic management (Gateway for cluster admins, HTTPRoute for app teams)
- First-class support for advanced routing (header-based, weighted, multi-backend)
- A more expressive API for modern service mesh and ingress controller capabilities

For a new IaC framework, designing an API that can output **HTTPRoute** resources in addition to Ingress would be a forward-looking, future-proof design.

---

## What Planton Supports

Planton's `HarborKubernetes` API is designed around the **Level 1: Externalized Dependencies** production pattern as the default, with support for in-cluster HA operators (Level 2) as an advanced option.

### Design Philosophy

The API is **not** a 1:1 mapping of the Helm chart's `values.yaml`. Instead, it is an opinionated, structured abstraction based on the **80/20 principle**: expose the 20% of configuration fields that 80% of production deployments actually need, and guide users toward production-ready patterns by default.

**Core principles:**

1. **Externalized State is Default**: The API uses protobuf `oneof` types to force an explicit choice between `external` (production) and `internal` (dev-only) for database, cache, and storage.

2. **Secrets-First**: Credentials for databases, caches, and object storage are referenced as Kubernetes Secrets, never inline strings.

3. **HA-Ready**: Replica counts for stateless components (Core, Registry, Jobservice) are exposed as top-level fields to make horizontal scaling straightforward.

4. **Security Hardened**: The API encourages non-root images, HTTPS-only `externalURL`, and OIDC authentication as first-class fields.

5. **Minimal Surface Area**: Advanced tunables (internal TLS, custom component images, deep Trivy configuration) are intentionally omitted. These are Day 2 configurations best managed via the Harbor web UI or secondary CRDs.

### What's Included (The "80%")

- **Service Exposure**: `externalURL`, ingress configuration, TLS settings
- **External Database**: PostgreSQL endpoint, credentials (via Secret), database name
- **External Cache**: Redis endpoint, credentials (via Secret)
- **External Object Storage**: S3/GCS/Azure configuration, credentials (via Secret), `disableRedirect` toggle
- **HA Replicas**: Replica counts for Core, Registry, Jobservice
- **Security**: Admin password (via Secret), OIDC provider configuration
- **Ingress Annotations**: For cert-manager, ingress class, and other controller-specific settings

### What's Omitted (The "20%")

- **Component Image Overrides**: Users should deploy tested, holistic Harbor versions, not mix-and-match component images.
- **Internal TLS**: Made obsolete by service mesh mTLS, which provides transparent pod-to-pod encryption.
- **Bundled Database Tunables**: The API exposes a simple "enable bundled" toggle for dev mode, not the dozens of tunables for in-cluster PostgreSQL.
- **Advanced Scanner Configuration**: Deep Trivy settings, custom CA bundles—these are Day 2 configurations managed via the Harbor UI.

### Example: Minimal Production Configuration

```protobuf
message HarborKubernetes {
  string external_url = 1;  // https://harbor.mycompany.com
  string admin_password_secret = 2;  // K8s Secret name

  message Exposure {
    message Ingress {
      string host = 1;
      map<string, string> annotations = 2;  // cert-manager, ingress class
      string tls_secret_name = 3;
    }
    oneof type {
      Ingress ingress = 1;
      NodePort node_port = 2;  // Dev-only
    }
  }
  Exposure exposure = 3;

  message Database {
    message External {
      string host = 1;
      int32 port = 2;
      string database_name = 3;
      string credentials_secret = 4;
    }
    message Internal {
      bool enabled = 1;  // Dev-only
    }
    oneof config {
      External external = 1;
      Internal internal = 2;
    }
  }
  Database database = 4;

  message Redis {
    message External {
      string host = 1;
      int32 port = 2;
      string credentials_secret = 3;
    }
    message Internal {
      bool enabled = 1;  // Dev-only
    }
    oneof config {
      External external = 1;
      Internal internal = 2;
    }
  }
  Redis redis = 5;

  message Storage {
    message S3 {
      string bucket = 1;
      string region = 2;
      string credentials_secret = 3;
    }
    message GCS {
      string bucket = 1;
      string credentials_secret_key_json = 2;
    }
    message Azure {
      string container = 1;
      string credentials_secret = 2;
    }
    message Filesystem {
      string size = 1;  // Dev-only
      string storage_class = 2;
    }
    oneof backend {
      S3 s3 = 1;
      GCS gcs = 2;
      Azure azure = 3;
      Filesystem filesystem = 4;
    }
    bool disable_redirect = 5;  // Critical for air-gapped/secure environments
  }
  Storage storage = 6;

  message Replicas {
    int32 core = 1 [default = 1];
    int32 jobservice = 2 [default = 1];
    int32 registry = 3 [default = 1];
  }
  Replicas replicas = 7;

  message OIDC {
    string provider_name = 1;
    string provider_endpoint = 2;
    string client_id = 3;
    string client_secret = 4;  // K8s Secret name
    string scope = 5;
    string username_claim = 6;
  }
  OIDC oidc = 8;
}
```

---

## Conclusion: The Paradigm Shift

The key insight for deploying Harbor on Kubernetes is this: **Harbor is not a stateful application**. It is a stateless orchestration layer that delegates all state management to purpose-built backends.

This architectural decision is what makes production-grade Harbor deployments viable on Kubernetes. The stateless components (Core, Registry, Portal, Jobservice) scale horizontally like any modern microservice. The stateful components (database, cache, artifacts) are managed by the tools designed for that job—managed cloud services, Kubernetes operators, or object storage.

The "all-in-one" bundled deployment is not a simplified version of production Harbor. It's a fundamentally different architecture that violates the core design principle: externalized state. Treating it as a "starting point" that can be "upgraded" to production is a mistake. Production Harbor is externalized from day one.

Planton's `HarborKubernetes` API codifies this principle. It makes the production pattern the natural path and requires explicit opt-in for dev-mode bundled components. This is how IaC should work: guiding users toward resilient architectures by default, not documenting anti-patterns alongside best practices and expecting users to choose correctly.

Deploy Harbor correctly, and you get a production-grade, highly available, secure container registry. Deploy it incorrectly, and you get a ticking time bomb waiting for the next `helm upgrade` to wipe your artifacts. Choose wisely.

---

## Further Reading

For deeper implementation details on specific deployment patterns, see:
- [Harbor Helm Chart Configuration Guide](./harbor-helm-guide.md) (planned)
- [Deploying Harbor with CloudNativePG](./cloudnativepg-integration.md) (planned)
- [Harbor Backup and Disaster Recovery Strategies](./backup-dr-guide.md) (planned)
- [Configuring OIDC Authentication Declaratively](./oidc-configuration.md) (planned)


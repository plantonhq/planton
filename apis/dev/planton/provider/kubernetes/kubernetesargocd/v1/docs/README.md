# Deploying Argo CD on Kubernetes: From Getting Started to Production

## Introduction: The GitOps Control Plane Dilemma

Argo CD is a declarative GitOps continuous delivery tool for Kubernetes. It's the control plane that watches your Git repositories and ensures your clusters match what's declared in code. But here's the irony: while Argo CD excels at managing *other* applications declaratively, deploying Argo CD *itself* is where many teams stumble.

The "getting started" guides make it look trivial—a single `kubectl apply` command. Yet production deployments require careful consideration of high availability, security, observability, and lifecycle management. The gap between "hello world" and "production ready" is substantial.

This document explains the deployment landscape for Argo CD, from anti-patterns to production-ready approaches, and clarifies what Planton abstracts and why.

## The Deployment Maturity Spectrum

### Level 0: The "Getting Started" Manifest Install (Anti-Pattern)

The most basic installation—widely cited in tutorials—involves applying raw YAML manifests directly:

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

This imperative approach is a **production anti-pattern** for three critical reasons:

**No Versioning**: The URL points to a `stable` or `main` branch, not an immutable release. You're applying "whatever is currently at that URL," which can introduce breaking changes without warning. This is not idempotent or reproducible.

**Configuration Hell**: Customizing this installation means downloading a multi-thousand-line YAML file, manually editing it, and maintaining your "forked" version. When Argo CD releases updates, you must manually rebase your changes—a significant maintenance burden.

**No Lifecycle Management**: The upgrade path is to `kubectl apply` a *new* version of the manifest, with no transactional guarantees and no simple rollback path.

**Verdict**: Suitable only for temporary demos and local development. Never use this in production.

### Level 1: The Kustomize-Based Install

Kustomize provides a declarative way to customize Kubernetes manifests using patches. The official Argo CD documentation supports this approach by allowing you to reference the official manifest URL as a remote resource:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: argocd
resources:
  - https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.2/manifests/install.yaml

patches:
  - patch: |-
      - op: replace
        path: /spec/replicas
        value: 3
    target:
      kind: Deployment
      name: argocd-repo-server
```

This is a significant improvement over raw manifests—it allows version-controlled, declarative customization. Teams standardized on Kustomize can manage Argo CD using familiar patch-based workflows.

**However**, it still has limitations:
- Still relies on a remote manifest URL, which can be fragile
- No built-in versioning or rollback of the "package"
- Lacks the robust packaging and lifecycle management of Helm
- Less community adoption for the specific task of *installing* Argo CD (vs. deploying apps *with* Argo CD)

**Verdict**: A viable option for teams deeply committed to pure-Kustomize environments, but Helm is a more robust choice for package management.

### Level 2: The Production Standard (Helm Chart)

The **most common, flexible, and robust method** for production deployments is the official, community-maintained Helm chart: `argo/argo-cd`.

Helm functions as a **package manager for Kubernetes**, solving all the deficiencies of manual manifest installs:

**Versioning**: Deployments are pinned to a specific, immutable chart version:
```bash
helm install argo-cd argo/argo-cd --version 5.51.6
```

**Configuration**: The entire configuration surface of Argo CD—including HA, SSO, Ingress, metrics, and resource limits—is abstracted into a single, well-documented `values.yaml` file. You customize by overriding values, not by patching multi-thousand-line manifests.

**Lifecycle Management**: Helm provides managed `helm upgrade` and `helm rollback` commands, treating the Argo CD installation as a versioned "release" with clear upgrade and rollback paths.

The official Helm chart is:
- Hosted at [https://argoproj.github.io/argo-helm](https://argoproj.github.io/argo-helm)
- Designated as "official" by Artifact Hub
- The de facto standard in the Argo CD community

**For installing a complex, third-party application like Argo CD, Helm is the purpose-built tool.** While Kustomize is powerful for *patching*, Helm is designed for *packaging* and distributing configurable software releases.

**Verdict**: This is the production standard and what Planton abstracts.

### Level 3: The Operator-Managed Install (Delegated Lifecycle)

A Kubernetes Operator is a custom controller that extends the Kubernetes API to manage an application's lifecycle automatically. The `argocd-operator` (in the argoproj-labs GitHub organization) introduces an `ArgoCD` Custom Resource Definition (CRD). You define a single high-level `ArgoCD` resource, and the operator handles provisioning, configuring, and upgrading all underlying components.

Two operator variants exist:

**Open Source**: `argoproj-labs/argocd-operator` - Community-driven, available for any Kubernetes cluster.

**Enterprise**: Red Hat OpenShift GitOps Operator - Commercially supported, deeply integrated into the OpenShift platform.

The operator provides the most "hands-off" automated lifecycle, actively reconciling cluster state to match your `ArgoCD` resource definition.

**Why not abstract the Operator?**

The Operator's `ArgoCD` CRD is *already an abstraction* over the Helm chart. It's a higher-level API with ~100 fields, compared to the Helm chart's 1000+ lines of configuration. Building a Planton abstraction on top of the Operator would be an **abstraction of an abstraction**—an architectural anti-pattern that:
- Limits configuration granularity
- Creates debugging complexity (three layers to troubleshoot)
- Depends on the operator's update cycle for new Argo CD features
- Has lower community adoption than Helm

**Verdict**: Excellent for managed, "hands-off" platforms (especially OpenShift), but Helm provides better alignment with Planton's philosophy of abstracting the most widely-used, production-ready patterns.

### Level 4: GitOps Bootstrapping (Self-Management)

This is an advanced, fully-declarative pattern where **Argo CD manages its own deployment**:

1. Install a minimal Argo CD instance (using Helm or Kustomize)
2. Point it to a Git repository containing an Application manifest
3. That Application defines the *full, production-grade* Argo CD deployment
4. The minimal instance "upgrades" itself to the full version, which then takes over self-management

This is the core of the **"App of Apps" pattern**, where a single "root" Application deploys all other applications, including the cluster's core infrastructure.

**ApplicationSets** are a more powerful, generator-based evolution of this pattern, enabling multi-cluster, multi-environment deployments from a single template.

**Argo CD Autopilot** was an opinionated tool created to automate this bootstrapping. However, community feedback suggests it hasn't gained significant traction, with reports of issues and stalled development. The *pattern* (App of Apps) is sound, but the `argocd-autopilot` tool may not be a stable foundation.

**Verdict**: The ultimate GitOps goal—Argo CD managing itself. Requires maturity in GitOps practices and is best implemented manually using the App of Apps pattern rather than relying on autopilot tooling.

## Production-Ready Requirements

A production Argo CD deployment is not just about getting it installed—it's about ensuring it's reliable, secure, and observable.

### High Availability (HA)

Argo CD is designed to be "largely stateless." All core data (Applications, AppProjects, clusters, repositories) are stored as Kubernetes resources in etcd. The built-in Redis instance is a *non-essential cache* for login sessions, UI state, and diff results. If Redis is lost, it can be rebuilt without data loss.

**HA Components**:

**Application Controller**: Responsible for reconciliation. Scale to 2+ replicas for HA. For massive-scale deployments (thousands of applications), the controller supports sharding and tuning of processing queues.
- Helm key: `controller.replicas: 2`

**Repo Server**: Clones Git repositories and generates manifests. Often CPU and memory-intensive. Scale to 2+ replicas for HA.
- Helm key: `repoServer.replicas: 2`

**Redis**: A single-replica Redis is a single point of failure. The official Helm chart includes a `redis-ha` sub-chart that deploys a highly available Redis cluster (typically 3 replicas) with pod anti-affinity.
- Helm key: `redis-ha.enabled: true`

The official HA manifests and Helm chart include pod anti-affinity rules to ensure replicas are scheduled on different nodes, preventing a single node failure from taking down the service.

### Security Configuration

**The Default Admin Pitfall**: The most common security mistake is leaving the default `admin` user enabled with its auto-generated password. This must be disabled once SSO is configured.

**SSO Integration (OIDC/SAML)**: This is a **non-negotiable production requirement**. Argo CD integrates with external OIDC providers (Okta, Auth0, Google, Keycloak) by delegating authentication to its bundled Dex instance or connecting directly to the provider.

Configuration is in the `argocd-cm` ConfigMap:
- Helm key: `configs.cm."oidc.config"`

**RBAC Configuration**: SSO only handles *authentication* (verifying who you are). RBAC handles *authorization* (defining what you can do). Configuration is in the `argocd-rbac-cm` ConfigMap.

A production setup must map SSO groups to Argo CD roles:

```csv
# Map SSO group 'platform-admins' to built-in role 'role:admin'
g, platform-admins, role:admin

# Map SSO group 'developers' to read-only access
g, developers, role:readonly
```

Helm key: `configs.rbac."policy.csv"`

### Service Exposure: Mastering Ingress

Exposing the `argocd-server` component is notoriously tricky. The server serves both **gRPC traffic** (for the CLI) and **HTTPS traffic** (for the Web UI) over the *same port* (443). This dual-protocol requirement breaks most standard Layer-7 Ingress controllers, which expect to terminate HTTP/S traffic.

**Two Solutions**:

**Solution 1: SSL Passthrough (Recommended)**

This is the cleanest and most common solution. The Ingress controller operates at Layer 4, passing raw, encrypted TCP traffic directly to the `argocd-server` pod. The server itself handles TLS termination and protocol detection (gRPC vs. HTTPS).

Requirements:
- Annotation: `nginx.ingress.kubernetes.io/ssl-passthrough: "true"` on the Ingress object
- The `ingress-nginx` controller must be started with the `--enable-ssl-passthrough` flag

**Solution 2: Dual Ingress / Dual Hostname**

Terminate SSL at the Ingress controller using two separate Ingress objects on different hostnames:
- `argocd.example.com` for the HTTPS UI
- `grpc.argocd.example.com` for the gRPC CLI

This is more complex—requiring two DNS records, two TLS certificates, and users must configure their CLI to point to the separate gRPC hostname.

**Planton's approach**: The abstraction is built around the simpler and more common **SSL Passthrough pattern**.

### Observability

Argo CD exposes comprehensive Prometheus metrics from its components:

**Application Controller** (`argocd-metrics:8082`): Exposes the most critical metrics:
- `argocd_app_info`: Reports health and sync status of all applications
- `argocd_app_sync_total`: Counter for sync operations

**API Server** (`argocd-server-metrics:8083`): Exposes metrics for API requests, gRPC calls, and login attempts.

The Helm chart can automatically create `ServiceMonitor` resources, enabling seamless metric scraping by the Prometheus Operator. A community Grafana dashboard is available in the Argo CD repository.

### Backup and Disaster Recovery

**Critical Insight**: The `argocd admin export` command is often cited as the backup method, but this is misleading. This command **does not** back up the Kubernetes resources deployed by Argo CD.

The true, **GitOps-native disaster recovery strategy** is to **treat Git as the single source of truth and the backup**. In a proper GitOps model, the cluster state is ephemeral and can be completely rebuilt from Git.

**A robust DR plan requires**:
1. All `Application` and `AppProject` definitions stored in Git
2. The Argo CD installation itself (Helm `values.yaml`) stored in Git and applied via the App of Apps pattern
3. All secrets (repository SSH keys, cluster credentials) managed declaratively via ExternalSecrets or SealedSecrets

If all configuration is in code, there is nothing stateful to back up. The `argocd admin export` command is only useful for backing up state that was *not* declaratively defined in Git (e.g., GPG keys, or repositories added manually via the UI).

## Production Configuration: The 80/20 Analysis

The official `argo/argo-cd` Helm chart's `values.yaml` is over 1000 lines long. A successful production deployment depends on configuring the **essential 20%** of fields, while the remaining 80% of advanced, niche fields can be left as defaults.

### The Essential 20% (Must Configure)

These fields are what virtually all production users must set:

**High Availability & Replicas**:
- `controller.replicas`: Default 1, must be 2+ for production
- `repoServer.replicas`: Default 1, must be 2+ for production
- `server.replicas`: Default 1, must be 2+ for production
- `redis-ha.enabled`: Default false, must be true for production

**Ingress & Server**:
- `server.ingress.enabled`: Default false, must be true
- `server.ingress.hosts`: The FQDN (e.g., `argocd.mycompany.com`)
- `server.ingress.tls.secretName`: Name of the Kubernetes TLS secret
- `server.ingress.annotations`: Critical for enabling SSL Passthrough

**SSO & RBAC**:
- `configs.cm."oidc.config"`: Multi-line string for OIDC provider details
- `configs.rbac."policy.csv"`: Multi-line string for mapping SSO groups to roles
- `configs.cm."admin.enabled"`: Default true, should be false once SSO is configured

**Resource Sizing**:
- `controller.resources`, `repoServer.resources`, `server.resources`: The chart provides **no default resource requests or limits**—a critical omission. Safe starting point: 250m-500m CPU and 256Mi-512Mi memory for each component. Rule of thumb: provision at least 300MB of memory per managed cluster.

### The Advanced 80% (Can Skip or Default)

These fields are powerful but only needed for specific, advanced use cases:

- `configs.cm."configManagementPlugins"`: For custom manifest-generation tools (e.g., sops-kustomize)
- `configs.cm."resource.customizations"`: For custom health checks for CRDs or ignoring fields during diffing
- `controller.sharding.enabled`: For sharding the controller across thousands of applications
- `repoServer.parallelismLimit`: Tuning for performance-intensive monorepo operations
- `server.extensions`: Tech-preview feature for UI plugins

## Reference Configurations

### Example 1: Dev/Staging Configuration

Minimal setup, disables HA, uses basic Ingress:

```yaml
# values-dev.yaml
# Non-HA setup: single replicas, standard Redis
redis-ha:
  enabled: false

controller:
  replicas: 1
repoServer:
  replicas: 1
server:
  replicas: 1

# Basic Ingress (simple TLS termination)
server:
  ingress:
    enabled: true
    ingressClassName: nginx
    hosts:
      - argocd-dev.mycompany.com
    tls:
      - secretName: argocd-dev-tls
        hosts:
          - argocd-dev.mycompany.com

# Use default admin for simplicity
configs:
  cm:
    admin.enabled: true
```

### Example 2: Production HA/SSO Configuration

Full production setup with HA, SSL Passthrough, OIDC, and RBAC:

```yaml
# values-prod.yaml
# --- High Availability ---
redis-ha:
  enabled: true

controller:
  replicas: 2
repoServer:
  replicas: 2
server:
  replicas: 2

# --- Production-Grade Resources ---
controller:
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

repoServer:
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

# --- Ingress with SSL Passthrough ---
server:
  ingress:
    enabled: true
    ingressClassName: nginx
    hosts:
      - argocd.mycompany.com
    annotations:
      nginx.ingress.kubernetes.io/ssl-passthrough: "true"
      nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    tls:
      - secretName: argocd-prod-tls
        hosts:
          - argocd.mycompany.com

# --- SSO & RBAC Configuration ---
configs:
  cm:
    # Disable the local 'admin' user
    admin.enabled: false

    # OIDC Configuration (Example for Okta)
    oidc.config: |
      name: Okta
      issuer: https://my-org.okta.com
      clientID: 0oa123abcdeFGHIJklm
      clientSecret: $oidc.okta.clientSecret  # From argocd-secret
      requestedScopes:
        - openid
        - profile
        - email
        - groups

  # RBAC policy to map SSO groups to Argo CD roles
  rbac:
    policy.csv: |
      # Give 'platform-admins' SSO group full admin rights
      g, platform-admins, role:admin
      
      # Give 'app-developers' SSO group read-only rights
      g, app-developers, role:readonly
```

## Production Readiness Checklist

| Category | Best Practice | Common Pitfall | Helm Key(s) |
|----------|---------------|----------------|-------------|
| **HA** | Enable HA Redis and multi-replica controllers | Running single-replica components | `redis-ha.enabled: true`<br/>`controller.replicas: 2`<br/>`repoServer.replicas: 2` |
| **Security (Admin)** | Disable default admin user after SSO is configured | Leaving default admin password active | `configs.cm."admin.enabled": false` |
| **Security (SSO)** | Configure OIDC/SAML integration with corporate IdP | No SSO, using shared admin account | `configs.cm."oidc.config": "..."` |
| **Security (RBAC)** | Map SSO groups to Argo CD roles | Giving all SSO users `role:admin` | `configs.rbac."policy.csv": "g,..."` |
| **Ingress** | Use SSL Passthrough for dual gRPC/HTTPS | Using NodePort or LoadBalancer, or fighting with L7 Ingress | `server.ingress.enabled: true`<br/>`server.ingress.annotations` |
| **DR** | Store all configs (Apps, Projects, Argo install) in Git | Relying on `argocd admin export` for DR | N/A (GitOps pattern) |
| **Metrics** | Enable ServiceMonitor for Prometheus | No monitoring of app sync/health state | `controller.metrics.serviceMonitor.enabled: true` |

## Planton's Choice: Helm Chart Abstraction

Planton's `ArgocdKubernetes` resource abstracts the official **Helm chart's `values.yaml` API**, not the Operator's CRD.

**Why Helm over the Operator?**

1. **Widest Community Adoption**: Helm is the de facto standard for installing Argo CD in production
2. **Full Configuration Surface**: Provides 100% of the configuration granularity required for production
3. **Direct Abstraction**: Avoids the anti-pattern of building an abstraction on top of another abstraction (the Operator)
4. **Stability**: The Helm chart's API is mature and stable; the Operator introduces an additional moving part

**The Planton API Philosophy**: Our protobuf API is not a 1:1 mapping of the Helm values. Instead, it's a **structured, opinionated API that abstracts user intent**. For example:

- `ha.enabled: true` in our API translates to setting multiple Helm keys: `controller.replicas`, `repoServer.replicas`, `redis-ha.enabled`
- `sso` and `rbac_policies` fields abstract the complex multi-line string configurations in `configs.cm` and `configs.rbac`
- Default SSL Passthrough Ingress pattern with simple hostname and TLS secret configuration

This simplifies the user experience and enforces production best practices by default, while still allowing advanced users to pass through custom Ingress annotations or override specific Helm values for specialized use cases.

## Licensing and Distribution

**License**: Argo CD is licensed under **Apache 2.0**, a permissive open-source license that allows for commercial use, modification, and distribution. Fully compatible with open-source platforms like Planton.

**Container Images**: Official images are hosted at `quay.io/argoproj/argocd`. **Critical note**: Images on Docker Hub are deprecated and no longer updated—always use the quay.io registry.

**Commercial Offerings**: 
- **Akuity**: Primary commercial vendor for Argo CD, co-founded by Argo project creator Jesse Suen. Offers managed, enterprise-grade Argo CD platform.
- **Red Hat OpenShift GitOps**: Commercially supported, packaged version of the Argo CD Operator, deeply integrated into the OpenShift platform.

## Conclusion: The GitOps Meta-Lesson

There's a certain elegance in using Argo CD to manage Argo CD itself—the App of Apps pattern taken to its logical conclusion. But before achieving that GitOps nirvana, you must first deploy Argo CD in a way that's reliable, secure, and maintainable.

The path from "getting started" to production-ready is not about adding complexity for its own sake. It's about understanding that the control plane for your entire deployment infrastructure deserves the same rigor you apply to your applications: high availability, strong authentication, declarative configuration, and infrastructure as code.

Helm provides the robust package management foundation for this journey. The Operator offers an alternative for fully automated lifecycle management. Kustomize serves teams committed to patch-based workflows. And raw manifests... well, they serve as a reminder that "simple to start" doesn't always mean "safe to run."

Planton abstracts the Helm chart—the production standard—while providing the structured, opinionated API that makes production best practices the default, not an afterthought. Because the best platform is one where doing the right thing is also the easiest thing.


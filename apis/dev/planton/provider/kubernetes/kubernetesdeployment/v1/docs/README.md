# Deploying Microservices to Kubernetes: The Evolution from Simple to Production-Ready

## Introduction

For years, the conventional wisdom about deploying microservices to Kubernetes has been: "Just create a Deployment, set an image, and you're good to go." This is technically true—your service will run. But there's a massive gap between "it runs" and "it runs reliably in production."

The uncomfortable reality is that a basic Kubernetes Deployment with just an image and replica count is an **anti-pattern for production**. Without health probes, the orchestrator routes traffic to dead pods. Without resource limits, you're assigned to the BestEffort QoS class, meaning your service is the first to be evicted when the node runs out of memory. Without a Pod Disruption Budget, a routine node drain can terminate all your replicas simultaneously, causing an outage.

This document explores the maturity spectrum of Kubernetes microservice deployments—from these dangerous anti-patterns to production-ready configurations that deliver genuine zero-downtime reliability. More importantly, it explains **why** Planton's KubernetesDeployment API is designed the way it is: to make the **secure, resilient path the easiest path**, by encoding battle-tested production practices into sensible defaults.

## The Deployment Maturity Spectrum

Understanding deployment approaches as a progression from basic to production-ready helps clarify what truly matters.

### Level 0: The Anti-Pattern

A minimal Deployment manifest looks deceptively simple:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: app
        image: myapp:v1.0
        ports:
        - containerPort: 8080
```

This will deploy. It will even serve traffic. But it's a production disaster waiting to happen:

- **No health probes**: Kubernetes has no way to know if your application is alive or ready. Dead pods stay in the load balancer rotation, causing user-facing errors.
- **No resource requests or limits**: Your pod is assigned to the **BestEffort** QoS class, making it the first candidate for eviction when the node experiences resource pressure.
- **Single replica**: Zero high availability. Any update or node failure guarantees downtime.
- **No graceful shutdown**: Rolling updates will drop in-flight connections as pods are abruptly terminated.

This is the "quickstart" example you see in tutorials. It should never reach production.

### Level 1: Basic Health and Resources

The first step toward production readiness is adding observability and resource guarantees:

```yaml
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: app
        image: myapp:v1.0
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

**What this solves:**
- **Health probes**: Kubernetes now knows when your app is alive (liveness) and ready to serve traffic (readiness).
- **Resource management**: Your pod has a guaranteed CPU and memory allocation and is promoted to the **Burstable** QoS class.
- **Multiple replicas**: Basic high availability during rolling updates.

**What's still missing:**
- **Startup probe**: If your application takes 30+ seconds to start, the liveness probe may prematurely kill it, creating a CrashLoopBackOff death spiral.
- **Graceful shutdown**: There's no preStop hook to drain connections before SIGTERM.
- **Pod Disruption Budget**: A node drain can still take down multiple replicas simultaneously.

### Level 2: Zero-Downtime Updates

Production services require carefully orchestrated updates that never drop a single request. This demands four interlocking components:

**1. Startup Probe Protection**

Slow-starting applications need a startup probe to protect them from premature liveness checks:

```yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 8080
  failureThreshold: 30
  periodSeconds: 10
```

This gives your application 5 minutes (30 × 10 seconds) to complete initialization before liveness checks begin.

**2. Readiness-Driven Rolling Updates**

The rolling update strategy must be configured to wait for new pods to pass readiness before terminating old ones:

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
```

With `maxUnavailable: 0`, Kubernetes guarantees that it will **never** terminate an old pod until a new pod is running and ready. This is critical for single-replica deployments, where the default `maxUnavailable: 25%` rounds up to 1, causing the only pod to be killed before the replacement is ready—guaranteed downtime.

**3. Graceful Shutdown Lifecycle**

When a pod is terminated, it's immediately removed from Service endpoints, but there's a race condition: Ingress controllers and kube-proxy need time to observe this change and update their routing tables. A `preStop` hook provides this critical delay:

```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 10"]
terminationGracePeriodSeconds: 45
```

The 10-second sleep keeps the pod alive long enough for routing tables to update. Meanwhile, your application must handle SIGTERM by stopping new requests and completing in-flight work within the termination grace period.

**4. Pod Disruption Budgets**

During voluntary disruptions (node maintenance, cluster scaling), a Pod Disruption Budget (PDB) prevents all replicas from being evicted simultaneously:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: my-service
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: my-service
```

This ensures node drains evict pods serially, maintaining high availability.

**Verdict:** This is the minimum configuration for genuine zero-downtime production deployments. However, it's still manually managed. Production teams need automation.

### Level 3: The Production Solution

A mature platform approach abstracts the complexity of Level 2 into intelligent defaults and environment-specific profiles.

**The Challenge:**
Most engineers understand the importance of these patterns, but implementing them consistently across dozens of microservices is error-prone. Forgetting a PDB or misconfiguring `maxUnavailable` leads to outages during cluster maintenance. Writing and maintaining these multi-resource manifests (Deployment, Service, Ingress, PDB, ServiceAccount, NetworkPolicy, ServiceMonitor) for every service is toil.

**The Solution: Opinionated IaC APIs**

Planton's KubernetesDeployment API solves this by providing environment profiles that encode production best practices into defaults:

**Development Profile**: Fast iteration, minimal overhead
- 1 replica
- No health probes (faster startup)
- No resource limits (BestEffort QoS)
- No PDB or network policies

**Production Profile**: Zero-downtime, security-first
- 3+ replicas (or HPA-managed)
- All three probes (startup, liveness, readiness) with battle-tested timeouts
- Guaranteed QoS (requests == limits for memory)
- Automatic PDB creation (`maxUnavailable: 1`)
- Graceful shutdown enabled (preStop hook, 45s termination grace)
- Non-root security context (`runAsNonRoot: true`)
- Default-deny network policies
- Automatic ServiceMonitor for Prometheus integration

By selecting `environment: production` in your configuration, you get all of these patterns automatically, without needing to understand the intricacies of PDBs or rolling update strategies. The platform makes the **correct choice the easy choice**.

## Configuration Management: Kustomize vs Helm

Once you understand what a production deployment requires, the next question is how to manage configuration across environments.

### Kustomize: YAML Patches

Kustomize is a Kubernetes-native tool that applies patches to a base configuration for different environments. It's "template-free"—you work directly with valid YAML.

**Pattern:**
```
base/
  deployment.yaml
  service.yaml
overlays/
  dev/
    kustomization.yaml (patches for dev)
  prod/
    kustomization.yaml (patches for prod)
```

**Strengths:**
- No templating language to learn
- Clean separation of concerns
- Native kubectl support (`kubectl apply -k`)

**Weaknesses:**
- Verbose for managing many microservices
- No release versioning or rollback mechanism
- DRY violations across service manifests

### Helm: Package Management

Helm packages applications into versioned charts and uses Go templates to generate manifests.

**The Library Chart Pattern:**
For microservice architectures, the **Helm Library Chart** pattern eliminates duplication. A single "library" chart defines templates for Deployment, Service, Ingress, etc. Individual microservice charts import this library and pass only service-specific values:

```yaml
# my-service/Chart.yaml
dependencies:
  - name: microservice-library
    version: 1.0.0
    repository: https://charts.company.com

# my-service/values.yaml
image: myapp:v1.0
port: 8080
replicas: 3
```

**Strengths:**
- DRY: One set of templates for all microservices
- Versioned releases with easy rollbacks
- Widely adopted ecosystem

**Weaknesses:**
- Go templating (`{{ include "..." }}`) is fragile and hard to debug
- Complexity increases with deeply nested templates
- Type safety is non-existent

**Planton's Position:**
The Helm Library Chart pattern and Planton's Protobuf-based API solve the **same problem**—abstracting common Kubernetes boilerplate into reusable definitions—but Protobuf APIs provide superior **type safety**, **programmatic validation**, and **cross-language stub generation**. Think of it as "Helm Library Charts, but as compiled code instead of brittle string templates."

## GitOps: Continuous Deployment at Scale

GitOps treats Git as the single source of truth for declarative infrastructure. A control-loop agent continuously reconciles cluster state to match Git, providing auditability and automated rollout.

### ArgoCD vs Flux CD

| Feature | ArgoCD | Flux CD |
|---------|---------|---------|
| **UI** | Comprehensive web UI for visualizing sync status, diffs, and app health | No UI (relies on CLI and Kubernetes APIs) |
| **Architecture** | Standalone application with its own RBAC system | Kubernetes-native, uses CRDs and standard RBAC |
| **Multi-tenancy** | First-class support via Application and AppProject CRDs | Manual setup via namespaces and RBAC |
| **Tooling** | Works with Helm, Kustomize, and raw YAML | Same (via CRDs: HelmRelease, Kustomization) |
| **Adoption** | Popular for teams transitioning to GitOps (visual UI eases onboarding) | Favored by platform engineers who prefer "building block" tooling |

**Both are production-ready.** ArgoCD's UI is invaluable for organizations prioritizing observability. Flux's CRD-driven model is ideal for teams who prefer declarative, `kubectl`-native workflows.

**Planton Integration:**
KubernetesDeployment generates standard Kubernetes manifests that work seamlessly with both tools:
1. Developer defines service in a Protobuf API
2. CI pipeline runs Planton to generate Kubernetes YAML
3. YAML is committed to a Git repository
4. ArgoCD or Flux detects changes and deploys automatically

## Autoscaling: Responding to Load

### Horizontal Pod Autoscaler (HPA)

HPA scales replicas based on observed metrics. It uses a simple ratio:

```
desiredReplicas = ceil(currentReplicas × (currentMetric / targetMetric))
```

**Key Insight: HPA scales based on `resource.requests`, not `limits`.**

This means CPU and memory requests must be set accurately. If requests are too low, HPA scales aggressively; too high, and it scales conservatively.

**The Average Trap:**
Scaling on average CPU is flawed. If one pod is pegged at 100% CPU while nine others are idle at 10%, the average is 19%—perfectly healthy by HPA's standards—but 10% of users experience catastrophic latency.

**Best Practice:**
Scale on **leading, business-relevant metrics** like request queue length, P99 latency, or message backlog, not lagging infrastructure metrics like CPU.

**Typical HPA Configuration:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-service
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

**Planton Default (Prod Profile):**
- Enabled by default
- Min replicas: 3, max replicas: 10
- Target CPU: 70% (conservative to allow burst capacity)

### Vertical Pod Autoscaler (VPA)

VPA adjusts resource requests and limits (scaling **up**) rather than adding replicas (scaling **out**). It's ideal for right-sizing workloads.

**Critical Limitation:**
VPA in "auto" mode **restarts pods** to apply new resource values, making it unsuitable for production without careful orchestration.

**Production Use Case:**
Run VPA in **recommendation-only mode** (`updateMode: "Off"`). It observes your workload and generates recommended resource values, which you can then use to optimize your Deployment's requests and fine-tune HPA thresholds.

### KEDA: Event-Driven Scaling

KEDA extends HPA to scale based on external event sources (Kafka topics, SQS queues, Pub/Sub subscriptions). Its "killer feature" is **scale-to-zero**: when the queue is empty, replicas drop to 0, saving costs. When messages arrive, KEDA scales up instantly.

**When to Use:**
- Asynchronous background workers
- Queue processors
- Batch jobs
- Any non-HTTP workload where scale-to-zero provides cost savings

**When NOT to Use:**
- HTTP APIs (you need at least 1 replica for immediate response)

| Autoscaler | Scales | Based On | Key Feature | Best For |
|------------|--------|----------|-------------|----------|
| **HPA** | Out (more pods) | CPU/Memory/Custom Metrics | Load-based scaling | Web services, APIs |
| **VPA** | Up (bigger pods) | CPU/Memory usage over time | Right-sizing | Stateful services, tuning HPA |
| **KEDA** | Out (more pods) | External events (queues, topics) | **Scale-to-zero** | Background workers, queue processors |

## Ingress and Traffic Management

### Ingress Controllers

An Ingress resource defines HTTP routing rules. An Ingress Controller is the reverse proxy that implements those rules.

**NGINX Ingress Controller:**
- De facto standard
- High performance, mature, widely supported
- **Use when:** You need a proven, universal solution

**Traefik:**
- Modern, cloud-native
- Automatic service discovery and dynamic configuration
- **Use when:** You prioritize ease of use and dynamic routing

**Cloud-Native Controllers (AWS ALB, GCP LBC):**
- Provisions managed cloud load balancers (e.g., AWS Application Load Balancer)
- Native integration with WAF, ACM, cloud IAM
- **Caution:** Can be expensive; some controllers create one LB per Ingress resource

**Planton Approach:**
Generate standard Ingress resources that work with **any** controller. The cluster administrator selects and configures the controller; the microservice author is decoupled from this infrastructure decision.

### TLS Certificate Automation

**cert-manager** is the standard for automating TLS certificates in Kubernetes.

**Pattern:**
1. Install cert-manager in the cluster
2. Create a `ClusterIssuer` for Let's Encrypt (or another CA)
3. Add a TLS block to your Ingress:
   ```yaml
   tls:
   - hosts:
     - myapp.example.com
     secretName: myapp-tls
   metadata:
     annotations:
       cert-manager.io/cluster-issuer: letsencrypt-prod
   ```
4. cert-manager automatically requests, validates, and renews certificates

**Planton Simplification:**
Setting `ingress.tls: true` in your KubernetesDeployment spec automatically generates the TLS block and cert-manager annotations, eliminating manual configuration.

### Service Mesh: When Is It Worth the Complexity?

Service meshes (Istio, Linkerd) provide three capabilities that are difficult to achieve otherwise:

1. **Observability**: Uniform metrics, logging, and tracing for all service-to-service communication
2. **Security**: Automatic mutual TLS (mTLS) for encrypted, authenticated inter-service communication
3. **Traffic Control**: Advanced L7 routing, retries, circuit breakers, canary deployments

**The Trade-Off:**
Service meshes add significant operational complexity and resource overhead (sidecar proxies in every pod).

**Istio:**
- Feature-rich, powerful, battle-tested (uses Envoy proxy)
- Supports VMs, not just Kubernetes
- Complex to install, operate, and debug

**Linkerd:**
- Lightweight, fast, easy to use
- Kubernetes-only, custom-built proxy
- Less flexible than Istio, but simpler

**The 80/20 Verdict:**
**80% of users do not need a service mesh.** The operational cost is immense. Only adopt a mesh if you have a **specific, critical need**:
- Hard compliance requirement for mTLS
- gRPC services at scale (gRPC's persistent HTTP/2 connections break L4 load balancing; meshes provide L7-aware gRPC load balancing)
- Fine-grained canary traffic splitting requirements

For HTTP microservices with standard observability tooling (Prometheus, Loki, OpenTelemetry), a mesh is overkill.

## Resource Management and Quality of Service

Kubernetes assigns every pod a **QoS class** based on its resource configuration. This class determines eviction priority when nodes run out of resources.

| QoS Class | Memory Configuration | CPU Configuration | Eviction Priority | Use Case |
|-----------|---------------------|-------------------|-------------------|----------|
| **Guaranteed** | requests == limits | requests == limits | Last (highest priority) | Critical production workloads (databases, stateful services) |
| **Burstable** | requests < limits | requests < limits or unset | Medium | Stateless microservices (allows CPU bursting) |
| **BestEffort** | No requests or limits | No requests or limits | First (lowest priority) | Dev/test, batch jobs |

### Production Recommendations

**Memory:**
- **Always** set `memory.requests` and `memory.limits` to the **same value** (requests == limits)
- Memory is incompressible; running out causes OOMKilled errors (exit code 137)
- This ensures Guaranteed QoS for memory, protecting your pods from eviction

**CPU:**
- **Always** set `cpu.requests` (required for HPA and scheduling)
- **Optionally** set `cpu.limits` to 2-3x the request, or leave it unset
- CPU is compressible; limits cause throttling, which can hurt latency
- Allowing bursting during traffic spikes is often preferable to hard throttling

**Typical Resource Allocations:**

| Workload Size | CPU Request | Memory Request/Limit | Use Case |
|---------------|-------------|----------------------|----------|
| **Small** | 100m | 128Mi | Lightweight APIs, sidecars |
| **Medium** | 500m | 512Mi | Standard microservices |
| **Large** | 1000m (1 CPU) | 1Gi | Compute-intensive services |

**Planton Prod Profile Defaults:**
- CPU request: 500m
- CPU limit: Unset (allows bursting)
- Memory request: 512Mi
- Memory limit: 512Mi (Guaranteed QoS for memory)

Developers should override these based on actual usage, but they provide a safe, production-ready starting point.

## Security: The 20% of Features That Matter

Most Kubernetes security guides are overwhelming. Here's the 20% that delivers 80% of the protection.

### 1. Run as Non-Root

The single most important security setting:

```yaml
securityContext:
  runAsNonRoot: true
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
```

This prevents privilege escalation attacks and container breakouts.

### 2. Network Policies: Default-Deny

By default, Kubernetes networking is "flat"—all pods can talk to all other pods. This is a security nightmare.

**Best Practice:**
Apply a **default-deny** NetworkPolicy to every namespace:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

Then explicitly allow-list only necessary traffic (e.g., "allow ingress from NGINX controller," "allow egress to DNS").

**Planton Prod Profile:**
Automatically generates default-deny policies and baseline allow-list rules. This is a profound security improvement delivered "for free."

### 3. Secrets Management

Kubernetes native Secrets are Base64-encoded (not encrypted) and stored in etcd. They're only secure if etcd encryption-at-rest is enabled and RBAC is locked down.

**Production Pattern: External Secrets Operator (ESO)**

1. Store secrets in a dedicated secret manager (Vault, AWS Secrets Manager, GCP Secret Manager)
2. Install the External Secrets Operator in your cluster
3. Grant the operator read-only access to the secret store
4. Create an `ExternalSecret` CRD that references the secret path
5. ESO fetches the secret and creates a native Kubernetes Secret
6. Your application mounts this Secret normally

This provides centralized, auditable secret management while keeping applications decoupled from the secret store.

### 4. Dedicated Service Accounts

By default, pods run as the `default` ServiceAccount, which often has excessive permissions.

**Best Practice:**
- Create a **dedicated ServiceAccount** for each microservice
- Set `automountServiceAccountToken: false` (only mount the token if the pod needs Kubernetes API access)
- If API access is needed, create a minimal Role with only the required permissions
- Bind the Role to the ServiceAccount via RoleBinding

**Planton:**
Automatically generates a dedicated ServiceAccount for every microservice with `automountServiceAccountToken: false` by default.

## Observability: Logs, Metrics, and Traces

Production microservices require three pillars of observability.

### Logging: Fluent Bit + Loki

**Application Responsibility:**
- Log to **stdout/stderr** (never to files)
- Use **structured logging** (JSON format)

**Platform Responsibility:**
- Run a lightweight log collector (Fluent Bit) as a DaemonSet on every node
- Forward logs to an aggregation backend (Loki is the modern, Prometheus-like solution)
- Visualize logs in Grafana

### Metrics: Prometheus + ServiceMonitor

**Application Responsibility:**
- Expose a `/metrics` endpoint in Prometheus text format
- Emit key business metrics (request rate, error rate, latency) and infrastructure metrics (memory usage, goroutines, etc.)

**Platform Responsibility:**
- Run Prometheus Operator in the cluster
- Create a `ServiceMonitor` CRD that tells Prometheus how to discover and scrape the service

**Planton Value-Add:**
Automatically generates both the Service and the ServiceMonitor for every microservice. This loops the service into the cluster's monitoring stack with **zero developer effort**.

### Tracing: OpenTelemetry + Jaeger/Tempo

**Application Responsibility:**
- Instrument code with the OpenTelemetry (OTel) SDK
- Propagate trace context (W3C Trace-Context headers) across service boundaries

**Platform Responsibility:**
- Run an OpenTelemetry Collector (as a DaemonSet or sidecar)
- Forward traces to a backend (Jaeger, Grafana Tempo)

**Planton Facilitation:**
Automatically injects environment variables for OTel SDK configuration (e.g., `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`).

| Pillar | Application Provides | Platform Provides | Planton Automates |
|--------|---------------------|-------------------|---------------------------|
| **Logging** | Logs to stdout/stderr (JSON) | Fluent Bit DaemonSet → Loki | N/A (cluster-level) |
| **Metrics** | `/metrics` endpoint (Prometheus format) | Prometheus Operator | Service + ServiceMonitor generation |
| **Tracing** | OTel SDK instrumentation | OTel Collector → Jaeger/Tempo | Inject OTEL_* environment variables |

## The 80/20 Configuration Principle

Just as APIs should focus on the 20% of configuration that 80% of users need, deployment documentation should clarify what's essential vs optional.

### Essential (Always Configure)

- **Image**: Container image repository and tag
- **Port**: Application listen port
- **Resources**: CPU and memory requests (minimum)
- **Readiness Probe**: At minimum, a readiness probe path

### Recommended (Production Deployments)

- **Replicas or HPA**: 3+ replicas, or HPA with min 3
- **Liveness and Startup Probes**: Complete health check triad
- **Resource Limits**: Memory limits (set equal to requests for Guaranteed QoS)
- **Graceful Shutdown**: Enabled (preStop hook + termination grace period)
- **PDB**: Automatic creation for multi-replica deployments
- **Security Context**: Non-root user, read-only filesystem
- **Network Policy**: Default-deny with explicit allow-list

### Advanced (Specific Use Cases)

- **Init Containers**: Database schema migrations, config pre-processing
- **Affinity/Anti-Affinity**: Topology constraints (spread across AZs, avoid noisy neighbors)
- **Custom Volumes**: Persistent storage, ConfigMap mounts beyond env vars
- **KEDA**: Scale-to-zero for queue processors

## Planton's Production-Ready Defaults

KubernetesDeployment is designed around the principle that **the secure, resilient path should be the easiest path**.

### Environment Profiles

| Configuration | dev | staging | prod |
|--------------|-----|---------|------|
| **Replicas** | 1 | 2 | 3 (HPA-managed) |
| **Probes** | None (fast iteration) | Readiness only | Startup + Readiness + Liveness |
| **QoS** | BestEffort (no resources) | Burstable | Guaranteed (memory), Burstable (CPU) |
| **HPA** | Disabled | Enabled (CPU @ 80%) | Enabled (CPU @ 70%, conservative) |
| **PDB** | Disabled | Disabled | Enabled (`maxUnavailable: 1`) |
| **Graceful Shutdown** | Disabled | Disabled | Enabled (preStop: sleep 10, 45s termination grace) |
| **Security Context** | Default (may run as root) | `runAsNonRoot: true` | `runAsNonRoot: true`, `readOnlyRootFilesystem: true` |
| **Network Policy** | None | None | Default-deny + baseline allow-list |
| **Observability** | Basic (logs to stdout) | Metrics (ServiceMonitor) | Full (logs + metrics + OTel env vars) |

### What This Means for Developers

**Development:**
```yaml
# Minimal config for fast local iteration
environment: dev
image: myapp:latest
port: 8080
```

**Production:**
```yaml
# Comprehensive production deployment with all best practices
environment: prod
image: myapp:v1.2.3
port: 8080
resources:
  cpu: 500m
  memory: 512Mi
readinessProbe:
  path: /ready
livenessProbe:
  path: /healthz
ingress:
  host: myapp.example.com
  tls: true
```

Setting `environment: prod` automatically enables:
- 3 replicas (or HPA management)
- All three health probes with production-tested timeouts
- Graceful shutdown lifecycle (preStop hook, 45s termination grace)
- Pod Disruption Budget (`maxUnavailable: 1`)
- Guaranteed QoS for memory
- Non-root security context
- Default-deny network policies
- Prometheus ServiceMonitor
- Dedicated ServiceAccount with `automountServiceAccountToken: false`

**Zero configuration overhead. Maximum reliability.**

## Conclusion: Complexity Hidden, Resilience Delivered

The gap between "runs in Kubernetes" and "runs reliably in production" is enormous. That gap is filled with health probes, Pod Disruption Budgets, graceful shutdown hooks, QoS classes, network policies, and observability integrations—dozens of interconnected resources and settings that must all work in concert.

Manually configuring this for every microservice is error-prone and toil-heavy. Copy-pasting manifests leads to configuration drift. Helm templates become fragile mazes of nested conditionals.

Planton's KubernetesDeployment API solves this by encoding production best practices into **environment-aware defaults**. Developers declare their intent at a high level; the platform generates complete, battle-tested Kubernetes manifests.

This is not about hiding Kubernetes. It's about **hiding complexity while preserving power**. Advanced users can override any default. But for the 80% use case—deploying a stateless HTTP microservice with production-grade reliability—the secure, resilient path is now also the **easiest** path.

The result: fewer outages, faster time-to-production, and engineering teams focused on building features instead of debugging PDB misconfigurations.

## Further Reading

- [Zero-Downtime Deployments Deep Dive](./zero-downtime-deployments.md) *(coming soon)*
- [Autoscaling Strategies for Different Workload Types](./autoscaling-strategies.md) *(coming soon)*
- [Security Hardening Checklist](./security-hardening.md) *(coming soon)*
- [Observability Stack Integration Guide](./observability-integration.md) *(coming soon)*


# Deploying Tekton on Kubernetes: From kubectl to Production

## The Production Reality

Five months running Tekton in production taught us: **installation is the easy part**. The hard part is integrating Tekton with your platform—streaming events to webhook receivers, exposing the dashboard securely, forwarding pod logs to your observability stack in real-time. When you `kubectl apply` Tekton's release YAML, you get a CI/CD engine. But you don't get the connective tissue that makes it production-ready.

This document explores how to deploy Tekton on Kubernetes, from the simplest kubectl approach to fully integrated production deployments. We'll examine the Tekton Operator (which automates lifecycle management), the critical configuration that most deployments need, and the integrations that transform Tekton from an isolated system into a platform component.

## The Deployment Spectrum

### Level 0: Manual Release YAML (Quick Start, Not Production)

The Tekton documentation starts here:

```bash
# Install Tekton Pipelines v1.11.2
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v1.11.2/release.yaml

# Install Tekton Triggers
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/previous/v0.25.0/release.yaml

# Install Tekton Dashboard
kubectl apply -f https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml
```

**What you get**: Tekton Pipelines controller, webhook, and CRDs in the `tekton-pipelines` namespace. Triggers and Dashboard in their respective namespaces. Everything runs with default settings.

**What you don't get**:
- No ingress for the dashboard (ClusterIP service only)
- No cloud events configuration (no webhook integration)
- No custom timeouts or service accounts
- No integration with your logging stack
- Manual upgrades (remember to update CRDs in correct order)

**Verdict**: Excellent for learning Tekton, running local experiments, or proof-of-concept demos. Not sustainable for production. Every configuration change means editing ConfigMaps post-install. Upgrades require coordination across multiple manifests. Suitable only if you have <5 pipelines and a single operator who knows kubectl.

### Level 1: Helm Charts (Better, Still Manual)

Community Helm charts exist for Tekton Pipelines:

```bash
helm repo add eddycharly https://eddycharly.github.io/helm-charts
helm install tekton-pipelines eddycharly/pipelines --version 1.15.0
```

**What you gain**:
- Values-based configuration (set defaults via `values.yaml`)
- Versioned deployments (Helm tracks what you installed)
- Easier upgrades (`helm upgrade`)
- Parameterization (ingress, service accounts, etc.)

**What you still manage manually**:
- ConfigMap settings for cloud events sink, timeouts, feature flags
- Dashboard ingress and authentication
- Log collection integration
- Ensuring compatibility between Pipelines/Triggers/Dashboard versions

**Verdict**: Improvement over raw kubectl. Good for teams already using Helm for everything. But Helm charts for Tekton are community-maintained (not official Tekton project), which means version lag and potential abandonment. Most Tekton users don't use Helm—they use the official manifests or the Operator.

### Level 2: Tekton Operator (Production Standard)

The Tekton Operator is an official Kubernetes operator that manages Tekton component lifecycle via CRDs:

```yaml
# Install Tekton Operator
kubectl apply -f https://storage.googleapis.com/tekton-releases/operator/previous/v0.78.0/release.yaml

# Configure Tekton via TektonConfig CRD
apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: config
spec:
  profile: all  # Installs Pipelines, Triggers, Dashboard
  targetNamespace: tekton-pipelines
```

**What the Operator manages**:
- Installation of Pipelines, Triggers, Dashboard with version compatibility
- Automatic creation of CRDs, controllers, webhooks, RBAC
- Unified configuration via TektonConfig CR
- Self-healing (recreates components if deleted)
- Simplified upgrades (update operator version or TektonConfig)

**Key insight**: The TektonConfig CR exposes most Tekton ConfigMap settings as first-class fields. Instead of editing `config-defaults` manually, you set:

```yaml
spec:
  pipeline:
    defaultServiceAccount: pipeline
    defaultTimeoutMinutes: 60
    defaultCloudEventsSink: "http://webhook-receiver.default.svc:8080/tekton"
    defaultPodTemplate:
      nodeSelector:
        workload: builds
```

The operator propagates these into Tekton's ConfigMaps. Change the CR, and the operator reconciles the ConfigMaps automatically.

**Advanced customization**: For settings the operator doesn't expose directly, use the `options` field:

```yaml
spec:
  pipeline:
    options:
      configMaps:
        config-events:  # Custom ConfigMap creation
          data:
            sink: "http://custom-sink.company.net"
            formats: "tektonv1"
      deployments:
        tekton-pipelines-controller:  # Inject env vars into controller
          spec:
            template:
              spec:
                containers:
                  - name: tekton-pipelines-controller
                    env:
                      - name: CONFIG_LOGGING_NAME
                        value: pipeline-config-logging
```

This pattern allows **operator-managed deployment with manual customization** where needed.

**What you still handle**:
- Dashboard ingress (operator deploys the dashboard pod, not the ingress)
- Authentication (Tekton Dashboard has no built-in auth)
- Log collection (DaemonSet for pod logs)
- Integration with your broader platform

**Verdict**: This is the production standard. The operator reduces operational toil while maintaining flexibility. Used extensively in OpenShift (Red Hat's distribution of Tekton) and increasingly in vanilla Kubernetes. If you're running Tekton long-term, the operator pays for itself in upgrade ease and config consistency.

## Critical Production Integrations

Installing Tekton is table stakes. Making it production-ready requires three integrations.

### 1. Cloud Events Sink: Wiring Tekton into Your Platform

Tekton can emit CloudEvents for every pipeline lifecycle transition:
- PipelineRun Started
- PipelineRun Running
- PipelineRun Succeeded/Failed
- TaskRun Started/Running/Succeeded/Failed

**Why this matters**: Your platform needs to react to pipeline results. When a build succeeds, update a deployment record. When a build fails, notify the developer. When a pipeline starts, show "building..." in your UI.

**Configuration (Tekton v1.8+)**:

Create or update the `config-events` ConfigMap (or use TektonConfig's `defaultCloudEventsSink` field):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-events
  namespace: tekton-pipelines
data:
  sink: "http://main.service-app-prod-tekton-webhooks-receiver/service-hub/tekton/cloud-event"
  formats: "tektonv1"
```

**How it works**: Tekton's controller sends HTTP POSTs to your sink URL. Each event is a CloudEvents v1.0 envelope with:
- Headers: `Ce-Type` (event type), `Ce-Source` (resource URI), `Ce-Id` (UUID), `Ce-Time`
- Body: Full PipelineRun or TaskRun JSON at the time of the event

**Example payload** for a successful PipelineRun:

```json
{
  "pipelineRun": {
    "metadata": { "name": "build-api-123", ... },
    "spec": { ... },
    "status": {
      "conditions": [{"type": "Succeeded", "status": "True"}],
      "startTime": "...",
      "completionTime": "..."
    }
  }
}
```

Your webhook receiver parses this to extract pipeline results, update databases, send notifications, etc.

**Reliability**: Tekton uses asynchronous delivery with exponential backoff retry. If your sink is down, Tekton will retry sending events without blocking pipeline execution. Events may arrive out-of-order during retries. For critical audit trails, consider also using Tekton Results (stores run history in a database) as a backup.

**Real-world usage at Planton**: The sink points to an internal service that updates the ServiceHub database with pipeline status and triggers downstream actions (deployment, notifications). This integration is **essential**—without it, Tekton runs in isolation with no way for the platform to know when builds complete.

### 2. Dashboard Exposure and Authentication

The Tekton Dashboard provides a web UI for viewing pipelines, logs, and run history. Out of the box, it's a ClusterIP service on port 9097—only accessible via `kubectl port-forward`.

**Production requirements**:
1. **External access** via Ingress, Gateway, or LoadBalancer
2. **Authentication** (Tekton Dashboard has no built-in auth)
3. **Read-only mode** for most users (safety)

**Exposing the Dashboard**:

Create an Ingress (or Gateway with HTTPRoute for Gateway API):

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tekton-dashboard
  namespace: tekton-pipelines
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - tekton.company.com
      secretName: tekton-dashboard-tls
  rules:
    - host: tekton.company.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: tekton-dashboard
                port:
                  number: 9097
```

**Authentication patterns**:

The Dashboard itself doesn't authenticate users. You must layer auth in front:

**Option A: OAuth2 Proxy (Recommended)**

Deploy an OAuth2 proxy that requires login (Google, Okta, GitHub):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oauth2-proxy
spec:
  template:
    spec:
      containers:
        - name: oauth2-proxy
          image: quay.io/oauth2-proxy/oauth2-proxy:latest
          args:
            - --provider=oidc
            - --oidc-issuer-url=https://accounts.google.com
            - --email-domain=company.com
            - --upstream=http://tekton-dashboard.tekton-pipelines.svc:9097
            - --http-address=0.0.0.0:4180
```

Point your Ingress to the proxy service instead of the dashboard directly. Users must authenticate before accessing the dashboard.

**Option B: Network-Based (Simpler, Less Secure)**

Expose dashboard only on internal network (VPN required) or use basic HTTP auth at the ingress. Quick to set up but lacks fine-grained user tracking.

**Read-only vs Read-write**:

Tekton Dashboard ships with two installation manifests:
- `release.yaml` - Read-only mode (default, safer)
- `release-full.yaml` - Read-write mode (allows creating/deleting resources)

In read-only mode, the dashboard's service account can only view Tekton resources. Users see pipeline status and logs but can't trigger new runs or delete pipelines. **Use read-only for most users**. Grant read-write access only to CI admins who need to manage pipelines via UI.

**Real-world at Planton**: Dashboard is critical for developers to debug failed builds. We expose it via Istio Gateway with HTTPS. Authentication is handled at the gateway level. Most users have read-only access; only platform engineers have write access via CLI/GitOps.

### 3. Log Collection: Real-Time Streaming and Long-Term Storage

Tekton task logs flow through Kubernetes container stdout/stderr. By default, you view them with `kubectl logs`. For production, you need centralized log collection and real-time streaming.

**The Log Pipeline**:

```
Tekton TaskRun Pods (stdout/stderr)
         ↓
Log Collector DaemonSet (Vector or Fluent Bit)
         ↓
Message Bus (NATS) or Log Store (Elasticsearch, Loki)
         ↓
Real-Time UI + Long-Term Archive
```

**Vector vs Fluent Bit**:

Both work as DaemonSets that tail pod logs and forward them. Key differences:

| Aspect | Vector | Fluent Bit |
|--------|--------|------------|
| **Language** | Rust (modern, high-performance) | C (battle-tested, low footprint) |
| **Configuration** | VRL (Vector Remap Language) - powerful transforms | Fluent config (simpler but less flexible) |
| **NATS Support** | Native NATS sink | NATS output plugin |
| **Memory** | ~50-100MB per node | ~30-50MB per node |
| **Community Momentum** | Growing (CNCF, backed by Datadog) | Established (widely deployed) |
| **Production Maturity** | Proven at scale, newer codebase | Proven for years, very stable |

**Planton's choice**: Migrated from Fluent Bit to Vector when switching from Kafka to NATS. Vector's native NATS sink and powerful transforms aligned better with the new architecture. Teams report that Vector "just works" with complex routing and is easier to debug than Fluent Bit's configuration syntax.

**Vector configuration for Tekton + NATS**:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vector-config
  namespace: logging
data:
  vector.yaml: |
    sources:
      kubernetes_logs:
        type: kubernetes_logs
        namespace_labels:
          - namespace
        pod_labels:
          - tekton.dev/pipelineRun
          - tekton.dev/taskRun
    
    transforms:
      filter_tekton:
        type: filter
        inputs: [kubernetes_logs]
        condition: '.kubernetes.namespace_labels.namespace == "tekton-pipelines"'
      
      parse_tekton_metadata:
        type: remap
        inputs: [filter_tekton]
        source: |
          .pipeline_run = .kubernetes.pod_labels."tekton.dev/pipelineRun"
          .task_run = .kubernetes.pod_labels."tekton.dev/taskRun"
    
    sinks:
      nats_stream:
        type: nats
        inputs: [parse_tekton_metadata]
        url: "nats://nats.messaging.svc:4222"
        subject: "tekton.logs.{{ pipeline_run }}"
        encoding:
          codec: json
```

**How this enables real-time logs**:
1. Vector captures all pod logs in `tekton-pipelines` namespace
2. Extracts pipeline/task metadata from pod labels
3. Publishes to NATS subject `tekton.logs.<pipeline-run-id>`
4. Your web console subscribes to that subject for live log streaming
5. Logs also archived to object storage (R2, S3) after pipeline completes

**DaemonSet deployment**: Deploy Vector as a DaemonSet with node-level access to `/var/log/containers`. Each node runs one Vector pod that tails logs for all pods on that node. This is the standard Kubernetes logging pattern—no changes to Tekton pods required.

**Alternative: Fluent Bit** works similarly. If you already run Fluent Bit cluster-wide, configure it to forward Tekton logs to your chosen destination. The configuration is more verbose but well-documented.

## Tekton Operator: Managed Lifecycle with Configuration Flexibility

The Tekton Operator shifts Tekton from "installed software" to "managed resource." Instead of applying release YAMLs, you create a TektonConfig CR:

```yaml
apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: config
spec:
  profile: all  # Pipelines + Triggers + Dashboard
  targetNamespace: tekton-pipelines
  
  pipeline:
    # Defaults for all PipelineRuns
    defaultServiceAccount: pipeline
    defaultTimeoutMinutes: 60
    defaultCloudEventsSink: "http://webhook.company.svc/tekton-events"
    
    # Global pod template
    defaultPodTemplate:
      nodeSelector:
        workload: ci-builds
      tolerations:
        - key: ci
          operator: Equal
          value: "true"
          effect: NoSchedule
    
    # Feature flags
    enableApiFields: stable  # or beta, alpha
    enableTektonOciBundles: true
```

**What happens**:
1. Operator reads TektonConfig
2. Determines which components to install (based on `profile`)
3. Deploys Tekton Pipelines, Triggers, Dashboard as specified
4. Creates ConfigMaps (`config-defaults`, `config-features`, etc.) with your settings
5. Monitors components, recreates if deleted
6. Handles upgrades when you update TektonConfig

**Upgrade workflow**:

```bash
# Update operator version
kubectl apply -f https://storage.googleapis.com/tekton-releases/operator/previous/v0.80.0/release.yaml

# Operator auto-updates TektonPipeline to compatible version
# Or explicitly set version in TektonConfig if you want control
```

**Advanced configuration via `options`**:

For settings not exposed as first-class TektonConfig fields, use `options` to inject custom ConfigMaps or environment variables:

```yaml
spec:
  pipeline:
    options:
      configMaps:
        config-logging:
          data:
            loglevel.controller: "info"
            loglevel.webhook: "debug"
            zap-logger-config: |
              {
                "level": "info",
                "encoding": "json",
                "outputPaths": ["stdout"],
                "errorOutputPaths": ["stderr"]
              }
      
      deployments:
        tekton-pipelines-controller:
          spec:
            template:
              spec:
                containers:
                  - name: tekton-pipelines-controller
                    env:
                      - name: CONFIG_LOGGING_NAME
                        value: config-logging
```

This override mechanism provides **escape hatches** for any setting. Most production deployments use a combination of first-class fields (for common settings) and `options` (for niche requirements).

## The 80/20 Configuration: What Production Actually Needs

After surveying production Tekton deployments, most configure the same 20% of settings:

### Essential Settings

**1. Service Account** (`defaultServiceAccount`)

Don't use the `default` service account. Create a dedicated `pipeline` service account with:
- ImagePullSecrets for your container registry
- RBAC to access Secrets, ConfigMaps, and deploy resources
- Cloud provider IAM bindings (if using Workload Identity or IRSA)

**2. Timeouts** (`defaultTimeoutMinutes`)

Tekton's default is 60 minutes. Adjust based on your workloads:
- Backend builds: 15-30 minutes
- Integration tests: 60-90 minutes
- Long-running migrations: 120+ minutes

Set a sensible default to prevent hung builds from consuming resources indefinitely.

**3. Cloud Events Sink** (`defaultCloudEventsSink`)

If you integrate Tekton with external systems (which most production platforms do), this is **mandatory**. Point it to your webhook receiver, event bus, or notification service.

**4. Resource Limits** (`defaultContainerResourceRequirements`)

Prevent runaway builds from hogging nodes:

```yaml
defaultContainerResourceRequirements:
  default:  # Applied to all step containers without explicit resources
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "2000m"
      memory: "2Gi"
```

This ensures every Tekton step has a CPU/memory ceiling. Adjust based on your typical build patterns.

**5. Node Placement** (`defaultPodTemplate`)

If you have dedicated build nodes (common in production):

```yaml
defaultPodTemplate:
  nodeSelector:
    pool: tekton-builds
  tolerations:
    - key: build-workload
      operator: Equal
      value: "true"
      effect: NoSchedule
```

All Tekton tasks run on your build node pool, isolated from application workloads.

### Dashboard Configuration

**6. Ingress Host** (Dashboard access)

Users need a URL like `https://tekton.company.com`. This requires an Ingress or Gateway resource pointing to `tekton-dashboard` service.

**7. Authentication** (Security)

Layer auth via OAuth2 Proxy or restrict network access. Never expose an unauthenticated Tekton Dashboard publicly.

**8. Read-Only Mode** (Default)

Install Dashboard with read-only permissions unless write access is explicitly needed. Most users just need to view logs and pipeline status.

### Log Collection

**9. Vector/Fluent Bit DaemonSet** (Log aggregation)

Deploy a cluster-wide log collector configured to:
- Tail logs from `tekton-pipelines` namespace
- Extract pipeline/task metadata from pod labels
- Forward to your logging backend (NATS, Elasticsearch, Loki)

**10. NATS Integration** (If using event-driven architecture)

Configure Vector to publish logs to NATS subjects:

```
tekton.logs.<pipeline-run-id>
```

Your UI subscribes to these subjects for real-time log streaming.

## What Planton Supports

The KubernetesTektonOperator component deploys Tekton via the official Tekton Operator with an opinionated, production-ready configuration:

**Deployment method**: Tekton Operator (Level 2)

**Why**: The operator handles component lifecycle, version compatibility, and upgrades. We layer additional configuration on top via TektonConfig and custom resources.

**Exposed configuration**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-ci
spec:
  operatorVersion: v0.78.0
  
  container:
    resources:  # Operator pod resources
      requests: {cpu: 100m, memory: 128Mi}
      limits: {cpu: 500m, memory: 512Mi}
  
  components:
    pipelines: true
    triggers: true
    dashboard: true
  
  # Future fields (to be added):
  # cloudEventsSink: "http://webhook-receiver.svc:8080/tekton"
  # dashboardIngress:
  #   enabled: true
  #   host: "tekton.company.com"
  #   tls: true
  # logCollection:
  #   enabled: true
  #   backend: nats
  #   natsURL: "nats://nats.messaging.svc:4222"
```

**Current limitations** (to be addressed):
- Cloud events sink must be configured manually via TektonConfig patch
- Dashboard ingress must be created separately
- Log collection (Vector/Fluent Bit) must be deployed independently

**Roadmap**: Add first-class fields for cloud events, dashboard exposure, and log collection to enable fully automated, production-ready Tekton deployment in a single manifest.

## Migration from Manual to Operator-Managed

If you're running a manual Tekton installation and want to migrate to the operator:

### Pre-Migration Checklist

1. **Export current configuration**:
   ```bash
   kubectl get configmap -n tekton-pipelines config-defaults -o yaml > config-backup.yaml
   kubectl get configmap -n tekton-pipelines feature-flags -o yaml > features-backup.yaml
   ```

2. **Document custom settings**:
   - Cloud events sink URL
   - Default service account name
   - Custom timeouts or pod templates
   - Enabled feature flags

3. **Prepare TektonConfig CR** with these settings translated to TektonConfig fields

### Migration Steps (5-10 minutes downtime)

**1. Quiesce pipelines**:
- Stop new pipeline triggers
- Wait for running PipelineRuns to complete
- Verify: `kubectl get pipelinerun -A --field-selector=status.conditions[0].status!=False,status.conditions[0].status!=True`

**2. Deploy Tekton Operator**:
```bash
kubectl apply -f https://storage.googleapis.com/tekton-releases/operator/previous/v0.78.0/release.yaml
```

**3. Remove old Tekton controllers** (but keep CRDs and CRs):
```bash
kubectl delete deployment -n tekton-pipelines tekton-pipelines-controller
kubectl delete deployment -n tekton-pipelines tekton-pipelines-webhook
kubectl delete deployment -n tekton-triggers tekton-triggers-controller
# DO NOT delete CRDs or PipelineRun/TaskRun resources
```

**4. Create TektonConfig CR** with your settings:
```bash
kubectl apply -f tekton-config.yaml
```

**5. Wait for operator to install components**:
```bash
kubectl get tektonconfig config -w
# Wait for READY=True
```

**6. Verify**:
```bash
# Check controllers running
kubectl get pods -n tekton-pipelines

# Verify ConfigMaps updated
kubectl get configmap -n tekton-pipelines config-defaults -o yaml

# Run test pipeline
kubectl create -f test-pipeline-run.yaml
```

### Rollback Plan

If migration fails:
1. Delete TektonConfig: `kubectl delete tektonconfig config`
2. Re-apply old Tekton manifests: `kubectl apply -f tekton-v1.1.0-release.yaml`
3. Controllers restart, pick up existing PipelineRuns

**Safety**: Don't delete CRDs or pipeline resources during migration. The operator uses the same CRDs as manual installation.

**Testing**: Practice this in a staging cluster first. The actual migration takes 5-10 minutes, but preparation prevents surprises.

## Production Best Practices

### Configuration Management

**Store TektonConfig in Git**: Treat TektonConfig like application code. Changes go through review, testing, GitOps deployment.

**Use `options` for custom settings**: If a ConfigMap field isn't exposed by TektonConfig, use `options` to inject it. This keeps configuration declarative.

**Version pinning**: Pin `operatorVersion` to a known-good version. Test upgrades in dev before rolling to production.

### Security

**Service accounts**: Use dedicated service accounts with least privilege. Don't grant cluster-admin to pipeline service accounts.

**Network policies**: Restrict which namespaces Tekton pods can communicate with. Pipelines shouldn't access production databases directly.

**Secret handling**: Store credentials in Kubernetes Secrets, reference in PipelineRuns. Never hardcode secrets in pipeline YAML.

### Monitoring

**Key metrics**:
- PipelineRun success/failure rate
- Average pipeline duration
- Queue depth (PipelineRuns pending)
- Webhook availability (for cloud events sink)

**Logging**:
- Aggregate Tekton controller logs (for debugging reconciliation issues)
- Aggregate task pod logs (for developer debugging)
- Alert on repeated controller errors

**Dashboard uptime**: Monitor dashboard availability and response time. It's a critical developer tool.

### Capacity Planning

**Node pool sizing**: If you have dedicated build nodes:
- Start with 3 nodes minimum (n2-standard-16 or similar)
- Configure autoscaling (1-10 nodes based on pipeline load)
- Use regular instances (not spot) if build latency matters

**PVC cleanup**: Tekton can accumulate workspace PVCs. Use TektonPruner or a cron job to clean up old PipelineRun artifacts.

**Event sink capacity**: Ensure your webhook receiver can handle event bursts. During high pipeline activity, Tekton emits hundreds of events per minute.

## Common Pitfalls

**1. Missing cloud events sink**: Pipelines run, but platform doesn't know when they complete. Always configure the sink if you integrate Tekton with external systems.

**2. Dashboard without auth**: Exposing dashboard publicly without authentication is a security risk. Even read-only mode reveals pipeline details.

**3. No resource limits**: Runaway builds can starve cluster resources. Always set default CPU/memory limits for step containers.

**4. Ignoring controller logs**: Tekton controller errors manifest as PipelineRuns stuck in pending. Monitor controller logs for early warning signs.

**5. Manual ConfigMap edits with Operator**: If you edit ConfigMaps directly while using the Operator, the operator will overwrite your changes. Use TektonConfig or `options` instead.

## The Paradigm Shift

Traditional CI/CD systems (Jenkins, GitLab CI) are monolithic: you install the system, and it comes with a UI, event hooks, log storage, everything. Tekton takes the opposite approach: it's a **building block**. You get pipeline execution (the engine) and must integrate the peripherals yourself.

This feels like extra work initially. But it's actually **freedom**. Want logs in Elasticsearch? Configure it. Want events in Kafka? Route them there. Want a custom dashboard? Build one using Tekton's API. Tekton doesn't lock you into a vendor's choices—it gives you a standards-based engine and lets you compose the rest.

The Tekton Operator makes this composition **declarative**. Instead of a bash script that installs Tekton, creates ConfigMaps, deploys dashboard, sets up ingress, you have a TektonConfig CR that describes the desired state. GitOps tools (ArgoCD, Flux) can manage it. Your platform can generate it. It's infrastructure as code, not imperative installation.

For production CI/CD platforms like Planton, this matters. We don't want a one-size-fits-all CI system. We want pipeline execution that integrates with our webhook architecture, our NATS-based log streaming, our Istio ingress. Tekton + Operator gives us that integration surface while handling the undifferentiated heavy lifting of component lifecycle management.

## What's Next

This document covered the deployment landscape. For deeper guides:

- **[Tekton Configuration Guide](./tekton-configuration.md)** - Comprehensive ConfigMap reference and TektonConfig field mapping
- **[Dashboard Security Guide](./dashboard-security.md)** - OAuth2 Proxy setup, RBAC, and authentication patterns
- **[Log Collection with Vector](./vector-integration.md)** - Complete Vector configuration for Tekton + NATS streaming
- **[Cloud Events Architecture](./cloud-events.md)** - Event payload structure, webhook receiver patterns, retry behavior

These guides are placeholders—to be written as needed based on user questions and production learnings.

---

**This document is grounded in**: Official Tekton documentation, Red Hat OpenShift Pipelines guides, production deployments at Planton (5 months running), and research into Vector/Fluent Bit integration patterns.

**Status**: Living document, updated as Tekton and the KubernetesTektonOperator component evolve.

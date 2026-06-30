# KubernetesTekton Research Documentation

This document provides comprehensive research on deploying Tekton using direct manifests, covering architecture decisions, comparison with operator-based deployment, and integration patterns.

## Deployment Approaches Comparison

### Direct Manifest Deployment (This Component)

**How it works:**
```bash
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl apply --filename https://infra.tekton.dev/tekton-releases/dashboard/latest/release.yaml
```

**Pros:**
- Simple and transparent - you see exactly what's being deployed
- Easy to debug - direct manifest inspection
- No operator overhead
- Faster initial deployment
- No TektonConfig CRD abstraction layer

**Cons:**
- Manual upgrades required
- Configuration via ConfigMap patches
- No automated lifecycle management

### Operator-Based Deployment (KubernetesTektonOperator)

**How it works:**
1. Deploy Tekton Operator
2. Create TektonConfig CRD
3. Operator reconciles and installs components

**Pros:**
- Automated upgrades
- Unified configuration via TektonConfig
- Health monitoring and self-healing
- Component selection via profiles (all/basic/lite)

**Cons:**
- More moving parts
- Abstraction can hide issues
- Operator itself needs maintenance
- Slower initial deployment (CRD registration, reconciliation)

## When to Use Each Approach

| Use Case | Recommended |
|----------|-------------|
| Learning/Experimentation | KubernetesTekton (manifest) |
| Development environments | KubernetesTekton (manifest) |
| Production with GitOps | Either (both work with ArgoCD/Flux) |
| Need automated upgrades | KubernetesTektonOperator |
| Debugging CI/CD issues | KubernetesTekton (manifest) |
| Enterprise with many clusters | KubernetesTektonOperator |

## CloudEvents Integration

### How CloudEvents Work in Tekton

Tekton's CloudEvents feature sends HTTP POST requests to a configured sink URL when TaskRun or PipelineRun resources change state.

**Event Types:**
- `dev.tekton.event.taskrun.started.v1`
- `dev.tekton.event.taskrun.running.v1`
- `dev.tekton.event.taskrun.successful.v1`
- `dev.tekton.event.taskrun.failed.v1`
- `dev.tekton.event.pipelinerun.started.v1`
- `dev.tekton.event.pipelinerun.running.v1`
- `dev.tekton.event.pipelinerun.successful.v1`
- `dev.tekton.event.pipelinerun.failed.v1`

### Configuration

The sink URL is configured in the `config-defaults` ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-defaults
  namespace: tekton-pipelines
data:
  default-cloud-events-sink: "http://my-service.namespace.svc.cluster.local/events"
```

### Receiver Implementation

A CloudEvents receiver needs to:
1. Accept HTTP POST on the configured path
2. Parse CloudEvent headers (`ce-type`, `ce-source`, `ce-id`, etc.)
3. Process the JSON body containing run details
4. Return 2xx status code

Example receiver endpoint in Go:

```go
func handleCloudEvent(w http.ResponseWriter, r *http.Request) {
    eventType := r.Header.Get("ce-type")
    eventSource := r.Header.Get("ce-source")
    
    var payload map[string]interface{}
    json.NewDecoder(r.Body).Decode(&payload)
    
    // Process event based on type
    switch eventType {
    case "dev.tekton.event.pipelinerun.successful.v1":
        // Update status, trigger deployment, send notification
    case "dev.tekton.event.pipelinerun.failed.v1":
        // Alert, update dashboard
    }
    
    w.WriteHeader(http.StatusOK)
}
```

## Dashboard Architecture

### Components

The Tekton Dashboard consists of:
- **Frontend**: React SPA served at port 9097
- **Backend**: Go API server proxying Kubernetes API

### Accessing the Dashboard

**Option 1: Port Forward (Development)**
```bash
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097
```

**Option 2: LoadBalancer Service**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: tekton-dashboard-lb
  namespace: tekton-pipelines
spec:
  type: LoadBalancer
  selector:
    app: tekton-dashboard
  ports:
    - port: 80
      targetPort: 9097
```

**Option 3: Gateway API (This Component)**
Uses Certificate → Gateway → HTTPRoute pattern for TLS-terminated HTTPS access.

### Authentication

Tekton Dashboard has **no built-in authentication**. Production deployments must add auth:

1. **OAuth2 Proxy**: Deploy oauth2-proxy in front of dashboard
2. **Istio AuthorizationPolicy**: If using Istio service mesh
3. **Ingress Auth Annotations**: nginx.ingress auth annotations

## Manifest URLs

### Tekton Pipelines
- Latest: `https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml`
- Versioned: `https://storage.googleapis.com/tekton-releases/pipeline/v0.65.2/release.yaml`
- Previous: `https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.64.0/release.yaml`

### Tekton Dashboard
- Latest: `https://infra.tekton.dev/tekton-releases/dashboard/latest/release.yaml`  
- Versioned: `https://infra.tekton.dev/tekton-releases/dashboard/v0.53.0/release.yaml`

### Tekton Triggers (Not in this component)
- Latest: `https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml`

## Upgrade Path

### From Manifest Deployment

```bash
# 1. Check current version
kubectl get deployment tekton-pipelines-controller -n tekton-pipelines -o jsonpath='{.spec.template.spec.containers[0].image}'

# 2. Apply new version
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/v0.66.0/release.yaml

# 3. Verify
kubectl rollout status deployment/tekton-pipelines-controller -n tekton-pipelines
```

### Migrating to Operator

If you later want operator-managed lifecycle:

1. Delete manifest-deployed resources (but keep PipelineRuns/TaskRuns)
2. Deploy KubernetesTektonOperator
3. Operator will adopt existing CRDs

## Troubleshooting

### Common Issues

**Pipeline controller not starting:**
```bash
kubectl logs -n tekton-pipelines -l app=tekton-pipelines-controller
kubectl describe pod -n tekton-pipelines -l app=tekton-pipelines-controller
```

**CloudEvents not being sent:**
```bash
# Verify config
kubectl get cm config-defaults -n tekton-pipelines -o yaml | grep cloud-events

# Check controller logs for CloudEvent errors
kubectl logs -n tekton-pipelines -l app=tekton-pipelines-controller | grep -i cloudevent
```

**Dashboard not accessible:**
```bash
# Check dashboard pod
kubectl get pods -n tekton-pipelines -l app=tekton-dashboard

# Check service
kubectl get svc -n tekton-pipelines tekton-dashboard

# Test connectivity
kubectl run curl --rm -it --restart=Never --image=curlimages/curl -- \
  curl -v http://tekton-dashboard.tekton-pipelines:9097
```

## References

- [Tekton Pipelines Releases](https://github.com/tektoncd/pipeline/releases)
- [Tekton Dashboard Releases](https://github.com/tektoncd/dashboard/releases)
- [Tekton CloudEvents Documentation](https://tekton.dev/docs/pipelines/events/)
- [Tekton ConfigMap Configuration](https://tekton.dev/docs/pipelines/additional-configs/)

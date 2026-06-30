# KubernetesTektonOperator: Dashboard Ingress and CloudEvents Support

**Date**: December 20, 2025
**Type**: Enhancement
**Components**: Kubernetes Provider, API Definitions, Pulumi CLI Integration, Gateway API

## Summary

Extended the `KubernetesTektonOperator` component to support dashboard ingress via Kubernetes Gateway API and CloudEvents sink URL configuration. These features bring feature parity with the manifest-based `KubernetesTekton` component, enabling production-ready Tekton deployments with external dashboard access and pipeline event notifications.

## Problem Statement / Motivation

The `KubernetesTektonOperator` component provided a simplified way to deploy Tekton using the operator pattern, but lacked two critical features available in the sibling `KubernetesTekton` component:

### Pain Points

- **No External Dashboard Access**: Users had to manually set up ingress or use port-forwarding to access the Tekton Dashboard
- **No Pipeline Notifications**: CloudEvents integration for pipeline lifecycle events wasn't available, limiting observability and automation capabilities
- **Feature Gap**: Inconsistent capabilities between the two Tekton deployment methods

## Solution / What's New

Added two new configuration options to `KubernetesTektonOperatorSpec`:

### Dashboard Ingress

Exposes the Tekton Dashboard externally via Kubernetes Gateway API with:
- TLS certificate provisioning via cert-manager
- HTTPS listener with automatic HTTP→HTTPS redirect
- HTTPRoute configuration for traffic routing

### CloudEvents Sink URL

Configures Tekton Pipelines to emit CloudEvents for:
- TaskRun state changes (started, running, succeeded, failed)
- PipelineRun lifecycle events

## Implementation Details

### Proto Schema Changes

Added new fields to `spec.proto`:

```protobuf
message KubernetesTektonOperatorSpec {
  // ... existing fields ...
  
  // Dashboard ingress configuration
  KubernetesTektonOperatorDashboardIngress dashboard_ingress = 5;
  
  // CloudEvents sink URL for pipeline notifications
  string cloud_events_sink_url = 6;
}

message KubernetesTektonOperatorDashboardIngress {
  bool enabled = 1;
  string hostname = 2;
}
```

### TektonConfig Enhancement

Modified `buildTektonConfigYAML()` to include the cloud events sink when configured:

```yaml
spec:
  profile: all
  targetNamespace: tekton-pipelines
  pipeline:
    default-cloud-events-sink: "http://my-receiver.svc.cluster.local"
```

### Gateway API Resources

Created `ingress.go` implementing the standard Planton ingress pattern:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Certificate   │───▶│     Gateway     │───▶│   HTTPRoutes    │
│  (cert-manager) │    │  (HTTPS/HTTP)   │    │ (redirect/route)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │ tekton-dashboard│
                                              │   :9097         │
                                              └─────────────────┘
```

### Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Added `dashboard_ingress` and `cloud_events_sink_url` fields |
| `locals.go` | Added ingress and CloudEvents computed values, exports |
| `outputs.go` | Added `OpCloudEventsSinkURL`, `OpDashboardExternalHostname` |
| `vars.go` | Added Gateway/Istio constants, Dashboard service details |
| `tekton_operator.go` | Updated TektonConfig builder for cloud events |
| `ingress.go` | **New** - Gateway API ingress resources |
| `main.go` | Updated orchestration to include ingress creation |

## Usage Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  operatorVersion: v0.78.0
  components:
    pipelines: true
    triggers: true
    dashboard: true
  dashboardIngress:
    enabled: true
    hostname: "tekton-dashboard.example.com"
  cloudEventsSinkUrl: "http://events-receiver.monitoring.svc.cluster.local/tekton"
```

## Benefits

### For Platform Teams
- **External Dashboard Access**: Expose Tekton Dashboard with TLS without manual ingress configuration
- **Pipeline Observability**: Enable event-driven monitoring and alerting via CloudEvents
- **Consistent Experience**: Same ingress pattern as other Planton components

### For Developers
- **Simple Configuration**: Two fields to enable powerful features
- **Secure by Default**: Automatic TLS via cert-manager integration
- **Event Integration**: Connect pipelines to notification systems, audit logs, or automation

## Impact

### User Experience
- Dashboard accessible via custom hostname with HTTPS
- Pipeline events routed to external systems for monitoring/automation

### Compatibility
- Fully backward compatible - new fields are optional
- Existing deployments continue to work unchanged
- Ingress only created when `dashboard_ingress.enabled: true`

### Prerequisites
- Istio ingress gateway installed
- cert-manager with ClusterIssuer matching the domain
- Gateway API CRDs installed

## Related Work

- **yaml/v2 Fix** (earlier in session): Fixed CRD timing issues using Pulumi's yaml/v2
- **KubernetesTekton Component**: Reference implementation with same features
- **Gateway API Pattern**: Follows established Planton ingress conventions

---

**Status**: ✅ Production Ready
**Timeline**: ~30 minutes implementation


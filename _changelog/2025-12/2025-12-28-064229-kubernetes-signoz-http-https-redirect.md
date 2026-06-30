# KubernetesSignoz: Add HTTP to HTTPS Redirect Support

**Date**: December 28, 2025
**Type**: Enhancement
**Components**: Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added HTTP to HTTPS redirect support for the KubernetesSignoz deployment component. Both the Pulumi and Terraform IaC modules now create Gateway API resources with HTTP listeners (port 80) and redirect HTTPRoutes that perform 301 redirects to HTTPS, matching the pattern used across other Kubernetes deployment components.

## Problem Statement / Motivation

When deploying SigNoz using the KubernetesSignoz module, external ingress only supported HTTPS access. Users accessing the HTTP URL were not redirected to HTTPS, leading to a poor user experience and inconsistency with other deployment components.

### Pain Points

- HTTP requests to SigNoz UI returned errors instead of redirecting to HTTPS
- OTEL Collector HTTP ingestion endpoint lacked HTTP redirect
- Inconsistent behavior compared to other Kubernetes deployment components (Jenkins, Tekton, OpenFGA, Solr, Locust)
- Terraform module was missing Gateway API resources entirely

## Solution / What's New

Updated both Pulumi and Terraform IaC modules to include:

1. **HTTP Listener (port 80)** on Gateway resources
2. **HTTP-to-HTTPS Redirect HTTPRoute** using Gateway API `RequestRedirect` filter with 301 status code
3. **Terraform Gateway API Resources** (previously missing from the module)

### Gateway Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Gateway                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Listeners:                                               ││
│  │   - https-external (port 443, TLS Terminate)             ││
│  │   - http-external (port 80, HTTP)                        ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                               │
          ┌────────────────────┴────────────────────┐
          │                                         │
          ▼                                         ▼
┌─────────────────────────┐           ┌─────────────────────────┐
│   HTTP Redirect Route   │           │      HTTPS Route        │
│  (http-external → 301)  │           │   (→ backend service)   │
│                         │           │                         │
│  Filter: RequestRedirect│           │  BackendRef: frontend   │
│    scheme: https        │           │    port: 3301           │
│    statusCode: 301      │           │                         │
└─────────────────────────┘           └─────────────────────────┘
```

## Implementation Details

### Pulumi Module Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/pulumi/module/locals.go`
- Added `SignozHttpRedirectRouteName` and `OtelHttpRedirectRouteName` fields to Locals struct
- Initialized computed route names: `{name}-signoz-http-redirect` and `{name}-otel-http-redirect`

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/pulumi/module/ingress_signoz.go`
- Added HTTP listener (`http-external`) to SigNoz UI Gateway
- Added HTTP redirect HTTPRoute with `RequestRedirect` filter
- Field order aligned with majority pattern: `RequestRedirect` before `Type`

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/pulumi/module/ingress_otel.go`
- Added HTTP listener (`http-otel-http`) to OTEL Collector Gateway
- Added HTTP redirect HTTPRoute with `RequestRedirect` filter
- Field order aligned with majority pattern

### Terraform Module Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/tf/locals.tf`
- Added computed resource names for Certificates, Gateways, and HTTPRoutes
- Added ClusterIssuer name extraction from hostname
- Added Istio ingress constants

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/tf/ingress.tf` (new)
- Created complete Gateway API resources for SigNoz UI ingress
- Created complete Gateway API resources for OTEL Collector ingress
- Each includes: Certificate, Gateway (HTTP+HTTPS), HTTP redirect route, HTTPS route

### Consistency with Existing Modules

The implementation follows the established pattern from:
- `kubernetesopenfga` - OpenFGA ingress
- `kubernetestekton` - Tekton Dashboard ingress
- `kubernetessolr` - Solr ingress
- `kuberneteslocust` - Locust ingress
- `kubernetesjenkins` - Jenkins ingress

Key consistency points:
- Gateway listener names: `https-external`, `http-external`
- HTTPRoute filter field order: `RequestRedirect` before `Type`
- 301 status code for permanent redirect
- Gateway placed in `istio-ingress` namespace
- HTTPRoutes placed in application namespace

## Benefits

- **Improved UX**: Users accessing HTTP URLs are automatically redirected to HTTPS
- **Security**: Ensures all traffic is encrypted by forcing HTTPS
- **Consistency**: Matches behavior of all other Kubernetes deployment components
- **Feature Parity**: Both Pulumi and Terraform modules now have complete ingress support

## Impact

### For SigNoz Users
- HTTP access to SigNoz UI now redirects to HTTPS
- HTTP access to OTEL Collector ingestion endpoint redirects to HTTPS
- No action required - redirect is automatic on next deployment update

### For Developers
- Terraform module now has complete Gateway API support (was missing)
- Pattern is consistent with other Kubernetes deployment components

## Files Changed

```
apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/pulumi/module/
  ├── locals.go           (modified - added redirect route names)
  ├── ingress_signoz.go   (modified - added HTTP listener and redirect)
  └── ingress_otel.go     (modified - added HTTP listener and redirect)

apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/tf/
  ├── locals.tf           (modified - added ingress resource names)
  └── ingress.tf          (created - complete Gateway API resources)
```

## Related Work

This enhancement brings KubernetesSignoz in line with the HTTP redirect pattern established in:
- KubernetesJenkins (existing)
- KubernetesTekton (existing)
- KubernetesOpenFGA (existing)
- KubernetesSolr (existing)
- KubernetesLocust (existing)

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation


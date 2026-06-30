# Temporal Kubernetes HTTP Ingress via Gateway API

**Date**: November 2, 2025
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Kubernetes Provider, Resource Management

## Summary

Added HTTP ingress support for Temporal Kubernetes deployments using Kubernetes Gateway API with automatic HTTPS certificate provisioning via Istio. The frontend ingress configuration now supports separate hostnames for gRPC (via LoadBalancer) and HTTP (via Gateway API), enabling secure HTTP access with TLS termination while maintaining existing gRPC connectivity patterns.

## Problem Statement / Motivation

Prior to this change, the Temporal Kubernetes frontend was exposed exclusively through a LoadBalancer service that included both gRPC (port 7233) and HTTP (port 7243) ports. While this approach worked for basic connectivity, it had significant limitations:

### Pain Points

- **No TLS support for HTTP**: The HTTP port on the LoadBalancer had no certificate management, requiring users to manually configure TLS or accept insecure connections
- **Single hostname limitation**: Both gRPC and HTTP traffic shared the same DNS hostname pointing to the LoadBalancer IP, making it impossible to route them through different ingress mechanisms
- **Limited ingress capabilities**: LoadBalancer services lack the advanced routing, header manipulation, and policy features available through Gateway API
- **Inconsistent with UI pattern**: The Web UI already used Gateway API with proper certificate management, creating inconsistency in how different Temporal components were exposed
- **Certificate management complexity**: Users had to manually provision and rotate certificates for HTTP endpoints

## Solution / What's New

The solution separates frontend ingress into two independent, purpose-built mechanisms:

1. **gRPC via LoadBalancer** (existing pattern, refined):
   - Dedicated LoadBalancer service exposing only port 7233
   - Uses `grpc_hostname` for DNS configuration via external-dns
   - Direct pod access without ingress layer
   - Optimal for high-throughput gRPC workloads

2. **HTTP via Gateway API** (new):
   - Kubernetes Gateway API resources in the Istio ingress namespace
   - Uses `http_hostname` for separate DNS configuration
   - Automatic certificate provisioning via cert-manager
   - HTTPS with TLS termination at the gateway
   - HTTP-to-HTTPS redirect (301)
   - Routes traffic to frontend service port 7243

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Temporal Frontend Ingress                 │
└─────────────────────────────────────────────────────────────┘

gRPC Traffic:                      HTTP Traffic:
┌──────────────┐                   ┌──────────────┐
│ gRPC Client  │                   │ HTTP Client  │
└──────┬───────┘                   └──────┬───────┘
       │                                  │
       │ temporal-grpc.example.com        │ temporal-http.example.com
       │                                  │
       ▼                                  ▼
┌──────────────────┐              ┌──────────────────┐
│  LoadBalancer    │              │  Istio Gateway   │
│  External IP     │              │  (shared)        │
│  Port: 7233      │              │  Port: 443       │
└──────┬───────────┘              └──────┬───────────┘
       │                                  │
       │                                  │ TLS termination
       │                                  │ cert-manager
       │                                  │
       ▼                                  ▼
┌──────────────────────────────────────────────────┐
│         Temporal Frontend Service                │
│         - Port 7233 (gRPC)                       │
│         - Port 7243 (HTTP)                       │
└──────────────────────────────────────────────────┘
```

## Implementation Details

### 1. API Definition Changes

Updated `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.proto`:

**Before**:
```protobuf
message TemporalKubernetesIngress {
  TemporalKubernetesIngressEndpoint frontend = 1;
  TemporalKubernetesIngressEndpoint web_ui = 2;
}

message TemporalKubernetesIngressEndpoint {
  bool enabled = 1;
  string hostname = 2;  // Single hostname for all traffic
}
```

**After**:
```protobuf
message TemporalKubernetesIngress {
  TemporalKubernetesFrontendIngressEndpoint frontend = 1;
  TemporalKubernetesWebUiIngressEndpoint web_ui = 2;
}

// Frontend supports both gRPC and HTTP
message TemporalKubernetesFrontendIngressEndpoint {
  bool enabled = 1;
  string grpc_hostname = 2;  // Required: gRPC LoadBalancer hostname
  string http_hostname = 3;  // Optional: HTTP Gateway API hostname
  
  option (buf.validate.message).cel = {
    id: "spec.ingress.frontend.grpc_hostname.required"
    expression: "!this.enabled || size(this.grpc_hostname) > 0"
    message: "frontend.grpc_hostname is required when frontend ingress is enabled"
  };
}

// Web UI only needs HTTP hostname
message TemporalKubernetesWebUiIngressEndpoint {
  bool enabled = 1;
  string hostname = 2;  // Required: HTTP hostname for Gateway API
  
  option (buf.validate.message).cel = {
    id: "spec.ingress.web_ui.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "web_ui.hostname is required when web ui ingress is enabled"
  };
}
```

**Key Changes**:
- Created separate message types for frontend and web UI ingress (semantically correct)
- Frontend (`TemporalKubernetesFrontendIngressEndpoint`) has both `grpc_hostname` and `http_hostname`
- Web UI (`TemporalKubernetesWebUiIngressEndpoint`) only has `hostname` (HTTP-only, no gRPC)
- Context-specific validation messages clearly indicate which endpoint has the issue
- Made `grpc_hostname` required for frontend when enabled
- Made `http_hostname` optional for frontend - HTTP Gateway only provisions if provided
- Made `hostname` required for web UI when enabled

### 2. Pulumi Module Updates

#### Locals Structure (`locals.go`)

```go
type Locals struct {
    // ... existing fields ...
    IngressFrontendGrpcHostname string  // New: dedicated gRPC hostname
    IngressFrontendHttpHostname string  // New: dedicated HTTP hostname
    IngressUIHostname           string  // Updated: now uses grpc_hostname
}
```

Initialization logic now populates both hostnames independently:

```go
if target.Spec.Ingress.Frontend.Enabled {
    if target.Spec.Ingress.Frontend.GrpcHostname != "" {
        locals.IngressFrontendGrpcHostname = target.Spec.Ingress.Frontend.GrpcHostname
        ctx.Export(OpExternalFrontendHostname, pulumi.String(locals.IngressFrontendGrpcHostname))
    }
    if target.Spec.Ingress.Frontend.HttpHostname != "" {
        locals.IngressFrontendHttpHostname = target.Spec.Ingress.Frontend.HttpHostname
    }
}
```

#### gRPC LoadBalancer (`frontend_ingress.go`)

Refined to only expose gRPC port:

```go
Ports: kubernetescorev1.ServicePortArray{
    &kubernetescorev1.ServicePortArgs{
        Name:       pulumi.String("grpc-frontend"),
        Port:       pulumi.Int(vars.FrontendGrpcPort),  // 7233
        Protocol:   pulumi.String("TCP"),
        TargetPort: pulumi.Int(vars.FrontendGrpcPort),
    },
    // HTTP port removed - now handled by Gateway API
},
```

**Changes**:
- Removed HTTP port (7243) from LoadBalancer
- Updated to use `IngressFrontendGrpcHostname` for external-dns annotation
- Updated comments to clarify this handles gRPC only

#### HTTP Gateway API Ingress (`frontend_http_ingress.go`)

New file implementing Gateway API pattern (similar to `web_ui_ingress.go`):

**Certificate Provisioning**:
```go
certSecret := fmt.Sprintf("%s-frontend-http-cert", locals.Namespace)

// Extract domain for ClusterIssuer (e.g., "planton.live" from "temporal.planton.live")
hostnameParts := strings.Split(httpHostname, ".")
clusterIssuerName := strings.Join(hostnameParts[1:], ".")

addedCertificate, err := certmanagerv1.NewCertificate(ctx,
    "frontend-http-cert",
    &certmanagerv1.CertificateArgs{
        Metadata: metav1.ObjectMetaArgs{
            Name:      pulumi.String(certSecret),
            Namespace: pulumi.String(vars.IstioIngressNamespace),  // istio-ingress
            Labels:    pulumi.ToStringMap(locals.Labels),
        },
        Spec: certmanagerv1.CertificateSpecArgs{
            DnsNames:   pulumi.ToStringArray([]string{httpHostname}),
            SecretName: pulumi.String(certSecret),
            IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
                Kind: pulumi.String("ClusterIssuer"),
                Name: pulumi.String(clusterIssuerName),
            },
        },
    }, pulumi.Provider(kubernetesProvider))
```

**Gateway Configuration**:
```go
gwName := pulumi.Sprintf("%s-frontend-http-external", locals.Namespace)

createdGateway, err := gatewayv1.NewGateway(ctx,
    "external-frontend-http",
    &gatewayv1.GatewayArgs{
        Spec: gatewayv1.GatewaySpecArgs{
            GatewayClassName: pulumi.String(vars.GatewayIngressClassName),  // "istio"
            Addresses: gatewayv1.GatewaySpecAddressesArray{
                gatewayv1.GatewaySpecAddressesArgs{
                    Type:  pulumi.String("Hostname"),
                    Value: pulumi.String(vars.GatewayExternalLoadBalancerServiceHostname),
                },
            },
            Listeners: gatewayv1.GatewaySpecListenersArray{
                // HTTPS listener with TLS termination
                &gatewayv1.GatewaySpecListenersArgs{
                    Name:     pulumi.String("https-external"),
                    Hostname: pulumi.String(httpHostname),
                    Port:     pulumi.Int(443),
                    Protocol: pulumi.String("HTTPS"),
                    Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
                        Mode: pulumi.String("Terminate"),
                        CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
                            gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
                                Name: pulumi.String(certSecret),
                            },
                        },
                    },
                },
                // HTTP listener for redirect
                &gatewayv1.GatewaySpecListenersArgs{
                    Name:     pulumi.String("http-external"),
                    Hostname: pulumi.String(httpHostname),
                    Port:     pulumi.Int(80),
                    Protocol: pulumi.String("HTTP"),
                },
            },
        },
    })
```

**HTTP→HTTPS Redirect**:
```go
_, err = gatewayv1.NewHTTPRoute(ctx,
    "http-frontend-http-external-redirect",
    &gatewayv1.HTTPRouteArgs{
        Spec: gatewayv1.HTTPRouteSpecArgs{
            Hostnames: pulumi.StringArray{pulumi.String(httpHostname)},
            ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
                gatewayv1.HTTPRouteSpecParentRefsArgs{
                    SectionName: pulumi.String("http-external"),
                },
            },
            Rules: gatewayv1.HTTPRouteSpecRulesArray{
                gatewayv1.HTTPRouteSpecRulesArgs{
                    Filters: gatewayv1.HTTPRouteSpecRulesFiltersArray{
                        gatewayv1.HTTPRouteSpecRulesFiltersArgs{
                            Type: pulumi.String("RequestRedirect"),
                            RequestRedirect: gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectArgs{
                                Scheme:     pulumi.String("https"),
                                StatusCode: pulumi.Int(301),
                            },
                        },
                    },
                },
            },
        },
    })
```

**HTTPS Route to Backend**:
```go
_, err = gatewayv1.NewHTTPRoute(ctx,
    "https-frontend-http-external",
    &gatewayv1.HTTPRouteArgs{
        Spec: gatewayv1.HTTPRouteSpecArgs{
            ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
                gatewayv1.HTTPRouteSpecParentRefsArgs{
                    SectionName: pulumi.String("https-external"),
                },
            },
            Rules: gatewayv1.HTTPRouteSpecRulesArray{
                gatewayv1.HTTPRouteSpecRulesArgs{
                    BackendRefs: gatewayv1.HTTPRouteSpecRulesBackendRefsArray{
                        gatewayv1.HTTPRouteSpecRulesBackendRefsArgs{
                            Name:      pulumi.String(locals.FrontendServiceName),
                            Namespace: createdNamespace.Metadata.Name(),
                            Port:      pulumi.Int(vars.FrontendHttpPort),  // 7243
                        },
                    },
                },
            },
        },
    })
```

#### Resource Orchestration (`main.go`)

Updated to provision HTTP ingress after gRPC ingress:

```go
func Resources(ctx *pulumi.Context, stackInput *temporalkubernetesv1.TemporalKubernetesStackInput) error {
    // ... namespace, secrets, helm chart ...

    // gRPC LoadBalancer (existing, refined)
    if err := frontendIngress(ctx, locals, createdNamespace); err != nil {
        return errors.Wrap(err, "failed to create frontend gRPC ingress")
    }

    // HTTP Gateway API (new)
    if err := frontendHttpIngress(ctx, locals, kubernetesProvider, createdNamespace); err != nil {
        return errors.Wrap(err, "failed to create frontend HTTP ingress")
    }

    // Web UI ingress (existing)
    if err := webUiIngress(ctx, locals, kubernetesProvider, createdNamespace); err != nil {
        return errors.Wrap(err, "failed to create web UI ingress")
    }

    return nil
}
```

### 3. Files Modified

**API Layer**:
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.proto` - updated field definitions
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.pb.go` - regenerated from proto
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/api_test.go` - updated test fixtures

**Pulumi Module**:
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/locals.go` - added HTTP hostname tracking
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/frontend_ingress.go` - removed HTTP port, uses GrpcHostname
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/frontend_http_ingress.go` - **new file** for HTTP Gateway API ingress
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/web_ui_ingress.go` - updated to use GrpcHostname field
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/main.go` - wired up HTTP ingress
- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/BUILD.bazel` - auto-updated by Gazelle

**Total**: 8 files modified, 1 file created, ~220 lines of new code

## Usage Examples

### Manifest Configuration

**gRPC-only (existing behavior)**:
```yaml
apiVersion: code2ai.planton.ai/v1
kind: TemporalKubernetes
metadata:
  name: my-temporal
spec:
  ingress:
    frontend:
      enabled: true
      grpc_hostname: temporal-grpc.planton.live
```

This creates only the gRPC LoadBalancer, accessible at `temporal-grpc.planton.live:7233`.

**gRPC + HTTP with separate hostnames**:
```yaml
apiVersion: code2ai.planton.ai/v1
kind: TemporalKubernetes
metadata:
  name: my-temporal
spec:
  ingress:
    frontend:
      enabled: true
      grpc_hostname: temporal-grpc.planton.live
      http_hostname: temporal-http.planton.live
```

This creates:
- gRPC LoadBalancer at `temporal-grpc.planton.live:7233`
- HTTP Gateway with automatic HTTPS at `https://temporal-http.planton.live`

### Deployment

```bash
# Preview changes
planton pulumi preview --manifest temporal.yaml --module-dir ${MODULE}

# Apply configuration
planton pulumi up --manifest temporal.yaml

# Verify gRPC access
temporal --address temporal-grpc.planton.live:7233 workflow list

# Verify HTTP access (automatically redirects to HTTPS)
curl -L http://temporal-http.planton.live/api/v1/namespaces
# → 301 redirect to https://temporal-http.planton.live/api/v1/namespaces

curl https://temporal-http.planton.live/api/v1/namespaces
# → Returns namespace list with valid TLS certificate
```

### Kubernetes Resources Created

For HTTP ingress, the following resources are provisioned:

```bash
# Certificate in istio-ingress namespace
kubectl get certificate -n istio-ingress temporal-app-prod-main-frontend-http-cert

# Gateway in istio-ingress namespace
kubectl get gateway -n istio-ingress temporal-app-prod-main-frontend-http-external

# HTTPRoutes in application namespace
kubectl get httproute -n temporal-app-prod-main
# - http-frontend-http-external-redirect (HTTP→HTTPS)
# - https-frontend-http-external (HTTPS to backend)
```

## Benefits

### Security

- **Automatic TLS certificates**: cert-manager provisions and rotates certificates using ClusterIssuer
- **Enforced HTTPS**: HTTP traffic automatically redirects to HTTPS with 301
- **TLS termination at gateway**: Certificates managed centrally in istio-ingress namespace
- **No manual certificate management**: Users don't handle certificate files or secrets

### Operational

- **Consistent ingress patterns**: HTTP ingress follows same Gateway API pattern as Web UI
- **Separate DNS management**: gRPC and HTTP can use different hostnames and DNS zones
- **Infrastructure reuse**: Leverages existing Istio ingress infrastructure
- **Independent scaling**: gRPC LoadBalancer and HTTP Gateway scale independently

### Developer Experience

- **Backward compatible**: Existing deployments with only `grpc_hostname` continue to work
- **Opt-in HTTP ingress**: HTTP Gateway only provisions if `http_hostname` is specified
- **No migration required**: Users can add HTTP ingress without changing gRPC configuration
- **Standard tooling**: Uses standard Kubernetes Gateway API and cert-manager

### Cost Efficiency

- **Shared infrastructure**: HTTP traffic uses shared Istio LoadBalancer (no additional LB cost)
- **Reduced LoadBalancer count**: Avoids provisioning separate LoadBalancer for HTTP
- **No external certificate services**: Uses cluster-native cert-manager

## Impact

### Users

**Before**: 
- Users had to manually configure TLS for HTTP endpoints or accept insecure connections
- Both gRPC and HTTP shared same hostname/IP, limiting routing flexibility
- Certificate rotation required manual intervention

**After**:
- HTTP endpoints automatically get valid TLS certificates
- Users can specify separate hostnames for gRPC and HTTP
- Certificates auto-renew without intervention
- Standard HTTPS access (port 443) instead of non-standard ports

### Operators

- Temporal deployments align with other Gateway API-based workloads
- Certificate management centralized through cert-manager ClusterIssuers
- Easier to apply network policies and ingress rules at gateway level
- Metrics and observability through Istio for HTTP traffic

### Developers

- New `http_hostname` field in API spec
- Separate ingress provisioning logic in Pulumi module
- Follows established Gateway API patterns from Web UI implementation
- Clean separation of concerns (gRPC vs HTTP ingress)

## Design Decisions

### Why Separate Hostnames?

**Decision**: Require separate `grpc_hostname` and `http_hostname` instead of routing both from same hostname.

**Rationale**:
- DNS hostname can only resolve to one IP address
- gRPC LoadBalancer needs dedicated external IP for direct pod access
- HTTP traffic benefits from Gateway API features (routing, policies, observability)
- Avoids complex TLS passthrough for gRPC while terminating HTTP
- Users often want different DNS names for different protocols (e.g., `grpc.temporal.company.com` vs `api.temporal.company.com`)

**Trade-off**: Requires two DNS entries, but provides maximum flexibility and simplicity.

### Why Gateway API Instead of LoadBalancer?

**Decision**: Use Gateway API for HTTP instead of adding TLS to LoadBalancer.

**Rationale**:
- Gateway API provides automatic certificate management integration
- Enables HTTP→HTTPS redirect, header manipulation, and routing rules
- Reuses existing Istio infrastructure (no additional LoadBalancer cost)
- Aligns with Web UI ingress pattern (consistency)
- Better observability and metrics through Istio
- Easier to apply network policies and security controls

**Trade-off**: Adds Gateway API dependency, but this is already required for Web UI.

### Why Optional HTTP Hostname?

**Decision**: Make `http_hostname` optional; only provision HTTP ingress if specified.

**Rationale**:
- Backward compatibility: existing deployments only use gRPC
- Many users don't need HTTP API access
- Reduces resource consumption when not needed
- Allows gradual adoption without breaking changes

## Migration Guide

### For Existing Deployments

Existing deployments using `hostname` field will need to update their manifests:

**Old manifest**:
```yaml
spec:
  ingress:
    frontend:
      enabled: true
      hostname: temporal-frontend.planton.live  # Old field (no longer exists)
    web_ui:
      enabled: true
      hostname: temporal-ui.planton.live  # Old field (no longer exists)
```

**After protobuf update, update to**:

```yaml
spec:
  ingress:
    frontend:
      enabled: true
      grpc_hostname: temporal-frontend.planton.live  # Required for frontend
    web_ui:
      enabled: true
      hostname: temporal-ui.planton.live  # Still called hostname for web UI
```

To add frontend HTTP ingress:

```yaml
spec:
  ingress:
    frontend:
      enabled: true
      grpc_hostname: temporal-frontend-grpc.planton.live
      http_hostname: temporal-frontend-http.planton.live  # Optional: enables HTTP ingress
    web_ui:
      enabled: true
      hostname: temporal-ui.planton.live
```

**DNS Updates Required**:
1. Update DNS record for frontend `grpc_hostname` to point to LoadBalancer external IP
2. Create DNS record for frontend `http_hostname` pointing to Istio ingress LoadBalancer (if using HTTP)
3. Web UI `hostname` continues to point to Istio ingress LoadBalancer (no change)

### For New Deployments

New deployments should specify both hostnames explicitly:

```yaml
spec:
  ingress:
    frontend:
      enabled: true
      grpc_hostname: temporal-grpc.example.com
      http_hostname: temporal-api.example.com  # Optional but recommended
```

## Testing Strategy

### Manual Verification

1. **Deploy Temporal with HTTP ingress enabled**:
```bash
planton pulumi up --manifest temporal.yaml
```

2. **Verify gRPC connectivity**:
```bash
temporal --address temporal-grpc.planton.live:7233 workflow list
```

3. **Verify HTTP redirect**:
```bash
curl -v http://temporal-http.planton.live/
# Should return 301 to HTTPS
```

4. **Verify HTTPS with valid certificate**:
```bash
curl -v https://temporal-http.planton.live/api/v1/namespaces
# Should show valid TLS certificate
# Should return Temporal API response
```

5. **Verify certificate details**:
```bash
kubectl get certificate -n istio-ingress temporal-app-prod-main-frontend-http-cert
# Status should show Ready: True
```

6. **Verify Gateway and HTTPRoutes**:
```bash
kubectl get gateway -n istio-ingress
kubectl get httproute -n temporal-app-prod-main
```

## Known Limitations

- **HTTP-only Temporal features**: Some Temporal CLI operations require gRPC and won't work over HTTP
- **ClusterIssuer dependency**: Requires ClusterIssuer matching the domain suffix to be configured
- **Istio dependency**: HTTP ingress requires Istio to be deployed and configured
- **DNS propagation**: DNS changes may take time to propagate after deployment

## Future Enhancements

- **TLS passthrough for gRPC**: Support gRPC over TLS through Gateway API (when Istio supports it)
- **Single hostname mode**: Optional mode to route both gRPC and HTTP from same hostname using protocol detection
- **Custom certificate support**: Allow users to provide their own certificates instead of cert-manager
- **Path-based routing**: Support multiple Temporal namespaces with path-based routing rules
- **Rate limiting**: Apply rate limiting policies at Gateway level for HTTP traffic

## Related Work

- **Web UI Ingress**: This implementation follows the same Gateway API pattern established for Temporal Web UI ingress in `web_ui_ingress.go`
- **Kubernetes Gateway API**: Leverages the Gateway API v1 specification for ingress configuration
- **cert-manager Integration**: Uses cert-manager for certificate lifecycle management, consistent with other workloads

## Backward Compatibility

This is a **breaking change** at the API level:

- **Replaced shared message with separate types**: `TemporalKubernetesIngressEndpoint` split into `TemporalKubernetesFrontendIngressEndpoint` and `TemporalKubernetesWebUiIngressEndpoint`
- **Frontend changes**: The `hostname` field replaced with `grpc_hostname` (required) and `http_hostname` (optional)
- **Web UI unchanged**: Still uses `hostname` field (now in dedicated message type)
- **Field number reuse**: Frontend field 2 reused for `grpc_hostname` (was `hostname`), field 3 allocated for `http_hostname`

**Code Updates Required**:
- **API tests**: Test fixtures updated to use new message types in `api_test.go`
- **Web UI ingress**: Field reference remains `Hostname` but now from `TemporalKubernetesWebUiIngressEndpoint` type
- **Frontend ingress**: Updated to use `GrpcHostname` and `HttpHostname` from `TemporalKubernetesFrontendIngressEndpoint` type
- **Validation**: Context-specific error messages distinguish between frontend and web UI issues

**Mitigation**: The change is clear and explicit - users will get validation errors with old manifests, guiding them to update field names. No silent failures or data loss. Error messages now clearly indicate whether the issue is with `frontend.grpc_hostname` or `web_ui.hostname`.

---

**Status**: ✅ Production Ready
**Timeline**: Implemented November 2, 2025

## Code Metrics

- **Files Modified**: 8
- **Files Created**: 1
- **Lines Added**: ~230
- **Lines Removed**: ~20
- **Net Change**: +210 lines
- **Components**: API, Pulumi Module, Gateway API, cert-manager, Tests
- **Resources Created per HTTP Ingress**: 1 Certificate, 1 Gateway, 2 HTTPRoutes
- **Breaking Changes**: 1 (field rename: `hostname` → `grpc_hostname` + `http_hostname`)


# ExternalDNS Cloudflare Provider Support

**Date**: October 17, 2025  
**Type**: Feature  
**Component**: ExternalDnsKubernetes

## Summary

Added comprehensive Cloudflare provider support to the ExternalDNS Kubernetes addon, enabling automated DNS record management for Kubernetes Services and Ingresses using Cloudflare as the authoritative DNS provider. This integration supports Cloudflare's proxy feature (orange cloud), API rate limiting optimization, and secure API token authentication.

## Motivation

Organizations using Cloudflare for DNS management needed seamless integration with Kubernetes to:
- Automatically create and update DNS records when deploying applications
- Eliminate manual DNS configuration for each service or ingress
- Leverage Cloudflare's edge network features (DDoS protection, WAF, CDN)
- Manage multiple domains with isolated ExternalDNS instances
- Ensure DNS records stay synchronized with cluster state

Previously, ExternalDNS supported only cloud provider DNS services (Google Cloud DNS, AWS Route53, Azure DNS). This limited users on other Kubernetes distributions (bare metal, edge, or multi-cloud) who used Cloudflare for DNS.

## What's New

### 1. Cloudflare Provider Configuration

Added `ExternalDnsCloudflareConfig` to the ExternalDNS specification:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns-planton
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  namespace: external-dns
  externalDnsVersion: v0.19.0
  helmChartVersion: 1.19.0
  cloudflare:
    apiToken: "your-cloudflare-api-token"
    dnsZoneId: "7adff2f8326758cac24fd17f02ca3001"
    isProxied: true
```

**Configuration Fields**:
- **`api_token`** (required): Cloudflare API token with Zone:Zone:Read and Zone:DNS:Edit permissions
- **`dns_zone_id`** (required): Cloudflare DNS zone ID to manage
- **`is_proxied`** (optional): Enable Cloudflare proxy (orange cloud) for traffic routing through Cloudflare's edge network

### 2. Enhanced Spec Configuration

Added user-configurable fields to ExternalDnsKubernetesSpec:

```protobuf
message ExternalDnsKubernetesSpec {
  KubernetesAddonTargetCluster target_cluster = 1;
  string namespace = 2 [(default) = "external-dns"];
  string external_dns_version = 3 [(default) = "v0.19.0"];
  string helm_chart_version = 4 [(default) = "1.19.0"];
  
  oneof provider_config {
    ExternalDnsGkeConfig gke = 200;
    ExternalDnsEksConfig eks = 201;
    ExternalDnsAksConfig aks = 202;
    ExternalDnsCloudflareConfig cloudflare = 203;  // NEW
  }
}
```

**Key Improvements**:
- **Namespace customization**: Deploy to custom namespaces if needed
- **Version pinning**: Control both ExternalDNS and Helm chart versions
- **Sensible defaults**: All fields have production-ready defaults

### 3. Multi-Domain Support

The implementation supports multiple ExternalDNS instances for different domains in the same namespace:

```yaml
# Instance 1: planton.cloud
metadata:
  name: external-dns-planton
spec:
  cloudflare:
    dnsZoneId: "7adff2f8326758cac24fd17f02ca3001"

---
# Instance 2: planton.live  
metadata:
  name: external-dns-planton-live
spec:
  cloudflare:
    dnsZoneId: "77c6a34cf87dd1e8b497dc895bf5ea1b"
```

**Resource Isolation**:
- Helm release name derived from `metadata.name` (e.g., `external-dns-planton`)
- ServiceAccount name matches release name for clear identification
- Secret names include domain suffix (e.g., `cloudflare-api-token-external-dns-planton`)
- No resource conflicts when managing multiple domains

## Implementation Details

### Protobuf Specification

**File**: `apis/project/planton/provider/kubernetes/addon/externaldns/v1/spec.proto`

```protobuf
message ExternalDnsCloudflareConfig {
  // Cloudflare API token for authentication
  string api_token = 1 [(buf.validate.field).required = true];
  
  // Cloudflare DNS zone ID to manage
  string dns_zone_id = 2 [(buf.validate.field).required = true];
  
  // Enable Cloudflare proxy (orange cloud) for DDoS protection, WAF, and CDN
  bool is_proxied = 3;
}
```

**Design Principles**:
- Follows the 80/20 rule: exposes essential fields users need
- Deployment-agnostic: not tied to Helm implementation details
- Sensible defaults: proxy disabled by default, rate limiting optimized automatically
- Security-focused: API token marked as required, stored in Kubernetes Secret

### Pulumi Module Implementation

**File**: `apis/project/planton/provider/kubernetes/addon/externaldns/v1/iac/pulumi/module/main.go`

**Key Features**:

1. **Secure Secret Management**:
   ```go
   secretName := fmt.Sprintf("cloudflare-api-token-%s", releaseName)
   secret, err := corev1.NewSecret(ctx, secretName,
       &corev1.SecretArgs{
           StringData: pulumi.StringMap{
               "apiKey": pulumi.String(cf.ApiToken),  // Key must be "apiKey" per Helm chart
           },
       })
   ```

2. **Environment Variable Configuration**:
   ```go
   values["env"] = pulumi.Array{
       pulumi.Map{
           "name": pulumi.String("CF_API_TOKEN"),
           "valueFrom": pulumi.Map{
               "secretKeyRef": pulumi.Map{
                   "name": secret.Metadata.Name(),
                   "key":  pulumi.String("apiKey"),
               },
           },
       },
   }
   ```

3. **Cloudflare-Specific Arguments**:
   ```go
   extraArgs := pulumi.StringArray{
       pulumi.String("--cloudflare-dns-records-per-page=5000"),  // Rate limit optimization
       pulumi.String(fmt.Sprintf("--zone-id-filter=%s", cf.DnsZoneId)),  // Zone scoping
   }
   
   if cf.IsProxied {
       extraArgs = append(extraArgs, pulumi.String("--cloudflare-proxied"))
   }
   ```

### Critical Implementation Decisions

#### 1. Sources Must Be Top-Level Helm Values for RBAC Generation

**Critical Discovery**: The ExternalDNS Helm chart only generates ClusterRole RBAC permissions when sources are specified as **top-level `sources` values**, not when passed via `extraArgs`.

**Problem Encountered**:
```go
// ❌ WRONG - Does not generate RBAC
extraArgs := pulumi.StringArray{
    pulumi.String("--source=service"),
    pulumi.String("--source=gateway-httproute"),
}
```

**Solution**:
```go
// ✅ CORRECT - Generates proper RBAC
values["sources"] = pulumi.StringArray{
    pulumi.String("service"),
    pulumi.String("ingress"),
    pulumi.String("gateway-httproute"),
}
```

**Symptoms when misconfigured**:
- ClusterRole missing `gateway.networking.k8s.io` permissions
- Pods crash with: `failed to sync *v1beta1.Gateway: context deadline exceeded with timeout 1m0s`
- Gateway API client created but cannot list resources

**Root Cause**: Helm chart templates conditionally generate RBAC rules based on the `sources` array, not by parsing `extraArgs`. This is by design in the chart architecture.

**Source**: Research findings from kubernetes-sigs/external-dns Helm chart documentation and GitHub issue #5636

#### 2. Secret Key Name: "apiKey"
Despite using API tokens (not legacy API keys), the ExternalDNS Helm chart expects the secret key to be named `apiKey`. Using any other name (e.g., `apiToken`) causes authentication failures with cryptic error messages.

**Source**: Research document section 3.1, GitHub issue kubernetes-sigs/external-dns#4263

#### 3. Zone ID as ExtraArg
The zone ID must be passed as `--zone-id-filter` in `extraArgs`, not as a top-level Helm value `zoneIdFilters`. The latter is silently ignored, causing ExternalDNS to not scope to any zone.

#### 4. Rate Limiting Optimization
Set `--cloudflare-dns-records-per-page=5000` (the maximum) to minimize API calls. Cloudflare's API limit is 1,200 requests per 5 minutes, and without this optimization, large clusters can easily exceed it.

**Source**: Research document section 4.2

#### 5. Domain-Specific Resource Names
Resource names include the manifest's `metadata.name` to support multiple ExternalDNS instances:
- Helm release: `external-dns-planton`, `external-dns-planton-live`
- ServiceAccount: matches release name
- Secret: `cloudflare-api-token-<release-name>`

This enables managing multiple Cloudflare zones (different domains) in the same namespace without conflicts.

### Simplified Configuration Variables

**File**: `apis/project/planton/provider/kubernetes/addon/externaldns/v1/iac/pulumi/module/vars.go`

Removed hardcoded defaults in favor of protobuf-defined defaults:

```go
var vars = struct {
    HelmChartName string
    HelmChartRepo string
}{
    HelmChartName: "external-dns",
    HelmChartRepo: "https://kubernetes-sigs.github.io/external-dns/",
}
```

**Removed**:
- `Namespace` - now comes from spec (default: "external-dns")
- `DefaultChartVersion` - now comes from spec (default: "1.19.0")
- `KsaName` - now derived from `metadata.name`

## Usage Examples

### Basic Cloudflare Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns-example
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  cloudflare:
    apiToken: "YOUR_CLOUDFLARE_API_TOKEN"
    dnsZoneId: "YOUR_ZONE_ID"
    isProxied: true  # Enable orange cloud
```

### Multiple Domains

```yaml
---
# Domain 1: example.com
apiVersion: kubernetes.openmcf.org/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns-example-com
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  cloudflare:
    apiToken: "TOKEN_FOR_EXAMPLE_COM"
    dnsZoneId: "ZONE_ID_EXAMPLE_COM"
    isProxied: false

---
# Domain 2: example.org
apiVersion: kubernetes.openmcf.org/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns-example-org
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  cloudflare:
    apiToken: "TOKEN_FOR_EXAMPLE_ORG"
    dnsZoneId: "ZONE_ID_EXAMPLE_ORG"
    isProxied: true
```

### Using with Kubernetes Services

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
spec:
  type: LoadBalancer
  ports:
  - port: 80
  selector:
    app: my-app
```

When the service gets an external IP, ExternalDNS automatically:
1. Creates an A record: `app.example.com` → `<LoadBalancer-IP>`
2. Creates a TXT record for ownership tracking
3. Enables proxy if `isProxied: true`

### Using with Ingress Resources

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        backend:
          service:
            name: my-app
            port:
              number: 80
```

ExternalDNS automatically creates DNS records pointing to the Ingress controller's IP.

### Using with Gateway API (Istio, Kong, etc.)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: console-planton
  namespace: istio-ingress
  annotations:
    external-dns.alpha.kubernetes.io/hostname: console.planton.cloud
spec:
  addresses:
  - type: Hostname
    value: ingress-external.istio-ingress.svc.cluster.local
  gatewayClassName: istio
  listeners:
  - hostname: console.planton.cloud
    port: 443
    protocol: HTTPS
```

ExternalDNS automatically creates DNS records pointing to the Gateway's external IP.

**Supported Gateway API Sources**:
- `service` - Kubernetes Services (type: LoadBalancer)
- `ingress` - Kubernetes Ingress resources
- `gateway-httproute` - Gateway API Gateway and HTTPRoute resources

## Cloudflare Features

### 1. Proxy Mode (Orange Cloud)

When `isProxied: true`:
- Traffic routes through Cloudflare's global edge network
- Provides DDoS protection, WAF, and CDN caching
- Hides origin server IP from public internet
- Works only for HTTP/HTTPS traffic on standard ports

When `isProxied: false`:
- DNS returns the actual LoadBalancer IP
- Required for non-HTTP services (SSH, databases, custom protocols)

**Default**: `false` (disabled)

### 2. API Rate Limiting Optimization

Cloudflare enforces 1,200 API requests per 5 minutes. The implementation:
- Sets `--cloudflare-dns-records-per-page=5000` (maximum)
- Reduces pagination API calls for large DNS zones
- Prevents `429 Too Many Requests` errors
- Enables management of thousands of DNS records

### 3. Zone Scoping

Uses `--zone-id-filter` to restrict ExternalDNS to specific zones:
- Prevents accidental cross-zone modifications
- Essential for multi-tenant or multi-domain clusters
- Provides additional safety layer beyond API token scoping

## Security Best Practices

### 1. API Token Permissions

**Required Permissions**:
- **Zone** → **Zone** → **Read** (to list and identify zones)
- **Zone** → **DNS** → **Edit** (to create, update, delete records)

**Token Scoping**:
- Scope to specific zones, not "All zones"
- One token per domain for environment isolation
- Set expiration dates and rotate regularly

**How to Create**:
1. Go to Cloudflare Dashboard → My Profile → API Tokens
2. Click "Create Token"
3. Use template "Edit zone DNS" or create custom
4. Set permissions: Zone:Zone:Read + Zone:DNS:Edit
5. Scope to specific zone(s)
6. Copy token immediately (not shown again)

### 2. Kubernetes Secret Management

API tokens are stored in Kubernetes Secrets:
- Never commit tokens to Git repositories
- Use external secret managers (Vault, AWS Secrets Manager) for production
- Rotate tokens regularly and update secrets
- Limit Secret access via Kubernetes RBAC

### 3. Zone ID Filtering

Always specify `dns_zone_id` to:
- Prevent ExternalDNS from managing wrong zones
- Provide fail-safe even if API token has broader permissions
- Enable clear audit trails in Cloudflare logs

## Production Deployment

### Deployed Instances

Successfully deployed to production environment managing two domains:

**Instance 1: planton.cloud**
- Manifest: `external-dns-planton.yaml`
- Zone ID: `7adff2f8326758cac24fd17f02ca3001`
- Helm Release: `external-dns-planton`
- ServiceAccount: `external-dns-planton`
- Status: ✅ Running

**Instance 2: planton.live**
- Manifest: `external-dns-planton-live.yaml`
- Zone ID: `77c6a34cf87dd1e8b497dc895bf5ea1b`
- Helm Release: `external-dns-planton-live`
- ServiceAccount: `external-dns-planton-live`
- Status: ✅ Running

### Verification

```bash
# Check both instances running
kubectl get pods -n external-dns
# Output:
# external-dns-planton-xxx   1/1   Running
# external-dns-planton-live-xxx    1/1   Running

# Check Helm releases
helm ls -n external-dns
# Output:
# external-dns-planton   1.19.0   0.19.0
# external-dns-planton-live    1.19.0   0.19.0

# Check secrets created
kubectl get secret -n external-dns
# Output:
# cloudflare-api-token-external-dns-planton
# cloudflare-api-token-external-dns-planton-live

# Verify zone ID filters and sources in logs
stern external -n external-dns | grep -E "ZoneIDFilter|Sources"
# Should show:
# Sources:[service ingress gateway-httproute]
# ZoneIDFilter:[7adff2f8326758cac24fd17f02ca3001]
# ZoneIDFilter:[77c6a34cf87dd1e8b497dc895bf5ea1b]

# Verify Gateway API RBAC permissions
kubectl get clusterrole external-dns-planton -o yaml | grep -A 10 gateway.networking.k8s.io
# Should show Gateway API permissions
```

### Gateway API Support Testing

Successfully tested Gateway API integration with Istio:

**Test Case**: console.planton.cloud
- **Gateway**: `console-planton` in `istio-ingress` namespace
- **Service**: `ingress-external` with external IP `34.93.244.81`
- **Annotation**: `external-dns.alpha.kubernetes.io/hostname: console.planton.cloud`
- **Result**: ✅ DNS record created automatically in Cloudflare
- **Proxy**: Orange cloud enabled

**Verification Steps Performed**:
1. Added ExternalDNS annotation to Gateway resource
2. Applied manifest to cluster
3. Monitored ExternalDNS logs for record creation
4. Verified A record appeared in Cloudflare dashboard
5. Tested DNS resolution with `dig console.planton.cloud`
6. Confirmed HTTPS access working

## Architecture

### Component Interaction

```
┌─────────────────────────────────────┐
│  Kubernetes Cluster                 │
│                                     │
│  ┌──────────────────────────────┐   │
│  │ Service/Ingress with         │   │
│  │ external-dns annotation      │   │
│  │  hostname: app.example.com   │   │
│  └──────────┬───────────────────┘   │
│             │ Watched by             │
│             ▼                        │
│  ┌──────────────────────────────┐   │
│  │ ExternalDNS Pod              │   │
│  │  - Monitors Kubernetes API   │   │
│  │  - Reads zone ID filter      │   │
│  │  - Reads CF_API_TOKEN        │   │
│  └──────────┬───────────────────┘   │
└─────────────┼───────────────────────┘
              │ API calls
              ▼
   ┌──────────────────────────────┐
   │  Cloudflare API              │
   │  - Zone: example.com         │
   │  - Creates A record          │
   │  - Creates TXT record        │
   │  - Sets proxy status         │
   └──────────────────────────────┘
```

### DNS Record Lifecycle

1. **Service/Ingress Created**: User deploys with annotation
2. **ExternalDNS Detects**: Watches Kubernetes API for changes
3. **Desired State**: Extracts hostname and IP from resource
4. **API Call**: Queries Cloudflare for existing records
5. **Reconciliation**: Creates/updates/deletes records as needed
6. **Ownership**: Creates TXT record to track ownership
7. **Synchronization**: Repeats every 60 seconds (default)

### Ownership Management

ExternalDNS uses TXT records to track ownership:

**Example**:
- A record: `app.example.com` → `203.0.113.10`
- TXT record: `app.example.com` → `"heritage=external-dns,external-dns/owner=default"`

**Purpose**:
- Prevents ExternalDNS from deleting manually created records
- Enables safe operation in pre-existing DNS zones
- Allows multiple ExternalDNS instances with different owners

**Trade-off**: Doubles the number of DNS API operations (A + TXT for each record)

## Troubleshooting

### Common Issues Encountered

#### 1. Zone ID Filter Not Applied

**Symptom**: Logs show `ZoneIDFilter:[]` (empty)

**Cause**: Zone ID passed as Helm value `zoneIdFilters` instead of `extraArg`

**Fix**: Pass as `--zone-id-filter=<ID>` in extraArgs ✅ (Fixed in implementation)

#### 2. Authentication Failure (403 Forbidden)

**Symptom**: 
```
403 Forbidden: code 9109, Unauthorized to access requested resource
```

**Causes**:
- Invalid or expired API token
- Token missing required permissions (Zone:Zone:Read or Zone:DNS:Edit)
- Token not scoped to the target zone

**Fix**:
1. Verify token in Cloudflare Dashboard
2. Check permissions: Zone:Zone:Read + Zone:DNS:Edit
3. Ensure token is scoped to the correct zone
4. Regenerate token if needed and update manifest

#### 3. Secret Key Name Mismatch

**Symptom**: `Invalid request headers (6003)` in logs

**Cause**: Secret key named `apiToken` instead of `apiKey`

**Fix**: Implementation uses `apiKey` as the key name ✅ (Built into implementation)

#### 4. Gateway API Timeout Errors

**Symptom**: 
```
failed to sync *v1beta1.Gateway: context deadline exceeded with timeout 1m0s
failed to sync *v1beta1.HTTPRoute: context deadline exceeded with timeout 1m0s
```

**Cause**: Sources specified in `extraArgs` instead of top-level `sources` parameter. Helm chart does not generate Gateway API RBAC permissions when sources are only in `extraArgs`.

**Fix**: Use top-level `sources` parameter ✅ (Fixed in implementation)

**Verification**:
```bash
# Check ClusterRole has Gateway API permissions
kubectl get clusterrole -l app.kubernetes.io/name=external-dns -o yaml | grep gateway.networking.k8s.io

# Should show:
# - apiGroups: [gateway.networking.k8s.io]
#   resources: [gateways, httproutes, ...]
#   verbs: [get, watch, list]
```

**Source**: Research findings, GitHub issue kubernetes-sigs/external-dns#5636

#### 5. Rate Limiting (Error 1015)

**Symptom**: `429 Too Many Requests`, DNS updates fail

**Mitigation**: Set `--cloudflare-dns-records-per-page=5000` ✅ (Built into implementation)

## Benefits

### 1. Automated DNS Management
- Zero manual DNS configuration for services and ingresses
- DNS records created within seconds of service deployment
- Automatic cleanup when services are deleted
- Reduces operational toil and human errors

### 2. Multi-Cloud Flexibility
- Use Cloudflare DNS with any Kubernetes distribution
- Not limited to cloud provider DNS (GCP, AWS, Azure)
- Works on bare metal, edge, or hybrid cloud

### 3. Cloudflare Edge Network
- Optional proxy mode leverages Cloudflare's global CDN
- DDoS protection and WAF for HTTP/S services
- Performance optimization (caching, compression, protocol upgrades)
- Hide origin server IPs from public internet

### 4. Enterprise-Grade Reliability
- Optimized for large-scale deployments (thousands of records)
- API rate limiting mitigation built-in
- Ownership tracking prevents accidental deletions
- Supports multi-tenant and multi-domain architectures

### 5. Security Hardening
- API tokens follow least-privilege principle
- Tokens scoped to specific zones
- Credentials stored in Kubernetes Secrets
- Zone filtering provides defense in depth

## Migration Guide

### For New Deployments

1. **Create Cloudflare API Token**:
   - Dashboard → API Tokens → Create Token
   - Permissions: Zone:Zone:Read + Zone:DNS:Edit
   - Scope to your DNS zone

2. **Create Manifest**:
   ```yaml
   apiVersion: kubernetes.openmcf.org/v1
   kind: ExternalDnsKubernetes
   metadata:
     name: external-dns-my-domain
   spec:
     targetCluster:
       kubernetesProviderConfigId: <your-cluster-id>
     cloudflare:
       apiToken: "<your-token>"
       dnsZoneId: "<your-zone-id>"
       isProxied: true
   ```

3. **Deploy**:
   ```bash
   export EXTERNAL_DNS_MODULE=~/scm/github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/addon/externaldns/v1/iac/pulumi
   
   openmcf pulumi up \
     --manifest external-dns.yaml \
     --module-dir ${EXTERNAL_DNS_MODULE}
   ```

4. **Test**:
   - Deploy a LoadBalancer service with annotation
   - Verify DNS record created in Cloudflare dashboard

### For Existing Manual DNS Setups

1. **Document Current DNS Records**: Export from Cloudflare before enabling automation
2. **Deploy ExternalDNS**: Start with a test subdomain
3. **Gradual Migration**: Move services one at a time
4. **Verify**: Check each DNS record matches before deleting manual entries
5. **Full Automation**: Once confident, annotate all services

## Testing

### Test Service Deployment

Created test manifest for verification:

**File**: `apis/project/planton/provider/kubernetes/addon/externaldns/v1/_cursor/lb-service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: external-dns-test
  annotations:
    external-dns.alpha.kubernetes.io/hostname: test-external-dns.planton.live
spec:
  type: LoadBalancer
  ports:
  - port: 80
```

**Verification Steps**:
1. Applied test service
2. Verified LoadBalancer IP assigned
3. Checked ExternalDNS logs for record creation
4. Confirmed A record created in Cloudflare dashboard
5. Verified DNS resolution: `dig test-external-dns.planton.live`
6. Tested HTTP access: `curl http://test-external-dns.planton.live`
7. Deleted service and confirmed automatic DNS cleanup

**Result**: ✅ All tests passed

## Known Limitations

### 1. Proxy for HTTP/HTTPS Only
Cloudflare proxy works only for HTTP/HTTPS on standard ports (80, 443, 8080, 8443). Other protocols require Cloudflare Spectrum (separate product).

### 2. No DNS Tag Support
Cloudflare's native DNS Tags (key-value pairs for governance) are not supported. ExternalDNS only supports text comments, limiting infrastructure categorization capabilities.

**Source**: GitHub issue kubernetes-sigs/external-dns#5859

### 3. API Token Rotation
Manual process required:
1. Generate new token in Cloudflare
2. Update manifest with new token
3. Redeploy ExternalDNS stack

Future enhancement: Support secret references for automated rotation.

## Future Enhancements

1. **Secret Reference Support**: Accept reference to existing Kubernetes Secret instead of plain token value
2. **Per-Resource Proxy Control**: Annotation to override proxy setting per service/ingress
3. **Custom TXT Owner ID**: Configurable owner identifier for multi-instance scenarios
4. **Sync Policy Configuration**: Expose `upsert-only` vs `sync` policy in spec
5. **Domain Filters**: Support domain-based filtering in addition to zone ID
6. **DNS Tag Integration**: Support Cloudflare DNS Tags when feature becomes available

## Related Documentation

- **ExternalDNS Official Docs**: https://github.com/kubernetes-sigs/external-dns
- **Cloudflare Provider Guide**: https://kubernetes-sigs.github.io/external-dns/latest/docs/tutorials/cloudflare/
- **Cloudflare API Tokens**: https://developers.cloudflare.com/fundamentals/api/get-started/create-token/
- **Cloudflare Proxy Status**: https://developers.cloudflare.com/dns/proxy-status/
- **Research Document**: `apis/project/planton/provider/kubernetes/addon/externaldns/v1/_cursor/external-dns.cloudflare.research.md`

## Breaking Changes

None. This is a new provider addition with no impact on existing GKE, EKS, or AKS configurations.

## Deployment Status

✅ **Protobuf Specification**: Complete with validation  
✅ **Pulumi Module**: Implemented and tested  
✅ **Multi-Domain Support**: Verified with planton.cloud and planton.live  
✅ **Production Deployment**: Successfully deployed to app-prod cluster  
✅ **Authentication**: Cloudflare API integration working  
✅ **DNS Record Creation**: Automated record creation verified  
✅ **Gateway API Support**: Tested with Istio Gateway resources  
✅ **RBAC Configuration**: Proper permissions for all sources (Service, Ingress, Gateway API)  
✅ **Documentation**: README created with examples and troubleshooting  
✅ **Rate Limiting**: Optimized for production scale  
✅ **Production Testing**: console.planton.cloud DNS record created via Gateway annotation

## Lessons Learned

### 1. Helm Chart RBAC Generation Pattern
The ExternalDNS Helm chart uses the top-level `sources` parameter for RBAC template generation, not `extraArgs`. This is a critical distinction that affects ClusterRole permissions.

### 2. Research-Driven Development
When encountering cryptic errors (timeout on Gateway API sync), structured research prompts helped identify the root cause quickly. The issue was not version incompatibility but configuration placement.

### 3. Gateway API v1 Support
ExternalDNS v0.19.0 fully supports both Gateway API v1 and v1beta1, eliminating concerns about version mismatches with modern Kubernetes clusters.

---

**Next Steps**: Monitor ExternalDNS logs for any rate limiting issues in production, document operational runbooks for token rotation and troubleshooting.


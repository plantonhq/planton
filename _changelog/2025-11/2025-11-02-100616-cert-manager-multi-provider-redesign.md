# CertManager Multi-Provider Redesign with Domain-Named ClusterIssuers

**Date**: November 2, 2025  
**Type**: Breaking Change + Enhancement  
**Components**: API Definitions, Kubernetes Provider, IAC Stack Runner, Provider Framework, User Experience

## Summary

Redesigned the CertManagerKubernetes addon from a single-provider model to a multi-provider architecture that supports managing certificates across multiple DNS providers (Cloudflare, GCP Cloud DNS, AWS Route53, Azure DNS) in a single deployment. The addon now automatically creates one ClusterIssuer per domain, each named after the domain itself (e.g., `planton.ai`, `planton.live`), providing superior visibility and easier troubleshooting compared to the previous single multi-solver ClusterIssuer approach.

## Problem Statement / Motivation

The previous CertManagerKubernetes implementation had a fundamental architectural limitation: it could only configure a single DNS provider per deployment. Organizations managing domains across multiple DNS providers (e.g., public domains on Cloudflare, internal domains on GCP Cloud DNS) were forced to deploy multiple cert-manager instances or manually manage ClusterIssuers.

### Pain Points

- **Single Provider Limitation**: Could only configure one DNS provider (Cloudflare OR GCP OR AWS OR Azure)
- **Manual ClusterIssuer Management**: Users had to manually create ClusterIssuer resources after addon deployment
- **Poor Visibility**: Generic issuer name (`letsencrypt-cluster-issuer`) didn't indicate which domains it managed
- **Difficult Troubleshooting**: When debugging certificate issues, unclear which domain/provider was involved
- **No Multi-Cloud Support**: Organizations with hybrid DNS setups couldn't use a unified cert-manager deployment
- **Inflexible**: Adding a new domain required either redeploying the addon or manually creating issuers

## Solution / What's New

The redesigned addon introduces a **multi-provider configuration model** where users specify an array of DNS providers, each managing one or more DNS zones. The Pulumi implementation automatically creates infrastructure based on this declarative configuration, with a key UX improvement: **one ClusterIssuer per domain**, each named after the domain for immediate recognition.

### Architecture Changes

**Old Structure (Single Provider)**:
```protobuf
spec {
  oneof provider_config {
    CertManagerGkeConfig gke = 100;
    CertManagerEksConfig eks = 101;
    CertManagerAksConfig aks = 102;
    CertManagerCloudflareConfig cloudflare = 103;
  }
}
```

**New Structure (Multi-Provider)**:
```protobuf
spec {
  AcmeConfig acme = 6;  // Global ACME settings
  repeated DnsProviderConfig dns_providers = 7;  // Array of providers
}

message DnsProviderConfig {
  string name = 1;
  repeated string dns_zones = 2;
  oneof provider {
    GcpCloudDnsProvider gcp_cloud_dns = 100;
    AwsRoute53Provider aws_route53 = 101;
    AzureDnsProvider azure_dns = 102;
    CloudflareProvider cloudflare = 103;
  }
}
```

### Key Features

#### 1. Multi-Provider Support

Users can now configure multiple DNS providers in a single deployment:

```yaml
spec:
  acme:
    email: "admin@example.com"
    server: "https://acme-v02.api.letsencrypt.org/directory"
  
  dnsProviders:
    # Cloudflare for public domains
    - name: cloudflare-public
      dnsZones:
        - example.com
        - example.org
      cloudflare:
        apiToken: "..."
    
    # GCP Cloud DNS for internal domains
    - name: gcp-internal
      dnsZones:
        - internal.example.net
      gcpCloudDns:
        projectId: "my-project"
        serviceAccountEmail: "cert-manager@project.iam.gserviceaccount.com"
    
    # AWS Route53 for AWS-hosted domains
    - name: aws-prod
      dnsZones:
        - aws.example.com
      awsRoute53:
        region: "us-east-1"
        roleArn: "arn:aws:iam::123456789012:role/cert-manager"
```

#### 2. Domain-Named ClusterIssuers

Instead of creating a single `letsencrypt-cluster-issuer` with multiple solvers, the addon now creates **one ClusterIssuer per domain**, each named after the domain itself:

**Created Resources** (from example above):
- ClusterIssuer: `example.com`
- ClusterIssuer: `example.org`
- ClusterIssuer: `internal.example.net`
- ClusterIssuer: `aws.example.com`

**Benefits**:
- ✅ Immediate recognition: `kubectl get clusterissuer` shows domain names directly
- ✅ Easier troubleshooting: Know exactly which issuer manages which domain
- ✅ Better isolation: Each domain has its own ACME account and rate limits
- ✅ Simpler references: Just use the domain name in Certificate resources

#### 3. Automatic Resource Creation

The Pulumi module automatically creates all necessary infrastructure:

| Resource | Pattern | Example |
|----------|---------|---------|
| Secrets | `cert-manager-<provider-name>-credentials` | `cert-manager-cloudflare-public-credentials` |
| ClusterIssuers | `<domain>` | `example.com`, `example.org` |
| ACME Keys | `letsencrypt-<domain>-account-key` | `letsencrypt-example.com-account-key` |

#### 4. Global ACME Configuration

Centralized ACME settings apply to all domains:

```yaml
acme:
  email: "admin@example.com"
  server: "https://acme-v02.api.letsencrypt.org/directory"  # or staging
```

All ClusterIssuers use the same email and ACME server, simplifying management.

## Implementation Details

### Protobuf Schema Redesign

**File**: `apis/project/planton/provider/kubernetes/addon/certmanager/v1/spec.proto`

1. **Added `AcmeConfig` message**:
   - `email`: ACME account email (required)
   - `server`: ACME server URL (defaults to Let's Encrypt production)

2. **Added `DnsProviderConfig` message**:
   - `name`: Unique identifier for the provider configuration
   - `dns_zones`: Array of DNS zones this provider manages
   - `provider`: Oneof for provider-specific config

3. **Restructured provider messages**:
   - Simplified field names (e.g., `project_id` → `projectId`)
   - Removed unnecessary foreign key references
   - Focused on essential credentials only

4. **Updated `CertManagerKubernetesSpec`**:
   - Removed `oneof provider_config`
   - Added `AcmeConfig acme` (required)
   - Added `repeated DnsProviderConfig dns_providers` (min 1 required)

### Pulumi Implementation

**File**: `apis/project/planton/provider/kubernetes/addon/certmanager/v1/iac/pulumi/module/main.go`

**Key Changes**:

1. **Multi-Provider Secret Creation**:
```go
cloudflareSecrets := make(map[string]pulumi.StringOutput)
for _, dnsProvider := range spec.DnsProviders {
    if cf := dnsProvider.GetCloudflare(); cf != nil {
        secretName := fmt.Sprintf("cert-manager-%s-credentials", dnsProvider.Name)
        secret, err := corev1.NewSecret(ctx, secretName, ...)
        cloudflareSecrets[dnsProvider.Name] = secret.Metadata.Name().Elem()
    }
}
```

2. **ServiceAccount Annotation Aggregation**:
```go
annotations := pulumi.StringMap{}
for _, dnsProvider := range spec.DnsProviders {
    if gcp := dnsProvider.GetGcpCloudDns(); gcp != nil {
        annotations["iam.gke.io/gcp-service-account"] = pulumi.String(gcp.ServiceAccountEmail)
    } else if aws := dnsProvider.GetAwsRoute53(); aws != nil {
        annotations["eks.amazonaws.com/role-arn"] = pulumi.String(aws.RoleArn)
    }
    // ... etc
}
```

3. **Domain-Specific ClusterIssuer Creation**:
```go
for _, dnsProvider := range spec.DnsProviders {
    for _, dnsZone := range dnsProvider.DnsZones {
        err = createClusterIssuerForDomain(ctx, kubeProvider, helmRelease, 
            spec, cloudflareSecrets, dnsProvider, dnsZone)
    }
}
```

4. **Per-Domain Issuer Function**:
```go
func createClusterIssuerForDomain(
    ctx *pulumi.Context,
    kubeProvider *kubernetes.Provider,
    helmRelease *helm.Release,
    spec *certmanagerv1.CertManagerKubernetesSpec,
    cloudflareSecrets map[string]pulumi.StringOutput,
    dnsProvider *certmanagerv1.DnsProviderConfig,
    domain string,
) error {
    issuerName := domain  // ClusterIssuer named after domain
    
    // Build single solver for this domain
    var solverConfig map[string]interface{}
    // ... provider-specific solver configuration
    
    // Create ClusterIssuer
    _, err := apiextensionsv1.NewCustomResource(ctx, issuerName,
        &apiextensionsv1.CustomResourceArgs{
            ApiVersion: pulumi.String("cert-manager.io/v1"),
            Kind:       pulumi.String("ClusterIssuer"),
            Metadata: &metav1.ObjectMetaArgs{
                Name: pulumi.String(issuerName),  // e.g., "planton.ai"
            },
            OtherFields: map[string]interface{}{
                "spec": map[string]interface{}{
                    "acme": map[string]interface{}{
                        "email":  spec.Acme.Email,
                        "server": spec.Acme.GetServer(),
                        "privateKeySecretRef": map[string]interface{}{
                            "name": fmt.Sprintf("letsencrypt-%s-account-key", domain),
                        },
                        "solvers": []interface{}{solverConfig},
                    },
                },
            },
        },
        pulumi.Provider(kubeProvider),
        pulumi.DependsOn([]pulumi.Resource{helmRelease}))
    
    return err
}
```

5. **DNS Propagation Reliability**:
```go
// Always configure public recursive nameservers for reliable DNS checks
"extraArgs": pulumi.Array{
    pulumi.String("--dns01-recursive-nameservers-only"),
    pulumi.String("--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"),
}
```

### Documentation Overhaul

**File**: `apis/project/planton/provider/kubernetes/addon/certmanager/v1/README.md`

Complete rewrite (600+ lines) covering:
- Multi-provider configuration examples
- Domain-named ClusterIssuer pattern
- Certificate request examples per domain
- Updated troubleshooting commands
- FAQ updates for new architecture

## Breaking Changes

### Configuration Structure

**Old Configuration** (No Longer Valid):
```yaml
spec:
  cloudflare:
    apiToken: "..."
```

**New Configuration** (Required):
```yaml
spec:
  acme:
    email: "admin@example.com"
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
      cloudflare:
        apiToken: "..."
```

### ClusterIssuer Names

**Old**: Single issuer named `letsencrypt-cluster-issuer`  
**New**: One issuer per domain, named after the domain (e.g., `planton.ai`)

### Certificate Resource Updates

**Before**:
```yaml
issuerRef:
  name: letsencrypt-cluster-issuer
  kind: ClusterIssuer
```

**After**:
```yaml
issuerRef:
  name: planton.ai  # Use the domain name
  kind: ClusterIssuer
```

### Migration Guide

For existing deployments:

1. **Backup existing ClusterIssuers**:
```bash
kubectl get clusterissuer -o yaml > clusterissuers-backup.yaml
```

2. **Update addon configuration**:
```yaml
spec:
  acme:
    email: "your-email@example.com"
    server: "https://acme-v02.api.letsencrypt.org/directory"
  
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - your-domain.com
      cloudflare:
        apiToken: "your-existing-token"
```

3. **Redeploy addon**:
```bash
planton pulumi up --manifest cert-manager.yaml --module-dir ${MODULE}
```

4. **Update Certificate resources**:
```bash
# Find all certificates using old issuer
kubectl get certificate -A -o json | jq -r '.items[] | 
  select(.spec.issuerRef.name == "letsencrypt-cluster-issuer") | 
  "\(.metadata.namespace)/\(.metadata.name)"'

# Update each certificate to use domain-named issuer
kubectl patch certificate <name> -n <namespace> --type merge -p '
spec:
  issuerRef:
    name: your-domain.com
'
```

5. **Verify new issuers**:
```bash
kubectl get clusterissuer
# Should show: your-domain.com
```

6. **Clean up old issuer** (optional):
```bash
kubectl delete clusterissuer letsencrypt-cluster-issuer
```

## Benefits

### Developer Experience

1. **Immediate Clarity**: `kubectl get clusterissuer` output is self-documenting:
```
NAME              READY   AGE
planton.ai     True    5d
planton.live      True    5d
internal.corp     True    5d
```

2. **Reduced Cognitive Load**: No need to remember generic issuer names or check which domains they manage

3. **Faster Debugging**:
```bash
# Old way: Which issuer handles planton.ai?
kubectl describe clusterissuer letsencrypt-cluster-issuer | grep -A20 solvers

# New way: Direct inspection
kubectl describe clusterissuer planton.ai
```

### Operational Benefits

1. **Multi-Cloud DNS Management**: Single cert-manager deployment for hybrid infrastructure
2. **Better Isolation**: Each domain has separate ACME account and rate limits
3. **Cleaner Organization**: Domain-based resource naming aligns with DNS ownership
4. **Easier Onboarding**: New domains just add to `dnsProviders` array

### Technical Improvements

1. **Reliable DNS Propagation**: Public recursive nameservers (1.1.1.1, 8.8.8.8) configured by default
2. **Version Safety**: Minimum cert-manager v1.16.4 enforced for Cloudflare API compatibility
3. **Automatic Discovery**: Cloudflare zone IDs auto-discovered via API (no manual zone ID required)
4. **Flexible Credentials**: Mix of cloud provider identity (GKE/EKS/AKS) and API tokens (Cloudflare)

## Impact

### Who's Affected

- **Platform Engineers**: Must migrate existing configurations to new structure
- **DevOps Teams**: Certificate resources need issuer name updates
- **New Users**: Benefit from clearer, more intuitive configuration

### What Changes

**For Platform Engineers**:
- Configuration file structure (breaking change)
- New multi-provider capability
- Domain-named ClusterIssuer pattern

**For Application Developers**:
- Certificate `issuerRef.name` must use domain name instead of generic issuer
- Otherwise, no changes to Certificate creation workflow

**For Operations**:
- Better visibility into certificate management
- Easier troubleshooting with domain-named resources
- Simplified multi-domain management

### Backward Compatibility

⚠️ **This is a breaking change**. Existing configurations using the old single-provider model will not work with the new API structure. Migration is required.

**No gradual migration path**: The protobuf schema change is incompatible with old configurations.

## Testing Strategy

### Manual Verification

1. **Multi-Provider Configuration**:
```bash
# Deploy with Cloudflare and GCP providers
planton pulumi up --manifest cert-manager-multi.yaml

# Verify ClusterIssuers created
kubectl get clusterissuer
# Should show: planton.ai, planton.live, internal.corp

# Check Cloudflare secret
kubectl get secret -n cert-manager cert-manager-cloudflare-prod-credentials
```

2. **Certificate Issuance**:
```bash
# Create test certificate
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: test-cert
  namespace: default
spec:
  secretName: test-tls
  issuerRef:
    name: planton.ai
    kind: ClusterIssuer
  dnsNames:
    - test.planton.ai
EOF

# Monitor certificate creation
kubectl describe certificate test-cert -n default
kubectl get challenge -n default
kubectl logs -n cert-manager deployment/cert-manager -f
```

3. **DNS Propagation Verification**:
```bash
# Verify TXT record creation during challenge
dig _acme-challenge.test.planton.ai TXT @1.1.1.1 +short
```

### Integration Tests

- **Cloudflare Provider**: Verified with production Planton deployment (planton.ai, planton.live)
- **GCP Provider**: Tested with internal GKE cluster + Cloud DNS
- **Build Verification**: `go build` successful with no compilation errors
- **Lint Checks**: No linter errors in updated code

## Code Metrics

### Files Changed

**Core Implementation**:
- `spec.proto`: Complete restructure (85 → 119 lines)
- `iac/pulumi/module/main.go`: Major rewrite (141 → 250 lines)
- `iac/pulumi/module/outputs.go`: Updated exports

**Documentation & Examples**:
- `README.md`: Complete rewrite (610 → 674 lines)
- `hack/example-clusterissuer-cloudflare.yaml`: Updated for domain-named issuers
- `hack/test-cloudflare.yaml`: Updated to multi-provider structure

### Generated Code

- `spec.pb.go`: Regenerated from updated proto schema
- All protobuf-generated code updated via `make build`

### Net Changes

- **Protobuf API**: ~40% expansion (new messages, repeated fields)
- **Pulumi Module**: ~75% rewrite (multi-provider logic, per-domain issuers)
- **Documentation**: ~10% expansion (multi-provider examples, new patterns)
- **Overall LOC**: +~350 lines (primarily documentation and examples)

## Design Decisions

### Why One Issuer Per Domain?

**Considered Alternatives**:

1. **Single Multi-Solver ClusterIssuer** (Original Design):
   - ✅ Fewer Kubernetes resources
   - ❌ Poor visibility (generic name)
   - ❌ Harder troubleshooting
   - ❌ Shared ACME account across all domains

2. **One Issuer Per Provider** (Alternative):
   - ✅ Fewer issuers than per-domain
   - ❌ Still requires selector logic for routing
   - ❌ Not immediately clear which domains are managed

3. **Domain-Named Issuers** (Chosen):
   - ✅ Immediate visibility: `kubectl get clusterissuer` is self-documenting
   - ✅ Direct mapping: domain → issuer (no ambiguity)
   - ✅ Better isolation: separate ACME accounts and rate limits
   - ✅ Easier troubleshooting: `kubectl describe clusterissuer planton.ai`
   - ❌ More Kubernetes resources (acceptable trade-off)

**Rationale**: The visibility and troubleshooting benefits outweigh the resource overhead. Modern Kubernetes clusters easily handle hundreds of CRDs, and the operational clarity is invaluable.

### Why Global ACME Config?

All domains typically use the same ACME server (Let's Encrypt production or staging) and notification email. Duplicating this across providers would be:
- Redundant
- Error-prone (inconsistent emails)
- Harder to manage (switching prod ↔ staging)

Global configuration simplifies common use cases while still allowing staging/production separation via separate addon deployments.

### Why Not Auto-Create Certificates?

The addon creates ClusterIssuers but not Certificate resources because:
- Certificate creation is application-specific (DNS names, namespaces vary)
- cert-manager's Ingress annotation (`cert-manager.io/cluster-issuer`) provides automatic creation when needed
- Explicit Certificate resources give developers full control
- Aligns with Kubernetes declarative model

## Known Limitations

1. **Multi-Domain Certificates Across Providers**: A single Certificate resource cannot span domains from different providers (e.g., can't have `planton.ai` and `internal.corp` in same cert). Workaround: Create separate certificates.

2. **Provider Migration**: Changing a domain's DNS provider requires deleting and recreating the ClusterIssuer (Pulumi handles this automatically).

3. **Staging vs Production**: Requires deploying separate addon instances with different `acme.server` URLs. Cannot have both staging and production issuers for the same domain in one deployment.

## Future Enhancements

1. **Automatic Certificate Discovery**: Scan Ingress resources and auto-create Certificates
2. **Certificate Monitoring**: Prometheus metrics for expiry, renewal failures
3. **Multi-Account Cloudflare**: Support for multiple Cloudflare accounts (currently one token per provider)
4. **Certificate Consolidation**: Automatic merging of same-domain certificates
5. **ACME Account Reuse**: Option to share ACME accounts across domains

## Related Work

### Previous Context

- **Initial Cloudflare Support**: Research report at `_cursor/cert-manager-cloudflare-support.report.md` (580 lines)
- **Multi-Solver Research**: Confirmation at `_cursor/multiple-solvers.report.md` that single ClusterIssuer can handle multiple providers

### Complementary Features

- **ExternalDNS Integration**: Uses same Cloudflare API tokens (Zone:Zone:Read + Zone:DNS:Edit)
- **Kubernetes Addon Framework**: Follows established patterns for provider configuration

### Alignment with Technical Report

The implementation aligns with cert-manager best practices from the technical report:
- ✅ API Token authentication (not legacy API Keys)
- ✅ Public recursive nameservers for DNS propagation
- ✅ Minimum cert-manager v1.16.4 for Cloudflare API compatibility
- ✅ Zone ID auto-discovery (no manual zone IDs required)
- ✅ Secret-based credential management

**Deviation**: Report recommended user-managed ClusterIssuers; implementation auto-creates them for better UX.

---

**Status**: ✅ Production Ready  
**Migration Required**: Yes (breaking change)  
**Deployment**: Planton production (planton.ai, planton.live)


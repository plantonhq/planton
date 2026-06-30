# GcpDnsRecord: Technical Research Document

## Introduction

DNS (Domain Name System) record management is a foundational aspect of modern cloud infrastructure. Google Cloud DNS is a high-performance, resilient, global Domain Name System service that publishes domain names to the global DNS. This document explores the landscape of DNS record management in GCP and explains Planton's approach to providing a streamlined, declarative interface for DNS record operations.

## The Evolution of DNS Management

### Traditional Approaches

Historically, DNS management involved:
1. **Manual Console Updates**: Administrators logged into provider dashboards to add/modify records
2. **Zone File Editing**: Direct manipulation of BIND-style zone files
3. **CLI Tools**: Using `gcloud dns record-sets` commands for record management

These approaches suffered from:
- No version control for DNS changes
- Lack of auditability
- Manual error-prone processes
- Difficulty in maintaining consistency across environments

### Modern Infrastructure as Code

The shift to Infrastructure as Code (IaC) brought declarative DNS management:
- **Terraform**: Google provider's `google_dns_record_set` resource
- **Pulumi**: GCP classic provider's DNS record set resources
- **Crossplane**: Kubernetes-style CRDs for GCP DNS
- **Cloud Foundation Toolkit**: Google's Terraform modules

## Google Cloud DNS Record Types

Cloud DNS supports the following record types:

| Type | Purpose | Example Value |
|------|---------|---------------|
| A | IPv4 address | 192.0.2.1 |
| AAAA | IPv6 address | 2001:db8::1 |
| CNAME | Canonical name (alias) | target.example.com. |
| MX | Mail exchange | 10 mail.example.com. |
| TXT | Text record | "v=spf1 include:_spf.google.com ~all" |
| SRV | Service location | 10 5 5269 xmpp.example.com. |
| NS | Nameserver | ns-cloud-d1.googledomains.com. |
| PTR | Reverse DNS | host.example.com. |
| CAA | Certificate Authority Authorization | 0 issue "letsencrypt.org" |
| SOA | Start of Authority | ns.example.com. admin.example.com. 1 ... |

## Deployment Methods Comparison

### 1. Google Cloud Console

**Pros:**
- Visual interface for quick changes
- Built-in validation
- No tooling setup required

**Cons:**
- No version control
- Manual process prone to errors
- No automation capability
- Difficult to replicate across environments

### 2. gcloud CLI

```bash
gcloud dns record-sets create www.example.com. \
  --zone=example-zone \
  --type=A \
  --ttl=300 \
  --rrdatas="192.0.2.1"
```

**Pros:**
- Scriptable
- Can be integrated into CI/CD
- Quick for one-off changes

**Cons:**
- Imperative rather than declarative
- State not tracked
- Requires careful scripting for idempotency

### 3. Terraform

```hcl
resource "google_dns_record_set" "www" {
  name         = "www.example.com."
  type         = "A"
  ttl          = 300
  managed_zone = "example-zone"
  rrdatas      = ["192.0.2.1"]
}
```

**Pros:**
- Declarative configuration
- State management
- Drift detection
- Plan before apply

**Cons:**
- Requires Terraform expertise
- State file management overhead
- Provider version compatibility

### 4. Pulumi (Go)

```go
record, err := dns.NewRecordSet(ctx, "www", &dns.RecordSetArgs{
    Name:        pulumi.String("www.example.com."),
    Type:        pulumi.String("A"),
    Ttl:         pulumi.Int(300),
    ManagedZone: pulumi.String("example-zone"),
    Rrdatas:     pulumi.StringArray{pulumi.String("192.0.2.1")},
})
```

**Pros:**
- Full programming language capabilities
- Type safety
- Reusable components
- Testing support

**Cons:**
- Steeper learning curve
- More complex setup
- Requires programming knowledge

### 5. Planton

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsRecord
metadata:
  name: www-example
spec:
  projectId: my-project
  managedZone: example-zone
  recordType: A
  name: www.example.com.
  values:
    - 192.0.2.1
  ttlSeconds: 300
```

**Pros:**
- Simple YAML configuration
- KRM-style familiar to Kubernetes users
- Built-in validation
- Dual IaC backend (Pulumi + Terraform)
- No IaC expertise required

**Cons:**
- Abstraction over native tools
- Limited to supported record types

## Planton's Approach

### 80/20 Design Philosophy

The `GcpDnsRecord` component focuses on the most common DNS record management needs:

**In Scope (80% of use cases):**
- Standard record types (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA)
- TTL configuration
- Multiple values for round-robin
- Wildcard records
- Reference to existing managed zones

**Out of Scope (edge cases):**
- DNSSEC signing operations (managed at zone level)
- Private DNS forwarding rules
- DNS policies and logging
- Record set transactions/batching
- Response policies

### Why a Separate Record Component?

Planton provides both `GcpDnsZone` (which can include inline records) and `GcpDnsRecord` for different use cases:

| Use Case | Recommended Component |
|----------|----------------------|
| Zone + records managed together | GcpDnsZone with inline records |
| Records managed by different teams | GcpDnsRecord per team |
| Dynamic record creation | GcpDnsRecord |
| Fine-grained access control | GcpDnsRecord |
| Records in external zones | GcpDnsRecord |

### Validation Strategy

The component enforces:

1. **Required Fields**: project_id, managed_zone, record_type, name, values
2. **Format Validation**: 
   - DNS name must be FQDN (ending with dot)
   - Managed zone name follows GCP naming conventions
3. **Range Validation**: TTL between 1-86400 seconds
4. **Enum Validation**: Record type must be a valid DNS record type

## Implementation Landscape

### Pulumi Module Design

The Pulumi module:
1. Loads stack input from environment
2. Configures GCP provider with credentials
3. Creates `dns.RecordSet` resource
4. Exports outputs (FQDN, record type, TTL)

### Terraform Module Design

The Terraform module:
1. Accepts variables matching spec.proto
2. Configures google provider
3. Creates `google_dns_record_set` resource
4. Outputs record metadata

### Feature Parity

Both implementations:
- Create the same underlying `google_dns_record_set` resource
- Support all record types
- Handle multiple values identically
- Produce equivalent outputs

## Production Best Practices

### TTL Guidelines

| Scenario | Recommended TTL | Reason |
|----------|-----------------|--------|
| Production stable | 3600-86400 | Reduce DNS queries |
| Pre-migration | 60-300 | Enable quick failover |
| Development | 60-300 | Faster iteration |
| MX records | 3600+ | Email delivery stability |
| TXT (verification) | 3600 | Service provider caching |

### Security Considerations

1. **CAA Records**: Always configure CAA to limit certificate issuance
2. **SPF Records**: Prevent email spoofing with proper SPF
3. **DMARC**: Add DMARC records for email authentication
4. **Minimal Permissions**: Use service accounts with DNS Admin role only where needed

### Common Patterns

#### Blue-Green Deployments

```yaml
# Blue environment
apiVersion: gcp.planton.dev/v1
kind: GcpDnsRecord
metadata:
  name: api-blue
spec:
  projectId: my-project
  managedZone: example-zone
  recordType: CNAME
  name: api-blue.example.com.
  values:
    - blue-lb.example.com.
---
# Production (points to active)
apiVersion: gcp.planton.dev/v1
kind: GcpDnsRecord
metadata:
  name: api-production
spec:
  projectId: my-project
  managedZone: example-zone
  recordType: CNAME
  name: api.example.com.
  values:
    - api-blue.example.com.
  ttlSeconds: 60  # Low TTL for quick switch
```

#### Multi-Region with GeoDNS

Note: GCP Cloud DNS doesn't support GeoDNS natively. Use Cloud Load Balancing with Backend Services for geographic routing. DNS records should point to the load balancer.

## Common Pitfalls

1. **Missing Trailing Dot**: DNS names must end with a dot (FQDN format)
2. **TTL Too High**: High TTL during migration causes extended propagation
3. **Forgotten Reverse DNS**: PTR records often forgotten for mail servers
4. **Wildcard Conflicts**: Explicit records override wildcards
5. **Zone Mismatches**: Record must be within the managed zone's domain

## Conclusion

`GcpDnsRecord` provides a declarative, validated interface for managing DNS records in Google Cloud DNS. By focusing on the 80% of common use cases while maintaining feature parity between Pulumi and Terraform backends, it enables teams to manage DNS as code with confidence.

The component's design separates record management from zone management, enabling:
- Team-specific DNS management
- Granular access control
- Independent record lifecycle
- Integration with existing zones

For advanced DNS features like DNSSEC management or DNS policies, users should combine `GcpDnsRecord` with native GCP tooling or consider the `GcpDnsZone` component which provides zone-level features.

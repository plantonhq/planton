# AWS Route53 DNS Record - Technical Research Documentation

## Introduction

AWS Route53 is Amazon's highly available and scalable Domain Name System (DNS) web service. While Route53 offers a comprehensive set of features including domain registration, DNS routing, and health checking, this document focuses specifically on **DNS resource records** - the fundamental building blocks of DNS that map domain names to IP addresses or other resources.

This research document analyzes the DNS record management landscape, compares deployment approaches, and explains Planton's design decisions for the `AwsRoute53DnsRecord` component.

## The DNS Record Landscape

### What Are DNS Records?

DNS records are entries in a DNS zone that define how domain names are resolved. Common record types include:

| Type | Purpose | Example |
|------|---------|---------|
| A | Maps domain to IPv4 address | `www.example.com → 192.0.2.1` |
| AAAA | Maps domain to IPv6 address | `www.example.com → 2001:db8::1` |
| CNAME | Alias to another domain | `blog.example.com → example.ghost.io` |
| MX | Mail exchange servers | `example.com → 10 mail.example.com` |
| TXT | Text data (SPF, DKIM, verification) | `example.com → "v=spf1..."` |
| NS | Nameserver delegation | `example.com → ns1.awsdns-01.com` |
| CAA | Certificate Authority Authorization | `example.com → 0 issue "letsencrypt.org"` |

### Route53's Unique Features

Route53 offers capabilities beyond standard DNS:

**1. Alias Records**
- Route53's proprietary extension to DNS
- Can point zone apex (naked domain) to AWS resources
- Free queries for alias records pointing to AWS resources
- Automatic IP updates when target resource IPs change
- Supports: CloudFront, ALB/NLB, S3 website, API Gateway, Elastic Beanstalk, VPC endpoints

**2. Advanced Routing Policies**
- **Simple**: Standard DNS behavior (default)
- **Weighted**: Distribute traffic by percentage (0-255 weight)
- **Latency**: Route to lowest-latency endpoint
- **Failover**: Active-passive with health checks
- **Geolocation**: Route by continent/country/US state
- **Multivalue Answer**: Return multiple healthy IPs

**3. Health Checks Integration**
- HTTP/HTTPS/TCP endpoint monitoring
- CloudWatch alarm-based health checks
- Automatic failover when health checks fail

## Deployment Methods Analysis

### 1. AWS Console (Manual)

**Pros:**
- Visual interface for simple changes
- Good for learning and exploration
- Immediate feedback

**Cons:**
- No version control
- Error-prone for complex configurations
- Not scalable for many records
- No audit trail

**Use Case:** Ad-hoc changes, debugging, initial exploration

### 2. AWS CLI

```bash
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch file://record.json
```

**Pros:**
- Scriptable
- Can be version controlled
- Supports all features

**Cons:**
- Complex JSON format
- No state management
- Manual change tracking
- No drift detection

**Use Case:** Quick scripts, CI/CD integration

### 3. AWS CloudFormation

```yaml
Resources:
  DNSRecord:
    Type: AWS::Route53::RecordSet
    Properties:
      HostedZoneId: Z1234567890ABC
      Name: www.example.com
      Type: A
      TTL: 300
      ResourceRecords:
        - 192.0.2.1
```

**Pros:**
- Native AWS integration
- Stack-based management
- Rollback support

**Cons:**
- AWS-only
- YAML/JSON syntax verbose
- Limited cross-stack references
- Slower deployment cycles

### 4. Terraform/OpenTofu

```hcl
resource "aws_route53_record" "www" {
  zone_id = "Z1234567890ABC"
  name    = "www.example.com"
  type    = "A"
  ttl     = 300
  records = ["192.0.2.1"]
}
```

**Pros:**
- Multi-cloud support
- Excellent state management
- Large community
- Rich ecosystem

**Cons:**
- HCL learning curve
- State file management complexity
- Provider version compatibility

### 5. Pulumi

```go
record, err := route53.NewRecord(ctx, "www", &route53.RecordArgs{
    ZoneId:  pulumi.String("Z1234567890ABC"),
    Name:    pulumi.String("www.example.com"),
    Type:    pulumi.String("A"),
    Ttl:     pulumi.Int(300),
    Records: pulumi.StringArray{pulumi.String("192.0.2.1")},
})
```

**Pros:**
- Real programming languages
- Strong typing
- Excellent IDE support
- Component abstractions

**Cons:**
- Requires programming knowledge
- Smaller community than Terraform

### 6. ExternalDNS (Kubernetes)

ExternalDNS automatically creates DNS records from Kubernetes Service/Ingress annotations:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: www.example.com
spec:
  type: LoadBalancer
```

**Pros:**
- Automatic record management
- Kubernetes-native
- Supports multiple providers

**Cons:**
- Kubernetes-only
- Limited to service discovery use case
- No advanced routing policies
- Less control over record details

## Comparative Analysis

| Approach | Declarative | State Mgmt | Multi-Cloud | Learning Curve | Advanced Features |
|----------|------------|------------|-------------|----------------|-------------------|
| Console | ❌ | ❌ | ❌ | Low | ✅ |
| CLI | Partial | ❌ | ❌ | Medium | ✅ |
| CloudFormation | ✅ | ✅ | ❌ | Medium | ✅ |
| Terraform | ✅ | ✅ | ✅ | Medium | ✅ |
| Pulumi | ✅ | ✅ | ✅ | Medium-High | ✅ |
| ExternalDNS | ✅ | Partial | ✅ | Low | ❌ |

## Planton's Approach

### Design Philosophy

`AwsRoute53DnsRecord` follows the 80/20 principle - exposing the 20% of features that serve 80% of use cases while maintaining access to Route53's powerful capabilities.

### Why a Standalone DNS Record Component?

While `AwsRoute53Zone` can create records inline within a zone, a standalone record component provides:

1. **Granular Management**: Create/update/delete individual records without touching the zone
2. **Cross-Account Records**: Create records in zones owned by different AWS accounts
3. **Team Autonomy**: Application teams manage their records without zone access
4. **Modular Composition**: Combine with other components (ALB, CloudFront) in pipelines

### Scope Decisions

**In Scope (80/20 Focus):**
- Standard record types: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA
- Alias records (Route53's killer feature)
- All routing policies: weighted, latency, failover, geolocation
- Health check integration
- Wildcard records

**Out of Scope:**
- Traffic flow policies (complex visual routing)
- DNSSEC key management (handled at zone level)
- Private hosted zone creation (use AwsRoute53Zone)
- Domain registration (separate workflow)
- Health check creation (separate resource)

### Schema Design Rationale

**1. Using Shared DnsRecordType Enum**
Instead of creating an AWS-specific enum, we use the shared `DnsRecordType` from `dev.planton.shared.networking.enums.dnsrecordtype`. This provides consistency across providers (AWS, GCP, Cloudflare) and enables potential cross-provider DNS abstractions.

**2. Values vs Alias as Mutually Exclusive**
The schema enforces that `values` (for standard records) and `alias_target` (for alias records) cannot both be specified. This matches Route53's API behavior and prevents configuration errors.

**3. Routing Policy as Oneof**
Only one routing policy can be active per record. The schema uses protobuf `oneof` to enforce this at the type level, preventing invalid combinations.

**4. Set Identifier Validation**
The schema validates that `set_identifier` is required when using non-simple routing policies, preventing deployment failures.

## Implementation Landscape

### Pulumi Implementation

The Pulumi module uses the `aws.route53.Record` resource:

```go
recordArgs := &route53.RecordArgs{
    ZoneId: pulumi.String(spec.HostedZoneId),
    Name:   pulumi.String(spec.Name),
    Type:   pulumi.String(spec.Type.String()),
}

if len(spec.Values) > 0 {
    recordArgs.Ttl = pulumi.Int(spec.Ttl)
    recordArgs.Records = pulumi.ToStringArray(spec.Values)
} else if spec.AliasTarget != nil {
    recordArgs.Aliases = route53.RecordAliasArray{
        &route53.RecordAliasArgs{
            Name:                 pulumi.String(spec.AliasTarget.DnsName),
            ZoneId:               pulumi.String(spec.AliasTarget.HostedZoneId),
            EvaluateTargetHealth: pulumi.Bool(spec.AliasTarget.EvaluateTargetHealth),
        },
    }
}
```

### Terraform Implementation

The Terraform module uses `aws_route53_record`:

```hcl
resource "aws_route53_record" "record" {
  zone_id = var.hosted_zone_id
  name    = var.name
  type    = var.type
  
  dynamic "alias" {
    for_each = var.alias_target != null ? [var.alias_target] : []
    content {
      name                   = alias.value.dns_name
      zone_id                = alias.value.hosted_zone_id
      evaluate_target_health = alias.value.evaluate_target_health
    }
  }
  
  ttl     = var.alias_target == null ? var.ttl : null
  records = var.alias_target == null ? var.values : null
}
```

## Production Best Practices

### 1. TTL Strategy

| Record Purpose | Recommended TTL | Rationale |
|----------------|-----------------|-----------|
| Stable records (MX, NS) | 86400 (1 day) | Reduce query costs |
| Standard records | 300 (5 min) | Balance caching and agility |
| Records that may change | 60 (1 min) | Faster failover |
| During migrations | 60 or lower | Quick cutover |

### 2. Alias Records Best Practices

- **Always use alias for AWS resources**: Free queries, automatic IP updates
- **Always use alias for zone apex**: CNAME not allowed at apex
- **Enable evaluate_target_health for failover**: Ensures unhealthy targets are skipped

### 3. Routing Policy Guidelines

- **Weighted**: Start with 95/5 split for canary, gradually shift
- **Latency**: Ensure endpoints are actually in the specified regions
- **Failover**: Always configure health checks for primary
- **Geolocation**: Always include a default record for unmatched locations

### 4. Health Check Integration

- Create health checks as separate resources
- Use 3+ regions for health check origins
- Set appropriate failure thresholds (typically 3)
- Monitor health check status in CloudWatch

## Common Pitfalls

1. **Forgetting trailing dot**: Route53 adds it automatically, but inconsistency can cause issues
2. **Using CNAME at apex**: Not allowed by DNS spec - use alias instead
3. **Missing set_identifier**: Required for all non-simple routing policies
4. **TTL on alias records**: TTL is ignored for alias records (uses target's TTL)
5. **Wrong hosted zone ID for alias**: The alias `hosted_zone_id` is the target service's zone, not your zone

## Cost Considerations

| Operation | Cost (as of 2024) |
|-----------|-------------------|
| Hosted zone | $0.50/month |
| Standard queries | $0.40/million |
| Latency queries | $0.60/million |
| Geo DNS queries | $0.70/million |
| Alias queries to AWS | Free |
| Health checks | $0.50-$2.00/month |

**Cost Optimization Tips:**
- Use alias records for AWS resources (free queries)
- Higher TTLs reduce query volume
- Consolidate related records to reduce zone count

## Conclusion

`AwsRoute53DnsRecord` provides a focused, declarative interface for managing Route53 DNS records. By supporting the full range of record types, alias records, and routing policies, it covers the vast majority of DNS management needs while maintaining simplicity through careful API design and built-in validation.

The component integrates seamlessly with other Planton AWS components, enabling sophisticated infrastructure patterns like multi-region deployments, blue/green releases, and automatic failover - all managed as code through a consistent, validated interface.

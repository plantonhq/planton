# Cloudflare DNS Record: Technical Research Documentation

## Introduction

DNS records are the foundational building blocks of the Domain Name System—they translate human-readable domain names into machine-usable addresses and routing information. Every website, email system, and internet-connected service relies on DNS records to function.

Cloudflare DNS offers a unique proposition: authoritative DNS served from 330+ global edge locations with built-in DDoS protection, zero per-query charges, and the ability to proxy traffic through Cloudflare's CDN/WAF infrastructure. This combination of DNS and edge services is what makes Cloudflare DNS records special—a simple A record isn't just DNS resolution, it can be a gateway to Cloudflare's entire security and performance platform.

This document provides comprehensive research into DNS record management on Cloudflare, examining the deployment landscape from manual operations to Infrastructure-as-Code, and explaining why Planton's approach offers the optimal balance of simplicity and power.

## The Evolution of DNS Management

### Early Days: Manual Zone Files

DNS configuration started with hand-edited zone files on BIND nameservers:

```
$TTL 86400
@   IN  SOA ns1.example.com. admin.example.com. (
            2024012301 ; Serial
            3600       ; Refresh
            1800       ; Retry
            604800     ; Expire
            86400 )    ; Minimum TTL

@   IN  NS  ns1.example.com.
@   IN  NS  ns2.example.com.
www IN  A   192.0.2.1
```

This approach required:
- Direct server access
- Manual file editing
- Service restarts to apply changes
- No validation before changes went live
- Expertise in zone file syntax

### Cloud DNS Services

Cloud providers (AWS Route53, Google Cloud DNS, Azure DNS) introduced API-driven DNS management:
- Web consoles for visual management
- APIs for programmatic access
- Instant propagation (no restart needed)
- Built-in validation
- Global distribution

### Cloudflare's Innovation

Cloudflare added a transformative feature: **the proxy toggle**. DNS records could now be more than just DNS—they could route traffic through Cloudflare's edge network, enabling:

- **Orange Cloud (Proxied)**: Traffic flows through Cloudflare's CDN, WAF, and DDoS protection
- **Grey Cloud (DNS-Only)**: Traditional DNS behavior

This duality means managing Cloudflare DNS records requires understanding not just DNS concepts, but also Cloudflare's edge services.

## Deployment Methods Landscape

### Level 0: Manual (Cloudflare Dashboard)

**Workflow:**
1. Log into Cloudflare Dashboard
2. Navigate to DNS → Records
3. Click "Add record"
4. Fill in type, name, value, TTL, proxy status
5. Save

**Pros:**
- Visual interface, no technical expertise needed
- Immediate feedback and validation
- Quick for one-off changes

**Cons:**
- No version control
- No reproducibility
- No audit trail
- Human error prone (typos, wrong values)
- Doesn't scale beyond a handful of records
- No automation

**Verdict:** Acceptable for initial exploration or emergency fixes. Not suitable for production infrastructure management.

### Level 1: CLI (Cloudflare API via curl)

**Example:**
```bash
# Create an A record
curl -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
     -H "Authorization: Bearer ${API_TOKEN}" \
     -H "Content-Type: application/json" \
     --data '{
       "type": "A",
       "name": "www",
       "content": "192.0.2.1",
       "ttl": 1,
       "proxied": true
     }'

# Update a record
curl -X PUT "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${RECORD_ID}" \
     -H "Authorization: Bearer ${API_TOKEN}" \
     -H "Content-Type: application/json" \
     --data '{
       "type": "A",
       "name": "www",
       "content": "192.0.2.2",
       "ttl": 1,
       "proxied": true
     }'

# Delete a record
curl -X DELETE "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${RECORD_ID}" \
     -H "Authorization: Bearer ${API_TOKEN}"
```

**Pros:**
- Scriptable and automatable
- Can be versioned in scripts
- Works in CI/CD pipelines

**Cons:**
- Imperative (must manage state manually)
- No idempotency (running twice creates duplicates)
- Complex record management (need to track record IDs)
- Error handling is manual
- JSON responses require parsing

**Verdict:** Useful for simple automation scripts. Not recommended for managing multiple records or production infrastructure.

### Level 2: Infrastructure-as-Code (Terraform)

**Example:**
```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

# A Record
resource "cloudflare_dns_record" "www" {
  zone_id = var.zone_id
  name    = "www"
  type    = "A"
  content = "192.0.2.1"
  proxied = true
  ttl     = 1
  comment = "Primary web server"
}

# CNAME Record
resource "cloudflare_dns_record" "api" {
  zone_id = var.zone_id
  name    = "api"
  type    = "CNAME"
  content = "api-lb.example.com"
  proxied = true
  ttl     = 1
}

# MX Record (note: cannot be proxied)
resource "cloudflare_dns_record" "mx_primary" {
  zone_id  = var.zone_id
  name     = "@"
  type     = "MX"
  content  = "mail.example.com"
  ttl      = 1
  priority = 10
}
```

**Pros:**
- Declarative configuration
- State management (tracks what exists)
- Plan before apply (see changes before making them)
- Version control friendly
- Idempotent (safe to run multiple times)
- Large community and ecosystem

**Cons:**
- Requires learning HCL syntax
- State file management complexity
- State locking for team collaboration
- Provider version management
- No type safety (values are strings)
- Verbose for simple records

**Verdict:** Industry standard for DNS IaC. Good choice for teams already using Terraform.

### Level 3: Infrastructure-as-Code (Pulumi)

**Example (Go):**
```go
package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// A Record
		www, err := cloudflare.NewDnsRecord(ctx, "www-record", &cloudflare.DnsRecordArgs{
			ZoneId:  pulumi.String(zoneId),
			Name:    pulumi.String("www"),
			Type:    pulumi.String("A"),
			Content: pulumi.String("192.0.2.1"),
			Proxied: pulumi.Bool(true),
			Ttl:     pulumi.Float64(1),
			Comment: pulumi.String("Primary web server"),
		})
		if err != nil {
			return err
		}

		ctx.Export("recordId", www.ID())
		return nil
	})
}
```

**Pros:**
- Real programming language (Go, Python, TypeScript, etc.)
- Type safety and IDE support
- Testing with standard test frameworks
- Loops, conditionals, functions
- Reusable components
- Same state management benefits as Terraform

**Cons:**
- Steeper learning curve than HCL
- Requires language expertise
- Still verbose for simple records
- Less community adoption than Terraform

**Verdict:** Excellent choice for teams with programming expertise. Preferred when DNS configuration needs complex logic or integration with other services.

### Other Methods

**Ansible:**
```yaml
- name: Create DNS record
  cloudflare_dns:
    zone: example.com
    record: www
    type: A
    value: 192.0.2.1
    proxied: yes
  environment:
    CLOUDFLARE_API_TOKEN: "{{ cloudflare_token }}"
```

Better for configuration management workflows. Limited DNS-specific features.

**Crossplane:**
```yaml
apiVersion: dns.cloudflare.upbound.io/v1alpha1
kind: Record
metadata:
  name: www-record
spec:
  forProvider:
    zoneId: abc123
    name: www
    type: A
    value: 192.0.2.1
    proxied: true
```

Kubernetes-native approach. Good for Kubernetes-centric teams but adds complexity for simple DNS management.

## Comparative Analysis

| Method | Learning Curve | Reproducibility | Team Scale | Automation | Best For |
|--------|---------------|-----------------|------------|------------|----------|
| **Dashboard** | Low | None | Individual | None | Exploration, emergencies |
| **curl/API** | Medium | Poor | Individual | Basic | Simple scripts |
| **Terraform** | Medium | High | Team | Excellent | Standard IaC workflows |
| **Pulumi** | High | High | Team | Excellent | Complex logic, type safety |
| **Ansible** | Medium | Good | Team | Good | Config management |
| **Crossplane** | High | High | Team | Excellent | Kubernetes-native |
| **Planton** | Low | High | Team | Excellent | Unified deployment |

## The Planton Approach

Planton applies the **Kubernetes Resource Model (KRM)** philosophy to DNS records, providing a clean, declarative API that abstracts away the complexity of different IaC tools.

### Design Philosophy

**1. Complete coverage, coherent surface:**
The component covers the full Cloudflare record surface against the provider schema as the floor:
- `zone_id` - Where to create the record
- `name` - The subdomain (or @ for root)
- `type` - any of the 21 Cloudflare record types
- `content` - presentation-format value for simple types
- `data` - typed per-type block for structured types (SRV, CAA, DS, HTTPS, ...)
- `proxied` - Orange cloud or grey cloud
- `ttl` - Time to live
- `priority` - For MX records
- `tags`, `settings` - record tags and proxied-record behavior
- `comment` - Documentation

Rather than a flat bag of every possible attribute, structured record data is modeled as a typed `data` oneof so each record type exposes only its relevant, validated fields.

**2. Unified API:**
```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsRecord
metadata:
  name: www-a-record
spec:
  zone_id: "abc123"
  name: "www"
  type: A
  content: "192.0.2.1"
  proxied: true
```

This manifest works with both Pulumi and Terraform backends—users don't need to know which IaC tool is executing.

**3. Built-in Validation:**
The protobuf schema enforces rules at definition time:
- Required fields (zone_id, name, type)
- Exactly one of `content` or a `data` block, matching the record type
- TTL range validation (0/1 auto, or 30-86400)
- Priority range (0-65535) and priority required for MX
- Cross-field validation (proxied only for A/AAAA/CNAME)

**4. Consistent Outputs:**
Every deployment produces the same outputs:
- `record_id` - Cloudflare record identifier
- `record_name` - The record name as stored by Cloudflare
- `record_type` - The created record type
- `proxied` - Whether traffic is proxied

### Why This Matters

**Before Planton:**
```bash
# Team A uses Terraform
terraform init
terraform plan -var="zone_id=abc" -var="name=www" ...
terraform apply

# Team B uses Pulumi
pulumi stack init
pulumi config set zone_id abc
pulumi up

# Team C uses curl scripts
./create-dns-record.sh www A 192.0.2.1
```

**With Planton:**
```bash
# Everyone uses the same command
planton apply -f record.yaml

# Switch IaC backend without changing manifests
planton apply -f record.yaml --iac terraform
```

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi implementation follows a modular structure:

```
iac/pulumi/
├── main.go           # Entry point, loads stack input
├── Pulumi.yaml       # Project configuration
├── Makefile          # Build and test targets
└── module/
    ├── main.go       # Resource orchestration
    ├── locals.go     # Data transformation
    ├── outputs.go    # Output constants
    └── dns_record.go # Cloudflare record creation
```

**Key implementation details:**
1. Stack input deserialization from environment variable
2. Cloudflare provider setup with credentials
3. Record creation with proper type mapping
4. Output export for downstream consumers

### Terraform Module Architecture

The Terraform implementation maintains feature parity:

```
iac/tf/
├── provider.tf      # Cloudflare provider configuration
├── variables.tf     # Input variables (generated from spec.proto)
├── locals.tf        # Computed values
├── main.tf          # Record resource definition
├── outputs.tf       # Output values
└── README.md        # Module documentation
```

**Key implementation details:**
1. Variables mirror the protobuf spec exactly
2. Locals handle type conversion and defaults
3. Single `cloudflare_dns_record` resource
4. Outputs match `stack_outputs.proto`

## Production Best Practices

### Record Type Selection

| Use Case | Record Type | Proxied | Notes |
|----------|-------------|---------|-------|
| Web server | A/AAAA | Yes | CDN + WAF |
| API endpoint | A/AAAA or CNAME | Yes | Rate limiting |
| Email (MX) | MX | No | Cannot proxy email |
| SPF/DKIM | TXT | No | Email authentication |
| Alias | CNAME | Yes | CNAME flattening at root |
| Certificate control | CAA | No | Restrict CAs |

### TTL Guidelines

| Scenario | Recommended TTL | Reason |
|----------|-----------------|--------|
| Proxied records | 1 (auto) | Cloudflare manages caching |
| Frequently changing | 60-300 | Quick propagation |
| Stable records | 3600-86400 | Reduce DNS queries |
| Pre-migration | 60 | Prepare for fast switch |

### Security Considerations

1. **Use Proxied Records**: Hides origin IP, enables WAF
2. **CAA Records**: Restrict certificate issuance to specific CAs
3. **SPF/DKIM/DMARC**: Complete email authentication chain
4. **API Token Scope**: Use tokens with minimal permissions (Zone:DNS:Edit)

### Common Pitfalls

**1. Proxying MX Records:**
MX records cannot be proxied. Attempting to set `proxied: true` for MX will fail or cause email issues.

**2. Multiple SPF Records:**
Only one SPF TXT record per domain. Multiple records cause validation failures.

**3. Root Domain CNAME:**
Traditional DNS doesn't allow CNAME at root. Cloudflare supports this via CNAME flattening, but other providers may not.

**4. TTL on Proxied Records:**
Proxied records always show TTL as "Auto" in dashboard because Cloudflare manages caching. Your specified TTL only applies to grey-cloud records.

### Monitoring and Alerting

1. **DNS Resolution Monitoring**: Set up external monitors to verify records resolve correctly
2. **Certificate Expiry**: Monitor SSL certificates for proxied records
3. **Email Deliverability**: Monitor SPF/DKIM pass rates
4. **Change Tracking**: Use IaC commit history as audit trail

## Conclusion

DNS record management has evolved from manual zone file editing to sophisticated Infrastructure-as-Code workflows. Cloudflare's unique proxy feature adds another dimension—DNS records become entry points to a full edge security and performance platform.

Planton's CloudflareDnsRecord component offers the optimal balance:
- **Simplicity**: Clean protobuf API with only essential fields
- **Safety**: Built-in validation catches errors before deployment
- **Flexibility**: Works with both Pulumi and Terraform
- **Consistency**: Same manifest format across all environments
- **Production-Ready**: Comprehensive outputs for integration

For teams managing DNS across multiple domains and environments, Planton provides the abstraction layer that makes DNS management predictable, reproducible, and maintainable.

## References

- [Cloudflare DNS Documentation](https://developers.cloudflare.com/dns/)
- [Cloudflare API Reference](https://developers.cloudflare.com/api/)
- [Terraform Cloudflare Provider](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- [Pulumi Cloudflare Provider](https://www.pulumi.com/registry/packages/cloudflare/)
- [RFC 1035 - DNS](https://datatracker.ietf.org/doc/html/rfc1035)
- [RFC 7208 - SPF](https://datatracker.ietf.org/doc/html/rfc7208)
- [RFC 8659 - CAA](https://datatracker.ietf.org/doc/html/rfc8659)

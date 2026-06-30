# DigitalOcean DNS Record: Technical Research Documentation

## Introduction

DNS records are the foundational building blocks of the Domain Name System—they translate human-readable domain names into machine-usable addresses and routing information. Every website, email system, and internet-connected service relies on DNS records to function.

DigitalOcean DNS provides a simple, reliable managed DNS service that integrates seamlessly with DigitalOcean's cloud infrastructure. While it lacks some advanced features of specialized DNS providers (like Cloudflare's proxy capabilities), it offers straightforward DNS management with DigitalOcean's characteristic ease of use.

This document provides comprehensive research into DNS record management on DigitalOcean, examining the deployment landscape from manual operations to Infrastructure-as-Code, and explaining why Planton's approach offers the optimal balance of simplicity and power.

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

Cloud providers introduced API-driven DNS management:
- Web consoles for visual management
- APIs for programmatic access
- Instant propagation (no restart needed)
- Built-in validation
- Global distribution

### DigitalOcean DNS

DigitalOcean's DNS service provides:
- Simple, intuitive API
- Integration with Droplets and other DO resources
- Anycast DNS with global distribution
- Free with any DigitalOcean account
- Support for all major record types

## Deployment Methods Landscape

### Level 0: Manual (DigitalOcean Control Panel)

**Workflow:**
1. Log into DigitalOcean Control Panel
2. Navigate to Networking → Domains
3. Select domain
4. Click "Add record"
5. Fill in type, hostname, value, TTL
6. Save

**Pros:**
- Visual interface, no technical expertise needed
- Immediate feedback and validation
- Quick for one-off changes

**Cons:**
- No version control
- No reproducibility
- No audit trail
- Human error prone
- Doesn't scale

**Verdict:** Acceptable for initial exploration. Not suitable for production infrastructure management.

### Level 1: CLI (doctl)

**Example:**
```bash
# Create an A record
doctl compute domain records create example.com \
  --record-type A \
  --record-name www \
  --record-data 192.0.2.1 \
  --record-ttl 3600

# List records
doctl compute domain records list example.com

# Delete a record
doctl compute domain records delete example.com RECORD_ID
```

**Pros:**
- Scriptable
- Can be version controlled (scripts)
- Faster than UI for bulk operations

**Cons:**
- Still imperative (describes actions, not desired state)
- No drift detection
- Manual sequencing of changes
- Scripts become complex for multi-record setups

**Verdict:** Good for scripting simple tasks. Not ideal for complex infrastructure.

### Level 2: Infrastructure-as-Code (Terraform/Pulumi)

**Terraform Example:**
```hcl
resource "digitalocean_record" "www" {
  domain = "example.com"
  type   = "A"
  name   = "www"
  value  = "192.0.2.1"
  ttl    = 3600
}

resource "digitalocean_record" "mail" {
  domain   = "example.com"
  type     = "MX"
  name     = "@"
  value    = "mail.example.com."
  priority = 10
}
```

**Pulumi Example (Go):**
```go
record, err := digitalocean.NewDnsRecord(ctx, "www", &digitalocean.DnsRecordArgs{
    Domain: pulumi.String("example.com"),
    Type:   pulumi.String("A"),
    Name:   pulumi.String("www"),
    Value:  pulumi.String("192.0.2.1"),
    Ttl:    pulumi.Int(3600),
})
```

**Pros:**
- Declarative (describes desired state)
- Version controlled
- Drift detection
- Planning before apply
- State management
- Modular and reusable

**Cons:**
- Learning curve for IaC tools
- State file management overhead
- Need to understand provider-specific APIs

**Verdict:** Industry standard for production infrastructure.

### Level 3: Planton (Unified API)

**Example:**
```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-record
spec:
  domain: "example.com"
  name: "www"
  type: A
  value: "192.0.2.1"
  ttl_seconds: 3600
```

**Deploy:**
```bash
planton apply -f record.yaml
```

**Advantages over raw IaC:**
- **Schema-first**: Protobuf definitions provide compile-time validation
- **Multi-IaC support**: Same manifest works with Pulumi or Terraform
- **Built-in validation**: Field-level and cross-field validation rules
- **Consistent patterns**: Same structure across all cloud providers
- **Documentation as code**: Self-documenting manifests

## DigitalOcean DNS Record Types

### A Record (Address)
Maps a hostname to an IPv4 address.
```yaml
spec:
  name: "www"
  type: A
  value: "192.0.2.1"
```

### AAAA Record (IPv6 Address)
Maps a hostname to an IPv6 address.
```yaml
spec:
  name: "www"
  type: AAAA
  value: "2001:db8::1"
```

### CNAME Record (Canonical Name)
Creates an alias pointing to another hostname.
```yaml
spec:
  name: "blog"
  type: CNAME
  value: "www.example.com"
```

### MX Record (Mail Exchange)
Routes email to mail servers.
```yaml
spec:
  name: "@"
  type: MX
  value: "mail.example.com"
  priority: 10
```

### TXT Record (Text)
Stores arbitrary text, commonly used for:
- SPF (email authentication)
- DKIM (email signing)
- Domain verification
- DMARC (email policy)

```yaml
spec:
  name: "@"
  type: TXT
  value: "v=spf1 include:_spf.google.com ~all"
```

### SRV Record (Service)
Specifies service location with priority, weight, and port.
```yaml
spec:
  name: "_sip._tcp"
  type: SRV
  value: "sipserver.example.com"
  priority: 10
  weight: 5
  port: 5060
```

### NS Record (Nameserver)
Delegates a subdomain to other nameservers.
```yaml
spec:
  name: "sub"
  type: NS
  value: "ns1.otherprovider.com"
```

### CAA Record (Certificate Authority Authorization)
Controls which CAs can issue certificates.
```yaml
spec:
  name: "@"
  type: CAA
  value: "letsencrypt.org"
  flags: 0
  tag: "issue"
```

## Production Best Practices

### 1. TTL Strategy

| Use Case | Recommended TTL |
|----------|-----------------|
| Stable records | 3600-86400 (1h-1d) |
| Frequently changing | 60-300 (1-5min) |
| Pre-migration | Lower TTL before changes |
| Post-migration | Raise TTL after stable |

### 2. Email Authentication Stack

For proper email deliverability, implement:
1. **SPF**: Authorize sending servers
2. **DKIM**: Sign outgoing messages
3. **DMARC**: Define policy for failures

### 3. Security Considerations

- Use CAA records to restrict certificate issuance
- Implement DNSSEC if available
- Monitor for unauthorized DNS changes
- Use separate API tokens with minimal permissions

### 4. High Availability

- Use multiple records for round-robin load balancing
- Implement health checks at the application level
- Consider using DigitalOcean Load Balancers with DNS

## Why Planton?

### 80/20 Principle

Planton's DigitalOceanDnsRecord component exposes the 20% of configuration that covers 80% of use cases:

**Included:**
- Record type (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA)
- Name (hostname/subdomain)
- Value (IP, hostname, text)
- TTL
- Priority, weight, port (for MX/SRV)
- Flags, tag (for CAA)

**Not Included (advanced/rare):**
- DNSSEC configuration
- Secondary DNS
- Bulk record operations

### Validation

Built-in protobuf validation ensures:
- Required fields are present
- TTL is within valid range (30-86400)
- Priority, weight, port are valid (0-65535)
- CAA flags are valid (0-255)
- Port is required for SRV records
- Tag is required for CAA records

### Multi-IaC Support

The same manifest works with:
- **Pulumi**: `planton pulumi up -f record.yaml`
- **Terraform**: `planton tofu apply -f record.yaml`

## Conclusion

DNS record management has evolved from hand-editing zone files to declarative Infrastructure-as-Code. DigitalOcean DNS provides a straightforward, reliable service that integrates well with the broader DigitalOcean ecosystem.

Planton's DigitalOceanDnsRecord component brings the best of IaC to DNS management:
- **Simplicity**: Clean YAML manifests
- **Validation**: Schema-enforced correctness
- **Flexibility**: Choose your IaC tool
- **Consistency**: Same patterns across all providers

For teams already using DigitalOcean infrastructure, this component provides seamless DNS management that fits into existing workflows and toolchains.

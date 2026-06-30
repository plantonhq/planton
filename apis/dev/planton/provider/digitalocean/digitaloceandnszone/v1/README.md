# DigitalOcean DNS Zone

## Overview

The **DigitalOcean DNS Zone API Resource** provides a consistent, declarative interface for creating and managing DNS zones (domains) on DigitalOcean. This resource simplifies DNS management by enabling Infrastructure as Code (IaC) workflows for domain configuration, allowing you to version control your DNS records alongside your application infrastructure.

## Purpose

We developed this API resource to streamline DNS management on DigitalOcean. By offering a protobuf-based, declarative interface, it reduces the complexity of DNS provisioning and enables teams to:

- **Manage DNS as Code**: Define DNS zones and records in YAML manifests that can be version-controlled in Git
- **Simplify Zone Creation**: Provision DigitalOcean domains with their initial DNS records in a single atomic operation
- **Standardize Configuration**: Use a consistent API that aligns with other Planton components
- **Enable Automation**: Integrate DNS management into CI/CD pipelines with declarative manifests
- **Cross-Reference Resources**: Reference values from other infrastructure resources using `StringValueOrRef`

## Key Features

- **Declarative DNS Management**: Define DNS zones and records in YAML with full validation and type safety
- **Comprehensive Record Type Support**: Supports A, AAAA, CNAME, MX, TXT, SRV, CAA, and NS records
- **Advanced Record Configuration**: Full support for MX priorities, SRV weight/port, and CAA flags/tags
- **Infrastructure Integration**: Seamlessly integrates with DigitalOcean Droplets, Load Balancers, and Spaces
- **Dual IaC Backend**: Deploy using either Pulumi (Go) or Terraform with identical specifications
- **Flexible Value References**: Use literal values or cross-resource references for dynamic DNS configuration

## Use Cases

### Simple Website Deployment

Create a DNS zone for a website with apex and www records pointing to a DigitalOcean Droplet or Load Balancer.

### Email Configuration

Configure MX records for email services (Google Workspace, Office 365) along with SPF, DKIM, and DMARC TXT records for email authentication.

### Multi-Environment Applications

Manage DNS for development, staging, and production environments with environment-specific subdomains.

### CDN and Static Assets

Point subdomains to DigitalOcean Spaces CDN endpoints for optimized static asset delivery.

### Certificate Authority Authorization

Use CAA records to authorize specific certificate authorities (like Let's Encrypt) for enhanced security.

## Important Considerations

### DNS Delegation

To use DigitalOcean DNS for a domain:

1. Register the domain with a third-party registrar (Namecheap, Google Domains, etc.)
2. Create the DNS zone using this API resource
3. Update the domain's nameservers at the registrar to:
   - `ns1.digitalocean.com`
   - `ns2.digitalocean.com`
   - `ns3.digitalocean.com`

**Critical**: If your domain has DNSSEC enabled, you **must disable it** at the registrar before delegating to DigitalOcean. DigitalOcean DNS does not support DNSSEC, and failure to disable it will cause DNS resolution failures.

### Platform Limitations

DigitalOcean DNS is designed for simplicity, not advanced features:

- **No DNSSEC**: DNSSEC is not supported. For domains requiring DNSSEC, use Cloudflare or AWS Route 53 instead
- **No Zone Import**: There is no API to import BIND-style zone files. Migrations require recreating all records via the API
- **No Traffic Routing**: Geo-routing, weighted routing, latency-based routing, and failover policies are not available
- **API Rate Limits**: The API is limited to 250 requests per minute, which may impact automation at scale

### When to Choose DigitalOcean DNS

DigitalOcean DNS is ideal when:

- Your infrastructure is primarily hosted on DigitalOcean
- You value simplicity and cost predictability (DigitalOcean DNS is currently free)
- You don't require advanced DNS features like DNSSEC or traffic routing
- You want tight integration with Droplets, Load Balancers, and Spaces

For teams requiring DNSSEC, global Anycast performance, or advanced routing, consider using Cloudflare or AWS Route 53 instead. A hybrid approach (Cloudflare for DNS, DigitalOcean for compute) is also common.

## Getting Started

See the `examples.md` file in this directory for practical configuration examples, including:

- Simple website (apex + www)
- Email configuration (MX + SPF + DMARC)
- Load balancer integration
- CDN endpoints
- CAA records for Let's Encrypt

For detailed implementation guidance, refer to:

- `docs/README.md` - Comprehensive research and design decisions
- `iac/pulumi/README.md` - Pulumi-specific usage instructions
- `iac/tf/README.md` - Terraform-specific usage instructions

## DNS Record Types Reference

This API resource supports the following DNS record types:

| Type | Purpose | Example Use Case |
|------|---------|-----------------|
| A | IPv4 address mapping | Point domain to Droplet IP |
| AAAA | IPv6 address mapping | IPv6-enabled services |
| CNAME | Canonical name (alias) | www → apex, or subdomain → Load Balancer |
| MX | Mail exchange | Google Workspace, Office 365 |
| TXT | Text records | SPF, DKIM, DMARC, domain verification |
| SRV | Service locator | SIP, XMPP, Minecraft servers |
| CAA | Certificate authority authorization | Restrict certificate issuance to Let's Encrypt |
| NS | Nameserver delegation | Delegate subdomain to another DNS provider |

## Production Best Practices

### TTL Strategy

- **Standard (3600s)**: Default for most records
- **Low TTL (300s)**: Use during migrations or cutover events for fast rollback
- **High TTL (86400s)**: Use for static records like CDN CNAMEs or SPF records

Lower TTLs *before* making critical changes, then raise them after confirming success.

### Monitoring

DigitalOcean's native monitoring doesn't cover DNS health. Use external monitoring services like:

- **UptimeRobot**: Track specific DNS records and alert on changes
- **Pingdom**: Monitor DNS resolution time and availability
- **DNSPerf**: Benchmark global DNS propagation

### Backup and Recovery

Your YAML manifests *are* your backup. Store them in Git for version control and disaster recovery. If you need to recreate DNS zones in a new account, simply apply the manifests with the Planton CLI.

For manual backups, use DigitalOcean's control panel to export a BIND-style zone file (note: restore is manual as there's no import function).

## Next Steps

1. Review `examples.md` for practical DNS zone configurations
2. Read `docs/README.md` for in-depth platform analysis and design rationale
3. Choose your deployment method:
   - **Pulumi**: See `iac/pulumi/README.md`
   - **Terraform**: See `iac/tf/README.md`
4. Deploy a test domain and verify delegation at your registrar
5. Configure monitoring for production domains

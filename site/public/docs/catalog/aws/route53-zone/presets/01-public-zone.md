---
title: "Public DNS Zone"
description: "This preset creates a public Route53 hosted zone for managing DNS records that resolve globally on the internet. Public zones are the most common type and are used for any domain that needs to be..."
type: "preset"
rank: "01"
presetSlug: "01-public-zone"
componentSlug: "route53-zone"
componentTitle: "Route53 Zone"
provider: "aws"
icon: "package"
order: 1
---

# Public DNS Zone

This preset creates a public Route53 hosted zone for managing DNS records that resolve globally on the internet. Public zones are the most common type and are used for any domain that needs to be reachable from the internet. DNS records (A, CNAME, MX, TXT, etc.) are managed separately via `AwsRoute53DnsRecord` for composability.

## When to Use

- Hosting DNS for an internet-facing domain (e.g., `example.com`)
- Any domain that needs publicly resolvable DNS records
- Foundation for ACM certificate validation, ALB DNS aliases, and other AWS integrations

## Key Configuration Choices

- **Public zone** (`isPrivate: false`) -- DNS records resolve globally; anyone on the internet can query this zone
- **No inline records** -- DNS records are managed as standalone `AwsRoute53DnsRecord` resources for better composability and independent lifecycle management
- **No DNSSEC** -- DNSSEC requires additional configuration at the domain registrar; enable separately if needed for your compliance requirements
- **No query logging** -- Query logging is disabled by default; enable for debugging or security monitoring of high-traffic domains

## Placeholders to Replace

This preset has no placeholders. The zone name is derived from `metadata.name`. After deployment, use the zone ID from status outputs to create DNS records via `AwsRoute53DnsRecord`.

## Related Presets

- **02-private-vpc-zone** -- Use instead for split-horizon DNS that resolves only within specific VPCs

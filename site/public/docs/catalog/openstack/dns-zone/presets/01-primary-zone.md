---
title: "Primary DNS Zone"
description: "This preset creates a primary DNS zone in Designate. The zone is authoritative for the specified domain and can host A, AAAA, CNAME, MX, TXT, and other record types. Records can be added inline via..."
type: "preset"
rank: "01"
presetSlug: "01-primary-zone"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "openstack"
icon: "package"
order: 1
---

# Primary DNS Zone

This preset creates a primary DNS zone in Designate. The zone is authoritative for the specified domain and can host A, AAAA, CNAME, MX, TXT, and other record types. Records can be added inline via the `records` field or as standalone `OpenStackDnsRecord` resources.

## When to Use

- Hosting DNS for a domain managed by the OpenStack deployment's Designate service
- Any workload that needs DNS records resolvable from within (or outside) the OpenStack environment
- The first step before creating DNS records for instances, load balancers, or services

## Key Configuration Choices

- **Primary zone** -- this is the authoritative zone (not a secondary/slave zone)
- **1-hour TTL** (`ttl: 3600`) -- default TTL applied to records that do not specify their own; adjust based on change frequency
- **Trailing dot** -- `domainName` must end with a dot (DNS convention for fully qualified domain names)
- **SOA email** -- administrative contact for the zone's SOA record

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-domain.com.>` | Fully qualified domain name with trailing dot (e.g., `example.com.`) | Your domain registrar |
| `admin@<your-domain.com>` | SOA admin email for the zone | Your team's DNS contact |

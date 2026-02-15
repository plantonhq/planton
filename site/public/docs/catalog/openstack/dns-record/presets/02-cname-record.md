---
title: "CNAME Record"
description: "This preset creates a standalone DNS CNAME record in Designate, aliasing one hostname to another. The target can be within the same zone or an external domain. CNAME records are commonly used for..."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "openstack"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a standalone DNS CNAME record in Designate, aliasing one hostname to another. The target can be within the same zone or an external domain. CNAME records are commonly used for vanity names, service aliases, and load balancer endpoints.

## When to Use

- Creating a friendly alias (e.g., `www.example.com` -> `app.example.com`)
- Pointing to an external service endpoint (e.g., `cdn.example.com` -> `d1234.cloudfront.net`)
- Any hostname that should resolve to another hostname rather than a direct IP

## Key Configuration Choices

- **CNAME record** (`type: CNAME`) -- aliases one hostname to another
- **5-minute TTL** (`ttl: 300`) -- moderate caching
- **Trailing dots** -- both `recordName` and target value must be FQDNs with trailing dots
- **Single value** -- CNAME records must have exactly one target value

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<zone-id>` | ID of the DNS zone this record belongs to | OpenStack console or `OpenStackDnsZone` status outputs |
| `<alias.your-domain.com.>` | The alias hostname with trailing dot | Your DNS naming convention |
| `<target.your-domain.com.>` | The target hostname with trailing dot | The canonical name of the service |

## Related Presets

- **01-a-record** -- Use instead when mapping a hostname directly to an IP address

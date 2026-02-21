---
title: "A Record"
description: "This preset creates a DNS A record that maps a subdomain to an IPv4 address. A records are the most common DNS record type, used to point domain names at web servers, application servers, or any host..."
type: "preset"
rank: "01"
presetSlug: "01-a-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "alicloud"
icon: "package"
order: 1
---

# A Record

This preset creates a DNS A record that maps a subdomain to an IPv4 address. A records are the most common DNS record type, used to point domain names at web servers, application servers, or any host with a static IP.

## When to Use

- Pointing a subdomain (e.g., `www`, `api`, `app`) to a server's IPv4 address
- Setting up apex domain resolution (use `@` as the host record)
- Any scenario where you need a domain name to resolve to a specific IP address

## Key Configuration Choices

- **Minimal fields** -- only the required fields are specified, keeping the manifest simple
- **Default TTL** -- uses the provider default of 600 seconds (10 minutes). Increase for stable records, decrease for records that change frequently.
- **Default status** -- ENABLE (the record is immediately active)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`). Alidns is global, but the provider requires a region. | Your deployment region strategy |
| `<your-domain-name>` | The parent domain (e.g., `example.com`). Must already exist in Alidns. | Your AliCloudDnsZone resource or Alidns console |
| `<host-record>` | Subdomain label (e.g., `www`, `api`, `@` for apex, `*` for wildcard) | Your DNS design |
| `<ipv4-address>` | Target IPv4 address (e.g., `203.0.113.10`) | Your server or load balancer IP |

## Post-Deployment Steps

1. Deploy the manifest to create the DNS record
2. DNS propagation depends on the TTL of any existing record for this name
3. Verify with `dig <host-record>.<your-domain-name>` or `nslookup`

## Related Presets

- **02-cname-record** -- use instead to alias a subdomain to another domain name
- **03-mx-record** -- use for email routing with priority

---
title: "Root DNS Zone"
description: "This preset creates a Scaleway DNS zone for a root domain with no inline records. Records are managed separately as standalone `ScalewayDnsRecord` resources, which is the recommended pattern for..."
type: "preset"
rank: "01"
presetSlug: "01-root-zone"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "scaleway"
icon: "package"
order: 1
---

# Root DNS Zone

This preset creates a Scaleway DNS zone for a root domain with no inline records. Records are managed separately as standalone `ScalewayDnsRecord` resources, which is the recommended pattern for production because it creates explicit dependency edges in infra chart DAGs and allows records to reference outputs from other resources.

## When to Use

- Setting up DNS management for a domain on Scaleway
- Production environments where DNS records are created independently (e.g., A records pointing to a Load Balancer, CNAMEs to a Kapsule endpoint)
- Any domain where records will reference other infrastructure resources via `valueFrom`

## Key Configuration Choices

- **Root zone** (no `subdomain`) -- manages the apex of the domain (e.g., `example.com`); set `subdomain` to a value like `staging` for delegated subzones
- **No inline records** -- records are managed as standalone `ScalewayDnsRecord` resources for better DAG visibility and cross-resource wiring

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-domain.com>` | The domain name to create a zone for (e.g., `example.com`) | Your domain registrar |

## Related Presets

After creating the zone, configure the nameservers from `status.outputs.name_servers` at your domain registrar to delegate DNS resolution to Scaleway.

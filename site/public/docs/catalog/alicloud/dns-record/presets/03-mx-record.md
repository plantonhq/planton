---
title: "MX Record"
description: "This preset creates a DNS MX (Mail Exchange) record that routes email for a domain to a mail server. MX records use a priority value to determine the order in which mail servers are tried."
type: "preset"
rank: "03"
presetSlug: "03-mx-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "alicloud"
icon: "package"
order: 3
---

# MX Record

This preset creates a DNS MX (Mail Exchange) record that routes email for a domain to a mail server. MX records use a priority value to determine the order in which mail servers are tried.

## When to Use

- Setting up email delivery for a domain (e.g., routing mail to your organization's mail server)
- Configuring multiple mail servers with failover (create multiple MX records with different priorities)
- Integrating with third-party email services (e.g., Alibaba Cloud Enterprise Mail, Google Workspace, Microsoft 365)

## Key Configuration Choices

- **`rr: "@"`** -- MX records are typically set at the apex (bare domain) level. Use `@` so that mail to `user@example.com` is routed correctly.
- **`priority`** -- required for MX records. Range 1 (highest priority) to 10 (lowest). Mail is delivered to the lowest-numbered server first; higher numbers are fallbacks.
- **`ttl: 3600`** -- mail routing changes are infrequent, so a 1-hour TTL balances cacheability with change propagation speed.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region strategy |
| `<your-domain-name>` | The parent domain (e.g., `example.com`). Must already exist in Alidns. | Your AliCloudDnsZone resource or Alidns console |
| `<mail-server-domain>` | Mail server hostname (e.g., `mx1.example.com`, `aspmx.l.google.com`) | Your email provider documentation |
| `<priority-1-to-10>` | MX priority: 1 = highest, 10 = lowest. Use 5 for a single server, or 1/5/10 for primary/secondary/tertiary. | Your email architecture |

## Post-Deployment Steps

1. Deploy the manifest to create the MX record
2. For multiple mail servers, create additional MX records with different priorities
3. Add a TXT record for SPF to prevent email spoofing: `"v=spf1 include:example.com ~all"`
4. Verify with `dig MX <your-domain-name>`

## Related Presets

- **01-a-record** -- use for web server DNS records
- **02-cname-record** -- use for domain aliasing

---
title: "Website + Email DNS Zone"
description: "This preset creates a DNS zone for a domain that serves both a website and receives email. Includes an apex A record, www CNAME, MX record for mail routing, and a TXT record for SPF email..."
type: "preset"
rank: "02"
presetSlug: "02-with-email"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "civo"
icon: "package"
order: 2
---

# Website + Email DNS Zone

This preset creates a DNS zone for a domain that serves both a website and receives email. Includes an apex A record, www CNAME, MX record for mail routing, and a TXT record for SPF email authentication. Covers the majority of small business and SaaS domain configurations.

## When to Use

- Domains that host both a website and receive email (Google Workspace, Microsoft 365, self-hosted mail)
- Standard business domain setup with web + email
- Domains that need SPF records to prevent email spoofing

## Key Configuration Choices

- **Apex A + www CNAME** -- standard website DNS (same as 01-simple-website)
- **MX record** (`type: MX`) -- routes email to your mail server or email provider
- **SPF TXT record** (`type: TXT`) -- authenticates which servers can send email for your domain; prevents spoofing
- **1-hour TTL** (`ttlSeconds: 3600`) -- recommended default for all records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `example.com` | Your domain name (FQDN) | Your domain registrar |
| `<server-ipv4-address>` | Public IP of your web server | `CivoComputeInstance` or `CivoIpAddress` status outputs |
| `<mail-server.example.com>` | Mail server hostname | Your email provider's MX setup guide |
| `<your-email-provider>` | SPF include domain (e.g., `_spf.google.com`) | Your email provider's SPF setup guide |

## Related Presets

- **01-simple-website** -- Use instead when the domain only needs web DNS without email

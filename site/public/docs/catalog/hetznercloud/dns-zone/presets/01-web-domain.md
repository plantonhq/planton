---
title: "Web Domain"
description: "This preset creates a primary DNS zone on Hetzner Cloud's authoritative nameservers for a production website domain. It includes the standard record set a real website needs: an apex A record, a www..."
type: "preset"
rank: "01"
presetSlug: "01-web-domain"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Web Domain

This preset creates a primary DNS zone on Hetzner Cloud's authoritative nameservers for a production website domain. It includes the standard record set a real website needs: an apex A record, a www CNAME, an MX record for email delivery, SPF and DMARC TXT records for email authentication, and a CAA record restricting certificate issuance to Let's Encrypt. Delete protection is enabled to prevent accidental removal.

After provisioning, point your domain's NS records at the Hetzner nameservers returned in `status.outputs.nameservers` (via your domain registrar) to activate the zone.

## When to Use

- Hosting a production website or web application on Hetzner Cloud servers
- Any domain that needs both web traffic resolution and email authentication records
- Domains that will use a HetznerCloudCertificate (managed Let's Encrypt) for HTTPS -- the CAA record pre-authorizes Let's Encrypt issuance

## Key Configuration Choices

- **Primary mode** (`mode: primary`) -- Hetzner Cloud is the authoritative nameserver; all records are managed directly through this manifest
- **Explicit default TTL** (`ttl: 3600`) -- one-hour TTL matches the spec default; lower to 300 during migrations, raise to 86400 for stable records
- **Delete protection enabled** (`deleteProtection: true`) -- prevents accidental deletion of a production zone through the API or console
- **Apex A record** (`@` / `A`) -- points the bare domain to your server's IPv4 address; add a second entry for round-robin if you have multiple servers
- **WWW CNAME** (`www` / `CNAME`) -- aliases `www.<your-domain>` to the apex domain so both resolve to the same server
- **MX record** (`@` / `MX` with `ttl: 3600`) -- routes email for the domain to your mail server; the priority prefix (`10`) follows standard MX convention
- **SPF record** (`@` / `TXT`) -- declares which mail servers are authorized to send email for your domain; prevents spoofing
- **DMARC record** (`_dmarc` / `TXT` with `p=reject`) -- instructs receiving mail servers to reject unauthenticated email; the strictest policy for production domains
- **CAA record** (`@` / `CAA`) -- restricts TLS certificate issuance to Let's Encrypt only; aligns with the `01-managed-lets-encrypt` HetznerCloudCertificate preset

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-domain>` | The DNS domain name (e.g., `example.com`) | Your domain registrar |
| `<server-ipv4-address>` | Public IPv4 address of your web server | The `status.outputs.ipv4_address` of your HetznerCloudServer resource, or the Hetzner Cloud Console |
| `<mail-server-hostname>` | Hostname of your mail server (e.g., `mail.example.com`) | Your email provider's setup instructions |
| `<spf-include-domain>` | SPF include domain for your email provider (e.g., `_spf.google.com` for Google Workspace, `spf.protection.outlook.com` for Microsoft 365) | Your email provider's DNS setup guide |

## Related Presets

- **02-secondary-zone** -- use instead when Hetzner Cloud should act as a secondary nameserver syncing from an external primary
- **03-simple-zone** -- use instead when you only need basic A/CNAME records without email or security records

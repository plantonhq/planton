---
title: "Mail Failover IPv4 with Reverse DNS"
description: "This preset allocates an IPv4 Floating IP with a reverse DNS (rDNS) record, assigns it to a server, and enables delete protection. It is designed for mail servers that need both failover capability..."
type: "preset"
rank: "03"
presetSlug: "03-mail-failover-ipv4"
componentSlug: "hetzner-cloud-floating-ip"
componentTitle: "Hetzner Cloud Floating IP"
provider: "hetznercloud"
icon: "package"
order: 3
---

# Mail Failover IPv4 with Reverse DNS

This preset allocates an IPv4 Floating IP with a reverse DNS (rDNS) record, assigns it to a server, and enables delete protection. It is designed for mail servers that need both failover capability and verifiable rDNS. The IaC module creates an `hcloud_floating_ip` with a server assignment and an `hcloud_rdns` resource linking the PTR record to the allocated IP address.

Mail deliverability depends on matching forward and reverse DNS. Receiving mail servers reject messages from IPs whose rDNS does not resolve to the sending domain. A Floating IP adds failover resilience on top of this: if the primary mail server goes down, keepalived (or equivalent) reassigns the IP -- and its rDNS -- to the standby server without any DNS propagation delay. Losing a mail server's IP means losing months of accumulated sender reputation, so delete protection is warranted.

## When to Use

- Mail servers (Postfix, Exim, Exchange) in an active/standby failover pair where SPF, DKIM, and DMARC verification require matching forward and reverse DNS
- SMTP relay or transactional email services that need both a clean rDNS-verified IP and failover resilience
- Any service where clients perform reverse DNS lookups for identity verification and downtime from server failure is unacceptable

## Key Configuration Choices

- **IPv4** (`type: ipv4`) -- mail delivery overwhelmingly relies on IPv4; most spam blocklists and reputation systems track IPv4 addresses
- **Reverse DNS** (`dnsPtr: <mail-server-hostname>`) -- sets the PTR record for the allocated IP; must match the hostname in your MX record and the server's HELO/EHLO greeting for reliable delivery
- **Server assignment** (`serverId`) -- attaches the IP to the primary mail server at creation time; keepalived or equivalent handles reassignment on failover
- **Delete protection** (`deleteProtection: true`) -- prevents accidental destruction of an IP whose sender reputation has been built over time; must be explicitly disabled before the resource can be removed
- **Falkenstein location** (`homeLocation: fsn1`) -- Hetzner's largest datacenter; change to match the location of your mail server pair since Floating IPs can only be assigned to servers in the same location

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<hetzner-server-id>` | Numeric ID of the primary mail server to assign this IP to (as a string) | Hetzner Cloud Console server details page, or `HetznerCloudServer` resource outputs (`status.outputs.server_id`) |
| `<mail-server-hostname>` | Fully qualified hostname the IP should reverse-resolve to (e.g., `mail.example.com`) | Your DNS zone's MX record target; must also match the server's HELO/EHLO identity |

## Related Presets

- **01-reserved-ipv4** -- minimal variant for reserving an IP before any infrastructure is in place
- **02-failover-ipv4** -- production failover without rDNS, for services that do not require reverse DNS verification

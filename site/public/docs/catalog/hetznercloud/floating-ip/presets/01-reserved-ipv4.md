---
title: "Reserved IPv4 Floating IP"
description: "This preset allocates an unassigned IPv4 Floating IP in Hetzner Cloud. The IP is reserved but not attached to any server, making it available for later assignment when your failover infrastructure..."
type: "preset"
rank: "01"
presetSlug: "01-reserved-ipv4"
componentSlug: "floating-ip"
componentTitle: "Floating IP"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Reserved IPv4 Floating IP

This preset allocates an unassigned IPv4 Floating IP in Hetzner Cloud. The IP is reserved but not attached to any server, making it available for later assignment when your failover infrastructure (keepalived, CARP, Pacemaker) is ready. The IaC module creates an `hcloud_floating_ip` resource with no server assignment and no reverse DNS.

Unlike a Primary IP, a Floating IP is not auto-configured on the server's network interface. Once assigned, you must configure the IP on the target server (typically via cloud-init or a keepalived VIP). This decoupling is what enables seamless failover between servers.

## When to Use

- Reserving a stable public IPv4 address before your server infrastructure is provisioned
- Planning a failover setup where the Floating IP will be assigned to the active server later
- Allocating IPs in advance for capacity planning or IP allowlisting with external services

## Key Configuration Choices

- **IPv4** (`type: ipv4`) -- allocates a single public address; the most common choice since most failover scenarios involve IPv4 endpoints
- **Falkenstein location** (`homeLocation: fsn1`) -- Hetzner's largest datacenter in the eu-central network zone; change to `nbg1` (Nuremberg), `hel1` (Helsinki), `ash` (Ashburn), `hil` (Hillsboro), or `sin` (Singapore) to match your target servers
- **No server assignment** -- the IP is created unassigned; add `serverId` later when your failover mechanism is configured
- **No delete protection** -- allows easy teardown during development; set `deleteProtection: true` before promoting to production

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired resource name.

## Related Presets

- **02-failover-ipv4** -- assigns the Floating IP to a server with delete protection, for production failover setups
- **03-mail-failover-ipv4** -- adds reverse DNS and server assignment for mail servers requiring rDNS verification

---
title: "Failover IPv4 Floating IP"
description: "This preset allocates an IPv4 Floating IP and assigns it to a server, with delete protection enabled. It represents the standard production failover configuration where a stable public endpoint must..."
type: "preset"
rank: "02"
presetSlug: "02-failover-ipv4"
componentSlug: "floating-ip"
componentTitle: "Floating IP"
provider: "hetznercloud"
icon: "package"
order: 2
---

# Failover IPv4 Floating IP

This preset allocates an IPv4 Floating IP and assigns it to a server, with delete protection enabled. It represents the standard production failover configuration where a stable public endpoint must survive server replacement. The IaC module creates an `hcloud_floating_ip` resource with a server assignment via `hcloud_floating_ip_assignment`.

In a typical failover setup, two or more servers in the same location run keepalived (or CARP/Pacemaker). The Floating IP starts assigned to the primary server. When keepalived detects the primary is down, it reassigns the Floating IP to the standby server via the Hetzner Cloud API, restoring service without a DNS change or client-side retry.

## When to Use

- Active/standby server pairs where a single public IPv4 endpoint must survive primary server failure
- Services behind a Floating IP managed by keepalived, CARP, Pacemaker, or a custom failover script
- Any production workload where losing the public IP address during maintenance or failure is unacceptable

## Key Configuration Choices

- **IPv4** (`type: ipv4`) -- allocates a single public address; the standard choice for failover endpoints that clients reach directly
- **Server assignment** (`serverId`) -- attaches the IP to a server at creation time; the server must be in the same location as `homeLocation`
- **Delete protection** (`deleteProtection: true`) -- prevents accidental destruction of the failover IP during infrastructure teardown; must be explicitly disabled before the resource can be removed
- **Falkenstein location** (`homeLocation: fsn1`) -- Hetzner's largest datacenter; change to match the location of your server pair since Floating IPs can only be assigned to servers in the same location
- **No rDNS** -- omitted because most failover endpoints do not require reverse DNS; add `dnsPtr` if your use case needs it (see the `03-mail-failover-ipv4` preset)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<hetzner-server-id>` | Numeric ID of the Hetzner Cloud server to assign this IP to (as a string) | Hetzner Cloud Console server details page, or `HetznerCloudServer` resource outputs (`status.outputs.server_id`) |

## Related Presets

- **01-reserved-ipv4** -- simpler variant without server assignment or delete protection, for reserving an IP before the failover infrastructure is ready
- **03-mail-failover-ipv4** -- adds reverse DNS for mail servers that need both failover capability and rDNS verification

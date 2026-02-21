---
title: "Production Web Server"
description: "This preset creates a production-hardened Hetzner Cloud server for web workloads. It combines a firewall for inbound traffic control, a private network attachment for backend communication, automatic..."
type: "preset"
rank: "02"
presetSlug: "02-production-web"
componentSlug: "server"
componentTitle: "Server"
provider: "hetznercloud"
icon: "package"
order: 2
---

# Production Web Server

This preset creates a production-hardened Hetzner Cloud server for web workloads. It combines a firewall for inbound traffic control, a private network attachment for backend communication, automatic daily backups, and all available protection flags. The cloud-init script applies OS-level security updates on first boot and enables unattended upgrades for ongoing patch management.

The cx32 server type (4 shared vCPU, 8 GB RAM, 80 GB NVMe) is right-sized for moderate web traffic. Scale up to `cx42` or `cx52` for heavier workloads, or switch to `ccx` (dedicated vCPU) if you need guaranteed CPU performance.

## When to Use

- Single-server production web applications, API backends, or reverse proxies
- Servers that need both public reachability (for HTTP/HTTPS traffic) and private connectivity (for database or cache access)
- Any workload where accidental deletion, unplanned rebuilds, or data loss would cause an outage

## Key Configuration Choices

- **Firewall attached** (`firewallIds`) -- restricts inbound traffic to only the ports you explicitly allow; pair with the HetznerCloudFirewall `01-web-server` preset for SSH + HTTP + HTTPS + ICMP
- **Private network** (`networks`) -- attaches the server to a Hetzner Cloud network for private communication with databases, caches, and other backend services; Hetzner auto-assigns an IP from the subnet range
- **Daily backups enabled** (`backups: true`) -- Hetzner retains 14 days of daily server images at a 20% surcharge on the server price; this is the simplest disaster recovery option available
- **Delete protection** (`deleteProtection: true`) -- prevents accidental destruction via API or Console; must be explicitly disabled before the server can be removed
- **Rebuild protection** (`rebuildProtection: true`) -- prevents accidental re-imaging of the server, which would wipe all data on the local disk
- **Graceful shutdown** (`shutdownBeforeDeletion: true`) -- sends an ACPI shutdown signal and waits for the OS to power off cleanly before Terraform destroys the server, reducing the risk of filesystem corruption
- **Cloud-init hardening** (`userData`) -- runs `apt-get upgrade` and installs `unattended-upgrades` on first boot to close known CVEs immediately; replace or extend with your own provisioning script
- **Auto-assigned public IPs** -- `publicNet` is omitted so the server gets ephemeral public IPv4 and IPv6; for stable IPs that survive server replacement, allocate a HetznerCloudPrimaryIp and wire it via `publicNet.ipv4`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<ssh-key-name>` | Name or numeric ID of an SSH key registered in Hetzner Cloud | The `metadata.name` of your HetznerCloudSshKey resource, or the SSH Keys page in the Hetzner Cloud Console |
| `<firewall-id>` | Numeric ID of a Hetzner Cloud firewall | The `status.outputs.firewall_id` of your HetznerCloudFirewall resource, or the Firewalls page in the Hetzner Cloud Console |
| `<network-id>` | Numeric ID of a Hetzner Cloud network | The `status.outputs.network_id` of your HetznerCloudNetwork resource, or the Networks page in the Hetzner Cloud Console |

## Related Presets

- **01-quick-start** -- minimal server without production hardening, for development and experimentation
- **03-private-backend** -- disables public networking entirely for servers that must not be internet-facing

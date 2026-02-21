# Quick-Start Server

This preset creates the simplest possible Hetzner Cloud server: a shared-vCPU instance running Ubuntu with SSH access and auto-assigned public IPv4 and IPv6 addresses. It provisions a single `hcloud_server` resource with no firewalls, private networks, or production hardening -- just a working VM you can SSH into immediately.

The cx22 server type (2 shared vCPU, 4 GB RAM, 40 GB NVMe) is Hetzner's cheapest x86 option. For ARM workloads, substitute `cax11` (2 Ampere vCPU, 4 GB RAM). For dedicated vCPUs, use the `ccx` or `cpx` families.

## When to Use

- Getting started with Hetzner Cloud and OpenMCF for the first time
- Development, experimentation, or throwaway environments where speed matters more than security
- Learning how HetznerCloudServer manifests work before layering on production features

## Key Configuration Choices

- **Shared vCPU** (`serverType: cx22`) -- the lowest-cost x86 server type; sufficient for light workloads, CI runners, and development
- **Ubuntu 24.04 LTS** (`image: ubuntu-24.04`) -- the most widely supported Linux distribution on Hetzner Cloud; change to `debian-12`, `rocky-9`, or a snapshot ID for other OS requirements
- **Falkenstein location** (`location: fsn1`) -- Hetzner's largest datacenter in the eu-central zone; change to `nbg1` (Nuremberg), `hel1` (Helsinki), `ash` (Ashburn), `hil` (Hillsboro), or `sin` (Singapore) to match your latency needs
- **Auto-assigned public IPs** -- `publicNet` is omitted, so the server receives ephemeral public IPv4 and IPv6 addresses automatically; these change if the server is deleted and recreated
- **No firewall** -- all inbound ports are open; apply a HetznerCloudFirewall (see the `01-web-server` or `02-ssh-only` firewall presets) before exposing any service to the internet
- **No backups or protections** -- keeps the preset minimal; see the `02-production-web` server preset for the full production configuration

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<ssh-key-name>` | Name or numeric ID of an SSH key registered in Hetzner Cloud | The `metadata.name` of your HetznerCloudSshKey resource, or the SSH Keys page in the Hetzner Cloud Console |

## Related Presets

- **02-production-web** -- adds firewall, private network, backups, and protection flags for production web servers
- **03-private-backend** -- disables public networking entirely for internal services reachable only via private network

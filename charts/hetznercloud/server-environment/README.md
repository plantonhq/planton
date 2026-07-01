# Hetzner Cloud Server Environment

Single-server environment on Hetzner Cloud with private networking, SSH key
authentication, firewall rules, and optional block volume storage.

## Use Case

Deploy a single Hetzner Cloud server with a private network, SSH access, and
basic firewall rules. Ideal for developer sandboxes, small applications, CI
runners, and hobby deployments.

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Private Network | `HetznerCloudNetwork` | Always |
| SSH Key | `HetznerCloudSshKey` | Always |
| Firewall | `HetznerCloudFirewall` | Always |
| Server | `HetznerCloudServer` | Always |
| Block Volume | `HetznerCloudVolume` | `enable_volume` |

## Default Firewall Rules

- **SSH (TCP 22)**: Inbound from anywhere (IPv4 + IPv6)
- **ICMP**: Inbound from anywhere (ping)

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `location` | Hetzner Cloud location | `fsn1` |
| `network_ip_range` | Private network CIDR | `10.0.0.0/16` |
| `subnet_ip_range` | Subnet CIDR | `10.0.1.0/24` |
| `subnet_network_zone` | Network zone | `eu-central` |
| `ssh_public_key` | SSH public key | -- |
| `server_type` | Server type (vCPU/RAM) | `cx22` |
| `image` | OS image | `ubuntu-24.04` |
| `user_data` | Cloud-init script | -- |
| `enable_volume` | Create block volume | `false` |
| `volume_size` | Volume size in GB | `10` |
| `volume_format` | Volume filesystem | `ext4` |

## Dependency Graph

```
Network ─┐
SSH Key ──┼─→ Server ──→ Volume (optional)
Firewall ─┘
```

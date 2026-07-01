# Hetzner Cloud Load-Balanced App

Multi-server web application on Hetzner Cloud with a load balancer, private
networking, SSH key authentication, and firewall rules. Optional HTTPS
termination with managed Let's Encrypt certificates and DNS zone for custom
domain routing.

## Use Case

Deploy a production web application across multiple Hetzner Cloud servers behind
a load balancer. Traffic is distributed via round-robin across all servers on a
private network. Enable DNS and HTTPS for custom domain routing with automatic
TLS certificate management.

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Private Network | `HetznerCloudNetwork` | Always |
| SSH Key | `HetznerCloudSshKey` | Always |
| Firewall | `HetznerCloudFirewall` | Always |
| App Servers | `HetznerCloudServer` x `server_count` | Always |
| DNS Zone | `HetznerCloudDnsZone` | `enable_dns` |
| TLS Certificate | `HetznerCloudCertificate` | `enable_https` |
| Load Balancer | `HetznerCloudLoadBalancer` | Always |

## Default Firewall Rules

- **SSH (TCP 22)**: Inbound from anywhere (IPv4 + IPv6)
- **HTTP (TCP 80)**: Inbound from anywhere
- **HTTPS (TCP 443)**: Inbound from anywhere
- **App Port**: Inbound from anywhere (configurable via `app_port`)
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
| `server_count` | Number of app servers | `2` |
| `app_port` | Application listen port | `8080` |
| `load_balancer_type` | LB size (lb11/lb21/lb31) | `lb11` |
| `enable_dns` | Create DNS zone | `false` |
| `domain_name` | Domain for DNS zone | `example.com` |
| `enable_https` | Enable HTTPS with Let's Encrypt | `false` |

## Dependency Graph

```
Network ─────┐
SSH Key ──────┼─→ Server(s) ──→ Load Balancer
Firewall ─────┘                    ↑
Certificate (opt) ─────────────────┘
DnsZone (opt, independent)
```

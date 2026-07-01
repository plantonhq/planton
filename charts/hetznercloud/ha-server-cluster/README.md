# Hetzner Cloud HA Server Cluster

High-availability server cluster on Hetzner Cloud with anti-affinity placement,
a load balancer for traffic distribution, and a floating IP for failover.

## Use Case

Deploy a cluster of servers guaranteed to run on different physical hosts via a
placement group. A load balancer distributes traffic using least-connections
algorithm, and a floating IP provides a stable address that can be reassigned
between nodes for failover without DNS changes. Ideal for databases (PostgreSQL,
MySQL), key-value stores (Redis, etcd), and other stateful services that require
high availability.

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Private Network | `HetznerCloudNetwork` | Always |
| SSH Key | `HetznerCloudSshKey` | Always |
| Firewall | `HetznerCloudFirewall` | Always |
| Placement Group | `HetznerCloudPlacementGroup` | Always |
| Cluster Nodes | `HetznerCloudServer` x `server_count` | Always |
| Floating IP | `HetznerCloudFloatingIp` | Always |
| Load Balancer | `HetznerCloudLoadBalancer` | Always |

## Default Firewall Rules

- **SSH (TCP 22)**: Inbound from anywhere (IPv4 + IPv6)
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
| `server_count` | Number of cluster nodes | `3` |
| `app_port` | Service listen port | `5432` |
| `load_balancer_type` | LB size (lb11/lb21/lb31) | `lb11` |
| `floating_ip_type` | Floating IP version | `ipv4` |

## Dependency Graph

```
Network ──────────┐
SSH Key ───────────┤
Firewall ──────────┼─→ Node(s) ──→ Load Balancer
Placement Group ───┘
Floating IP (independent)
```

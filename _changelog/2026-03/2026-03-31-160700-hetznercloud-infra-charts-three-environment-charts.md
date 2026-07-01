# Hetzner Cloud InfraCharts: Three Environment Charts

**Date**: March 31, 2026
**Type**: New Chart
**Provider**: Hetzner Cloud
**Chart(s)**: server-environment, load-balanced-app, ha-server-cluster

## Summary

Added three production-ready InfraCharts for Hetzner Cloud, covering the core
deployment patterns for bare-metal-grade cloud servers: single-server
environments, load-balanced web applications, and high-availability server
clusters. These charts bring Hetzner Cloud to first-class InfraChart status
alongside AWS, GCP, Azure, OCI, Scaleway, and Civo.

## Problem Statement / Motivation

Hetzner Cloud had 12 fully implemented Planton deployment components (SSH keys,
networks, firewalls, servers, volumes, load balancers, certificates, DNS zones,
placement groups, floating IPs, primary IPs, snapshots) plus a complete provider
connection API, but zero InfraCharts. Without charts, users had to manually
compose individual resources and wire network IDs to servers, server IDs to load
balancer targets, and certificate IDs to HTTPS services.

### Pain Points

- No one-click Hetzner Cloud environment provisioning despite having all
  building blocks
- Manual valueFrom wiring between 5-7 resources per environment is error-prone
- No cross-provider parity -- AWS, GCP, Azure, OCI, and Scaleway had
  environment charts but Hetzner Cloud didn't
- Hetzner Cloud's clean, focused resource set makes it an ideal candidate for
  opinionated environment charts

## Solution / What's New

Three charts under `hetznercloud/` in the infra-charts repo, following the
conventions established by `oci/compute-environment`, `aws/ecs-environment`,
and `scaleway/kapsule-environment`.

### Chart Structure

```
hetznercloud/
├── server-environment/      # Single server with networking and storage
├── load-balanced-app/       # Multi-server web app with LB, DNS, HTTPS
└── ha-server-cluster/       # HA cluster with anti-affinity and failover
```

## Implementation Details

### Resources Included

| Chart | Resources | Templates | Params | Conditional |
|-------|-----------|-----------|--------|-------------|
| server-environment | Network, SshKey, Firewall, Server, Volume | 4 | 11 | Volume |
| load-balanced-app | Network, SshKey, Firewall, Server(s), DnsZone, Certificate, LB | 6 | 14 | DNS, HTTPS |
| ha-server-cluster | Network, SshKey, Firewall, PlacementGroup, Server(s), FloatingIp, LB | 5 | 12 | None |

### Hetzner Cloud-Specific Patterns

**Multi-server via Jinja loops**: The load-balanced-app and ha-server-cluster
charts introduce `{% for %}` loops to create a configurable number of servers
from a single `server_count` parameter. This is a new pattern for the chart
library -- previous charts created at most one instance of each resource kind.

**Private network traffic**: Both load-balanced charts route traffic through the
private network (`usePrivateIp: true` on server targets) instead of the public
internet, following Hetzner Cloud best practices.

**Algorithm selection**: load-balanced-app uses `round_robin` for even web
traffic distribution. ha-server-cluster uses `least_connections` for stateful
services where connection duration varies.

**Placement group anti-affinity**: ha-server-cluster assigns all nodes to a
`spread` placement group, guaranteeing they run on different physical hosts.

### Conditional Resources

```yaml
# Block volume (server-environment)
{% if values.enable_volume | bool %}
kind: HetznerCloudVolume
{% endif %}

# DNS zone (load-balanced-app)
{% if values.enable_dns | bool %}
kind: HetznerCloudDnsZone
{% endif %}

# HTTPS certificate (load-balanced-app)
{% if values.enable_https | bool %}
kind: HetznerCloudCertificate
{% endif %}
```

### Resource Relationships (valueFrom Wiring)

All charts wire resources through `valueFrom` references:

```
HetznerCloudNetwork (outputs: network_id)
HetznerCloudSshKey (outputs: ssh_key_id)
HetznerCloudFirewall (outputs: firewall_id)
  └─ HetznerCloudServer (inputs: networks[].networkId, sshKeys[], firewallIds[])
       └─ HetznerCloudLoadBalancer (inputs: network.networkId, serverTargets[].serverId)
       └─ HetznerCloudVolume (inputs: serverId)
```

## Benefits

- **5-minute Hetzner Cloud environments**: From zero to a running server, a
  load-balanced web app, or an HA database cluster
- **Cross-provider parity**: Hetzner Cloud joins 8 other providers as a
  first-class InfraChart provider
- **Hetzner Cloud best practices baked in**: Private networking, SSH-only
  access, firewall rules, anti-affinity placement
- **Composable toggles**: Optional volume, DNS, HTTPS via boolean flags
- **24 files across 3 charts**: Complete chart family

## Impact

Platform users can now provision Hetzner Cloud environments through InfraCharts
with the same experience as AWS ECS, OCI Compute, or Scaleway Kapsule
environments. The three charts cover the most common Hetzner Cloud deployment
patterns from developer sandbox to production HA cluster.

## Usage Example

```bash
# Preview a chart
planton chart build hetznercloud/server-environment

# Create a project from the chart
planton project create --from-chart hetznercloud/load-balanced-app \
  --name my-web-app \
  --values ./my-values.yaml
```

Example `values.yaml` override for load-balanced-app:

```yaml
params:
  - name: location
    value: fsn1
  - name: ssh_public_key
    value: "ssh-ed25519 AAAA... user@host"
  - name: server_type
    value: cpx21
  - name: server_count
    value: "3"
  - name: app_port
    value: "3000"
  - name: enable_dns
    value: true
  - name: domain_name
    value: myapp.example.com
  - name: enable_https
    value: true
```

## Related Work

- Hetzner Cloud Planton components: 12 resource kinds in planton-hetzner-cloud
- Planton monorepo assets: deployment-component.yaml, iac-modules.yaml for all
  12 components (v0.3.57)
- HetznerCloudProviderConnection API: complete across all platform layers
- Parent project: 20260219.03.sp.hetznercloud-resource-expansion

---

**Status**: Production Ready
**Timeline**: Created 2026-03-31

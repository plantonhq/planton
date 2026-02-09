# OpenStack Load Balancer Monitor

Provision and manage Octavia health monitors in OpenStack using OpenMCF's unified API.

## Overview

An Octavia health monitor periodically checks the health of pool members and removes unhealthy members from the pool's rotation until they recover. Monitors support HTTP, HTTPS, PING, TCP, TLS-HELLO, and UDP-CONNECT check types.

This component creates an `openstack_lb_monitor_v2` resource through both Pulumi and Terraform IaC modules with full feature parity. The monitor name is derived from `metadata.name`.

**Important**: Health monitors do NOT support tags in the Terraform OpenStack provider.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Octavia
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)
4. **Pool**: An existing OpenStack load balancer pool (see `OpenStackLoadBalancerPool`)

## Quick Start

### Minimal HTTP Monitor

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-health
spec:
  pool_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  type: "HTTP"
  delay: 5
  timeout: 10
  max_retries: 3
```

### Deploy

```bash
openmcf apply --manifest monitor.yaml -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pool_id` | StringValueOrRef | Yes | Pool FK. ForceNew |
| `type` | string | Yes | Check type: HTTP, HTTPS, PING, TCP, TLS-HELLO, UDP-CONNECT. ForceNew |
| `delay` | int32 | Yes | Interval between checks (seconds) |
| `timeout` | int32 | Yes | Timeout per check (seconds) |
| `max_retries` | int32 | Yes | Consecutive successes for healthy (1-10) |
| `max_retries_down` | int32 | No | Consecutive failures for unhealthy (1-10) |
| `url_path` | string | No | URL path for HTTP/HTTPS checks |
| `http_method` | string | No | HTTP method for HTTP/HTTPS checks |
| `expected_codes` | string | No | Expected HTTP codes (e.g., "200", "200-299") |
| `admin_state_up` | bool | No | Administrative state. Default: true |
| `region` | string | No | Override region from provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `monitor_id` | UUID of the health monitor |
| `name` | Monitor name |
| `type` | Health check type |
| `pool_id` | Monitored pool ID |
| `region` | Region where the monitor was created |

## Foreign Key Relationships

**Inbound:** `pool_id` -> `OpenStackLoadBalancerPool.status.outputs.pool_id`

## IaC Implementations

- **Terraform resource**: `openstack_lb_monitor_v2`
- **Pulumi resource**: `openstack.loadbalancer.Monitor`

## Notes

- **No tags support**: Health monitors do NOT support tags in the Terraform OpenStack provider.
- **HTTP fields**: url_path, http_method, and expected_codes are only valid for HTTP/HTTPS monitors.
- **CEL validation**: Setting HTTP fields on non-HTTP monitors is rejected at the API level.
- **max_retries_down**: Defaults to the same value as max_retries when not set.
- **Admin state**: Set admin_state_up false to stop health checking without deleting the monitor.

# OpenStack Load Balancer Monitor

Deploys an Octavia health monitor that periodically probes pool members to determine their health status. Unhealthy members are automatically removed from the pool's traffic rotation until they recover. Monitors support HTTP, HTTPS, PING, TCP, TLS-HELLO, and UDP-CONNECT check types.

## What Gets Created

When you deploy an OpenStackLoadBalancerMonitor resource, OpenMCF provisions:

- **Octavia Health Monitor** — a `loadbalancer.Monitor` resource attached to the specified pool. The monitor sends periodic probes to each pool member at the configured interval. After the required number of consecutive failures, a member is removed from rotation. After the required number of consecutive successes, it is restored. HTTP and HTTPS monitors validate response codes against configurable expectations; PING, TCP, TLS-HELLO, and UDP-CONNECT monitors check connectivity only.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An Octavia pool** to attach the monitor to (each pool supports at most one monitor)

## Quick Start

Create a file `monitor.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: http-health
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackLoadBalancerMonitor.http-health
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  type: HTTP
  delay: 5
  timeout: 3
  maxRetries: 3
  urlPath: /healthz
  expectedCodes: "200"
```

Deploy:

```shell
openmcf apply -f monitor.yaml
```

This creates an HTTP health monitor that checks `/healthz` on each pool member every 5 seconds, expects a 200 response within 3 seconds, and requires 3 consecutive successes to mark a member healthy.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `poolId` | `StringValueOrRef` | UUID of the Octavia pool to monitor. Can reference an OpenStackLoadBalancerPool resource via `valueFrom`. ForceNew: changing this recreates the monitor. | Required |
| `type` | `string` | The type of health check to perform: `HTTP`, `HTTPS`, `PING`, `TCP`, `TLS-HELLO`, or `UDP-CONNECT`. ForceNew: changing this recreates the monitor. | Must be one of the listed values |
| `delay` | `int32` | Interval in seconds between consecutive health checks sent to each member. | Required |
| `timeout` | `int32` | Maximum time in seconds to wait for a health check response. If a member does not respond within this time, the check is considered failed. | Required |
| `maxRetries` | `int32` | Number of consecutive successful health checks required before a member is considered healthy and returned to rotation. | Required; must be between 1 and 10 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maxRetriesDown` | `int32` | same as `maxRetries` | Number of consecutive failed health checks required before a member is considered unhealthy and removed from rotation. Must be between 1 and 10. |
| `urlPath` | `string` | — | URL path to request for HTTP/HTTPS health checks (e.g., `/healthz`). Only valid when `type` is `HTTP` or `HTTPS`. |
| `httpMethod` | `string` | — | HTTP method for HTTP/HTTPS health checks: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`, `CONNECT`, or `TRACE`. Only valid when `type` is `HTTP` or `HTTPS`. |
| `expectedCodes` | `string` | — | Expected HTTP response codes for a healthy member. Supports single codes (`"200"`), ranges (`"200-299"`), and comma-separated lists (`"200,202"`). Only valid when `type` is `HTTP` or `HTTPS`. |
| `adminStateUp` | `bool` | `true` | Administrative state of the monitor. When `false`, the monitor stops checking members without being deleted. |
| `region` | `string` | provider default | Overrides the region from the provider config for this monitor. |

## Examples

### HTTP Health Monitor

An HTTP monitor that checks a health endpoint on each pool member, suitable for web application pools:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: web-http-health
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackLoadBalancerMonitor.web-http-health
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  type: HTTP
  delay: 10
  timeout: 5
  maxRetries: 3
  urlPath: /healthz
  httpMethod: GET
  expectedCodes: "200"
```

### TCP Health Monitor

A TCP monitor for non-HTTP services. The monitor attempts a TCP connection to each member and marks it healthy if the connection succeeds:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: db-tcp-check
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancerMonitor.db-tcp-check
spec:
  poolId: 7d8e9f0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a
  type: TCP
  delay: 5
  timeout: 3
  maxRetries: 3
```

### HTTP Monitor with Asymmetric Thresholds

An HTTP monitor with different thresholds for failure detection and recovery. Members are removed after 2 consecutive failures (fast detection) but require 5 consecutive successes to return to rotation (cautious recovery):

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: api-health
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancerMonitor.api-health
spec:
  poolId: 12345678-abcd-efgh-ijkl-123456789abc
  type: HTTPS
  delay: 10
  timeout: 5
  maxRetries: 5
  maxRetriesDown: 2
  urlPath: /api/health
  httpMethod: GET
  expectedCodes: "200-299"
```

### Using Foreign Key References

Reference an OpenMCF-managed pool instead of hardcoding UUIDs:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerMonitor
metadata:
  name: ref-monitor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancerMonitor.ref-monitor
spec:
  poolId:
    valueFrom:
      kind: OpenStackLoadBalancerPool
      name: web-pool
      field: status.outputs.pool_id
  type: HTTP
  delay: 5
  timeout: 3
  maxRetries: 3
  urlPath: /healthz
  expectedCodes: "200"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `monitor_id` | `string` | UUID of the created health monitor |
| `name` | `string` | Name of the monitor, derived from `metadata.name` |
| `type` | `string` | Health check type (HTTP, HTTPS, PING, TCP, TLS-HELLO, or UDP-CONNECT) |
| `pool_id` | `string` | UUID of the monitored pool |
| `region` | `string` | OpenStack region where the monitor was created |

## Related Components

- [OpenStackLoadBalancer](/docs/catalog/openstack/openstackloadbalancer) — provides the top-level VIP that receives client traffic
- [OpenStackLoadBalancerListener](/docs/catalog/openstack/openstackloadbalancerlistener) — binds a protocol and port on the load balancer to a pool
- [OpenStackLoadBalancerPool](/docs/catalog/openstack/openstackloadbalancerpool) — the pool whose members this monitor checks
- [OpenStackLoadBalancerMember](/docs/catalog/openstack/openstackloadbalancermember) — the backend servers being health-checked by this monitor

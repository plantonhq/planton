# OpenStack Load Balancer Pool

Deploys an Octavia backend pool in OpenStack that groups members behind a listener, defining the backend protocol, load-balancing algorithm, and optional session persistence for traffic distribution.

## What Gets Created

When you deploy an OpenStackLoadBalancerPool resource, Planton provisions:

- **Octavia Pool** — a `loadbalancer.Pool` resource attached to the specified listener, configured with the chosen protocol, load-balancing algorithm, and optional session persistence. Members and health monitors attach to this pool to complete the backend configuration.

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **An existing listener** — the pool requires a listener UUID (from an OpenStackLoadBalancerListener resource or provisioned externally)
- **Octavia service** enabled in the target OpenStack project

## Quick Start

Create a file `pool.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: web-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackLoadBalancerPool.web-pool
spec:
  listenerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: HTTP
  lbMethod: ROUND_ROBIN
```

Deploy:

```shell
planton apply -f pool.yaml
```

This creates an Octavia pool using round-robin distribution for HTTP traffic, attached to the specified listener.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `listenerId` | `StringValueOrRef` | UUID of the listener this pool is the default pool for. ForceNew: changing this recreates the pool. Can reference an OpenStackLoadBalancerListener resource via `valueFrom`. | Required |
| `protocol` | `string` | The protocol used by pool members to receive traffic. ForceNew: changing this recreates the pool. | One of: `HTTP`, `HTTPS`, `TCP`, `UDP`, `PROXY` |
| `lbMethod` | `string` | The load-balancing algorithm to distribute traffic across members. | One of: `ROUND_ROBIN`, `LEAST_CONNECTIONS`, `SOURCE_IP`, `SOURCE_IP_PORT` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `persistence` | `SessionPersistence` | — | Session persistence configuration. Ensures requests from the same client are routed to the same backend member. See sub-fields below. |
| `persistence.type` | `string` | — | The persistence type. Required when `persistence` is set. One of: `SOURCE_IP` (hash client IP), `HTTP_COOKIE` (Octavia-managed cookie), `APP_COOKIE` (application-managed cookie, requires `cookieName`). |
| `persistence.cookieName` | `string` | — | The application cookie name for session affinity. Only valid when `persistence.type` is `APP_COOKIE`. Validated by a CEL rule that rejects this field for other persistence types. |
| `description` | `string` | — | Human-readable description of the pool. |
| `adminStateUp` | `bool` | `true` | Administrative state of the pool. When `false`, the pool stops receiving traffic. |
| `tags` | `string[]` | `[]` | Tags applied to the pool in OpenStack for filtering and organization. Must be unique within this resource. |
| `region` | `string` | provider default | Overrides the region from the provider config for this pool. |

## Examples

### Basic HTTP Pool with Round-Robin

A minimal pool distributing HTTP traffic evenly across members:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: web-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackLoadBalancerPool.web-pool
spec:
  listenerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: HTTP
  lbMethod: ROUND_ROBIN
  description: "Web backend pool"
  tags:
    - web
    - frontend
```

### Pool with Application Cookie Persistence

A pool with `APP_COOKIE` session persistence, routing clients with the same `JSESSIONID` cookie to the same backend member. Useful for Java-based applications with server-side sessions:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: app-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OpenStackLoadBalancerPool.app-pool
spec:
  listenerId: 7d8e9f0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a
  protocol: HTTP
  lbMethod: LEAST_CONNECTIONS
  persistence:
    type: APP_COOKIE
    cookieName: JSESSIONID
  tags:
    - production
    - app-tier
```

### TCP Pool with Source IP Affinity

A TCP pool using `SOURCE_IP` load balancing to route all connections from a given client IP to the same backend. Suitable for non-HTTP protocols or stateful TCP services:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: tcp-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OpenStackLoadBalancerPool.tcp-pool
spec:
  listenerId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
  protocol: TCP
  lbMethod: SOURCE_IP
  persistence:
    type: SOURCE_IP
  adminStateUp: true
  region: RegionOne
```

### Using Foreign Key References

Reference an Planton-managed listener instead of hardcoding the UUID:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: ref-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OpenStackLoadBalancerPool.ref-pool
spec:
  listenerId:
    valueFrom:
      kind: OpenStackLoadBalancerListener
      name: http-listener
      field: status.outputs.listener_id
  protocol: HTTP
  lbMethod: ROUND_ROBIN
  persistence:
    type: HTTP_COOKIE
  description: "Pool with FK reference and Octavia-managed cookie"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `poolId` | `string` | UUID of the created Octavia pool. Used as a foreign key by OpenStackLoadBalancerMember and OpenStackLoadBalancerMonitor. |
| `name` | `string` | Name of the pool, derived from `metadata.name` |
| `protocol` | `string` | Backend protocol of the pool (`HTTP`, `HTTPS`, `TCP`, `UDP`, or `PROXY`) |
| `lbMethod` | `string` | Load-balancing algorithm of the pool |
| `region` | `string` | OpenStack region where the pool was created |

## Related Components

- [OpenStackLoadBalancer](/docs/catalog/openstack/openstackloadbalancer) — provides the top-level load balancer (VIP) that owns listeners
- [OpenStackLoadBalancerListener](/docs/catalog/openstack/openstackloadbalancerlistener) — provides the listener that this pool attaches to
- [OpenStackLoadBalancerMember](/docs/catalog/openstack/openstackloadbalancermember) — adds backend servers to this pool
- [OpenStackLoadBalancerMonitor](/docs/catalog/openstack/openstackloadbalancermonitor) — attaches health checks to this pool

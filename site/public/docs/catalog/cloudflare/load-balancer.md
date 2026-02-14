---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareloadbalancer"
---

# Cloudflare Load Balancer

Deploys a Cloudflare Load Balancer with an associated origin pool and health monitor. The component distributes traffic across one or more origin servers, with configurable health checks, session affinity, and traffic steering policies.

## What Gets Created

When you deploy a CloudflareLoadBalancer resource, OpenMCF provisions:

- **Load Balancer Monitor** — an HTTP health check that probes each origin at the configured `healthProbePath`, expecting `2xx` responses, with 2 retries and a 5-second timeout
- **Load Balancer Pool** — a pool named `{metadata.name}-pool` containing all declared origins with their respective weights, linked to the monitor
- **Load Balancer** — a `cloudflare_load_balancer` resource bound to the specified hostname and zone, using the pool as both the default and fallback pool, with the configured proxy, steering, and session affinity settings

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **An existing Cloudflare DNS zone** — either the zone ID as a literal string or a deployed CloudflareDnsZone resource to reference
- **Appropriate permissions** — the API token must have Load Balancing edit access for the target zone
- **Cloudflare Load Balancing add-on** — Load Balancing must be enabled on your Cloudflare account (it is a paid feature)

## Quick Start

Create a file `load-balancer.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: my-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareLoadBalancer.my-lb
spec:
  hostname: app.example.com
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  origins:
    - name: primary
      address: 203.0.113.10
      weight: 1
  proxied: true
  healthProbePath: /healthz
```

Deploy:

```shell
openmcf apply -f load-balancer.yaml
```

This creates a proxied load balancer for `app.example.com` with a single origin, health-checked at `/healthz`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `hostname` | `string` | The DNS hostname that the load balancer will serve (e.g., `app.example.com`). | Required |
| `zoneId` | `StringValueOrRef` | The Cloudflare Zone ID containing the hostname. Accepts a literal `value` string or a `valueFrom` reference to a CloudflareDnsZone resource. | Required |
| `origins` | `list` | List of origin servers behind the load balancer. Each origin has `name`, `address`, and optional `weight`. | Required, minimum 1 item |

Each entry in `origins` has the following fields:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `origins[].name` | `string` | A label to identify the origin within the pool. | Required |
| `origins[].address` | `string` | The origin server address (IP or DNS hostname) reachable via HTTP(S). | Required |
| `origins[].weight` | `int32` | Relative traffic weight for this origin. Higher values receive more traffic. | Default: `1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `proxied` | `bool` | `true` | Whether to route traffic through Cloudflare's CDN and WAF (orange-cloud mode). |
| `healthProbePath` | `string` | `"/"` | The HTTP path used by the health monitor to check origin availability. The monitor sends `GET` requests and expects `2xx` responses. |
| `sessionAffinity` | `enum` | `none` | Session affinity mode. `none` disables affinity. `cookie` pins clients to an origin using a Cloudflare-managed cookie. |
| `steeringPolicy` | `enum` | `off` | Traffic steering policy. `off` for static failover (origins tried in order), `geo` for geographic routing, `random` for weighted random distribution. |

### Zone ID Reference

The `zoneId` field accepts either a literal value or a cross-resource reference.

**Literal value:**

```yaml
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

**Reference to a CloudflareDnsZone resource:**

```yaml
spec:
  zoneId:
    valueFrom:
      name: my-zone
```

When using `valueFrom`, the `kind` defaults to `CloudflareDnsZone` and the `fieldPath` defaults to `status.outputs.zone_id`, so only the resource `name` is required. You may also specify `env` to reference a zone deployed in a different environment.

## Examples

### Two-Origin Failover

A load balancer with a primary and secondary origin. The default steering policy (`off`) means Cloudflare tries origins in order, falling back to the secondary if the primary fails health checks:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: api-failover
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareLoadBalancer.api-failover
spec:
  hostname: api.example.com
  zoneId:
    valueFrom:
      name: prod-zone
  origins:
    - name: us-east
      address: 198.51.100.10
      weight: 1
    - name: us-west
      address: 198.51.100.20
      weight: 1
  proxied: true
  healthProbePath: /health
```

### Weighted Random Distribution with Session Affinity

A load balancer that distributes traffic randomly by weight across three origins, using cookie-based session affinity to keep returning users on the same origin:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: web-weighted
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareLoadBalancer.web-weighted
spec:
  hostname: www.example.com
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  origins:
    - name: origin-a
      address: 203.0.113.1
      weight: 3
    - name: origin-b
      address: 203.0.113.2
      weight: 2
    - name: origin-c
      address: 203.0.113.3
      weight: 1
  proxied: true
  healthProbePath: /up
  sessionAffinity: cookie
  steeringPolicy: random
```

### Geo-Routed Load Balancer with Zone Reference

A load balancer that uses geo steering to route users to the nearest origin, referencing a CloudflareDnsZone resource for the zone ID:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: global-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareLoadBalancer.global-lb
spec:
  hostname: app.example.com
  zoneId:
    valueFrom:
      name: prod-zone
      env: prod
  origins:
    - name: eu-west
      address: 192.0.2.10
      weight: 1
    - name: ap-southeast
      address: 192.0.2.20
      weight: 1
    - name: us-east
      address: 192.0.2.30
      weight: 1
  proxied: true
  healthProbePath: /ping
  steeringPolicy: geo
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `loadBalancerId` | `string` | The unique identifier of the created Cloudflare Load Balancer |
| `loadBalancerDnsRecordName` | `string` | The hostname DNS record associated with the load balancer |
| `loadBalancerCnameTarget` | `string` | The canonical CNAME target that the hostname resolves to (the Cloudflare endpoint) |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/dns-zone) — manages the parent DNS zone; its `zone_id` output can be referenced by this component via `valueFrom`
- [CloudflareDnsRecord](/docs/catalog/cloudflare/dns-record) — manages individual DNS records; commonly used alongside load balancers for non-load-balanced hostnames
- [CloudflareWorker](/docs/catalog/cloudflare/worker) — serverless functions that can act as origins or be paired with load-balanced routes

---
title: "HTTP Listener"
description: "This preset creates an HTTP listener on port 80. It accepts unencrypted HTTP traffic and forwards it to a backend pool. This is the simplest and most common listener configuration -- suitable for..."
type: "preset"
rank: "01"
presetSlug: "01-http"
componentSlug: "load-balancer-listener"
componentTitle: "Load Balancer Listener"
provider: "openstack"
icon: "package"
order: 1
---

# HTTP Listener

This preset creates an HTTP listener on port 80. It accepts unencrypted HTTP traffic and forwards it to a backend pool. This is the simplest and most common listener configuration -- suitable for internal services, HTTP-to-HTTPS redirects, or environments where TLS is terminated upstream.

## When to Use

- Internal services that communicate over plain HTTP within a VPC
- HTTP-to-HTTPS redirect listeners (paired with an application-level redirect)
- Development and testing environments where TLS overhead is unnecessary

## Key Configuration Choices

- **HTTP protocol** (`protocol: HTTP`) -- Layer 7, unencrypted
- **Port 80** -- standard HTTP port
- **Unlimited connections** -- no `connectionLimit` set; Octavia's default applies
- **No CIDR restrictions** -- all sources can reach the listener (add `allowedCidrs` to restrict)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<loadbalancer-id>` | ID of the load balancer to attach this listener to | OpenStack console or `OpenStackLoadBalancer` status outputs |

## Related Presets

- **02-https-terminated** -- Use instead when TLS should be terminated at the load balancer
- **03-tcp-passthrough** -- Use instead for raw TCP traffic (databases, message queues)

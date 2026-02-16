---
title: "TCP Health Check Monitor"
description: "This preset creates a TCP health monitor that checks backend members by attempting a TCP connection every 10 seconds. If the connection succeeds, the member is healthy. This is the standard monitor..."
type: "preset"
rank: "02"
presetSlug: "02-tcp-health-check"
componentSlug: "load-balancer-monitor"
componentTitle: "Load Balancer Monitor"
provider: "openstack"
icon: "package"
order: 2
---

# TCP Health Check Monitor

This preset creates a TCP health monitor that checks backend members by attempting a TCP connection every 10 seconds. If the connection succeeds, the member is healthy. This is the standard monitor for non-HTTP services where application-level health checks are not available.

## When to Use

- Database pools (PostgreSQL, MySQL, Redis) behind a TCP listener
- Message queue services (Kafka, NATS, RabbitMQ)
- Any TCP service that does not expose an HTTP health endpoint

## Key Configuration Choices

- **TCP type** -- verifies the port is reachable (TCP handshake) without any application-level protocol
- **10s interval** (`delay: 10`) -- check every 10 seconds
- **5s timeout** (`timeout: 5`) -- fail if TCP handshake does not complete within 5 seconds
- **3 retries** (`maxRetries: 3`) -- 3 consecutive passes to mark healthy, 3 consecutive failures to mark unhealthy

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<pool-id>` | ID of the pool to monitor | OpenStack console or `OpenStackLoadBalancerPool` status outputs |

## Related Presets

- **01-http-health-check** -- Use instead for HTTP services with a dedicated health endpoint

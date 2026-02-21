---
title: "TCP Pass-Through"
description: "This preset creates a layer-4 TCP load balancer that forwards raw TCP connections to backend servers without any application-layer inspection. The load balancer does not parse HTTP headers, manage..."
type: "preset"
rank: "03"
presetSlug: "03-tcp-pass-through"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "hetznercloud"
icon: "package"
order: 3
---

# TCP Pass-Through

This preset creates a layer-4 TCP load balancer that forwards raw TCP connections to backend servers without any application-layer inspection. The load balancer does not parse HTTP headers, manage cookies, or terminate TLS -- it simply relays bytes between clients and targets. This is the correct pattern for databases, message queues, mail servers, game servers, or any protocol that is not HTTP.

The least_connections algorithm is chosen because TCP services like databases and connection pools typically have varying response times. Sending each new connection to the target with the fewest active connections produces better utilization than round_robin for these workloads.

## When to Use

- Database connection pooling (PostgreSQL on 5432, MySQL on 3306, Redis on 6379)
- Message queue frontends (RabbitMQ on 5672, Kafka on 9092)
- Mail servers, game servers, or any custom TCP protocol
- Any service where the load balancer must not inspect or modify the application payload

## Key Configuration Choices

- **TCP protocol** (`protocol: tcp`) -- layer-4 pass-through with no HTTP inspection; the load balancer forwards raw TCP connections
- **Explicit ports required** (`listenPort`, `destinationPort`) -- unlike HTTP and HTTPS, TCP has no default ports; both must be set (enforced by CEL validation in the spec)
- **Least connections** (`algorithm: least_connections`) -- routes each new connection to the target with the fewest active connections; better than round_robin for backends with uneven processing times
- **No health check block** -- omitting the health check lets the provider auto-create a TCP connection check on the destination port; this is correct for generic TCP services where there is no HTTP endpoint to probe
- **No delete protection** -- this preset omits `deleteProtection` to keep it minimal; add `deleteProtection: true` for production deployments

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<listen-port>` | Port the load balancer listens on for incoming TCP connections | Determined by your protocol (e.g., 5432 for PostgreSQL, 6379 for Redis) |
| `<destination-port>` | Port on backend servers that receives forwarded connections | Usually the same as listen-port unless backends use a non-standard port |
| `<backend-server-id-1>` | Numeric ID of the first backend server | The `status.outputs.server_id` of your HetznerCloudServer resource, or the Servers page in the Hetzner Cloud Console |
| `<backend-server-id-2>` | Numeric ID of the second backend server | Same as above for your second backend server |

## Related Presets

- **01-https-web-app** -- HTTPS load balancer with TLS termination for web applications
- **02-private-internal** -- private-network-only load balancer for internal HTTP services

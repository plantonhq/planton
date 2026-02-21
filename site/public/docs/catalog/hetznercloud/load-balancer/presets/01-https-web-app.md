---
title: "HTTPS Web Application"
description: "This preset creates a public-facing HTTPS load balancer that terminates TLS at the edge and forwards plain HTTP to backend servers. It automatically redirects all HTTP traffic to HTTPS, discovers..."
type: "preset"
rank: "01"
presetSlug: "01-https-web-app"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "hetznercloud"
icon: "package"
order: 1
---

# HTTPS Web Application

This preset creates a public-facing HTTPS load balancer that terminates TLS at the edge and forwards plain HTTP to backend servers. It automatically redirects all HTTP traffic to HTTPS, discovers backends dynamically via a Hetzner Cloud label selector, and runs an HTTP health check against each target's application port. This is the standard production pattern for web applications, APIs, and any service that needs encrypted public traffic.

The lb11 load balancer type supports up to 25 targets and 10,000 concurrent connections per second. Scale to lb21 (75 targets, 20k conn/s) or lb31 (150 targets, 40k conn/s) as traffic grows.

## When to Use

- Public-facing web applications or APIs that require HTTPS
- Services where backend servers are managed dynamically and identified by Hetzner Cloud labels
- Any deployment where the load balancer should handle TLS termination so backends serve plain HTTP

## Key Configuration Choices

- **HTTPS with redirect** (`protocol: https`, `redirectHttp: true`) -- the load balancer listens on 443 (the protocol default) and automatically creates an HTTP-to-HTTPS redirect on port 80; clients never reach the application over plain HTTP
- **TLS termination at the edge** (`destinationPort: 8080`) -- the LB decrypts TLS and forwards plain HTTP to backends on port 8080; backends do not need their own certificates
- **Label selector targets** (`labelSelectorTargets`) -- backends are discovered dynamically by matching a Hetzner Cloud label expression (e.g., `env=production,role=web`); servers matching the selector are added and removed automatically as labels change
- **HTTP health check on /health** (`healthCheck.http.path: /health`) -- the LB probes each backend's application port (8080) with an HTTP GET; targets that fail 3 consecutive checks (the default) are removed from rotation until they recover
- **Delete protection** (`deleteProtection: true`) -- prevents accidental destruction of a production load balancer via the API or Console; must be explicitly disabled before the resource can be removed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<certificate-id>` | Numeric ID of a Hetzner Cloud certificate (managed or uploaded) for TLS termination | The `status.outputs.certificate_id` of your HetznerCloudCertificate resource, or the Certificates page in the Hetzner Cloud Console |
| `<label-selector>` | Hetzner Cloud label selector expression that matches your backend servers | The labels you assign to your HetznerCloudServer resources (e.g., `env=production,role=web`) |

## Related Presets

- **02-private-internal** -- load balancer on a private network with no public interface, for internal service routing
- **03-tcp-pass-through** -- layer-4 TCP balancing for non-HTTP protocols like databases or message queues

# HTTP Health Check Monitor

This preset creates an HTTP health monitor that checks backend members by sending a GET request to `/healthz` every 10 seconds. A member is considered healthy after 3 consecutive successful responses (HTTP 200) and unhealthy after 3 consecutive failures. This is the most common monitor for web services.

## When to Use

- HTTP/HTTPS web applications with a dedicated health check endpoint
- REST APIs that expose a `/healthz` or `/health` endpoint
- Any HTTP pool where application-level health checking is needed

## Key Configuration Choices

- **HTTP type** -- sends real HTTP requests, validating application health (not just TCP connectivity)
- **GET /healthz** -- standard Kubernetes-style health endpoint; change `urlPath` to match your application
- **10s interval** (`delay: 10`) -- check every 10 seconds
- **5s timeout** (`timeout: 5`) -- fail if no response within 5 seconds
- **3 retries** (`maxRetries: 3`) -- 3 consecutive passes to mark healthy, 3 consecutive failures to mark unhealthy
- **HTTP 200 expected** -- only 200 is considered healthy; use `200-299` for broader acceptance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<pool-id>` | ID of the pool to monitor | OpenStack console or `OpenStackLoadBalancerPool` status outputs |

## Related Presets

- **02-tcp-health-check** -- Use instead for non-HTTP services (databases, message queues)

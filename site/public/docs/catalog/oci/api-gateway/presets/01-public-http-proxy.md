---
title: "Public HTTP Proxy"
description: "This preset creates a public-facing OCI API Gateway with a single deployment that proxies HTTP requests to a backend service. It includes CORS configuration for browser-based clients, a health check..."
type: "preset"
rank: "01"
presetSlug: "01-public-http-proxy"
componentSlug: "api-gateway"
componentTitle: "API Gateway"
provider: "oci"
icon: "package"
order: 1
---

# Public HTTP Proxy

This preset creates a public-facing OCI API Gateway with a single deployment that proxies HTTP requests to a backend service. It includes CORS configuration for browser-based clients, a health check endpoint returning a stock response, a wildcard catch-all route forwarding to the backend, and full access/execution logging. This is the simplest and most common pattern: expose a backend service to the internet through a managed gateway.

## When to Use

- Exposing a backend REST API to the internet with CORS support for browser-based frontends
- Adding a managed API gateway layer in front of existing microservices without modifying the backend
- Projects that need request logging, CORS, and timeouts without implementing them in application code
- Development and staging environments where authentication is handled at the application layer rather than the gateway

## Key Configuration Choices

- **Public endpoint** (`endpointType: endpoint_type_public`) -- the gateway is accessible from the internet. It must be placed in a public subnet. The gateway's public IP is available in status outputs and can be mapped to a custom domain via DNS.
- **NSG protection** (`networkSecurityGroupIds`) -- restricts network access to the gateway. Configure ingress rules allowing TCP port 443 from `0.0.0.0/0` for public APIs, or restrict to specific CIDR blocks for partner-only APIs.
- **CORS policy** (`requestPolicies.cors`) -- configured for a single frontend domain with credentials support. The preflight response is cached for 1 hour (`maxAgeInSeconds: 3600`). Adjust `allowedOrigins` to list your actual frontend domains; use `["*"]` only for fully public APIs where any origin is acceptable.
- **Health check route** (`/health` with stock response) -- returns `{"status":"ok"}` without forwarding to the backend. Useful for load balancer health probes, uptime monitors, and smoke tests that verify the gateway itself is operational.
- **Wildcard catch-all route** (`/{path*}`) -- forwards all other requests to the backend URL. The `{path*}` wildcard matches any path segment after the deployment's `pathPrefix`. The full request path is forwarded to the backend.
- **Timeouts** (`connectTimeoutInSeconds: 10`, `readTimeoutInSeconds: 30`, `sendTimeoutInSeconds: 30`) -- prevent hung connections. Connect timeout is aggressive (10s) to fail fast when the backend is unreachable. Read/send timeouts accommodate typical API response times. Increase `readTimeoutInSeconds` for long-running operations.
- **Access and execution logging** -- access logs capture every request/response pair for audit and analytics. Execution logs at `info` level capture gateway processing details for debugging routing and policy evaluation.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the gateway will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid>` | OCID of a public subnet for the gateway endpoint | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<gateway-nsg-ocid>` | OCID of the NSG allowing HTTPS ingress to the gateway | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<frontend-domain>` | Origin domain for CORS (e.g., `app.example.com`) | Your frontend deployment configuration |
| `<backend-url>` | Backend service URL (e.g., `https://backend.example.com:8080`) | Your backend service deployment |

## Related Presets

- **02-jwt-authenticated-api** -- Use instead when the gateway should validate JWT tokens before forwarding to the backend
- **03-private-functions-backend** -- Use instead for internal APIs backed by OCI Functions with no public internet exposure

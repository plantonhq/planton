# Internal GRPC ALB

This preset creates a VPC-internal ALB for service-to-service GRPC communication. The ALB is not accessible from the internet.

## When to Use

- Microservices communicating via GRPC within a VPC
- Internal API gateways that don't need public exposure
- Service mesh entry points using GRPC/HTTP2

## Key Configuration Choices

- **Intranet address type** -- VPC-internal only, no public DNS
- **Basic edition** -- sufficient for internal GRPC traffic, lower cost
- **GRPC protocol** -- server group configured for GRPC backends
- **Wlc scheduler** -- routes to the least-busy server, optimal for GRPC's long-lived connections
- **GRPC health checks** -- uses GRPC health checking protocol
- **HTTP/2 enabled** -- required for GRPC transport

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-alb-name>` | ALB name (2-128 chars) | Choose a descriptive name |
| `<alibaba-cloud-region>` | Region code | Your deployment region |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<zone-a>`, `<zone-b>` | Availability zones | Region's available zones |
| `<your-vswitch-id-a>`, `<your-vswitch-id-b>` | VSwitch IDs | `AliCloudVswitch` stack outputs |
| `<your-internal-cert-id>` | Internal certificate ID | Alibaba Cloud CAS or self-signed |

## Related Presets

- **01-internet-http** -- Public HTTP ALB
- **02-https-production** -- Public HTTPS ALB with WAF

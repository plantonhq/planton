# TLS Internet-Facing NLB

This preset creates an internet-facing Network Load Balancer with TLS termination on port 443. The NLB decrypts incoming TLS and forwards plaintext TCP to targets on your application port. Includes deletion protection and a modern TLS security policy.

## When to Use

- Public-facing APIs or services that require HTTPS at the edge
- gRPC workloads where TLS termination at the NLB simplifies certificate management
- Any Layer 4 workload that needs TLS termination without HTTP-level routing (e.g., ALB path-based routing)
- Scenarios where you want TLS offload without the overhead of ALB (lower latency, static IP support)

## Key Configuration Choices

- **Internet-facing** (`internal: false`) — NLB is created in public subnets; receives a public DNS name and is reachable from the internet
- **TLS termination** — Listener uses `protocol: TLS` with an ACM certificate; NLB decrypts and forwards plaintext TCP to targets
- **Deletion protection** (`deleteProtectionEnabled: true`) — Prevents accidental deletion of the load balancer
- **Modern TLS policy** (`ELBSecurityPolicy-TLS13-1-2-2021-06`) — Enables TLS 1.3 and 1.2; disables older protocols
- **Two public subnets** — Required for cross-AZ high availability

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<nlb-name>` | Unique name for the NLB (lowercase, alphanumeric, hyphens) | Choose a descriptive name (e.g., `api-public`, `grpc-gateway`) |
| `<public-subnet-id-az1>` | Public subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<public-subnet-id-az2>` | Public subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<acm-certificate-arn>` | ARN of an ACM certificate covering your domain | AWS ACM console or `AwsCertManagerCert` status outputs |
| `<application-port>` | Port on targets to forward decrypted traffic to (e.g., 8080, 9090) | Your application's listening port |

## Common Additions

- Add `dns` section with `route53ZoneId` and `hostnames` for automatic Route53 alias records
- Add `allocationId` to subnet mappings for static public IPs (see **03-static-ip-production**)
- Add `securityGroups` to restrict inbound traffic
- Add `crossZoneLoadBalancingEnabled: true` if targets are unevenly distributed across AZs
- Use `valueFrom` to reference `AwsCertManagerCert` for certificate ARN

## Related Presets

- **01-tcp-internal** — Use for internal-only microservice communication
- **03-static-ip-production** — Use when you need static IPs, DNS, HTTP health checks, and full production hardening

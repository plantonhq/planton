# AwsNetworkLoadBalancer

The **AwsNetworkLoadBalancer** resource provides a standardized way to provision and manage AWS Network Load Balancers (NLBs) through OpenMCF. NLBs operate at Layer 4 (TCP/UDP/TLS), offering ultra-low latency, static IP addresses per Availability Zone, and the ability to handle millions of connections per second.

## NLB vs ALB: When to Use Which

| Aspect | Network Load Balancer (NLB) | Application Load Balancer (ALB) |
|--------|----------------------------|----------------------------------|
| **OSI Layer** | Layer 4 (Transport) | Layer 7 (Application) |
| **Protocols** | TCP, UDP, TLS, TCP_UDP | HTTP, HTTPS, WebSockets |
| **Static IPs** | Yes — Elastic IP per AZ | No — dynamic DNS only |
| **Routing** | Port and protocol only | Path, host, headers, query string |
| **Latency** | Ultra-low (microseconds) | Slightly higher |
| **Security Groups** | Optional | Effectively required |
| **TLS Termination** | At listener level | At HTTP level |
| **Use Case** | Non-HTTP traffic, static IP allowlisting, TCP passthrough | Web apps, APIs, microservices |

**Choose NLB when** you need static IPs for allowlisting, TCP/UDP passthrough, TLS termination at Layer 4, or ultra-low latency for non-HTTP workloads. **Choose ALB when** you need HTTP/HTTPS routing, path-based routing, or Lambda/container targets with content-based rules.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **subnetMappings**: List of subnet mappings defining where NLB nodes are placed. Each mapping has a `subnetId` and optionally an `allocationId` (Elastic IP) for internet-facing NLBs or `privateIpv4Address` for internal NLBs. At least one mapping required; AWS recommends two for high availability.
- **listeners**: List of listeners, each with a `name`, `port`, `protocol` (TCP, UDP, TLS, TCP_UDP), and inline `targetGroup`. Each listener forwards traffic to its target group. At least one listener required.
- **internal**: When `true`, creates an internal NLB accessible only within the VPC. When `false` (default), creates an internet-facing NLB.
- **targetGroup** (per listener): Defines `port`, `protocol`, and optionally `healthCheck`, `deregistrationDelaySeconds`, `preserveClientIp`, `proxyProtocolV2`, `connectionTermination`, `stickinessEnabled`.

### Advanced Fields (20% Use Case)

- **securityGroups**: Optional list of security group IDs. Unlike ALB, NLBs can run without security groups. Once attached, at least one must remain (cannot fully remove).
- **deleteProtectionEnabled**: Prevents accidental deletion when enabled.
- **crossZoneLoadBalancingEnabled**: Distributes traffic across all AZs. Default `false` for NLB (unlike ALB). Enable when target distribution is uneven.
- **ipAddressType**: `ipv4` (default) or `dualstack`.
- **dnsRecordClientRoutingPolicy**: `any_availability_zone` (default), `availability_zone_affinity`, or `partial_availability_zone_affinity` — controls how DNS routes clients to NLB nodes.
- **dns**: Route53 configuration with `enabled`, `route53ZoneId`, and `hostnames` for alias records.
- **subnetMapping.allocationId**: Elastic IP allocation ID for static public IP per AZ (internet-facing only).
- **listener.tls**: Required when `protocol` is TLS. Includes `certificateArn` and optional `sslPolicy`.
- **listener.tcpIdleTimeoutSeconds**: TCP idle timeout (60–6000s). Only for TCP protocol.
- **listener.alpnPolicy**: For TLS listeners — HTTP1Only, HTTP2Only, HTTP2Optional, HTTP2Preferred, None.
- **targetGroup.targetType**: `instance`, `ip`, or `alb` (for NLB-in-front-of-ALB pattern).

## Stack Outputs

After provisioning, the AwsNetworkLoadBalancer resource provides:

- **load_balancer_arn**: ARN of the Network Load Balancer.
- **load_balancer_name**: Name assigned to the NLB.
- **load_balancer_dns_name**: DNS name assigned by AWS (e.g., `my-nlb-abc123.elb.us-east-1.amazonaws.com`).
- **load_balancer_hosted_zone_id**: Route53 hosted zone ID for the NLB's DNS name (for alias records).
- **listener_arns**: Map of listener name → listener ARN (e.g., `status.outputs.listener_arns.tcp-443`).
- **target_group_arns**: Map of listener name → target group ARN. Primary output for ECS, EKS, or auto-scaling groups to register targets (e.g., `status.outputs.target_group_arns.tcp-443`).

## How It Works

When you define an AwsNetworkLoadBalancer resource, OpenMCF:

1. **Creates the NLB**: Provisions a Network Load Balancer in the specified subnets with optional Elastic IPs per AZ.
2. **Configures subnet mappings**: Places NLB nodes in each subnet; for internet-facing NLBs with `allocationId`, assigns static public IPs.
3. **Creates listeners and target groups**: Each listener has an inline target group. NLB only supports forward actions, so every listener must have a target group.
4. **Registers targets externally**: Targets (EC2 instances, IPs, or ALBs) are registered by downstream services (ECS, EKS, auto-scaling groups) using `status.outputs.target_group_arns.{listener_name}`.
5. **Manages DNS (optional)**: If DNS is enabled, creates Route53 alias A records for the specified hostnames.
6. **Applies advanced options**: Security groups, cross-zone load balancing, deletion protection, and target group options as configured.

The resource uses Pulumi or Terraform under the hood to provision all necessary AWS resources consistently and reliably.

## Use Cases

### Static IP Allowlisting

Use Elastic IPs per subnet mapping for internet-facing NLBs. Partners, firewalls, or legacy systems can allowlist these static IPs without DNS changes.

### TCP Passthrough

Forward raw TCP/UDP traffic without HTTP inspection. Ideal for databases, game servers, IoT protocols, or any non-HTTP workload.

### TLS Termination at NLB

Use a TLS listener with an ACM certificate. The NLB terminates TLS and forwards plaintext TCP to targets, offloading certificate management from application servers.

### NLB in Front of ALB

Register an ALB as a target (`targetType: alb`). The NLB provides static IPs and Layer 4 entry; the ALB handles Layer 7 routing, path-based rules, and SSL.

### Internal Microservices

Deploy an internal NLB for service-to-service traffic. Optionally pin private IPs per subnet with `privateIpv4Address`.

## References

- [AWS Network Load Balancer Documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/introduction.html)
- [NLB vs ALB: Choosing the Right Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/network-load-balancers.html)
- [Target Groups for Your Network Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html)
- [Elastic IP Addresses for NLB](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/network-load-balancers.html#nlb-enable-static-ip)

# AWS Network Load Balancer

Deploys an AWS Network Load Balancer with bundled listeners and target groups, optional static IP addresses via Elastic IP allocation, TLS termination, and Route53 DNS record management. The component operates at Layer 4 (TCP/UDP/TLS) and requires at least one subnet mapping and one listener.

## What Gets Created

When you deploy an AwsNetworkLoadBalancer resource, Planton provisions:

- **Network Load Balancer** — an `aws_lb` resource of type `network`, placed in the specified subnets via subnet mappings with attached security groups (optional), cross-zone load balancing configuration, and DNS client routing policy
- **Target Groups** — one `aws_lb_target_group` per listener, configured with the specified protocol, port, target type, health check, deregistration delay, and connection settings (preserve client IP, Proxy Protocol v2, connection termination, source IP stickiness)
- **Listeners** — one `aws_lb_listener` per spec listener, forwarding traffic to its corresponding target group with optional TLS configuration (certificate, SSL policy, ALPN policy) and TCP idle timeout
- **Route53 A Records** — created only when DNS is enabled, one alias record per hostname pointing to the NLB's DNS name with target health evaluation enabled

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **At least one subnet** in a VPC (public subnets for internet-facing, private for internal). AWS recommends two or more across Availability Zones for high availability
- **An Elastic IP allocation ID** per subnet if you need static public IP addresses (internet-facing NLBs only)
- **An ACM certificate ARN** if enabling TLS termination on a listener
- **A Route53 hosted zone** if enabling DNS management
- **A security group** if you want to filter inbound traffic (optional for NLB, unlike ALB)

## Quick Start

Create a file `nlb.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: my-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsNetworkLoadBalancer.my-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-0a1b2c3d4e5f00001
    - subnetId: subnet-0a1b2c3d4e5f00002
  listeners:
    - name: tcp-80
      port: 80
      protocol: TCP
      targetGroup:
        port: 80
        protocol: TCP
```

Deploy:

```shell
planton apply -f nlb.yaml
```

This creates an internet-facing NLB with a TCP listener on port 80 forwarding to a target group on port 80, deployed across two subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the Network Load Balancer will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `subnetMappings` | `object[]` | Subnet mappings defining where NLB nodes are placed and optional static IPs | Minimum 1 item required |
| `subnetMappings[].subnetId` | `string` | Subnet ID for the NLB node. Can reference an AwsVpc resource via `valueFrom` | Required |
| `listeners` | `object[]` | Listener configurations with inline target groups | Minimum 1 item required |
| `listeners[].name` | `string` | Unique name for the listener, used as key in output maps | Required. Lowercase alphanumeric and hyphens, starts with letter, max 63 chars |
| `listeners[].port` | `int` | Port the listener accepts traffic on | Required. Range: 1–65535 |
| `listeners[].protocol` | `string` | Listener protocol | Required. One of: `TCP`, `UDP`, `TLS`, `TCP_UDP` |
| `listeners[].targetGroup` | `object` | Target group configuration for this listener | Required |
| `listeners[].targetGroup.port` | `int` | Port to route traffic to on targets | Required. Range: 1–65535 |
| `listeners[].targetGroup.protocol` | `string` | Protocol for communication with targets | Required. One of: `TCP`, `UDP`, `TLS`, `TCP_UDP` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subnetMappings[].allocationId` | `string` | — | Elastic IP allocation ID for a static public IP on this NLB node. Internet-facing only. |
| `subnetMappings[].privateIpv4Address` | `string` | — | Specific private IP for the NLB node. Internal NLBs only. |
| `securityGroups` | `string[]` | `[]` | Security group IDs. Optional for NLB. Can reference AwsSecurityGroup via `valueFrom`. |
| `internal` | `bool` | `false` | When `true`, creates an internal NLB accessible only within the VPC. |
| `deleteProtectionEnabled` | `bool` | `false` | Prevents accidental deletion when enabled. |
| `crossZoneLoadBalancingEnabled` | `bool` | `false` | Distributes traffic evenly across all targets in all AZs. |
| `ipAddressType` | `string` | `ipv4` | IP address type. One of: `ipv4`, `dualstack`. |
| `dnsRecordClientRoutingPolicy` | `string` | `any_availability_zone` | DNS routing policy. One of: `any_availability_zone`, `availability_zone_affinity`, `partial_availability_zone_affinity`. |
| `listeners[].tls` | `object` | — | TLS configuration. Required when protocol is `TLS`. |
| `listeners[].tls.certificateArn` | `string` | — | ACM certificate ARN for TLS termination. Can reference AwsCertManagerCert via `valueFrom`. |
| `listeners[].tls.sslPolicy` | `string` | AWS default | TLS security policy (e.g., `ELBSecurityPolicy-TLS13-1-2-2021-06`). |
| `listeners[].tcpIdleTimeoutSeconds` | `int` | `350` | TCP idle timeout. Only valid for TCP protocol. Range: 60–6000. |
| `listeners[].alpnPolicy` | `string` | — | ALPN policy. Only valid for TLS protocol. One of: `HTTP1Only`, `HTTP2Only`, `HTTP2Optional`, `HTTP2Preferred`, `None`. |
| `listeners[].targetGroup.targetType` | `string` | `instance` | Target type. One of: `instance`, `ip`, `alb`. |
| `listeners[].targetGroup.healthCheck` | `object` | TCP on traffic-port | Health check configuration. |
| `listeners[].targetGroup.healthCheck.protocol` | `string` | `TCP` | Health check protocol. One of: `TCP`, `HTTP`, `HTTPS`. |
| `listeners[].targetGroup.healthCheck.port` | `string` | `traffic-port` | Health check port. `traffic-port` or 1–65535. |
| `listeners[].targetGroup.healthCheck.path` | `string` | `/` | Path for HTTP/HTTPS health checks. Required when protocol is HTTP or HTTPS. |
| `listeners[].targetGroup.healthCheck.healthyThreshold` | `int` | `3` | Consecutive successes before healthy. Range: 2–10. |
| `listeners[].targetGroup.healthCheck.unhealthyThreshold` | `int` | `3` | Consecutive failures before unhealthy. Range: 2–10. |
| `listeners[].targetGroup.healthCheck.intervalSeconds` | `int` | `30` | Seconds between health checks. Range: 5–300. |
| `listeners[].targetGroup.healthCheck.timeoutSeconds` | `int` | `10` | Seconds before timeout. Range: 2–120. |
| `listeners[].targetGroup.healthCheck.matcher` | `string` | `200-399` | HTTP response codes for healthy. Only for HTTP/HTTPS. |
| `listeners[].targetGroup.deregistrationDelaySeconds` | `int` | `300` | Seconds to wait before deregistering draining targets. Range: 0–3600. |
| `listeners[].targetGroup.preserveClientIp` | `bool` | `false` | Preserve original client IP in the IP header. |
| `listeners[].targetGroup.proxyProtocolV2` | `bool` | `false` | Enable Proxy Protocol v2 header with client metadata. |
| `listeners[].targetGroup.connectionTermination` | `bool` | `false` | Terminate connections on deregistration delay expiry. |
| `listeners[].targetGroup.stickinessEnabled` | `bool` | `false` | Source-IP-based sticky sessions. |
| `dns.enabled` | `bool` | `false` | Enable Route53 DNS record creation. |
| `dns.route53ZoneId` | `string` | — | Route53 hosted zone ID. Can reference AwsRoute53Zone via `valueFrom`. |
| `dns.hostnames` | `string[]` | `[]` | Domain names to point to the NLB. Must be unique. |

## Examples

### Internal TCP NLB

An NLB accessible only within the VPC for internal microservice communication:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: internal-tcp-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsNetworkLoadBalancer.internal-tcp-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-private-az1
    - subnetId: subnet-private-az2
  internal: true
  listeners:
    - name: tcp-8080
      port: 8080
      protocol: TCP
      targetGroup:
        port: 8080
        protocol: TCP
        targetType: ip
```

### TLS Termination with ACM Certificate

Internet-facing NLB that terminates TLS and forwards plaintext TCP to application servers:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: tls-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNetworkLoadBalancer.tls-nlb
spec:
  region: us-east-1
  subnetMappings:
    - subnetId: subnet-public-az1
    - subnetId: subnet-public-az2
  deleteProtectionEnabled: true
  listeners:
    - name: tls-443
      port: 443
      protocol: TLS
      tls:
        certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abc-12345
        sslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
      targetGroup:
        port: 8080
        protocol: TCP
        healthCheck:
          protocol: HTTP
          path: /health
          intervalSeconds: 10
          healthyThreshold: 3
```

### Static IPs with Elastic IPs

Internet-facing NLB with static public IPs for firewall allowlisting:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: static-ip-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNetworkLoadBalancer.static-ip-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-public-az1
      allocationId: eipalloc-abc123
    - subnetId: subnet-public-az2
      allocationId: eipalloc-def456
  crossZoneLoadBalancingEnabled: true
  listeners:
    - name: tcp-443
      port: 443
      protocol: TCP
      targetGroup:
        port: 443
        protocol: TCP
        preserveClientIp: true
```

### Full-Featured with Foreign Key References

Production NLB referencing other Planton-managed resources:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: prod-nlb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsNetworkLoadBalancer.prod-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          field: status.outputs.public_subnets[0].id
      allocationId: eipalloc-prod-az1
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          field: status.outputs.public_subnets[1].id
      allocationId: eipalloc-prod-az2
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: nlb-sg
        field: status.outputs.security_group_id
  deleteProtectionEnabled: true
  crossZoneLoadBalancingEnabled: true
  listeners:
    - name: tls-443
      port: 443
      protocol: TLS
      tls:
        certificateArn:
          valueFrom:
            kind: AwsCertManagerCert
            name: prod-cert
            field: status.outputs.cert_arn
        sslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
      alpnPolicy: HTTP2Preferred
      targetGroup:
        port: 8443
        protocol: TCP
        targetType: ip
        deregistrationDelaySeconds: 60
        preserveClientIp: true
        connectionTermination: true
        healthCheck:
          protocol: HTTPS
          path: /healthz
          healthyThreshold: 3
          unhealthyThreshold: 3
          intervalSeconds: 10
          timeoutSeconds: 6
          matcher: "200"
  dns:
    enabled: true
    route53ZoneId:
      valueFrom:
        kind: AwsRoute53Zone
        name: prod-zone
        field: status.outputs.zone_id
    hostnames:
      - api.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_arn` | `string` | ARN of the Network Load Balancer |
| `load_balancer_name` | `string` | Name assigned to the NLB (may differ from `metadata.name`) |
| `load_balancer_dns_name` | `string` | DNS name assigned by AWS (e.g., `my-nlb-abc123.elb.us-east-1.amazonaws.com`) |
| `load_balancer_hosted_zone_id` | `string` | Route53 hosted zone ID for the NLB's DNS entry, used for creating alias records |
| `listener_arns` | `map<string,string>` | Map of listener name to listener ARN (e.g., `listener_arns.tls-443`) |
| `target_group_arns` | `map<string,string>` | Map of listener name to target group ARN (e.g., `target_group_arns.tls-443`). Used by downstream services to register targets |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the subnets for NLB placement via subnet mappings
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — optional inbound traffic filtering for the NLB
- [AwsRoute53Zone](/docs/catalog/aws/awsroute53zone) — hosts the DNS zone for alias records
- [AwsCertManagerCert](/docs/catalog/aws/awscertmanagercert) — provides the ACM certificate for TLS termination on listeners
- [AwsAlb](/docs/catalog/aws/awsalb) — Application Load Balancer for Layer 7 (HTTP/HTTPS) routing; can be used as a target type for NLB-in-front-of-ALB patterns

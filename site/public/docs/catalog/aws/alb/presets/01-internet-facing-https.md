---
title: "Internet-Facing HTTPS ALB"
description: "This preset creates an internet-facing Application Load Balancer with HTTPS termination and Route53 DNS management. It enables deletion protection and uses the AWS-recommended 60-second idle timeout...."
type: "preset"
rank: "01"
presetSlug: "01-internet-facing-https"
componentSlug: "alb"
componentTitle: "ALB"
provider: "aws"
icon: "package"
order: 1
---

# Internet-Facing HTTPS ALB

This preset creates an internet-facing Application Load Balancer with HTTPS termination and Route53 DNS management. It enables deletion protection and uses the AWS-recommended 60-second idle timeout. This is the most common production ALB configuration.

## When to Use

- Public-facing web applications or APIs that need HTTPS
- Production workloads requiring automatic DNS management via Route53
- Standard HTTP/HTTPS load balancing without specialized protocol requirements (gRPC, WebSocket)

## Key Configuration Choices

- **Internet-facing** (`internal: false`) -- accessible from the public internet via public subnets
- **HTTPS enabled** (`ssl.enabled: true`) -- terminates TLS at the ALB using an ACM certificate
- **DNS management** (`dns.enabled: true`) -- automatically creates Route53 alias records pointing to the ALB
- **Deletion protection** (`deleteProtectionEnabled: true`) -- prevents accidental deletion of a production load balancer
- **60-second idle timeout** (`idleTimeoutSeconds: 60`) -- AWS recommended default, suitable for most HTTP/HTTPS traffic
- **Two subnets across AZs** -- minimum required by AWS for cross-AZ high availability

## Placeholders to Replace

| Placeholder                | Description                                                 | Where to Find                                          |
| -------------------------- | ----------------------------------------------------------- | ------------------------------------------------------ |
| `<public-subnet-id-az1>`   | Public subnet in the first Availability Zone                | AWS VPC console or `AwsVpc` status outputs             |
| `<public-subnet-id-az2>`   | Public subnet in the second Availability Zone               | AWS VPC console or `AwsVpc` status outputs             |
| `<alb-security-group-id>`  | Security group allowing inbound traffic on ports 80 and 443 | AWS EC2 console or `AwsSecurityGroup` status outputs   |
| `<acm-certificate-arn>`    | ARN of an ACM certificate covering your domain              | AWS ACM console or `AwsCertManagerCert` status outputs |
| `<route53-hosted-zone-id>` | ID of the Route53 hosted zone for your domain               | AWS Route53 console or `AwsRoute53Zone` status outputs |
| `<your-domain.com>`        | The domain name that should point to this ALB               | Your domain registrar or DNS provider                  |

## Related Presets

- **02-internal-http** -- Use instead when the ALB should only be accessible within the VPC (e.g., for internal microservice communication)

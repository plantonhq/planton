# AwsAlb

The **AwsAlb** resource provides a standardized way to provision and manage AWS Application Load Balancers (ALBs) through Planton. It simplifies ALB configuration by focusing on essential fields while providing flexibility for common networking, SSL, and DNS requirements.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **subnets**: List of subnet IDs where the ALB will be created (minimum 2 required for high availability across multiple Availability Zones). Use private subnets for internal ALBs or public subnets for internet-facing ALBs.
- **security_groups**: List of security group IDs to attach to the ALB for controlling inbound/outbound traffic.
- **internal**: Boolean flag indicating whether the ALB is internal (true) or internet-facing (false). Defaults to false for internet-facing.
- **ssl.enabled**: Boolean to enable SSL/TLS termination on the ALB.
- **ssl.certificate_arn**: ARN of the AWS Certificate Manager certificate to use for HTTPS listeners (required when ssl.enabled is true).

### Advanced Fields (20% Use Case)

- **delete_protection_enabled**: Boolean to enable deletion protection, preventing accidental deletion of the ALB.
- **idle_timeout_seconds**: Connection idle timeout in seconds (AWS default is 60 seconds if omitted).
- **dns.enabled**: Boolean to enable automatic Route53 DNS record creation for the ALB.
- **dns.route53_zone_id**: Route53 Hosted Zone ID where DNS records will be created (required when dns.enabled is true).
- **dns.hostnames**: List of domain names (e.g., ["app.example.com", "api.example.com"]) that will point to the ALB's DNS name.

## Stack Outputs

After provisioning, the AwsAlb resource provides the following outputs:

- **alb_arn**: The ARN of the created Application Load Balancer.
- **alb_dns_name**: The DNS name assigned by AWS for accessing the ALB.
- **alb_zone_id**: The Route53 zone ID of the ALB (useful for creating alias records).

## How It Works

When you define an AwsAlb resource, Planton:

1. **Creates the ALB**: Provisions an Application Load Balancer in the specified subnets with the configured security groups.
2. **Configures Networking**: Sets up the ALB as either internet-facing or internal based on the `internal` flag.
3. **Applies Security**: Attaches security groups to control traffic flow to and from the ALB.
4. **Enables SSL (Optional)**: If SSL is enabled, configures HTTPS listeners with the specified certificate.
5. **Manages DNS (Optional)**: If DNS is enabled, creates Route53 A/AAAA records pointing the specified hostnames to the ALB.
6. **Sets Advanced Options**: Applies deletion protection and idle timeout settings as configured.

The resource uses Pulumi or Terraform under the hood (depending on your stack configuration) to provision all necessary AWS resources consistently and reliably.

## Use Cases

### Internet-Facing Web Application
Deploy a public-facing ALB with SSL termination for a web application accessible from the internet.

### Internal Microservices
Create an internal ALB for routing traffic between microservices within a private VPC, without exposing services to the internet.

### Multi-Domain Hosting
Use DNS configuration to map multiple domain names to a single ALB, ideal for hosting multiple applications or environments.

### High Availability Architecture
Leverage multi-subnet deployment (required minimum of 2) to ensure ALB availability across multiple Availability Zones.

## References

- [AWS Application Load Balancer Documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html)
- [AWS ALB Best Practices](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/application-load-balancer-best-practices.html)
- [AWS Certificate Manager](https://docs.aws.amazon.com/acm/)
- [Route53 DNS Management](https://docs.aws.amazon.com/route53/)

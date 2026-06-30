# AwsCertManagerCert

The **AwsCertManagerCert** resource provides a standardized way to provision and manage SSL/TLS certificates through AWS Certificate Manager (ACM) using DNS validation. It simplifies certificate management by automating DNS validation records in Route53 and supporting both single and multi-domain certificates.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **primary_domain_name**: The main domain name for the certificate (e.g., "example.com" or "*.example.com" for wildcard). This is required and serves as the primary Subject name on the certificate.
- **route53_hosted_zone_id**: The Route53 Hosted Zone ID where DNS validation records will be automatically created. Must be a public hosted zone that matches the domain names. This is required for DNS validation.

### Advanced Fields (20% Use Case)

- **alternate_domain_names**: Optional list of Subject Alternative Names (SANs) to include on the certificate. Each domain must be unique and follow the same format as the primary domain (supports wildcards). The primary domain should not be duplicated in this list.
- **validation_method**: Method for domain ownership verification. Options are "DNS" (default and recommended) or "EMAIL". DNS validation is preferred as it automates the validation process through Route53.

## Stack Outputs

After provisioning, the AwsCertManagerCert resource provides the following outputs:

- **cert_arn**: The Amazon Resource Name (ARN) of the created certificate, used to reference the certificate in other AWS resources like ALBs, CloudFront distributions, or API Gateways.
- **validation_status**: Current status of the certificate validation (e.g., "PENDING_VALIDATION", "ISSUED").
- **domain_validation_records**: Details of the DNS validation records created in Route53.

## How It Works

When you define an AwsCertManagerCert resource, Planton:

1. **Requests Certificate**: Creates an SSL/TLS certificate in AWS Certificate Manager for the specified primary domain and any alternate domains.
2. **Validates Ownership**: Automatically creates DNS validation records (CNAME records) in the specified Route53 hosted zone to prove domain ownership.
3. **Waits for Issuance**: Monitors the certificate status until AWS validates ownership and issues the certificate.
4. **Manages Lifecycle**: Handles certificate renewal automatically—ACM certificates are valid for 13 months and renew automatically if DNS validation records remain in place.

The DNS validation approach is preferred because it's automated, doesn't require manual intervention, and supports wildcard certificates.

## Use Cases

### Wildcard Certificate for Subdomains
Create a single wildcard certificate (*.example.com) to secure all subdomains under your primary domain.

### Multi-Domain Certificate
Issue one certificate covering multiple distinct domains (e.g., example.com, api.example.com, www.example.com) to simplify certificate management.

### CloudFront and ALB SSL
Provision certificates for use with CloudFront distributions (requires us-east-1 region) or Application Load Balancers in any region.

### Automated Certificate Management
Leverage ACM's automatic renewal and Route53 DNS validation for zero-touch certificate lifecycle management.

## Important Notes

### DNS Validation Requirements
- The Route53 hosted zone must be publicly accessible and match the domain names.
- DNS validation records are automatically created as CNAME records in the hosted zone.
- Wildcard certificates (*.example.com) require DNS validation; email validation is not supported for wildcards.

### Regional Considerations
- For CloudFront distributions, certificates must be created in the us-east-1 region.
- For other services (ALB, API Gateway), create certificates in the same region as the service.

### Certificate Renewal
- ACM certificates automatically renew every 13 months if DNS validation records remain in Route53.
- No action required for renewal as long as validation records are not deleted.

## References

- [AWS Certificate Manager Documentation](https://docs.aws.amazon.com/acm/)
- [DNS Validation for ACM Certificates](https://docs.aws.amazon.com/acm/latest/userguide/dns-validation.html)
- [ACM Certificate Characteristics](https://docs.aws.amazon.com/acm/latest/userguide/acm-certificate.html)
- [Route53 DNS Management](https://docs.aws.amazon.com/route53/)

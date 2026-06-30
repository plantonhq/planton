# AWS Certificate Manager: From Manual Burden to Automated Elegance

## Introduction

For years, managing SSL/TLS certificates was a necessary evil—a routine that involved setting calendar reminders, maintaining secure storage for private keys, orchestrating complex renewal workflows, and praying that nothing would break during the annual scramble. The emergence of Let's Encrypt democratized certificate acquisition, but it still required running and maintaining ACME clients on your infrastructure. Then AWS Certificate Manager (ACM) arrived with a compelling proposition: **completely free, fully automated certificate lifecycle management**—but only if you're willing to stay within the AWS ecosystem.

ACM is not just a certificate authority; it's a deeply integrated service that handles provisioning, renewal, and deployment of SSL/TLS certificates for AWS resources like Application Load Balancers, CloudFront distributions, and API Gateway custom domains. Its promise of "set-it-and-forget-it" renewal has made it the default choice for AWS-native applications. However, realizing this promise requires understanding a critical truth: **the method you choose to validate domain ownership determines whether you achieve true automation or inherit a ticking time bomb**.

This document explores the landscape of ACM deployment approaches, from manual console workflows to sophisticated Infrastructure as Code (IaC) patterns. It examines why **DNS validation is non-negotiable** for production systems, how different IaC tools abstract the validation workflow, and the production patterns that separate robust implementations from brittle ones. Most importantly, it explains how Planton's `AwsCertManagerCert` resource delivers the simplest possible API while handling the complex orchestration beneath the surface.

## The Evolution of ACM Deployment: From Console to Code

The journey from manual certificate management to fully automated IaC reveals a progression in both automation sophistication and production readiness. Understanding this evolution helps clarify why certain approaches are superior and what trade-offs exist at each level.

### Level 0: Email Validation – The Production Anti-Pattern

**What it is:** When requesting a certificate, ACM can send validation emails to administrative addresses for the domain (like `admin@example.com` or `webmaster@example.com`). A human must receive this email and click an approval link to prove domain ownership.

**Why it exists:** Email validation predates DNS validation and was designed for scenarios where users might not have direct access to DNS management. It feels "simple" on the surface—just click a link in an email.

**Why it fails in production:** This approach is fundamentally incompatible with Infrastructure as Code and automated deployments:

1. **Initial Issuance Requires Human Intervention:** Your IaC pipeline will stall, waiting for someone to check their inbox and click a link. This breaks continuous deployment.

2. **Renewal is Also Manual:** ACM sends renewal approval emails 45 days before expiration. If this email is missed, filtered to spam, or sent to an unmaintained inbox, the certificate **will expire**, causing an immediate production outage.

3. **No Migration Path:** A certificate validated via email cannot be switched to DNS validation. You must request a new certificate, update all associated resources, and destroy the old one—a risky, error-prone process.

**Verdict:** Email validation is acceptable only for temporary, non-production testing. Using it in production guarantees a future incident and should be considered a critical misconfiguration.

### Level 1: Manual Console Workflow – Learning the Fundamentals

**What it is:** Using the AWS Console to request a certificate, select DNS validation, and manually create the CNAME records that ACM generates. If the domain is in Route 53, the console offers a convenient "Create records in Route 53" button that automates the DNS record creation.

**What it teaches:** This workflow makes the certificate lifecycle explicit. You see that ACM generates a unique CNAME record for each domain (including wildcards and SANs), that these records must be added to the **public** DNS zone, and that the certificate doesn't issue until ACM can query DNS and verify the records exist.

**Why it's limited:** Manual workflows don't scale. They're not repeatable, not version-controlled, and not auditable. Every new certificate requires the same manual steps, and there's no record of what was configured or why. More critically, the CNAME validation records **must persist forever**—not just until the certificate issues. ACM checks for these records during its automated renewal process 60 days before expiration. Deleting them breaks renewal and causes certificate expiration.

**Verdict:** Essential for learning and one-off testing. Unacceptable for production environments where repeatability, auditability, and disaster recovery depend on infrastructure being defined as code.

### Level 2: Scripted Deployment – The Imperative Approach

**What it is:** Writing custom scripts using the AWS CLI or SDKs (like Boto3) to automate the certificate request, CNAME extraction, DNS record creation, and validation waiting. This typically involves:

1. Calling `aws acm request-certificate` with DNS validation specified
2. Polling `aws acm describe-certificate` to wait for the `DomainValidationOptions` array to populate
3. Parsing this array to extract the CNAME name and value for each domain
4. Calling `aws route53 change-resource-record-sets` to create the DNS records
5. Using `aws acm wait certificate-validated` to block until the certificate status becomes "ISSUED"

**Why it's better:** This approach is repeatable and can be integrated into CI/CD pipelines. It represents a genuine automation effort.

**Why it's still problematic:** Imperative scripts are fragile and hard to maintain. They require explicit error handling, state management, and retry logic. If the script fails mid-execution, there's no way to determine what state was created. The script doesn't track dependencies between resources or handle updates gracefully. Most critically, these scripts often lack idempotency—running them twice may create duplicate records or fail in unexpected ways.

**Verdict:** A stepping stone toward true Infrastructure as Code, but not production-grade. Scripts are difficult to test, hard to review, and lack the declarative guarantees and state management that make infrastructure reliable.

### Level 3: Declarative IaC – The Production Standard

**What it is:** Using declarative Infrastructure as Code tools (Terraform, Pulumi, CloudFormation, AWS CDK) to define the desired state of the certificate and its validation infrastructure. The tool handles the imperative orchestration, state tracking, and idempotent operations.

**Why it's superior:** Declarative IaC solves the fundamental problems of scripts:

- **State Management:** The tool tracks what resources exist, what state they're in, and what changes are needed.
- **Idempotency:** Running the same configuration multiple times produces the same result.
- **Dependency Management:** The tool understands that the load balancer listener depends on the validated certificate, not just the certificate request.
- **Auditability and Review:** Infrastructure changes go through code review and version control.
- **Disaster Recovery:** The entire infrastructure can be recreated from the IaC repository.

**The Critical Split:** Here, the abstraction model diverges. Different IaC tools take fundamentally different approaches to the validation workflow, revealing a philosophical divide between AWS-native tools and multi-cloud platforms.

## The Great IaC Divide: Implicit vs. Explicit Validation

The most interesting technical aspect of ACM automation is how different IaC tools abstract—or choose not to abstract—the validation workflow. This is not just a matter of syntax; it reflects a fundamental difference in philosophy about what abstraction means.

### The AWS-Native Model: Implicit Integration (CloudFormation & CDK)

CloudFormation and the AWS CDK, being AWS's own IaC tools, take full advantage of their deep integration with Route 53 to provide the simplest possible user experience.

**How it works:**

- **Single Resource:** A single `AWS::CertificateManager::Certificate` resource (in CloudFormation) or `DnsValidatedCertificate` construct (in CDK) handles the entire workflow.
- **The Magic Property:** The `DomainValidationOptions` property (CloudFormation) or `hostedZone` parameter (CDK) is where the abstraction happens. The user maps each domain name to its Route 53 Hosted Zone ID.
- **Hidden Orchestration:** When CloudFormation processes this resource, it internally:
  1. Requests the certificate from ACM
  2. Receives the CNAME validation data
  3. **Implicitly creates** the CNAME records in the specified hosted zones
  4. **Implicitly waits** for ACM to validate and issue the certificate
  
This all happens within a single logical resource. The user never sees the CNAME records as separate infrastructure—they're an implementation detail managed by the CloudFormation service backend.

**Example (CloudFormation YAML):**

```yaml
Resources:
  MyCertificate:
    Type: AWS::CertificateManager::Certificate
    Properties:
      DomainName: example.com
      SubjectAlternativeNames:
        - "*.example.com"
      ValidationMethod: DNS
      DomainValidationOptions:
        - DomainName: example.com
          HostedZoneId: Z2KZ5YTUFZNC7H
        - DomainName: "*.example.com"
          HostedZoneId: Z2KZ5YTUFZNC7H
```

**Strengths:**

- **Minimal Boilerplate:** The user expresses intent ("I want a cert for example.com, validated via this hosted zone") and the tool handles the rest.
- **Fewer Moving Parts:** No need to manage intermediate resources or wire up dependencies.
- **AWS-Optimized:** Perfect for teams fully committed to AWS infrastructure.

**Limitations:**

- **Route 53 Lock-In:** This implicit automation only works if DNS is managed in Route 53. If your domain is managed by Cloudflare, Google Domains, or any other provider, you're forced to manage the validation records manually outside the IaC stack.
- **Less Flexibility:** You can't inspect or customize the validation records. The abstraction is opaque.

### The Multi-Cloud Model: Explicit Composition (Terraform & Pulumi)

Terraform, Pulumi, and OpenTofu take a different philosophical stance. They're designed to work across multiple cloud providers, so they avoid AWS-specific "magic" and instead favor explicit, composable resources.

**How it works:**

Automation requires orchestrating **three separate, explicit resources** in a specific dependency chain:

1. **`aws_acm_certificate`** (Terraform) / `aws.acm.Certificate` (Pulumi)
   - Requests the certificate with DNS validation
   - Exports a crucial attribute: `domain_validation_options`
   - This is a list of objects, one per domain, containing the CNAME name and value

2. **`aws_route53_record`** (Terraform) / `aws.route53.Record` (Pulumi)
   - Creates the DNS validation records
   - Uses a `for_each` loop or similar construct to iterate over `domain_validation_options`
   - Creates one CNAME record per domain in the certificate request

3. **`aws_acm_certificate_validation`** (Terraform) / `aws.acm.CertificateValidation` (Pulumi)
   - This is a "waiter" resource—it doesn't create anything in AWS
   - It creates an explicit dependency that blocks downstream resources
   - Takes the certificate ARN (from resource #1) and the FQDNs of the validation records (from resource #2)
   - Completes only when ACM reports the certificate as "ISSUED"

**Critical Insight:** Downstream resources (like a load balancer listener) must depend on the `certificate_arn` output from the **validation resource** (resource #3), not the certificate request (resource #1). This ensures the certificate is fully issued before it's attached to any service.

**Example (Terraform HCL):**

```hcl
# 1. Request the certificate
resource "aws_acm_certificate" "cert" {
  domain_name               = "example.com"
  subject_alternative_names = ["*.example.com"]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# 2. Create the DNS validation records
resource "aws_route53_record" "validation" {
  for_each = {
    for dvo in aws_acm_certificate.cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.primary.zone_id
}

# 3. Wait for validation to complete
resource "aws_acm_certificate_validation" "cert" {
  certificate_arn         = aws_acm_certificate.cert.arn
  validation_record_fqdns = [for record in aws_route53_record.validation : record.fqdn]
}

# 4. Use the validated certificate (example: ALB listener)
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  
  # Depend on the validation resource, not the certificate request
  certificate_arn   = aws_acm_certificate_validation.cert.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}
```

**Strengths:**

- **Multi-Cloud Flexibility:** The explicit pattern allows swapping `aws_route53_record` for `cloudflare_record` or any other DNS provider. This enables cross-cloud DNS validation scenarios.
- **Transparency:** Every resource is visible and customizable. You can inspect the validation records, adjust TTLs, or add custom logic.
- **Composability:** The modular design fits Terraform's philosophy of resource composition.

**Limitations:**

- **Verbose Boilerplate:** The three-resource pattern is verbose, especially when managing multiple certificates.
- **Steeper Learning Curve:** Understanding why the validation resource exists and why dependencies must be wired correctly requires deeper knowledge of the ACM lifecycle.
- **Footguns:** It's easy to create the wrong dependency (e.g., depending on the certificate request instead of the validation resource), which can cause race conditions or deployment failures.

### Comparison Summary

| Aspect | CloudFormation/CDK (Implicit) | Terraform/Pulumi (Explicit) |
|--------|------------------------------|----------------------------|
| **Resource Count** | 1 logical resource | 3 explicit resources |
| **DNS Provider Support** | Route 53 only (implicit automation) | Any DNS provider (explicit composition) |
| **Abstraction Level** | High (validation is hidden) | Low (every step is visible) |
| **Best For** | AWS-only infrastructure, simple use cases | Multi-cloud, hybrid DNS, advanced customization |
| **Boilerplate** | Minimal | Moderate to high |
| **Dependency Management** | Automatic (implicit) | Manual (explicit `depends_on` wiring) |

Both approaches are valid and production-ready. The choice depends on your infrastructure architecture and DNS management strategy.

## Production Patterns and Critical Constraints

Deploying ACM certificates in production requires understanding several non-obvious patterns and pitfalls that can cause outages or operational friction if overlooked.

### The CloudFront us-east-1 Imperative

This is the most frequently encountered constraint and the source of countless production issues:

**The Rule:** Amazon CloudFront is a global service. To use an ACM certificate with a CloudFront distribution, the certificate **must** be provisioned in the **us-east-1 (N. Virginia)** region. This is non-negotiable and is a consequence of how CloudFront's global infrastructure is architected.

**The Contradiction:** All other ACM-integrated services (Application Load Balancers, Network Load Balancers, API Gateway) are **regional**. An ALB in `eu-west-1` requires a certificate in `eu-west-1`. An API Gateway in `ap-southeast-1` requires a certificate in `ap-southeast-1`.

**The Operational Reality:** If you have a multi-region application architecture that uses both regional load balancers and a global CloudFront distribution, you will need to provision **multiple identical certificates**:

- One certificate in each regional endpoint's region (for the regional ALB/NLB)
- One certificate in `us-east-1` (for CloudFront)

This doubles the management surface and requires careful IaC design to prevent confusion.

**Implementation Note:** In Terraform/Pulumi, this requires using provider aliases to manage resources in multiple regions simultaneously. In CloudFormation/CDK, this requires either cross-region stack dependencies or deploying separate stacks per region.

### Domain Strategy: Apex + Wildcard is King

ACM allows both wildcard certificates (`*.example.com`) and multi-domain certificates (Subject Alternative Names). Understanding their interaction is critical for minimizing certificate count while maximizing coverage.

**Pattern 1: Wildcard Only (`*.example.com`)**

- Covers: `api.example.com`, `www.example.com`, `app.example.com`, etc.
- **Does NOT cover:** `example.com` (the apex/naked domain)

**Pattern 2: Apex Only (`example.com`)**

- Covers: `example.com`
- **Does NOT cover:** Any subdomain

**Pattern 3: Apex + Wildcard (Recommended)**

This is the most common and practical production pattern:

- `domain_name`: `example.com`
- `subject_alternative_names`: `["*.example.com"]`

This single certificate covers both the apex domain and all first-level subdomains, which satisfies 90% of production use cases.

**When to Use Multi-Domain SANs:** If you need to cover multiple unrelated domains or second-level subdomains (e.g., `*.api.example.com`), add them as explicit SANs. However, this increases the validation complexity, as ACM generates a separate CNAME record for each unique domain/subdomain pattern.

### Certificate Granularity: Isolation vs. Convenience

**Monolithic Approach:** Use a single `*.example.com` wildcard certificate for all services (dev, staging, prod, api, www, etc.).

- **Pros:** Simple to manage; only one certificate to monitor.
- **Cons:** Large blast radius. If the certificate must be revoked or rotated, all services are impacted. ACM certificates cannot be shared across AWS accounts or regions, so this approach limits architectural flexibility.

**Per-Service Approach:** Create distinct certificates for distinct applications or environments (e.g., `api.example.com`, `app.production.example.com`, `staging.example.com`).

- **Pros:** Follows the principle of least privilege. Certificate lifecycle is tied to the application's IaC stack. Reduces operational risk and blast radius.
- **Cons:** More certificates to track (though ACM's auto-renewal mitigates this burden).

**Best Practice:** Use **account-level separation** for environments. Keep production certificates in a separate AWS account from development/testing. Within an account, favor **per-application certificates** to reduce coupling and blast radius.

### The Deletion Lifecycle Trap

This is a critical operational hazard that can cause IaC destroy operations to fail catastrophically:

**The Problem:** ACM's API prevents deletion of a certificate that is currently associated with any AWS resource. The API returns a `ResourceInUseException` error.

**The IaC Deadlock:** When you run `terraform destroy` or delete a CloudFormation stack:

1. The IaC tool plans to delete the load balancer listener first, then the certificate.
2. The tool sends the API call to delete the listener. This is often asynchronous—the API accepts the request, but the deletion takes time.
3. The tool immediately sends the API call to delete the certificate.
4. The ACM API rejects the deletion with `ResourceInUseException` because the listener, while "deleting," has not yet fully disassociated from the certificate.
5. The destroy operation fails, leaving the infrastructure in a partially destroyed state.

**The "Shadow Resource" Nightmare:** This problem is exponentially worse when the associated resource is invisible:

- **API Gateway:** When you create a custom domain in API Gateway, AWS **implicitly provisions** an Application Load Balancer in an AWS-managed account and associates it with your certificate. You cannot see or manage this "shadow" ALB.
- **Amazon Cognito:** When you add a custom domain to a User Pool, AWS **implicitly provisions** a CloudFront distribution and associates it with your certificate.

When you delete the API Gateway custom domain or Cognito User Pool, the "shadow" resource is deleted, but the association with the certificate may linger. This can permanently block certificate deletion and often requires filing an AWS support ticket to manually break the association.

**Mitigation Strategies:**

1. **Resource Protection:** Use `lifecycle { prevent_destroy = true }` (Terraform), `DeletionPolicy: Retain` (CloudFormation), or `protect: true` (Pulumi) on all production certificate resources. This prevents accidental deletion.

2. **Explicit Dependencies:** In Terraform, use explicit `depends_on` to ensure certificates are deleted last, after all dependent resources.

3. **Manual Disassociation:** Before destroying a stack with ACM certificates, manually check for associations using the AWS Console or `aws acm describe-certificate --certificate-arn <arn>` and disassociate them first.

4. **Retain Policy by Default:** Design IaC modules to retain certificates by default, requiring explicit opt-in to allow deletion.

### Monitoring and Renewal

While ACM's auto-renewal for DNS-validated certificates is highly reliable, it's not infallible:

**Monitoring Best Practices:**

1. **EventBridge Alerts:** Configure AWS EventBridge to alert on "ACM Certificate Approaching Expiration" and "ACM Certificate Renewal Action Required" events. This is **critical** for imported certificates and email-validated certificates (if you must use them), as these require manual renewal.

2. **CNAME Persistence Checks:** Implement automated checks (e.g., via AWS Config rules or custom Lambda functions) to verify that the `_acm-validation` CNAME records remain in DNS. Accidental deletion of these records breaks renewal.

3. **Certificate Inventory:** Maintain a centralized inventory of all ACM certificates, their expiration dates, associated resources, and validation methods. This can be automated using AWS Config aggregators or third-party tools.

## Planton's Choice: Simplicity Through Intelligent Orchestration

The `AwsCertManagerCert` resource in Planton is designed to deliver the **best user experience** by learning from both the AWS-native and multi-cloud IaC approaches. It aims for CloudFormation/CDK-level simplicity while maintaining the flexibility and transparency that advanced users expect.

### Design Philosophy

1. **DNS Validation as the Default (and Strong Recommendation):** The API defaults to DNS validation and strongly discourages email validation through documentation. Email validation is supported only because the AWS API allows it, but users are warned that it's a production anti-pattern.

2. **Route 53 Automation:** When a `route53_hosted_zone_id` is provided, the resource **automatically orchestrates** the creation of the CNAME validation records, emulating the CloudFormation `DomainValidationOptions` behavior. Users don't need to manage the validation records as separate infrastructure.

3. **Explicit Region Control:** The API includes a `region` field (inherited from the resource's environment/credential configuration) to handle the critical CloudFront `us-east-1` use case. This is not an optional enhancement—it's a mandatory feature for real-world production deployments.

4. **Minimal API Surface:** Following the 80/20 principle, the API exposes only the fields that 80% of users need 80% of the time. Advanced fields (like key algorithm, private CA ARN, or certificate transparency settings) are either omitted or nested in an optional `advanced_options` block to keep the primary API clean and approachable.

### Core API Fields

The `AwsCertManagerCertSpec` protobuf message defines:

1. **`primary_domain_name`** (string, required): The main domain for the certificate. Can be an apex domain (`example.com`) or a wildcard (`*.example.com`).

2. **`alternate_domain_names`** (repeated string, optional): Subject Alternative Names (SANs). This allows a single certificate to cover multiple domains or subdomain patterns. The most common pattern is `primary_domain_name: "example.com"` with `alternate_domain_names: ["*.example.com"]`.

3. **`route53_hosted_zone_id`** (StringValueOrRef, required): The Route 53 Hosted Zone ID where the DNS validation CNAME records will be created automatically. This is the key to automation. If the domain is not in Route 53, users must manage validation records externally (and the resource documents this as an advanced, unsupported use case).

4. **`validation_method`** (string, optional, default: "DNS"): Allows selection between DNS and EMAIL validation. Defaults to DNS and documentation strongly recommends never changing this.

### What's Handled Automatically

When you define an `AwsCertManagerCert` resource with a `route53_hosted_zone_id`, the underlying Pulumi or Terraform module:

1. Requests the certificate via the ACM API with DNS validation
2. Extracts the `domain_validation_options` from the ACM response
3. Iterates over each domain and creates the required CNAME records in the specified Route 53 hosted zone
4. Waits for ACM to validate the certificate and report status as "ISSUED"
5. Exports the certificate ARN and primary domain name as stack outputs

This orchestration mirrors the Terraform "three-resource pattern" under the hood but presents it as a single, logical resource to the user.

### Stack Outputs

The `AwsCertManagerCertStackOutputs` message provides:

1. **`cert_arn`**: The ARN of the issued certificate. This is the value you'll reference when attaching the certificate to a load balancer, CloudFront distribution, or API Gateway custom domain.

2. **`certificate_domain_name`**: The primary domain name for which the certificate was issued. This is useful for documentation, auditing, and debugging.

### Example: Typical Production Configuration

```yaml
apiVersion: aws.planton.ai/v1
kind: AwsCertManagerCert
metadata:
  name: example-com-cert
spec:
  primary_domain_name: example.com
  alternate_domain_names:
    - "*.example.com"
  route53_hosted_zone_id: Z2KZ5YTUFZNC7H  # Resolved from Route53 resource
  validation_method: DNS  # Explicit, though this is the default
```

This single YAML definition results in a fully validated, production-ready certificate covering `example.com` and all first-level subdomains, with zero manual DNS intervention required.

## Conclusion: Automation Requires the Right Foundation

The history of ACM deployment methods reveals a fundamental truth: **automation is only as reliable as the validation method you choose**. Email validation is a trap disguised as simplicity—a manual process that will inevitably cause a production outage. DNS validation, when implemented correctly through declarative IaC, is the only path to true "set-it-and-forget-it" certificate management.

The divide between implicit (AWS-native) and explicit (multi-cloud) IaC approaches is not about which is "better" in an absolute sense. CloudFormation and CDK offer the simplest experience for teams fully committed to AWS. Terraform and Pulumi offer the transparency and flexibility needed for multi-cloud architectures or hybrid DNS scenarios. Both are production-ready when used with the right patterns.

Planton's `AwsCertManagerCert` resource synthesizes the best of both worlds: it delivers CloudFormation-level simplicity for the common case (AWS domain in Route 53) while maintaining the transparency and control that advanced users expect. It defaults to the secure path, strongly recommends against anti-patterns, and handles the complex orchestration—CNAME record creation, validation waiting, dependency management—so you don't have to.

In a world where SSL/TLS certificates are no longer an operational burden but a transparent infrastructure capability, ACM stands out as one of AWS's most successful managed services. When combined with intelligent IaC abstractions, it transforms certificate management from a recurring task into a solved problem.


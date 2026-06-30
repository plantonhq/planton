# Overview

The **GcpCertManagerCert** API resource provides a standardized and straightforward way to provision SSL/TLS certificates on Google Cloud Platform (GCP) using either Google Certificate Manager or Google-managed SSL certificates for load balancers. By focusing on essential configurations like domain names, DNS validation, and certificate type, it makes managing SSL/TLS certificates on GCP far more accessible within the Planton multi-cloud deployment framework.

## Purpose

Deploying SSL/TLS certificates typically involves handling multiple moving parts—DNS validation records, certificate provisioning, domain verification, and more. The **GcpCertManagerCert** resource aims to streamline that process by:

- **Simplifying Certificate Deployments**: Offer an easy-to-use interface for provisioning SSL/TLS certificates on GCP.
- **Aligning with Best Practices**: Provide recommended defaults (e.g., DNS validation, Certificate Manager) to ensure users have a production-ready baseline without repetitive configuration.
- **Promoting Consistency**: Enforce standardized naming and validations, reducing misconfigurations across multiple certificates and environments.
- **Supporting Multiple Certificate Types**: Choose between Certificate Manager (modern, feature-rich) or Load Balancer certificates (classic, load balancer-specific) based on your needs.

## Key Features

### Dual Certificate Type Support

- **Certificate Manager (MANAGED)**: Uses Google Certificate Manager for modern, flexible certificate management with advanced features.
- **Load Balancer Certificates**: Uses Google-managed SSL certificates optimized for load balancers, providing a classic approach.

### Automatic DNS Validation

- **Cloud DNS Integration**: Automatically creates DNS validation records in Google Cloud DNS for domain verification.
- **Seamless Validation**: Handles the entire validation process without manual intervention.

### Multi-Domain Support

- **Primary Domain**: Define your main domain (apex or wildcard).
- **Subject Alternative Names (SANs)**: Add multiple alternate domains to a single certificate for comprehensive coverage.
- **Wildcard Support**: Protect entire subdomains with wildcard certificates (e.g., `*.example.com`).

### Environment Management

- **GCP Project Isolation**: Deploy certificates in specific GCP projects for proper environment separation.
- **Cloud DNS Zones**: Integrate with existing Cloud DNS managed zones for validation.

### Seamless Integration

- **Planton CLI**: Deploy the same resource across multiple stacks using either Pulumi or Terraform under the hood.
- **Multi-Cloud Ready**: Combine GcpCertManagerCert on GCP with other providers in the same manifest, adopting Planton's uniform resource model.

## Benefits

- **Reduced Complexity**: A single definition for your certificate—domain names, validation settings, and type—means fewer files and less overhead.
- **Automatic Validation**: DNS validation records are created automatically, eliminating manual DNS configuration steps.
- **Infrastructure Consistency**: Enforce naming conventions, validations, and recommended defaults so your deployments remain predictable and repeatable.
- **Enhanced Security**: Automated certificate provisioning and renewal reduce the risk of expired certificates causing outages.
- **Flexibility**: Choose the certificate type that best fits your architecture—Certificate Manager for general use or Load Balancer certificates for LB-specific scenarios.

## Example Usage

Below is a minimal YAML snippet demonstrating how to configure and deploy a GCP certificate using Planton:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: my-gcp-cert
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: example.com
  alternateDomainNames:
    - www.example.com
    - api.example.com
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: MANAGED
  validationMethod: DNS
```

### Deploying with Planton

Once your YAML manifest is ready, you can deploy using Planton's CLI. Planton will validate the manifest against the Protobuf schema and orchestrate everything in Pulumi or Terraform.

- **Using Pulumi**:
  ```bash
  planton pulumi up --manifest gcpcertmanagercert.yaml --stack org/project/my-stack
  ```
- **Using Terraform**:
  ```bash
  planton terraform apply --manifest gcpcertmanagercert.yaml --stack org/project/my-stack
  ```

Planton will provision the certificate, create DNS validation records in Cloud DNS, and ensure the certificate is validated and ready for use.

## Certificate Types

### Certificate Manager (MANAGED)

Google Certificate Manager is the modern approach to SSL/TLS certificate management on GCP. It provides:

- **Advanced Features**: Certificate maps, regional certificates, and more.
- **Flexible Integration**: Works with various GCP services beyond just load balancers.
- **Automatic Renewal**: Handles certificate renewal automatically.
- **DNS Authorization**: Explicit DNS authorization records for domain validation.

**Use when**: You need a modern, flexible certificate management solution that works across multiple GCP services.

### Load Balancer Certificates (LOAD_BALANCER)

Google-managed SSL certificates are optimized for load balancers. They provide:

- **Classic Approach**: Traditional SSL certificate management for load balancers.
- **Automatic Provisioning**: Certificates are provisioned when attached to a load balancer.
- **Simplified Management**: Less configuration required for basic load balancer scenarios.

**Use when**: You're specifically using GCP load balancers and want a straightforward certificate solution.

## DNS Validation

The resource uses DNS validation through Google Cloud DNS:

1. **DNS Authorization**: For Certificate Manager certificates, DNS authorization records are created.
2. **Validation Records**: DNS TXT or CNAME records are automatically added to your Cloud DNS zone.
3. **Automatic Verification**: GCP verifies domain ownership through these DNS records.
4. **Certificate Issuance**: Once validated, the certificate is issued and ready for use.

## Best Practices

- **Use Certificate Manager**: For new deployments, prefer Certificate Manager (MANAGED) for its advanced features.
- **Wildcard Certificates**: Use wildcards (`*.example.com`) to cover multiple subdomains with a single certificate.
- **Separate Environments**: Use different GCP projects for development, staging, and production certificates.
- **Monitor Expiration**: While certificates auto-renew, monitor their status for any issues.

---

Happy deploying! If you have questions or run into issues, feel free to open an issue on our GitHub repository or reach out through our community channels for support.


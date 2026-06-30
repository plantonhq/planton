# AWS Certificate

Deploys a public SSL/TLS certificate through AWS Certificate Manager (ACM) with automatic DNS validation via Route53. Planton creates the certificate, provisions the required CNAME validation records in the specified hosted zone, and waits for ACM to confirm domain ownership before marking the deployment complete.

## What Gets Created

When you deploy an AwsCertManagerCert resource, Planton provisions:

- **ACM Certificate** — an `acm.Certificate` resource requesting a public certificate for the primary domain and any alternate domain names, validated via DNS
- **Route53 CNAME Records** — one `route53.Record` per unique domain validation option, created in the specified hosted zone with a TTL of 300 seconds, used by ACM to verify domain ownership
- **Certificate Validation** — an `acm.CertificateValidation` resource that blocks until ACM confirms all DNS validation records have been verified and the certificate is issued

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **A Route53 public hosted zone** that is authoritative for the domain names on the certificate
- **Domain ownership** — the hosted zone must be able to serve the CNAME records that ACM requires for validation

## Quick Start

Create a file `cert.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCertManagerCert
metadata:
  name: my-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCertManagerCert.my-cert
spec:
  region: us-east-1
  primaryDomainName: example.com
  route53HostedZoneId: Z0123456789ABCDEFGHIJ
```

Deploy:

```shell
planton apply -f cert.yaml
```

This creates an ACM certificate for `example.com`, adds the DNS validation CNAME record to the specified Route53 zone, and waits for validation to complete.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the certificate will be created (e.g., `us-east-1`, `eu-west-1`). | Required; non-empty |
| `primaryDomainName` | `string` | Main domain name for the certificate. Supports wildcard prefixes (e.g., `*.example.com`). | Must match pattern `^(?:\*\.[A-Za-z0-9\-\.]+\|[A-Za-z0-9\-\.]+\.[A-Za-z]{2,})$` |
| `route53HostedZoneId` | `StringValueOrRef` | ID of the Route53 public hosted zone where DNS validation records are created. Can reference an AwsRoute53Zone resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `alternateDomainNames` | `string[]` | `[]` | Subject Alternative Names (SANs) for the certificate. Each entry follows the same pattern as `primaryDomainName`. Must not contain duplicates. Do not repeat the primary domain here. |
| `validationMethod` | `string` | `"DNS"` | How ACM verifies domain ownership. Valid values: `DNS`, `EMAIL`. |

## Examples

### Single Domain Certificate

A certificate for a single apex domain:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCertManagerCert
metadata:
  name: apex-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCertManagerCert.apex-cert
spec:
  region: us-east-1
  primaryDomainName: example.com
  route53HostedZoneId: Z0123456789ABCDEFGHIJ
```

### Wildcard Certificate

A wildcard certificate covering all subdomains of a domain:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCertManagerCert
metadata:
  name: wildcard-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCertManagerCert.wildcard-cert
spec:
  region: us-east-1
  primaryDomainName: "*.example.com"
  route53HostedZoneId: Z0123456789ABCDEFGHIJ
```

### Certificate with Subject Alternative Names

A certificate covering the apex domain and multiple specific subdomains:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCertManagerCert
metadata:
  name: multi-domain-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCertManagerCert.multi-domain-cert
spec:
  region: us-east-1
  primaryDomainName: example.com
  alternateDomainNames:
    - www.example.com
    - api.example.com
    - admin.example.com
  route53HostedZoneId: Z0123456789ABCDEFGHIJ
```

### Using Foreign Key References

Reference an Planton-managed Route53 zone instead of hardcoding the zone ID:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCertManagerCert
metadata:
  name: ref-cert
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCertManagerCert.ref-cert
spec:
  region: us-east-1
  primaryDomainName: "*.example.com"
  alternateDomainNames:
    - example.com
  route53HostedZoneId:
    valueFrom:
      kind: AwsRoute53Zone
      name: my-zone
      field: status.outputs.zone_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cert_arn` | `string` | ARN of the issued ACM certificate, used to attach the certificate to ALBs, CloudFront distributions, or API Gateways |
| `certificate_domain_name` | `string` | The primary domain name for which the certificate was issued |

## Related Components

- [AwsRoute53Zone](/docs/catalog/aws/awsroute53zone) — provides the hosted zone where DNS validation records are created
- [AwsAlb](/docs/catalog/aws/awsalb) — uses the certificate ARN for SSL termination on HTTPS listeners
- [AwsCloudfront](/docs/catalog/aws/awscloudfront) — uses the certificate ARN for HTTPS on CloudFront distributions

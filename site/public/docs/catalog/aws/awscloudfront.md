---
title: "CloudFront"
description: "CloudFront deployment documentation"
icon: "package"
order: 100
componentName: "awscloudfront"
---

# AWS CloudFront

Deploys an AWS CloudFront distribution with one or more origins, a default cache behavior that redirects HTTP to HTTPS, and optional custom-domain SSL via ACM. The component uses a custom origin configuration with TLSv1.2 and applies no geo-restrictions by default.

## What Gets Created

When you deploy an AwsCloudFront resource, OpenMCF provisions:

- **CloudFront Distribution** — a `cloudfront.Distribution` with the specified origins, a default cache behavior (GET/HEAD, redirect-to-https, query-string forwarding disabled, cookie forwarding disabled, default TTL 3600s, max TTL 86400s), and geo-restrictions set to `none`
- **Custom Origin Configuration** — each origin is configured with `https-only` protocol policy, ports 80/443, and `TLSv1.2` SSL protocol
- **Viewer Certificate** — if `certificateArn` is provided, uses SNI-only with minimum protocol `TLSv1.2_2021`; otherwise uses the CloudFront default certificate (`*.cloudfront.net`)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An origin endpoint** (e.g., an S3 bucket website endpoint, an ALB DNS name, or any HTTPS-capable domain)
- **An ACM certificate ARN in us-east-1** if using custom domain aliases (CloudFront requires certificates in us-east-1 regardless of origin region)

## Quick Start

Create a file `cloudfront.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: my-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsCloudFront.my-cdn
spec:
  enabled: true
  origins:
    - domainName: my-bucket.s3.us-east-1.amazonaws.com
      isDefault: true
```

Deploy:

```shell
openmcf apply -f cloudfront.yaml
```

This creates a CloudFront distribution with a single origin, using the default CloudFront certificate and `PriceClass_All` edge locations.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `origins` | `Origin[]` | List of origins available to the distribution. Must contain at least one origin with exactly one marked as default. | Minimum 1 item; exactly one must have `isDefault: true` |
| `origins[].domainName` | `string` | DNS name of the origin, e.g., `my-bucket.s3.amazonaws.com`. | Minimum 1 character; must be a valid domain name |
| `origins[].isDefault` | `bool` | Whether this origin is the default for the distribution. Exactly one origin must be marked as default. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `bool` | `false` | Whether the distribution is enabled and serving traffic. Set to `true` for the distribution to accept requests. |
| `aliases` | `string[]` | `[]` | Alternate domain names (CNAMEs) for the distribution, e.g., `cdn.example.com`. Must be unique. Requires `certificateArn` to be set. |
| `certificateArn` | `string` | — | ACM certificate ARN for custom domains. Must be in the `us-east-1` region. Required when `aliases` are provided. When omitted, the distribution uses the default `*.cloudfront.net` certificate. |
| `priceClass` | `enum` | `PRICE_CLASS_ALL` | Controls which CloudFront edge locations serve content. Valid values: `PRICE_CLASS_100` (US, Canada, Europe), `PRICE_CLASS_200` (adds Asia, Middle East, Africa), `PRICE_CLASS_ALL` (all edge locations). |
| `origins[].originPath` | `string` | `""` | Path that CloudFront appends to origin requests, e.g., `/production`. Must start with `/` if set. |
| `defaultRootObject` | `string` | `""` | Object returned when a viewer requests the root URL, e.g., `index.html`. |

## Examples

### Static Website with S3 Origin

A distribution serving a static site from an S3 bucket with `index.html` as the root object:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: static-site-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsCloudFront.static-site-cdn
spec:
  enabled: true
  defaultRootObject: index.html
  origins:
    - domainName: my-website-bucket.s3-website-us-east-1.amazonaws.com
      isDefault: true
```

### Custom Domain with SSL

A distribution with a custom domain name and ACM certificate for HTTPS:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: branded-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsCloudFront.branded-cdn
spec:
  enabled: true
  aliases:
    - cdn.example.com
    - assets.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abc-12345
  defaultRootObject: index.html
  origins:
    - domainName: my-bucket.s3.us-east-1.amazonaws.com
      isDefault: true
```

### Cost-Optimized Distribution

A distribution restricted to US, Canada, and Europe edge locations to reduce costs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: regional-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsCloudFront.regional-cdn
spec:
  enabled: true
  priceClass: PRICE_CLASS_100
  origins:
    - domainName: api.internal.example.com
      isDefault: true
```

### Multi-Origin Distribution

A distribution with multiple origins, routing to different backends. The default origin serves the main site, while a second origin serves content from a subdirectory:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudFront
metadata:
  name: multi-origin-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsCloudFront.multi-origin-cdn
spec:
  enabled: true
  aliases:
    - www.example.com
  certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/prod-cert
  priceClass: PRICE_CLASS_200
  defaultRootObject: index.html
  origins:
    - domainName: website-bucket.s3.us-east-1.amazonaws.com
      isDefault: true
    - domainName: api-alb.us-east-1.elb.amazonaws.com
      originPath: /v1
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `distribution_id` | `string` | CloudFront distribution ID (e.g., `E123ABCXYZ`) |
| `domain_name` | `string` | CloudFront-assigned domain name (e.g., `d123abc.cloudfront.net`) |
| `hosted_zone_id` | `string` | Route53 hosted zone ID for aliasing to CloudFront (always `Z2FDTNDATAQYW2`) |

## Related Components

- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — common origin for static website hosting
- [AwsCertManagerCert](/docs/catalog/aws/awscertmanagercert) — provides ACM certificates for custom domain SSL
- [AwsRoute53Zone](/docs/catalog/aws/awsroute53zone) — hosts DNS zones for creating alias records pointing to the distribution
- [AwsRoute53DnsRecord](/docs/catalog/aws/awsroute53dnsrecord) — creates alias records pointing custom domains to the CloudFront distribution

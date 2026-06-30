# AWS Route53 DNS Record

## Overview

The `AwsRoute53DnsRecord` component enables declarative management of individual DNS records in AWS Route53 hosted zones. This component is designed for users who need fine-grained control over DNS records, including support for advanced Route53 features like alias records and routing policies.

Route53's alias records are a powerful AWS-specific feature that allows pointing zone apex domains (like `example.com`) directly to AWS resources without the restrictions of CNAME records, and without incurring Route53 query charges for alias queries to AWS resources.

## Purpose

This component simplifies DNS record management by:

- **Declarative Record Management**: Define DNS records as code with full validation
- **Resource References**: Wire records to zones and alias targets using `value_from` references
- **Alias Record Support**: Leverage Route53's killer feature for pointing domains to AWS resources (CloudFront, ALB/NLB, S3, API Gateway)
- **Advanced Routing Policies**: Configure weighted, latency-based, failover, and geolocation routing
- **Validation**: Built-in schema validation ensures correct configuration before deployment

## Key Features

- **Standard DNS Records**: A, AAAA, CNAME, MX, TXT, NS, SRV, CAA
- **Alias Records**: Point zone apex to AWS resources with zero query charges
- **Value References**: Use `value_from` to wire records to other resources (zones, ALBs)
- **Weighted Routing**: Split traffic for blue/green or canary deployments
- **Latency-Based Routing**: Route users to lowest-latency endpoint
- **Failover Routing**: Automatic failover with health check integration
- **Geolocation Routing**: Route based on user location (GDPR, localization)
- **Wildcard Support**: Create `*.example.com` catch-all records

## Benefits

- **No CNAME at Apex Limitation**: Use alias records for zone apex domains
- **Cost Efficiency**: Alias queries to AWS resources are free
- **High Availability**: Built-in support for failover and multi-region routing
- **Infrastructure as Code**: Full GitOps workflow with validation
- **Seamless Wiring**: Reference other resources directly with `value_from`

## Example Usage

### Basic A Record

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-example
spec:
  zone_id:
    value: Z1234567890ABC
  name: www.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.1
    - 192.0.2.2
```

### A Record with Zone Reference

Reference an existing `AwsRoute53Zone` resource:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-example
spec:
  zone_id:
    value_from:
      name: my-zone  # References AwsRoute53Zone named "my-zone"
  name: www.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.1
```

### Alias Record to ALB

Wire directly to an `AwsAlb` resource:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-alb-alias
spec:
  zone_id:
    value_from:
      name: my-zone
  name: api.example.com
  type: A
  alias_target:
    dns_name:
      value_from:
        name: my-alb  # References AwsAlb named "my-alb"
    zone_id:
      value_from:
        name: my-alb
    evaluate_target_health: true
```

### Alias Record to CloudFront (Literal Values)

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsRoute53DnsRecord
metadata:
  name: apex-cloudfront
spec:
  zone_id:
    value: Z1234567890ABC
  name: example.com
  type: A
  alias_target:
    dns_name:
      value: d1234abcd.cloudfront.net
    zone_id:
      value: Z2FDTNDATAQYW2
    evaluate_target_health: false
```

Deploy using the Planton CLI:

```bash
planton pulumi up --manifest dns-record.yaml
```

## Best Practices

1. **Use Alias for AWS Resources**: Alias records are free and automatically update when IPs change
2. **Use value_from for Resource Wiring**: Reference zones and targets by name for maintainability
3. **Set Appropriate TTLs**: Lower TTL (60s) for records you might change during incidents
4. **Use Health Checks with Failover**: Combine failover routing with health checks for automatic recovery
5. **Document Records**: Use meaningful resource names that describe the record's purpose

## Related Components

- **AwsRoute53Zone**: Create and manage Route53 hosted zones
- **AwsAlb**: Application Load Balancers (common alias targets with value_from support)
- **AwsCertManagerCert**: SSL/TLS certificates for your domains
- **AwsCloudFront**: CloudFront distributions (common alias targets)

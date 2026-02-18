# AWS Route53 DNS Record Examples

This document provides working, copy-paste ready examples for common Route53 DNS record configurations.

## Basic A Record (Literal Zone ID)

Simple A record pointing a subdomain to an IP address.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-a-record
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: www.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.1
```

## A Record with Zone Reference

Reference an existing `AwsRoute53Zone` resource using `value_from`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-a-record
spec:
  region: us-east-1
  zone_id:
    value_from:
      name: my-zone  # References AwsRoute53Zone named "my-zone"
  name: www.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.1
```

## A Record with Multiple IPs (Round Robin)

DNS-based load balancing across multiple servers.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-round-robin
spec:
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.1
    - 192.0.2.2
    - 192.0.2.3
```

## CNAME Record

Alias a subdomain to another domain name.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: blog-cname
spec:
  zone_id:
    value: Z1234567890ABC
  name: blog.example.com
  type: CNAME
  ttl: 300
  values:
    - example.ghost.io
```

## MX Records for Email

Configure mail exchange servers with priority.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: mail-mx-records
spec:
  zone_id:
    value: Z1234567890ABC
  name: example.com
  type: MX
  ttl: 3600
  values:
    - "10 mail1.example.com"
    - "20 mail2.example.com"
    - "30 mail3.example.com"
```

## TXT Record for SPF

Email authentication SPF record.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: spf-record
spec:
  zone_id:
    value: Z1234567890ABC
  name: example.com
  type: TXT
  ttl: 300
  values:
    - "v=spf1 include:_spf.google.com include:servers.mcsv.net ~all"
```

## Wildcard A Record

Catch-all record for any subdomain.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: wildcard-record
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: "*.example.com"
  type: A
  ttl: 300
  values:
    - 192.0.2.1
```

## Alias Record to ALB (with value_from)

Wire directly to an `AwsAlb` resource - the most common pattern for web applications.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-alb-alias
spec:
  zone_id:
    value_from:
      name: my-zone  # References AwsRoute53Zone
  name: api.example.com
  type: A
  alias_target:
    dns_name:
      value_from:
        name: my-alb  # References AwsAlb - gets load_balancer_dns_name
    zone_id:
      value_from:
        name: my-alb  # References AwsAlb - gets load_balancer_hosted_zone_id
    evaluate_target_health: true
```

## Alias Record to CloudFront Distribution (Literal)

Point zone apex to CloudFront (free queries, no CNAME restriction).

```yaml
apiVersion: aws.openmcf.org/v1
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
      value: Z2FDTNDATAQYW2  # CloudFront's global hosted zone ID
    evaluate_target_health: false
```

## Alias Record to S3 Website

Point domain to S3 static website hosting.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: static-s3-alias
spec:
  zone_id:
    value: Z1234567890ABC
  name: static.example.com
  type: A
  alias_target:
    dns_name:
      value: my-bucket.s3-website-us-east-1.amazonaws.com
    zone_id:
      value: Z3AQBSTGFYJSTF  # S3 us-east-1 hosted zone ID
    evaluate_target_health: false
```

## Weighted Routing - Blue/Green Deployment

Split traffic between two versions (70% blue, 30% green).

**Blue environment (70% traffic):**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-weighted-blue
spec:
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.1
  routing_policy:
    weighted:
      weight: 70
  set_identifier: blue
```

**Green environment (30% traffic):**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-weighted-green
spec:
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.2
  routing_policy:
    weighted:
      weight: 30
  set_identifier: green
```

## Latency-Based Routing - Multi-Region

Route users to the lowest-latency endpoint.

**US East endpoint:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-latency-us-east
spec:
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.1
  routing_policy:
    latency:
      region: us-east-1
  set_identifier: us-east-1
```

**EU West endpoint:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-latency-eu-west
spec:
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.2
  routing_policy:
    latency:
      region: eu-west-1
  set_identifier: eu-west-1
```

## Failover Routing - Disaster Recovery

Automatic failover when primary fails health check.

**Primary record:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-failover-primary
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: www.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.1
  routing_policy:
    failover:
      failover_type: primary
  set_identifier: primary
  health_check_id: abcd1234-5678-90ab-cdef-example
```

**Secondary record:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: www-failover-secondary
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: www.example.com
  type: A
  ttl: 60
  values:
    - 192.0.2.2
  routing_policy:
    failover:
      failover_type: secondary
  set_identifier: secondary
```

## Geolocation Routing - GDPR Compliance

Route EU users to EU servers for data residency.

**EU users:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-geo-eu
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.1
  routing_policy:
    geolocation:
      continent: EU
  set_identifier: europe
```

**US users:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: api-geo-us
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: api.example.com
  type: A
  ttl: 300
  values:
    - 192.0.2.2
  routing_policy:
    geolocation:
      country: US
  set_identifier: us
```

## CAA Record - Certificate Authority Authorization

Restrict which CAs can issue certificates for your domain.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53DnsRecord
metadata:
  name: caa-record
spec:
  region: us-east-1
  zone_id:
    value: Z1234567890ABC
  name: example.com
  type: CAA
  ttl: 3600
  values:
    - '0 issue "letsencrypt.org"'
    - '0 issue "amazon.com"'
    - '0 issuewild ";"'
```

## Deployment

Deploy any of these examples using the OpenMCF CLI:

```bash
# Deploy a single record
openmcf pulumi up --manifest dns-record.yaml

# Deploy with OpenTofu/Terraform
openmcf tofu apply --manifest dns-record.yaml
```

## value_from Defaults

When using `value_from`, the following defaults are applied:

| Field | Default Kind | Default Field Path |
|-------|--------------|-------------------|
| `spec.zone_id` | `AwsRoute53Zone` | `status.outputs.zone_id` |
| `alias_target.dns_name` | `AwsAlb` | `status.outputs.load_balancer_dns_name` |
| `alias_target.zone_id` | `AwsAlb` | `status.outputs.load_balancer_hosted_zone_id` |

You can override these by specifying `kind` and `field_path` in the `value_from` block.

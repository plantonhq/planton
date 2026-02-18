# AWS Route53 Zone Examples

Below are several examples demonstrating how to define an AWS Route53 Zone component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Route53 Zone

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "example.com."
      values:
        - "192.0.2.1"
      ttlSeconds: 300
```

This example creates a basic Route53 hosted zone:
• Domain name: example.com
• A record pointing to IP address 192.0.2.1
• 300-second TTL for caching
• Suitable for simple website hosting.

---

## Route53 Zone with Multiple Records

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example-multi.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "www.example-multi.com."
      values:
        - "192.0.2.10"
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "api.example-multi.com."
      values:
        - "www.example-multi.com."
      ttlSeconds: 300
    - recordType: "AAAA"
      name: "www.example-multi.com."
      values:
        - "2001:db8::1"
      ttlSeconds: 300
```

This example includes multiple record types:
• A record for IPv4 address resolution
• CNAME record for subdomain aliasing
• AAAA record for IPv6 address resolution
• Consistent TTL settings across records.

---

## Route53 Zone with Email Configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: mail-example.com
spec:
  region: us-east-1
  records:
    - recordType: "MX"
      name: "mail-example.com."
      values:
        - "10 mailserver1.example.org."
        - "20 mailserver2.example.org."
      ttlSeconds: 3600
    - recordType: "TXT"
      name: "mail-example.com."
      values:
        - "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
    - recordType: "A"
      name: "mail.mail-example.com."
      values:
        - "192.0.2.100"
      ttlSeconds: 300
```

This example configures email services:
• MX records for mail server priorities
• TXT record for SPF email authentication
• A record for mail server IP address
• Longer TTL for email records (3600 seconds).

---

## Route53 Zone with Load Balancer Integration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: app-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "app-example.com."
      values:
        - "dualstack.alb-123456789.us-east-1.elb.amazonaws.com."
      ttlSeconds: 300
    - recordType: "A"
      name: "api.app-example.com."
      values:
        - "dualstack.alb-987654321.us-east-1.elb.amazonaws.com."
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "www.app-example.com."
      values:
        - "app-example.com."
      ttlSeconds: 300
```

This example integrates with AWS load balancers:
• A records pointing to Application Load Balancers
• CNAME for www subdomain redirection
• Dual-stack ALB endpoints for IPv4/IPv6 support
• API subdomain for separate service routing.

---

## Route53 Zone with CDN Integration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: cdn-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "cdn-example.com."
      values:
        - "d1234abcd.cloudfront.net."
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "www.cdn-example.com."
      values:
        - "cdn-example.com."
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "static.cdn-example.com."
      values:
        - "d5678efgh.cloudfront.net."
      ttlSeconds: 3600
```

This example integrates with CloudFront CDN:
• A record for main domain pointing to CloudFront
• CNAME for www subdomain redirection
• Separate CNAME for static assets CDN
• Longer TTL for static content (3600 seconds).

---

## Route53 Zone with Subdomain Management

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: subdomain-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "subdomain-example.com."
      values:
        - "192.0.2.1"
      ttlSeconds: 300
    - recordType: "A"
      name: "dev.subdomain-example.com."
      values:
        - "192.0.2.10"
      ttlSeconds: 300
    - recordType: "A"
      name: "staging.subdomain-example.com."
      values:
        - "192.0.2.20"
      ttlSeconds: 300
    - recordType: "A"
      name: "prod.subdomain-example.com."
      values:
        - "192.0.2.30"
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "*.subdomain-example.com."
      values:
        - "subdomain-example.com."
      ttlSeconds: 300
```

This example manages multiple environments:
• Main domain A record
• Environment-specific subdomains (dev, staging, prod)
• Wildcard CNAME for catch-all subdomains
• Consistent TTL settings across environments.

---

## Route53 Zone with Security Records

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: secure-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "secure-example.com."
      values:
        - "192.0.2.1"
      ttlSeconds: 300
    - recordType: "TXT"
      name: "secure-example.com."
      values:
        - "v=spf1 include:_spf.google.com ~all"
        - "google-site-verification=abc123def456"
      ttlSeconds: 3600
    - recordType: "CAA"
      name: "secure-example.com."
      values:
        - "0 issue \"letsencrypt.org\""
        - "0 issue \"amazontrust.com\""
      ttlSeconds: 3600
    - recordType: "TXT"
      name: "_dmarc.secure-example.com."
      values:
        - "v=DMARC1; p=quarantine; rua=mailto:dmarc@secure-example.com"
      ttlSeconds: 3600
```

This example includes security-focused records:
• SPF record for email authentication
• Google site verification for search console
• CAA records for certificate authority authorization
• DMARC record for email policy enforcement
• Longer TTL for security records (3600 seconds).

---

## Route53 Zone with Service Discovery

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: services-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "services-example.com."
      values:
        - "192.0.2.1"
      ttlSeconds: 300
    - recordType: "SRV"
      name: "_sip._tcp.services-example.com."
      values:
        - "0 5 5060 sipserver.services-example.com."
      ttlSeconds: 3600
    - recordType: "SRV"
      name: "_ldap._tcp.services-example.com."
      values:
        - "0 10 389 ldapserver.services-example.com."
      ttlSeconds: 3600
    - recordType: "A"
      name: "sipserver.services-example.com."
      values:
        - "192.0.2.100"
      ttlSeconds: 300
    - recordType: "A"
      name: "ldapserver.services-example.com."
      values:
        - "192.0.2.200"
      ttlSeconds: 300
```

This example includes service discovery records:
• SRV records for SIP and LDAP services
• A records for service server IPs
• Service-specific subdomains
• Longer TTL for service records (3600 seconds).

---

## Route53 Zone with Geographic Routing

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: geo-example.com
spec:
  region: us-east-1
  records:
    - recordType: "A"
      name: "geo-example.com."
      values:
        - "192.0.2.1"
      ttlSeconds: 300
    - recordType: "A"
      name: "us.geo-example.com."
      values:
        - "192.0.2.10"
      ttlSeconds: 300
    - recordType: "A"
      name: "eu.geo-example.com."
      values:
        - "192.0.2.20"
      ttlSeconds: 300
    - recordType: "A"
      name: "asia.geo-example.com."
      values:
        - "192.0.2.30"
      ttlSeconds: 300
    - recordType: "CNAME"
      name: "www.geo-example.com."
      values:
        - "geo-example.com."
      ttlSeconds: 300
```

This example supports geographic routing:
• Main domain A record
• Regional subdomains (us, eu, asia)
• CNAME for www subdomain
• Consistent TTL settings across regions.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the Route53 zone is active via the AWS console or by
using the AWS CLI:

```shell
aws route53 list-hosted-zones
```

For detailed zone information:

```shell
aws route53 get-hosted-zone --id <your-zone-id>
```

To list DNS records in the zone:

```shell
aws route53 list-resource-record-sets --hosted-zone-id <your-zone-id>
```

This will show the Route53 zone details including nameservers, record sets, and configuration information.

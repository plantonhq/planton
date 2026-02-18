# AWS Route53 Zone Examples

## Create Using CLI

Create a YAML manifest using one of the examples below. After the YAML is created, apply it with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

(You can also use the shorter form `planton apply -f <yaml-path>` if your environment is configured accordingly.)

---

## Basic Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
spec:
  region: us-east-1
  records:
    - recordType: A
      name: example.com.
      values:
        - "192.0.2.1"
      ttlSeconds: 300
```

This sets up a Route53 hosted zone for `example.com` and creates a simple A record pointing to the IP address
`192.0.2.1`.

---

## Multiple Records Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example-multi.com
spec:
  region: us-east-1
  records:
    - recordType: A
      name: www.example-multi.com.
      values:
        - "192.0.2.10"
      ttlSeconds: 300
    - recordType: CNAME
      name: api.example-multi.com.
      values:
        - "www.example-multi.com."
      ttlSeconds: 300
```

In this manifest, the hosted zone for `example-multi.com` has:
• An A record to route `www.example-multi.com` to `192.0.2.10`.
• A CNAME record `api.example-multi.com` that points to `www.example-multi.com.`.

---

## MX Record Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: mail-example.com
spec:
  region: us-east-1
  records:
    - recordType: MX
      name: mail-example.com.
      values:
        - "10 mailserver1.example.org."
        - "20 mailserver2.example.org."
      ttlSeconds: 3600
```

Here, the zone `mail-example.com` includes an MX record with two mail servers:
• Priority 10: `mailserver1.example.org.`
• Priority 20: `mailserver2.example.org.`

---

## TXT Record with Multiple Values

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: txt-example.com
spec:
  region: us-east-1
  records:
    - recordType: TXT
      name: txt-example.com.
      values:
        - "\"google-site-verification=abc123\""
        - "\"v=spf1 include:_spf.example.com ~all\""
      ttlSeconds: 300
```

This example adds a TXT record with multiple strings, which can be used for verification tokens or SPF records.

---

## Minimal Example (No Records)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: minimal.com
spec:
  region: us-east-1
  records: [ ]
```

This creates a hosted zone named `minimal.com` with no DNS records. It's useful when you simply want to provision the
zone first and add records later.

---

## Alias Record to CloudFront Distribution

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: cdn.example.com
spec:
  region: us-east-1
  records:
    - recordType: A
      name: cdn.example.com
      aliasTarget:
        dnsName: d1234abcd.cloudfront.net
        hostedZoneId: Z2FDTNDATAQYW2  # CloudFront's global hosted zone ID
        evaluateTargetHealth: false
```

This example creates an alias record at the zone apex pointing to a CloudFront distribution. Alias records are free (no query charges) and work at the apex where CNAMEs are not allowed.

---

## Alias Record to Application Load Balancer

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: api.example.com
spec:
  region: us-east-1
  records:
    - recordType: A
      name: api.example.com
      aliasTarget:
        dnsName: my-alb-1234567890.us-east-1.elb.amazonaws.com
        hostedZoneId: Z35SXDOTRQ7X7K  # ALB hosted zone ID for us-east-1
        evaluateTargetHealth: true
```

This routes traffic to an Application Load Balancer with health check evaluation enabled. Route53 will check the ALB's health before responding to queries.

---

## Weighted Routing for Blue/Green Deployment

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: app.example.com
spec:
  region: us-east-1
  records:
    # 90% of traffic to old version (blue)
    - recordType: A
      name: app.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 60
      setIdentifier: blue-90
      routingPolicy:
        weighted:
          weight: 90
    # 10% of traffic to new version (green)
    - recordType: A
      name: app.example.com
      values:
        - "192.0.2.2"
      ttlSeconds: 60
      setIdentifier: green-10
      routingPolicy:
        weighted:
          weight: 10
```

Weighted routing distributes traffic across multiple resources. Perfect for blue/green deployments, canary releases, or A/B testing. Adjust weights to gradually shift traffic from old to new versions.

---

## Latency-Based Routing for Global Applications

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: global.example.com
spec:
  region: us-east-1
  records:
    # US East region
    - recordType: A
      name: global.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 60
      setIdentifier: us-east-1
      routingPolicy:
        latency:
          region: us-east-1
    # Europe region
    - recordType: A
      name: global.example.com
      values:
        - "198.51.100.1"
      ttlSeconds: 60
      setIdentifier: eu-west-1
      routingPolicy:
        latency:
          region: eu-west-1
    # Asia Pacific region
    - recordType: A
      name: global.example.com
      values:
        - "203.0.113.1"
      ttlSeconds: 60
      setIdentifier: ap-southeast-1
      routingPolicy:
        latency:
          region: ap-southeast-1
```

Latency-based routing automatically directs users to the AWS region with the lowest network latency. Ideal for global applications serving users from multiple regions.

---

## Failover Routing with Health Checks

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: api.example.com
spec:
  region: us-east-1
  records:
    # Primary endpoint with health check
    - recordType: A
      name: api.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 60
      setIdentifier: primary
      healthCheckId: abc123-health-check-id
      routingPolicy:
        failover:
          type: PRIMARY
    # Secondary endpoint (disaster recovery)
    - recordType: A
      name: api.example.com
      values:
        - "192.0.2.2"
      ttlSeconds: 60
      setIdentifier: secondary
      routingPolicy:
        failover:
          type: SECONDARY
```

Failover routing enables active-passive disaster recovery. Route53 monitors the primary endpoint via health check and automatically fails over to the secondary if the primary becomes unhealthy.

**Note:** You must create the health check separately and provide its ID in the `healthCheckId` field.

---

## Geolocation Routing for Compliance

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: www.example.com
spec:
  region: us-east-1
  records:
    # EU users go to EU endpoint (GDPR compliance)
    - recordType: A
      name: www.example.com
      values:
        - "198.51.100.1"
      ttlSeconds: 300
      setIdentifier: europe
      routingPolicy:
        geolocation:
          continent: EU
    # US users go to US endpoint
    - recordType: A
      name: www.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 300
      setIdentifier: north-america
      routingPolicy:
        geolocation:
          continent: NA
    # Default for all other locations
    - recordType: A
      name: www.example.com
      values:
        - "203.0.113.1"
      ttlSeconds: 300
      setIdentifier: default
```

Geolocation routing directs traffic based on user's geographic location. Use for GDPR compliance, localized content, or geographic restrictions. You can route by continent, country, or US state.

---

## Private Hosted Zone with VPC Associations

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: internal.example.com
spec:
  region: us-east-1
  isPrivate: true
  vpcAssociations:
    - vpcId: vpc-12345678
      vpcRegion: us-east-1
    - vpcId: vpc-87654321
      vpcRegion: us-west-2
  records:
    - recordType: A
      name: db.internal.example.com
      values:
        - "10.0.1.100"
      ttlSeconds: 300
    - recordType: A
      name: cache.internal.example.com
      values:
        - "10.0.2.200"
      ttlSeconds: 300
```

Private hosted zones resolve only within associated VPCs, enabling split-horizon DNS for internal services. Perfect for microservices that should not be accessible from the internet.

**Requirements:**
- VPCs must have `enableDnsHostnames` and `enableDnsSupport` enabled
- At least one VPC association is required for private zones

---

## Zone with Query Logging and DNSSEC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: secure.example.com
spec:
  region: us-east-1
  enableQueryLogging: true
  queryLogGroupName: /aws/route53/secure.example.com
  enableDnssec: true
  records:
    - recordType: A
      name: secure.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 300
```

This example enables advanced security features:
- **Query Logging**: DNS queries are sent to CloudWatch Logs for debugging, security monitoring, and analytics
- **DNSSEC**: Cryptographic signatures prevent DNS spoofing attacks

**Prerequisites:**
- CloudWatch Log Group must exist before enabling query logging
- DNSSEC requires additional configuration at your domain registrar

**Warning:** High-traffic domains generate large query logs. Set CloudWatch Logs retention policies to avoid surprise bills.

---

## Complete Production Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: production.example.com
spec:
  region: us-east-1
  enableQueryLogging: true
  queryLogGroupName: /aws/route53/production
  records:
    # Apex domain aliased to CloudFront
    - recordType: A
      name: production.example.com
      aliasTarget:
        dnsName: d1234abcd.cloudfront.net
        hostedZoneId: Z2FDTNDATAQYW2
        evaluateTargetHealth: false
    
    # API with failover
    - recordType: A
      name: api.production.example.com
      values:
        - "192.0.2.1"
      ttlSeconds: 60
      setIdentifier: api-primary
      healthCheckId: abc123-primary-health
      routingPolicy:
        failover:
          type: PRIMARY
    
    - recordType: A
      name: api.production.example.com
      values:
        - "192.0.2.2"
      ttlSeconds: 60
      setIdentifier: api-secondary
      routingPolicy:
        failover:
          type: SECONDARY
    
    # Email with MX records
    - recordType: MX
      name: production.example.com
      values:
        - "10 mail1.production.example.com"
        - "20 mail2.production.example.com"
      ttlSeconds: 3600
    
    # SPF and DKIM for email security
    - recordType: TXT
      name: production.example.com
      values:
        - "v=spf1 include:_spf.example.com ~all"
      ttlSeconds: 300
    
    - recordType: TXT
      name: _dmarc.production.example.com
      values:
        - "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
      ttlSeconds: 300
```

This production-ready example demonstrates:
- Alias record for apex domain to CloudFront (cost-effective and performant)
- Failover routing for API with health checks (high availability)
- MX records for email routing
- TXT records for email security (SPF, DKIM, DMARC)
- Query logging for monitoring and debugging

---

## After Deploying

Once you've chosen one of these examples, apply it with OpenMCF using Pulumi or Terraform:

```shell
openmcf pulumi up --manifest aws-route53-zone.yaml --stack myorg/dev
```

or

```shell
openmcf terraform apply --manifest aws-route53-zone.yaml --stack myorg/dev
```

When the command completes successfully, your Route53 zone will be created in AWS. You can confirm by checking the AWS
console or by using the AWS CLI:

```shell
aws route53 list-hosted-zones
```

You should see your new hosted zone, along with any records that you defined in the manifest.

---

## Important Notes

### For Public Zones
After creating a public hosted zone, you **must** update your domain registrar with the Route53 name servers. You can find these in the stack outputs:

```shell
openmcf pulumi output nameservers --stack myorg/dev
```

Then configure these name servers at your domain registrar (GoDaddy, Namecheap, etc.).

### For Private Zones
Ensure your VPCs have DNS resolution enabled:

```shell
aws ec2 modify-vpc-attribute --vpc-id vpc-12345678 --enable-dns-hostnames
aws ec2 modify-vpc-attribute --vpc-id vpc-12345678 --enable-dns-support
```

### TTL Strategy
- **60 seconds**: Records you might change during incidents (failover, blue/green)
- **300 seconds** (5 min): Default for most records
- **3600 seconds** (1 hour): Static records (MX, NS)
- **86400 seconds** (1 day): Very static records to reduce query costs

**Pro Tip:** Lower TTL a day before planned changes, then raise it back after verification.

---

Happy routing with OpenMCF!

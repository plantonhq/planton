---
title: "Route53 Zone"
description: "Route53 Zone deployment documentation"
icon: "package"
order: 100
componentName: "awsroute53zone"
---

# AWS Route53 Zone

Deploys an AWS Route53 hosted zone with optional private zone configuration, VPC associations, DNSSEC, query logging, and inline DNS records supporting alias targets, routing policies, and health checks.

## What Gets Created

When you deploy an AwsRoute53Zone resource, OpenMCF provisions:

- **Route53 Hosted Zone** — a `route53.HostedZone` (AWS Native) resource, either public (internet-resolvable) or private (VPC-scoped), with the zone name derived from `metadata.name`
- **VPC Associations** — for private zones only, each `vpcAssociations` entry associates the zone with a VPC in a specified region, enabling DNS resolution from that VPC
- **DNSSEC Configuration** — when `enableDnssec` is `true`, a `route53.HostedZoneDnsSec` (AWS Classic) resource is created to enable cryptographic signing of DNS records
- **Query Logging** — when `enableQueryLogging` is `true`, a `route53.QueryLog` (AWS Classic) resource is created linking the zone to the specified CloudWatch Log Group
- **DNS Records** — for each entry in `records`, a `route53.Record` (AWS Classic) resource is created with support for basic records (A, AAAA, CNAME, MX, TXT, etc.), alias records pointing to AWS resources, and advanced routing policies (weighted, latency, failover, geolocation)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A domain name** that you own or control, used as `metadata.name` (e.g., `example.com`)
- **At least one VPC** with `enableDnsHostnames` and `enableDnsSupport` enabled, if creating a private zone
- **A CloudWatch Log Group** already created in the target region, if enabling query logging
- **Additional registrar configuration** for public zones — after deployment, update your domain registrar's nameservers to match the output nameservers

## Quick Start

Create a file `route53-zone.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRoute53Zone.example-com
spec: {}
```

Deploy:

```shell
openmcf apply -f route53-zone.yaml
```

This creates a public hosted zone for `example.com`. After deployment, update your domain registrar with the nameservers from the stack outputs.

## Configuration Reference

### Required Fields

No fields in `spec` are strictly required. A minimal `spec: {}` creates a public hosted zone using `metadata.name` as the zone name.

| Field | Type | Description |
|-------|------|-------------|
| `metadata.name` | `string` | The domain name for the hosted zone (e.g., `example.com`). Also used as the zone name in Route53. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isPrivate` | `bool` | `false` | When `true`, creates a private hosted zone that resolves only within associated VPCs. When `false`, creates a public zone resolvable on the internet. |
| `vpcAssociations` | `object[]` | `[]` | VPC associations for private zones. Required when `isPrivate` is `true`. At least one association must be specified for private zones. |
| `vpcAssociations[].vpcId` | `string` | — | The VPC ID to associate with the private zone. Required per association. |
| `vpcAssociations[].vpcRegion` | `string` | — | The AWS region where the VPC is located (e.g., `us-east-1`). Required per association. |
| `enableQueryLogging` | `bool` | `false` | Enables DNS query logging to CloudWatch Logs. Useful for debugging, security monitoring, and understanding query patterns. |
| `queryLogGroupName` | `string` | — | CloudWatch Log Group name for query logs. Required when `enableQueryLogging` is `true`. The log group must already exist. |
| `enableDnssec` | `bool` | `false` | Enables DNSSEC for the hosted zone. Adds cryptographic signatures to DNS records to prevent spoofing attacks. Requires additional configuration at the domain registrar level. |
| `records` | `object[]` | `[]` | Inline DNS records to create in the zone. Each entry creates a Route53 record resource. |
| `records[].recordType` | `DnsRecordType` | — | The DNS record type. Valid values: `A`, `AAAA`, `ALIAS`, `CNAME`, `MX`, `NS`, `PTR`, `SOA`, `SRV`, `TXT`, `CAA`. Required per record. |
| `records[].name` | `string` | — | The DNS record name (e.g., `www.example.com` or `*.example.com` for wildcards). Required per record. |
| `records[].ttlSeconds` | `int32` | `300` | Time to live in seconds. Ignored for alias records. Common values: 60 (fast changes), 300 (default), 86400 (static records). |
| `records[].values` | `string[]` | `[]` | Record values. Mutually exclusive with `aliasTarget`. Format depends on record type (e.g., `["192.0.2.1"]` for A records). |
| `records[].aliasTarget` | `object` | — | Alias target for Route53 alias records. Mutually exclusive with `values`. |
| `records[].aliasTarget.dnsName` | `string` | — | DNS name of the target resource (e.g., ALB DNS name). Required for alias records. |
| `records[].aliasTarget.hostedZoneId` | `string` | — | Hosted zone ID of the target AWS service (not your Route53 zone). Required for alias records. |
| `records[].aliasTarget.evaluateTargetHealth` | `bool` | `false` | When `true`, Route53 checks the health of the target before responding to queries. |
| `records[].routingPolicy` | `object` | — | Routing policy configuration. If not specified, simple routing is used. Only one policy type can be set. |
| `records[].routingPolicy.weighted` | `object` | — | Weighted routing. Distributes traffic based on assigned weights. |
| `records[].routingPolicy.weighted.weight` | `int32` | — | Weight value (0-255). Higher weight means more traffic. Weight of 0 stops traffic to this record. |
| `records[].routingPolicy.latency` | `object` | — | Latency-based routing. Routes to the lowest-latency endpoint. |
| `records[].routingPolicy.latency.region` | `string` | — | AWS region where this resource is located (e.g., `us-east-1`). Required for latency routing. |
| `records[].routingPolicy.failover` | `object` | — | Failover routing. Automatic failover to secondary when primary fails. |
| `records[].routingPolicy.failover.type` | `FailoverRecordType` | — | Must be `PRIMARY` or `SECONDARY`. Required for failover routing. |
| `records[].routingPolicy.geolocation` | `object` | — | Geolocation routing. Routes based on user location. |
| `records[].routingPolicy.geolocation.continent` | `string` | — | Two-letter continent code (e.g., `NA`, `EU`, `AS`). Use continent or country, not both. |
| `records[].routingPolicy.geolocation.country` | `string` | — | Two-letter ISO 3166-1 country code (e.g., `US`, `GB`, `DE`). |
| `records[].routingPolicy.geolocation.subdivision` | `string` | — | US state code (e.g., `CA`, `NY`). Only valid when country is `US`. |
| `records[].healthCheckId` | `string` | — | Route53 health check ID for failover routing. Only used with failover routing policy. |
| `records[].setIdentifier` | `string` | — | Unique identifier for routing policies. Required for weighted, latency, failover, and geolocation routing. Must be unique among records with the same name and type. |

## Examples

### Public Zone with A Records

A public zone with basic A and CNAME records:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRoute53Zone.example-com
spec:
  records:
    - recordType: A
      name: example.com
      ttlSeconds: 300
      values:
        - "203.0.113.10"
    - recordType: CNAME
      name: www.example.com
      ttlSeconds: 300
      values:
        - "example.com"
    - recordType: MX
      name: example.com
      ttlSeconds: 86400
      values:
        - "10 mail1.example.com"
        - "20 mail2.example.com"
    - recordType: TXT
      name: example.com
      ttlSeconds: 300
      values:
        - "v=spf1 include:_spf.google.com ~all"
```

### Private Zone with VPC Associations

A private zone for internal service discovery across two VPCs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: internal.example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRoute53Zone.internal-example-com
spec:
  isPrivate: true
  vpcAssociations:
    - vpcId: vpc-0a1b2c3d4e5f00001
      vpcRegion: us-east-1
    - vpcId: vpc-0a1b2c3d4e5f00002
      vpcRegion: us-west-2
  records:
    - recordType: A
      name: postgres.internal.example.com
      ttlSeconds: 60
      values:
        - "10.0.1.50"
    - recordType: A
      name: redis.internal.example.com
      ttlSeconds: 60
      values:
        - "10.0.1.51"
```

### Zone with Alias Records

A public zone with alias records pointing to an ALB and CloudFront:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRoute53Zone.example-com
spec:
  enableDnssec: true
  records:
    - recordType: A
      name: example.com
      aliasTarget:
        dnsName: my-alb-123456.us-east-1.elb.amazonaws.com
        hostedZoneId: Z35SXDOTRQ7X7K
        evaluateTargetHealth: true
    - recordType: A
      name: cdn.example.com
      aliasTarget:
        dnsName: d1234abcd.cloudfront.net
        hostedZoneId: Z2FDTNDATAQYW2
        evaluateTargetHealth: false
    - recordType: TXT
      name: example.com
      ttlSeconds: 300
      values:
        - "v=spf1 include:_spf.google.com ~all"
```

### Weighted Routing for Blue/Green Deployment

Split traffic between two endpoints using weighted routing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRoute53Zone.example-com
spec:
  records:
    - recordType: A
      name: api.example.com
      ttlSeconds: 60
      values:
        - "203.0.113.10"
      routingPolicy:
        weighted:
          weight: 90
      setIdentifier: blue
    - recordType: A
      name: api.example.com
      ttlSeconds: 60
      values:
        - "203.0.113.20"
      routingPolicy:
        weighted:
          weight: 10
      setIdentifier: green
```

### Failover Routing with Health Checks

Active-passive failover between primary and secondary endpoints:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRoute53Zone.example-com
spec:
  enableQueryLogging: true
  queryLogGroupName: /aws/route53/example-com
  records:
    - recordType: A
      name: app.example.com
      ttlSeconds: 60
      values:
        - "203.0.113.10"
      routingPolicy:
        failover:
          type: PRIMARY
      healthCheckId: hc-primary-12345
      setIdentifier: primary
    - recordType: A
      name: app.example.com
      ttlSeconds: 60
      values:
        - "203.0.113.20"
      routingPolicy:
        failover:
          type: SECONDARY
      setIdentifier: secondary
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The Route53 hosted zone ID (e.g., `Z1234567890ABCDEF`) |
| `zone_name` | `string` | The hosted zone name (matches `metadata.name`) |
| `nameservers` | `string[]` | The list of nameservers assigned to the zone. For public zones, update your domain registrar with these values. |

## Related Components

- [AwsRoute53DnsRecord](/docs/catalog/aws/route53-dns-record) — creates standalone DNS records in an existing hosted zone, useful for cross-team or cross-account record management
- [AwsAlb](/docs/catalog/aws/alb) — provides the DNS name and hosted zone ID for alias record targets
- [AwsCloudfront](/docs/catalog/aws/cloudfront) — provides the distribution DNS name for alias record targets
- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC IDs for private zone associations
- [AwsCertManagerCert](/docs/catalog/aws/certificate-manager-certificate) — DNS validation of ACM certificates requires records in the zone

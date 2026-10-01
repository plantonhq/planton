# AWS Route 53 DNS Record

Deploys a single DNS record in a Route53 hosted zone with support for every Route53 record type (A, AAAA, CNAME, MX, TXT, SRV, NS, SOA, PTR, CAA, DS, NAPTR, SPF, HTTPS, SVCB, SSHFP, TLSA), alias records pointing to AWS resources, and the full set of routing policies (weighted, latency, failover, geolocation, geoproximity, CIDR, multivalue answer). A record's identity in AWS is the (zone, name, type, setIdentifier) tuple -- changing the zone or the name replaces the record, while everything else updates in place.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Route53 DNS Record** -- a resource record in the specified hosted zone, either as a standard record with TTL and values or as an alias record pointing to an AWS resource (ALB, CloudFront, S3, API Gateway). With `allowOverwrite`, creation can adopt an existing record set with the same name and type instead of failing on the collision
- **Routing Policy Configuration** -- created only when `routingPolicy` is specified; configures weighted, latency-based, failover, geolocation, geoproximity (distance with a bias dial), CIDR (per-network), or multivalue-answer routing for the record
- **Health Check Association** -- configured only when `healthCheckId` is provided (literal ID or AwsRoute53HealthCheck reference); Route53 only serves the record while the health check passes
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Route53 hosted zone** -- the zone where the DNS record will be created. Provide the zone ID directly or reference an AwsRoute53Zone Cloud Resource via ValueFromRef.
- **An AWS resource** (optional, for alias records) -- the target resource (ALB, CloudFront distribution, S3 bucket, API Gateway) that the alias record will point to. Provide the DNS name and hosted zone ID directly or reference via ValueFromRef.
- **A Route53 health check** (optional) -- pairs with failover routing (and any non-simple policy where unhealthy answers should drop out). Reference an AwsRoute53HealthCheck Cloud Resource or pass a health check ID.
- **A Route53 CIDR collection** (optional, CIDR routing only) -- created through the Route53 API outside this resource; the record references it by collection ID and location name.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 DNS Record**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Alias Record to ALB** preset in the [Presets](#presets) tab to pre-populate a working alias record configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53DnsRecord
metadata:
  name: www-example
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  zoneId:
    value: "Z1234567890ABC"
  name: www.example.com
  type: A
  ttl: 300
  values:
    - value: "192.0.2.1"
```

```shell
planton apply -f dns-record.yaml
```

This creates a standard A record pointing `www.example.com` to the specified IP address with a 5-minute TTL. No alias target or routing policy is configured. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DNS record to a hosted zone and ALB deployed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: AwsRoute53Zone
      name: example-zone
      fieldPath: status.outputs.zone_id
  name: app.example.com
  type: A
  aliasTarget:
    dnsName:
      valueFrom:
        kind: AwsAlb
        name: app-alb
        fieldPath: status.outputs.load_balancer_dns_name
    zoneId:
      valueFrom:
        kind: AwsAlb
        name: app-alb
        fieldPath: status.outputs.load_balancer_hosted_zone_id
    evaluateTargetHealth: true
```

The InfraPipeline resolves the dependency graph, deploys the hosted zone and ALB first, then provisions the DNS record with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Route53 DNS record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Standard vs. alias record** -- Use `values` for standard records with explicit TTL. Use `aliasTarget` for alias records pointing to AWS resources -- alias records work at zone apex (where CNAME is prohibited), incur no Route53 query charges for AWS targets, and automatically track IP changes. The two are mutually exclusive.

**Routing policy** -- Defaults to simple routing (single-value response). Choose weighted for traffic splitting (blue/green, canary), latency for routing to the nearest region, failover for active-passive DR, geolocation for location-based routing (GDPR compliance, localized content), geoproximity for distance-based routing with a bias dial (gradual traffic shifts between neighboring regions), CIDR for per-network routing off a Route53 CIDR collection, or multivalue answer for client-side load balancing across up to eight healthy records (not compatible with alias records). Non-simple routing requires a `setIdentifier` -- and the coupling holds both ways: a `setIdentifier` on a simple record is rejected at authoring time, because AWS rejects it at change time.

**TTL** -- Required for standard records (`values`); forbidden on alias records (the target's TTL applies). AWS accepts 0 through 2147483647: an explicit `0` means "never cache" (every lookup hits Route 53 -- valid, billed per query). Use lower values (60s) for records you might change during incidents -- and for health-checked records, where slow caches defeat the point of failover. Use higher values (86400s) for static records like MX or NS to reduce query costs.

**Health checks** -- Attach a `healthCheckId` (literal ID or AwsRoute53HealthCheck reference) so Route53 only serves the record while the check passes. Most commonly paired with failover routing, but valid with any non-simple policy -- e.g. weighted records that drop out of rotation when unhealthy.

**Adopting existing records** -- Set `allowOverwrite: true` to take over a record set that already exists in the zone (auto-created NS/SOA records, or a record created outside the resource graph) instead of failing on the name/type collision. The existing values are replaced on the first deploy.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsRoute53Zone** | `zoneId` | `status.outputs.zone_id` |
| **AwsAlb** (optional) | `aliasTarget.dnsName` | `status.outputs.load_balancer_dns_name` |
| **AwsAlb** (optional) | `aliasTarget.zoneId` | `status.outputs.load_balancer_hosted_zone_id` |
| **AwsRoute53HealthCheck** (optional) | `healthCheckId` | `status.outputs.health_check_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `fqdn` | Fully qualified domain name of the created record | Application configuration, SSL certificate validation |
| `record_type` | DNS record type that was created (A, CNAME, etc.) | Debugging and inventory |
| `zone_id` | Hosted zone ID where the record was created | Record grouping and management |
| `is_alias` | Whether this is an alias record | Debugging and inventory |
| `set_identifier` | Set identifier for routing policies | Routing policy management |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Alias to ALB** -- An alias A record pointing a domain to an Application Load Balancer with target health evaluation enabled. Suitable for web applications and APIs behind an ALB where you need zone apex support and automatic failover. Start from the **Alias Record to ALB** preset.

**Standard A record** -- A basic A record mapping a subdomain to one or more IP addresses with a configurable TTL. Suitable for pointing to external services, on-premises resources, or static IP addresses. Start from the **Simple A Record** preset.

## Works With

- [**AWS Route 53 Zone**](/cloud-catalog/aws-route53-zone) -- provides the hosted zone where the DNS record is created
- [**AWS ALB**](/cloud-catalog/aws-alb) -- provides the load balancer DNS name and hosted zone ID for alias records
- [**AWS Route 53 Health Check**](/cloud-catalog/aws-route53-health-check) -- gates the record's answers on endpoint health
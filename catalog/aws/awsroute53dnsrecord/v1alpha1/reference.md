# AwsRoute53DnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRoute53DnsRecordSpec defines one DNS resource record set in a Route 53
hosted zone.

A record is either a STANDARD record (values + ttl) or an ALIAS record
(alias_target) — exactly one of the two. Alias records are Route 53's
killer feature: they point a name (including the zone apex, where CNAME is
forbidden) at an AWS resource such as an ALB, CloudFront distribution, S3
website, or API Gateway, with free queries and automatic target updates.

Beyond simple resolution, a record can carry ONE routing policy — weighted
(traffic splitting), latency (nearest region), failover (active-passive),
geolocation (user's location), geoproximity (bias-adjustable distance),
CIDR (per-subnet routing), or multivalue answer (client-side load
balancing). Records sharing a name and type but carrying different
set_identifier values form one routing group.

The record's identity in AWS is (zone, name, type, set_identifier) —
changing zone_id or name replaces the record (ForceNew).

## Example

```yaml
# AWS Route 53 DNS record — examples
#
# Usage:
#   planton apply -f manifest.yaml

apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53DnsRecord
metadata:
  name: test-dns-record
spec:
  region: us-east-1
  zoneId:
    value: Z0123456789ABCDEFGHIJ
  name: test.example.com
  type: A
  ttl: 300
  values:
    - value: 192.0.2.1

---
# Weighted canary record: 10% of traffic to the new stack. A second record
# with the same name/type and weight 90 completes the routing group.

apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53DnsRecord
metadata:
  name: canary-dns-record
spec:
  region: us-east-1
  zoneId:
    value: Z0123456789ABCDEFGHIJ
  name: api.example.com
  type: A
  ttl: 60
  values:
    - value: 192.0.2.10
  setIdentifier: canary
  routingPolicy:
    weighted:
      weight: 10
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.zoneId` | `string \| valueFrom` | yes |  | AwsRoute53Zone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.ttl` | `int32` |  |  |  |
| `spec.values` | `[]string \| valueFrom` |  |  |  |
| `spec.aliasTarget` | `AwsRoute53AliasTarget` |  |  |  |
| `spec.aliasTarget.dnsName` | `string \| valueFrom` | yes |  | AwsAlb (`status.outputs.load_balancer_dns_name`) |
| `spec.aliasTarget.zoneId` | `string \| valueFrom` | yes |  | AwsAlb (`status.outputs.load_balancer_hosted_zone_id`) |
| `spec.aliasTarget.evaluateTargetHealth` | `bool` |  |  |  |
| `spec.routingPolicy` | `AwsRoute53RoutingPolicy` |  |  |  |
| `spec.routingPolicy.weighted` | `AwsRoute53WeightedPolicy` |  |  |  |
| `spec.routingPolicy.weighted.weight` | `int32` |  |  |  |
| `spec.routingPolicy.latency` | `AwsRoute53LatencyPolicy` |  |  |  |
| `spec.routingPolicy.latency.region` | `string` | yes |  |  |
| `spec.routingPolicy.failover` | `AwsRoute53FailoverPolicy` |  |  |  |
| `spec.routingPolicy.failover.failoverType` | `string` | yes |  |  |
| `spec.routingPolicy.geolocation` | `AwsRoute53GeolocationPolicy` |  |  |  |
| `spec.routingPolicy.geolocation.continent` | `string` |  |  |  |
| `spec.routingPolicy.geolocation.country` | `string` |  |  |  |
| `spec.routingPolicy.geolocation.subdivision` | `string` |  |  |  |
| `spec.routingPolicy.geoproximity` | `AwsRoute53GeoproximityPolicy` |  |  |  |
| `spec.routingPolicy.geoproximity.awsRegion` | `string` |  |  |  |
| `spec.routingPolicy.geoproximity.coordinates` | `AwsRoute53Coordinates` |  |  |  |
| `spec.routingPolicy.geoproximity.coordinates.latitude` | `string` | yes |  |  |
| `spec.routingPolicy.geoproximity.coordinates.longitude` | `string` | yes |  |  |
| `spec.routingPolicy.geoproximity.localZoneGroup` | `string` |  |  |  |
| `spec.routingPolicy.geoproximity.bias` | `int32` |  |  |  |
| `spec.routingPolicy.cidr` | `AwsRoute53CidrPolicy` |  |  |  |
| `spec.routingPolicy.cidr.collectionId` | `string` | yes |  |  |
| `spec.routingPolicy.cidr.locationName` | `string` | yes |  |  |
| `spec.routingPolicy.multivalueAnswer` | `AwsRoute53MultivalueAnswerPolicy` |  |  |  |
| `spec.healthCheckId` | `string \| valueFrom` |  |  | AwsRoute53HealthCheck (`status.outputs.health_check_id`) |
| `spec.setIdentifier` | `string` |  |  |  |
| `spec.allowOverwrite` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Route 53 is a global service; this selects the region used for provider
API calls.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.zoneId

`string | valueFrom` · required

The Route 53 hosted zone that owns this record. Create-time immutable
(ForceNew). Can reference an AwsRoute53Zone resource — the default field
path wires to "status.outputs.zone_id".

- references: AwsRoute53Zone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRoute53Zone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The record name (fully qualified domain name or subdomain). Create-time
immutable (ForceNew).
Examples:
  - "example.com" for the zone apex
  - "www.example.com" for a subdomain
  - "*.example.com" for a wildcard (catch-all subdomains)
  - "_dmarc.example.com" / "_sip._tcp.example.com" — underscore-prefixed
    labels, the convention for records that configure services rather
    than name hosts (DMARC/DKIM email authentication, SRV service
    discovery, ACME challenges); Route 53 accepts them like any label.
Route 53 normalizes the trailing dot automatically.

- rule: {"required":true,"string":{"pattern":"^(?:\\*\\.[A-Za-z0-9_\\-\\.]+|[A-Za-z0-9_\\-\\.]+\\.[A-Za-z]{2,}|[A-Za-z0-9_\\-\\.]+)$"}}

### spec.type

`string` · required

The DNS record type. All Route 53 resource record set types are
supported:
- "A" / "AAAA": IPv4 / IPv6 addresses (also the alias record types).
- "CNAME": canonical-name redirection (not at the zone apex — use an
  A/AAAA alias there).
- "MX": mail exchangers ("<priority> <host>" values).
- "TXT": text data (SPF, DKIM, domain verification).
- "NS" / "SOA": delegation and authority records (usually managed by the
  zone itself; override TTLs only with care).
- "SRV": service locators ("<priority> <weight> <port> <target>").
- "PTR": reverse-DNS pointers.
- "CAA": certificate-authority authorization.
- "DS": delegation signer, for DNSSEC-signed child zone delegation.
- "NAPTR": name authority pointers (telephony/SIP).
- "SPF": legacy sender-policy type (RFC 7208 deprecates it — use TXT).
- "HTTPS" / "SVCB": service binding records (protocol hints, ECH).
- "SSHFP": SSH host key fingerprints.
- "TLSA": DANE TLS certificate association.

- rule: {"required":true,"string":{"in":["A","AAAA","CAA","CNAME","DS","HTTPS","MX","NAPTR","NS","PTR","SOA","SPF","SRV","SSHFP","SVCB","TLSA","TXT"]}}

### spec.ttl

`int32` · optional (explicit presence)

Time to live in seconds — how long resolvers cache the record. Required
for standard records; must be omitted for alias records (the target's
TTL applies). AWS accepts 0 to 2147483647; an explicit 0 means "never
cache" (every lookup hits Route 53 — valid, but billed per query).
Common values: 60 (fast cutover during incidents, and the recommended
ceiling for health-checked records), 300 (general default), 86400
(static records like MX/NS).

- rule: ttl must be between 0 and 2147483647 seconds (AWS's contract; 0 = never cache)

### spec.values

`[]string | valueFrom`

The record data for standard records. Format depends on type:
  - A: IPv4 addresses (e.g. ["192.0.2.1", "192.0.2.2"])
  - AAAA: IPv6 addresses (e.g. ["2001:db8::1"])
  - CNAME: one target hostname (e.g. ["target.example.com"])
  - MX: priority + mail server (e.g. ["10 mail1.example.com"])
  - TXT: text values (e.g. ["v=spf1 include:_spf.google.com ~all"])
Each value is at most 4,000 characters (AWS's per-value limit).
Mutually exclusive with alias_target — a record is standard or alias,
never both. Each entry can be a literal or a reference to another
resource's output — a load balancer's address, a zone's name server for
a delegation, a domain-verification record's value.

- rule: each literal value is at most 4000 characters (AWS's per-value limit)
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.aliasTarget

`AwsRoute53AliasTarget`

Alias target for A/AAAA alias records: point this name at an AWS
resource (ALB, NLB, CloudFront, S3 website, API Gateway, another record
in the zone) instead of literal addresses. Works at the zone apex,
queries are free, and Route 53 tracks the target's address changes
automatically. Mutually exclusive with values (and ttl).

### spec.aliasTarget.dnsName

`string | valueFrom` · required

The DNS name of the target resource.
Example literals:
  - CloudFront: "d1234abcd.cloudfront.net"
  - ALB: "my-alb-1234567890.us-east-1.elb.amazonaws.com"
  - S3 website: "my-bucket.s3-website-us-east-1.amazonaws.com"

- references: AwsAlb (`status.outputs.load_balancer_dns_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAlb, name: <that resource's name>, fieldPath: status.outputs.load_balancer_dns_name}} -- a bare string does not parse

### spec.aliasTarget.zoneId

`string | valueFrom` · required

The target AWS service's hosted zone ID for the target's region.
Example literals: CloudFront "Z2FDTNDATAQYW2" (global), ALB us-east-1
"Z35SXDOTRQ7X7K", S3 website us-east-1 "Z3AQBSTGFYJSTF".

- references: AwsAlb (`status.outputs.load_balancer_hosted_zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAlb, name: <that resource's name>, fieldPath: status.outputs.load_balancer_hosted_zone_id}} -- a bare string does not parse

### spec.aliasTarget.evaluateTargetHealth

`bool`

When true, Route 53 checks the target's own health (e.g. an ALB with no
healthy targets is treated as failed) before answering with this record.
The building block for alias-based failover.

### spec.routingPolicy

`AwsRoute53RoutingPolicy`

Routing policy for advanced traffic management. At most one policy per
record; records with the same name and type but different set_identifier
values combine into one routing group. Omit for simple routing.

### spec.routingPolicy.weighted

`AwsRoute53WeightedPolicy`

Weighted: split traffic across group members proportionally to their
weights. Blue/green deployments, canary releases, gradual migrations.

### spec.routingPolicy.weighted.weight

`int32`

Relative weight (0–255). Higher gets more traffic; 0 drains this record
(it is answered only if every group member is at 0 or unhealthy).

- rule: {"int32":{"lte":255,"gte":0}}

### spec.routingPolicy.latency

`AwsRoute53LatencyPolicy`

Latency: answer with the group member whose AWS region has the lowest
measured latency to the resolver. Multi-region applications.

### spec.routingPolicy.latency.region

`string` · required

The AWS region this record's resource lives in — the region whose
latency measurements represent this group member.
Example: "us-east-1", "eu-west-1", "ap-southeast-1"

Format-checked rather than enumerated on purpose: the provider's region
enum grows with every AWS region launch, and freezing today's list here
would reject tomorrow's regions until the next pin sweep. AWS itself
rejects unknown regions at change time.

- rule: {"required":true,"string":{"pattern":"^[a-z]{2,4}(-[a-z]+)+-[0-9]+$"}}

### spec.routingPolicy.failover

`AwsRoute53FailoverPolicy`

Failover: active-passive — the PRIMARY answers while healthy, the
SECONDARY takes over when the primary's health check fails.

### spec.routingPolicy.failover.failoverType

`string` · required

This record's role in the failover pair:
- "PRIMARY": answered while its health check passes.
- "SECONDARY": answered when the primary is unhealthy.

- rule: {"required":true,"string":{"in":["PRIMARY","SECONDARY"]}}

### spec.routingPolicy.geolocation

`AwsRoute53GeolocationPolicy`

Geolocation: answer based on WHERE the user is (continent, country, or
US state). Compliance boundaries, localized content.

- rule: geolocation routing requires continent, country, or subdivision

### spec.routingPolicy.geolocation.continent

`string`

Two-letter continent code: "AF", "AN", "AS", "EU", "OC", "NA", "SA"
(a frozen set — continents don't churn). Use continent OR country, not
both.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AF","AN","AS","EU","OC","NA","SA"]}}

### spec.routingPolicy.geolocation.country

`string`

Two-letter ISO 3166-1 alpha-2 country code (e.g. "US", "GB", "DE"), or
"*" for the default record that answers unmatched locations.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(\\*|[A-Z]{2})$"}}

### spec.routingPolicy.geolocation.subdivision

`string`

Subdivision code — US states only (e.g. "CA", "NY"); requires country
"US". At most 3 characters (AWS's limit).

- rule: {"string":{"maxLen":"3"}}

### spec.routingPolicy.geoproximity

`AwsRoute53GeoproximityPolicy`

Geoproximity: answer based on DISTANCE between the user and the
resource, with a bias dial to grow or shrink each resource's catchment
area. Traffic shifting between nearby regions.

- rule: geoproximity routing requires exactly one of aws_region, coordinates, or local_zone_group

### spec.routingPolicy.geoproximity.awsRegion

`string`

The AWS region hosting the resource (for resources in AWS regions).
Format-checked rather than enumerated (the region list grows; AWS
rejects unknown regions at change time).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]{2,4}(-[a-z]+)+-[0-9]+$"}}

### spec.routingPolicy.geoproximity.coordinates

`AwsRoute53Coordinates`

Latitude/longitude of the resource (for resources outside AWS).

### spec.routingPolicy.geoproximity.coordinates.latitude

`string` · required

Latitude in decimal degrees, "-90" to "90" (e.g. "40.71").

- rule: {"required":true}

### spec.routingPolicy.geoproximity.coordinates.longitude

`string` · required

Longitude in decimal degrees, "-180" to "180" (e.g. "-74.01").

- rule: {"required":true}

### spec.routingPolicy.geoproximity.localZoneGroup

`string`

The AWS Local Zone group hosting the resource (for Local Zone
deployments). Example: "us-east-1-bue-1".

### spec.routingPolicy.geoproximity.bias

`int32`

Expands (positive) or shrinks (negative) the geographic area this
resource answers for, from -99 to 99. Use it to shift traffic gradually
between neighboring locations.

- rule: {"int32":{"lte":99,"gte":-99}}

### spec.routingPolicy.cidr

`AwsRoute53CidrPolicy`

CIDR: answer based on the resolver's IP block, using a Route 53 CIDR
collection. Fine-grained per-network routing (e.g. ISP or office
egress ranges).

### spec.routingPolicy.cidr.collectionId

`string` · required

ID of the Route 53 CIDR collection holding the location's CIDR blocks.

- rule: {"required":true}

### spec.routingPolicy.cidr.locationName

`string` · required

Name of the location within the collection this record answers for, or
"*" for the collection's default location.

- rule: {"required":true}

### spec.routingPolicy.multivalueAnswer

`AwsRoute53MultivalueAnswerPolicy`

Multivalue answer: return up to eight healthy records so clients pick
among them — poor-man's load balancing with health-check awareness.
Not compatible with alias records.

### spec.healthCheckId

`string | valueFrom`

Health check gating this record's answers: Route 53 only serves the
record while the health check passes. Most commonly paired with failover
routing (primary/secondary), but valid with any non-simple routing
policy — e.g. weighted records that drop out of rotation when unhealthy.
Can reference an AwsRoute53HealthCheck resource.

- references: AwsRoute53HealthCheck (`status.outputs.health_check_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRoute53HealthCheck, name: <that resource's name>, fieldPath: status.outputs.health_check_id}} -- a bare string does not parse

### spec.setIdentifier

`string`

Distinguishes this record among records with the same name and type in a
routing group. Required by every routing policy; must be unique within
the group. Examples: "primary", "secondary", "us-east-1", "weight-70".

- rule: {"string":{"maxLen":"128"}}

### spec.allowOverwrite

`bool`

When true, creating this record overwrites an existing record set with
the same name and type instead of failing. Useful when adopting records
that were created outside the resource graph (e.g. a zone's auto-created
records or a manually-created record being brought under management).

## Validation Rules

- `values_or_alias_exclusive`: values and alias_target are mutually exclusive — a record is either a standard record or an alias record
- `values_or_alias_required`: either values (standard record) or alias_target (alias record) must be specified
- `alias_forbids_ttl`: ttl cannot be set on alias records (Route 53 uses the alias target's TTL)
- `values_require_ttl`: ttl is required for standard records (set 300 if unsure; 0 means never cache)
- `set_identifier_for_routing_policy`: set_identifier is required when a routing_policy is set (it distinguishes this record within its routing group)
- `set_identifier_requires_routing_policy`: set_identifier is only valid with a routing_policy (AWS rejects it on simple records)
- `multivalue_forbids_alias`: multivalue answer routing cannot be used with alias records (AWS limitation)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRoute53DnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.fqdn` | `string` | The fully qualified domain name (FQDN) of the created DNS record. Example: "www.example.com" or "example.com" for apex records. |
| `status.outputs.record_type` | `string` | The DNS record type that was created. Example: "A", "AAAA", "CNAME", "MX", "TXT" |
| `status.outputs.zone_id` | `string` | The hosted zone ID where the record was created. |
| `status.outputs.is_alias` | `bool` | Whether this is an alias record (pointing to an AWS resource). |
| `status.outputs.set_identifier` | `string` | The set identifier (if using routing policies). Empty for simple routing. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | AwsRoute53Zone | `status.outputs.zone_id` |
| `spec.aliasTarget.dnsName` | AwsAlb | `status.outputs.load_balancer_dns_name` |
| `spec.aliasTarget.zoneId` | AwsAlb | `status.outputs.load_balancer_hosted_zone_id` |
| `spec.healthCheckId` | AwsRoute53HealthCheck | `status.outputs.health_check_id` |

## See Also

- [Overview](../README.md)

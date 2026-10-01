# GcpDnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDnsRecordSpec creates one DNS record set inside an existing Cloud DNS
managed zone. A record set is the unit Cloud DNS manages: one (name, type)
pair holding either a static list of values (round-robin) or exactly one
routing policy (weighted, geolocation, or failover).

The record answers queries in one of two mutually exclusive ways:
  - values: static RRDATA returned to every resolver (the common case).
  - routing_policy: Cloud DNS steers each query by weight, caller
    geography, or primary/backup health — only one policy style per record.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: test-dns-record
spec:
  projectId:
    value: test-gcp-project
  managedZone:
    value: test-zone
  type: A
  name:
    value: test.example.com.
  values:
    - value: 192.0.2.1
  ttlSeconds: 300

  # Ephemeral test fixture: delete the record set on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.managedZone` | `string \| valueFrom` | yes |  | GcpDnsZone (`status.outputs.zone_name`) |
| `spec.type` | `string` | yes |  |  |
| `spec.name` | `string \| valueFrom` | yes |  |  |
| `spec.values` | `[]string \| valueFrom` |  |  |  |
| `spec.ttlSeconds` | `int32` |  | `300` |  |
| `spec.routingPolicy` | `GcpDnsRecordRoutingPolicy` |  |  |  |
| `spec.routingPolicy.wrr` | `[]GcpDnsRecordWrrPolicyItem` |  |  |  |
| `spec.routingPolicy.wrr[].weight` | `double` | yes |  |  |
| `spec.routingPolicy.wrr[].values` | `[]string \| valueFrom` |  |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets` | `GcpDnsRecordHealthCheckedTargets` |  |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers` | `[]GcpDnsRecordInternalLoadBalancerTarget` |  |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `string \| valueFrom` | yes |  | GcpAddress (`status.outputs.address`) |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipProtocol` | `string` | yes |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType` | `string` |  |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].port` | `string` | yes |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].region` | `string` |  |  |  |
| `spec.routingPolicy.wrr[].healthCheckedTargets.externalEndpoints` | `[]string` |  |  |  |
| `spec.routingPolicy.geo` | `[]GcpDnsRecordGeoPolicyItem` |  |  |  |
| `spec.routingPolicy.geo[].location` | `string` | yes |  |  |
| `spec.routingPolicy.geo[].values` | `[]string \| valueFrom` |  |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets` | `GcpDnsRecordHealthCheckedTargets` |  |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers` | `[]GcpDnsRecordInternalLoadBalancerTarget` |  |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `string \| valueFrom` | yes |  | GcpAddress (`status.outputs.address`) |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipProtocol` | `string` | yes |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType` | `string` |  |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].port` | `string` | yes |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].region` | `string` |  |  |  |
| `spec.routingPolicy.geo[].healthCheckedTargets.externalEndpoints` | `[]string` |  |  |  |
| `spec.routingPolicy.enableGeoFencing` | `bool` |  |  |  |
| `spec.routingPolicy.primaryBackup` | `GcpDnsRecordPrimaryBackupPolicy` |  |  |  |
| `spec.routingPolicy.primaryBackup.primary` | `GcpDnsRecordHealthCheckedTargets` | yes |  |  |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers` | `[]GcpDnsRecordInternalLoadBalancerTarget` |  |  |  |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipAddress` | `string \| valueFrom` | yes |  | GcpAddress (`status.outputs.address`) |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipProtocol` | `string` | yes |  |  |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].loadBalancerType` | `string` |  |  |  |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].networkUrl` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].port` | `string` | yes |  |  |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].region` | `string` |  |  |  |
| `spec.routingPolicy.primaryBackup.primary.externalEndpoints` | `[]string` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo` | `[]GcpDnsRecordGeoPolicyItem` | yes |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].location` | `string` | yes |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].values` | `[]string \| valueFrom` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets` | `GcpDnsRecordHealthCheckedTargets` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers` | `[]GcpDnsRecordInternalLoadBalancerTarget` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `string \| valueFrom` | yes |  | GcpAddress (`status.outputs.address`) |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipProtocol` | `string` | yes |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType` | `string` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].port` | `string` | yes |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].region` | `string` |  |  |  |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.externalEndpoints` | `[]string` |  |  |  |
| `spec.routingPolicy.primaryBackup.trickleRatio` | `double` |  |  |  |
| `spec.routingPolicy.primaryBackup.enableGeoFencingForBackups` | `bool` |  |  |  |
| `spec.routingPolicy.healthCheck` | `string \| valueFrom` |  |  | GcpHealthCheck (`status.outputs.self_link`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that hosts the managed zone.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used (ambient credentials
decide).

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.managedZone

`string | valueFrom` · required

The name of the managed zone this record lives in.
This is the zone RESOURCE name (e.g. "prod-example-zone"), not the DNS
name — Cloud DNS addresses zones by resource name.
Can be a literal value or a reference to a GcpDnsZone resource.

- references: GcpDnsZone (`status.outputs.zone_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_name}} -- a bare string does not parse

### spec.type

`string` · required

The DNS record type, uppercase (e.g. "A", "AAAA", "CNAME", "MX", "TXT",
"SRV", "NS", "PTR", "CAA", "SOA", "HTTPS", "SVCB", "DS", "DNSKEY",
"TLSA", "SSHFP", "NAPTR"). Cloud DNS accepts any record type the API
supports; the string is passed through as-is, so new types need no
spec change.

- rule: {"required":true,"string":{"pattern":"^[A-Z0-9]{1,10}$"}}

### spec.name

`string | valueFrom` · required

The fully qualified domain name this record set applies to.
Must end with a trailing dot (e.g. "www.example.com.").
A leading "*." creates a wildcard record; leading underscores support
service labels such as "_dmarc" and "_acme-challenge".
A literal FQDN or a reference to ANY kind's output (no default kind,
since the name can come from anywhere) -- e.g. a
GcpCertManagerDnsAuthorization's status.outputs.dns_record_name for
its _acme-challenge validation record. A referenced record is created
after the resource it reads.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.values

`[]string | valueFrom`

Static values (RRDATA) for the record set — the meaning depends on type:
  A: IPv4 addresses ("192.0.2.1")
  AAAA: IPv6 addresses ("2001:db8::1")
  CNAME: target hostname WITH trailing dot ("target.example.com.")
  MX: "priority mailserver." ("10 mail.example.com.")
  TXT: text values; values containing spaces need surrounding \" quotes,
       and single values longer than 255 characters (DKIM keys) must be
       split with "" between chunks.
Multiple values answer as a round-robin set. Mutually exclusive with
routing_policy. Each entry is a literal or a reference to ANY kind's
output, so a record can publish a value that only exists once that
resource does: a reserved address (a GcpGlobalAddress's or a
GcpAddress's status.outputs.address), one of a zone's name servers for
a subdomain delegation (a GcpDnsZone's status.outputs.nameservers.0,
.1, ...), or a validation target (a GcpCertManagerDnsAuthorization's
status.outputs.dns_record_data). A referenced record is created after
the resource it reads.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ttlSeconds

`int32` · optional (explicit presence)

Time to live in seconds — how long resolvers cache this record.
Common values: 60 (fast failover), 300 (default), 3600, 86400; NS
records conventionally use 172800 (2 days). Lower TTLs propagate
changes faster at the cost of more query load.

- default: `300`
- rule: {"int32":{"gte":0}}

### spec.routingPolicy

`GcpDnsRecordRoutingPolicy`

Query-steering policy — Cloud DNS answers each query based on weights,
caller geography, or target health instead of returning static values.
Mutually exclusive with values.

- rule: exactly one of wrr, geo, or primary_backup must be set
- rule: enable_geo_fencing applies only to geolocation routing (set geo entries)

### spec.routingPolicy.wrr

`[]GcpDnsRecordWrrPolicyItem`

Weighted round robin: traffic splits across entries in proportion to
their weights. Useful for canary rollouts and A/B traffic splitting.

### spec.routingPolicy.wrr[].weight

`double` · required · optional (explicit presence)

The ratio of traffic routed to this entry, relative to the sum of all
weights. A weight of 0 receives no traffic (useful for staging an
entry before shifting traffic onto it) — declared optional so the
explicit 0 is expressible while the field itself stays required.

- rule: {"required":true,"double":{"gte":0}}

### spec.routingPolicy.wrr[].values

`[]string | valueFrom`

Static values (RRDATA) answered for this entry.
If the zone has DNSSEC enabled, an entry may set only one of values or
health_checked_targets; otherwise both may be combined.
Each entry is a literal or a reference to ANY kind's output, as the
record's own values are -- e.g. a GcpGlobalAddress's
status.outputs.address for the backend this weight steers to.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.routingPolicy.wrr[].healthCheckedTargets

`GcpDnsRecordHealthCheckedTargets`

Load-balancer targets health-checked for this entry (A/AAAA records
only). Unhealthy targets are withdrawn from answers automatically.

- rule: at least one of internal_load_balancers or external_endpoints must be set

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers

`[]GcpDnsRecordInternalLoadBalancerTarget`

Internal load balancer frontends to health check. Cloud DNS reads the
load balancer's own health signal — no separate health check resource
is needed for these.

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipAddress

`string | valueFrom` · required

The frontend IP address of the load balancer.
Can be a literal IP or a reference to a GcpAddress resource (the
reserved internal VIP the forwarding rule serves on).

- references: GcpAddress (`status.outputs.address`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipProtocol

`string` · required

The IP protocol the load balancer frontend is configured for.
Case-sensitive: "tcp" or "udp".

- rule: {"required":true,"string":{"in":["tcp","udp"]}}

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType

`string`

The type of load balancer. Case-sensitive: "regionalL4ilb",
"regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.

- rule: load_balancer_type must be one of regionalL4ilb, regionalL7ilb, globalL7ilb

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].networkUrl

`string | valueFrom` · required

The fully qualified self-link URL of the VPC network the load balancer
belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
Can be a literal URL or a reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].port

`string` · required

The configured port of the load balancer frontend.

- rule: {"required":true}

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].project

`string | valueFrom` · required

The ID of the project the load balancer belongs to — may differ from
the record's project in Shared-VPC and cross-project topologies.
Can be a literal project ID or a reference to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].region

`string`

The region of the load balancer. Required for regional load balancers,
omitted for global ones.

### spec.routingPolicy.wrr[].healthCheckedTargets.externalEndpoints

`[]string`

Public internet IP addresses to health check. Requires the routing
policy's health_check to be set.

### spec.routingPolicy.geo

`[]GcpDnsRecordGeoPolicyItem`

Geolocation routing: each entry answers queries originating nearest to
its location. Useful for latency-sensitive multi-region serving.

### spec.routingPolicy.geo[].location

`string` · required

The Google Cloud location name this entry serves (e.g. "us-east1",
"europe-west3"). Queries are routed to the entry nearest the caller.

- rule: {"required":true}

### spec.routingPolicy.geo[].values

`[]string | valueFrom`

Static values (RRDATA) answered for this location.
Each entry is a literal or a reference to ANY kind's output, as the
record's own values are -- e.g. a regional GcpAddress's
status.outputs.address for the location's frontend.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.routingPolicy.geo[].healthCheckedTargets

`GcpDnsRecordHealthCheckedTargets`

Load-balancer targets health-checked for this location (A/AAAA records
only). Unhealthy targets are withdrawn from answers automatically.

- rule: at least one of internal_load_balancers or external_endpoints must be set

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers

`[]GcpDnsRecordInternalLoadBalancerTarget`

Internal load balancer frontends to health check. Cloud DNS reads the
load balancer's own health signal — no separate health check resource
is needed for these.

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipAddress

`string | valueFrom` · required

The frontend IP address of the load balancer.
Can be a literal IP or a reference to a GcpAddress resource (the
reserved internal VIP the forwarding rule serves on).

- references: GcpAddress (`status.outputs.address`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipProtocol

`string` · required

The IP protocol the load balancer frontend is configured for.
Case-sensitive: "tcp" or "udp".

- rule: {"required":true,"string":{"in":["tcp","udp"]}}

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType

`string`

The type of load balancer. Case-sensitive: "regionalL4ilb",
"regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.

- rule: load_balancer_type must be one of regionalL4ilb, regionalL7ilb, globalL7ilb

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].networkUrl

`string | valueFrom` · required

The fully qualified self-link URL of the VPC network the load balancer
belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
Can be a literal URL or a reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].port

`string` · required

The configured port of the load balancer frontend.

- rule: {"required":true}

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].project

`string | valueFrom` · required

The ID of the project the load balancer belongs to — may differ from
the record's project in Shared-VPC and cross-project topologies.
Can be a literal project ID or a reference to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].region

`string`

The region of the load balancer. Required for regional load balancers,
omitted for global ones.

### spec.routingPolicy.geo[].healthCheckedTargets.externalEndpoints

`[]string`

Public internet IP addresses to health check. Requires the routing
policy's health_check to be set.

### spec.routingPolicy.enableGeoFencing

`bool`

When true, geo queries are fenced: a location with unhealthy targets
keeps answering (with the unhealthy answer) instead of failing over to
the next-closest location. Applies to geo routing only.

### spec.routingPolicy.primaryBackup

`GcpDnsRecordPrimaryBackupPolicy`

Failover routing: queries are answered with the primary targets while
any are healthy, then fall back to a regional geo policy.

### spec.routingPolicy.primaryBackup.primary

`GcpDnsRecordHealthCheckedTargets` · required

The global primary targets. Queries are answered from these while any
target is healthy.

- rule: {"required":true}
- rule: at least one of internal_load_balancers or external_endpoints must be set

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers

`[]GcpDnsRecordInternalLoadBalancerTarget`

Internal load balancer frontends to health check. Cloud DNS reads the
load balancer's own health signal — no separate health check resource
is needed for these.

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipAddress

`string | valueFrom` · required

The frontend IP address of the load balancer.
Can be a literal IP or a reference to a GcpAddress resource (the
reserved internal VIP the forwarding rule serves on).

- references: GcpAddress (`status.outputs.address`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipProtocol

`string` · required

The IP protocol the load balancer frontend is configured for.
Case-sensitive: "tcp" or "udp".

- rule: {"required":true,"string":{"in":["tcp","udp"]}}

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].loadBalancerType

`string`

The type of load balancer. Case-sensitive: "regionalL4ilb",
"regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.

- rule: load_balancer_type must be one of regionalL4ilb, regionalL7ilb, globalL7ilb

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].networkUrl

`string | valueFrom` · required

The fully qualified self-link URL of the VPC network the load balancer
belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
Can be a literal URL or a reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].port

`string` · required

The configured port of the load balancer frontend.

- rule: {"required":true}

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].project

`string | valueFrom` · required

The ID of the project the load balancer belongs to — may differ from
the record's project in Shared-VPC and cross-project topologies.
Can be a literal project ID or a reference to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].region

`string`

The region of the load balancer. Required for regional load balancers,
omitted for global ones.

### spec.routingPolicy.primaryBackup.primary.externalEndpoints

`[]string`

Public internet IP addresses to health check. Requires the routing
policy's health_check to be set.

### spec.routingPolicy.primaryBackup.backupGeo

`[]GcpDnsRecordGeoPolicyItem` · required

The regional failover policy used when no primary target is healthy.

- rule: {"repeated":{"minItems":"1"}}

### spec.routingPolicy.primaryBackup.backupGeo[].location

`string` · required

The Google Cloud location name this entry serves (e.g. "us-east1",
"europe-west3"). Queries are routed to the entry nearest the caller.

- rule: {"required":true}

### spec.routingPolicy.primaryBackup.backupGeo[].values

`[]string | valueFrom`

Static values (RRDATA) answered for this location.
Each entry is a literal or a reference to ANY kind's output, as the
record's own values are -- e.g. a regional GcpAddress's
status.outputs.address for the location's frontend.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets

`GcpDnsRecordHealthCheckedTargets`

Load-balancer targets health-checked for this location (A/AAAA records
only). Unhealthy targets are withdrawn from answers automatically.

- rule: at least one of internal_load_balancers or external_endpoints must be set

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers

`[]GcpDnsRecordInternalLoadBalancerTarget`

Internal load balancer frontends to health check. Cloud DNS reads the
load balancer's own health signal — no separate health check resource
is needed for these.

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipAddress

`string | valueFrom` · required

The frontend IP address of the load balancer.
Can be a literal IP or a reference to a GcpAddress resource (the
reserved internal VIP the forwarding rule serves on).

- references: GcpAddress (`status.outputs.address`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipProtocol

`string` · required

The IP protocol the load balancer frontend is configured for.
Case-sensitive: "tcp" or "udp".

- rule: {"required":true,"string":{"in":["tcp","udp"]}}

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].loadBalancerType

`string`

The type of load balancer. Case-sensitive: "regionalL4ilb",
"regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.

- rule: load_balancer_type must be one of regionalL4ilb, regionalL7ilb, globalL7ilb

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].networkUrl

`string | valueFrom` · required

The fully qualified self-link URL of the VPC network the load balancer
belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
Can be a literal URL or a reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].port

`string` · required

The configured port of the load balancer frontend.

- rule: {"required":true}

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].project

`string | valueFrom` · required

The ID of the project the load balancer belongs to — may differ from
the record's project in Shared-VPC and cross-project topologies.
Can be a literal project ID or a reference to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].region

`string`

The region of the load balancer. Required for regional load balancers,
omitted for global ones.

### spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.externalEndpoints

`[]string`

Public internet IP addresses to health check. Requires the routing
policy's health_check to be set.

### spec.routingPolicy.primaryBackup.trickleRatio

`double` · optional (explicit presence)

Ratio of traffic (0.0–1.0) trickled to the backup targets even while
the primaries are healthy — keeps backup paths warm and verifiable.

- rule: {"double":{"lte":1,"gte":0}}

### spec.routingPolicy.primaryBackup.enableGeoFencingForBackups

`bool`

When true, backup geo queries are fenced (see
GcpDnsRecordRoutingPolicy.enable_geo_fencing).

### spec.routingPolicy.healthCheck

`string | valueFrom`

Health check used for public-IP health checking of routing-policy
targets. Can be a literal self-link or a reference to a GcpHealthCheck
resource. Internal load balancer targets carry their own implicit
health checking and do not need this.

- references: GcpHealthCheck (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHealthCheck, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.deletionPolicy

`string`

Deletion policy for the record set — what happens on destroy:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the record set is deleted; resolvers stop getting
               answers for this name/type as caches expire
  "PREVENT" -- destroy FAILS; protects a name other systems resolve
               (mail routing, domain verification, service discovery)
  "ABANDON" -- the record set is removed from management but keeps
               answering queries in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.values_xor_routing_policy`: exactly one of values or routing_policy must be set
- `spec.name_valid_fqdn_when_literal`: name must be a valid DNS domain name ending with a trailing dot when specified as a literal value (e.g., www.example.com.)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.fqdn` | `string` | The fully qualified domain name of the created DNS record. Example: "www.example.com." or "api.example.com." |
| `status.outputs.record_type` | `string` | The DNS record type that was created. Example: "A", "AAAA", "CNAME", "MX", "TXT". |
| `status.outputs.managed_zone` | `string` | The name of the managed zone containing this record. |
| `status.outputs.project_id` | `string` | The GCP project ID where the record was created. |
| `status.outputs.ttl_seconds` | `int32` | The TTL (time to live) in seconds for the DNS record. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.managedZone` | GcpDnsZone | `status.outputs.zone_name` |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | GcpAddress | `status.outputs.address` |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].project` | GcpProject | `status.outputs.project_id` |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | GcpAddress | `status.outputs.address` |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].project` | GcpProject | `status.outputs.project_id` |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipAddress` | GcpAddress | `status.outputs.address` |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].networkUrl` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].project` | GcpProject | `status.outputs.project_id` |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | GcpAddress | `status.outputs.address` |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].project` | GcpProject | `status.outputs.project_id` |
| `spec.routingPolicy.healthCheck` | GcpHealthCheck | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)

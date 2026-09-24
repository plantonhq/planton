# DigitalOceanLoadBalancer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanLoadBalancerSpec models the full digitalocean_loadbalancer
resource surface: regional and global balancer types, sizing (slug or
scaling units), forwarding rules with TLS termination or passthrough,
health checks with full threshold tuning, cookie-based sticky sessions,
backend targeting by Droplet references or tag, VPC/subnet placement,
network visibility and stack, an LB-level firewall, HTTPS redirect,
PROXY protocol, backend keepalive, idle-timeout tuning, TLS cipher
policy, project placement, bring-your-own-IP, and the global load
balancer's domains, target balancers, CDN, and regional failover
settings.

## Example

```yaml
# Example DigitalOceanLoadBalancer manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real balancer (regional HTTP, no VPC,
# no Droplets). The second exercises the regional surface: explicit VPC,
# size_unit, PROXY protocol, keepalive, idle-timeout, tag targeting,
# health-check thresholds, cookie sticky sessions, and a firewall.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanLoadBalancer
metadata:
  name: example-dolb-minimal
spec:
  loadBalancerName: example-dolb-minimal
  region: nyc3
  forwardingRules:
    - entryPort: 80
      entryProtocol: http
      targetPort: 80
      targetProtocol: http
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanLoadBalancer
metadata:
  name: example-dolb-full
spec:
  loadBalancerName: example-dolb-full
  region: nyc3
  vpc:
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  sizeUnit: 1
  enableProxyProtocol: true
  enableBackendKeepalive: true
  httpIdleTimeoutSeconds: 90
  dropletTag: web
  forwardingRules:
    - entryPort: 80
      entryProtocol: http
      targetPort: 8080
      targetProtocol: http
  healthCheck:
    port: 8080
    protocol: http
    path: /health
    checkIntervalSec: 15
    responseTimeoutSeconds: 5
    unhealthyThreshold: 3
    healthyThreshold: 5
  stickySessions:
    type: cookies
    cookieName: DO-LB
    cookieTtlSeconds: 300
  firewall:
    allow:
      - cidr:10.0.0.0/8
    deny:
      - cidr:192.0.2.0/24
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.loadBalancerName` | `string` | yes |  |  |
| `spec.region` | `enum` |  |  |  |
| `spec.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.forwardingRules` | `[]DigitalOceanLoadBalancerForwardingRule` |  |  |  |
| `spec.forwardingRules[].entryPort` | `uint32` | yes |  |  |
| `spec.forwardingRules[].entryProtocol` | `enum` | yes |  |  |
| `spec.forwardingRules[].targetPort` | `uint32` | yes |  |  |
| `spec.forwardingRules[].targetProtocol` | `enum` | yes |  |  |
| `spec.forwardingRules[].tlsPassthrough` | `bool` |  |  |  |
| `spec.forwardingRules[].certificateName` | `string \| valueFrom` |  |  | DigitalOceanCertificate (`status.outputs.certificate_id`) |
| `spec.healthCheck` | `DigitalOceanLoadBalancerHealthCheck` |  |  |  |
| `spec.healthCheck.port` | `uint32` | yes |  |  |
| `spec.healthCheck.protocol` | `enum` | yes |  |  |
| `spec.healthCheck.path` | `string` |  |  |  |
| `spec.healthCheck.checkIntervalSec` | `uint32` |  | `10` |  |
| `spec.healthCheck.responseTimeoutSeconds` | `uint32` |  | `5` |  |
| `spec.healthCheck.unhealthyThreshold` | `uint32` |  | `3` |  |
| `spec.healthCheck.healthyThreshold` | `uint32` |  | `5` |  |
| `spec.dropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |
| `spec.dropletTag` | `string` |  |  |  |
| `spec.stickySessions` | `DigitalOceanLoadBalancerStickySessions` |  |  |  |
| `spec.stickySessions.type` | `string` | yes |  |  |
| `spec.stickySessions.cookieName` | `string` |  |  |  |
| `spec.stickySessions.cookieTtlSeconds` | `uint32` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.size` | `string` |  |  |  |
| `spec.sizeUnit` | `uint32` |  |  |  |
| `spec.redirectHttpToHttps` | `bool` |  |  |  |
| `spec.enableProxyProtocol` | `bool` |  |  |  |
| `spec.enableBackendKeepalive` | `bool` |  |  |  |
| `spec.disableLetsEncryptDnsRecords` | `bool` |  |  |  |
| `spec.httpIdleTimeoutSeconds` | `uint32` |  |  |  |
| `spec.tlsCipherPolicy` | `string` |  |  |  |
| `spec.network` | `string` |  |  |  |
| `spec.networkStack` | `string` |  |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | DigitalOceanProject (`status.outputs.project_id`) |
| `spec.subnetUuid` | `string` |  |  |  |
| `spec.ip` | `string` |  |  |  |
| `spec.targetLoadBalancerIds` | `[]string \| valueFrom` |  |  | DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`) |
| `spec.firewall` | `DigitalOceanLoadBalancerFirewall` |  |  |  |
| `spec.firewall.allow` | `[]string` |  |  |  |
| `spec.firewall.deny` | `[]string` |  |  |  |
| `spec.domains` | `[]DigitalOceanLoadBalancerDomain` |  |  |  |
| `spec.domains[].name` | `string` | yes |  |  |
| `spec.domains[].isManaged` | `bool` |  |  |  |
| `spec.domains[].certificateName` | `string \| valueFrom` |  |  | DigitalOceanCertificate (`status.outputs.certificate_id`) |
| `spec.glbSettings` | `DigitalOceanLoadBalancerGlbSettings` |  |  |  |
| `spec.glbSettings.targetProtocol` | `string` | yes |  |  |
| `spec.glbSettings.targetPort` | `uint32` | yes |  |  |
| `spec.glbSettings.regionPriorities` | `map<string, uint32>` |  |  |  |
| `spec.glbSettings.failoverThreshold` | `uint32` |  |  |  |
| `spec.glbSettings.cdn` | `DigitalOceanLoadBalancerGlbCdn` |  |  |  |
| `spec.glbSettings.cdn.isEnabled` | `bool` |  |  |  |

## Field Details

### spec.loadBalancerName

`string` · required

The name of the load balancer. Must be unique per account.
Constraints: 1-64 characters, lowercase alphanumeric and hyphens.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[a-z0-9-]+$"}}

### spec.region

`enum`

The DigitalOcean region for a REGIONAL or REGIONAL_NETWORK balancer.
Required for those types and forbidden for GLOBAL balancers, which are
anycast and have no home region. Cannot be changed after creation.

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.vpc

`string | valueFrom`

(Optional) Reference to the DigitalOcean VPC in which to create the
load balancer. When unset, DigitalOcean places a regional balancer in
the region's default VPC; GLOBAL balancers take no VPC at all. Cannot
be changed after creation.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.forwardingRules

`[]DigitalOceanLoadBalancerForwardingRule`

Forwarding rules routing traffic from the balancer to the backends.
The regional balancer's core configuration; mutually exclusive with
glb_settings (global balancers route by domain instead).

- rule: http3 is valid only as an entry protocol

### spec.forwardingRules[].entryPort

`uint32` · required

Port on the load balancer that listens for incoming traffic.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.forwardingRules[].entryProtocol

`enum` · required

Protocol for incoming traffic on the entry port.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digitalocean_load_balancer_protocol_unspecified`
- `http`
- `https`
- `tcp`
- `http2`
- `http3`
- `udp`

### spec.forwardingRules[].targetPort

`uint32` · required

Port on the backend that receives forwarded traffic.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.forwardingRules[].targetProtocol

`enum` · required

Protocol for traffic between the balancer and the backend. http3 is
not valid here.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digitalocean_load_balancer_protocol_unspecified`
- `http`
- `https`
- `tcp`
- `http2`
- `http3`
- `udp`

### spec.forwardingRules[].tlsPassthrough

`bool`

(Optional) Whether TLS is passed through to the backends without
termination at the balancer. A passthrough rule carries no certificate.

### spec.forwardingRules[].certificateName

`string | valueFrom`

(Optional) TLS certificate for SSL termination, as a literal
certificate NAME or a reference to a DigitalOceanCertificate resource.
Required by the API when entry_protocol is https without
tls_passthrough. DigitalOcean identifies certificates by name here
because certificate UUIDs rotate when Let's Encrypt auto-renews;
the referenced kind's certificate_id output carries that stable name.

- references: DigitalOceanCertificate (`status.outputs.certificate_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanCertificate, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

### spec.healthCheck

`DigitalOceanLoadBalancerHealthCheck`

(Optional) Health check probing backend Droplets. When unset,
DigitalOcean applies a TCP check against the first forwarding rule's
target port.

- rule: path is required for http/https health checks and not allowed for tcp
- rule: health check protocol must be http, https, or tcp

### spec.healthCheck.port

`uint32` · required

The port on the backend to probe.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.healthCheck.protocol

`enum` · required

Protocol to probe with: http, https, or tcp.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digitalocean_load_balancer_protocol_unspecified`
- `http`
- `https`
- `tcp`
- `http2`
- `http3`
- `udp`

### spec.healthCheck.path

`string`

Request path for http/https probes (e.g. "/health"). Required for
http/https and not allowed for tcp.

### spec.healthCheck.checkIntervalSec

`uint32`

(Optional) Seconds between probes (3-300). 0 (unset) defers to
DigitalOcean's default of 10.

- default: `10`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":300,"gte":3}}

### spec.healthCheck.responseTimeoutSeconds

`uint32`

(Optional) Seconds to wait for a probe response (3-300). 0 (unset)
defers to DigitalOcean's default of 5.

- default: `5`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":300,"gte":3}}

### spec.healthCheck.unhealthyThreshold

`uint32`

(Optional) Consecutive probe failures (2-10) before a backend stops
receiving traffic. 0 (unset) defers to DigitalOcean's default of 3.

- default: `3`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":10,"gte":2}}

### spec.healthCheck.healthyThreshold

`uint32`

(Optional) Consecutive probe successes (2-10) before a backend receives
traffic again. 0 (unset) defers to DigitalOcean's default of 5.

- default: `5`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":10,"gte":2}}

### spec.dropletIds

`[]string | valueFrom`

(Optional) Specific Droplets to attach, as literal numeric Droplet IDs
or references to DigitalOceanDroplet resources. Mutually exclusive with
droplet_tag.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

### spec.dropletTag

`string`

(Optional) A Droplet tag: every Droplet carrying it is attached, and
membership follows the tag automatically as Droplets come and go.
The tag does not have to exist yet -- DigitalOcean stores it as a
selector and accepts a tag no Droplet carries (unlike a firewall, which
rejects an unknown tag); the balancer serves nothing until a Droplet in
its region picks up the tag. Mutually exclusive with droplet_ids.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.stickySessions

`DigitalOceanLoadBalancerStickySessions`

(Optional) Cookie-based session affinity. When unset, DigitalOcean
defaults to no sticky sessions ("none").

- rule: type cookies requires cookie_name and cookie_ttl_seconds; type none forbids them

### spec.stickySessions.type

`string` · required

Affinity mode: cookies pins clients to backends, none disables
affinity (the API default, useful to assert explicitly).

- rule: {"required":true,"string":{"in":["cookies","none"]}}

### spec.stickySessions.cookieName

`string`

Name of the affinity cookie (2-40 characters). Required with type
cookies.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"40"}}

### spec.stickySessions.cookieTtlSeconds

`uint32`

Lifetime of the affinity cookie in seconds. Required with type cookies.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.type

`string`

(Optional) The balancer type. REGIONAL (the default when unset) fronts
backends in one region; REGIONAL_NETWORK is the regional network
balancer; GLOBAL is the anycast balancer configured through
glb_settings, domains, and target_load_balancer_ids. Cannot be changed
after creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REGIONAL","GLOBAL","REGIONAL_NETWORK"]}}

### spec.size

`string`

(Optional) Balancer size slug. lb-small, lb-medium, and lb-large are
equivalent to size_unit 1, 3, and 6. Mutually exclusive with size_unit;
when both are unset DigitalOcean provisions lb-small.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["lb-small","lb-medium","lb-large"]}}

### spec.sizeUnit

`uint32`

(Optional) Balancer capacity in scaling units (1-200). Finer-grained
than the three size slugs and the only way past lb-large capacity.
Mutually exclusive with size.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":200,"gte":1}}

### spec.redirectHttpToHttps

`bool`

(Optional) Whether HTTP requests to port 80 are redirected to HTTPS on
port 443. Requires an HTTPS forwarding rule to redirect to.

### spec.enableProxyProtocol

`bool`

(Optional) Whether the PROXY protocol header is sent to the backends,
passing the client's real source address through. Backends must be
configured to accept the header.

### spec.enableBackendKeepalive

`bool`

(Optional) Whether HTTP keepalive connections are maintained to target
Droplets.

### spec.disableLetsEncryptDnsRecords

`bool`

(Optional) Whether the automatic DNS records for Let's Encrypt
certificates are NOT created. Only meaningful with a Let's Encrypt
certificate on a forwarding rule.

### spec.httpIdleTimeoutSeconds

`uint32`

(Optional) Seconds an idle HTTP connection stays open (the API accepts
30-600). 0 (unset) defers to DigitalOcean's default of 60.

### spec.tlsCipherPolicy

`string`

(Optional) TLS cipher policy for HTTPS/TLS forwarding rules. DEFAULT
accepts a broad cipher set; STRONG restricts to modern ciphers.
When unset, DigitalOcean applies DEFAULT.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DEFAULT","STRONG"]}}

### spec.network

`string`

(Optional) Network visibility: EXTERNAL (the default when unset) gives
the balancer a public address; INTERNAL keeps it reachable only inside
the VPC. Cannot be changed after creation. Never reported back by the
API, so drift on it is invisible.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["EXTERNAL","INTERNAL"]}}

### spec.networkStack

`string`

(Optional) IP stack: IPV4 (the default when unset) or DUALSTACK for
IPv4+IPv6. Cannot be changed after creation. Never reported back by
the API, so drift on it is invisible.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["IPV4","DUALSTACK"]}}

### spec.projectId

`string | valueFrom`

(Optional) The project the balancer is created in. Reference a
DigitalOceanProject resource (the default wiring resolves its
project_id output) or pass a literal project UUID. When unset, the
account's default project is used.

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.subnetUuid

`string`

(Optional) UUID of the DigitalOcean-managed VPC subnet to place the
balancer in. Literal only: subnets are DigitalOcean-assigned network
slices, not a Planton-managed kind. Requires vpc; cannot be changed
after creation.

### spec.ip

`string`

(Optional) Bring-your-own IP: an unassigned BYOIP address on the
account, in the balancer's region, assigned at creation. Consumed only
at create time; when unset DigitalOcean allocates the address. The
assigned address is exported as the ip stack output either way.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ip":true}}

### spec.targetLoadBalancerIds

`[]string | valueFrom`

(Optional) For GLOBAL balancers: the regional load balancers that
receive the routed traffic, as literal balancer UUIDs or references to
DigitalOceanLoadBalancer resources.

- references: DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.load_balancer_id}} -- a bare string does not parse

### spec.firewall

`DigitalOceanLoadBalancerFirewall`

(Optional) Balancer-level firewall controlling which sources may reach
it. When unset, all traffic is admitted.

### spec.firewall.allow

`[]string`

Sources ALLOWED to reach the balancer, each "ip:<address>" or
"cidr:<block>". The provider validates nothing here; catching a typo
before apply beats debugging silently-open traffic after.
matches() only: substring() is a Go-CEL extension and does not compile
on protovalidate-java (the control-plane engine). IPv4 is range-checked;
IPv6 is colon-hex (the provider accepts both).

- rule: {"repeated":{"items":{"cel":[{"id":"ip_or_cidr_rule","message":"must be 'ip:<address>' or 'cidr:<block>'","expression":"this.matches('^ip:((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$') || this.matches('^ip:[0-9a-fA-F:]*:[0-9a-fA-F:]+$') || this.matches('^cidr:((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/(3[0-2]|[12]?[0-9])$') || this.matches('^cidr:[0-9a-fA-F:]*:[0-9a-fA-F:]+/[0-9]{1,3}$')"}]}}}

### spec.firewall.deny

`[]string`

Sources DENIED from reaching the balancer, each "ip:<address>" or
"cidr:<block>".

- rule: {"repeated":{"items":{"cel":[{"id":"ip_or_cidr_rule","message":"must be 'ip:<address>' or 'cidr:<block>'","expression":"this.matches('^ip:((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$') || this.matches('^ip:[0-9a-fA-F:]*:[0-9a-fA-F:]+$') || this.matches('^cidr:((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/(3[0-2]|[12]?[0-9])$') || this.matches('^cidr:[0-9a-fA-F:]*:[0-9a-fA-F:]+/[0-9]{1,3}$')"}]}}}

### spec.domains

`[]DigitalOceanLoadBalancerDomain`

(Optional) For GLOBAL balancers: the domains that ingress traffic to
the balancer.

### spec.domains[].name

`string` · required

The domain name.

- rule: {"required":true,"string":{"hostname":true}}

### spec.domains[].isManaged

`bool`

(Optional) Whether the domain is managed by DigitalOcean (a DigitalOcean
DNS zone), letting the balancer manage its records.

### spec.domains[].certificateName

`string | valueFrom`

(Optional) TLS certificate for the domain's HTTPS handshake, as a
literal certificate NAME or a reference to a DigitalOceanCertificate
resource. DigitalOcean identifies certificates by name because
certificate UUIDs rotate when Let's Encrypt auto-renews; the referenced
kind's certificate_id output carries that stable name.

- references: DigitalOceanCertificate (`status.outputs.certificate_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanCertificate, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

### spec.glbSettings

`DigitalOceanLoadBalancerGlbSettings`

(Optional) For GLOBAL balancers: routing configuration. Mutually
exclusive with forwarding_rules.

### spec.glbSettings.targetProtocol

`string` · required

Protocol used toward the regional targets: http or https.

- rule: {"required":true,"string":{"in":["http","https"]}}

### spec.glbSettings.targetPort

`uint32` · required

Port used toward the regional targets: 80 or 443.

- rule: {"required":true,"uint32":{"in":[80,443]}}

### spec.glbSettings.regionPriorities

`map<string, uint32>`

(Optional) Priority per region slug (lower is preferred) for steering
traffic across the regional targets.

### spec.glbSettings.failoverThreshold

`uint32`

(Optional) Percentage of failed requests (1-99) that triggers failover
away from a region. The API only reports it back when
region_priorities is also set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":99,"gte":1}}

### spec.glbSettings.cdn

`DigitalOceanLoadBalancerGlbCdn`

(Optional) CDN caching at the global edge.

### spec.glbSettings.cdn.isEnabled

`bool`

Whether edge caching is enabled.

## Validation Rules

- `droplet_ids_xor_tag`: droplet_ids and droplet_tag are mutually exclusive
- `forwarding_rules_xor_glb_settings`: exactly one of forwarding_rules or glb_settings must be set
- `region_by_type`: region must be empty for GLOBAL balancers and set for all other types
- `subnet_requires_vpc`: subnet_uuid requires vpc to be set
- `size_xor_size_unit`: size and size_unit are mutually exclusive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanLoadBalancer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.load_balancer_id` | `string` | The unique identifier (UUID) of the created DigitalOcean Load Balancer. |
| `status.outputs.ip` | `string` | The public IPv4 address assigned to the Load Balancer. |
| `status.outputs.urn` | `string` | The uniform resource name (URN) of the Load Balancer, usable with DigitalOcean project resources. |
| `status.outputs.ipv6` | `string` | The IPv6 address of the Load Balancer, populated when network_stack is DUALSTACK. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.forwardingRules[].certificateName` | DigitalOceanCertificate | `status.outputs.certificate_id` |
| `spec.dropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |
| `spec.projectId` | DigitalOceanProject | `status.outputs.project_id` |
| `spec.targetLoadBalancerIds` | DigitalOceanLoadBalancer | `status.outputs.load_balancer_id` |
| `spec.domains[].certificateName` | DigitalOceanCertificate | `status.outputs.certificate_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanFirewall | `spec.inboundRules[].sourceLoadBalancerUids` | `status.outputs.load_balancer_id` |
| DigitalOceanFirewall | `spec.outboundRules[].destinationLoadBalancerUids` | `status.outputs.load_balancer_id` |
| DigitalOceanLoadBalancer | `spec.targetLoadBalancerIds` | `status.outputs.load_balancer_id` |
| DigitalOceanMonitorAlert | `spec.loadBalancerIds` | `status.outputs.load_balancer_id` |

## See Also

- [Overview](../README.md)

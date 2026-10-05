# CloudflareLoadBalancerPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareLoadBalancerPoolSpec defines an account-scoped pool of origin servers
for Cloudflare Load Balancing. A pool groups origins, health-checks them via a
referenced monitor, and is selected by one or more zone-scoped
CloudflareLoadBalancers (as default, fallback, or geo-routed pools). Pools are
reusable across load balancers and have an independent lifecycle.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancerPool
metadata:
  name: test-pool
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: web-pool
  origins:
    - name: origin-1
      address:
        value: "203.0.113.10"
      weight: 1
      enabled: true
    - name: origin-2
      address:
        value: "203.0.113.11"
      weight: 1
      enabled: true
  monitor:
    value: "a1b2c3d4e5f60718293a4b5c6d7e8f90"
  minimumOrigins: 1
  checkRegions:
    - WNAM
    - WEU
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.origins` | `[]CloudflareLoadBalancerPoolOrigin` | yes |  |  |
| `spec.origins[].name` | `string` | yes |  |  |
| `spec.origins[].address` | `string \| valueFrom` | yes |  |  |
| `spec.origins[].weight` | `double` |  | `1` |  |
| `spec.origins[].enabled` | `bool` |  | `true` |  |
| `spec.origins[].port` | `int32` |  |  |  |
| `spec.origins[].hostHeader` | `string` |  |  |  |
| `spec.origins[].virtualNetworkId` | `string` |  |  |  |
| `spec.origins[].flattenCname` | `bool` |  | `true` |  |
| `spec.monitor` | `string \| valueFrom` |  |  | CloudflareLoadBalancerMonitor (`status.outputs.monitor_id`) |
| `spec.checkRegions` | `[]enum` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.minimumOrigins` | `int32` |  | `1` |  |
| `spec.latitude` | `double` |  |  |  |
| `spec.longitude` | `double` |  |  |  |
| `spec.loadShedding` | `CloudflareLoadBalancerPoolLoadShedding` |  |  |  |
| `spec.loadShedding.defaultPercent` | `double` |  |  |  |
| `spec.loadShedding.defaultPolicy` | `string` |  |  |  |
| `spec.loadShedding.sessionPercent` | `double` |  |  |  |
| `spec.loadShedding.sessionPolicy` | `string` |  |  |  |
| `spec.originSteering` | `CloudflareLoadBalancerPoolOriginSteering` |  |  |  |
| `spec.originSteering.policy` | `string` |  |  |  |
| `spec.notificationFilter` | `CloudflareLoadBalancerPoolNotificationFilter` |  |  |  |
| `spec.notificationFilter.origin` | `CloudflareLoadBalancerPoolNotificationFilterRule` |  |  |  |
| `spec.notificationFilter.origin.disable` | `bool` |  |  |  |
| `spec.notificationFilter.origin.healthy` | `bool` |  |  |  |
| `spec.notificationFilter.pool` | `CloudflareLoadBalancerPoolNotificationFilterRule` |  |  |  |
| `spec.notificationFilter.pool.disable` | `bool` |  |  |  |
| `spec.notificationFilter.pool.healthy` | `bool` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this pool. Must match the account of the
referenced monitor and of any load balancer that uses this pool.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

A short name (tag) for the pool. Only alphanumeric characters, hyphens, and
underscores are allowed.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32","pattern":"^[A-Za-z0-9_-]+$"}}

### spec.origins

`[]CloudflareLoadBalancerPoolOrigin` · required

The origin servers in this pool. Traffic to a healthy pool is balanced
across its currently-healthy origins. At least one origin is required.

- rule: {"repeated":{"minItems":"1"}}

### spec.origins[].name

`string` · required

A human-identifiable name for the origin.

- rule: {"required":true}

### spec.origins[].address

`string | valueFrom` · required

The origin address: an IP (v4/v6) or a publicly addressable hostname that
resolves directly to the origin (not a Cloudflare-proxied hostname). Accepts
a literal or a reference to another resource's output (e.g. a compute
instance's public IP or a load balancer hostname), so origins compose in the
resource graph. To use an internal/reserved address, also set
virtual_network_id.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.origins[].weight

`double` · optional (explicit presence)

The weight of this origin relative to others in the pool (0.0–1.0). Defaults
to 1 when unset. Used by weighted origin-steering policies.

- default: `1`
- rule: weight must be between 0 and 1

### spec.origins[].enabled

`bool` · optional (explicit presence)

Whether this origin is enabled within the pool. A disabled origin receives no
traffic and is excluded from health checks. Defaults to true when unset.

- default: `true`

### spec.origins[].port

`int32`

The upstream port for connections. Leave 0 to use the default port for the
monitor's protocol.

- rule: port must be 0 (protocol default) or between 1 and 65535

### spec.origins[].hostHeader

`string`

Override the HTTP Host header sent to this origin (Cloudflare supports a
single Host override per origin). Leave empty to send the default Host.

### spec.origins[].virtualNetworkId

`string`

The virtual-network ID this origin belongs to, required when address is an
internal/reserved address reachable over a Cloudflare Tunnel virtual network.

### spec.origins[].flattenCname

`bool` · optional (explicit presence)

Whether to flatten CNAME origin addresses to their A/AAAA records before
returning to the client. Defaults to true when unset.

- default: `true`

### spec.monitor

`string | valueFrom`

The monitor that health-checks this pool's origins, as a literal monitor ID
or a reference to a CloudflareLoadBalancerMonitor. When omitted, origins are
always considered healthy (not recommended for production).

- references: CloudflareLoadBalancerMonitor (`status.outputs.monitor_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerMonitor, name: <that resource's name>, fieldPath: status.outputs.monitor_id}} -- a bare string does not parse

### spec.checkRegions

`[]enum`

Regions from which to run health checks. Leave empty to check from every
Cloudflare data center.

- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `check_region_unspecified` -- Unspecified — health checks run from every Cloudflare data center.
- `WNAM`
- `ENAM`
- `WEU`
- `EEU`
- `NSAM`
- `SSAM`
- `OC`
- `ME`
- `NAF`
- `SAF`
- `SAS`
- `SEAS`
- `NEAS`
- `ALL_REGIONS`

### spec.description

`string`

Human-readable description shown in the Cloudflare dashboard.

### spec.enabled

`bool` · optional (explicit presence)

Whether the pool is enabled. A disabled pool receives no traffic and is
excluded from health checks. Defaults to true when unset.

- default: `true`

### spec.minimumOrigins

`int32`

The minimum number of origins that must be healthy for the pool to serve
traffic; below this the pool fails over. Leave 0 for the Cloudflare default (1).

- default: `1`
- rule: minimum_origins must be 0 (default) or a positive number

### spec.latitude

`double` · optional (explicit presence)

Latitude (decimal degrees) of the data center hosting these origins, used by
proximity steering. If set, longitude must also be set.

- rule: latitude must be between -90 and 90

### spec.longitude

`double` · optional (explicit presence)

Longitude (decimal degrees) of the data center hosting these origins, used by
proximity steering. If set, latitude must also be set.

- rule: longitude must be between -180 and 180

### spec.loadShedding

`CloudflareLoadBalancerPoolLoadShedding`

Load-shedding policy: sheds a configurable percentage of traffic away from
the pool. Omit to shed nothing.

### spec.loadShedding.defaultPercent

`double`

Percent of new sessions / unaffinitized traffic to shed (0–100).

- rule: default_percent must be between 0 and 100

### spec.loadShedding.defaultPolicy

`string`

Policy for shedding default traffic: "random" sheds a random percent of
requests; "hash" sheds all requests from a percent of client IPs. Empty
defaults to "random".

- rule: default_policy must be "random" or "hash"

### spec.loadShedding.sessionPercent

`double`

Percent of existing sessions to shed (0–100).

- rule: session_percent must be between 0 and 100

### spec.loadShedding.sessionPolicy

`string`

Policy for shedding existing sessions. Only "hash" is supported (to avoid
exponential decay). Empty defaults to "hash".

- rule: session_policy must be "hash"

### spec.originSteering

`CloudflareLoadBalancerPoolOriginSteering`

Origin-steering policy: how origins are selected for new sessions and traffic
without session affinity. Omit for Cloudflare's default (random).

### spec.originSteering.policy

`string`

Origin-steering policy. One of "random", "hash",
"least_outstanding_requests", or "least_connections". Empty defaults to
"random".

- rule: policy must be one of "random", "hash", "least_outstanding_requests", "least_connections"

### spec.notificationFilter

`CloudflareLoadBalancerPoolNotificationFilter`

Filters pool/origin health-status notifications. Omit to use account defaults.

### spec.notificationFilter.origin

`CloudflareLoadBalancerPoolNotificationFilterRule`

Notification filter for individual origins.

### spec.notificationFilter.origin.disable

`bool`

Disable notifications for this resource type entirely.

### spec.notificationFilter.origin.healthy

`bool` · optional (explicit presence)

When set, send notifications only for this health status (e.g. false to
notify only on DOWN events). Omit to notify on all events.

### spec.notificationFilter.pool

`CloudflareLoadBalancerPoolNotificationFilterRule`

Notification filter for the pool as a whole.

### spec.notificationFilter.pool.disable

`bool`

Disable notifications for this resource type entirely.

### spec.notificationFilter.pool.healthy

`bool` · optional (explicit presence)

When set, send notifications only for this health status (e.g. false to
notify only on DOWN events). Omit to notify on all events.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareLoadBalancerPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pool_id` | `string` | The Cloudflare-assigned identifier of the pool. A CloudflareLoadBalancer references this value in its default_pools, fallback_pool, or geo-pool maps. |
| `status.outputs.pool_name` | `string` | The pool name (echoed for convenience). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.monitor` | CloudflareLoadBalancerMonitor | `status.outputs.monitor_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareLoadBalancer | `spec.defaultPools` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.fallbackPool` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.regionPools[].poolIds` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.countryPools[].poolIds` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.popPools[].poolIds` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.rules[].overrides.countryPools[].poolIds` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.rules[].overrides.defaultPools` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.rules[].overrides.fallbackPool` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.rules[].overrides.popPools[].poolIds` | `status.outputs.pool_id` |
| CloudflareLoadBalancer | `spec.rules[].overrides.regionPools[].poolIds` | `status.outputs.pool_id` |

## See Also

- [Overview](../README.md)

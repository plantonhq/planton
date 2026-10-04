# AzureTrafficManagerProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureTrafficManagerProfileSpec** defines a Traffic Manager profile
-- Azure's DNS-based traffic director. The profile owns a public DNS
name ({relative_name}.trafficmanager.net) and answers each lookup
with the address of one of its endpoints
(AzureTrafficManagerEndpoint), chosen by the routing method:
Performance (closest), Priority (failover order), Weighted
(proportional spread), Geographic (by the caller's region), Subnet
(by the caller's source IP), or MultiValue (all healthy addresses at
once). Because the steering happens in DNS, Traffic Manager fronts
anything with a resolvable address -- across regions, clouds, and
on-premises -- and is not in the data path.

Traffic Manager is a GLOBAL service: a profile lives in no Azure
region (its resource group's region is just where the metadata
record sits). This is why this spec omits the `region` field that
other Azure resources include.

The profile continuously health-probes every endpoint per
monitor_config and only answers with healthy ones. Endpoint objects
are separate resources (AzureTrafficManagerEndpoint) referencing
this profile -- one profile serves many endpoints.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: Performance
# routing, the complete monitor shape (HTTPS probe with path, custom
# status ranges, a Host header, zero-tolerance failures), an explicit
# low DNS TTL, Traffic View, and user tags merged over the derived
# ones. (max_return belongs to MultiValue routing -- its shape rides a
# dedicated offline plan per the profile's record.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureTrafficManagerProfile
metadata:
  name: test-traffic-manager-profile
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: platform-rg
  name: app-director
  routingMethod: Performance
  dnsConfig:
    relativeName: contoso-app-director
    ttlSeconds: 30
  monitorConfig:
    protocol: HTTPS
    port: 443
    path: /healthz
    intervalInSeconds: 30
    timeoutInSeconds: 9
    toleratedNumberOfFailures: 0
    expectedStatusCodeRanges:
      - 200-299
      - 301-301
    customHeaders:
      - name: Host
        value: app.contoso.com
  trafficViewEnabled: true
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.routingMethod` | `string` | yes |  |  |
| `spec.dnsConfig` | `AzureTrafficManagerDnsConfig` | yes |  |  |
| `spec.dnsConfig.relativeName` | `string` | yes |  |  |
| `spec.dnsConfig.ttlSeconds` | `int32` |  | `60` |  |
| `spec.monitorConfig` | `AzureTrafficManagerMonitorConfig` | yes |  |  |
| `spec.monitorConfig.protocol` | `string` | yes |  |  |
| `spec.monitorConfig.port` | `int32` | yes |  |  |
| `spec.monitorConfig.path` | `string` |  |  |  |
| `spec.monitorConfig.intervalInSeconds` | `int32` |  | `30` |  |
| `spec.monitorConfig.timeoutInSeconds` | `int32` |  | `10` |  |
| `spec.monitorConfig.toleratedNumberOfFailures` | `int32` |  | `3` |  |
| `spec.monitorConfig.expectedStatusCodeRanges` | `[]string` |  |  |  |
| `spec.monitorConfig.customHeaders` | `[]AzureTrafficManagerCustomHeader` |  |  |  |
| `spec.monitorConfig.customHeaders[].name` | `string` | yes |  |  |
| `spec.monitorConfig.customHeaders[].value` | `string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.maxReturn` | `int32` |  |  |  |
| `spec.trafficViewEnabled` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the profile's metadata record lives in
(the profile itself is global). Can be a literal resource-group
name or a reference to an AzureResourceGroup's name output.
Changing it replaces the profile.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The profile's resource name, unique within the resource group.
Changing the name replaces the profile. (The DNS name users
resolve is dns_config.relative_name, not this.)

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingMethod

`string` · required

How lookups are answered. Updatable in place -- switching methods
re-steers traffic without touching endpoints.
  - "Performance": the endpoint with the lowest network latency to
    the caller.
  - "Priority": the healthy endpoint with the lowest priority
    value (active/passive failover).
  - "Weighted": endpoints in proportion to their weight values.
  - "Geographic": the endpoint whose geo_mappings claim the
    caller's region (every region used must be claimed by exactly
    one endpoint).
  - "Subnet": the endpoint whose subnets claim the caller's source
    IP.
  - "MultiValue": ALL healthy endpoint addresses in one answer
    (max_return caps how many; endpoints must be external IPv4/
    IPv6 addresses).

- rule: {"required":true,"string":{"in":["Performance","Priority","Weighted","Geographic","Subnet","MultiValue"]}}

### spec.dnsConfig

`AzureTrafficManagerDnsConfig` · required

The profile's DNS identity: the name it answers on and how long
resolvers may cache the answer.

- rule: {"required":true}

### spec.dnsConfig.relativeName

`string` · required

The DNS label the profile answers on: {relative_name}
.trafficmanager.net. Globally unique across ALL of Azure (the
trafficmanager.net namespace is shared) -- Azure rejects a taken
name at apply time. Changing it replaces the profile. Point your
own domain at the profile with a CNAME to this generated name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dnsConfig.ttlSeconds

`int32` · optional (explicit presence)

Time to live in seconds for the profile's DNS answers: how long
resolvers cache an answer -- and therefore how long clients keep
using an endpoint after Traffic Manager stops handing it out. Low
TTLs (30-120) make failover visible quickly at the cost of more
queries (queries are the billing meter). Unspecified applies 60,
the Azure portal's own default -- the modules always send the
effective value explicitly.

- default: `60`
- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.monitorConfig

`AzureTrafficManagerMonitorConfig` · required

How the profile health-probes its endpoints. Only healthy
endpoints enter DNS answers.

- rule: {"required":true}
- rule: With interval_in_seconds 10 (fast interval), set timeout_in_seconds explicitly to 5-9 -- the default 10 does not fit inside the probe window

### spec.monitorConfig.protocol

`string` · required

The probe protocol. HTTP/HTTPS probe path and expect a status in
expected_status_code_ranges (or 200 when unset); TCP considers a
completed handshake healthy.

- rule: {"required":true,"string":{"in":["HTTP","HTTPS","TCP"]}}

### spec.monitorConfig.port

`int32` · required · optional (explicit presence)

The port probed on each endpoint (e.g. 443 for HTTPS).

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.monitorConfig.path

`string`

The path probed on HTTP/HTTPS (e.g. "/healthz"). Meaningless for
TCP probes. Unset probes "/".

### spec.monitorConfig.intervalInSeconds

`int32` · optional (explicit presence)

Seconds between probes of each endpoint. Azure offers exactly two
cadences: 30 (the default, included in the profile price) and 10
(fast interval, billed extra per endpoint). Unspecified applies
30 -- the modules always send the effective value explicitly.

- default: `30`
- rule: {"int32":{"in":[10,30]}}

### spec.monitorConfig.timeoutInSeconds

`int32` · optional (explicit presence)

Seconds a probe waits for a response before counting a failure.
5-10 with the 30-second interval; the 10-second fast interval
narrows it to 5-9 (set it explicitly then -- the 10 default no
longer fits). Unspecified applies 10 -- the modules always send
the effective value explicitly.

- default: `10`
- rule: {"int32":{"lte":10,"gte":5}}

### spec.monitorConfig.toleratedNumberOfFailures

`int32` · optional (explicit presence)

Consecutive probe failures before an endpoint is marked degraded
and leaves DNS answers. 0 degrades on the first failure.
Unspecified applies 3 -- the modules always send the effective
value explicitly.

- default: `3`
- rule: {"int32":{"lte":9,"gte":0}}

### spec.monitorConfig.expectedStatusCodeRanges

`[]string`

HTTP status ranges counted healthy, each as "min-max" (e.g.
"200-299", "301-301"). Unset expects exactly 200. HTTP/HTTPS only.

- rule: {"repeated":{"items":{"string":{"pattern":"^[0-9]{3}-[0-9]{3}$"}}}}

### spec.monitorConfig.customHeaders

`[]AzureTrafficManagerCustomHeader`

Headers sent with every HTTP/HTTPS probe (e.g. a Host header for
name-based virtual hosting). Endpoints can override per-endpoint.

### spec.monitorConfig.customHeaders[].name

`string` · required

The header name (e.g. "Host").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.monitorConfig.customHeaders[].value

`string` · required

The header value.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the profile answers queries. Unspecified applies true, the
provider's default -- set false to park the profile (its DNS name
stops resolving to endpoints) without deleting it. Updatable in
place.

- default: `true`

### spec.maxReturn

`int32` · optional (explicit presence)

For MultiValue routing only: the maximum number of healthy
endpoint addresses returned in one DNS answer (1-8). REQUIRED when
routing_method is "MultiValue" -- the provider rejects a
MultiValue profile without it.

- rule: {"int32":{"lte":8,"gte":1}}

### spec.trafficViewEnabled

`bool`

Enables Traffic View: Azure aggregates the profile's DNS query
patterns into per-region latency and volume analytics (billed per
million queries processed). Off by default.

### spec.tags

`map<string, string>`

Free-form tags applied to the profile, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Updatable in place.

## Validation Rules

- `azure_traffic_manager_multivalue_requires_max_return`: MultiValue routing requires max_return (1-8) -- the number of healthy addresses returned per answer

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureTrafficManagerProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.traffic_manager_profile_id` | `string` | The profile's ARM resource ID (.../providers/Microsoft.Network/trafficManagerProfiles/{name}) -- what AzureTrafficManagerEndpoint references, and what alias DNS records target. |
| `status.outputs.traffic_manager_profile_name` | `string` | The profile's resource name within its resource group. |
| `status.outputs.fqdn` | `string` | The profile's public DNS name ({relative_name}.trafficmanager.net) -- what users resolve, and what your own domain CNAMEs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureTrafficManagerEndpoint | `spec.profileId` | `status.outputs.traffic_manager_profile_id` |
| AzureTrafficManagerEndpoint | `spec.nested.targetProfileId` | `status.outputs.traffic_manager_profile_id` |

## See Also

- [Overview](../README.md)

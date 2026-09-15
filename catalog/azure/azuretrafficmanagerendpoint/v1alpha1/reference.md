# AzureTrafficManagerEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureTrafficManagerEndpointSpec** defines one destination a Traffic
Manager profile (AzureTrafficManagerProfile) can steer traffic to.

The endpoint type is declared by which variant block is present: set
exactly one of `azure`, `external`, or `nested` -- mirroring Azure's
three endpoint types.
  - `azure`: a public Azure resource, by ARM ID (a Public IP, an App
    Service, another profile's fronted service). Azure tracks the
    target's address itself, so the endpoint follows the resource.
  - `external`: anything with a DNS name or IP address -- services in
    other clouds, on-premises, or Azure resources you prefer to
    address by name.
  - `nested`: ANOTHER Traffic Manager profile, composing routing
    methods into trees (e.g. a Performance parent choosing between
    regional Weighted children).

Fields every type shares (weight, priority, enabled, health-probe
headers, geo/subnet claims) live at the spec root; each variant block
carries only what its type actually accepts.

**Which shared fields matter depends on the PROFILE's routing
method** (Azure evaluates them there; nothing enforces them here):
weight steers Weighted profiles, priority steers Priority profiles,
geo_mappings steer Geographic profiles (every region must be claimed
by exactly one endpoint), subnets steer Subnet profiles (ranges must
not overlap across endpoints). MultiValue profiles require external
endpoints whose targets are literal IPv4/IPv6 addresses.

## Example

```yaml
# Offline-plan test manifest. Exercises the external variant at depth:
# an explicit endpoint location (the Performance-routing requirement),
# always-serve, an explicit weight and priority, and a per-endpoint
# probe Host header. (The azure and nested variants ride dedicated
# offline plans per the profile's record; subnet and geo claims belong
# to Subnet/Geographic-routed profiles and ride the subnet-claims
# offline plan.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureTrafficManagerEndpoint
metadata:
  name: test-traffic-manager-endpoint
  org: test-org
  env: dev
spec:
  profileId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/trafficManagerProfiles/app-director
  name: eastus-app
  external:
    target:
      value: app-eastus.contoso.com
    endpointLocation: eastus
    alwaysServeEnabled: true
  weight: 100
  priority: 1
  customHeaders:
    - name: Host
      value: app-eastus.contoso.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.profileId` | `string \| valueFrom` | yes |  | AzureTrafficManagerProfile (`status.outputs.traffic_manager_profile_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.azure` | `AzureTrafficManagerAzureEndpoint` |  |  |  |
| `spec.azure.targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.azure.alwaysServeEnabled` | `bool` |  | `false` |  |
| `spec.external` | `AzureTrafficManagerExternalEndpoint` |  |  |  |
| `spec.external.target` | `string \| valueFrom` | yes |  |  |
| `spec.external.endpointLocation` | `string` |  |  |  |
| `spec.external.alwaysServeEnabled` | `bool` |  | `false` |  |
| `spec.nested` | `AzureTrafficManagerNestedEndpoint` |  |  |  |
| `spec.nested.targetProfileId` | `string \| valueFrom` | yes |  | AzureTrafficManagerProfile (`status.outputs.traffic_manager_profile_id`) |
| `spec.nested.minimumChildEndpoints` | `int32` | yes |  |  |
| `spec.nested.minimumRequiredChildEndpointsIpv4` | `int32` |  |  |  |
| `spec.nested.minimumRequiredChildEndpointsIpv6` | `int32` |  |  |  |
| `spec.nested.endpointLocation` | `string` |  |  |  |
| `spec.weight` | `int32` |  | `1` |  |
| `spec.priority` | `int32` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.geoMappings` | `[]string` |  |  |  |
| `spec.subnets` | `[]AzureTrafficManagerEndpointSubnet` |  |  |  |
| `spec.subnets[].first` | `string` | yes |  |  |
| `spec.subnets[].last` | `string` |  |  |  |
| `spec.subnets[].scope` | `int32` |  |  |  |
| `spec.customHeaders` | `[]AzureTrafficManagerEndpointCustomHeader` |  |  |  |
| `spec.customHeaders[].name` | `string` | yes |  |  |
| `spec.customHeaders[].value` | `string` | yes |  |  |

## Field Details

### spec.profileId

`string | valueFrom` · required

The Traffic Manager profile this endpoint belongs to, by ARM
resource ID -- defaults to referencing an
AzureTrafficManagerProfile's traffic_manager_profile_id output.
Changing it replaces the endpoint.

- references: AzureTrafficManagerProfile (`status.outputs.traffic_manager_profile_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureTrafficManagerProfile, name: <that resource's name>, fieldPath: status.outputs.traffic_manager_profile_id}} -- a bare string does not parse

### spec.name

`string` · required

The endpoint's name, unique per type within the profile. Changing
it replaces the endpoint.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.azure

`AzureTrafficManagerAzureEndpoint`

A public Azure resource as the destination. Set exactly one
variant block on this spec.

### spec.azure.targetResourceId

`string | valueFrom` · required

The target resource's ARM ID. The resource must expose a public IP
address (a Public IP, an App Service, ...). Reference the owning
kind's id output with an explicit valueFrom (e.g. an
AzurePublicIp's public_ip_id) or pass a literal ID -- no kind
dominates Azure endpoint targets, so there is no default.
Retargeting updates the endpoint in place.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.azure.alwaysServeEnabled

`bool` · optional (explicit presence)

Serve this endpoint even when probes call it unhealthy (health
checks are disabled). Useful when probes cannot reach the target
but users can. Off by default. Updatable in place.

- default: `false`

### spec.external

`AzureTrafficManagerExternalEndpoint`

A DNS name or IP address as the destination. Set exactly one
variant block on this spec.

### spec.external.target

`string | valueFrom` · required

The destination: a fully qualified DNS name or an IP address
(MultiValue profiles require literal IPv4/IPv6 addresses). A
reference or a literal -- no kind dominates external targets, so
references declare their kind explicitly. Retargeting updates the
endpoint in place.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.external.endpointLocation

`string`

The Azure region Traffic Manager treats as this endpoint's
location, for latency-based decisions (e.g. "eastus"). REQUIRED by
the service when the profile routes by Performance (external
targets carry no discoverable region); ignored otherwise --
enforced apply-time, the shape only the service can check.

### spec.external.alwaysServeEnabled

`bool` · optional (explicit presence)

Serve this endpoint even when probes call it unhealthy (health
checks are disabled). Useful when probes cannot reach the target
but users can. Off by default. Updatable in place.

- default: `false`

### spec.nested

`AzureTrafficManagerNestedEndpoint`

Another Traffic Manager profile as the destination. Set exactly
one variant block on this spec.

### spec.nested.targetProfileId

`string | valueFrom` · required

The CHILD profile's ARM resource ID -- defaults to referencing an
AzureTrafficManagerProfile's traffic_manager_profile_id output.
Retargeting updates the endpoint in place. The endpoint lives in its
own parent profile (profile_id) and merely points at this one, so the
reference is access, not placement, on a diagram.

- references: AzureTrafficManagerProfile (`status.outputs.traffic_manager_profile_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureTrafficManagerProfile, name: <that resource's name>, fieldPath: status.outputs.traffic_manager_profile_id}} -- a bare string does not parse

### spec.nested.minimumChildEndpoints

`int32` · required · optional (explicit presence)

The minimum number of healthy endpoints the child profile must
hold for THIS endpoint to count as healthy (at least 1). Updatable
in place.

- rule: {"required":true,"int32":{"gte":1}}

### spec.nested.minimumRequiredChildEndpointsIpv4

`int32` · optional (explicit presence)

Of the child's healthy endpoints, how many must be IPv4-addressed
for this endpoint to count as healthy. 0 (or unset) applies no
IPv4 floor -- the modules send the value only when positive,
mirroring the provider. Updatable in place.

- rule: {"int32":{"gte":0}}

### spec.nested.minimumRequiredChildEndpointsIpv6

`int32` · optional (explicit presence)

Of the child's healthy endpoints, how many must be IPv6-addressed
for this endpoint to count as healthy. 0 (or unset) applies no
IPv6 floor -- the modules send the value only when positive,
mirroring the provider. Updatable in place.

- rule: {"int32":{"gte":0}}

### spec.nested.endpointLocation

`string`

The Azure region Traffic Manager treats as this endpoint's
location, for latency-based decisions (e.g. "eastus"). REQUIRED by
the service when the PARENT profile routes by Performance; ignored
otherwise -- enforced apply-time, the shape only the service can
check.

### spec.weight

`int32` · optional (explicit presence)

For Weighted profiles: this endpoint's share of traffic relative
to its siblings (1-1000). Unspecified applies 1, the provider's
default -- the modules always send the effective value explicitly.
Updatable in place.

- default: `1`
- rule: {"int32":{"lte":1000,"gte":1}}

### spec.priority

`int32` · optional (explicit presence)

For Priority profiles: this endpoint's failover position (1-1000,
lower serves first, no two endpoints may share a value). Leave
unset to let Azure assign the next free value in creation order --
the service owns the default, so the modules send it only when
set. Updatable in place.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the endpoint participates in routing. Unspecified applies
true, the provider's default -- set false to drain the endpoint
(it leaves DNS answers) without deleting it. Updatable in place.

- default: `true`

### spec.geoMappings

`[]string`

For Geographic profiles: the regions this endpoint answers for, as
Traffic Manager geographic-hierarchy codes ("WORLD", "GEO-EU",
"DE", ...). Azure requires every code to be claimed by exactly ONE
endpoint in the profile -- claims are validated apply-time against
the live hierarchy. Updatable in place.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.subnets

`[]AzureTrafficManagerEndpointSubnet`

For Subnet profiles: the caller source-IP ranges this endpoint
answers for. Ranges must not overlap across the profile's
endpoints (Azure enforces apply-time). FIXED AT CREATION -- any
change replaces the endpoint.

### spec.subnets[].first

`string` · required

The range's first IPv4 address (or its only one).

- rule: {"required":true,"string":{"ipv4":true}}

### spec.subnets[].last

`string`

The range's last IPv4 address, inclusive (range form).

- rule: last must be an IPv4 address

### spec.subnets[].scope

`int32` · optional (explicit presence)

The CIDR prefix length applied to first (CIDR form), 0-32.

- rule: {"int32":{"lte":32,"gte":0}}

### spec.customHeaders

`[]AzureTrafficManagerEndpointCustomHeader`

Health-probe headers for THIS endpoint, overriding the profile's
monitor_config headers on collision (e.g. a per-endpoint Host
header). Updatable in place.

### spec.customHeaders[].name

`string` · required

The header name (e.g. "Host").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customHeaders[].value

`string` · required

The header value.

- rule: {"required":true,"string":{"minLen":"1"}}

## Validation Rules

- `azure_traffic_manager_endpoint_exactly_one_variant`: Set exactly one endpoint variant -- azure, external, or nested -- the variant determines the endpoint type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureTrafficManagerEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint_id` | `string` | The endpoint's ARM resource ID. Format: {profile_id}/{TYPE}/{name} where {TYPE} is AzureEndpoints, ExternalEndpoints, or NestedEndpoints per the spec's variant. |
| `status.outputs.endpoint_name` | `string` | The endpoint's name within its profile. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.profileId` | AzureTrafficManagerProfile | `status.outputs.traffic_manager_profile_id` |
| `spec.nested.targetProfileId` | AzureTrafficManagerProfile | `status.outputs.traffic_manager_profile_id` |

## See Also

- [Overview](../README.md)
